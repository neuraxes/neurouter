// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openai

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/neuraxes/neurouter/internal/conf"
)

// mockHTTPClient is a mock implementation of option.HTTPClient for testing.
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, errors.New("DoFunc is not set")
}

func TestNewOpenAIUpstream(t *testing.T) {
	Convey("Given a configuration and logger", t, func() {
		config := &conf.OpenAIConfig{
			BaseUrl: "https://api.openai.com/v1/",
			ApiKey:  "test-key",
		}

		Convey("When newOpenAIUpstream is called", func() {
			repo, err := newOpenAIUpstreamWithClient(config, nil, log.DefaultLogger)

			Convey("Then it should return a new upstream and no error", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				So(repo.config.BaseUrl, ShouldEqual, "https://api.openai.com/v1/")
				So(repo.config.ApiKey, ShouldEqual, "test-key")
				So(repo.client, ShouldNotBeNil)
			})
		})

		Convey("When newOpenAIUpstreamWithClient is called", func() {
			mockClient := &mockHTTPClient{}
			repo, err := newOpenAIUpstreamWithClient(config, mockClient, log.DefaultLogger)

			Convey("Then it should return a new upstream with the custom client", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				So(repo.config.BaseUrl, ShouldEqual, "https://api.openai.com/v1/")
				So(repo.config.ApiKey, ShouldEqual, "test-key")
				So(repo.client, ShouldNotBeNil)
			})
		})
	})
}
