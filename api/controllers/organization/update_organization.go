package organization

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils/responses"
	"github.com/tsujio/x-base/logging"
)

func (controller *OrganizationController) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	// Get organization id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "organizationID", &id)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid organization id", err)
		return
	}

	// Decode request body
	var input schemas.UpdateOrganizationInput
	err = schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Fetch
	organization, err := (&models.Organization{ID: models.UUID(id)}).Get(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			responses.SendErrorResponse(w, r, http.StatusNotFound, "Not found", nil)
			return
		}
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get organization", err)
		return
	}

	// Update
	if result := models.ValidateProperties(input.Properties); result != "" {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, result, nil)
		return
	}
	for k, v := range input.Properties {
		organization.Properties[k] = v
	}
	err = organization.Save(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to save organization", err)
		return
	}

	// Convert to output schema
	var output schemas.Organization
	err = copier.Copy(&output, &organization)
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
