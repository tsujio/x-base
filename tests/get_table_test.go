package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestGetTable(t *testing.T) {
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
			Output: map[string]interface{}{
				"id":              testutils.GetUUID("table-01"),
				"organization_id": testutils.GetUUID("org1"),
				"name":            "table-01",
				"type":            "table",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.GetUUID("table-01"),
						"name": "table-01",
						"type": "table",
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
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
		tc.Method = http.MethodGet
		testutils.RunTestCase(t, tc)
	}
}
