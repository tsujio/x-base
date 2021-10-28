package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

const ColumnTailIndex = 9999

type Column struct {
	ID         UUID
	TableID    UUID
	Index      int
	Properties Properties
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func moveColumnIndicesToTemporaryAddress(db *gorm.DB, tableID UUID) error {
	if err := db.Model(&Column{}).
		Where("table_id = ?", tableID).
		Order("`index` DESC").
		UpdateColumn("`index`", gorm.Expr("`index` + ?", ColumnTailIndex)).
		Error; err != nil {
		return xerrors.Errorf("Failed to execute query: %w", err)
	}
	return nil
}

func CanonicalizeColumnIndices(db *gorm.DB, tableID UUID) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := moveColumnIndicesToTemporaryAddress(tx, tableID); err != nil {
			return xerrors.Errorf("Failed to shift column indices: %w", err)
		}

		err := tx.Exec(`
		UPDATE columns AS c
		INNER JOIN (
		    SELECT id, ROW_NUMBER() OVER (PARTITION BY table_id ORDER BY `+"`index`"+`) - 1 AS `+"`index`"+`
		    FROM columns
		    WHERE table_id = ?
		) AS t
		USING (id)
		SET c.`+"`index`"+` = t.`+"`index`"+`
		WHERE c.table_id = ?
		`, tableID, tableID).Error
		if err != nil {
			return xerrors.Errorf("Failed to reset column indices: %w", err)
		}

		return nil
	})
}

func ReorderColumns(db *gorm.DB, tableID UUID, order []UUID) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := moveColumnIndicesToTemporaryAddress(tx, tableID); err != nil {
			return xerrors.Errorf("Failed to shift column indices: %w", err)
		}

		var cases string
		var params []interface{}
		for i, id := range order {
			cases += " WHEN ? THEN ? "
			params = append(params, id, i)
			if i > 0 && (i%10 == 0 || i == len(order)-1) {
				params = append(params, tableID)
				err := tx.Exec(fmt.Sprintf(`
				UPDATE columns
				SET `+"`index`"+` = CASE id %s ELSE `+"`index`"+` END
				WHERE table_id = ?
				`, cases), params...).Error
				if err != nil {
					return xerrors.Errorf("Failed to update column indices: %w", err)
				}
				cases = ""
				params = []interface{}{}
			}
		}

		if err := CanonicalizeColumnIndices(tx, tableID); err != nil {
			return xerrors.Errorf("Failed to canonicalize column indices: %w", err)
		}

		return nil
	})
}

func (c *Column) BeforeSave(db *gorm.DB) error {
	// Shift indices
	if err := db.Model(&Column{}).
		Where("table_id = ? AND `index` >= ?", c.TableID, c.Index).
		Order("`index` DESC").
		UpdateColumn("`index`", gorm.Expr("`index` + ?", 1)).
		Error; err != nil {
		return xerrors.Errorf("Failed to shift column indices: %w", err)
	}

	return nil
}

func (c *Column) Create(db *gorm.DB, skipCanonicalizeIndices bool) error {
	if c.ID == UUID(uuid.Nil) {
		id, err := uuid.NewRandom()
		if err != nil {
			return xerrors.Errorf("Failed to generate id: %w", err)
		}
		c.ID = UUID(id)
	}

	err := db.Create(c).Error
	if err != nil {
		return xerrors.Errorf("Failed to create model: %w", err)
	}

	if !skipCanonicalizeIndices {
		if err := CanonicalizeColumnIndices(db, c.TableID); err != nil {
			return xerrors.Errorf("Failed to canonicalize column indices: %w", err)
		}
		col, err := c.Get(db)
		if err != nil {
			return xerrors.Errorf("Failed to get column: %w", err)
		}
		*c = *col
	}

	return nil
}

func (c *Column) Save(db *gorm.DB, skipCanonicalizeIndices bool) error {
	if c.ID == UUID(uuid.Nil) {
		return fmt.Errorf("Empty id")
	}
	err := db.Save(c).Error
	if err != nil {
		return xerrors.Errorf("Failed to save model: %w", err)
	}

	if !skipCanonicalizeIndices {
		if err := CanonicalizeColumnIndices(db, c.TableID); err != nil {
			return xerrors.Errorf("Failed to canonicalize column indices: %w", err)
		}
		col, err := c.Get(db)
		if err != nil {
			return xerrors.Errorf("Failed to get column: %w", err)
		}
		*c = *col
	}

	return nil
}

func (c *Column) Get(db *gorm.DB) (*Column, error) {
	err := db.Where("id = ?", c.ID).First(c).Error
	if err != nil {
		return nil, xerrors.Errorf("Failed to get model: %w", err)
	}
	return c, nil
}

func (c *Column) Delete(db *gorm.DB, skipCanonicalizeIndices bool) error {
	err := db.Where("id = ?", c.ID).Delete(c).Error
	if err != nil {
		return xerrors.Errorf("Failed to delete model: %w", err)
	}

	if !skipCanonicalizeIndices {
		if err := CanonicalizeColumnIndices(db, c.TableID); err != nil {
			return xerrors.Errorf("Failed to canonicalize column indices: %w", err)
		}
	}

	return nil
}
