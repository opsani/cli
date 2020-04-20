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

package opsani_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/opsani/cli/opsani"
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	client := opsani.NewClient()
	assert.NotNil(t, client)
}

// Test Initialization Behaviors

func TestThatNewClientSetsJSONContentTypeHeader(t *testing.T) {
	client := opsani.NewClient()
	assert.Equal(t, "application/json", client.GetHeaders().Get("Content-Type"))
}

func TestThatNewClientSetsJSONAcceptHeader(t *testing.T) {
	client := opsani.NewClient()
	assert.Equal(t, "application/json", client.GetHeaders().Get("Accept"))
}

func TestThatNewClientSetsBaseURL(t *testing.T) {
	client := opsani.NewClient()
	assert.Equal(t, "https://api.opsani.com", client.GetBaseURL())
}

func TestThatSettingBaseURLTrimsTrailingSlash(t *testing.T) {
	client := opsani.NewClient()
	client.SetBaseURL("https://api.opsani.com/")
	assert.Equal(t, "https://api.opsani.com", client.GetBaseURL())
}

// Test API Interactions

func TestStartApp(t *testing.T) {
	okResponse := map[string]interface{}{"status": "ok"}
	okJSON, _ := json.Marshal(okResponse)
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		w.Write(okJSON)
	}))
	defer ts.Close()

	client := opsani.NewWithHTTPClient(ts.Client())
	client.SetBaseURL(ts.URL)
	resp, err := client.StartApp()
	assert.Nil(t, err)
	result := resp.Result()
	assert.NotNil(t, result)
	assert.Equal(t, &okResponse, result)
}
