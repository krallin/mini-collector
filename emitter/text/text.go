package text

import (
	"fmt"
	"github.com/aptible/mini-collector/batch"
)

type textEmitter struct{}

func New() *textEmitter {
	return &textEmitter{}
}

func (t *textEmitter) Emit(batch []batch.Entry) error {
	for _, entry := range batch {
		fmt.Printf("%+v\n", entry)
	}

	return nil
}
