package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestReorderColumn(t *testing.T) {
	makePath := func(tableID uuid.UUID) string {
		return fmt.Sprintf("/tables/%s/columns/reorder", tableID)
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
				      - id: table-02
				        columns:
				          - id: column-04
				          - id: column-05
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"order": []interface{}{
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-02"),
					testutils.GetUUID("column-01"),
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"columns": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("column-03"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(0),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-02"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(1),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-01"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(2),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-02"),
					testutils.GetUUID("column-01"),
				})
				testColumnOrder(tc, router, testutils.GetUUID("table-02"), []uuid.UUID{
					testutils.GetUUID("column-04"),
					testutils.GetUUID("column-05"),
				})
			},
		},
		{
			Title: "Contains unknown column id",
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"order": []interface{}{
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-02"),
					testutils.GetUUID("column-01"),
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"columns": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("column-02"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(0),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-01"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(1),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-02"),
					testutils.GetUUID("column-01"),
				})
			},
		},
		{
			Title: "Missing some ids",
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"order": []interface{}{
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-01"),
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"columns": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("column-03"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(0),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-01"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(1),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-02"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(2),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-01"),
					testutils.GetUUID("column-02"),
				})
			},
		},
		{
			Title: "Contains other table's columns",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				          - id: column-02
				      - id: table-02
				        columns:
				          - id: column-03
				          - id: column-04
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"order": []interface{}{
					testutils.GetUUID("column-02"),
					testutils.GetUUID("column-01"),
					testutils.GetUUID("column-03"),
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"columns": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("column-02"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(0),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-01"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(1),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-02"),
					testutils.GetUUID("column-01"),
				})
				testColumnOrder(tc, router, testutils.GetUUID("table-02"), []uuid.UUID{
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-04"),
				})
			},
		},
		{
			Title: "Many columns",
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
				          - id: column-05
				          - id: column-06
				          - id: column-07
				          - id: column-08
				          - id: column-09
				          - id: column-10
				          - id: column-11
				          - id: column-12
				          - id: column-13
				          - id: column-14
				          - id: column-15
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"order": []interface{}{
					testutils.GetUUID("column-15"),
					testutils.GetUUID("column-14"),
					testutils.GetUUID("column-13"),
					testutils.GetUUID("column-12"),
					testutils.GetUUID("column-11"),
					testutils.GetUUID("column-10"),
					testutils.GetUUID("column-09"),
					testutils.GetUUID("column-08"),
					testutils.GetUUID("column-07"),
					testutils.GetUUID("column-06"),
					testutils.GetUUID("column-05"),
					testutils.GetUUID("column-04"),
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-02"),
					testutils.GetUUID("column-01"),
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"columns": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("column-15"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(0),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-14"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(1),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-13"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(2),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-12"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(3),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-11"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(4),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-10"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(5),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-09"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(6),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-08"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(7),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-07"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(8),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-06"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(9),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-05"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(10),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-04"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(11),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-03"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(12),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-02"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(13),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("column-01"),
						"tableId":    testutils.GetUUID("table-01"),
						"index":      float64(14),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check columns
				testColumnOrder(tc, router, testutils.GetUUID("table-01"), []uuid.UUID{
					testutils.GetUUID("column-15"),
					testutils.GetUUID("column-14"),
					testutils.GetUUID("column-13"),
					testutils.GetUUID("column-12"),
					testutils.GetUUID("column-11"),
					testutils.GetUUID("column-10"),
					testutils.GetUUID("column-09"),
					testutils.GetUUID("column-08"),
					testutils.GetUUID("column-07"),
					testutils.GetUUID("column-06"),
					testutils.GetUUID("column-05"),
					testutils.GetUUID("column-04"),
					testutils.GetUUID("column-03"),
					testutils.GetUUID("column-02"),
					testutils.GetUUID("column-01"),
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"order": []interface{}{},
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
