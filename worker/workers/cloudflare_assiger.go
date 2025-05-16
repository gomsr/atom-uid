package workers

import (
	"github.com/gomsr/atom-cloudflare/kvs/worker"
)

type CloudflareAssigner struct{}

func (c *CloudflareAssigner) NextWorkerId() int64 {
	for i := 0; i < 10; i++ {
		if id, err := worker.NextWorkerID(); err != nil {
			continue
		} else {
			return id
		}
	}
	panic("Could not assign worker id")
}
