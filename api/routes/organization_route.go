package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/controllers/organization"
)

func SetOrganizationRoutes(router *mux.Router, db *gorm.DB) {
	controller := organization.OrganizationController{
		DB: db,
	}

	router.HandleFunc("", controller.CreateOrganization).Methods(http.MethodPost)
	router.HandleFunc("", controller.GetOrganizationList).Methods(http.MethodGet)
	router.HandleFunc("/{organizationID}", controller.GetOrganization).Methods(http.MethodGet)
	router.HandleFunc("/{organizationID}", controller.UpdateOrganization).Methods(http.MethodPatch)
	router.HandleFunc("/{organizationID}", controller.DeleteOrganization).Methods(http.MethodDelete)
}
