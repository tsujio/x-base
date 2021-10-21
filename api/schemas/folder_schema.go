package schemas

import (
	"time"

	"github.com/google/uuid"
)

type CreateFolderInput struct {
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
	Name           string     `json:"name" validate:"required,lte=100"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
}

type UpdateFolderInput struct {
	Name           *string    `json:"name" validate:"omitempty,gt=0,lte=100"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
}

type GetFolderChildrenInput struct {
	PaginationInput
	OrganizationID uuid.UUID `schema:"organizationId"`
}

type Folder struct {
	ID             uuid.UUID                  `json:"id"`
	OrganizationID uuid.UUID                  `json:"organization_id"`
	Type           string                     `json:"type"`
	Name           string                     `json:"name"`
	Path           []TableFilesystemPathEntry `json:"path"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

type FolderChildren struct {
	PaginatedList
	Children []interface{} `json:"children"`
}
