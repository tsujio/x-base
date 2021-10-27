package organization

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jinzhu/copier"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils/responses"
	"github.com/tsujio/x-base/logging"
)

func (controller *OrganizationController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var input schemas.CreateOrganizationInput
	err := schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create organization
	o := models.Organization{}
	err = o.Create(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to create organization", err)
		return
	}

	// Convert to output schema
	var output schemas.Organization
	err = copier.Copy(&output, &o)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
		return
	}

	// Send response
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
