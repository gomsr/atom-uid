package workers

import "math/rand"

type LocalAssigner struct{}

func (c *LocalAssigner) NextWorkerId() int64 {
	return rand.Int63n(512)
}
