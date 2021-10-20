package organization

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/logging"
)

func (controller *OrganizationController) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	// Get organization id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "organizationID", &id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&schemas.Error{Message: fmt.Sprintf("Invalid organization id: %s", err)})
		return
	}

	// Fetch
	organization, err := (&models.Organization{ID: models.UUID(id)}).Get(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&schemas.Error{Message: "Not found"})
			return
		}

		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&schemas.Error{Message: fmt.Sprintf("Failed to get organization: %s", err)})
		return
	}

	// Delete
	err = organization.Delete(controller.DB)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&schemas.Error{Message: fmt.Sprintf("Failed to delete organization: %s", err)})
		return
	}
}