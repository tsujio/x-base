package folder

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils/responses"
	"github.com/tsujio/x-base/logging"
)

var (
	defaultPage     = 1
	defaultPageSize = 10
)

func (controller *FolderController) GetFolderChildren(w http.ResponseWriter, r *http.Request) {
	// Get folder id
	vars := mux.Vars(r)
	var id uuid.UUID
	err := schemas.DecodeUUID(vars, "folderID", &id)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid folder id", err)
		return
	}

	// Decode request parameters
	var input schemas.GetFolderChildrenInput
	err = schemas.DecodeQuery(r.URL.Query(), &input)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request parameter", err)
		return
	}

	// Decode sort key
	var sortKeys []schemas.GetListSortKey
	err = schemas.DecodeGetListSort(input.Sort, &sortKeys)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid sort parameter", err)
		return
	}

	if input.Page == nil {
		input.Page = &defaultPage
	}
	if input.PageSize == nil {
		input.PageSize = &defaultPageSize
	}

	// Fetch
	var folder *models.Folder
	if id == uuid.Nil {
		if input.OrganizationID == uuid.Nil {
			responses.SendErrorResponse(w, r, http.StatusBadRequest, "Organization id is required for root folder", nil)
			return
		}
		folder = &models.Folder{}
		folder.OrganizationID = models.UUID(input.OrganizationID)
	} else {
		f, err := (&models.TableFilesystemEntry{ID: models.UUID(id)}).GetFolder(controller.DB)
		if err != nil {
			if xerrors.Is(err, gorm.ErrRecordNotFound) {
				responses.SendErrorResponse(w, r, http.StatusNotFound, "Not found", nil)
				return
			}
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get folder", err)
			return
		}
		folder = f
	}

	// Prepare get opts
	var sortKeyOpt []models.GetFolderChildrenSortKey
	for _, k := range sortKeys {
		sortKeyOpt = append(sortKeyOpt, models.GetFolderChildrenSortKey{
			Key:         k.Key,
			OrderAsc:    k.OrderAsc,
			OrderDesc:   k.OrderDesc,
			OrderValues: k.OrderValues,
		})
	}
	opts := models.GetFolderChildrenOpts{
		Sort:        sortKeyOpt,
		Offset:      (*input.Page - 1) * *input.PageSize,
		Limit:       *input.PageSize + 1,
		ComputePath: true,
	}

	// Get children
	children, totalCount, err := folder.GetChildren(controller.DB, &opts)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get children", err)
		return
	}
	hasNext := len(children) > *input.PageSize
	if hasNext {
		children = children[:len(children)-1]
	}

	// Convert to output schema
	var output schemas.FolderChildren
	var c []schemas.FolderChild
	if err := copier.Copy(&c, &children); err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
	}
	output.Children = c
	output.TotalCount = totalCount
	output.HasNext = hasNext

	// Send response
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
