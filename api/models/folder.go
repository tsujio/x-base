package models

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type Folder struct {
	TableFilesystemEntry
}

func (f *Folder) BeforeSave(*gorm.DB) error {
	f.Type = "folder"
	return nil
}

func (f *Folder) Create(db *gorm.DB) error {
	if f.ID == UUID(uuid.Nil) {
		id, err := uuid.NewRandom()
		if err != nil {
			return xerrors.Errorf("Failed to generate id: %w", err)
		}
		f.ID = UUID(id)
	}

	err := db.Create(f).Error
	if err != nil {
		return xerrors.Errorf("Failed to create model: %w", err)
	}
	return nil
}

func (f *Folder) Save(db *gorm.DB) error {
	if f.ID == UUID(uuid.Nil) {
		return fmt.Errorf("Empty id")
	}
	err := db.Save(f).Error
	if err != nil {
		return xerrors.Errorf("Failed to save model: %w", err)
	}
	return nil
}

func (f *Folder) Get(db *gorm.DB) (*Folder, error) {
	err := db.Where("id = ?", f.ID).First(f).Error
	if err != nil {
		return nil, xerrors.Errorf("Failed to get model: %w", err)
	}
	return f, nil
}

func (f *Folder) Delete(db *gorm.DB) error {
	err := db.Where("id = ?", f.ID).Delete(f).Error
	if err != nil {
		return xerrors.Errorf("Failed to delete model: %w", err)
	}
	return nil
}
