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

var unitOfMeasures = []models.UnitOfMeasure{
	{UnitName: "EACH"},
	{UnitName: "PACK"},
	{UnitName: "FEET"},
	{UnitName: "DOZEN"},
	{UnitName: "BOTTLE"},
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
		ID:          "P001",
		BrandName:   "CROWN",
		BuyPrice:    22000,
		IsActive:    true,
		ProductName: "Cement 5.25 CROWN",
		// ReorderLvl:       10,
		// QtyOnHand:        50,
		// DerivedQtyOnHand: 15,
		SellPriceLevel1: 28000,
		SellPriceLevel2: 28000,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P002",
		BrandName:   "MATO",
		BuyPrice:    30000,
		IsActive:    true,
		ProductName: "ToiletBowl MATO big",
		// ReorderLvl:       5,
		// QtyOnHand:        50,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 35000,
		SellPriceLevel2: 35000,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P003",
		BrandName:   "SOGO",
		BuyPrice:    35000,
		IsActive:    true,
		ProductName: "PVC 4Inch Class 8.5 SOGO",
		// ReorderLvl:       3,
		// QtyOnHand:        50,
		// DerivedQtyOnHand: 16,
		SellPriceLevel1: 35000,
		SellPriceLevel2: 35000,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P004",
		BrandName:   "SOGO",
		BuyPrice:    2000,
		IsActive:    true,
		ProductName: "PVC 4Inch SK 8.5",
		// ReorderLvl:       3,
		// QtyOnHand:        20,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 2500,
		SellPriceLevel2: 2500,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P005",
		BrandName:   "CROWN",
		BuyPrice:    25000,
		IsActive:    true,
		ProductName: "Cement 4.25 APACHE",
		// ReorderLvl:       10,
		// QtyOnHand:        50,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 28000,
		SellPriceLevel2: 27000,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P006",
		BrandName:   "n/a",
		BuyPrice:    2500,
		IsActive:    true,
		ProductName: "ထုံးအိတ်",
		// ReorderLvl:       10,
		// QtyOnHand:        50,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 3000,
		SellPriceLevel2: 3000,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P007",
		BrandName:   "SOGO",
		BuyPrice:    2500,
		IsActive:    true,
		ProductName: "PVC Fitting 2Inch Tee",
		// ReorderLvl:       2,
		// QtyOnHand:        63,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 3500,
		SellPriceLevel2: 3500,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P008",
		BrandName:   "SOGO",
		BuyPrice:    2000,
		IsActive:    true,
		ProductName: "PVC Fitting 1-1/5Inch SK",
		// ReorderLvl:       5,
		// QtyOnHand:        33,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 2500,
		SellPriceLevel2: 2500,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P009",
		BrandName:   "n/a",
		BuyPrice:    2000,
		IsActive:    true,
		ProductName: "Glue 502",
		// ReorderLvl:       8,
		// QtyOnHand:        66,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 2500,
		SellPriceLevel2: 2500,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P010",
		BrandName:   "n/a",
		BuyPrice:    12000,
		IsActive:    true,
		ProductName: "Glue P-brand (Large)",
		// ReorderLvl:       4,
		// QtyOnHand:        98,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 15000,
		SellPriceLevel2: 15000,
		Uom:             "EACH",
		CategoryId:      1,
	},
	{
		ID:          "P011",
		BrandName:   "n/a",
		BuyPrice:    7500,
		IsActive:    true,
		ProductName: "Glue P-brand (Mideam)",
		// ReorderLvl:       8,
		// QtyOnHand:        53,
		// DerivedQtyOnHand: 0,
		SellPriceLevel1: 8500,
		SellPriceLevel2: 8500,
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

	fmt.Println("Seeding unit of measures data ....")
	db.Create(&unitOfMeasures)

	fmt.Println("..... Seeding completed .....")
}
