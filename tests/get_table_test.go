package tests

import (
	"fmt"
	"net/http"
	"net/url"
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
				            properties:
				              key: value
				        properties:
				          key1: value1
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
				"path":           []interface{}{},
				"columns": []interface{}{
					map[string]interface{}{
						"id":      testutils.GetUUID("column-01"),
						"tableId": testutils.GetUUID("table-01"),
						"index":   float64(0),
						"properties": map[string]interface{}{
							"key": "value",
						},
						"createdAt": testutils.Timestamp{},
						"updatedAt": testutils.Timestamp{},
					},
				},
				"properties": map[string]interface{}{
					"key1": "value1",
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
		{
			Title: "Properties",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder
				        properties:
				          key1: value
				        children:
				          - id: table-01
				            properties:
				              key1: value1
				              key2: value2
				              key3: value3
				            columns:
				              - id: column-01
				                properties:
				                  key2: c2
				`)
			},
			Query: url.Values{
				"properties":       []string{"key1,key2"},
				"columnProperties": []string{"key2,key3"},
			},
			Path:       makePath(testutils.GetUUID("table-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.GetUUID("table-01"),
				"organizationId": testutils.GetUUID("org1"),
				"type":           "table",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.GetUUID("folder"),
						"type": "folder",
						"properties": map[string]interface{}{
							"key1": "value",
							"key2": nil,
						},
					},
				},
				"columns": []interface{}{
					map[string]interface{}{
						"id":      testutils.GetUUID("column-01"),
						"tableId": testutils.GetUUID("table-01"),
						"index":   float64(0),
						"properties": map[string]interface{}{
							"key2": "c2",
							"key3": nil,
						},
						"createdAt": testutils.Timestamp{},
						"updatedAt": testutils.Timestamp{},
					},
				},
				"properties": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
				"createdAt": testutils.Timestamp{},
				"updatedAt": testutils.Timestamp{},
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodGet
		testutils.RunTestCase(t, tc)
	}
}
