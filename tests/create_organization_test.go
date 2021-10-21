package tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/tsujio/x-base/tests/testutils"
)

func TestCreateOrganization(t *testing.T) {
	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Body: map[string]interface{}{
				"name": "organization-01",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"name":       "organization-01",
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Empty name",
			Body: map[string]interface{}{
				"name": "",
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bName\b`},
			},
		},
		{
			Title: "Name length=100",
			Body: map[string]interface{}{
				"name": strings.Repeat("あ", 100),
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"name":       strings.Repeat("あ", 100),
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Name length=101",
			Body: map[string]interface{}{
				"name": strings.Repeat("あ", 101),
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bName\b`},
			},
		},
	}

	for _, tc := range testCases {
		tc.Path = "/organizations"
		tc.Method = http.MethodPost
		testutils.RunTestCase(t, tc)
	}
}
