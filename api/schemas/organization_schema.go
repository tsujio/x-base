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
}

type UpdateOrganizationInput struct {
}

type Organization struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
