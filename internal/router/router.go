package router

import (
	"log"

	"github.com/gofiber/fiber/v2"
	authDi "github.com/sankangkin/di-rest-api/internal/auth/di"
	categoryDi "github.com/sankangkin/di-rest-api/internal/domain/category/di"
	customerDi "github.com/sankangkin/di-rest-api/internal/domain/customer/di"
	inventoryDi "github.com/sankangkin/di-rest-api/internal/domain/inventory/di"
	transactionDi "github.com/sankangkin/di-rest-api/internal/domain/itemtransactions/di"
	productDi "github.com/sankangkin/di-rest-api/internal/domain/product/di"
	productpriceDi "github.com/sankangkin/di-rest-api/internal/domain/productprice/di"
	productStockDi "github.com/sankangkin/di-rest-api/internal/domain/productstock/di"
	purchaseDi "github.com/sankangkin/di-rest-api/internal/domain/purchase/di"
	saleDi "github.com/sankangkin/di-rest-api/internal/domain/sale/di"
	supplierDi "github.com/sankangkin/di-rest-api/internal/domain/supplier/di"
	unitconversionDi "github.com/sankangkin/di-rest-api/internal/domain/unitconversion/di"
	unitofmeasurementDi "github.com/sankangkin/di-rest-api/internal/domain/unitofmeasurement/di"
	"github.com/sankangkin/di-rest-api/internal/middleware"
)

func Initialize(app *fiber.App) {

	api := app.Group("/api")
	api.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("---->  Hello from stt api using go fiber framework <-- ")
	})

	// authentication di
	authService, err := authDi.InitAuth()
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}
	// auth route
	auth := api.Group("/auth")
	auth.Post("/register", authService.SignUp)
	auth.Post("/login", authService.SignIn)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)

	// category di
	catService, err := categoryDi.InitCategory()
	if err != nil {
		log.Fatalf("Failed to initialize category service: %v", err)
	}
	// category route
	categories := api.Group("/categories")
	// categories.Use(middleware.Protected())
	categories.Post("/", catService.CreateCategory)
	categories.Get("/", catService.GetAllCategorie)
	categories.Get("/:id", catService.GetCategoryById)
	categories.Put("/:id", catService.UpdateCatagory)
	categories.Delete("/:id", catService.DeleteCategory)

	// product di
	productService, err := productDi.InitProductDI()
	if err != nil {
		log.Fatalf("Failed to initialize product service: %v", err)
	}

	products := api.Group("/products")
	products.Use(middleware.Protected())

	products.Post("/", productService.CreateProduct)
	products.Get("/", productService.GetAllProducts)

	// products.Get("/stocks", productService.GetAllProductStocks)
	// products.Get("/stocks/:id", productService.GetProductStocksById)
	products.Get("/prices", productService.GetAllProductPrices)
	products.Get("/prices/:id", productService.GetProductUnitPricesById)

	products.Put("/:id", productService.UpdateProduct)
	products.Delete("/:id", productService.DeleteProduct)
	products.Get("/:id", productService.GetProductById) // ❗️Keep this at the BOTTOM

	// unitconversion di
	unitConversionService, err := unitconversionDi.InitUnitConversionDI()
	if err != nil {
		log.Fatalf("Failed to initialize unit conversion service: %v", err)
	}
	unitconversion := api.Group("/unitconversions")
	unitconversion.Use(middleware.Protected())
	unitconversion.Post("/", unitConversionService.CreateUnitConversion)
	unitconversion.Get("/", unitConversionService.GetAllUnitConversions)
	unitconversion.Get("/:id", unitConversionService.GetUnitConversionById)
	unitconversion.Put("/:id", unitConversionService.UpdateUnitConversion)
	unitconversion.Delete("/:id", unitConversionService.DeleteUnitConversion)
	// unit conversion route

	//unit of measurement di
	unitOfMeasurementService, err := unitofmeasurementDi.InitUnitOfMeasurementDI()
	if err != nil {
		log.Fatalf("Failed to initialize unit of measurement service: %v", err)
	}
	unitofmeasurement := api.Group("/uoms")
	unitofmeasurement.Use(middleware.Protected())
	unitofmeasurement.Post("/", unitOfMeasurementService.CreateUnitOfMeasurement)
	unitofmeasurement.Get("/", unitOfMeasurementService.GetAllUnitOfMeasurement)
	unitofmeasurement.Get("/:id", unitOfMeasurementService.GetUnitOfMeasurementById)
	unitofmeasurement.Put("/:id", unitOfMeasurementService.UpdateUnitOfMeasurement)
	unitofmeasurement.Delete("/:id", unitOfMeasurementService.DeleteUnitOfMeasurement)
	// unit of measurement route

	// productstock di
	productStockService, err := productStockDi.InitProductStockDI()
	if err != nil {
		log.Fatalf("Failed to initialize product stock service: %v", err)
	}
	productstocks := api.Group("/productstocks")
	productstocks.Use(middleware.Protected())
	productstocks.Post("/", productStockService.CreateProductStocks)
	productstocks.Get("/", productStockService.GetAllProductStocks)
	productstocks.Get("/:id", productStockService.GetProductStocksById)
	productstocks.Put("/:id", productStockService.UpdateProductStocksById)

	//productprice di
	productPriceService, err := productpriceDi.InitProductPriceDI()
	if err != nil {
		log.Fatalf("Failed to initialize product price service: %v", err)
	}
	productprices := api.Group("/productprices")
	productprices.Use(middleware.Protected())
	productprices.Post("/", productPriceService.CreateProductPrice)
	productprices.Get("/", productPriceService.GetAllProductPrices)
	productprices.Get("/:id", productPriceService.GetProductPriceById)
	productprices.Put("/:id", productPriceService.UpdateProductPrice)

	// item transactions di
	transactionService, err := transactionDi.InitTransactionDI()
	if err != nil {
		log.Fatalf("Failed to initialize transaction service: %v", err)
	}
	// item transactions route
	transactions := api.Group("/transactions")
	transactions.Use(middleware.Protected())
	transactions.Get("/", transactionService.GetAll)
	transactions.Get("/by-product/:productId", transactionService.GetTransactionsByProductId)
	transactions.Get("/by-type/:tranType", transactionService.GetTransactionsByTransactionType)
	transactions.Get("/by-product-type/:productId/:tranType", transactionService.GetByProductIdAndTranType)
	transactions.Post("/adjustment", transactionService.CreateAdjustmentTransaction)

	// customer di
	customerService, err := customerDi.InitCustomer()
	if err != nil {
		log.Fatalf("Failed to initialize customer service: %v", err)
	}
	// customer route
	customer := api.Group("/customers")
	customer.Use(middleware.Protected())
	customer.Post("/", customerService.CreateCustomer)
	customer.Get("/", customerService.GetAllCustomers)
	customer.Get("/:id", customerService.GetCustomerById)
	customer.Put("/:id", customerService.UpdateCustomer)
	customer.Delete("/:id", customerService.DeleteCustomer)

	// supplier di
	supplierService, err := supplierDi.InitSupplier()
	if err != nil {
		log.Fatalf("Failed to initialize supplier service: %v", err)
	}
	// supplier route
	supplier := api.Group("/suppliers")
	supplier.Use(middleware.Protected())
	supplier.Post("/", supplierService.CreateSupplier)
	supplier.Get("/", supplierService.GetAllSuppliers)
	supplier.Get("/:id", supplierService.GetSupplierById)
	supplier.Put("/:id", supplierService.UpdateSupplier)
	supplier.Delete("/:id", supplierService.DeleteSupplier)

	// inventory di
	inventoryService, err := inventoryDi.InitInventoryDI()
	if err != nil {
		log.Fatalf("Failed to initialize inventory service: %v", err)
	}
	// inventory route
	inventory := api.Group("/inventories")
	inventory.Use(middleware.Protected())
	inventory.Get("/", inventoryService.GetAllInventories)
	inventory.Post("/increase", inventoryService.IncreaseInventory)
	inventory.Post("/decrease", inventoryService.DecreaseInventory)

	// sale di
	saleService, err := saleDi.InitSaleDI()
	if err != nil {
		log.Fatalf("Failed to initialize sale service: %v", err)
	}
	// sale route
	sale := api.Group("/sales")
	sale.Use(middleware.Protected())
	sale.Post("/", saleService.CreateSale)
	sale.Get("/", saleService.GetAllSales)
	sale.Get("/:id", saleService.GetById)

	// purchase di
	purchaseService, err := purchaseDi.InitPurchaseDI()
	if err != nil {
		log.Fatalf("Failed to initialize purchase service: %v", err)
	}
	// purchase route
	purchase := api.Group("/purchases")
	purchase.Use(middleware.Protected())
	purchase.Post("/", purchaseService.CreatePurchase)
	purchase.Get("/", purchaseService.GetAllPurchases)
	purchase.Get("/:id", purchaseService.GetById)
}
