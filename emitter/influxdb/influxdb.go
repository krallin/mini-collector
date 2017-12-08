package influxdb

import (
	"encoding/json"
	"fmt"
	"github.com/aptible/mini-collector/batch"
	client "github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
)

const (
	bufferSize   = 10
	maxSendCount = 5
)

type influxDbConfiguration struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type batchForResend struct {
	batch     []batch.Entry
	sendCount int
}

type influxdbEmitter struct {
	client   client.Client
	database string

	sendBuffer   chan []batch.Entry
	resendBuffer chan batchForResend
}

func New(jsonConfiguration string) (*influxdbEmitter, error) {
	config := influxDbConfiguration{}
	err := json.Unmarshal([]byte(jsonConfiguration), &config)

	if err != nil {
		return nil, fmt.Errorf("could not decode InfluxDB jsonConfiguration: %v", err)
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Password,
	})

	if err != nil {
		return nil, fmt.Errorf("invalid InfluxDb configuration: %v", err)
	}

	e := &influxdbEmitter{
		client:       c,
		database:     config.Database,
		sendBuffer:   make(chan []batch.Entry, bufferSize),
		resendBuffer: make(chan batchForResend, bufferSize),
	}

	go e.run()

	return e, nil
}

func (e *influxdbEmitter) Emit(batch []batch.Entry) error {
	if len(batch) <= 0 {
		// Nothing to emit
		return nil
	}

	select {
	case e.sendBuffer <- batch:
		return nil
	default:
		return fmt.Errorf("buffer is full")
	}
}

func (e *influxdbEmitter) run() {
	for {
		select {
		case batch := <-e.sendBuffer:
			log.Infof("send new batch to InfluxDB: %d items", len(batch))
			err := e.sendBatch(batch)
			if err != nil {
				e.resendBatch(batchForResend{batch: batch})
			}
		case batchForResend := <-e.resendBuffer:
			log.Infof("send failed batch to InfluxDB: %d items", len(batchForResend.batch))
			err := e.sendBatch(batchForResend.batch)
			if err != nil {
				e.resendBatch(batchForResend)
			}
		}
	}
}

func (e *influxdbEmitter) sendBatch(batch []batch.Entry) error {
	bp := buildBatchPoints("drain-dev", batch)
	err := e.client.Write(bp)

	if err != nil {
		return fmt.Errorf("failed to send batch: %+v", err)
	}

	return nil
}

func (e *influxdbEmitter) resendBatch(p batchForResend) {
	p.sendCount++

	if p.sendCount > maxSendCount {
		log.Errorf("batch lost: retry exceeded")
		return
	}

	select {
	case e.resendBuffer <- p:
		log.Infof("scheduled batch for resend")
	default:
		log.Errorf("batch lost: resend buffer full")
	}
}
