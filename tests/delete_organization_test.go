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

func TestDeleteOrganization(t *testing.T) {
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
			Title: "Delete",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(2)
				return err
			},
			Path:       makePath(uuids[0]),
			StatusCode: http.StatusOK,
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire
				res := testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", uuids[0]), nil)
				if res != nil {
					t.Errorf("[%s] Not deleted", tc.Title)
				}

				// Did not delete other data
				res = testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", uuids[1]), nil)
				if res == nil {
					t.Errorf("[%s] Deleted other data", tc.Title)
				}
			},
		},
		{
			Title: "Not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(2)
				return err
			},
			Path:       makePath(uuids[3]),
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Not found",
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodDelete
		testutils.RunTestCase(t, tc)
	}
}
