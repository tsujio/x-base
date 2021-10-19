package organization

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jinzhu/copier"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/logging"
)

func (controller *OrganizationController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var input schemas.CreateOrganizationInput
	err := schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&schemas.Error{Message: fmt.Sprintf("Invalid request body: %s", err)})
		return
	}

	// Create organization
	o := models.Organization{
		Name: input.Name,
	}
	err = o.Create(controller.DB)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&schemas.Error{Message: fmt.Sprintf("Failed to create organization: %s", err)})
		return
	}

	// Convert to output schema
	var output schemas.Organization
	err = copier.Copy(&output, &o)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&schemas.Error{Message: fmt.Sprintf("Failed to make output data: %s", err)})
		return
	}

	// Send response
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
