package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type Column struct {
	ID        UUID
	TableID   UUID
	Index     int
	Name      string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func IsValidColumnType(typ string) bool {
	for _, t := range []string{"string"} {
		if t == typ {
			return true
		}
	}
	return false
}

func (c *Column) BeforeSave(db *gorm.DB) error {
	if !IsValidColumnType(c.Type) {
		return fmt.Errorf("Invalid column type (%s)", c.Type)
	}

	if err := db.Model(c).
		Where("table_id = ? AND `index` >= ?", c.TableID, c.Index).
		UpdateColumn("`index`", gorm.Expr("`index` + ?", 1)).
		Error; err != nil {
		return err
	}

	return nil
}

func CanonicalizeColumnIndices(db *gorm.DB, tableID UUID) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Column{}).
			Where("table_id = ?", tableID).
			UpdateColumn("`index`", gorm.Expr("`index` + ?", 9999)).
			Error; err != nil {
			return err
		}

		return tx.Exec(`
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
	})
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
			return err
		}
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
			return err
		}
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
			return err
		}
	}

	return nil
}
