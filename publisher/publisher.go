package publisher

import (
	"golang.org/x/net/context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"github.com/aptible/mini-collector/api"
	"github.com/aptible/mini-collector/collector"
)

// TODO: Make it an inteface
type Publisher struct {
	connection *grpc.ClientConn
	publishChannel chan collector.Point
}

func Open() (*Publisher, error) {
	addr := "localhost:50051"

	connection, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		// TODO: Wrap error
		return nil, err
	}

	publishChannel := make(chan collector.Point, 0)

	publisher := &Publisher{
		connection: connection,
		publishChannel: publishChannel,
	}

	go publisher.start()

	return publisher, nil
}

func (p *Publisher) start() {
	// TODO: Opening connection should probably just go here
	// TODO: https://bbengfort.github.io/programmer/2017/03/03/secure-grpc.html
	defer p.connection.Close()

	md := metadata.New(map[string]string {
		"environment": "foo",
		"app": "bar",
		"service": "qux",
		"docker_name": "0123",
	})

	// TODO: Probably not background
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	for {
		client := api.NewAggregatorClient(p.connection)

		stream, err := client.Publish(ctx)
		if err != nil {
			// TODO: This will loop dumb.
			fmt.Printf("stream open failed: %+v\n", err)
			continue
		}

		fmt.Printf("have stream\n")

		for {
			point := <-p.publishChannel
			fmt.Printf("have point\n")

			req := api.PublishRequest{
				TimeNs: 123,
				CpuUsage: float32(point.CpuUsage),
			}

			err := stream.Send(&req)
			if err != nil {
				fmt.Printf("failed to deliver: %+v", err)
			}
		}
	}
}

func (p *Publisher) openStream() (api.Aggregator_PublishClient, error) {
	client := api.NewAggregatorClient(p.connection)

	stream, err := client.Publish(context.Background())
	if err != nil {
		return nil, err
	}

	req := api.PublishRequest{
		Environment: "foo",
		App: "bar",
		Service: "qux",
		DockerName: "0123",
	}

	err = stream.Send(&req)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (p *Publisher) Publish(point collector.Point) {
	p.publishChannel <- point
}
