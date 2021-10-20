package tests

import (
	"net/http"
	"net/url"
	"testing"

	"gorm.io/gorm"

	"github.com/tsujio/x-base/tests/testutils"
)

func TestGetOrganizationList(t *testing.T) {
	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				    createdAt: 2021-01-01T00:00:00Z
				  - id: organization-02
				    createdAt: 2021-01-02T00:00:00Z
				  - id: organization-03
				    createdAt: 2021-01-03T00:00:00Z
				  - id: organization-04
				    createdAt: 2021-01-04T00:00:00Z
				  - id: organization-05
				    createdAt: 2021-01-05T00:00:00Z
				  `)
			},
			Query: url.Values{
				"page":     []string{"2"},
				"pageSize": []string{"2"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("organization-03"),
						"name":       "organization-03",
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("organization-04"),
						"name":       "organization-04",
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
				},
				"total_count": float64(5),
				"has_next":    true,
			},
		},
		{
			Title: "hasNext=false",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				    createdAt: 2021-01-01T00:00:00Z
				  - id: organization-02
				    createdAt: 2021-01-02T00:00:00Z
			`)
			},
			Query: url.Values{
				"page":     []string{"2"},
				"pageSize": []string{"1"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("organization-02"),
						"name":       "organization-02",
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
				},
				"total_count": float64(2),
				"has_next":    false,
			},
		},
		{
			Title: "Too large page",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				  - id: organization-02
				`)
			},
			Query: url.Values{
				"page":     []string{"3"},
				"pageSize": []string{"1"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{},
				"total_count":   float64(2),
				"has_next":      false,
			},
		},
		{
			Title: "pageSize=0",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				`)
			},
			Query: url.Values{
				"page":     []string{"1"},
				"pageSize": []string{"0"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{},
				"total_count":   float64(1),
				"has_next":      true,
			},
		},
		{
			Title: "page=0",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				  - id: organization-02
				`)
			},
			Query: url.Values{
				"page":     []string{"0"},
				"pageSize": []string{"1"},
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bPage\b`},
			},
		},
		{
			Title: "pageSize=-1",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				  - id: organization-02
				`)
			},
			Query: url.Values{
				"page":     []string{"1"},
				"pageSize": []string{"-1"},
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bPageSize\b`},
			},
		},
	}

	for _, tc := range testCases {
		tc.Path = "/organizations"
		tc.Method = http.MethodGet
		testutils.RunTestCase(t, tc)
	}
}
