package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestUpdateTable(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/tables/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "Name",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				        children:
				          - id: table-01
				          - id: table-02
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": "new-table",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.GetUUID("table-01"),
				"organization_id": testutils.GetUUID("org1"),
				"name":            "new-table",
				"type":            "table",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.GetUUID("folder-01"),
						"name": "folder-01",
						"type": "folder",
					},
					map[string]interface{}{
						"id":   testutils.GetUUID("table-01"),
						"name": "new-table",
						"type": "table",
					},
				},
				"columns":    []interface{}{},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-02")), nil)
				if res["name"] != "table-02" {
					t.Errorf("[%s] Modified other data:\n%s", tc.Title, res["name"])
				}
			},
		},
		{
			Title: "ParentFolderID",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				      - id: table-01
				      - id: table-02
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"parent_folder_id": testutils.GetUUID("folder-01"),
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.GetUUID("table-01"),
				"organization_id": testutils.GetUUID("org1"),
				"name":            "table-01",
				"type":            "table",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.GetUUID("folder-01"),
						"name": "folder-01",
						"type": "folder",
					},
					map[string]interface{}{
						"id":   testutils.GetUUID("table-01"),
						"name": "table-01",
						"type": "table",
					},
				},
				"columns":    []interface{}{},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-02")), nil)
				if len(res["path"].([]interface{})) != 1 {
					t.Errorf("[%s] Modified other data:\n%s", tc.Title, res["path"])
				}
			},
		},
		{
			Title: "Move to root folder",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				        children:
				          - id: table-01
				          - id: table-02
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"parent_folder_id": "00000000-0000-0000-0000-000000000000",
			},
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
				"columns":    []interface{}{},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-02")), nil)
				if len(res["path"].([]interface{})) != 2 {
					t.Errorf("[%s] Modified other data:\n%s", tc.Title, res["path"])
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
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"name": "new-table",
			},
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Not found",
			},
		},
		{
			Title: "Parent not found",
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
				"parent_folder_id": testutils.GetUUID("folder-01"),
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": "Destination folder not found",
			},
		},
		{
			Title: "Parent's organization mismatch",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				  - id: org2
				    tables:
				      - id: folder-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: map[string]interface{}{
				"parent_folder_id": testutils.GetUUID("folder-01"),
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": "Cannot move to another organization",
			},
		},
		{
			Title: "No update",
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
				"columns": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("column-01"),
						"table_id":   testutils.GetUUID("table-01"),
						"index":      float64(0),
						"name":       "column-01",
						"type":       "string",
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodPatch
		testutils.RunTestCase(t, tc)
	}
}
