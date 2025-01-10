package workers

type DbAssigner struct{}

func (c *DbAssigner) NextWorkerId() int64 {
	panic("Could not assign worker id")
}
