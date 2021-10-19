package testutils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/tsujio/x-base/databases"
)

var db *gorm.DB

func makeDBConfig() *databases.DBConfig {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		dbPort = 3306
	}
	dbName := os.Getenv("DB_NAME")
	dbType := os.Getenv("DB_TYPE")

	if !strings.HasSuffix(dbName, "_test") {
		log.Fatal("Database name must have suffix '_test' for testing")
	}

	return &databases.DBConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		DBType:   dbType,
	}
}

func GetDB() *gorm.DB {
	return db
}

func RefreshDB() {
	dbConfig := makeDBConfig()

	// Get table names
	rows, err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = ?", dbConfig.DBName).Rows()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	err = db.Exec("SET foreign_key_checks = 0").Error
	if err != nil {
		log.Fatal(err)
	}

	// Drop tables
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)).Error
		if err != nil {
			log.Fatal(err)
		}
	}

	err = db.Exec("SET foreign_key_checks = 1").Error
	if err != nil {
		log.Fatal(err)
	}

	// Create tables
	err = databases.Setup(dbConfig, os.Getenv("MIGRATIONS_DIR"))
	if err != nil {
		log.Fatal(err)
	}
}
