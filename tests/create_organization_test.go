package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/tsujio/x-base/tests/testutils"
)

func TestCreateOrganization(t *testing.T) {
	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Body: map[string]interface{}{
				"properties": map[string]interface{}{
					"name": "organization-01",
				},
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id": testutils.UUID{},
				"properties": map[string]interface{}{
					"name": "organization-01",
				},
				"createdAt": testutils.Timestamp{},
				"updatedAt": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title:      "Empty property",
			Body:       map[string]interface{}{},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":         testutils.UUID{},
				"properties": map[string]interface{}{},
				"createdAt":  testutils.Timestamp{},
				"updatedAt":  testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/organizations/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Invalid property key",
			Body: map[string]interface{}{
				"properties": map[string]interface{}{
					"prop key": "value",
				},
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `Invalid property key`},
			},
		},
	}

	for _, tc := range testCases {
		tc.Path = "/organizations"
		tc.Method = http.MethodPost
		testutils.RunTestCase(t, tc)
	}
}
