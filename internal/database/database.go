package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database interface {
	NewDB() (*gorm.DB, error)
}

func NewDB() (*gorm.DB, error) {

	log.Println("Starting NewDB .......")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf(err.Error())
	}

	Host := os.Getenv("DB_HOST")
	Port := os.Getenv("POSTGRES_PORT")
	Password := os.Getenv("POSTGRES_PASSWORD")
	User := os.Getenv("POSTGRES_USER")
	DBName := os.Getenv("POSTGRES_DB")
	SSLMode := os.Getenv("SSLMODE")

	var dsn = fmt.Sprintf(
		"host=%s port=%s password=%s user=%s dbname=%s sslmode=%s",
		Host, Port, Password, User, DBName, SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Println("Migration start.....")
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
		&models.User{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Migration done.....")
    return db, nil
}


