package table

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

func (controller *TableController) DeleteColumn(w http.ResponseWriter, r *http.Request) {
	// Get table id and column id
	vars := mux.Vars(r)
	var tableID, columnID uuid.UUID
	err := schemas.DecodeUUID(vars, "tableID", &tableID)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid table id", err)
		return
	}
	err = schemas.DecodeUUID(vars, "columnID", &columnID)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid column id", err)
		return
	}

	// Fetch column
	column, err := (&models.Column{ID: models.UUID(columnID)}).Get(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			responses.SendErrorResponse(w, r, http.StatusNotFound, "Not found", nil)
			return
		}
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get column", err)
		return
	}
	if column.TableID != models.UUID(tableID) {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Wrong table id", nil)
		return
	}

	// Delete
	err = controller.DB.Transaction(func(tx *gorm.DB) error {
		return column.Delete(tx, false)
	})
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to delete column", err)
		return
	}
}
