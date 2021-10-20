package table

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils"
)

func (controller *TableController) DeleteTable(w http.ResponseWriter, r *http.Request) {
	// Get table id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "tableID", &id)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid table id", err)
		return
	}

	// Fetch
	table, err := (&models.TableFilesystemEntry{ID: models.UUID(id)}).GetTable(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(w, r, http.StatusNotFound, "Not found", nil)
			return
		}
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get table", err)
		return
	}

	// Delete
	err = table.Delete(controller.DB)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to delete table", err)
		return
	}
}
