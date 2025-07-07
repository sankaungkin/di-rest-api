package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
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
	{ID: 1, UnitName: "EACH"},
	{ID: 2, UnitName: "PACK"},
	{ID: 3, UnitName: "FEET"},
	{ID: 4, UnitName: "DOZEN"},
	{ID: 5, UnitName: "BOTTLE"},
}

var categories = []models.Category{
	{
		ID: 1, CategoryName: "Construction Materials",
	},
	{
		ID: 2, CategoryName: "Sanitary Ware",
	},
	{
		ID: 3, CategoryName: "PVC Pipe",
	},
	{

		ID: 4, CategoryName: "PVC Fitting",
	},
	{
		ID: 5, CategoryName: "GI Fitting",
	},
	{
		ID: 6, CategoryName: "ရေသလျောက်",
	},
	{
		ID: 7, CategoryName: "Glass Block",
	},
	{
		ID: 8, CategoryName: "တိုင်ခေါင်း",
	},
	{
		ID: 9, CategoryName: "Nail",
	},
	{
		ID: 10, CategoryName: "Concrete Nail",
	},
	{
		ID: 11, CategoryName: "Water Tap",
	},
	{
		ID: 12, CategoryName: "Water Spray",
	},
	{
		ID: 13, CategoryName: "Adhesive",
	},
	{
		ID: 14, CategoryName: "Tape",
	},
	{
		ID: 15, CategoryName: "Concrete Pole",
	},
	{
		ID: 16, CategoryName: "Concrete Block",
	},
	{
		ID: 17, CategoryName: "ကုန်မာ",
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

var products = []models.Product{
	{
		ID:              "P001",
		ProductName:     "Cement 5.25 CROWN",
		CategoryId:      1,
		Uom:             "EACH",
		DeriveUom:       "PACK",
		UomId:           1,
		DeriveUomId:     2,
		BuyPrice:        22000,
		SellPriceLevel1: 28000,
		DeriveUnitPrice: 28000,
		BrandName:       "CROWN",
		IsActive:        true,
	},
	// ('P001','Cement 5.25 CROWN',1,'','',1,2,22000,28000,28000,'CROWN',true)
	{
		ID:              "P002",
		BrandName:       "MATO",
		BuyPrice:        30000,
		IsActive:        true,
		ProductName:     "ToiletBowl MATO big",
		SellPriceLevel1: 35000,
		DeriveUnitPrice: 35000,
		Uom:             "EACH",
		DeriveUom:       "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     1,
	},
	{
		ID:              "P003",
		BrandName:       "SOGO",
		DeriveUom:       "FEET",
		BuyPrice:        35000,
		IsActive:        true,
		ProductName:     "PVC 4Inch Class 8.5 SOGO",
		SellPriceLevel1: 35000,
		DeriveUnitPrice: 35000,
		Uom:             "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     3,
	},
	{
		ID:              "P004",
		BrandName:       "SOGO",
		BuyPrice:        2000,
		IsActive:        true,
		ProductName:     "PVC 4Inch SK 8.5",
		SellPriceLevel1: 2500,
		DeriveUnitPrice: 2500,
		Uom:             "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     3,
		DeriveUom:       "FEET",
	},
	{
		ID:              "P005",
		BrandName:       "CROWN",
		BuyPrice:        25000,
		IsActive:        true,
		ProductName:     "Cement 4.25 APACHE",
		SellPriceLevel1: 28000,
		DeriveUnitPrice: 27000,
		Uom:             "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     2,
		DeriveUom:       "PACK",
	},
	{
		ID:              "P006",
		BrandName:       "n/a",
		BuyPrice:        2500,
		IsActive:        true,
		ProductName:     "ထုံးအိတ်",
		SellPriceLevel1: 3000,
		DeriveUnitPrice: 3000,
		Uom:             "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     1,
		DeriveUom:       "EACH",
	},
	{
		ID:              "P007",
		BrandName:       "SOGO",
		BuyPrice:        2500,
		IsActive:        true,
		ProductName:     "PVC Fitting 2Inch Tee",
		SellPriceLevel1: 3500,
		DeriveUnitPrice: 3500,
		Uom:             "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     1,
		DeriveUom:       "EACH",
	},
	{
		ID:              "P008",
		BrandName:       "SOGO",
		BuyPrice:        2000,
		IsActive:        true,
		ProductName:     "PVC Fitting 1-1/5Inch SK",
		DeriveUom:       "EACH",
		SellPriceLevel1: 2500,
		DeriveUnitPrice: 2500,
		Uom:             "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     1,
	},
	{
		ID:              "P009",
		BrandName:       "n/a",
		BuyPrice:        2000,
		IsActive:        true,
		ProductName:     "Glue 502",
		SellPriceLevel1: 2500,
		DeriveUnitPrice: 2500,
		Uom:             "EACH",
		DeriveUom:       "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     1,
	},
	{
		ID:              "P010",
		BrandName:       "n/a",
		BuyPrice:        12000,
		IsActive:        true,
		ProductName:     "Glue P-brand (Large)",
		SellPriceLevel1: 15000,
		DeriveUnitPrice: 15000,
		Uom:             "EACH",
		DeriveUom:       "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     1,
	},
	{
		ID:              "P011",
		BrandName:       "n/a",
		BuyPrice:        7500,
		IsActive:        true,
		ProductName:     "Glue P-brand (Mideam)",
		SellPriceLevel1: 8500,
		DeriveUnitPrice: 8500,
		Uom:             "EACH",
		DeriveUom:       "EACH",
		CategoryId:      1,
		UomId:           1,
		DeriveUomId:     1,
	},
}

var unitConversions = []models.UnitConversion{
	{
		Description:  "EACH-TO-EACH",
		ProductId:    "P001",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-TO-PACK",
		ProductId:    "P001",
		BaseUnit:     "EACH",
		DeriveUnit:   "PACK",
		BaseUnitId:   1,
		DeriveUnitId: 3,
		Factor:       19,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P002",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-TO-FEET",
		ProductId:    "P003",
		BaseUnit:     "EACH",
		DeriveUnit:   "FEET",
		BaseUnitId:   1,
		DeriveUnitId: 2,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P004",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P005",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P005",
		BaseUnit:     "EACH",
		DeriveUnit:   "PACK",
		BaseUnitId:   1,
		DeriveUnitId: 2,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P006",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P007",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P008",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P009",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P010",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
	{
		Description:  "EACH-ONLY",
		ProductId:    "P011",
		BaseUnit:     "EACH",
		DeriveUnit:   "EACH",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		Factor:       1,
	},
}

var productPrices = []models.ProductPrice{
	{
		ProductId: "P001",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 22000,
	},
	{
		ProductId: "P001",
		UnitId:    2,
		PriceType: "SELL",
		UnitPrice: 2500,
	},
	{
		ProductId: "P001",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 18500,
	},
	{
		ProductId: "P002",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 35000,
	},
	{
		ProductId: "P002",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 32000,
	},
	{
		ProductId: "P003",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 30000,
	},
	{
		ProductId: "P003",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 32000,
	},
	{
		ProductId: "P003",
		UnitId:    3,
		PriceType: "SELL",
		UnitPrice: 1800,
	},
	{
		ProductId: "P004",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 2300,
	},
	{
		ProductId: "P004",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 2500,
	},
	{
		ProductId: "P005",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 19500,
	},
	{
		ProductId: "P005",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 21000,
	},
	{
		ProductId: "P005",
		UnitId:    2,
		PriceType: "SELL",
		UnitPrice: 2000,
	},
	{
		ProductId: "P006",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 2500,
	},
	{
		ProductId: "P006",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 3000,
	},
	{
		ProductId: "P007",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 3000,
	},
	{
		ProductId: "P007",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 3500,
	},
	{
		ProductId: "P008",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 2000,
	},
	{
		ProductId: "P008",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 2500,
	},
	{
		ProductId: "P009",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 2300,
	},
	{
		ProductId: "P009",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 2500,
	},
	{
		ProductId: "P010",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 13000,
	},
	{
		ProductId: "P010",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 15000,
	},
	{
		ProductId: "P011",
		UnitId:    1,
		PriceType: "BUY",
		UnitPrice: 7500,
	},
	{
		ProductId: "P011",
		UnitId:    1,
		PriceType: "SELL",
		UnitPrice: 8500,
	},
}

var productStocks = []models.ProductStock{
	{
		ProductId:    "P001",
		BaseUnitId:   1,
		DeriveUnitId: 2,
		BaseQty:      50,
		DerivedQty:   6,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P002",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P003",
		BaseUnitId:   1,
		DeriveUnitId: 3,
		BaseQty:      50,
		DerivedQty:   10,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P004",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P005",
		BaseUnitId:   1,
		DeriveUnitId: 2,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P006",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   10,
	},
	{
		ProductId:    "P007",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P008",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P009",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P010",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   5,
	},
	{
		ProductId:    "P011",
		BaseUnitId:   1,
		DeriveUnitId: 1,
		BaseQty:      50,
		DerivedQty:   0,
		ReorderLvl:   5,
	},
}

func load() {
	fmt.Println("......Seeding data ....")

	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		fmt.Println("Seeding unit of measures data ....")
		if err := tx.Create(&unitOfMeasures).Error; err != nil {
			return fmt.Errorf("failed to seed unitOfMeasures: %w", err)
		}

		fmt.Println("Seeding categories data ....")
		if err := tx.Create(&categories).Error; err != nil {
			return fmt.Errorf("failed to seed categories: %w", err)
		}

		fmt.Println("Seeding customers data ....")
		if err := tx.Create(&customers).Error; err != nil {
			return fmt.Errorf("failed to seed customers: %w", err)
		}

		fmt.Println("Seeding suppliers data ....")
		if err := tx.Create(&suppliers).Error; err != nil {
			return fmt.Errorf("failed to seed suppliers: %w", err)
		}

		fmt.Println("Seeding products data ....")
		if err := tx.Create(&products).Error; err != nil {
			return fmt.Errorf("failed to seed products: %w", err)
		}

		fmt.Println("Seeding product prices data ....")
		if err := tx.Create(&productPrices).Error; err != nil {
			return fmt.Errorf("failed to seed productPrices: %w", err)
		}

		fmt.Println("Seeding product stocks data ....")
		if err := tx.Create(&productStocks).Error; err != nil {
			return fmt.Errorf("failed to seed productStocks: %w", err)
		}

		fmt.Println("Seeding unit conversions data ....")
		if err := tx.Create(&unitConversions).Error; err != nil {
			return fmt.Errorf("failed to seed unitConversions: %w", err)
		}

		// If everything is successful
		fmt.Println("..... Seeding completed .....")
		return nil
	})

	if err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}
}
