package schemas

import (
	"time"

	"github.com/google/uuid"
)

type GetOrganizationListInput struct {
	PaginationInput
}

type CreateOrganizationInput struct {
	Name string `json:"name" validate:"required"`
}

type UpdateOrganizationInput struct {
	Name *string `json:"name" validate:"omitempty,gt=0"`
}

type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrganizationList struct {
	PaginatedList
	Organizations []Organization `json:"organizations"`
}
