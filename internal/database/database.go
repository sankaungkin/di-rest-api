package database

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseInterface interface {
	NewDB() (*gorm.DB, error)
}

var (
	db     *gorm.DB
	dbOnce sync.Once
	Blue   = "\033[34m"
	Reset  = "\033[0m"
)

func NewDB() (*gorm.DB, error) {

	dbOnce.Do(func() {
		log.Println(Blue + "------> NewDB constructor is called <-----" + Reset)
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal(err)
		}

		Host := os.Getenv("DB_HOST")
		Port := os.Getenv("POSTGRES_PORT")
		Password := os.Getenv("POSTGRES_PASSWORD")
		User := os.Getenv("POSTGRES_USER")
		DBName := os.Getenv("POSTGRES_DB")
		SSLMode := os.Getenv("SSLMODE")
		// Set default values if empty
		if SSLMode == "" {
			SSLMode = "disable" // or "prefer" depending on your requirements
		}
		timeZone := "Asia/Yangon"

		var dsn = fmt.Sprintf(
			"host=%s port=%s password=%s user=%s dbname=%s sslmode=%s timezone=%s",
			Host, Port, Password, User, DBName, SSLMode, timeZone)

		log.Print(dsn)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			// return nil, err
			log.Fatal(err)
		}
		err = db.AutoMigrate(
			&models.Category{},
			&models.Customer{},
			&models.Supplier{},
			&models.UnitOfMeasure{},
			&models.UnitConversion{},
			&models.Product{},
			&models.ProductPrice{},
			&models.ProductPriceHistory{},
			&models.ProductStock{},
			&models.Inventory{},
			&models.Sale{},
			&models.SaleDetail{},
			&models.SaleReturn{},
			&models.SaleReturnItem{},
			&models.ProductUnit{},
			&models.Purchase{},
			&models.PurchaseDetail{},
			&models.ItemTransaction{},
			&models.DailySummaries{},
			&models.Cashbook{},
			&models.Expense{},

			&models.User{})
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Migration done.....")
	})
	return db, nil

}
