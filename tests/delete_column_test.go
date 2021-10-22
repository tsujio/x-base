package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestDeleteColumn(t *testing.T) {
	makePath := func(tableID, columnID uuid.UUID) string {
		return fmt.Sprintf("/tables/%s/columns/%s", tableID, columnID)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "General case",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				          - id: column-02
				          - id: column-03
				          - id: column-04
				`)
			},
			Path:       makePath(testutils.GetUUID("table-01"), testutils.GetUUID("column-02")),
			StatusCode: http.StatusOK,
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Check columns
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", testutils.GetUUID("table-01")), nil)
				columnIDs := []uuid.UUID{testutils.GetUUID("column-01"), testutils.GetUUID("column-03"), testutils.GetUUID("column-04")}
				for i, col := range res["columns"].([]interface{}) {
					c := col.(map[string]interface{})
					if c["id"] != columnIDs[i].String() || c["index"] != float64(i) {
						t.Errorf("[%s] Got unexpected columns:\n%s", tc.Title, res["columns"])
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodDelete
		testutils.RunTestCase(t, tc)
	}
}
