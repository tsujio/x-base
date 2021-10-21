package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/tsujio/x-base/tests/testutils"
)

func TestFolderTable(t *testing.T) {
	testCases := []testutils.APITestCase{
		{
			Title: "Create at root",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{testutils.GetUUID("org1").String()},
			},
			Body: map[string]interface{}{
				"name": "folder-01",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.UUID{},
				"organization_id": testutils.GetUUID("org1"),
				"name":            "folder-01",
				"type":            "folder",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.UUID{},
						"name": "folder-01",
						"type": "folder",
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
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
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{testutils.GetUUID("org1").String()},
			},
			Body: map[string]interface{}{
				"name":             "folder-01",
				"parent_folder_id": "00000000-0000-0000-0000-000000000000",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.UUID{},
				"organization_id": testutils.GetUUID("org1"),
				"name":            "folder-01",
				"type":            "folder",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.UUID{},
						"name": "folder-01",
						"type": "folder",
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
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
				        type: folder
				        children:
				          - id: folder-02
				            type: folder
				          - id: folder-03
				            type: folder				        
				`)
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{testutils.GetUUID("org1").String()},
			},
			Body: map[string]interface{}{
				"name":             "folder-04",
				"parent_folder_id": testutils.GetUUID("folder-02"),
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.UUID{},
				"organization_id": testutils.GetUUID("org1"),
				"name":            "folder-04",
				"type":            "folder",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.GetUUID("folder-01"),
						"name": "folder-01",
						"type": "folder",
					},
					map[string]interface{}{
						"id":   testutils.GetUUID("folder-02"),
						"name": "folder-02",
						"type": "folder",
					},
					map[string]interface{}{
						"id":   testutils.UUID{},
						"name": "folder-04",
						"type": "folder",
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/folders/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Empty name",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{testutils.GetUUID("org1").String()},
			},
			Body: map[string]interface{}{
				"name": "",
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bName\b`},
			},
		},
	}

	for _, tc := range testCases {
		tc.Path = "/folders"
		tc.Method = http.MethodPost
		testutils.RunTestCase(t, tc)
	}
}
