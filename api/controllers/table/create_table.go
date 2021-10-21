package table

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
	"github.com/tsujio/x-base/api/utils"
	"github.com/tsujio/x-base/logging"
)

func (controller *TableController) CreateTable(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var input schemas.CreateTableInput
	err := schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Check parent folder
	if input.ParentFolderID != nil && *input.ParentFolderID != uuid.Nil {
		parent, err := (&models.TableFilesystemEntry{ID: models.UUID(*input.ParentFolderID)}).GetFolder(controller.DB)
		if err != nil {
			if xerrors.Is(err, gorm.ErrRecordNotFound) {
				utils.SendErrorResponse(w, r, http.StatusBadRequest, "Parent folder not found", nil)
				return
			}
			utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get parent folder", err)
			return
		}

		if parent.OrganizationID != models.UUID(input.OrganizationID) {
			utils.SendErrorResponse(w, r, http.StatusBadRequest, "Cannot create table as a child of another organization's folder", nil)
			return
		}
	}

	// Create table
	t := models.Table{
		TableFilesystemEntry: models.TableFilesystemEntry{
			OrganizationID: models.UUID(input.OrganizationID),
			Name:           input.Name,
			ParentFolderID: (*models.UUID)(input.ParentFolderID),
		},
	}
	err = t.Create(controller.DB)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to create table", err)
		return
	}

	// Convert to output schema
	var output schemas.Table
	err = t.ComputePath(controller.DB)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get path", err)
		return
	}
	err = copier.Copy(&output, &t)
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
