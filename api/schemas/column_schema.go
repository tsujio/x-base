package schemas

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateColumnInput struct {
	Index      *int                   `json:"index" validate:"omitempty,gte=0,lte=999"`
	Properties map[string]interface{} `json:"properties"`
}

type UpdateColumnInput struct {
	Index      *int                   `json:"index" validate:"omitempty,gte=0,lte=999"`
	Properties map[string]interface{} `json:"properties"`
}

type ReorderColumnInput struct {
	Order []uuid.UUID `json:"order"`
}

type Column struct {
	ID         uuid.UUID              `json:"id"`
	TableID    uuid.UUID              `json:"tableId"`
	Index      int                    `json:"index"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

func (c Column) MarshalJSON() ([]byte, error) {
	if c.Properties == nil {
		c.Properties = make(map[string]interface{})
	}
	type Alias Column
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(c)})
}

type ColumnList struct {
	Columns []Column `json:"columns"`
}

func (c ColumnList) MarshalJSON() ([]byte, error) {
	if c.Columns == nil {
		c.Columns = []Column{}
	}
	type Alias ColumnList
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(c)})
}
