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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	"github.com/neuraxes/neurouter/internal/biz/entity"
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
			repo, err := newOpenAIUpstream(config, log.DefaultLogger)

			Convey("Then it should return a new upstream and no error", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				upstream, ok := repo.(*upstream)
				So(ok, ShouldBeTrue)
				So(upstream.config.BaseUrl, ShouldEqual, "https://api.openai.com/v1/")
				So(upstream.config.ApiKey, ShouldEqual, "test-key")
				So(upstream.client, ShouldNotBeNil)
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

func TestChat(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client", t, func() {
		config := &conf.OpenAIConfig{
			BaseUrl: "https://api.openai.com/v1/",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := newOpenAIUpstreamWithClient(config, mockClient, log.DefaultLogger)
		So(err, ShouldBeNil)

		Convey("When Chat is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "https://api.openai.com/v1/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				body, err := io.ReadAll(httpReq.Body)
				So(err, ShouldBeNil)

				var reqMap map[string]any
				err = json.Unmarshal(body, &reqMap)
				So(err, ShouldBeNil)

				var expectedMap map[string]any
				err = json.Unmarshal([]byte(mockChatCompletionRequestBody), &expectedMap)
				So(err, ShouldBeNil)

				So(reqMap, ShouldResemble, expectedMap)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatCompletionResponseBody)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), mockChatReq)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(len(resp.Message.Id), ShouldEqual, 36)
				resp.Message.Id = "mock_message_id"
				So(proto.Equal(resp, mockChatResp), ShouldBeTrue)
			})
		})

		Convey("When the API call fails", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			_, err := repo.Chat(context.Background(), mockChatReq)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}

func TestChatStream(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client for streaming", t, func() {
		config := &conf.OpenAIConfig{
			BaseUrl: "https://api.openai.com/v1/",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := newOpenAIUpstreamWithClient(config, mockClient, log.DefaultLogger)
		So(err, ShouldBeNil)

		Convey("When ChatStream is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "https://api.openai.com/v1/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				body, err := io.ReadAll(httpReq.Body)
				So(err, ShouldBeNil)

				var reqMap map[string]any
				err = json.Unmarshal(body, &reqMap)
				So(err, ShouldBeNil)

				var expectedMap map[string]any
				err = json.Unmarshal([]byte(mockChatCompletionStreamRequestBody), &expectedMap)
				So(err, ShouldBeNil)

				So(reqMap, ShouldResemble, expectedMap)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/event-stream"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatCompletionStreamResponseBody)),
				}, nil
			}

			seq := repo.ChatStream(context.Background(), mockChatReq)

			Convey("Then it should return a sequence and no error", func() {
				So(seq, ShouldNotBeNil)

				var responses []*entity.ChatResp
				for resp, err := range seq {
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					So(resp.Message.Id, ShouldHaveLength, 36)

					resp.Message.Id = "mock_message_id"
					responses = append(responses, resp)
				}

				So(responses, ShouldHaveLength, len(mockChatStreamResp))

				for i, resp := range responses {
					if !proto.Equal(resp, mockChatStreamResp[i]) {
						fmt.Println("\n", resp.String(), "\n", mockChatStreamResp[i].String())
					}
					So(proto.Equal(resp, mockChatStreamResp[i]), ShouldBeTrue)
				}
			})
		})

		Convey("When the API call fails", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			seq := repo.ChatStream(context.Background(), mockChatReq)

			Convey("Then it should return an error", func() {
				So(seq, ShouldNotBeNil)

				for resp, err := range seq {
					So(resp, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "network error")
				}
			})
		})
	})
}
