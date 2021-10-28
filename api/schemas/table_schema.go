package schemas

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CreateTableInput struct {
	OrganizationID uuid.UUID              `json:"organizationId" validate:"required"`
	ParentFolderID *uuid.UUID             `json:"parentFolderId"`
	Columns        []CreateColumnInput    `json:"columns"`
	Properties     map[string]interface{} `json:"properties"`
}

type UpdateTableInput struct {
	ParentFolderID *uuid.UUID             `json:"parentFolderId"`
	Properties     map[string]interface{} `json:"properties"`
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
	if t.Properties == nil {
		t.Properties = make(map[string]interface{})
	}
	type Alias Table
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(t)})
}
