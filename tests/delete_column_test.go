package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestDeleteColumn(t *testing.T) {
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
			Title: "General case",
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
				          - id: column-04
				`)
			},
			Path:       makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-02")),
			StatusCode: http.StatusOK,
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-01"),
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-04"),
				})
			},
		},
		{
			Title: "Table not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-02
				        columns:
				          - id: column-01
				`)
			},
			Path:       makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-01")),
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Table not found",
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check other table's columns not affected
				testColumnOrder(tc, router, testutils.GetUUID("table-02"), []uuid.UUID{
					testutils.GetUUID("column-01"),
				})
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
				        columns:
				          - id: column-01
				`)
			},
			Path:       makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-02")),
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Column not found",
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check other columns not affected
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-01"),
				})
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
			Path:       makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-02")),
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Column not found",
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-01"),
				})

				// Check other columns not affected
				testColumnOrder(tc, router, testutils.GetUUID("table-02"), []uuid.UUID{
					testutils.GetUUID("column-02"),
				})
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodDelete
		testutils.RunTestCase(t, tc)
	}
}
