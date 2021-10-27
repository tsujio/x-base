package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestUpdateOrganization(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/organizations/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "No update",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				  - id: organization-02
				`)
			},
			Path:       makePath(testutils.GetUUID("organization-01")),
			Body:       map[string]interface{}{},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":        testutils.GetUUID("organization-01"),
				"createdAt": testutils.Timestamp{},
				"updatedAt": testutils.Timestamp{},
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
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				`)
			},
			Path:       makePath(testutils.GetUUID("organization-02")),
			Body:       map[string]interface{}{},
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
