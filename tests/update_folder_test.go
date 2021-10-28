package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestUpdateFolder(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/folders/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "ParentFolderID",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				      - id: folder-02
				      - id: folder-03
				`)
			},
			Path: makePath(testutils.GetUUID("folder-02")),
			Body: map[string]interface{}{
				"parentFolderId": testutils.GetUUID("folder-01"),
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.GetUUID("folder-02"),
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("folder-01"),
						"type":       "folder",
						"properties": map[string]interface{}{},
					},
				},
				"properties": map[string]interface{}{},
				"createdAt":  testutils.Timestamp{},
				"updatedAt":  testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/folders/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/folders/%s", testutils.GetUUID("folder-03")), nil)
				if len(res["path"].([]interface{})) != 0 {
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
				          - id: folder-02
				          - id: folder-03
				`)
			},
			Path: makePath(testutils.GetUUID("folder-02")),
			Body: map[string]interface{}{
				"parentFolderId": "00000000-0000-0000-0000-000000000000",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.GetUUID("folder-02"),
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path":           []interface{}{},
				"properties":     map[string]interface{}{},
				"createdAt":      testutils.Timestamp{},
				"updatedAt":      testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/folders/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/folders/%s", testutils.GetUUID("folder-03")), nil)
				if len(res["path"].([]interface{})) != 1 {
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
			Path: makePath(testutils.GetUUID("folder-01")),
			Body: map[string]interface{}{
				"name": "new-folder",
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
				      - id: folder-01
				`)
			},
			Path: makePath(testutils.GetUUID("folder-01")),
			Body: map[string]interface{}{
				"parentFolderId": testutils.GetUUID("folder-02"),
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
				      - id: folder-01
				  - id: org2
				    tables:
				      - id: folder-02
				`)
			},
			Path: makePath(testutils.GetUUID("folder-01")),
			Body: map[string]interface{}{
				"parentFolderId": testutils.GetUUID("folder-02"),
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": "Cannot move to another organization",
			},
		},
		{
			Title: "Path loop",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				        children:
				          - id: folder-02
				`)
			},
			Path: makePath(testutils.GetUUID("folder-01")),
			Body: map[string]interface{}{
				"parentFolderId": testutils.GetUUID("folder-02"),
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": "Cannot move to sub folder",
			},
		},
		{
			Title: "Properties",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				        properties:
				          key1: value1
				          key2: value2
				          key3: value3
				`)
			},
			Path: makePath(testutils.GetUUID("folder-01")),
			Body: map[string]interface{}{
				"properties": map[string]interface{}{
					"key1": "new-key",
					"key2": nil,
					"key4": "value4",
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.GetUUID("folder-01"),
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path":           []interface{}{},
				"properties": map[string]interface{}{
					"key1": "new-key",
					"key3": "value3",
					"key4": "value4",
				},
				"createdAt": testutils.Timestamp{},
				"updatedAt": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/folders/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Invalid property key",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				`)
			},
			Path: makePath(testutils.GetUUID("folder-01")),
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
		tc.Method = http.MethodPatch
		testutils.RunTestCase(t, tc)
	}
}
