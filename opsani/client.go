package opsani

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"k8s.io/apimachinery/pkg/util/json"
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

func (c *Client) resourceURLPath(resource string) string {
	return fmt.Sprintf("/accounts/%s/applications/%s/%s", c.appDomain, c.appName, resource)
}

/**
Configuration
*/
func (c *Client) configURLPath() string {
	return c.resourceURLPath("config")
}

// GetConfig retrieves the Opsani app configuration from the API
func (c *Client) GetConfig() (interface{}, error) {
	var result map[string]interface{}
	resp, err := c.restyClient.R().
		SetResult(&result).
		Get(c.configURLPath())
	return resp.Result(), err
}

// GetConfigToOutput retrieves the Opsani app configuration from the API and writes it to a file
func (c *Client) GetConfigToOutput(filename string) error {
	_, err := c.restyClient.R().
		SetOutput(filename).
		Get(c.configURLPath())
	return err
}

/**
Lifecycle
*/

// APIError represents an error returned by the Opsani API
type APIError struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Traceback string `json:"traceback"`
	Version   string `json:"version"`
}

func (c *Client) stateURLPath() string {
	return c.resourceURLPath("state")
}

func (c *Client) responseOutcome(resp *resty.Response, respErr error) (interface{}, error) {
	if respErr != nil {
		return nil, respErr
	}

	if resp.IsSuccess() {
		if r := resp.Result(); r != nil {
			return r, nil
		}
	} else if resp.IsError() {
		if e := resp.Error(); e != nil {
			return e, nil
		}
	}
	var result map[string]interface{}
	err := json.Unmarshal(resp.Body(), result)
	return result, err
}

// StartApp starts a stopped Opsani app
func (c *Client) StartApp() (interface{}, error) {
	var result map[string]interface{}
	var apiError APIError
	resp, err := c.restyClient.R().
		SetBody(`{"target_state": "running"}`).
		SetResult(&result).
		SetError(&apiError).
		Patch(c.stateURLPath())
	return c.responseOutcome(resp, err)
}

// StopApp stops a running Opsani app
func (c *Client) StopApp() (interface{}, error) {
	var result map[string]interface{}
	var apiError APIError
	resp, err := c.restyClient.R().
		SetBody(`{"target_state": "stopped"}`).
		SetResult(&result).
		SetError(&apiError).
		Patch(c.stateURLPath())
	return c.responseOutcome(resp, err)
}

// RestartApp stops a running Opsani app
func (c *Client) RestartApp() (interface{}, error) {
	var result map[string]interface{}
	var apiError APIError
	resp, err := c.restyClient.R().
		SetHeader("Content-Type", "application/merge-patch+json").
		SetQueryParams(map[string]string{
			"reset": "true",
			"patch": "true",
		}).
		SetBody(`{}`).
		SetResult(&result).
		SetError(&apiError).
		Put(c.configURLPath())
	return c.responseOutcome(resp, err)
}

// GetAppStatus retrieves the status of the Opsani app from the API
func (c *Client) GetAppStatus() (interface{}, error) {
	var result map[string]interface{}
	var apiError APIError
	resp, err := c.restyClient.R().
		SetResult(&result).
		SetError(&apiError).
		Get(c.stateURLPath())
	return c.responseOutcome(resp, err)
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

// SetOutputDirectory sets the output directory for saving API responses
func (c *Client) SetOutputDirectory(dir string) {
	c.restyClient.SetOutputDirectory(dir)
}

// WriteOutputToFile configures a request handler to write responses to the specified file
func (c *Client) WriteOutputToFile(filename string) {
	c.restyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetOutput(filename)
		return nil // if its success otherwise return error
	})
}
