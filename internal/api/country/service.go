package country

import (
	"github.com/vuduongtp/go-core/internal/model"
	dbutil "github.com/vuduongtp/go-core/pkg/util/db"

	"gorm.io/gorm"
)

// New creates new country application service
func New(db *gorm.DB, cdb dbutil.Intf) *Country {
	return &Country{
		db:  db,
		cdb: cdb,
	}
}

// Country represents country application service
type Country struct {
	db  *gorm.DB
	cdb dbutil.Intf
}

// NewDB returns a new country database instance
func NewDB() *dbutil.DB {
	return dbutil.NewDB(model.Country{})
}
