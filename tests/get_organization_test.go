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

func TestGetOrganization(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/organizations/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				  - id: organization-02
				`)
			},
			Path:       makePath(testutils.GetUUID("organization-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.GetUUID("organization-01"),
				"properties": map[string]interface{}{},
				"createdAt":  testutils.Timestamp{},
				"updatedAt":  testutils.Timestamp{},
			},
		},
		{
			Title: "Properties",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				    properties:
				      key1: value1
				      key2: value2
				      key3: value3
				  - id: organization-02
				`)
			},
			Query: url.Values{
				"properties": []string{"key1,key2,key4"},
			},
			Path:       makePath(testutils.GetUUID("organization-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id": testutils.GetUUID("organization-01"),
				"properties": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key4": nil,
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
				  - id: organization-01
				  - id: organization-02
				`)
			},
			Path:       makePath(testutils.GetUUID("organization-03")),
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
