package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/models"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	database.NewDB()
	load()
}

var categories = []models.Category{
	{
		CategoryName: "Construction Materials",
	},
	{
		CategoryName: "Sanitary Ware",
	},
	{
		CategoryName: "PVC Pipe",
	},
	{
		CategoryName: "PVC Fitting",
	},
	{
		CategoryName: "GI Fitting",
	},
	{
		CategoryName: "ရေသလျောက်",
	},
	{
		CategoryName: "Glass Block",
	},
	{
		CategoryName: "တိုင်ခေါင်း",
	},
	{
		CategoryName: "Nail",
	},
	{
		CategoryName: "Concrete Nail",
	},
	{
		CategoryName: "Water Tap",
	},
	{
		CategoryName: "Water Spray",
	},
	{
		CategoryName: "Adhesive",
	},
	{
		CategoryName: "Tape",
	},
	{
		CategoryName: "Concrete Pole",
	},
	{
		CategoryName: "Concrete Block",
	},
	{
		CategoryName: "ကုန်မာ",
	},
}

var products = []models.Product{
	{
		ID:              "P001",
		BrandName:       "CROWN",
		BuyPrice:        7500,
		IsActive:        true,
		ProductName:     "Cement 4.25 CROWN",
		ReorderLvl:      10,
		QtyOnHand:       50,
		SellPriceLevel1: 8300,
		SellPriceLevel2: 8000,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:              "P002",
		BrandName:       "MATO",
		BuyPrice:        25000,
		IsActive:        true,
		ProductName:     "ToiletBowl MATO big",
		ReorderLvl:      5,
		QtyOnHand:       50,
		SellPriceLevel1: 31000,
		SellPriceLevel2: 30000,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:              "P003",
		BrandName:       "SOGO",
		BuyPrice:        18000,
		IsActive:        true,
		ProductName:     "PVC 4Inch Class 8.5 SOGO",
		ReorderLvl:      3,
		QtyOnHand:       50,
		SellPriceLevel1: 21000,
		SellPriceLevel2: 20000,
		Uom:             "EACH",
		CategoryId:      1,
	},
}

var customers = []models.Customer{
	{
		Name:    "Work-In Customer",
		Address: "Work In",
		Phone:   "09-12346",
	},
	{
		Name:    "ရာပြည့် ကွန်ကရစ်",
		Address: "19 Street",
		Phone:   "09-45645666",
	},
	{
		Name:    "သန်းထိုက်စံ",
		Address: "19 Street",
		Phone:   "09-4566332",
	},
}

var suppliers = []models.Supplier{
	{
		Name:    "999",
		Address: "24th street",
		Phone:   "09-12346",
	},
	{
		Name:    "OSCAR TRADING",
		Address: "81st street",
		Phone:   "09-45645666",
	},
	{
		Name:    "တော်ဝင်",
		Address: "24 Street",
		Phone:   "09-4566332",
	},
}

func load() {

	fmt.Println("......Seeding data ....")
	db, _ := database.NewDB()

	fmt.Println("Seeding categories data ....")
	db.Create(&categories)

	fmt.Println("Seeding products data ....")
	db.Create(&products)

	fmt.Println("Seeding customers data ....")
	db.Create(&customers)

	fmt.Println("Seeding suppliers data ....")
	db.Create(&suppliers)

	fmt.Println("..... Seeding completed .....")
}
