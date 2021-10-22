package schemas

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/tsujio/x-base/api/models"
)

func init() {
	validator.New().RegisterValidation("columntype", func(fl validator.FieldLevel) bool {
		return models.IsValidColumnType(fl.Field().String())
	})
}

type CreateColumnInput struct {
	Index *int   `json:"index" validate:"omitempty,gte=0"`
	Name  string `json:"name" validate:"required"`
	Type  string `json:"type" validate:"required,columntype"`
}

type UpdateColumnInput struct {
	Index *int    `json:"index" validate:"omitempty,gte=0"`
	Name  *string `json:"name" validate:"omitempty,gt=0"`
	Type  *string `json:"type" validate:"omitempty,columntype"`
}

type ReorderColumnInput struct {
	Order []uuid.UUID `json:"order"`
}

type Column struct {
	ID        uuid.UUID `json:"id"`
	TableID   uuid.UUID `json:"table_id"`
	Index     int       `json:"index"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
