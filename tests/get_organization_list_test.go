package tests

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestGetOrganizationList(t *testing.T) {
	createOrganizations := func(n int) ([]models.Organization, error) {
		var organizations []models.Organization
		for i := 0; i < n; i++ {
			o := models.Organization{
				Name:      fmt.Sprintf("organization-%02d", i+1),
				CreatedAt: time.Now().AddDate(0, 0, i),
			}
			err := o.Create(testutils.GetDB())
			if err != nil {
				return nil, err
			}
			organizations = append(organizations, o)
		}
		return organizations, nil
	}

	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(5)
				return err
			},
			Query: url.Values{
				"page":     []string{"2"},
				"pageSize": []string{"2"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{
					map[string]interface{}{
						"id":         testutils.UUID{},
						"name":       "organization-03",
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
					map[string]interface{}{
						"id":         testutils.UUID{},
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
				_, err := createOrganizations(5)
				return err
			},
			Query: url.Values{
				"page":     []string{"5"},
				"pageSize": []string{"1"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{
					map[string]interface{}{
						"id":         testutils.UUID{},
						"name":       "organization-05",
						"created_at": testutils.Timestamp{},
						"updated_at": testutils.Timestamp{},
					},
				},
				"total_count": float64(5),
				"has_next":    false,
			},
		},
		{
			Title: "Too large page",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(5)
				return err
			},
			Query: url.Values{
				"page":     []string{"6"},
				"pageSize": []string{"1"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{},
				"total_count":   float64(5),
				"has_next":      false,
			},
		},
		{
			Title: "pageSize=0",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(5)
				return err
			},
			Query: url.Values{
				"page":     []string{"1"},
				"pageSize": []string{"0"},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"organizations": []interface{}{},
				"total_count":   float64(5),
				"has_next":      true,
			},
		},
		{
			Title: "page=0",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				_, err := createOrganizations(5)
				return err
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
				_, err := createOrganizations(5)
				return err
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
