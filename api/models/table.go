package models

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type Table struct {
	TableFilesystemEntry
	Columns []Column `gorm:"-"`
}

func (t *Table) Create(db *gorm.DB) error {
	if t.ID == UUID(uuid.Nil) {
		id, err := uuid.NewRandom()
		if err != nil {
			return xerrors.Errorf("Failed to generate id: %w", err)
		}
		t.ID = UUID(id)
	}
	t.Type = "table"

	err := db.Create(&t.TableFilesystemEntry).Error
	if err != nil {
		return xerrors.Errorf("Failed to create base model: %w", err)
	}
	err = db.Select("ID").Omit("CreatedAt", "UpdatedAt").Create(&t).Error
	if err != nil {
		return xerrors.Errorf("Failed to create model: %w", err)
	}
	return nil
}

func (t *Table) Save(db *gorm.DB) error {
	if t.ID == UUID(uuid.Nil) {
		return fmt.Errorf("Empty id")
	}
	t.Type = "table"
	err := db.Save(&t.TableFilesystemEntry).Error
	if err != nil {
		return xerrors.Errorf("Failed to save model: %w", err)
	}
	return nil
}

func (t *Table) Get(db *gorm.DB) (*Table, error) {
	err := db.Model(&TableFilesystemEntry{}).
		Where("id = ?", t.ID).
		Joins("INNER JOIN tables USING (id)").
		First(&t.TableFilesystemEntry).
		Error
	if err != nil {
		return nil, xerrors.Errorf("Failed to get model: %w", err)
	}
	return t, nil
}

func (t *Table) Delete(db *gorm.DB) error {
	err := db.Where("id = ?", t.ID).Delete(&t.TableFilesystemEntry).Error
	if err != nil {
		return xerrors.Errorf("Failed to delete model: %w", err)
	}
	return nil
}

func (t *Table) FetchColumns(db *gorm.DB) error {
	var columns []Column
	err := db.Where("table_id = ?", t.ID).Order("`index`").Find(&columns).Error
	if err != nil {
		return xerrors.Errorf("Failed to get columns: %w", err)
	}
	t.Columns = columns
	return nil
}
