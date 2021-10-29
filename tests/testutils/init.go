package testutils

import (
	"log"
	"os"
	"time"

	"github.com/tsujio/x-base/databases"
	"github.com/tsujio/x-base/logging"
	"gorm.io/gorm/logger"
)

func init() {
	// Set logger
	logging.SetLogger(&logging.DefaultLogger{})

	// Open db
	_db, err := databases.Open(makeDBConfig(), logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	}))
	if err != nil {
		log.Fatal(err)
	}
	db = _db
}
