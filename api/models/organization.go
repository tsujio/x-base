package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/utils"
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
		Order(utils.ToSnakeCase(opts.Sort)).Offset(opts.Offset).Limit(opts.Limit).Find(&organizations).
		Error
	if err != nil {
		return nil, 0, xerrors.Errorf("Failed to get organizations: %w", err)
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

	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC().Truncate(time.Second)
	}
	o.UpdatedAt = o.CreatedAt

	err := db.Create(o).Error
	if err != nil {
		return xerrors.Errorf("Failed to create organization: %w", err)
	}
	return nil
}

func (o *Organization) Get(db *gorm.DB) (*Organization, error) {
	err := db.Model(o).Where("id = ?", o.ID).First(o).Error
	if err != nil {
		return nil, xerrors.Errorf("Failed to get organization: %w", err)
	}
	return o, nil
}
