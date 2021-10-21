package models

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/utils"
)

type Folder struct {
	TableFilesystemEntry
}

type GetFolderChildrenOpts struct {
	Sort          string
	Offset, Limit int
	ComputePath   bool
}

func (f *Folder) GetChildren(db *gorm.DB, opts *GetFolderChildrenOpts) ([]interface{}, int64, error) {
	var parentPath []TableFilesystemPathEntry
	if opts.ComputePath && f.ID != UUID(uuid.Nil) {
		e := &TableFilesystemEntry{ID: f.ID, Type: "folder"}
		if err := e.ComputePath(db); err != nil {
			return nil, 0, xerrors.Errorf("Failed to get path: %w", err)
		}
		parentPath = e.Path
	}

	parentFolderIDCond := func(db *gorm.DB) *gorm.DB {
		if f.ID == UUID(uuid.Nil) {
			return db.Where("parent_folder_id IS NULL")
		} else {
			return db.Where("parent_folder_id = ?", f.ID)
		}
	}

	var ret []interface{}
	var totalCount int64

	// Fetch folders
	var folders []Folder
	var folderTotalCount int64
	err := db.Model(&Folder{}).Scopes(parentFolderIDCond).
		Count(&folderTotalCount).
		Order(utils.ToSnakeCase(opts.Sort)).Offset(opts.Offset).Limit(opts.Limit).Find(&folders).
		Error
	if err != nil {
		return nil, 0, xerrors.Errorf("Failed to get folders: %w", err)
	}
	for _, f := range folders {
		if opts.ComputePath {
			f.Path = append([]TableFilesystemPathEntry{}, parentPath...)
			f.Path = append(f.Path, TableFilesystemPathEntry{
				ID:   f.ID,
				Type: f.Type,
				Name: f.Name,
			})
		}
		ret = append(ret, f)
	}
	totalCount += folderTotalCount

	// Fetch tables
	offset := opts.Offset - int(totalCount)
	if offset < 0 {
		offset = 0
	}
	limit := opts.Limit - len(ret)
	if limit < 0 {
		limit = 0
	}
	var tables []Table
	var tableTotalCount int64
	err = db.Model(&Table{}).Scopes(parentFolderIDCond).
		Count(&tableTotalCount).
		Order(utils.ToSnakeCase(opts.Sort)).Offset(offset).Limit(limit).Find(&tables).
		Error
	if err != nil {
		return nil, 0, xerrors.Errorf("Failed to get tables: %w", err)
	}
	for _, t := range tables {
		if opts.ComputePath {
			t.Path = append([]TableFilesystemPathEntry{}, parentPath...)
			t.Path = append(t.Path, TableFilesystemPathEntry{
				ID:   t.ID,
				Type: t.Type,
				Name: t.Name,
			})
		}
		ret = append(ret, t)
	}
	totalCount += tableTotalCount

	return ret, totalCount, nil
}

func (f *Folder) BeforeSave(db *gorm.DB) error {
	if err := f.TableFilesystemEntry.BeforeSave(db); err != nil {
		return err
	}

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
