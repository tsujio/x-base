package schemas

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateColumnInput struct {
	Index *int   `json:"index" validate:"omitempty,gte=0,lte=999"`
	Name  string `json:"name" validate:"required,lte=100"`
	Type  string `json:"type" validate:"required,columntype"`
}

type UpdateColumnInput struct {
	Index *int    `json:"index" validate:"omitempty,gte=0,lte=999"`
	Name  *string `json:"name" validate:"omitempty,gt=0,lte=100"`
	Type  *string `json:"type" validate:"omitempty,columntype"`
}

type ReorderColumnInput struct {
	Order []uuid.UUID `json:"order"`
}

type Column struct {
	ID        uuid.UUID `json:"id"`
	TableID   uuid.UUID `json:"tableId"`
	Index     int       `json:"index"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
