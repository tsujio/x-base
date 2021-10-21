package tests

import (
	"fmt"
	"net/http"
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
				      - id: folder-02
				`)
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{testutils.GetUUID("org1").String()},
			},
			Path:       makePath(testutils.GetUUID("folder-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.GetUUID("folder-01"),
				"organization_id": testutils.GetUUID("org1"),
				"name":            "folder-01",
				"type":            "folder",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.GetUUID("folder-01"),
						"name": "folder-01",
						"type": "folder",
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
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
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{testutils.GetUUID("org1").String()},
			},
			Path:       makePath(testutils.GetUUID("folder-02")),
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Not found",
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodGet
		testutils.RunTestCase(t, tc)
	}
}
