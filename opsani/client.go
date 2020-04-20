// Copyright 2020 Opsani
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opsani

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"encoding/json"

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
Request Primitives
*/

func (c *Client) newRequest() *resty.Request {
	req := c.restyClient.NewRequest()
	req.SetResult(&map[string]interface{}{})
	req.SetError(&APIError{})
	return req
}

func (c *Client) appResourceURLPath(resource string) string {
	return fmt.Sprintf("/accounts/%s/applications/%s/%s", c.appDomain, c.appName, resource)
}

/**
Configuration
*/
func (c *Client) appConfigURLPath() string {
	return c.appResourceURLPath("config")
}

// GetConfig retrieves the Opsani app configuration from the API
func (c *Client) GetConfig() (*resty.Response, error) {
	return c.newRequest().
		Get(c.appConfigURLPath())
}

// SetConfigFromBody sets the app configuration from the given body, overwriting the existing configuration
func (c *Client) SetConfigFromBody(body interface{}, apply bool) (*resty.Response, error) {
	return c.newRequest().
		SetQueryParams(map[string]string{
			"reset": strconv.FormatBool(apply),
		}).
		SetBody(body).
		Put(c.appConfigURLPath())
}

// PatchConfigFromBody patches the existing app configuration from the given body producing a merged configuration
func (c *Client) PatchConfigFromBody(body interface{}, apply bool) (*resty.Response, error) {
	return c.newRequest().
		SetHeader("Content-Type", "application/merge-patch+json").
		SetQueryParams(map[string]string{
			"reset": strconv.FormatBool(apply),
			"patch": "true",
		}).
		SetBody(body).
		Put(c.appConfigURLPath())
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
	return c.appResourceURLPath("state")
}

// StartApp starts a stopped Opsani app
func (c *Client) StartApp() (*resty.Response, error) {
	return c.newRequest().
		SetBody(`{"target_state": "running"}`).
		Patch(c.stateURLPath())
}

// StopApp stops a running Opsani app
func (c *Client) StopApp() (*resty.Response, error) {
	return c.newRequest().
		SetBody(`{"target_state": "stopped"}`).
		Patch(c.stateURLPath())
}

// RestartApp stops a running Opsani app
func (c *Client) RestartApp() (*resty.Response, error) {
	return c.newRequest().
		SetHeader("Content-Type", "application/merge-patch+json").
		SetQueryParams(map[string]string{
			"reset": "true",
			"patch": "true",
		}).
		SetBody(`{}`).
		Put(c.appConfigURLPath())
}

// GetAppStatus retrieves the status of the Opsani app from the API
func (c *Client) GetAppStatus() (*resty.Response, error) {
	return c.newRequest().
		Get(c.stateURLPath())
}

/**
Authentication actions
*/

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

// AfterResponseWriteBodyToFile configures a request handler to write responses to the specified file
func (c *Client) AfterResponseWriteBodyToFile(filename string) {
	c.restyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetOutput(filename)
		return nil
	})
}

// AfterResponseWritePrettyJSONBodyToFile appends response middleware to write pretty JSON responses to the specified file
func (c *Client) AfterResponseWritePrettyJSONBodyToFile(filename string) {
	c.restyClient.OnAfterResponse(func(rc *resty.Client, resp *resty.Response) error {
		return WritePrettyJSONBytesToFile(resp.Body(), filename)
	})
}

// WritePrettyJSONBytesToFile writes an array of unformmated JSON bytes in pretty printed formatted to the target file
func WritePrettyJSONBytesToFile(body []byte, file string) error {
	outputPath := ""

	if !filepath.IsAbs(file) {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		outputPath += pwd + string(filepath.Separator)
	}

	outputPath = filepath.Clean(outputPath + file)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	var formattedJSON bytes.Buffer
	err = json.Indent(&formattedJSON, body, "", "    ")
	if err != nil {
		return err
	}
	reader := bufio.NewReader(&formattedJSON)
	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}

	return err
}

// WritePrettyJSONStringsToFile writes an array of unformatted JSON strings in pretty printed formatted to the target file
func WritePrettyJSONStringsToFile(strings []string, filename string) error {
	outputPath := ""

	if !filepath.IsAbs(filename) {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		outputPath += pwd + string(filepath.Separator)
	}

	outputPath = filepath.Clean(outputPath + filename)

	file, err := os.OpenFile(outputPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	datawriter := bufio.NewWriter(file)
	for _, jsonString := range strings {
		var formattedJSON bytes.Buffer
		err = json.Indent(&formattedJSON, []byte(jsonString), "", "    ")
		if err != nil {
			return err
		}
		formattedJSON.WriteString("\n")
		reader := bufio.NewReader(&formattedJSON)
		_, err = datawriter.ReadFrom(reader)
		if err != nil {
			return err
		}
	}

	datawriter.Flush()
	return nil
}
