package opsani

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Client provides a high level interface to the Opsani API
type Client struct {
	restyClient *resty.Client
	appDomain   string
	appName     string
}

// NewClient creates a new Opsani API client.
func NewClient() *Client {
	rc := resty.New().
		SetHeaders(map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json",
			"User-Agent":   "Opsani CLI",
		}).
		SetHostURL("https://api.opsani.com")
	return createClientWithRestyClient(rc)
}

// NewWithHTTPClient creates a new Opsani client with given `http.Client`.
func NewWithHTTPClient(hc *http.Client) *Client {
	return createClientWithHTTPClient(hc)
}

// NewWithRestyClient creates a new Opsani client with given `http.Client`.
func NewWithRestyClient(rc *resty.Client) *Client {
	return createClientWithRestyClient(rc)
}

func createClientWithRestyClient(rc *resty.Client) *Client {
	return &Client{
		restyClient: rc,
	}
}

func createClientWithHTTPClient(hc *http.Client) *Client {
	restyClient := resty.NewWithClient(hc)
	return createClientWithRestyClient(restyClient)
}

// GetClient returns the current `resty.Client` used by the opsani client.
func (c *Client) GetClient() *resty.Client {
	return c.restyClient
}

// SetApp sets the Opsani app that the client is interacting with
func (c *Client) SetApp(app string) *Client {
	c.appDomain = filepath.Dir(app)
	c.appName = filepath.Base(app)
	return c
}

// GetApp gets the Opsani app that the client is interacting with
func (c *Client) GetApp(app string) string {
	return filepath.Join(c.appDomain, c.appName)
}

/**
Configuration
*/
func (c *Client) configURLPath() string {
	return fmt.Sprintf("/accounts/%s/applications/%s/config", c.appDomain, c.appName)
}

// GetConfig retrieves the Opsani app configuration from the API
func (c *Client) GetConfig() (interface{}, error) {
	var result map[string]interface{}
	resp, err := c.restyClient.R().
		SetResult(&result).
		Get(c.configURLPath())
	return resp.Result(), err
}

/**
Authentication actions
*/

// func (c *Client) LoginWithCredentials(username string, password string) {

// }

// func (c *Client) LoginWithLink(link string) {

// }

// func (c *Client) Logout() {

// }

// IsAuthenticated returns true if an authentication token has been set.
func (c *Client) IsAuthenticated() bool {
	return c.restyClient.Token == ""
}

// SetBaseURL method sets the base URL against which all requests are executed
func (c *Client) SetBaseURL(URL string) *Client {
	c.restyClient.HostURL = strings.TrimRight(URL, "/")
	return c
}

// SetAuthToken method sets the auth token of the `Authorization` header for all API requests.
func (c *Client) SetAuthToken(token string) *Client {
	c.restyClient.SetAuthToken(token)
	return c
}

// SetDebug controls whether or not debugging is enabled on the API client
func (c *Client) SetDebug(enabled bool) *Client {
	c.restyClient.SetDebug(enabled)
	return c
}

// EnableTrace enables tracing information for all requests
func (c *Client) EnableTrace() *Client {
	c.restyClient.EnableTrace()
	c.restyClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		// Explore response object
		fmt.Println("Response Info:")
		fmt.Println("Error      :", resp.Error())
		fmt.Println("Status Code:", resp.StatusCode())
		fmt.Println("Status     :", resp.Status())
		fmt.Println("Time       :", resp.Time())
		fmt.Println("Received At:", resp.ReceivedAt())
		fmt.Println("Body       :\n", resp)
		fmt.Println()

		// Explore trace info
		fmt.Println("Request Trace Info:")
		ti := resp.Request.TraceInfo()
		fmt.Println("DNSLookup    :", ti.DNSLookup)
		fmt.Println("ConnTime     :", ti.ConnTime)
		fmt.Println("TLSHandshake :", ti.TLSHandshake)
		fmt.Println("ServerTime   :", ti.ServerTime)
		fmt.Println("ResponseTime :", ti.ResponseTime)
		fmt.Println("TotalTime    :", ti.TotalTime)
		fmt.Println("IsConnReused :", ti.IsConnReused)
		fmt.Println("IsConnWasIdle:", ti.IsConnWasIdle)
		fmt.Println("ConnIdleTime :", ti.ConnIdleTime)

		return nil
	})
	return c
}
