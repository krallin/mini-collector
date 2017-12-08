package batch

import (
	"github.com/aptible/mini-collector/api"
	"time"
)

type Entry struct {
	Time time.Time
	Tags map[string]string
	api.PublishRequest
}
