package folder

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils/responses"
	"github.com/tsujio/x-base/logging"
)

func (controller *FolderController) CreateFolder(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var input schemas.CreateFolderInput
	err := schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Check parent folder
	if input.ParentFolderID != nil && *input.ParentFolderID != uuid.Nil {
		parent, err := (&models.TableFilesystemEntry{ID: models.UUID(*input.ParentFolderID)}).GetFolder(controller.DB)
		if err != nil {
			if xerrors.Is(err, gorm.ErrRecordNotFound) {
				responses.SendErrorResponse(w, r, http.StatusBadRequest, "Parent folder not found", nil)
				return
			}
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get parent folder", err)
			return
		}

		if parent.OrganizationID != models.UUID(input.OrganizationID) {
			responses.SendErrorResponse(w, r, http.StatusBadRequest, "Cannot create folder as a child of another organization's folder", nil)
			return
		}
	}

	// Create folder
	f := models.Folder{
		TableFilesystemEntry: models.TableFilesystemEntry{
			OrganizationID: models.UUID(input.OrganizationID),
			Name:           input.Name,
			ParentFolderID: (*models.UUID)(input.ParentFolderID),
		},
	}
	err = f.Create(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to create folder", err)
		return
	}

	// Convert to output schema
	var output schemas.Folder
	err = f.ComputePath(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get path", err)
		return
	}
	err = copier.Copy(&output, &f)
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
