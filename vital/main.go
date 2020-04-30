package main

import (
	"fmt"
	"io/ioutil"

	"github.com/gofiber/fiber"
	"gopkg.in/yaml.v2"
)

// ClientProfile is a configuration for an Opsani client
type ClientProfile struct {
	InitToken string `yaml:"init_token"`
	BaseURL   string `yaml:"base_url"`
	AppID     string `yaml:"app_id"`
	APIToken  string `yaml:"api_token"`
}

// AppConfig represents data from the .config.yaml file
type AppConfig struct {
	// ProfilesByToken is a map of single use tokens to client profiles
	Profiles []ClientProfile `yaml:"profiles"`
}

func main() {
	app := fiber.New()

	// Serve static assets
	app.Static("/", "./assets")

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
	app.Get("/init/:token", func(c *fiber.Ctx) {
		config := AppConfig{}
		data, err := ioutil.ReadFile(".config.yaml")
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal([]byte(data), &config)
		if err != nil {
			panic(err)
		}

		var profile *ClientProfile
		for _, p := range config.Profiles {
			if p.InitToken == c.Params("token") {
				profile = &p
			}
		}

		if profile != nil {
			c.JSON(fiber.Map{
				"base_url": profile.BaseURL,
				"app":      profile.AppID,
				"token":    profile.APIToken,
			})
		} else {
			c.Send("Unknown token")
			c.SendStatus(404)
		}
	})

	app.Listen(8080)
}
