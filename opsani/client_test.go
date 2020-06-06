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
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

// Test Initialization Behaviors

func (s *ClientTestSuite) TestThatNewClientSetsJSONContentTypeHeader() {
	client := opsani.NewClient()
	s.Require().Equal("application/json", client.GetHeaders().Get("Content-Type"))
}

func (s *ClientTestSuite) TestThatNewClientSetsJSONAcceptHeader() {
	client := opsani.NewClient()
	s.Require().Equal("application/json", client.GetHeaders().Get("Accept"))
}

func (s *ClientTestSuite) TestThatNewClientSetsBaseURL() {
	client := opsani.NewClient()
	s.Require().Equal("https://api.opsani.com", client.GetBaseURL())
}

func (s *ClientTestSuite) TestThatSettingBaseURLTrimsTrailingSlash() {
	client := opsani.NewClient()
	client.SetBaseURL("https://api.opsani.com/")
	s.Require().Equal("https://api.opsani.com", client.GetBaseURL())
}

// Test API Interactions

func (s *ClientTestSuite) TestStartAppSuccess() {
	responseObj := map[string]interface{}{"status": "ok"}
	JSON, _ := json.Marshal(responseObj)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		w.Write(JSON)
	}))
	defer ts.Close()

	client := opsani.NewClient()
	client.SetBaseURL(ts.URL)
	resp, err := client.StartApp()
	s.Require().Nil(err)
	result := resp.Result()
	s.Require().NotNil(result)
	s.Require().Equal(&responseObj, result)
}

func (s *ClientTestSuite) TestStartAppAlreadyStartedError() {
	responseObj := opsani.APIError{
		Message:   "Server can only be started if it is in stopped or final state",
		Status:    "400 Bad Request",
		Traceback: "...",
		Version:   "18.4.0"}
	JSON, _ := json.Marshal(responseObj)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON)
	}))
	defer ts.Close()

	client := opsani.NewClient()
	client.SetBaseURL(ts.URL)
	resp, err := client.StartApp()
	s.Require().Error(err)
	result := resp.Result()
	s.Require().Empty(result)
	s.Require().Equal(&responseObj, err)
}
