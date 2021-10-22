package table

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

func (controller *TableController) UpdateTable(w http.ResponseWriter, r *http.Request) {
	// Get table id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "tableID", &id)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid table id", err)
		return
	}

	// Decode request body
	var input schemas.UpdateTableInput
	err = schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Fetch
	table, err := (&models.TableFilesystemEntry{ID: models.UUID(id)}).GetTable(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			responses.SendErrorResponse(w, r, http.StatusNotFound, "Not found", nil)
			return
		}
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get table", err)
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

		if parent.OrganizationID != models.UUID(table.OrganizationID) {
			responses.SendErrorResponse(w, r, http.StatusBadRequest, "Cannot move to another organization", nil)
			return
		}
	}

	// Update
	if input.Name != nil {
		table.Name = *input.Name
	}
	if input.ParentFolderID != nil {
		table.ParentFolderID = (*models.UUID)(input.ParentFolderID)
	}
	err = table.Save(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Invalid to save table", err)
		return
	}

	// Convert to output schema
	var output schemas.Table
	err = table.ComputePath(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get path", err)
		return
	}
	err = copier.Copy(&output, &table)
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
