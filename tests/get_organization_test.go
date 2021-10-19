package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestGetOrganization(t *testing.T) {
	var uuids []uuid.UUID
	for i := 0; i < 10; i++ {
		uuids = append(uuids, uuid.New())
	}

	createOrganizations := func(n int) ([]models.Organization, error) {
		var organizations []models.Organization
		for i := 0; i < n; i++ {
			o := models.Organization{
				ID:   models.UUID(uuids[i]),
				Name: fmt.Sprintf("organization-%02d", i+1),
			}
			err := o.Create(testutils.GetDB())
			if err != nil {
				return nil, err
			}
			organizations = append(organizations, o)
		}
		return organizations, nil
	}

	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/organizations/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Prepare: func(db *gorm.DB) error {
				_, err := createOrganizations(3)
				return err
			},
			Path:       makePath(uuids[1]),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         uuids[1].String(),
				"name":       "organization-02",
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
		},
		{
			Title: "Not found",
			Prepare: func(db *gorm.DB) error {
				_, err := createOrganizations(3)
				return err
			},
			Path:       makePath(uuids[4]),
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Not found",
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodGet

		testutils.RunTestCase(t, tc)
	}
}
