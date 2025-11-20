package types

// SoftDeleteFilter defines the filtering behavior for soft-deleted entities.
type SoftDeleteFilter string

const (
	// SoftDeleteFilterExcludeDeleted excludes soft-deleted entities (default behavior).
	SoftDeleteFilterExcludeDeleted SoftDeleteFilter = "exclude_deleted"
	// SoftDeleteFilterOnlyDeleted returns only soft-deleted entities.
	SoftDeleteFilterOnlyDeleted SoftDeleteFilter = "only_deleted"
	// SoftDeleteFilterAll returns all entities regardless of soft-delete status.
	SoftDeleteFilterAll SoftDeleteFilter = "all"
)
