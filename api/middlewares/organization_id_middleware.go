package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func OrganizationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if values, exists := r.Header["X-ORGANIZATION-ID"]; exists && len(values) > 0 {
			organizationID, err := uuid.Parse(values[0])
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "organizationID", organizationID))
			}
		}

		next.ServeHTTP(w, r)
	})
}

func GetOrganizationID(r *http.Request) uuid.UUID {
	id := r.Context().Value("organizationID")
	if id == nil {
		return uuid.Nil
	}
	return id.(uuid.UUID)
}
