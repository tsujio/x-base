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

func (controller *TableController) UpdateColumn(w http.ResponseWriter, r *http.Request) {
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

	// Decode request body
	var input schemas.UpdateColumnInput
	err = schemas.DecodeJSON(r.Body, &input)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if result := models.ValidateProperties(input.Properties); result != "" {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, result, nil)
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

	// Update
	if input.Index != nil {
		column.Index = *input.Index
	}
	for k, v := range input.Properties {
		column.Properties[k] = v
	}
	err = controller.DB.Transaction(func(tx *gorm.DB) error {
		return column.Save(tx, false)
	})
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to update column", err)
		return
	}

	// Convert to output schema
	var output schemas.Column
	err = copier.Copy(&output, &column)
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
