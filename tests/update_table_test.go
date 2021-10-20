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

func TestUpdateTable(t *testing.T) {
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

	makePath := func(id uuid.UUID) string {
		return fmt.Sprintf("/tables/%s", id)
	}

	testCases := []testutils.APITestCase{
		{
			Title: "Name",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				if err := createOrganizations(1); err != nil {
					return err
				}
				if err := (&models.Folder{
					TableFilesystemEntry: models.TableFilesystemEntry{
						ID:             models.UUID(uuids[1]),
						OrganizationID: models.UUID(uuids[0]),
						Name:           "folder-01",
					},
				}).Create(db); err != nil {
					return err
				}
				for i := 0; i < 2; i++ {
					o := models.Table{
						TableFilesystemEntry: models.TableFilesystemEntry{
							ID:             models.UUID(uuids[2+i]),
							OrganizationID: models.UUID(uuids[0]),
							Name:           fmt.Sprintf("table-%02d", i+1),
							ParentFolderID: (*models.UUID)(&uuids[1]),
						},
					}
					if err := o.Create(testutils.GetDB()); err != nil {
						return err
					}
				}
				return nil
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
			},
			Path: makePath(uuids[2]),
			Body: map[string]interface{}{
				"name": "new-table",
			},
			StatusCode: http.StatusOK,
			Output: map[string]interface{}{
				"id":              testutils.UUID{},
				"organization_id": uuids[0].String(),
				"name":            "new-table",
				"type":            "table",
				"path": []interface{}{
					map[string]interface{}{
						"id":   testutils.UUID{},
						"name": "folder-01",
						"type": "folder",
					},
					map[string]interface{}{
						"id":   testutils.UUID{},
						"name": "new-table",
						"type": "table",
					},
				},
				"created_at": testutils.Timestamp{},
				"updated_at": testutils.Timestamp{},
			},
			PostCheck: func(tc *testutils.APITestCase, router http.Handler, output map[string]interface{}) {
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/tables/%s", uuids[3]), nil)
				if res["name"] != "table-02" {
					t.Errorf("[%s] Modified other data:\n%s", tc.Title, res["name"])
				}
			},
		},
		{
			Title: "ParentFolderID",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				if err := createOrganizations(1); err != nil {
					return err
				}
				if err := (&models.Folder{
					TableFilesystemEntry: models.TableFilesystemEntry{
						ID:             models.UUID(uuids[1]),
						OrganizationID: models.UUID(uuids[0]),
						Name:           "folder-01",
					},
				}).Create(db); err != nil {
					return err
				}
				for i := 0; i < 2; i++ {
					o := models.Table{
						TableFilesystemEntry: models.TableFilesystemEntry{
							ID:             models.UUID(uuids[2+i]),
							OrganizationID: models.UUID(uuids[0]),
							Name:           fmt.Sprintf("table-%02d", i+1),
						},
					}
					if err := o.Create(testutils.GetDB()); err != nil {
						return err
					}
				}
				return nil
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
			},
			Path: makePath(uuids[2]),
			Body: map[string]interface{}{
				"parent_folder_id": uuids[1].String(),
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
						"name": "folder-01",
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
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/tables/%s", uuids[3]), nil)
				if len(res["path"].([]interface{})) != 1 {
					t.Errorf("[%s] Modified other data:\n%s", tc.Title, res["path"])
				}
			},
		},
		{
			Title: "Move to root folder",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				if err := createOrganizations(1); err != nil {
					return err
				}
				if err := (&models.Folder{
					TableFilesystemEntry: models.TableFilesystemEntry{
						ID:             models.UUID(uuids[1]),
						OrganizationID: models.UUID(uuids[0]),
						Name:           "folder-01",
					},
				}).Create(db); err != nil {
					return err
				}
				for i := 0; i < 2; i++ {
					o := models.Table{
						TableFilesystemEntry: models.TableFilesystemEntry{
							ID:             models.UUID(uuids[2+i]),
							OrganizationID: models.UUID(uuids[0]),
							Name:           fmt.Sprintf("table-%02d", i+1),
							ParentFolderID: (*models.UUID)(&uuids[1]),
						},
					}
					if err := o.Create(testutils.GetDB()); err != nil {
						return err
					}
				}
				return nil
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
			},
			Path: makePath(uuids[2]),
			Body: map[string]interface{}{
				"parent_folder_id": "00000000-0000-0000-0000-000000000000",
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
				// Reacquire and compare with the previous response
				res := testutils.ServeGet(router, fmt.Sprintf("/tables/%s", output["id"]), nil)
				if diff := testutils.CompareJson(output, res); diff != "" {
					t.Errorf("[%s] Reacquired response mismatch:\n%s", tc.Title, diff)
				}

				// Did not change other data
				res = testutils.ServeGet(router, fmt.Sprintf("/tables/%s", uuids[3]), nil)
				if len(res["path"].([]interface{})) != 2 {
					t.Errorf("[%s] Modified other data:\n%s", tc.Title, res["path"])
				}
			},
		},
		{
			Title: "Not found",
			Prepare: func(tc *testutils.APITestCase, db *gorm.DB) error {
				if err := createOrganizations(1); err != nil {
					return err
				}
				for i := 0; i < 2; i++ {
					o := models.Table{
						TableFilesystemEntry: models.TableFilesystemEntry{
							ID:             models.UUID(uuids[1+i]),
							OrganizationID: models.UUID(uuids[0]),
							Name:           fmt.Sprintf("table-%02d", i+1),
						},
					}
					if err := o.Create(testutils.GetDB()); err != nil {
						return err
					}
				}
				return nil
			},
			Header: http.Header{
				"X-ORGANIZATION-ID": []string{uuids[0].String()},
			},
			Path: makePath(uuids[3]),
			Body: map[string]interface{}{
				"name": "new-table",
			},
			StatusCode: http.StatusNotFound,
			Output: map[string]interface{}{
				"message": "Not found",
			},
		},
	}

	for _, tc := range testCases {
		tc.Method = http.MethodPatch
		testutils.RunTestCase(t, tc)
	}
}
