package types

// SoftDeleteFilter defines the filtering behavior for soft-deleted entities.
type SoftDeleteFilter int

const (
	// SoftDeleteFilterExcludeDeleted excludes soft-deleted entities (default behavior).
	SoftDeleteFilterExcludeDeleted SoftDeleteFilter = iota
	// SoftDeleteFilterOnlyDeleted returns only soft-deleted entities.
	SoftDeleteFilterOnlyDeleted
	// SoftDeleteFilterAll returns all entities regardless of soft-delete status.
	SoftDeleteFilterAll
)
