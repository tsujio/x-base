package folder

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

func (controller *FolderController) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	// Get folder id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "folderID", &id)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid folder id", err)
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

	// Delete
	err = folder.Delete(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to delete folder", err)
		return
	}
}
