package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestDeleteOrganization(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/organizations/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "Delete",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				  - id: org2
				`)
			},
			Path:       makePath(testutils.GetUUID("org1")),
			StatusCode: http.StatusOK,
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire
				res := testutils.ServeGet(router, makePath(testutils.GetUUID("org1")), nil)
				if res != nil {
					t.Errorf("[%s] Not deleted", tc.Title)
				}

				// Did not delete other data
				res = testutils.ServeGet(router, makePath(testutils.GetUUID("org2")), nil)
				if res == nil {
					t.Errorf("[%s] Deleted other data", tc.Title)
				}
			},
		},
		{
			Title: "Not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Path:       makePath(testutils.GetUUID("org2")),
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
