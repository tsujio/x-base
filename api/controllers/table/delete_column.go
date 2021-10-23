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

	// Fetch table
	table, err := (&models.TableFilesystemEntry{ID: models.UUID(tableID)}).GetTable(controller.DB)
	if err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			responses.SendErrorResponse(w, r, http.StatusNotFound, "Table not found", nil)
			return
		}
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get table", err)
		return
	}

	// Fetch columns
	err = table.FetchColumns(controller.DB)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get columns", err)
	}

	// Find column
	var column *models.Column
	for _, c := range table.Columns {
		if c.ID == models.UUID(columnID) {
			column = &c
			break
		}
	}
	if column == nil {
		responses.SendErrorResponse(w, r, http.StatusNotFound, "Column not found", nil)
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
