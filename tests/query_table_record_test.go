package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/ghodss/yaml"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/tests/testutils"
)

func makeJSON(ymlTmpl string, params map[string]interface{}) map[string]interface{} {
	tmpl, err := template.New("").Parse(ymlTmpl)
	if err != nil {
		panic(err)
	}
	var yml bytes.Buffer
	if err := tmpl.Execute(&yml, params); err != nil {
		panic(err)
	}
	var q map[string]interface{}
	if err := yaml.Unmarshal([]byte(testutils.Dedent(yml.String())), &q); err != nil {
		panic(err)
	}
	return q
}

func selectTable(router http.Handler, tableID uuid.UUID, query map[string]interface{}) map[string]interface{} {
	body, err := json.Marshal(&query)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/tables/%s/query", tableID), bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}

	r := httptest.NewRecorder()
	router.ServeHTTP(r, req)

	switch r.Code {
	case http.StatusOK:
		var result map[string]interface{}
		err = json.Unmarshal(r.Body.Bytes(), &result)
		if err != nil {
			log.Fatal(err)
		}
		return result
	case http.StatusNotFound:
		return nil
	default:
		log.Fatal(r.Code, " ", r.Body.String())
		return nil
	}
}

func TestQueryTableRecordInsert(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/tables/%s/query", id)
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
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			insert:
			  columns:
			    - column: {{ .column01 }}
			    - column: {{ .column02 }}
			  values:
			    - - value: v1-1
			      - value: v1-2
			    - - value: v2-1
			      - value: null
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
				"column02": testutils.GetUUID("column-02"),
			}),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"record_ids": []interface{}{
					testutils.UUID{},
					testutils.UUID{},
				},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Select from the table
				res := selectTable(router, testutils.GetUUID("table-01"), makeJSON(`
				select:
				  columns: [{metadata: id}, {column: {{ .column01 }} }, {column: {{ .column02 }} }]
				  order_by: [{key: {column: {{ .column01 }} }}]
				`, map[string]interface{}{
					"column01": testutils.GetUUID("column-01"),
					"column02": testutils.GetUUID("column-02"),
				}))
				if diff := testutils.CompareJson(res, map[string]interface{}{
					"records": []interface{}{
						[]interface{}{output["record_ids"].([]interface{})[0], "v1-1", "v1-2"},
						[]interface{}{output["record_ids"].([]interface{})[1], "v2-1", nil},
					},
				}); diff != "" {
					t.Errorf("[%s] Selected records mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				`)
			},
			Path: makePath(testutils.GetUUID("table-02")),
			Body: makeJSON(`
			insert:
			  columns:
			    - column: {{ .column01 }}
			    - column: {{ .column02 }}
			  values:
			    - - value: v1-1
			      - value: v1-2
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
				"column02": testutils.GetUUID("column-02"),
			}),
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Table not found",
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodPost
		testutils.RunTestCase(t, tc)
	}
}
