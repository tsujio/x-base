package schemas

import (
	"time"

	"github.com/google/uuid"
)

type CreateTableInput struct {
	Name           string     `json:"name" validate:"required"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
}

type UpdateTableInput struct {
	Name           *string    `json:"name" validate:"omitempty,gt=0"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
}

type Table struct {
	ID             uuid.UUID                  `json:"id"`
	OrganizationID uuid.UUID                  `json:"organization_id"`
	Type           string                     `json:"type"`
	Name           string                     `json:"name"`
	Path           []TableFilesystemPathEntry `json:"path"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

type TableFilesystemPathEntry struct {
	ID   uuid.UUID `json:"id"`
	Type string    `json:"type"`
	Name string    `json:"name"`
}
