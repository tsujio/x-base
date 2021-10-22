package schemas

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type GetOrganizationListInput struct {
	PaginationInput
}

type CreateOrganizationInput struct {
	Name string `json:"name" validate:"required,lte=100"`
}

type UpdateOrganizationInput struct {
	Name *string `json:"name" validate:"omitempty,gt=0,lte=100"`
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

func (o OrganizationList) MarshalJSON() ([]byte, error) {
	if o.Organizations == nil {
		o.Organizations = []Organization{}
	}
	type Alias OrganizationList
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(o)})
}
