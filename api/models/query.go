package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type SQLBuilder interface {
	BuildSQL() (string, []interface{}, error)
}

type InsertQuery struct {
	TableID UUID
	Columns []ColumnExpr
	Values  [][]ValueExpr
}

func (q *InsertQuery) Execute(db *gorm.DB) ([]UUID, error) {
	sql := `INSERT INTO table_records(id, table_id, data, created_at, updated_at)`

	now := time.Now().UTC().Truncate(time.Second)

	var ids []UUID
	var params []interface{}
	sql += ` VALUES`
	for i, record := range q.Values {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return nil, xerrors.Errorf("Failed to generate uuid: %w", err)
		}
		id := UUID(uuid)
		ids = append(ids, id)

		data := map[string]interface{}{}
		for j, v := range record {
			data[q.Columns[j].Column.ID.String()] = v.Value
		}
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return nil, xerrors.Errorf("Failed to serialize data: %w", err)
		}

		if i > 0 {
			sql += `,`
		}
		sql += ` (?, ?, ?, ?, ?)`
		params = append(params, id, q.TableID, dataJSON, now, now)
	}

	if err := db.Exec(sql, params...).Error; err != nil {
		return nil, xerrors.Errorf("Failed to execute query: %w", err)
	}

	return ids, nil
}

type SelectQuery struct {
	Columns []SelectColumn
	From    interface{}
	OrderBy []SortKey
	Offset  *int
	Limit   *int
}

func (q *SelectQuery) Execute(db *gorm.DB, dest interface{}) error {
	var sql string
	var params []interface{}
	switch t := q.From.(type) {
	case TableExpr:
		sql = `SELECT`
		for i, c := range q.Columns {
			if i > 0 {
				sql += `,`
			}
			s, p, err := c.BuildSQL()
			if err != nil {
				return xerrors.Errorf("Failed to build select column sql: %w", err)
			}
			sql += s
			params = append(params, p...)
		}

		sql += `
		FROM table_records
		WHERE table_id = ?
		`
		params = append(params, t.Table.ID)

		if len(q.OrderBy) > 0 {
			sql += ` ORDER BY`
			for i, o := range q.OrderBy {
				if i > 0 {
					sql += ","
				}
				s, p, err := o.BuildSQL()
				if err != nil {
					return xerrors.Errorf("Failed to build order by sql: %w", err)
				}
				sql += s
				params = append(params, p...)
			}
		}

		if q.Limit != nil {
			sql += fmt.Sprintf(" LIMIT %d", *q.Limit)
		}
		if q.Offset != nil {
			sql += fmt.Sprintf(" OFFSET %d", *q.Offset)
		}
	default:
		return fmt.Errorf("Invalid table type: %T", t)
	}

	// Execute query
	if err := db.Raw(sql, params...).Scan(dest).Error; err != nil {
		return xerrors.Errorf("Failed to execute query: %w", err)
	}

	return nil
}

type MetadataExprKey string

const (
	MetadataExprKeyID        MetadataExprKey = "id"
	MetadataExprKeyCreatedAt MetadataExprKey = "created_at"
)

type MetadataExpr struct {
	Key MetadataExprKey
}

func (e MetadataExpr) BuildSQL() (string, []interface{}, error) {
	switch e.Key {
	case MetadataExprKeyID:
		return " id_string ", nil, nil
	default:
		return " " + string(e.Key) + " ", nil, nil
	}
}

type ColumnExpr struct {
	Column Column
}

func (e ColumnExpr) BuildSQL() (string, []interface{}, error) {
	var typ string
	switch e.Column.Type {
	case "string":
		typ = "CHAR"
	case "integer", "boolean":
		typ = "SIGNED"
	case "float":
		typ = "DECIMAL(65, 30)"
	default:
		typ = "JSON"
	}

	id := e.Column.ID.String()
	sql := fmt.Sprintf(`
	CAST(CASE WHEN JSON_EXTRACT(data, '$."%s"') IS NULL OR JSON_TYPE(JSON_EXTRACT(data, '$."%s"')) = 'NULL' THEN NULL
	          WHEN JSON_TYPE(JSON_EXTRACT(data, '$."%s"')) = 'STRING' THEN JSON_UNQUOTE(JSON_EXTRACT(data, '$."%s"'))
	          WHEN JSON_TYPE(JSON_EXTRACT(data, '$."%s"')) = 'BOOLEAN' THEN CASE WHEN JSON_EXTRACT(data, '$."%s"') THEN TRUE ELSE FALSE END
	          ELSE JSON_EXTRACT(data, '$."%s"')
	     END AS %s)
	`, id, id, id, id, id, id, id, typ)

	// Convert DECIMAL to DOUBLE
	if e.Column.Type == "float" {
		sql += " + 0E0 "
	}

	return sql, nil, nil
}

type ValueExpr struct {
	Value interface{}
}

func (e ValueExpr) BuildSQL() (string, []interface{}, error) {
	return " ? ", []interface{}{e.Value}, nil
}

type TableExpr struct {
	Table Table
}

type SortKeyOrder string

const (
	SortKeyOrderAsc  SortKeyOrder = "ASC"
	SortKeyOrderDesc SortKeyOrder = "DESC"
)

type SelectColumn struct {
	Column SQLBuilder
	As     string
}

func (c SelectColumn) BuildSQL() (string, []interface{}, error) {
	sql, params, err := c.Column.BuildSQL()
	if err != nil {
		return "", nil, xerrors.Errorf("Failed to build column sql: %w", err)
	}

	if c.As != "" {
		if matched, err := regexp.Match(`^\w+$`, []byte(c.As)); err != nil || !matched {
			return "", nil, fmt.Errorf("Invalid select column alias")
		}
		sql += fmt.Sprintf(` AS %s `, c.As)
	}

	return sql, params, nil
}

type SortKey struct {
	Key   SQLBuilder
	Order SortKeyOrder
}

func (k SortKey) BuildSQL() (string, []interface{}, error) {
	sql, params, err := k.Key.BuildSQL()
	if err != nil {
		return "", nil, xerrors.Errorf("Failed to build sort key sql: %w", err)
	}

	sql += " " + string(k.Order) + " "

	return sql, params, nil
}
