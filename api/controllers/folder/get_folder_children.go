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

	// Get children
	opts := models.GetFolderChildrenOpts{
		Sort:        "Name ASC, ID ASC",
		Offset:      (*input.Page - 1) * *input.PageSize,
		Limit:       *input.PageSize + 1,
		ComputePath: true,
	}
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
	for _, child := range children {
		var schema schemas.FolderChild
		switch c := child.(type) {
		case models.Table:
			if err := copier.Copy(&schema, &c); err != nil {
				responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
			}
		case models.Folder:
			if err := copier.Copy(&schema, &c); err != nil {
				responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
			}
		default:
			responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data (invalid type)", nil)
			return
		}
		output.Children = append(output.Children, schema)
	}
	output.TotalCount = totalCount
	output.HasNext = hasNext

	// Send response
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
