package folder

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

func (controller *FolderController) UpdateFolder(w http.ResponseWriter, r *http.Request) {
	// Get folder id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "folderID", &id)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid folder id", err)
		return
	}

	// Decode request body
	var input schemas.UpdateFolderInput
	err = schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if result := models.ValidateProperties(input.Properties); result != "" {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, result, nil)
		return
	}

	// Fetch
	folder, err := (&models.TableFilesystemEntry{ID: models.UUID(id)}).GetFolder(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			responses.SendErrorResponse(w, r, http.StatusNotFound, "Not found", nil)
			return
		}
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get folder", err)
		return
	}

	// Check destination folder
	if input.ParentFolderID != nil && *input.ParentFolderID != uuid.Nil {
		parent, err := (&models.TableFilesystemEntry{ID: models.UUID(*input.ParentFolderID)}).GetFolder(controller.DB)
		if err != nil {
			if xerrors.Is(err, gorm.ErrRecordNotFound) {
				responses.SendErrorResponse(w, r, http.StatusBadRequest, "Destination folder not found", nil)
				return
			}
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get destination folder", err)
			return
		}

		if parent.OrganizationID != models.UUID(folder.OrganizationID) {
			responses.SendErrorResponse(w, r, http.StatusBadRequest, "Cannot move to another organization", nil)
			return
		}

		// Path loop check
		err = parent.ComputePath(controller.DB)
		if err != nil {
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get destination folder's path", err)
		}
		for _, e := range parent.Path {
			if e.ID == folder.ID {
				responses.SendErrorResponse(w, r, http.StatusBadRequest, "Cannot move to sub folder", nil)
				return
			}
		}
	}

	// Update
	if input.ParentFolderID != nil {
		folder.ParentFolderID = (*models.UUID)(input.ParentFolderID)
	}
	for k, v := range input.Properties {
		folder.Properties[k] = v
	}
	err = folder.Save(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Invalid to save folder", err)
		return
	}

	// Convert to output schema
	var output schemas.Folder
	err = folder.ComputePath(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get path", err)
		return
	}
	err = copier.Copy(&output, &folder)
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
