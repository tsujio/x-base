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
			Path:       makePath(testutils.GetUUID("folder-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":             testutils.GetUUID("folder-01"),
				"organizationId": testutils.GetUUID("org1"),
				"type":           "folder",
				"path":           []interface{}{},
				"properties":     map[string]interface{}{},
				"createdAt":      testutils.Timestamp{},
				"updatedAt":      testutils.Timestamp{},
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
	}

	for _, tc := range testCases {
		tc.Method = http.MethodGet
		testutils.RunTestCase(t, tc)
	}
}
