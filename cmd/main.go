package main

import (
	"log"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sankangkin/di-rest-api/internal/router"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := fiber.New()
	app.Use(cors.New())

	router.Initialize(app)
	app.Listen(":6666")
	
}