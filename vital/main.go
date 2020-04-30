package main

import (
	"fmt"
	"os"
	"github.com/gofiber/fiber"
)

func main() {
	app := fiber.New()

	// Serve static assets
	app.Static("/", "./assets")

	// TODO: Index serving static files
	// builds, index, script
	app.Post("/signup", func(c *fiber.Ctx) {
		fmt.Printf("Hello! name=%q, email=%q, app=%q", 
			c.FormValue("name"), c.FormValue("email"), c.FormValue("app_name"))
	})

	// TODO: Template the script to send
	// run `opsani init <token>`
	// app.Get("/install.sh/:token", func(c *fiber.Ctx) {
	// 	// Get the query param

	// 	fmt.Printf("Token: %s", c.Params("token"))
	// })
	app.Get("/token/:token", func(c *fiber.Ctx) {
		// Get the query param
		fmt.Printf("Token is %q", c.Params("token"))
	})	

	app.Listen(8080)
}
