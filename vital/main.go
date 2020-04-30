package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/jordan-wright/email"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
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

func loadConfig() *AppConfig {
	config := &AppConfig{}
	data, err := ioutil.ReadFile(".config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		panic(err)
	}

	return config
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getGmailService() *gmail.Service {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		panic(err)
	}
	httpClient := config.Client(context.Background(), tok)

	srv, err := gmail.New(httpClient)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
	return srv
}

func main() {
	app := fiber.New()

	// Serve static assets
	app.Static("/", "./assets")

	app.Post("/signup", func(c *fiber.Ctx) {
		name := c.FormValue("name")
		recipient := c.FormValue("email")
		// appName := c.FormValue("app_name")
		// Send the email via Gmail
		config := loadConfig()
		token := config.Profiles[0].InitToken
		gmailSvc := getGmailService()
		var message gmail.Message

		body := fmt.Sprintf(`Hi %s,
Thank you for your interest in Opsani Vital.
We are eager to share the impact of continuous optimization on your engineering practices and business.

The easiest way to get started is using the Opsani CLI. It will automatically link to your optimization
engine and provides tutorials and auto-discovery capabilities to make integrating your app a breeze.

> curl http://localhost:5678/install.sh/%s | sh

Just follow the instructions from there!

Cheers,
- The Opsani Team
`, name, token)
		e := email.NewEmail()
		e.From = "vital@opsani.com"
		e.To = []string{recipient}
		e.Subject = "Welcome to Opsani Vital!"
		e.Text = []byte(body)

		messagePayload, err := e.Bytes()
		if err != nil {
			panic(err)
		}
		message.Raw = base64.StdEncoding.EncodeToString(messagePayload)
		_, err = gmailSvc.Users.Messages.Send("me", &message).Do()
		if err != nil {
			log.Fatalf("Unable to send message: %v", err)
		}
		fmt.Println("Sent email:", string(messagePayload))
		c.Set("Content-Type", "text/html")
		c.SendString(`<html><body><p>Success! Check your email for further instructions.</p></body></html`)
	})

	// Returns an instance of the script that will round-trip the init token
	app.Get("/install.sh/:token", func(c *fiber.Ctx) {
		token := c.Params("token")
		data, err := ioutil.ReadFile("assets/install.sh")
		if err != nil {
			panic(err)
		}

		// Replace our token value so it gets back to the user
		tokenEnvVar := fmt.Sprintf("OPSANI_INIT_TOKEN=%s", token)
		script := strings.Replace(string(data), `OPSANI_INIT_TOKEN="${OPSANI_INIT_TOKEN:-}"`, tokenEnvVar, 1)
		c.SendString(script)
	})

	// 	fmt.Printf("Token: %s", c.Params("token"))
	// })
	app.Get("/init/:token", func(c *fiber.Ctx) {
		config := loadConfig()

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
