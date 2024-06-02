package main

import (
	_ "github.com/sankangkin/di-rest-api/cmd/docs"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/sankangkin/di-rest-api/internal/router"
)

// @title					REST-API with(golang fiber, google wire dependency injection)
// @version					1.0
// @description				This is an auto-generated API docs.
// @termsOfService				http://swagger.io/terms/
// @contact.name				API Support
// @contact.email				sankaungkin@gmail.com
// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.

func main() {

	// log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/swagger/*", swagger.HandlerDefault) // default

	router.Initialize(app)
	app.Listen(":3030")
	
}