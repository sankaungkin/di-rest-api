package router

import (
	"log"

	"github.com/gofiber/fiber/v2"
	categoryDi "github.com/sankangkin/di-rest-api/internal/domain/category/di"
	customerDi "github.com/sankangkin/di-rest-api/internal/domain/customer/di"
	inventoryDi "github.com/sankangkin/di-rest-api/internal/domain/inventory/di"
	productDi "github.com/sankangkin/di-rest-api/internal/domain/product/di"
	saleDI "github.com/sankangkin/di-rest-api/internal/domain/sale/di"
	supplierDi "github.com/sankangkin/di-rest-api/internal/domain/supplier/di"
)

func Initialize(app *fiber.App) {

	api := app.Group("/api")
	api.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("---->  Hello from stt api using go fiber framework <-- ")
	})

	// category di
	catService, err := categoryDi.InitCategory()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// category route
	categories := api.Group("/category")
	categories.Post("/", catService.CreateCategory)
	categories.Get("/",catService.GetAllCategorie)
	categories.Get("/:id", catService.GetCategoryById)
	categories.Put("/:id", catService.UpdateCatagory)
	categories.Delete("/:id", catService.DeleteCategory)


	// product di
	productService, err := productDi.InitProductDI()
	if err != nil {
		log.Fatalf(err.Error())
	} 
	// product route
	products := api.Group("/product")
	products.Post("/", productService.CreateProduct)
	products.Get("/", productService.GetAllProducts)
	products.Get("/:id", productService.GetProductById)
	products.Put("/:id", productService.UpdateProduct)
	products.Delete("/:id",productService.DeleteProduct)

	// customer di
	customerService, err := customerDi.InitCustomer()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// customer route
	customer := api.Group("/customer")
	customer.Post("/", customerService.CreateCustomer)
	customer.Get("/", customerService.GetAllCustomers)
	customer.Get("/:id", customerService.GetCustomerById)
	customer.Put("/:id", customerService.UpdateCustomer)
	customer.Delete("/:id", customerService.DeleteCustomer)

	// supplier di
	supplierService, err := supplierDi.InitSupplier()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// supplier route
	supplier := api.Group("/supplier")
	supplier.Post("/", supplierService.CreateSupplier)
	supplier.Get("/", supplierService.GetAllSuppliers)
	supplier.Get("/:id", supplierService.GetSupplierById)
	supplier.Put("/:id", supplierService.UpdateSupplier)
	supplier.Delete("/:id", supplierService.DeleteSupplier)

	// inventory di
	inventoryService, err := inventoryDi.InitInventoryDI()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// inventory route
	inventory := api.Group("/inventory")
	inventory.Get("/", inventoryService.GetInventory)
	inventory.Post("/increase", inventoryService.IncreaseInventory)
	inventory.Post("/decrease", inventoryService.DecreaseInventory)

	// sale di
	saleService, err := saleDI.InitSaleDI()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// sale route
	sale := api.Group("/sale")
	sale.Post("/", saleService.CreateSale)
	sale.Get("/", saleService.GetAllSales)
	sale.Get("/:id", saleService.GetById)
}