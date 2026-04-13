package postgres

import (
	"gorm.io/gorm"
)

type resourceDB struct {
	db *gorm.DB
}

func NewResourceDB(db *gorm.DB) *resourceDB {
	return &resourceDB{db: db}
}
