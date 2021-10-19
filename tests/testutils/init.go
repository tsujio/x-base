package testutils

import (
	"log"

	"github.com/tsujio/x-base/databases"
	"github.com/tsujio/x-base/logging"
)

func init() {
	// Set logger
	logging.SetLogger(&logging.DefaultLogger{})

	// Open db
	_db, err := databases.Open(makeDBConfig())
	if err != nil {
		log.Fatal(err)
	}
	db = _db
}
