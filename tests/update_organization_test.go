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

func TestUpdateOrganization(t *testing.T) {
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
			Title: "Update name",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(2)
				return err
			},
			Path: makePath(uuids[0]),
			Body: map[string]interface{}{
				"name": "new-organization",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         uuids[0].String(),
				"name":       "new-organization",
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", uuids[1]), nil)
				if res["name"] != "organization-02" {
					t.Errorf("[%s] Modified other data:\n%s", tc.Title, res["name"])
				}
			},
		},
		{
			Title: "Empty name",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(2)
				return err
			},
			Path: makePath(uuids[0]),
			Body: map[string]interface{}{
				"name": "",
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bName\b`},
			},
		},
		{
			Title: "No update",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(2)
				return err
			},
			Path:       makePath(uuids[0]),
			Body:       map[string]interface{}{},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         uuids[0].String(),
				"name":       "organization-01",
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(2)
				return err
			},
			Path: makePath(uuids[3]),
			Body: map[string]interface{}{
				"name": "new-organization",
			},
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Not found",
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodPatch
		testutils.RunTestCase(t, tc)
	}
}
