package organization

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jinzhu/copier"

	"github.com/tsujio/x-base/api/models"
	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/api/utils"
	"github.com/tsujio/x-base/logging"
)

func (controller *OrganizationController) GetOrganizationList(w http.ResponseWriter, r *http.Request) {
	// Decode request parameters
	var input schemas.GetOrganizationListInput
	err := schemas.DecodeQuery(r.URL.Query(), &input)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid request parameter", err)
		return
	}

	// Fetch
	opts := models.GetOrganizationListOpts{
		Sort:   "CreatedAt ASC, ID ASC",
		Offset: (input.Page - 1) * input.PageSize,
		Limit:  input.PageSize + 1,
	}
	organizations, totalCount, err := models.GetOrganizationList(controller.DB, &opts)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get organizations", err)
		return
	}
	hasNext := len(organizations) > input.PageSize
	if hasNext {
		organizations = organizations[:len(organizations)-1]
	}

	// Convert to output schema
	var output schemas.OrganizationList
	err = copier.Copy(&output.Organizations, &organizations)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Failed to make output data", err)
		return
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
