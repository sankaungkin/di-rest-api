package database

import (
	"log"

	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database interface {
	GetDB() *gorm.DB
}


type PostgresDB struct {
	db *gorm.DB
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.Category{},
		&models.Customer{},
		&models.Supplier{},
		&models.Product{},
		&models.Inventory{},
		&models.Sale{},
		&models.SaleDetail{},
		&models.Purchase{},
		&models.PurchaseDetail{},
		&models.ItemTransaction{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return &PostgresDB{db: db}, nil
	
}

func (db *PostgresDB)GetDB() *gorm.DB{
	return db.db
} 