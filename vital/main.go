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
	"github.com/matcornic/hermes/v2"
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
		config := loadConfig()
		token := config.Profiles[0].InitToken
		gmailSvc := getGmailService()
		var message gmail.Message

		h := hermes.Hermes{
			Product: hermes.Product{
				Name:      "Opsani",
				Link:      "https://www.opsani.com/",
				Copyright: "© Opsani All rights reserved 2020",
				Logo:      "http://34.222.186.235/opsani.png",
				// Logo:      "https://www.opsani.com/wp-content/uploads/2019/04/opsani_logo.svg",
			},
		}
		codeBlock := fmt.Sprintf("```bash\n$ curl http://localhost:5678/install.sh/%s | sh\n```", token)
		markdown := fmt.Sprintf(`
Cloud cost savings are close at hand.

To start optimizing, install the Opsani CLI:

%s

---

`, codeBlock)
		hermesEmail := hermes.Email{
			Body: hermes.Body{
				Name: name,
				Intros: []string{
					"Welcome to Opsani! We're very excited to have you on board.",
				},
				FreeMarkdown: hermes.Markdown(markdown),
				Outros: []string{
					"Need help or have questions? Just reply to this email and we are happy to help.",
				},
				Signature: "Cheers",
			},
		}

		// Generate an HTML email with the provided contents (for modern clients)
		emailBody, err := h.GenerateHTML(hermesEmail)
		if err != nil {
			panic(err) // Tip: Handle error with something else than a panic ;)
		}

		// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
		emailText, err := h.GeneratePlainText(hermesEmail)
		if err != nil {
			panic(err) // Tip: Handle error with something else than a panic ;)
		}

		// Send HTML and plain text emails via GMail
		e := email.NewEmail()
		e.From = "vital@opsani.com"
		e.To = []string{recipient}
		e.Subject = "Welcome to Opsani Vital!"
		e.Text = []byte(emailText)
		e.HTML = []byte(emailBody)

		messagePayload, err := e.Bytes()
		if err != nil {
			panic(err)
		}
		message.Raw = base64.URLEncoding.EncodeToString(messagePayload)
		_, err = gmailSvc.Users.Messages.Send("me", &message).Do()
		if err != nil {
			log.Printf("Unable to send message: %v\n", err)
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
				"base_url":  profile.BaseURL,
				"optimizer": profile.AppID,
				"token":     profile.APIToken,
			})
		} else {
			c.Send("Unknown token")
			c.SendStatus(404)
		}
	})

	app.Listen(8080)
}
