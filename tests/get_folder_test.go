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

func TestGetFolder(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/folders/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				        properties:
				          key1: value1
				  - id: folder-02
				`)
			},
			Path:       makePath(testutils.GetUUID("folder-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.GetUUID("folder-01"),
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path":           []interface{}{},
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
				      - id: folder-01
				`)
			},
			Path:       makePath(testutils.GetUUID("folder-02")),
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
				          - id: folder-01
				            properties:
				              key1: value1
				              key2: value2
				              key3: value3
				`)
			},
			Query: url.Values{
				"properties": []string{"key1,key2"},
			},
			Path:       makePath(testutils.GetUUID("folder-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.GetUUID("folder-01"),
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
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
