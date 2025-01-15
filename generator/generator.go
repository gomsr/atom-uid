package generator

type Type uint

const (
	DefaultUid Type = iota
	CachedUid
)

const (
	EpochStr       = "2024-01-01"
	EpochStrFormat = "2006-01-02"
)

// UidGenerator defines an interface for generating and parsing unique IDs.
type UidGenerator interface {
	// MustUID generates a unique ID.
	// Returns the generated UID or an error if the generation fails.
	MustUID() int64

	GetUID() (int64, error)

	// ParseUID parses the given UID into its components (e.g., timestamp, worker ID, sequence).
	// Returns the parsed information as a string.
	ParseUID(uid int64) string
}
