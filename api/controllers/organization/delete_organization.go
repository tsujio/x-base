package organization

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils/responses"
)

func (controller *OrganizationController) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	// Get organization id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "organizationID", &id)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid organization id", err)
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

	// Delete
	err = organization.Delete(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to delete organization", err)
		return
	}
}
