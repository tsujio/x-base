package tests

import (
	"fmt"
	"net/http"
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
			Path:       makePath(testutils.GetUUID("table-01")),
			Body:       map[string]interface{}{},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"tableId":    testutils.GetUUID("table-01"),
				"index":      float64(0),
				"properties": map[string]interface{}{},
				"createdAt":  testutils.Timestamp{},
				"updatedAt":  testutils.Timestamp{},
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
			Path:       makePath(testutils.GetUUID("table-01")),
			Body:       map[string]interface{}{},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"tableId":    testutils.GetUUID("table-01"),
				"index":      float64(1),
				"properties": map[string]interface{}{},
				"createdAt":  testutils.Timestamp{},
				"updatedAt":  testutils.Timestamp{},
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
				"index": 0,
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"tableId":    testutils.GetUUID("table-01"),
				"index":      float64(0),
				"properties": map[string]interface{}{},
				"createdAt":  testutils.Timestamp{},
				"updatedAt":  testutils.Timestamp{},
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
				"index": 999,
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"tableId":    testutils.GetUUID("table-01"),
				"index":      float64(1),
				"properties": map[string]interface{}{},
				"createdAt":  testutils.Timestamp{},
				"updatedAt":  testutils.Timestamp{},
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
			Path:       makePath(testutils.GetUUID("table-01")),
			Body:       map[string]interface{}{},
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Table not found",
			},
		},
		{
			Title: "Properties",
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
				"properties": map[string]interface{}{
					"key1": "value1",
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":      testutils.UUID{},
				"tableId": testutils.GetUUID("table-01"),
				"index":   float64(0),
				"properties": map[string]interface{}{
					"key1": "value1",
				},
				"createdAt": testutils.Timestamp{},
				"updatedAt": testutils.Timestamp{},
			},
		},
		{
			Title: "Invalid property key",
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
				"properties": map[string]interface{}{
					"prop key": "value1",
				},
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `Invalid property key`},
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodPost
		testutils.RunTestCase(t, tc)
	}
}
