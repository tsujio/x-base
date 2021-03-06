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
				"sort":     []string{"createdAt:asc"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("organization-03"),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.GetUUID("organization-04"),
						"properties": map[string]interface{}{},
						"createdAt":  testutils.Timestamp{},
						"updatedAt":  testutils.Timestamp{},
					},
				},
				"totalCount": float64(5),
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
				"page": []string{"3"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{},
				"totalCount":    float64(2),
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
				"pageSize": []string{"0"},
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bPageSize\b`},
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
				"page": []string{"0"},
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
				"pageSize": []string{"-1"},
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bPageSize\b`},
			},
		},
		{
			Title: "pageSize=100",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				`)
			},
			Query: url.Values{
				"pageSize": []string{"100"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{
					map[string]interface{}{
						"id":         testutils.GetUUID("organization-01"),
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
				  - id: organization-01
				`)
			},
			Query: url.Values{
				"pageSize": []string{"101"},
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bPageSize\b`},
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
				`)
			},
			Query: url.Values{
				"properties": []string{"key1,key2,key4"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{
					map[string]interface{}{
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
				"totalCount": float64(1),
			},
		},
		{
			Title: "Sort by property",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: organization-01
				    properties:
				      key1: 1
				  - id: organization-02
				    properties:
				      key1: 2
				`)
			},
			Query: url.Values{
				"sort": []string{"property.key1:desc"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{
					map[string]interface{}{
						"id": testutils.GetUUID("organization-02"),
						"properties": map[string]interface{}{
							"key1": float64(2),
						},
						"createdAt": testutils.Timestamp{},
						"updatedAt": testutils.Timestamp{},
					},
					map[string]interface{}{
						"id": testutils.GetUUID("organization-01"),
						"properties": map[string]interface{}{
							"key1": float64(1),
						},
						"createdAt": testutils.Timestamp{},
						"updatedAt": testutils.Timestamp{},
					},
				},
				"totalCount": float64(2),
			},
		},
	}

	for _, tc := range testCases {
		tc.Path = "/organizations"
		tc.Method = http.MethodGet
		testutils.RunTestCase(t, tc)
	}
}
