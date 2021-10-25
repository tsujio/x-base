package tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestCreateColumn(t *testing.T) {
	makePath := func(tableID uuid.UUID) string {
		return fmt.Sprintf("/tables/%s/columns", tableID)
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
			Title: "Create in empty table",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": "column-01",
				"type": "string",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"table_id":   testutils.GetUUID("table-01"),
				"index":      float64(0),
				"name":       "column-01",
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
					uuid.MustParse(output["id"].(string)),
				})
			},
		},
		{
			Title: "Create in table that has columns",
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": "column-02",
				"type": "string",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"table_id":   testutils.GetUUID("table-01"),
				"index":      float64(1),
				"name":       "column-02",
				"type":       "string",
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-01")), nil)
				if diff := testutils.CompareJson(output, res["columns"].([]interface{})[1]); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-01"),
					uuid.MustParse(output["id"].(string)),
				})
			},
		},
		{
			Title: "With index",
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name":  "column-02",
				"type":  "string",
				"index": 0,
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"table_id":   testutils.GetUUID("table-01"),
				"index":      float64(0),
				"name":       "column-02",
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
					uuid.MustParse(output["id"].(string)),
					testutils.GetUUID("column-01"),
				})
			},
		},
		{
			Title: "Empty name",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": "",
				"type": "string",
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bName\b`},
			},
		},
		{
			Title: "Name length=100",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": strings.Repeat("あ", 100),
				"type": "string",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"table_id":   testutils.GetUUID("table-01"),
				"index":      float64(0),
				"name":       strings.Repeat("あ", 100),
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
					uuid.MustParse(output["id"].(string)),
				})
			},
		},
		{
			Title: "Name length=101",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": strings.Repeat("あ", 101),
				"type": "string",
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bName\b`},
			},
		},
		{
			Title: "Invalid column type",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": "column-01",
				"type": "invalid",
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bType\b`},
			},
		},
		{
			Title: "Negative index",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name":  "column-01",
				"type":  "string",
				"index": -1,
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bIndex\b`},
			},
		},
		{
			Title: "Index=999",
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name":  "column-02",
				"type":  "string",
				"index": 999,
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"table_id":   testutils.GetUUID("table-01"),
				"index":      float64(1),
				"name":       "column-02",
				"type":       "string",
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-01")), nil)
				if diff := testutils.CompareJson(output, res["columns"].([]interface{})[1]); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-01"),
					uuid.MustParse(output["id"].(string)),
				})
			},
		},
		{
			Title: "Index=1000",
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name":  "column-02",
				"type":  "string",
				"index": 1000,
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bIndex\b`},
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": "column-01",
				"type": "string",
			},
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Table not found",
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodPost
		testutils.RunTestCase(t, tc)
	}
}