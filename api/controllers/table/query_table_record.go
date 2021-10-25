package table

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

func (controller *TableController) QueryTableRecord(w http.ResponseWriter, r *http.Request) {
	// Get table id
	vars := mux.Vars(r)
	var tableID uuid.UUID
	err := schemas.DecodeUUID(vars, "tableID", &tableID)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid table id", err)
		return
	}

	// Decode request body
	query, err := schemas.DecodeQueryTableRecordInput(r.Body)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request body", err)
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
	if err := table.FetchColumns(controller.DB); err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to fetch columns", err)
		return
	}

	// Query
	var output interface{}
	switch q := query.(type) {
	case *schemas.InsertQuery:
		// Convert
		iq, err := convertToInsertQuery(q, table)
		if err != nil {
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to convert query", err)
			return
		}

		// Execute
		ids, err := iq.Execute(controller.DB)
		if err != nil {
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to execute query", err)
			return
		}

		// Convert to output schema
		var schema schemas.InsertQueryResult
		err = copier.Copy(&schema.RecordIDs, &ids)
		if err != nil {
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
			return
		}
		output = schema
	case *schemas.SelectQuery:
		// Convert
		sq, err := convertToSelectQuery(q, table)
		if err != nil {
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to convert query", err)
			return
		}

		// Execute
		var result []map[string]interface{}
		err = sq.Execute(controller.DB, &result)
		if err != nil {
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to execute query", err)
			return
		}

		// Convert to output schema
		var schema schemas.SelectQueryResult
		var records [][]interface{}
		for _, row := range result {
			var record []interface{}
			for i := 0; i < len(row); i++ {
				record = append(record, row[fmt.Sprintf("_%d", i)])
			}
			records = append(records, record)
		}
		schema.Records = records
		output = schema
	case *schemas.UpdateQuery:
	case *schemas.DeleteQuery:
	default:
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Invalid query type (application error)", nil)
		return
	}

	// Send response
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func convertToInsertQuery(query *schemas.InsertQuery, table *models.Table) (*models.InsertQuery, error) {
	q := models.InsertQuery{}

	// TableID
	q.TableID = table.ID

	// Columns
	for _, c := range query.Columns {
		var column *models.Column
		for _, col := range table.Columns {
			if col.ID == models.UUID(c.ColumnID) {
				column = &col
				break
			}
		}
		if column == nil {
			return nil, fmt.Errorf("Column not found: id=%s", c.ColumnID)
		}
		q.Columns = append(q.Columns, models.ColumnExpr{
			Column: *column,
		})
	}

	// Values
	for _, row := range query.Values {
		var record []models.ValueExpr
		for _, v := range row {
			record = append(record, models.ValueExpr{
				Value: v.Value,
			})
		}
		if len(record) != len(q.Columns) {
			return nil, fmt.Errorf("Record length != # of columns")
		}
		q.Values = append(q.Values, record)
	}

	return &q, nil
}

func convertToSelectQuery(query *schemas.SelectQuery, table *models.Table) (*models.SelectQuery, error) {
	q := models.SelectQuery{}

	// Columns
	for i, c := range query.Columns {
		col, err := convertToExpr(c, table)
		if err != nil {
			return nil, xerrors.Errorf("Invalid column: %w", err)
		}
		q.Columns = append(q.Columns, models.SelectColumn{
			Column: col,
			As:     fmt.Sprintf("_%d", i),
		})
	}

	// From
	q.From = models.TableExpr{
		Table: *table,
	}

	// OrderBy
	for _, o := range query.OrderBy {
		key, err := convertToExpr(o.Key, table)
		if err != nil {
			return nil, xerrors.Errorf("Invalid sort key: %w", err)
		}

		var order models.SortKeyOrder
		if o.Order == "" || strings.ToLower(o.Order) == "asc" {
			order = models.SortKeyOrderAsc
		} else if strings.ToLower(o.Order) == "desc" {
			order = models.SortKeyOrderDesc
		} else {
			return nil, fmt.Errorf("Invalid sort order: %s", o.Order)
		}

		q.OrderBy = append(q.OrderBy, models.SortKey{
			Key:   key,
			Order: order,
		})
	}

	// Offset
	q.Offset = &query.Offset

	// Limit
	q.Limit = &query.Limit

	return &q, nil
}

func convertToExpr(schema interface{}, table *models.Table) (models.SQLBuilder, error) {
	switch s := schema.(type) {
	case schemas.MetadataExpr:
		return models.MetadataExpr{
			Key: models.MetadataExprKey(s.Metadata),
		}, nil
	case schemas.ColumnExpr:
		for _, col := range table.Columns {
			if col.ID == models.UUID(s.ColumnID) {
				return models.ColumnExpr{
					Column: col,
				}, nil
			}
		}
		return nil, fmt.Errorf("Column not found: id=%s", s.ColumnID)
	case schemas.ValueExpr:
		return models.ValueExpr{
			Value: s.Value,
		}, nil
	case schemas.FuncExpr:
		var expr models.FuncExpr
		switch s.Func {
		case "count":
			expr.Func = models.FuncExprFuncCount
		default:
			return nil, fmt.Errorf("Invalid func: %s", s.Func)
		}

		for _, arg := range s.Args {
			e, err := convertToExpr(arg, table)
			if err != nil {
				return nil, xerrors.Errorf("Failed to convert expr: %w", err)
			}
			expr.Args = append(expr.Args, e)
		}

		return expr, nil
	default:
		return nil, fmt.Errorf("Invalid expr type: %T", s)
	}
}
