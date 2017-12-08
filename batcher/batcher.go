package batcher

import (
	"github.com/aptible/mini-collector/batch"
	"github.com/aptible/mini-collector/emitter"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	bufferSize       = 10
	batchSize        = 100
	publishFrequency = 5 * time.Second // TODO: Tweak
)

type batcher struct {
	emitter emitter.Emitter
	buffer  chan *batch.Entry
}

func New(emitter emitter.Emitter) Batcher {
	b := &batcher{
		emitter: emitter,
		buffer:  make(chan *batch.Entry, bufferSize),
	}

	go b.start()

	return b
}

func (b *batcher) start() {
	for {
		lastPublish := time.Now()
		currentBatch := make([]batch.Entry, 0, batchSize)

	BatchLoop:
		for {
			select {
			case newEntry := <-b.buffer:
				currentBatch = append(currentBatch, *newEntry)
				if len(currentBatch) >= batchSize {
					break BatchLoop
				}
			case <-time.After(time.Until(lastPublish.Add(publishFrequency))):
				break BatchLoop
			}
		}

		b.emitBatch(currentBatch)
	}
}

func (b *batcher) emitBatch(batch []batch.Entry) {
	err := b.emitter.Emit(batch)
	if err != nil {
		log.Errorf("emitter rejected batch: %+v", err)
	}
}

func (b *batcher) Ingest(entry *batch.Entry) {
	b.buffer <- entry
}
