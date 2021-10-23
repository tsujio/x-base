package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/utils/strings"
)

type Organization struct {
	ID        UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetOrganizationListOpts struct {
	Sort          string
	Offset, Limit int
}

func GetOrganizationList(db *gorm.DB, opts *GetOrganizationListOpts) ([]Organization, int64, error) {
	var organizations []Organization
	var totalCount int64
	err := db.Model(&Organization{}).
		Count(&totalCount).
		Order(strings.ToSnakeCase(opts.Sort)).Offset(opts.Offset).Limit(opts.Limit).Find(&organizations).
		Error
	if err != nil {
		return nil, 0, xerrors.Errorf("Failed to get models: %w", err)
	}
	return organizations, totalCount, nil
}

func (o *Organization) Create(db *gorm.DB) error {
	if o.ID == UUID(uuid.Nil) {
		id, err := uuid.NewRandom()
		if err != nil {
			return xerrors.Errorf("Failed to generate id: %w", err)
		}
		o.ID = UUID(id)
	}

	err := db.Create(o).Error
	if err != nil {
		return xerrors.Errorf("Failed to create model: %w", err)
	}
	return nil
}

func (o *Organization) Save(db *gorm.DB) error {
	if o.ID == UUID(uuid.Nil) {
		return fmt.Errorf("Empty id")
	}
	err := db.Save(o).Error
	if err != nil {
		return xerrors.Errorf("Failed to save model: %w", err)
	}
	return nil
}

func (o *Organization) Get(db *gorm.DB) (*Organization, error) {
	err := db.Where("id = ?", o.ID).First(o).Error
	if err != nil {
		return nil, xerrors.Errorf("Failed to get model: %w", err)
	}
	return o, nil
}

func (o *Organization) Delete(db *gorm.DB) error {
	err := db.Where("id = ?", o.ID).Delete(o).Error
	if err != nil {
		return xerrors.Errorf("Failed to delete model: %w", err)
	}
	return nil
}
