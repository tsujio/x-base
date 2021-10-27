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
				        columns:
				          - id: column-01
				      - id: table-02
				        columns:
				          - id: column-02
				`)
			},
			Path:       makePath(testutils.GetUUID("table-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.GetUUID("table-01"),
				"organizationId": testutils.GetUUID("org1"),
				"type":           "table",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.GetUUID("table-01"),
						"type": "table",
					},
				},
				"columns": []interface{}{
					map[string]interface{}{
						"id":        testutils.GetUUID("column-01"),
						"tableId":   testutils.GetUUID("table-01"),
						"index":     float64(0),
						"createdAt": testutils.Timestamp{},
						"updatedAt": testutils.Timestamp{},
					},
				},
				"createdAt": testutils.Timestamp{},
				"updatedAt": testutils.Timestamp{},
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
				`)
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
