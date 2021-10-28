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
				            createdAt: "2021-10-01T00:00:00Z"
				          - id: folder-03
				            createdAt: "2021-10-02T00:00:00Z"
				          - id: folder-04
				            createdAt: "2021-10-03T00:00:00Z"
				            children:
				              - id: table-02
				      - id: folder-02
				`)
			},
			Path: makePath(testutils.GetUUID("folder-01")),
			Query: url.Values{
				"sort": []string{"type:(folder table),createdAt:asc"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
						"id":             testutils.GetUUID("folder-03"),
						"organizationId": testutils.GetUUID("org1"),
						"type":           "folder",
						"path": []interface{}{
							map[string]interface{}{
								"id":         testutils.GetUUID("folder-01"),
								"type":       "folder",
								"properties": map[string]interface{}{},
							},
							map[string]interface{}{
								"id":         testutils.GetUUID("folder-03"),
								"type":       "folder",
								"properties": map[string]interface{}{},
							},
						},
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":             testutils.GetUUID("folder-04"),
						"organizationId": testutils.GetUUID("org1"),
						"type":           "folder",
						"path": []interface{}{
							map[string]interface{}{
								"id":         testutils.GetUUID("folder-01"),
								"type":       "folder",
								"properties": map[string]interface{}{},
							},
							map[string]interface{}{
								"id":         testutils.GetUUID("folder-04"),
								"type":       "folder",
								"properties": map[string]interface{}{},
							},
						},
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":             testutils.GetUUID("table-01"),
						"organizationId": testutils.GetUUID("org1"),
						"type":           "table",
						"path": []interface{}{
							map[string]interface{}{
								"id":         testutils.GetUUID("folder-01"),
								"type":       "folder",
								"properties": map[string]interface{}{},
							},
							map[string]interface{}{
								"id":         testutils.GetUUID("table-01"),
								"type":       "table",
								"properties": map[string]interface{}{},
							},
						},
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
				"totalCount": float64(3),
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
				"sort":           []string{"type:(folder table)"},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
						"id":             testutils.GetUUID("folder-01"),
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
					map[string]interface{}{
						"id":             testutils.GetUUID("table-01"),
						"organizationId": testutils.GetUUID("org1"),
						"type":           "table",
						"path": []interface{}{
							map[string]interface{}{
								"id":         testutils.GetUUID("table-01"),
								"type":       "table",
								"properties": map[string]interface{}{},
							},
						},
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
				"totalCount": float64(2),
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
				"sort":           []string{"type:(folder table)"},
				"page":           []string{"1"},
				"pageSize":       []string{"1"},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
						"id":             testutils.GetUUID("folder-01"),
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
				},
				"totalCount": float64(2),
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
				"sort":           []string{"type:(folder table)"},
				"page":           []string{"2"},
				"pageSize":       []string{"1"},
			},
			Path:       makePath(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
						"id":             testutils.GetUUID("table-01"),
						"organizationId": testutils.GetUUID("org1"),
						"type":           "table",
						"path": []interface{}{
							map[string]interface{}{
								"id":         testutils.GetUUID("table-01"),
								"type":       "table",
								"properties": map[string]interface{}{},
							},
						},
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
				"totalCount": float64(2),
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
				"children":   []interface{}{},
				"totalCount": float64(0),
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
				"children":   []interface{}{},
				"totalCount": float64(0),
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
						"id":             testutils.GetUUID("table-01"),
						"organizationId": testutils.GetUUID("org1"),
						"type":           "table",
						"path": []interface{}{
							map[string]interface{}{
								"id":         testutils.GetUUID("table-01"),
								"type":       "table",
								"properties": map[string]interface{}{},
							},
						},
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
				"totalCount": float64(1),
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
