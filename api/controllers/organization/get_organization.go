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
	"github.com/tsujio/x-base/api/utils"
	"github.com/tsujio/x-base/logging"
)

func (controller *OrganizationController) GetOrganization(w http.ResponseWriter, r *http.Request) {
	// Get organization id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "organizationID", &id)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid organization id", err)
		return
	}

	// Fetch
	organization, err := (&models.Organization{ID: models.UUID(id)}).Get(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(w, r, http.StatusNotFound, "Not found", nil)
			return
		}
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get organization", err)
		return
	}

	// Convert to output schema
	var output schemas.Organization
	err = copier.Copy(&output, &organization)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
		return
	}

	// Send response
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
