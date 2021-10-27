package schemas

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateFolderInput struct {
	OrganizationID uuid.UUID  `json:"organizationId" validate:"required"`
	Name           string     `json:"name" validate:"required,lte=100"`
	ParentFolderID *uuid.UUID `json:"parentFolderId"`
}

type UpdateFolderInput struct {
	Name           *string    `json:"name" validate:"omitempty,gt=0,lte=100"`
	ParentFolderID *uuid.UUID `json:"parentFolderId"`
}

type GetFolderChildrenInput struct {
	PaginationInput
	OrganizationID uuid.UUID `schema:"organizationId"`
}

type Folder struct {
	TableFilesystemEntry
}

type FolderChildren struct {
	PaginatedList
	Children []FolderChild `json:"children"`
}

func (c FolderChildren) MarshalJSON() ([]byte, error) {
	if c.Children == nil {
		c.Children = []FolderChild{}
	}
	type Alias FolderChildren
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(c)})
}

type FolderChild TableFilesystemEntry

func (c FolderChild) MarshalJSON() ([]byte, error) {
	if c.Path == nil {
		c.Path = []TableFilesystemPathEntry{}
	}
	type Alias FolderChild
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(c)})
}

type TableFilesystemEntry struct {
	ID             uuid.UUID                  `json:"id"`
	OrganizationID uuid.UUID                  `json:"organizationId"`
	Type           string                     `json:"type"`
	Name           string                     `json:"name"`
	Path           []TableFilesystemPathEntry `json:"path"`
	CreatedAt      time.Time                  `json:"createdAt"`
	UpdatedAt      time.Time                  `json:"updatedAt"`
}

type TableFilesystemPathEntry struct {
	ID   uuid.UUID `json:"id"`
	Type string    `json:"type"`
	Name string    `json:"name"`
}
