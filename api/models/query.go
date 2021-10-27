package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type TableRecord struct {
	ID        UUID
	TableID   UUID
	Data      JSON
	CreatedAt time.Time
	UpdatedAt time.Time
}

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
	Where   SQLBuilder
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

		if q.Where != nil {
			s, p, err := q.Where.BuildSQL()
			if err != nil {
				return xerrors.Errorf("Failed to build where sql: %w", err)
			}
			sql += " AND (" + s + ") "
			params = append(params, p...)
		}

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
	rows, err := db.Raw(sql, params...).Rows()
	if err != nil {
		return xerrors.Errorf("Failed to execute query: %w", err)
	}
	defer rows.Close()
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return xerrors.Errorf("Failed to get column types: %w", err)
	}
	var records []map[string]interface{}
	for rows.Next() {
		record := map[string]interface{}{}
		if err := db.ScanRows(rows, &record); err != nil {
			return xerrors.Errorf("Failed to scan row: %w", err)
		}
		for _, t := range colTypes {
			if t.DatabaseTypeName() == "JSON" {
				if val, exists := record[t.Name()]; exists {
					if s, ok := val.(string); ok {
						var v interface{}
						if err := json.Unmarshal([]byte(s), &v); err == nil {
							record[t.Name()] = v
						}
					}
				}
			}
		}
		records = append(records, record)
	}
	reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(records))

	return nil
}

type UpdateQuery struct {
	Table interface{}
	Set   []UpdateSet
	Where SQLBuilder
}

func (q *UpdateQuery) Execute(db *gorm.DB) error {
	sql := "UPDATE "
	var params []interface{}

	switch t := q.Table.(type) {
	case TableExpr:
		sql += " table_records SET data = JSON_SET(data, "

		for i, us := range q.Set {
			if i > 0 {
				sql += `,`
			}

			s, p, err := us.Value.BuildSQL()
			if err != nil {
				return xerrors.Errorf("Failed to build value sql: %w", err)
			}

			sql += fmt.Sprintf(`'$."%s"', %s`, us.To.Column.ID, s)
			params = append(params, p...)
		}
		sql += ")"

		sql += `
		WHERE table_id = ?
		`
		params = append(params, t.Table.ID)

		s, p, err := q.Where.BuildSQL()
		if err != nil {
			return xerrors.Errorf("Failed to build where sql: %w", err)
		}
		sql += " AND (" + s + ") "
		params = append(params, p...)
	default:
		return fmt.Errorf("Invalid table type: %T", t)
	}

	// Execute query
	if err := db.Exec(sql, params...).Error; err != nil {
		return xerrors.Errorf("Failed to execute query: %w", err)
	}

	return nil
}

type DeleteQuery struct {
	Table interface{}
	Where SQLBuilder
}

func (q *DeleteQuery) Execute(db *gorm.DB) error {
	sql := "DELETE FROM "
	var params []interface{}

	switch t := q.Table.(type) {
	case TableExpr:
		sql += `
		table_records
		WHERE table_id = ?
		`
		params = append(params, t.Table.ID)

		s, p, err := q.Where.BuildSQL()
		if err != nil {
			return xerrors.Errorf("Failed to build where sql: %w", err)
		}
		sql += " AND (" + s + ") "
		params = append(params, p...)
	default:
		return fmt.Errorf("Invalid table type: %T", t)
	}

	// Execute query
	if err := db.Exec(sql, params...).Error; err != nil {
		return xerrors.Errorf("Failed to execute query: %w", err)
	}

	return nil
}

type MetadataExprKey int

const (
	MetadataExprKeyID MetadataExprKey = iota
	MetadataExprKeyCreatedAt
)

type MetadataExpr struct {
	Key MetadataExprKey
}

func (e MetadataExpr) BuildSQL() (string, []interface{}, error) {
	switch e.Key {
	case MetadataExprKeyID:
		return " id_string ", nil, nil
	case MetadataExprKeyCreatedAt:
		return " created_at ", nil, nil
	default:
		return "", nil, fmt.Errorf("Invalid metadata key: %v", e.Key)
	}
}

type ColumnExpr struct {
	Column Column
}

func (e ColumnExpr) BuildSQL() (string, []interface{}, error) {
	id := e.Column.ID.String()
	sql := fmt.Sprintf(`
	CAST(CASE WHEN JSON_EXTRACT(data, '$."%s"') IS NULL OR JSON_TYPE(JSON_EXTRACT(data, '$."%s"')) = 'NULL' THEN NULL
	          ELSE JSON_EXTRACT(data, '$."%s"')
	     END AS JSON)
	`, id, id, id)

	return sql, nil, nil
}

type ValueExpr struct {
	Value interface{}
}

func (e ValueExpr) BuildSQL() (string, []interface{}, error) {
	return " ? ", []interface{}{e.Value}, nil
}

type FuncExprFunc int

const (
	FuncExprFuncCount FuncExprFunc = iota
)

type FuncExpr struct {
	Func FuncExprFunc
	Args []SQLBuilder
}

func (e FuncExpr) BuildSQL() (string, []interface{}, error) {
	var args []string
	var params []interface{}
	for _, arg := range e.Args {
		s, p, err := arg.BuildSQL()
		if err != nil {
			return "", nil, xerrors.Errorf("Failed to build func arg sql: %w", err)
		}
		args = append(args, s)
		params = append(params, p...)
	}

	var fn string
	switch e.Func {
	case FuncExprFuncCount:
		fn = "COUNT"
	default:
		return "", nil, fmt.Errorf("Invalid func: %v", e.Func)
	}

	return fmt.Sprintf(" %s(%s) ", fn, strings.Join(args, ", ")), params, nil
}

type UnaryOpExpr struct {
	Op SQLBuilder
}

func (e UnaryOpExpr) BuildSQL() (string, []interface{}, error) {
	s, p, err := e.Op.BuildSQL()
	if err != nil {
		return "", nil, xerrors.Errorf("Failed to build operand sql: %w", err)
	}
	return s, p, nil
}

type BinOpExpr struct {
	Op1 SQLBuilder
	Op2 SQLBuilder
}

func (e BinOpExpr) BuildSQL() ([]string, [][]interface{}, error) {
	s1, p1, err := e.Op1.BuildSQL()
	if err != nil {
		return nil, nil, xerrors.Errorf("Failed to build first operand sql: %w", err)
	}
	s2, p2, err := e.Op2.BuildSQL()
	if err != nil {
		return nil, nil, xerrors.Errorf("Failed to build second operand sql: %w", err)
	}
	return []string{s1, s2}, [][]interface{}{p1, p2}, nil
}

type EqExpr BinOpExpr

func (e EqExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) = (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type NeExpr BinOpExpr

func (e NeExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) != (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type GtExpr BinOpExpr

func (e GtExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) > (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type GeExpr BinOpExpr

func (e GeExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) >= (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type LtExpr BinOpExpr

func (e LtExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) < (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type LeExpr BinOpExpr

func (e LeExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) <= (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type LikeExpr BinOpExpr

func (e LikeExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) LIKE (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type IsNullExpr UnaryOpExpr

func (e IsNullExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := UnaryOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) IS NULL ", s), p, nil
	}
}

type AndExpr BinOpExpr

func (e AndExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) AND (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type OrExpr BinOpExpr

func (e OrExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) OR (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type NotExpr UnaryOpExpr

func (e NotExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := UnaryOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" NOT (%s) ", s), p, nil
	}
}

type AddExpr BinOpExpr

func (e AddExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) + (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type SubExpr BinOpExpr

func (e SubExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) - (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type MulExpr BinOpExpr

func (e MulExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) * (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type DivExpr BinOpExpr

func (e DivExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) / (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type ModExpr BinOpExpr

func (e ModExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := BinOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" (%s) %% (%s) ", s[0], s[1]), append(append([]interface{}{}, p[0]...), p[1]...), nil
	}
}

type NegExpr UnaryOpExpr

func (e NegExpr) BuildSQL() (string, []interface{}, error) {
	if s, p, err := UnaryOpExpr(e).BuildSQL(); err != nil {
		return "", nil, err
	} else {
		return fmt.Sprintf(" - (%s) ", s), p, nil
	}
}

type TableExpr struct {
	Table Table
}

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

type SortKeyOrder int

const (
	SortKeyOrderAsc SortKeyOrder = iota
	SortKeyOrderDesc
)

type SortKey struct {
	Key   SQLBuilder
	Order SortKeyOrder
}

func (k SortKey) BuildSQL() (string, []interface{}, error) {
	sql, params, err := k.Key.BuildSQL()
	if err != nil {
		return "", nil, xerrors.Errorf("Failed to build sort key sql: %w", err)
	}

	switch k.Order {
	case SortKeyOrderAsc:
		sql += " ASC "
	case SortKeyOrderDesc:
		sql += " DESC "
	default:
		return "", nil, fmt.Errorf("Invalid sort order: %v", k.Order)
	}

	return sql, params, nil
}

type UpdateSet struct {
	To    ColumnExpr
	Value SQLBuilder
}
