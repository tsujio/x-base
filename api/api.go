package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/routes"
	"github.com/tsujio/x-base/logging"
)

func CreateRouter(db *gorm.DB) http.Handler {
	router := mux.NewRouter().
		StrictSlash(true)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Organization routes
	organizationRouter := router.PathPrefix("/organizations").Subrouter()
	routes.SetOrganizationRoutes(organizationRouter, db)

	// Table routes
	tableRouter := router.PathPrefix("/tables").Subrouter()
	routes.SetTableRoutes(tableRouter, db)

	handler := cors.AllowAll().Handler(router)

	return handler
}

func Run(host string, port int, db *gorm.DB) error {
	router := CreateRouter(db)

	addr := fmt.Sprintf("%s:%d", host, port)

	logging.Info(fmt.Sprintf("Listen on %s\n", addr), nil)

	err := http.ListenAndServe(addr, router)
	if err != nil {
		return xerrors.Errorf("Failed to start api: %w", err)
	}

	return nil
}
