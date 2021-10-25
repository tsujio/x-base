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
			data[q.Columns[j].ColumnID.String()] = v.Value
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
		for i, col := range q.Columns {
			if i > 0 {
				sql += `,`
			}

			switch c := col.Column.(type) {
			case MetadataExpr:
				switch c.Key {
				case MetadataExprKeyID:
					sql += " id_string"
				default:
					sql += " " + string(c.Key)
				}
			case ColumnExpr:
				sql += fmt.Sprintf(`
				CASE
				    WHEN JSON_TYPE(JSON_EXTRACT(data, '$."%s"')) = 'NULL' THEN NULL
				    ELSE JSON_UNQUOTE(JSON_EXTRACT(data, '$."%s"'))
				END
				`, c.ColumnID.String(), c.ColumnID.String())
			case ValueExpr:
				sql += ` ?`
				params = append(params, c.Value)
			default:
				return fmt.Errorf("Invalid column type: %T", c)
			}

			if col.As != "" {
				if matched, err := regexp.Match(`^\w+$`, []byte(col.As)); err != nil || !matched {
					return fmt.Errorf("Invalid column alias")
				}
				sql += fmt.Sprintf(` AS %s`, col.As)
			}
		}

		sql += `
		FROM table_records
		WHERE table_id = ?
		`
		params = append(params, t.TableID)

		if len(q.OrderBy) > 0 {
			sql += ` ORDER BY`
			for i, o := range q.OrderBy {
				if i > 0 {
					sql += ","
				}

				switch k := o.Key.(type) {
				case MetadataExpr:
					switch k.Key {
					case MetadataExprKeyID:
						sql += " id_string"
					default:
						sql += " " + string(k.Key)
					}
				case ColumnExpr:
					sql += fmt.Sprintf(`
					CASE
					    WHEN JSON_TYPE(JSON_EXTRACT(data, '$."%s"')) = 'NULL' THEN NULL
					    ELSE JSON_UNQUOTE(JSON_EXTRACT(data, '$."%s"'))
					END
					`, k.ColumnID.String(), k.ColumnID.String())
				case ValueExpr:
					sql += ` ?`
					params = append(params, k.Value)
				default:
					return fmt.Errorf("Invalid sort key type: %T", k)
				}

				sql += " " + string(o.Order)
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

type ColumnExpr struct {
	ColumnID UUID
}

type ValueExpr struct {
	Value interface{}
}

type TableExpr struct {
	TableID UUID
}

type SortKeyOrder string

const (
	SortKeyOrderAsc  SortKeyOrder = "ASC"
	SortKeyOrderDesc SortKeyOrder = "DESC"
)

type SelectColumn struct {
	Column interface{}
	As     string
}

type SortKey struct {
	Key   interface{}
	Order SortKeyOrder
}
