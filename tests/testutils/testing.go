package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime/debug"
	"testing"

	"gorm.io/gorm"

	"github.com/tsujio/x-base/api"
)

func ServeGet(router http.Handler, path string, query url.Values) map[string]interface{} {
	url := url.URL{
		Path: path,
	}
	if query != nil {
		url.RawQuery = query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
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

type APITestCase struct {
	Title      string
	Prepare    func(*APITestCase, *gorm.DB) error
	Method     string
	Path       string
	Query      url.Values
	Header     http.Header
	Body       map[string]interface{}
	StatusCode int
	Output     interface{}
	PostCheck  func(*APITestCase, http.Handler, map[string]interface{})
	Context    map[string]interface{}
}

func RunTestCase(t *testing.T, tc APITestCase) {
	defer (func() {
		if err := recover(); err != nil {
			fmt.Println(string(debug.Stack()))
			t.Fatalf("[%s] %v", tc.Title, err)
		}
	})()

	if tc.Context == nil {
		tc.Context = map[string]interface{}{}
	}

	// Prepare db
	RefreshDB()
	if tc.Prepare != nil {
		if err := tc.Prepare(&tc, GetDB()); err != nil {
			t.Fatalf("[%s] %+v", tc.Title, err)
		}
	}

	// Prepare request
	url := url.URL{
		Path:     tc.Path,
		RawQuery: tc.Query.Encode(),
	}
	var body io.Reader
	if tc.Body != nil {
		b, err := json.Marshal(&tc.Body)
		if err != nil {
			t.Fatalf("[%s] %+v", tc.Title, err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(tc.Method, url.String(), body)
	if err != nil {
		t.Fatalf("[%s] %+v", tc.Title, err)
	}
	req.Header = tc.Header

	// Serve request
	r := httptest.NewRecorder()
	router := api.CreateRouter(
		GetDB(),
	)
	router.ServeHTTP(r, req)

	// Check status code
	if r.Code != tc.StatusCode {
		t.Errorf("[%s] Status code mismatch: expected=%v, actual=%v", tc.Title, tc.StatusCode, r.Code)
	}

	// Check output
	var result map[string]interface{}
	if tc.Output != nil {
		err = json.Unmarshal(r.Body.Bytes(), &result)
		if err != nil {
			t.Fatalf("[%s] %+v", tc.Title, err)
		}
		if diff := CompareJson(tc.Output, result); diff != "" {
			t.Errorf("[%s] Response mismatch:\n%s", tc.Title, diff)
		}
	} else {
		if body := r.Body.String(); body != "" {
			t.Errorf("[%s] Got response: %s", tc.Title, body)
		}
	}

	if tc.PostCheck != nil {
		tc.PostCheck(&tc, router, result)
	}
}
