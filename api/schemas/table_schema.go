package schemas

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CreateTableInput struct {
	OrganizationID uuid.UUID           `json:"organization_id" validate:"required"`
	Name           string              `json:"name" validate:"required,lte=100"`
	ParentFolderID *uuid.UUID          `json:"parent_folder_id"`
	Columns        []CreateColumnInput `json:"columns"`
}

type UpdateTableInput struct {
	Name           *string    `json:"name" validate:"omitempty,gt=0,lte=100"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
}

type Table struct {
	TableFilesystemEntry
	Columns []Column `json:"columns"`
}

func (t Table) MarshalJSON() ([]byte, error) {
	if t.Columns == nil {
		t.Columns = []Column{}
	}
	if t.Path == nil {
		t.Path = []TableFilesystemPathEntry{}
	}
	type Alias Table
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(t)})
}
