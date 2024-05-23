package router

import (
	"log"

	"github.com/gofiber/fiber/v2"
	categoryDi "github.com/sankangkin/di-rest-api/internal/domain/category/di"
	customerDi "github.com/sankangkin/di-rest-api/internal/domain/customer/di"
	productDi "github.com/sankangkin/di-rest-api/internal/domain/product/di"
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
	categories.Get("/",catService.GetAllCategorie)
	categories.Get("/:id", catService.GetCategoryById)

	// product di
	productService, err := productDi.InitProductDI()
	if err != nil {
		log.Fatalf(err.Error())
	} 
	// product route
	products := api.Group("/product")
	products.Get("/", productService.GetAllProducts)
	products.Get("/:id", productService.GetProductById)

	// customer di
	customerService, err := customerDi.InitCustomer()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// customer route
	customer := api.Group("/customer")
	customer.Get("/", customerService.GetAllCustomers)
	customer.Get("/:id", customerService.GetCustomerById)

	// supplier di
	supplierService, err := supplierDi.InitSupplier()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// supplier route
	supplier := api.Group("/supplier")
	supplier.Get("/", supplierService.GetAllSuppliers)
	supplier.Get("/:id", supplierService.GetSupplierById)

}