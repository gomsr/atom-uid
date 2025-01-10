package generator

// UidGenerator defines an interface for generating and parsing unique IDs.
type UidGenerator interface {
	// GetUID generates a unique ID.
	// Returns the generated UID or an error if the generation fails.
	GetUID() (int64, error)

	// ParseUID parses the given UID into its components (e.g., timestamp, worker ID, sequence).
	// Returns the parsed information as a string.
	ParseUID(uid int64) string
}
