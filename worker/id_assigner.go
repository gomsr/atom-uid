package worker

import "github.com/gomsr/atom-uid/worker/workers"

type Type uint

const (
	LocalWorkerId Type = iota
	DbWorkerId
	CloudflareWorkerId
)

// IdAssigner defines an interface for assigning worker IDs.
type IdAssigner interface {

	// NextWorkerId assigns a worker ID for the DefaultUidGenerator.
	// Returns the assigned worker ID.
	NextWorkerId() int64
}

func (c Type) Instance() IdAssigner {
	var assigner IdAssigner
	switch c {
	case DbWorkerId:
		assigner = &workers.DbAssigner{}
	case CloudflareWorkerId:
		assigner = &workers.CloudflareAssigner{}
	default:
		assigner = &workers.LocalAssigner{}
	}

	return assigner
}
