package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestDeleteTable(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/tables/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        type: table
				      - id: table-02
				        type: table
				`)
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{testutils.GetUUID("org1").String()},
			},
			Path:       makePath(testutils.GetUUID("table-01")),
			StatusCode: http.StatusOK,
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire
				res := testutils.ServeGet(router, makePath(testutils.GetUUID("table-01")), nil)
				if res != nil {
					t.Errorf("[%s] Not deleted", tc.Title)
				}

				// Did not delete other data
				res = testutils.ServeGet(router, makePath(testutils.GetUUID("table-02")), nil)
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
				    tables:
				      - id: table-01
				        type: table
				`)
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{testutils.GetUUID("org1").String()},
			},
			Path:       makePath(testutils.GetUUID("table-02")),
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
