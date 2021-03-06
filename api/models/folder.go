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

type GetFolderChildrenOpts struct {
	Sort          []GetListSortKey
	Offset, Limit int
	ComputePath   bool
}

func (f *Folder) GetChildren(db *gorm.DB, opts *GetFolderChildrenOpts) ([]TableFilesystemEntry, int64, error) {
	var path []TableFilesystemPathEntry
	if opts.ComputePath && f.ID != UUID(uuid.Nil) {
		e := &TableFilesystemEntry{ID: f.ID}
		if err := e.ComputePath(db); err != nil {
			return nil, 0, xerrors.Errorf("Failed to get path: %w", err)
		}
		path = append(e.Path, TableFilesystemPathEntry{
			ID:         f.ID,
			Type:       f.Type,
			Properties: f.Properties,
		})
	}

	conds := []func(db *gorm.DB) *gorm.DB{
		func(db *gorm.DB) *gorm.DB {
			return db.Where("organization_id = ?", f.OrganizationID)
		},
		func(db *gorm.DB) *gorm.DB {
			if f.ID == UUID(uuid.Nil) {
				return db.Where("parent_folder_id IS NULL")
			} else {
				return db.Where("parent_folder_id = ?", f.ID)
			}
		},
	}

	order, err := convertGetListSortKeyToOrderString(opts.Sort, []string{"id", "type", "created_at", "updated_at"})
	if err != nil {
		return nil, 0, xerrors.Errorf("Failed to convert sort key: %w", err)
	}

	var children []TableFilesystemEntry
	var totalCount int64
	err = db.Model(&TableFilesystemEntry{}).Scopes(conds...).
		Count(&totalCount).
		Order(order).Offset(opts.Offset).Limit(opts.Limit).Find(&children).
		Error
	if err != nil {
		return nil, 0, xerrors.Errorf("Failed to children: %w", err)
	}
	if opts.ComputePath {
		for i := range children {
			children[i].Path = path
		}
	}

	return children, totalCount, nil
}

func (f *Folder) Create(db *gorm.DB) error {
	if f.ID == UUID(uuid.Nil) {
		id, err := uuid.NewRandom()
		if err != nil {
			return xerrors.Errorf("Failed to generate id: %w", err)
		}
		f.ID = UUID(id)
	}
	f.Type = "folder"

	err := db.Create(&f.TableFilesystemEntry).Error
	if err != nil {
		return xerrors.Errorf("Failed to create base model: %w", err)
	}
	err = db.Select("ID").Omit("CreatedAt", "UpdatedAt").Create(&f).Error
	if err != nil {
		return xerrors.Errorf("Failed to create model: %w", err)
	}
	return nil
}

func (f *Folder) Save(db *gorm.DB) error {
	if f.ID == UUID(uuid.Nil) {
		return fmt.Errorf("Empty id")
	}
	f.Type = "folder"
	err := db.Save(&f.TableFilesystemEntry).Error
	if err != nil {
		return xerrors.Errorf("Failed to save model: %w", err)
	}
	return nil
}

func (f *Folder) Get(db *gorm.DB) (*Folder, error) {
	err := db.Model(&TableFilesystemEntry{}).
		Where("id = ?", f.ID).
		Joins("INNER JOIN folders USING (id)").
		First(&f.TableFilesystemEntry).
		Error
	if err != nil {
		return nil, xerrors.Errorf("Failed to get model: %w", err)
	}
	return f, nil
}

func (f *Folder) Delete(db *gorm.DB) error {
	err := db.Where("id = ?", f.ID).Delete(&f.TableFilesystemEntry).Error
	if err != nil {
		return xerrors.Errorf("Failed to delete model: %w", err)
	}
	return nil
}
