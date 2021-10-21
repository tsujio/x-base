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
	"github.com/tsujio/x-base/api/utils"
	"github.com/tsujio/x-base/logging"
)

func (controller *FolderController) GetFolder(w http.ResponseWriter, r *http.Request) {
	// Get folder id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "folderID", &id)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid folder id", err)
		return
	}

	// Fetch
	folder, err := (&models.TableFilesystemEntry{ID: models.UUID(id)}).GetFolder(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(w, r, http.StatusNotFound, "Not found", nil)
			return
		}
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get folder", err)
		return
	}

	// Convert to output schema
	var output schemas.Folder
	err = copier.Copy(&output, &folder)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
		return
	}
	path, err := folder.GetPath(controller.DB)
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
