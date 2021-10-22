package responses

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tsujio/x-base/api/schemas"
	"github.com/tsujio/x-base/logging"
)

func SendErrorResponse(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	if status == http.StatusInternalServerError {
		if err != nil {
			logging.Error(fmt.Sprintf("%s: %+v", message, err), r)
		} else {
			logging.Error(message, r)
		}
	}

	w.WriteHeader(status)
	msg := message
	if err != nil {
		msg += ": " + err.Error()
	}
	if e := json.NewEncoder(w).Encode(&schemas.Error{
		Message: msg,
	}); e != nil {
		logging.Error(fmt.Sprintf("%+v", e), r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
