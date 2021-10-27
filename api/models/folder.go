package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/api/utils/strings"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type Folder struct {
	TableFilesystemEntry
}

type GetFolderChildrenSortKey struct {
	Key         string
	OrderAsc    bool
	OrderDesc   bool
	OrderValues []string
}

type GetFolderChildrenOpts struct {
	Sort          []GetFolderChildrenSortKey
	Offset, Limit int
	ComputePath   bool
}

func (f *Folder) GetChildren(db *gorm.DB, opts *GetFolderChildrenOpts) ([]TableFilesystemEntry, int64, error) {
	var parentPath []TableFilesystemPathEntry
	if opts.ComputePath && f.ID != UUID(uuid.Nil) {
		e := &TableFilesystemEntry{ID: f.ID}
		if err := e.ComputePath(db); err != nil {
			return nil, 0, xerrors.Errorf("Failed to get path: %w", err)
		}
		parentPath = e.Path
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

	var order string
	if len(opts.Sort) == 0 {
		order = "id"
	} else {
		for i, s := range opts.Sort {
			if i > 0 {
				order += ", "
			}
			k := strings.ToSnakeCase(s.Key)
			switch k {
			case "id", "created_at", "updated_at":
				var o string
				if s.OrderAsc {
					o = "ASC"
				} else if s.OrderDesc {
					o = "DESC"
				} else {
					return nil, 0, fmt.Errorf("Invalid sort option (expected 'asc' or 'desc')")
				}
				order += k + " " + o
			case "type":
				if len(s.OrderValues) == 0 {
					return nil, 0, fmt.Errorf("Invalid sort option (empty value list)")
				}
				order += "CASE"
				for i, v := range s.OrderValues {
					var match bool
					for _, t := range []string{"table", "folder"} {
						if v == t {
							match = true
							break
						}
					}
					if !match {
						return nil, 0, fmt.Errorf("Invalid sort option (value list)")
					}
					order += fmt.Sprintf(" WHEN type = '%s' THEN %d", v, i)
				}
				order += fmt.Sprintf(" ELSE %d END ASC", len(s.OrderValues))
			default:
				return nil, 0, fmt.Errorf("Invalid sort key: %s", s.Key)
			}
		}
	}

	var children []TableFilesystemEntry
	var totalCount int64
	err := db.Model(&TableFilesystemEntry{}).Scopes(conds...).
		Count(&totalCount).
		Order(order).Offset(opts.Offset).Limit(opts.Limit).Find(&children).
		Error
	if err != nil {
		return nil, 0, xerrors.Errorf("Failed to children: %w", err)
	}
	for i := range children {
		if opts.ComputePath {
			c := &children[i]
			c.Path = append(append([]TableFilesystemPathEntry{}, parentPath...), TableFilesystemPathEntry{
				ID:   c.ID,
				Type: c.Type,
			})
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
