package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/tsujio/x-base/api/controllers/folder"
)

func SetFolderRoutes(router *mux.Router, db *gorm.DB) {
	controller := folder.FolderController{
		DB: db,
	}

	router.HandleFunc("", controller.CreateFolder).Methods(http.MethodPost)
	router.HandleFunc("/{folderID}", controller.GetFolder).Methods(http.MethodGet)
	router.HandleFunc("/{folderID}", controller.UpdateFolder).Methods(http.MethodPatch)
	router.HandleFunc("/{folderID}", controller.DeleteFolder).Methods(http.MethodDelete)
	router.HandleFunc("/{folderID}/children", controller.GetFolderChildren).Methods(http.MethodGet)
}
