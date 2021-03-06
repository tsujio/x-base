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
				            type: string
				          - id: column-02
				            type: integer
				          - id: column-03
				            type: float
				          - id: column-04
				            type: boolean
				          - id: column-05
				            type: integer
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			insert:
			  columns:
			    - column: {{ .column01 }}
			    - column: {{ .column02 }}
			    - column: {{ .column03 }}
			    - column: {{ .column04 }}
			    - column: {{ .column05 }}
			    - property: key1
			  values:
			    - - value: v1-1
			      - value: 0
			      - value: 3.14
			      - value: true
			      - value: 1
			      - value: [1, 2, 3]
			    - - value: v2-1
			      - value: 1
			      - value: 2.71
			      - value: false
			      - value: 2
			      - value: {key: value}
			    - - value: null
			      - value: null
			      - value: null
			      - value: null
			      - value: 3
			      - value: [{k: v1}, {k: v2}]
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
				"column02": testutils.GetUUID("column-02"),
				"column03": testutils.GetUUID("column-03"),
				"column04": testutils.GetUUID("column-04"),
				"column05": testutils.GetUUID("column-05"),
			}),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"recordIds": []interface{}{
					testutils.UUID{},
					testutils.UUID{},
					testutils.UUID{},
				},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Select from the table
				res := selectTable(router, testutils.GetUUID("table-01"), makeJSON(`
				select:
				  columns:
				    - metadata: id
				    - column: {{ .column01 }}
				    - column: {{ .column02 }}
				    - column: {{ .column03 }}
				    - column: {{ .column04 }}
				    - property: key1
				    - property: key2
				  orderBy: [{key: {column: {{ .column05 }} }}]
				`, map[string]interface{}{
					"column01": testutils.GetUUID("column-01"),
					"column02": testutils.GetUUID("column-02"),
					"column03": testutils.GetUUID("column-03"),
					"column04": testutils.GetUUID("column-04"),
					"column05": testutils.GetUUID("column-05"),
				}))
				if diff := testutils.CompareJson(map[string]interface{}{
					"records": []interface{}{
						[]interface{}{
							output["recordIds"].([]interface{})[0],
							"v1-1",
							float64(0),
							float64(3.14),
							true,
							[]interface{}{float64(1), float64(2), float64(3)},
							nil,
						},
						[]interface{}{
							output["recordIds"].([]interface{})[1],
							"v2-1",
							float64(1),
							float64(2.71),
							false,
							map[string]interface{}{
								"key": "value",
							},
							nil,
						},
						[]interface{}{
							output["recordIds"].([]interface{})[2],
							nil,
							nil,
							nil,
							nil,
							[]interface{}{
								map[string]interface{}{
									"k": "v1",
								},
								map[string]interface{}{
									"k": "v2",
								},
							},
							nil,
						},
					},
					"limit": float64(10),
				}, res); diff != "" {
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

func TestQueryTableRecordSelect(t *testing.T) {
	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/tables/%s/query", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "Select column",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: string
				          - id: column-02
				            type: integer
				          - id: column-03
				            type: float
				          - id: column-04
				            type: boolean
				          - id: column-05
				            type: boolean
				        records:
				          - data: ["v1", -1, -3.14, true, false]
				            createdAt: "2021-10-01T00:00:00Z"
				          - data: [null, null, null, null, null]
				            createdAt: "2021-10-02T00:00:00Z"
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			select:
			  columns:
			    - column: {{ .column01 }}
			    - column: {{ .column02 }}
			    - column: {{ .column03 }}
			    - column: {{ .column04 }}
			    - column: {{ .column05 }}
			  orderBy:
			    - key: {metadata: createdAt}
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
				"column02": testutils.GetUUID("column-02"),
				"column03": testutils.GetUUID("column-03"),
				"column04": testutils.GetUUID("column-04"),
				"column05": testutils.GetUUID("column-05"),
			}),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"records": []interface{}{
					[]interface{}{"v1", float64(-1), float64(-3.14), true, false},
					[]interface{}{nil, nil, nil, nil, nil},
				},
				"limit": float64(10),
			},
		},
		{
			Title: "Select metadata",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - id: record-01
				            data: [1]
				            createdAt: "2021-10-01T00:00:00Z"
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			select:
			  columns: [{metadata: id}, {metadata: createdAt}]
			`, nil),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"records": []interface{}{
					[]interface{}{testutils.GetUUID("record-01"), "2021-10-01T00:00:00Z"},
				},
				"limit": float64(10),
			},
		},
		{
			Title: "Select property",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [1]
				            properties:
				              key: value
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			select:
			  columns: [{property: key}, {property: key2}]
			`, nil),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"records": []interface{}{
					[]interface{}{"value", nil},
				},
				"limit": float64(10),
			},
		},
		{
			Title: "Where",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [1]
				          - data: [2]
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			select:
			  columns: [{column: {{ .column01 }} }]
			  where: {eq: [{column: {{ .column01 }} }, {value: 2}]}
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
			}),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"records": []interface{}{
					[]interface{}{float64(2)},
				},
				"limit": float64(10),
			},
		},
		{
			Title: "Order by desc",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [1]
				          - data: [2]
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			select:
			  columns: [{column: {{ .column01 }} }]
			  orderBy: [{key: {column: {{ .column01 }} }, order: desc}]
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
			}),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"records": []interface{}{
					[]interface{}{float64(2)},
					[]interface{}{float64(1)},
				},
				"limit": float64(10),
			},
		},
		{
			Title: "Offset and limit",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [1]
				          - data: [2]
				          - data: [3]
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			select:
			  columns: [{column: {{ .column01 }} }]
			  orderBy: [{key: {column: {{ .column01 }} }}]
			  offset: 1
			  limit: 1
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
			}),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"records": []interface{}{
					[]interface{}{float64(2)},
				},
				"limit": float64(1),
			},
		},
		{
			Title: "Aggregate functions",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [1]
				          - data: [2]
				          - data: [3]
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			select:
			  columns:
			    - func: count
			      args: [{metadata: id}]
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
			}),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"records": []interface{}{
					[]interface{}{float64(3)},
				},
				"limit": float64(10),
			},
		},
		{
			Title: "Operators",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [null]
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			select:
			  columns:
			    - eq: [{value: 1}, {value: 1}]
			    - eq: [{value: 1}, {value: 2}]
			    - ne: [{value: 1}, {value: 1}]
			    - ne: [{value: 1}, {value: 2}]
			    - gt: [{value: 1}, {value: 1}]
			    - gt: [{value: 2}, {value: 1}]
			    - ge: [{value: 1}, {value: 1}]
			    - ge: [{value: 1}, {value: 2}]
			    - ge: [{value: 2}, {value: 1}]
			    - lt: [{value: 1}, {value: 1}]
			    - lt: [{value: 1}, {value: 2}]
			    - le: [{value: 1}, {value: 1}]
			    - le: [{value: 2}, {value: 1}]
			    - le: [{value: 1}, {value: 2}]
			    - like: [{value: abc}, {value: "ab%"}]
			    - like: [{value: abc}, {value: "ac%"}]
			    - isNull: {value: null}
			    - isNull: {value: "null"}
			    - and: [{value: true}, {value: true}]
			    - and: [{value: true}, {value: false}]
			    - or: [{value: false}, {value: false}]
			    - or: [{value: false}, {value: true}]
			    - not: {value: true}
			    - not: {value: false}
			    - add: [{value: 1}, {value: 2}]
			    - sub: [{value: 1}, {value: 2}]
			    - mul: [{value: 2}, {value: 2}]
			    - div: [{value: 1}, {value: 2}]
			    - mod: [{value: 3}, {value: 2}]
			    - neg: {value: 1}
			`, nil),
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"records": []interface{}{
					[]interface{}{
						float64(1),
						float64(0),
						float64(0),
						float64(1),
						float64(0),
						float64(1),
						float64(1),
						float64(0),
						float64(1),
						float64(0),
						float64(1),
						float64(1),
						float64(0),
						float64(1),
						float64(1),
						float64(0),
						float64(1),
						float64(0),
						float64(1),
						float64(0),
						float64(0),
						float64(1),
						float64(0),
						float64(1),
						float64(3),
						float64(-1),
						float64(4),
						float64(0.5),
						float64(1),
						float64(-1),
					},
				},
				"limit": float64(10),
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
			select:
			  columns:
			    - value: 1
			`, nil),
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

func TestQueryTableRecordUpdate(t *testing.T) {
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
				            type: integer
				          - id: column-02
				            type: float
				          - id: column-03
				            type: string
				        records:
				          - data: [1, 3.14, "v1"]
				          - data: [2, 2.71, "v2"]
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			update:
			  set:
			    - to: {column: {{ .column01 }} }
			      value: {add: [{column: {{ .column01 }} }, {value: 2}]}
			    - to: {column: {{ .column02 }} }
			      value: {value: 1.41}
			  where: {eq: [{column: {{ .column03 }} }, {value: "v1"}]}
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
				"column02": testutils.GetUUID("column-02"),
				"column03": testutils.GetUUID("column-03"),
			}),
			StatusCode: http.StatusOK,
			Output:     map[string]interface{}{},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Select from the table
				res := selectTable(router, testutils.GetUUID("table-01"), makeJSON(`
				select:
				  columns: [{column: {{ .column01 }} }, {column: {{ .column02 }} }, {column: {{ .column03 }} }]
				  orderBy: [{key: {column: {{ .column01 }} }}]
				`, map[string]interface{}{
					"column01": testutils.GetUUID("column-01"),
					"column02": testutils.GetUUID("column-02"),
					"column03": testutils.GetUUID("column-03"),
				}))
				if diff := testutils.CompareJson(map[string]interface{}{
					"records": []interface{}{
						[]interface{}{float64(2), float64(2.71), "v2"},
						[]interface{}{float64(3), float64(1.41), "v1"},
					},
					"limit": float64(10),
				}, res); diff != "" {
					t.Errorf("[%s] Selected records mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Update properties",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				return testutils.LoadFixture(`
				organizations:
				  - id: org1
				    tables:
				      - id: table-01
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [1]
				            properties:
				              key1: value1
				          - data: [2]
				            properties:
				              key1: value2
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			update:
			  set:
			    - to: {property: key1}
			      value: {value: new-value}
			    - to: {property: key2}
			      value: {value: 3.14}
			  where: {eq: [{column: {{ .column01 }} }, {value: 2}]}
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
			}),
			StatusCode: http.StatusOK,
			Output:     map[string]interface{}{},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Select from the table
				res := selectTable(router, testutils.GetUUID("table-01"), makeJSON(`
				select:
				  columns: [{property: key1}, {property: key2}]
				  orderBy: [{key: {column: {{ .column01 }} }}]
				`, map[string]interface{}{
					"column01": testutils.GetUUID("column-01"),
				}))
				if diff := testutils.CompareJson(map[string]interface{}{
					"records": []interface{}{
						[]interface{}{"value1", nil},
						[]interface{}{"new-value", float64(3.14)},
					},
					"limit": float64(10),
				}, res); diff != "" {
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
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [1]
				`)
			},
			Path: makePath(testutils.GetUUID("table-02")),
			Body: makeJSON(`
			update:
			  set:
			    - to: {column: {{ .column01 }} }
			      value: {value: 2}
			  where: {value: true}
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
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

func TestQueryTableRecordDelete(t *testing.T) {
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
				            type: integer
				        records:
				          - data: [1]
				          - data: [2]
				`)
			},
			Path: makePath(testutils.GetUUID("table-01")),
			Body: makeJSON(`
			delete:
			  where:
			    eq:
			      - {column: {{ .column01 }} }
			      - {value: 1}
			`, map[string]interface{}{
				"column01": testutils.GetUUID("column-01"),
			}),
			StatusCode: http.StatusOK,
			Output:     map[string]interface{}{},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Select from the table
				res := selectTable(router, testutils.GetUUID("table-01"), makeJSON(`
				select:
				  columns: [{column: {{ .column01 }} }]
				  orderBy: [{key: {column: {{ .column01 }} }}]
				`, map[string]interface{}{
					"column01": testutils.GetUUID("column-01"),
				}))
				if diff := testutils.CompareJson(map[string]interface{}{
					"records": []interface{}{
						[]interface{}{float64(2)},
					},
					"limit": float64(10),
				}, res); diff != "" {
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
				        columns:
				          - id: column-01
				            type: integer
				        records:
				          - data: [1]
				`)
			},
			Path: makePath(testutils.GetUUID("table-02")),
			Body: makeJSON(`
			delete:
			  where: {value: true}
			`, nil),
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
