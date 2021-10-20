package tests

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/tests/testutils"
)

func TestCreateTable(t *testing.T) {
	var uuids []uuid.UUID
	for i := 0; i < 10; i++ {
		uuids = append(uuids, uuid.New())
	}

	createOrganizations := func(n int) error {
		for i := 0; i < n; i++ {
			o := models.Organization{
				ID:   models.UUID(uuids[i]),
				Name: fmt.Sprintf("organization-%02d", i+1),
			}
			if err := o.Create(testutils.GetDB()); err != nil {
				return err
			}
		}
		return nil
	}

	createFolders := func(entries []models.TableFilesystemEntry) error {
		for _, e := range entries {
			if err := (&models.Folder{
				TableFilesystemEntry: e,
			}).Create(testutils.GetDB()); err != nil {
				return err
			}
		}
		return nil
	}

	testCases := []testutils.APITestCase{
		{
			Title: "Create at root",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				err := createOrganizations(1)
				return err
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
			},
			Body: map[string]interface{}{
				"name": "table-01",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.UUID{},
				"organization_id": uuids[0].String(),
				"name":            "table-01",
				"type":            "table",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.UUID{},
						"name": "table-01",
						"type": "table",
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Create at sub folder",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				if err := createOrganizations(1); err != nil {
					return err
				}
				if err := createFolders([]models.TableFilesystemEntry{
					{
						ID:             models.UUID(uuids[1]),
						OrganizationID: models.UUID(uuids[0]),
						Name:           "folder-01",
					},
					{
						ID:             models.UUID(uuids[2]),
						OrganizationID: models.UUID(uuids[0]),
						Name:           "folder-02",
						ParentFolderID: (*models.UUID)(&uuids[1]),
					},
					{
						ID:             models.UUID(uuids[3]),
						OrganizationID: models.UUID(uuids[0]),
						Name:           "folder-03",
						ParentFolderID: (*models.UUID)(&uuids[1]),
					},
				}); err != nil {
					return err
				}
				return nil
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
			},
			Body: map[string]interface{}{
				"name":             "table-01",
				"parent_folder_id": uuids[2].String(),
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.UUID{},
				"organization_id": uuids[0].String(),
				"name":            "table-01",
				"type":            "table",
				"path": []interface{}{
					map[string]interface{}{
						"id":   uuids[1].String(),
						"name": "folder-01",
						"type": "folder",
					},
					map[string]interface{}{
						"id":   uuids[2].String(),
						"name": "folder-02",
						"type": "folder",
					},
					map[string]interface{}{
						"id":   testutils.UUID{},
						"name": "table-01",
						"type": "table",
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}
			},
		},
		{
			Title: "Empty name",
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
			},
			Body: map[string]interface{}{
				"name": "",
			},
			StatusCode: http.StatusBadRequest,
			Output: map[string]interface{}{
				"message": testutils.Regexp{Pattern: `\bName\b`},
			},
		},
	}

	for _, tc := range testCases {
		tc.Path = "/tables"
		tc.Method = http.MethodPost
		testutils.RunTestCase(t, tc)
	}
}
