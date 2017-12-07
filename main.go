package main

import (
	"sync/atomic"
	"fmt"
	"time"
	"github.com/aptible/mini-collector/collector"
	"github.com/aptible/mini-collector/publisher"
)

func main() {
	// TODO: Need to push to remote
	// TODO: Need to have an ID on the stats to never push the same again (maybe just a timestamp - we need that anyway)
	// TODO: Volumes / configuration
	// TODO: Throttling stats
	// TODO: Handle sigterm / sigint

	var value atomic.Value
	readyChan := make(chan interface{}, 1)

	go func() {
		var point collector.Point
		lastState := collector.MakeNoContainerState()

		c := collector.NewCollector("1f58a43e2863fd73aebdf09a7dae6c47983af8fd7523a048e4b9bddcd4ee6f2f")

		for {
			point, lastState = c.GetPoint(lastState)
			value.Store(point)

			select {
			case readyChan <- nil:
				fmt.Println("pusher ok")
			default:
				// TODO: Better logging
				fmt.Println("pusher falling behind")
			}

			time.Sleep(1000 * time.Millisecond)
		}
	}()

	// TODO: better error handling here / needs to be in a retry loop
	publisher, err := publisher.Open()
	if err != nil {
		fmt.Printf("failed to create publisher: %+v", err)
		return
	}

	// defer publisher.Close() // TODO

	for {
		<-readyChan
		fmt.Printf("wake up\n")
		point := value.Load().(collector.Point)
		publisher.Publish(point)
	}
}
