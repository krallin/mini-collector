package emitter

import (
	"github.com/aptible/mini-collector/batch"
)

type Emitter interface {
	Emit(batch []batch.Entry) error
}
