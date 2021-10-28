package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/tsujio/x-base/tests/testutils"
)

func TestCreateFolder(t *testing.T) {
	testCases := []testutils.APITestCase{
		{
			Title: "Create at root",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Body: map[string]interface{}{
				"organizationId": testutils.GetUUID("org1"),
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.UUID{},
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path":           []interface{}{},
				"properties":     map[string]interface{}{},
				"createdAt":      testutils.Timestamp{},
				"updatedAt":      testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/folders/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Create at root by zero id",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Body: map[string]interface{}{
				"organizationId": testutils.GetUUID("org1"),
				"parentFolderId": "00000000-0000-0000-0000-000000000000",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.UUID{},
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path":           []interface{}{},
				"properties":     map[string]interface{}{},
				"createdAt":      testutils.Timestamp{},
				"updatedAt":      testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/folders/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Create at sub folder",
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
			Body: map[string]interface{}{
				"organizationId": testutils.GetUUID("org1"),
				"parentFolderId": testutils.GetUUID("folder-02"),
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.UUID{},
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("folder-01"),
						"type":       "folder",
						"properties": map[string]interface{}{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("folder-02"),
						"type":       "folder",
						"properties": map[string]interface{}{},
					},
				},
				"properties": map[string]interface{}{},
				"createdAt":  testutils.Timestamp{},
				"updatedAt":  testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/folders/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Parent not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Body: map[string]interface{}{
				"organizationId": testutils.GetUUID("org1"),
				"parentFolderId": testutils.GetUUID("folder-02"),
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": "Parent folder not found",
			},
		},
		{
			Title: "Parent's organization mismatch",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				  - id: org2
				    tables:
				      - id: folder-01
				`)
			},
			Body: map[string]interface{}{
				"organizationId": testutils.GetUUID("org1"),
				"parentFolderId": testutils.GetUUID("folder-01"),
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": "Cannot create folder as a child of another organization's folder",
			},
		},
		{
			Title: "Properties",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Body: map[string]interface{}{
				"organizationId": testutils.GetUUID("org1"),
				"properties": map[string]interface{}{
					"key1": "value1",
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.UUID{},
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path":           []interface{}{},
				"properties": map[string]interface{}{
					"key1": "value1",
				},
				"createdAt": testutils.Timestamp{},
				"updatedAt": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
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
				`)
			},
			Body: map[string]interface{}{
				"organizationId": testutils.GetUUID("org1"),
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
		tc.Path = "/folders"
		tc.Method = http.MethodPost
		testutils.RunTestCase(t, tc)
	}
}
