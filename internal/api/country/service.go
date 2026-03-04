package country

import (
	"github.com/vuduongtp/go-core/pkg/database"

	"gorm.io/gorm"
)

// New creates new country application service
func New(db *gorm.DB, cdb database.Intf) *Country {
	return &Country{
		db:  db,
		cdb: cdb,
	}
}

// Country represents country application service
type Country struct {
	db  *gorm.DB
	cdb database.Intf
}
