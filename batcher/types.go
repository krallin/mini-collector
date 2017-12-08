package batcher

import (
	"github.com/aptible/mini-collector/batch"
)

type Batcher interface {
	Ingest(entry *batch.Entry)
}
