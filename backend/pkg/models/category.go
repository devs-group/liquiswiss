package models

type Category struct {
	ID      int64  `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	CanEdit bool   `db:"can_edit" json:"canEdit"`
	// InUse marks categories still referenced by transactions of the user's
	// current organisation (delete requires reassigning them first)
	InUse bool `db:"in_use" json:"inUse"`
}

type CreateCategory struct {
	Name string `json:"name" validate:"required,max=100"`
}

type UpdateCategory struct {
	Name *string `json:"name" validate:"omitempty,max=100"`
}

type ReassignCategory struct {
	TargetID int64 `json:"targetId" validate:"required,gt=0"`
}
