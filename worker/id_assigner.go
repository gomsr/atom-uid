package worker

// IdAssigner defines an interface for assigning worker IDs.
type IdAssigner interface {

	// AssignWorkerId assigns a worker ID for the DefaultUidGenerator.
	// Returns the assigned worker ID.
	AssignWorkerId() int64
}
