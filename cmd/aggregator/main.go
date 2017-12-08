package main

import (
	"context"
	"fmt"
	"github.com/aptible/mini-collector/api"
	"github.com/aptible/mini-collector/batch"
	"github.com/aptible/mini-collector/batcher"
	"github.com/aptible/mini-collector/emitter"
	"github.com/aptible/mini-collector/emitter/influxdb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"time"
)

const (
	port = ":50051"
)

var (
	requiredTags = []string{
		"environment",
		"service",
		"container",
	}

	optionalTags = []string{
		"app",
		"database",
	}
)

type server struct {
	batcher batcher.Batcher
}

func (s *server) Publish(ctx context.Context, point *api.PublishRequest) (*api.PublishResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, fmt.Errorf("no metadata")
	}

	ts := time.Unix(int64(point.UnixTime), 0)

	tags := map[string]string{}

	for _, k := range requiredTags {
		v, ok := md[k]
		if !ok {
			return nil, fmt.Errorf("missing required metadata key: %s", k)
		}
		tags[k] = v[0]
	}

	for _, k := range optionalTags {
		v, ok := md[k]
		if !ok {
			continue
		}
		tags[k] = v[0]
	}

	s.batcher.Ingest(&batch.Entry{
		Time:           ts,
		Tags:           tags,
		PublishRequest: *point,
	})
	return &api.PublishResponse{}, nil
}

func getEmitter() (emitter.Emitter, error) {
	influxDbConfiguration, ok := os.LookupEnv("AGGREGATOR_INFLUXDB_CONFIGURATION")
	if ok {
		log.Infof("using InfluxDB emitter")
		return influxdb.New(influxDbConfiguration)
	}

	return nil, fmt.Errorf("no emitter configured")
}

func main() {
	emitter, err := getEmitter()
	if err != nil {
		log.Fatalf("failed to get emitter: %v", err)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	api.RegisterAggregatorServer(s, &server{
		batcher: batcher.New(emitter),
	})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
