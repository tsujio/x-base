package organization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/copier"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils/responses"
	"github.com/tsujio/x-base/logging"
)

var (
	defaultPage     = 1
	defaultPageSize = 10
)

func (controller *OrganizationController) GetOrganizationList(w http.ResponseWriter, r *http.Request) {
	// Decode request parameters
	var input schemas.GetOrganizationListInput
	err := schemas.DecodeQuery(r.URL.Query(), &input)
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
	var sortKeyOpt []models.GetListSortKey
	if err := copier.Copy(&sortKeyOpt, &sortKeys); err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make query option", err)
		return
	}
	opts := models.GetOrganizationListOpts{
		Sort:   sortKeyOpt,
		Offset: (*input.Page - 1) * *input.PageSize,
		Limit:  *input.PageSize,
	}
	organizations, totalCount, err := models.GetOrganizationList(controller.DB, &opts)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get organizations", err)
		return
	}

	// Convert to output schema
	var output schemas.OrganizationList
	err = copier.Copy(&output.Organizations, &organizations)
	if err != nil {
		responses.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
		return
	}
	if input.Properties != "" {
		keys := strings.Split(input.Properties, ",")
		for i := range output.Organizations {
			output.Organizations[i].Properties = organizations[i].Properties.SelectKeys(keys)
		}
	}
	output.TotalCount = totalCount

	// Send response
	err = json.NewEncoder(w).Encode(&output)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
