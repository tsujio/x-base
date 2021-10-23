package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestUpdateColumn(t *testing.T) {
	makePath := func(tableID, columnID uuid.UUID) string {
		return fmt.Sprintf("/tables/%s/columns/%s", tableID, columnID)
	}

	testColumnOrder := func(tc *testutils.APITestCase, router http.Handler, tableID uuid.UUID, columnIDs []uuid.UUID) {
		res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", tableID), nil)
		if len(columnIDs) != len(res["columns"].([]interface{})) {
			t.Errorf("[%s] # of Columns mismatch:\n%s", tc.Title, res["columns"])
		}
		for i, col := range res["columns"].([]interface{}) {
			c := col.(map[string]interface{})
			if c["id"] != columnIDs[i].String() || c["index"] != float64(i) {
				t.Errorf("[%s] Got unexpected columns:\n%s", tc.Title, res["columns"])
			}
		}
	}

	testCases := []testutils.APITestCase{
		{
			Title: "Name",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				          - id: column-02
				`)
			},
			Path: makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-01")),
			Body: map[string]interface{}{
				"name": "new-column",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.GetUUID("column-01"),
				"table_id":   testutils.GetUUID("table-01"),
				"index":      float64(0),
				"name":       "new-column",
				"type":       "string",
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-01")), nil)
				if diff := testutils.CompareJson(output, res["columns"].([]interface{})[0]); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-01"),
					testutils.GetUUID("column-02"),
				})
			},
		},
		{
			Title: "Move to tail",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				          - id: column-02
				          - id: column-03
				`)
			},
			Path: makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-02")),
			Body: map[string]interface{}{
				"index": 3,
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.GetUUID("column-02"),
				"table_id":   testutils.GetUUID("table-01"),
				"index":      float64(2),
				"name":       "column-02",
				"type":       "string",
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-01")), nil)
				if diff := testutils.CompareJson(output, res["columns"].([]interface{})[2]); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-01"),
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-02"),
				})
			},
		},
		{
			Title: "Table not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Path: makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-01")),
			Body: map[string]interface{}{
				"name": "new-column",
			},
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Table not found",
			},
		},
		{
			Title: "Table exists but column not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-01")),
			Body: map[string]interface{}{
				"name": "new-column",
			},
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Column not found",
			},
		},
		{
			Title: "Specify other table's column",
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
			Path: makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-02")),
			Body: map[string]interface{}{
				"name": "new-column",
			},
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Column not found",
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check other columns not affected
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-02")), nil)
				if res["columns"].([]interface{})[0].(map[string]interface{})["name"] != "column-02" {
					t.Errorf("[%s] Modified other table's column:\n%s", tc.Title, res["columns"])
				}
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodPatch
		testutils.RunTestCase(t, tc)
	}
}
