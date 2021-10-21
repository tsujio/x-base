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

func TestGetFolderChildren(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/folders/%s/children", id)
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
				        children:
				          - id: table-01
				          - id: folder-03
				          - id: folder-04
				            children:
				              - id: table-02
				      - id: folder-02
				`)
			},
			Path:       makePath(testutils.GetUUID("folder-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
						"id":              testutils.GetUUID("folder-03"),
						"organization_id": testutils.GetUUID("org1"),
						"name":            "folder-03",
						"type":            "folder",
						"path": []interface{}{
							map[string]interface{}{
								"id":   testutils.GetUUID("folder-01"),
								"name": "folder-01",
								"type": "folder",
							},
							map[string]interface{}{
								"id":   testutils.GetUUID("folder-03"),
								"name": "folder-03",
								"type": "folder",
							},
						},
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":              testutils.GetUUID("folder-04"),
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
								"id":   testutils.GetUUID("folder-04"),
								"name": "folder-04",
								"type": "folder",
							},
						},
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
					map[string]interface{}{
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
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
				},
				"total_count": float64(3),
				"has_next":    false,
			},
		},
		{
			Title: "Root folder",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				      - id: folder-01
				        children:
				          - id: table-02
				  - id: org2
				    tables:
				      - id: table-03
				`)
			},
			Query: url.Values{
				"organizationId": []string{testutils.GetUUID("org1").String()},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
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
					map[string]interface{}{
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
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
				},
				"total_count": float64(2),
				"has_next":    false,
			},
		},
		{
			Title: "Root folder without organizationId",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": "Organization id is required for root folder",
			},
		},
		{
			Title: "Fetche folders only",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				      - id: folder-01
				`)
			},
			Query: url.Values{
				"organizationId": []string{testutils.GetUUID("org1").String()},
				"page":           []string{"1"},
				"pageSize":       []string{"1"},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
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
				"total_count": float64(2),
				"has_next":    true,
			},
		},
		{
			Title: "Fetche tables only",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				      - id: folder-01
				`)
			},
			Query: url.Values{
				"organizationId": []string{testutils.GetUUID("org1").String()},
				"page":           []string{"2"},
				"pageSize":       []string{"1"},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
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
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
				},
				"total_count": float64(2),
				"has_next":    false,
			},
		},
		{
			Title: "Empty folder",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: folder-01
				`)
			},
			Path:       makePath(testutils.GetUUID("folder-01")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children":    []interface{}{},
				"total_count": float64(0),
				"has_next":    false,
			},
		},
		{
			Title: "Empty root",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				`)
			},
			Query: url.Values{
				"organizationId": []string{testutils.GetUUID("org1").String()},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children":    []interface{}{},
				"total_count": float64(0),
				"has_next":    false,
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
			Path:       makePath(testutils.GetUUID("folder-01")),
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Not found",
			},
		},
		{
			Title: "pageSize=100",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Query: url.Values{
				"organizationId": []string{testutils.GetUUID("org1").String()},
				"pageSize":       []string{"100"},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
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
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
				},
				"total_count": float64(1),
				"has_next":    false,
			},
		},
		{
			Title: "pageSize=101",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Query: url.Values{
				"organizationId": []string{testutils.GetUUID("org1").String()},
				"pageSize":       []string{"101"},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bPageSize\b`},
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodGet
		testutils.RunTestCase(t, tc)
	}
}
