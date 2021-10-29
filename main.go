package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tsujio/x-base/api"
	"github.com/tsujio/x-base/databases"
	"github.com/tsujio/x-base/logging"
)

func main() {
	// Initialize logger
	logging.SetLogger(logging.DefaultLogger{})

	// Set up db
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbType := os.Getenv("DB_TYPE")
	dbConfig := &databases.DBConfig{
		User:     dbUser,
		Password: dbPassword,
		Host:     dbHost,
		Port:     dbPort,
		DBName:   dbName,
		DBType:   dbType,
	}
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	err := databases.Setup(dbConfig, migrationsDir)
	if err != nil {
		logging.Error(fmt.Sprintf("Failed to set up db: %+v", err), nil)
		return
	}
	db, err := databases.Open(dbConfig, nil)
	if err != nil {
		logging.Error(fmt.Sprintf("Failed to open db: %+v", err), nil)
		return
	}

	host := os.Getenv("HOST")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8000
	}

	// Run api
	err = api.Run(host, port, db)
	if err != nil {
		logging.Error(fmt.Sprintf("Failed to run api: %+v", err), nil)
		os.Exit(1)
	}
}
