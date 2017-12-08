package main

import (
	"github.com/aptible/mini-collector/collector"
	"github.com/aptible/mini-collector/publisher"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	publisherBufferSize = 10
	// pollInterval        = 2 * time.Second // TODO
	pollInterval = 2000 * time.Millisecond // TODO
)

func getEnvOrFatal(k string) string {
	val, ok := os.LookupEnv(k)
	if !ok {
		log.Fatalf("%s must be set", k)
	}
	return val
}

func main() {
	// TODO: Volumes / configuration
	// TODO: Throttling stats
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	serverAddress := getEnvOrFatal("MINI_COLLECTOR_REMOTE_ADDRESS")
	containerId := getEnvOrFatal("MINI_COLLECTOR_CONTAINER_ID")
	environmentName := getEnvOrFatal("MINI_COLLECTOR_ENVIRONMENT_NAME")
	serviceName := getEnvOrFatal("MINI_COLLECTOR_SERVICE_NAME")

	tags := map[string]string{
		"environment": environmentName,
		"service":     serviceName,
		"container":   containerId,
	}

	appName, ok := os.LookupEnv("MINI_COLLECTOR_APP_NAME")
	if ok {
		tags["app"] = appName
	}

	databaseName, ok := os.LookupEnv("MINI_COLLECTOR_DATABASE_NAME")
	if ok {
		tags["database"] = databaseName
	}

	_, debug := os.LookupEnv("MINI_COLLECTOR_DEBUG")
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	publisher := publisher.Open(
		serverAddress,
		tags,
		20,
	)

	c := collector.NewCollector(containerId)

	lastState := collector.MakeNoContainerState()

MainLoop:
	for {
		select {
		case <-time.After(time.Until(lastState.Time.Add(pollInterval))):
			var point collector.Point
			point, lastState = c.GetPoint(lastState)
			err := publisher.Queue(lastState.Time, point)
			if err != nil {
				log.Warnf("publisher is failling behind: %v", err)
			}
		case <-termChan:
			// Exit
			log.Infof("exiting")
			break MainLoop
		}
	}

	publisher.Close()
}
