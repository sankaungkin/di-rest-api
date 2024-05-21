package router

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/di"
)

func Initialize(app *fiber.App) {

	api := app.Group("/api")
	api.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("---->  Hello from stt api using go fiber framework <-- ")
	})

	catService, err := di.InitCategory()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// route
	categories := api.Group("/category")
	categories.Get("/",catService.GetAllCategorie)
	categories.Get("/:id", catService.GetCategoryById)
}