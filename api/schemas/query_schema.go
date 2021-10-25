package schemas

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
)

func DecodeQueryTableRecordInput(sourcce io.Reader) (interface{}, error) {
	var input map[string]interface{}
	if err := json.NewDecoder(sourcce).Decode(&input); err != nil {
		return nil, xerrors.Errorf("Failed to decode input as json: %w", err)
	}

	if i, exists := input["insert"]; exists {
		return DecodeInsertQuery(i, "insert")
	} else if s, exists := input["select"]; exists {
		return DecodeSelectQuery(s, "select")
	} else if u, exists := input["update"]; exists {
		return DecodeUpdateQuery(u, "update")
	} else if d, exists := input["delete"]; exists {
		return DecodeDeleteQuery(d, "delete")
	} else {
		return nil, fmt.Errorf("Invalid query (expect \"insert\", \"select\", \"update\", or \"delete\")")
	}
}

type InsertQuery struct {
	Columns []ColumnExpr
	Values  [][]ValueExpr
}

type SelectQuery struct {
	Columns []interface{}
	OrderBy []SortKey
	Offset  int
	Limit   int
}

type UpdateQuery struct {
}

type DeleteQuery struct {
}

type MetadataExpr struct {
	Metadata string
}

type ColumnExpr struct {
	ColumnID uuid.UUID
}

type ValueExpr struct {
	Value interface{}
}

type FuncExpr struct {
	Func string
	Args []interface{}
}

type SortKey struct {
	Key   interface{}
	Order string
}

func DecodeInsertQuery(input interface{}, path string) (*InsertQuery, error) {
	in, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=object, got=%T, path=%s", input, path)
	}

	var query InsertQuery

	// columns
	columns, exists := in["columns"]
	if !exists {
		return nil, fmt.Errorf(".columns required: path=%s", path)
	}
	cols, ok := columns.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=array, got=%T, path=%s.columns", columns, path)
	}
	for i, c := range cols {
		expr, err := DecodeColumnExpr(c, fmt.Sprintf("%s.columns[%d]", path, i))
		if err != nil {
			return nil, err
		}
		query.Columns = append(query.Columns, *expr)
	}

	// values
	values, exists := in["values"]
	if !exists {
		return nil, fmt.Errorf(".values required: path=%s", path)
	}
	vals, ok := values.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=array, got=%T, path=%s.values", values, path)
	}
	for i, val := range vals {
		var record []ValueExpr
		vi, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid type: expected=array, got=%T, path=%s.values[%d]", val, path, i)
		}
		for j, vij := range vi {
			expr, err := DecodeValueExpr(vij, fmt.Sprintf("%s.values[%d][%d]", path, i, j))
			if err != nil {
				return nil, err
			}
			record = append(record, *expr)
		}
		query.Values = append(query.Values, record)
	}

	return &query, nil
}

func DecodeSelectQuery(input interface{}, path string) (*SelectQuery, error) {
	in, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=object, got=%T, path=%s", input, path)
	}

	var query SelectQuery

	// columns
	columns, exists := in["columns"]
	if !exists {
		return nil, fmt.Errorf(".columns required: path=%s", path)
	}
	cols, ok := columns.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=array, got=%T, path=%s.columns", columns, path)
	}
	for i, c := range cols {
		expr, err := DecodeExpr(c, fmt.Sprintf("%s.columns[%d]", path, i))
		if err != nil {
			return nil, err
		}
		query.Columns = append(query.Columns, reflect.ValueOf(expr).Elem().Interface())
	}

	// order_by
	orderBy, exists := in["order_by"]
	if exists {
		ob, ok := orderBy.([]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid type: expected=array, got=%T, path=%s.order_by", orderBy, path)
		}
		for i, o := range ob {
			k, err := DecodeSortKey(o, fmt.Sprintf("%s.order_by[%d]", path, i))
			if err != nil {
				return nil, err
			}
			query.OrderBy = append(query.OrderBy, *k)
		}
	} else {
		query.OrderBy = []SortKey{
			{
				Key:   MetadataExpr{Metadata: "created_at"},
				Order: "asc",
			},
			{
				Key:   MetadataExpr{Metadata: "id"},
				Order: "asc",
			},
		}
	}

	// offset
	offset, exists := in["offset"]
	if exists {
		o, ok := offset.(float64)
		if !ok {
			return nil, fmt.Errorf("Invalid type: expected=number, got=%T, path=%s.offset", offset, path)
		}
		query.Offset = int(o)
	} else {
		query.Offset = 0
	}

	// limit
	limit, exists := in["limit"]
	if exists {
		l, ok := limit.(float64)
		if !ok {
			return nil, fmt.Errorf("Invalid type: expected=number, got=%T, path=%s.limit", limit, path)
		}
		query.Limit = int(l)
	} else {
		query.Limit = 10
	}
	if query.Limit > 999 {
		return nil, fmt.Errorf("Invalid value (too large limit): path=%s.limit", path)
	}

	return &query, nil
}

func DecodeUpdateQuery(input interface{}, path string) (*UpdateQuery, error) {
	return nil, nil
}

func DecodeDeleteQuery(input interface{}, path string) (*DeleteQuery, error) {
	return nil, nil
}

func DecodeExpr(input interface{}, path string) (interface{}, error) {
	if e, err := DecodeMetadataExpr(input, path); err == nil {
		return e, nil
	}
	if e, err := DecodeColumnExpr(input, path); err == nil {
		return e, nil
	}
	if e, err := DecodeValueExpr(input, path); err == nil {
		return e, nil
	}
	if e, err := DecodeFuncExpr(input, path); err == nil {
		return e, nil
	}
	return nil, fmt.Errorf("Did not match any schema: path=%s", path)
}

func DecodeMetadataExpr(input interface{}, path string) (*MetadataExpr, error) {
	in, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=object, got=%T, path=%s", input, path)
	}

	var expr MetadataExpr

	// metadata
	metadata, exists := in["metadata"]
	if !exists {
		return nil, fmt.Errorf(".metadata required: path=%s", path)
	}
	d, ok := metadata.(string)
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=string, got=%T, path=%s.metadata", metadata, path)
	}
	for _, key := range []string{"id", "created_at"} {
		if key == d {
			expr.Metadata = d
		}
	}
	if expr.Metadata == "" {
		return nil, fmt.Errorf("Invalid value: path=%s.metadata", path)
	}

	return &expr, nil
}

func DecodeColumnExpr(input interface{}, path string) (*ColumnExpr, error) {
	in, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=object, got=%T, path=%s", input, path)
	}

	var expr ColumnExpr

	// column
	column, exists := in["column"]
	if !exists {
		return nil, fmt.Errorf(".column required: path=%s", path)
	}
	col, ok := column.(string)
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=string, got=%T, path=%s.column", column, path)
	}
	id, err := uuid.Parse(col)
	if err != nil {
		return nil, fmt.Errorf("Invalid column id format (uuid expected): path=%s.column", path)
	}
	expr.ColumnID = id

	return &expr, nil
}

func DecodeValueExpr(input interface{}, path string) (*ValueExpr, error) {
	in, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=object, got=%T, path=%s", input, path)
	}

	var expr ValueExpr

	// value
	value, exists := in["value"]
	if !exists {
		return nil, fmt.Errorf(".value required: path=%s", path)
	}
	expr.Value = value

	return &expr, nil
}

func DecodeFuncExpr(input interface{}, path string) (*FuncExpr, error) {
	in, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=object, got=%T, path=%s", input, path)
	}

	var expr FuncExpr

	// func
	fn, exists := in["func"]
	if !exists {
		return nil, fmt.Errorf(".func required: path=%s", path)
	}
	f, ok := fn.(string)
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=string, got=%T, path=%s.func", fn, path)
	}
	var found bool
	for _, s := range []string{"count"} {
		if s == f {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("Invalid func: path=%s.func", path)
	}
	expr.Func = f

	// args
	if args, exists := in["args"]; exists {
		ags, ok := args.([]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid type: expected=array, got=%T, path=%s.args", args, path)
		}
		for i, ag := range ags {
			e, err := DecodeExpr(ag, fmt.Sprintf("%s.args[%d]", path, i))
			if err != nil {
				return nil, err
			}
			expr.Args = append(expr.Args, reflect.ValueOf(e).Elem().Interface())
		}
	}

	return &expr, nil
}

func DecodeSortKey(input interface{}, path string) (*SortKey, error) {
	in, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid type: expected=object, got=%T, path=%s", input, path)
	}

	var expr SortKey

	// key
	key, exists := in["key"]
	if !exists {
		return nil, fmt.Errorf(".key required: path=%s", path)
	}
	e, err := DecodeExpr(key, fmt.Sprintf("%s.key", path))
	if err != nil {
		return nil, err
	}
	expr.Key = reflect.ValueOf(e).Elem().Interface()

	// order
	order, exists := in["order"]
	if exists {
		o, ok := order.(string)
		if !ok {
			return nil, fmt.Errorf("Invalid type: expected=string, got=%T, path=%s.order", input, path)
		}
		if o != "asc" && o != "desc" {
			return nil, fmt.Errorf("Invalid value (\"asc\" or \"desc\" expected): path=%s.order", path)
		}
		expr.Order = o
	} else {
		expr.Order = "asc"
	}

	return &expr, nil
}

type InsertQueryResult struct {
	RecordIDs []uuid.UUID `json:"record_ids"`
}

func (q InsertQueryResult) MarshalJSON() ([]byte, error) {
	if q.RecordIDs == nil {
		q.RecordIDs = []uuid.UUID{}
	}
	type Alias InsertQueryResult
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(q)})
}

type SelectQueryResult struct {
	Records [][]interface{} `json:"records"`
}

func (q SelectQueryResult) MarshalJSON() ([]byte, error) {
	if q.Records == nil {
		q.Records = [][]interface{}{}
	}
	type Alias SelectQueryResult
	return json.Marshal(&struct{ Alias }{Alias: (Alias)(q)})
}

type UpdateQueryResult struct {
}

type DeleteQueryResult struct {
}
