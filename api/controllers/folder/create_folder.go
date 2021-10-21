package folder

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"

	"github.com/tsujio/x-base/api/middlewares"
	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils"
	"github.com/tsujio/x-base/logging"
)

func (controller *FolderController) CreateFolder(w http.ResponseWriter, r *http.Request) {
	// Get organization id
	organizationID := middlewares.GetOrganizationID(r)
	if organizationID == uuid.Nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Organization id not specified", nil)
		return
	}

	// Decode request body
	var input schemas.CreateFolderInput
	err := schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create folder
	f := models.Folder{
		TableFilesystemEntry: models.TableFilesystemEntry{
			OrganizationID: models.UUID(organizationID),
			Name:           input.Name,
			ParentFolderID: (*models.UUID)(input.ParentFolderID),
		},
	}
	err = f.Create(controller.DB)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to create folder", err)
		return
	}

	// Convert to output schema
	var output schemas.Folder
	err = copier.Copy(&output, &f)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
		return
	}
	path, err := f.GetPath(controller.DB)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get path", err)
		return
	}
	err = copier.Copy(&output.Path, &path)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data (path)", err)
		return
	}

	// Send response
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
