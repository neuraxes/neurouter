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

package anthropic

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

func TestNewAnthropicUpstream(t *testing.T) {
	Convey("Given a configuration and logger", t, func() {
		config := &conf.AnthropicConfig{
			BaseUrl: "https://api.anthropic.com/",
			ApiKey:  "test-key",
		}

		Convey("When newAnthropicUpstream is called", func() {
			repo, err := newAnthropicUpstreamWithClient(config, nil, log.DefaultLogger)

			Convey("Then it should return a new upstream and no error", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				upstream, ok := repo.(*upstream)
				So(ok, ShouldBeTrue)
				So(upstream.config.BaseUrl, ShouldEqual, "https://api.anthropic.com/")
				So(upstream.config.ApiKey, ShouldEqual, "test-key")
				So(upstream.client, ShouldNotBeNil)
			})
		})

		Convey("When newAnthropicUpstreamWithClient is called", func() {
			mockClient := &mockHTTPClient{}
			repo, err := newAnthropicUpstreamWithClient(config, mockClient, log.DefaultLogger)

			Convey("Then it should return a new upstream with the custom client", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				upstream, ok := repo.(*upstream)
				So(ok, ShouldBeTrue)
				So(upstream.config.BaseUrl, ShouldEqual, "https://api.anthropic.com/")
				So(upstream.config.ApiKey, ShouldEqual, "test-key")
				So(upstream.client, ShouldNotBeNil)
			})
		})
	})
}
