package database

import (
	"github.com/imdatngo/gowhere"
	"github.com/kevinxvu/goone/pkg/logging"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// New creates new database connection to the database server
func New(dbType, dbPsn string, enableLog bool) (*gorm.DB, error) {
	config := new(gorm.Config)
	db, err := NewWithConfig(dbType, dbPsn, config)
	if err != nil {
		return nil, err
	}

	if enableLog {
		db.Logger = logging.NewGormLogger().LogMode(gormlogger.Info)
	} else {
		db.Logger = logging.NewGormLogger().LogMode(gormlogger.Silent)
	}

	return db, nil
}

// NewWithConfig creates new database connection using custom gorm config.
func NewWithConfig(dialect, dbPsn string, cfg *gorm.Config) (db *gorm.DB, err error) {
	switch dialect {
	case "mysql":
		gowhere.DefaultConfig.Dialect = gowhere.DialectMySQL
		db, err = gorm.Open(mysql.Open(dbPsn), cfg)
		if err != nil {
			return nil, err
		}
	case "postgres":
		gowhere.DefaultConfig.Dialect = gowhere.DialectPostgreSQL
		db, err = gorm.Open(postgres.Open(dbPsn), cfg)
		if err != nil {
			return nil, err
		}
	case "sqlite3":
		gowhere.DefaultConfig.Dialect = gowhere.DialectMySQL
		db, err = gorm.Open(sqlite.Open(dbPsn), cfg)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
