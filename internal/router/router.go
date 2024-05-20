package router

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/handler"
	"github.com/sankangkin/di-rest-api/internal/repository"
	"github.com/sankangkin/di-rest-api/internal/service"
)

func Initialize(app *fiber.App) {


	DB, err := database.NewDB()
	if err != nil {
		log.Fatalf(err.Error())
	}

	repo := repository.NewCategoryRepository(DB)
	srv := service.NewCategoryService(repo)
	hdl := handler.NewCategoryHandler(srv)

	api := app.Group("/api")
	api.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("---->  Hello from stt api using go fiber framework <-- ")
	})

	// category
	categories := api.Group("/category")
	categories.Get("/",hdl.GetAllCategorie)
}