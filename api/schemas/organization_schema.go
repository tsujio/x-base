package schemas

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type GetOrganizationListInput struct {
	PaginationInput
	Properties string `schema:"properties"`
}

type GetOrganizationInput struct {
	Properties string `schema:"properties"`
}

type CreateOrganizationInput struct {
	Properties map[string]interface{} `json:"properties"`
}

type UpdateOrganizationInput struct {
	Properties map[string]interface{} `json:"properties"`
}

type Organization struct {
	ID         uuid.UUID              `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

func (o Organization) MarshalJSON() ([]byte, error) {
	if o.Properties == nil {
		o.Properties = make(map[string]interface{})
	}
	type Alias Organization
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(o)})
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
