package publisher

import (
	"fmt"
	"github.com/aptible/mini-collector/api"
	"github.com/aptible/mini-collector/collector"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

// TODO: Make it an inteface
type Publisher struct {
	connectTimeout time.Duration
	publishTimeout time.Duration

	serverAddress string
	tags          map[string]string

	publishChannel chan *api.PublishRequest
	doneChannel    chan interface{}
	cancel         context.CancelFunc
}

func Open(serverAddress string, tags map[string]string, queueSize int) *Publisher {
	ctx, cancel := context.WithCancel(context.Background())

	publisher := &Publisher{
		connectTimeout: 5 * time.Second,
		publishTimeout: 2 * time.Second,

		serverAddress: serverAddress,
		tags:          tags,

		publishChannel: make(chan *api.PublishRequest, queueSize),
		doneChannel:    make(chan interface{}),
		cancel:         cancel,
	}

	go publisher.startPublisher(ctx)

	return publisher
}

func (p *Publisher) startPublisher(ctx context.Context) {
StartLoop:
	for {
		select {
		case <-ctx.Done():
			log.Debugf("shutdown publisher loop")
			break StartLoop
		default:
			p.startConnection(ctx)
		}
	}

	log.Debugf("shutdown publisher")
	p.doneChannel <- nil
}

func (p *Publisher) startConnection(ctx context.Context) {
	connection, err := func() (*grpc.ClientConn, error) {
		dialCtx, cancel := context.WithTimeout(ctx, p.connectTimeout)
		defer cancel()
		return grpc.DialContext(dialCtx, p.serverAddress, grpc.WithInsecure(), grpc.WithBlock())
	}()

	if err != nil {
		log.Errorf("could not connect to [%v]: %v", p.serverAddress, err)
		return
	}
	defer connection.Close()

	client := api.NewAggregatorClient(connection)

	md := metadata.New(p.tags)

	baseCtx := metadata.NewOutgoingContext(ctx, md)

PublishLoop:
	for {
		select {
		case payload := <-p.publishChannel:
			err := func() error {
				localCtx, cancel := context.WithTimeout(baseCtx, p.publishTimeout)
				defer cancel()
				_, err := client.Publish(localCtx, payload, grpc.FailFast(false))
				return err
			}()

			if err != nil {
				// Try to requeue the request. But, if the buffer is
				// full, just drop it (favor more recent data points).
				select {
				case p.publishChannel <- payload:
					log.Infof("requeued point [%v]: %v", (*payload).UnixTime, err)
				default:
					log.Warnf("dropped point [%v]: %v", (*payload).UnixTime, err)
				}

				continue
			}

			log.Debugf("delivered point [%v]", (*payload).UnixTime)
		case <-ctx.Done():
			log.Debugf("shutdown connection loop")
			break PublishLoop
		}
	}

	log.Debugf("shutdown connection")
}

func (p *Publisher) Queue(ts time.Time, point collector.Point) error {
	payload := api.PublishRequest{
		UnixTime:      uint64(ts.Unix()),
		MilliCpuUsage: point.MilliCpuUsage,
		MemoryTotalMb: point.MemoryTotalMb,
		MemoryRssMb:   point.MemoryRssMb,
		MemoryLimitMb: point.MemoryLimitMb,
		Running:       point.Running,
	}

	select {
	case p.publishChannel <- &payload:
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

func (p *Publisher) Close() {
	p.cancel()
	<-p.doneChannel
}
