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

func TestDeleteTable(t *testing.T) {
	var uuids []uuid.UUID
	for i := 0; i < 10; i++ {
		uuids = append(uuids, uuid.New())
	}

	createOrganizations := func(n int) error {
		for i := 0; i < n; i++ {
			o := models.Organization{
				ID:   models.UUID(uuids[i]),
				Name: fmt.Sprintf("organization-%02d", i+1),
			}
			if err := o.Create(testutils.GetDB()); err != nil {
				return err
			}
		}
		return nil
	}

	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/tables/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				if err := createOrganizations(1); err != nil {
					return err
				}
				for i := 0; i < 2; i++ {
					o := models.Table{
						TableFilesystemEntry: models.TableFilesystemEntry{
							ID:             models.UUID(uuids[1+i]),
							OrganizationID: models.UUID(uuids[0]),
							Name:           fmt.Sprintf("table-%02d", i+1),
						},
					}
					if err := o.Create(testutils.GetDB()); err != nil {
						return err
					}
				}
				return nil
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
			},
			Path:       makePath(uuids[1]),
			StatusCode: http.StatusOK,
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire
				res := testutils.ServeGet(router, makePath(uuids[1]), nil)
				if res != nil {
					t.Errorf("[%s] Not deleted", tc.Title)
				}

				// Did not delete other data
				res = testutils.ServeGet(router, makePath(uuids[2]), nil)
				if res == nil {
					t.Errorf("[%s] Deleted other data", tc.Title)
				}
			},
		},
		{
			Title: "Not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				if err := createOrganizations(1); err != nil {
					return err
				}
				for i := 0; i < 2; i++ {
					o := models.Table{
						TableFilesystemEntry: models.TableFilesystemEntry{
							ID:             models.UUID(uuids[1+i]),
							OrganizationID: models.UUID(uuids[0]),
							Name:           fmt.Sprintf("table-%02d", i+1),
						},
					}
					if err := o.Create(testutils.GetDB()); err != nil {
						return err
					}
				}
				return nil
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
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
