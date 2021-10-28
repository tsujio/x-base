package schemas

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateFolderInput struct {
	OrganizationID uuid.UUID              `json:"organizationId" validate:"required"`
	ParentFolderID *uuid.UUID             `json:"parentFolderId"`
	Properties     map[string]interface{} `json:"properties"`
}

type UpdateFolderInput struct {
	ParentFolderID *uuid.UUID             `json:"parentFolderId"`
	Properties     map[string]interface{} `json:"properties"`
}

type GetFolderChildrenInput struct {
	PaginationInput
	OrganizationID uuid.UUID `schema:"organizationId"`
	Sort           string    `schema:"sort"`
}

type Folder struct {
	TableFilesystemEntry
}

func (f Folder) MarshalJSON() ([]byte, error) {
	if f.Path == nil {
		f.Path = []TableFilesystemPathEntry{}
	}
	if f.Properties == nil {
		f.Properties = make(map[string]interface{})
	}
	type Alias Folder
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(f)})
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
	if c.Properties == nil {
		c.Properties = make(map[string]interface{})
	}
	type Alias FolderChild
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(c)})
}

type TableFilesystemEntry struct {
	ID             uuid.UUID                  `json:"id"`
	OrganizationID uuid.UUID                  `json:"organizationId"`
	Type           string                     `json:"type"`
	Path           []TableFilesystemPathEntry `json:"path"`
	Properties     map[string]interface{}     `json:"properties"`
	CreatedAt      time.Time                  `json:"createdAt"`
	UpdatedAt      time.Time                  `json:"updatedAt"`
}

type TableFilesystemPathEntry struct {
	ID         uuid.UUID              `json:"id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

func (e TableFilesystemPathEntry) MarshalJSON() ([]byte, error) {
	if e.Properties == nil {
		e.Properties = make(map[string]interface{})
	}
	type Alias TableFilesystemPathEntry
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(e)})
}
