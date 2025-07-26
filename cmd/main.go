package main

import (
	"log"

	_ "github.com/sankangkin/di-rest-api/cmd/docs"

	productStockDi "github.com/sankangkin/di-rest-api/internal/domain/productstock/di"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sankangkin/di-rest-api/internal/router"
	"github.com/sankangkin/di-rest-api/internal/websocket"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title					REST-API with(golang fiber, google wire dependency injection)
// @version					1.0
// @description				This is an auto-generated API docs.
// @termsOfService				http://swagger.io/terms/
// @contact.name				SanKaungKin
// @contact.email				sankaungkin@gmail.com
// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
// @host						localhost:5555

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description	Type "Bearer" followed by a space and JWT token.

func main() {

	// log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := fiber.New()
	// app.Use(cors.New())

	app.Use(cors.New(cors.Config{
		//TAKE NOTE don't put the space between the AllowOrigins *****************************
		AllowOrigins:     "http://localhost:4200,http://192.168.100.7:4200,http://127.0.0.1:4200,http://192.168.100.7:5555,http://localhost:5555", // your frontend addresses
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
	}))

	productStockRepo, err := productStockDi.InitProductStockRepoDI()
	if err != nil {
		log.Fatalf("Failed to initialize product stock service: %v", err)
	}

	hub := websocket.NewHub(productStockRepo)
	go hub.Run()

	// app.Get("/swagger/*", swagger.HandlerDefault) // default

	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	router.Initialize(app, hub)
	app.Listen("0.0.0.0:5555")

}
