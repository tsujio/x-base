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

	// Update
	if input.Name != nil {
		column.Name = *input.Name
	}
	if input.Type != nil {
		column.Type = *input.Type
	}
	if input.Index != nil {
		column.Index = *input.Index
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
