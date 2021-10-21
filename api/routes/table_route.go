package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/controllers/table"
)

func SetTableRoutes(router *mux.Router, db *gorm.DB) {
	controller := table.TableController{
		DB: db,
	}

	router.HandleFunc("", controller.CreateTable).Methods(http.MethodPost)
	router.HandleFunc("/{tableID}", controller.GetTable).Methods(http.MethodGet)
	router.HandleFunc("/{tableID}", controller.UpdateTable).Methods(http.MethodPatch)
	router.HandleFunc("/{tableID}", controller.DeleteTable).Methods(http.MethodDelete)
}
