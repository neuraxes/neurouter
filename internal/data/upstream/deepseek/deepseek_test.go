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

package deepseek

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
)

func TestNewDeepSeekChatRepo(t *testing.T) {
	Convey("Given a configuration and logger", t, func() {
		config := &conf.DeepSeekConfig{BaseUrl: "http://localhost:8080/"}
		logger := log.DefaultLogger

		Convey("When NewDeepSeekChatRepo is called", func() {
			repo, err := NewDeepSeekChatRepo(config, logger)

			Convey("Then it should return a new upstream and no error", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				chatRepo, ok := repo.(*upstream)
				So(ok, ShouldBeTrue)
				So(chatRepo.config.BaseUrl, ShouldEqual, "http://localhost:8080")
				So(chatRepo.client, ShouldEqual, http.DefaultClient)
			})
		})

		Convey("When NewDeepSeekChatRepoWithClient is called", func() {
			mockClient := &mockHTTPClient{}
			repo, err := NewDeepSeekChatRepoWithClient(config, logger, mockClient)

			Convey("Then it should return a new upstream with the custom client", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				chatRepo, ok := repo.(*upstream)
				So(ok, ShouldBeTrue)
				So(chatRepo.client, ShouldEqual, mockClient)
			})
		})
	})
}

func TestChat(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client", t, func() {
		config := &conf.DeepSeekConfig{
			BaseUrl: "http://localhost",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := NewDeepSeekChatRepoWithClient(config, log.DefaultLogger, mockClient)
		So(err, ShouldBeNil)

		chatRepo, ok := repo.(*upstream)
		So(ok, ShouldBeTrue)

		req := &entity.ChatReq{
			Id:    "test-req-id",
			Model: "deepseek-chat",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Hello"}}}},
			},
		}

		Convey("When Chat is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "http://localhost/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				body, _ := io.ReadAll(httpReq.Body)
				So(gjson.Get(string(body), "model").String(), ShouldEqual, "deepseek-chat")
				So(gjson.Get(string(body), "messages.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(string(body), "messages.0.content").String(), ShouldEqual, "Hello")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(mockChatCompletionResp)),
				}, nil
			}

			resp, err := chatRepo.Chat(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "test-req-id")
				So(resp.Model, ShouldEqual, "deepseek-chat")
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Id, ShouldEqual, "1c416330-2dca-4478-a1ac-1257d6512c7d")
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "Hello! How can I help you today?")
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage.PromptTokens, ShouldEqual, 12)
				So(resp.Statistics.Usage.CompletionTokens, ShouldEqual, 9)
			})
		})

		Convey("When the API call fails", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			_, err := chatRepo.Chat(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}

var mockChatStreamResp = []*entity.ChatResp{
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "Hello"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "!"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " How"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " can"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " I"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " help"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " you"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " today"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "?"},
			}},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "71b67039-9a15-4b3d-be53-2d1ce5847f2f",
			Role: v1.Role_MODEL,
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     12,
				CompletionTokens: 9,
			},
		},
	},
}

func TestChatStream(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client for streaming", t, func() {
		config := &conf.DeepSeekConfig{
			BaseUrl: "http://localhost",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := NewDeepSeekChatRepoWithClient(config, log.DefaultLogger, mockClient)
		So(err, ShouldBeNil)

		chatRepo, ok := repo.(*upstream)
		So(ok, ShouldBeTrue)

		req := &entity.ChatReq{
			Id:    "test-stream-req-id",
			Model: "deepseek-chat",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Hello"}}}},
			},
		}

		Convey("When ChatStream is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "http://localhost/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")

				body, _ := io.ReadAll(httpReq.Body)
				So(gjson.Get(string(body), "model").String(), ShouldEqual, "deepseek-chat")
				So(gjson.Get(string(body), "messages.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(string(body), "messages.0.content").String(), ShouldEqual, "Hello")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(mockChatCompletionStreamResp)),
				}, nil
			}

			streamClient, err := chatRepo.ChatStream(context.Background(), req)

			Convey("Then it should return a stream client and no error", func() {
				So(err, ShouldBeNil)
				So(streamClient, ShouldNotBeNil)
				defer streamClient.Close()

				var responses []*entity.ChatResp
				for {
					resp, err := streamClient.Recv()
					if err == io.EOF {
						break
					}
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					So(resp.Id, ShouldEqual, "test-stream-req-id")
					So(resp.Model, ShouldEqual, "deepseek-chat")

					responses = append(responses, resp)
				}

				for i, resp := range responses {
					So(proto.Equal(resp, mockChatStreamResp[i]), ShouldBeTrue)
				}
			})
		})

		Convey("When the API call fails", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			_, err := chatRepo.ChatStream(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}

func TestChatWithToolCall(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client for tool calls", t, func() {
		config := &conf.DeepSeekConfig{
			BaseUrl: "http://localhost",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := NewDeepSeekChatRepoWithClient(config, log.DefaultLogger, mockClient)
		So(err, ShouldBeNil)

		chatRepo, ok := repo.(*upstream)
		So(ok, ShouldBeTrue)

		req := &entity.ChatReq{
			Id:    "test-tool-call-req-id",
			Model: "deepseek-chat",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "What is the weather in Tokyo?"}}}},
			},
			Tools: []*v1.Tool{
				{
					Tool: &v1.Tool_Function_{
						Function: &v1.Tool_Function{
							Name:        "get_weather",
							Description: "Get the current weather for a city",
							Parameters: &v1.Schema{
								Type: v1.Schema_TYPE_OBJECT,
								Properties: map[string]*v1.Schema{
									"location": {
										Type:        v1.Schema_TYPE_STRING,
										Description: "City name",
									},
								},
								Required: []string{"location"},
							},
						},
					},
				},
			},
		}

		Convey("When Chat is called with tools and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "http://localhost/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				body, _ := io.ReadAll(httpReq.Body)
				So(gjson.Get(string(body), "model").String(), ShouldEqual, "deepseek-chat")
				So(gjson.Get(string(body), "messages.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(string(body), "messages.0.content").String(), ShouldEqual, "What is the weather in Tokyo?")
				So(gjson.Get(string(body), "tools.0.type").String(), ShouldEqual, "function")
				So(gjson.Get(string(body), "tools.0.function.name").String(), ShouldEqual, "get_weather")
				So(gjson.Get(string(body), "tools.0.function.description").String(), ShouldEqual, "Get the current weather for a city")
				So(gjson.Get(string(body), "tools.0.function.parameters.type").String(), ShouldEqual, "object")
				So(gjson.Get(string(body), "tools.0.function.parameters.properties.location.type").String(), ShouldEqual, "string")
				So(gjson.Get(string(body), "tools.0.function.parameters.properties.location.description").String(), ShouldEqual, "City name")
				So(gjson.Get(string(body), "tools.0.function.parameters.required.0").String(), ShouldEqual, "location")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(mockChatCompletionWithToolResp)),
				}, nil
			}

			resp, err := chatRepo.Chat(context.Background(), req)

			Convey("Then it should return a valid response with tool calls and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "test-tool-call-req-id")
				So(resp.Model, ShouldEqual, "deepseek-chat")
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Id, ShouldEqual, "fd8b3b63-8112-403f-aa52-57a46f320424")
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(len(resp.Message.Contents), ShouldEqual, 2)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "I'll check the current weather in Tokyo for you.")
				So(resp.Message.Contents[1].GetFunctionCall(), ShouldNotBeNil)
				fc := resp.Message.Contents[1].GetFunctionCall()
				So(fc.Id, ShouldEqual, "call_00_wVp0FIPEgzSN4qfP502y9zG8")
				So(fc.Name, ShouldEqual, "get_weather")
				So(fc.Arguments, ShouldEqual, `{"location": "Tokyo"}`)
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage.PromptTokens, ShouldEqual, 161)
				So(resp.Statistics.Usage.CompletionTokens, ShouldEqual, 25)
			})
		})
	})
}

var mockChatStreamWithToolCallResp = []*entity.ChatResp{
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "I"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "'ll"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " check"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " the"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " current"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " weather"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " in"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " Tokyo"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " for"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " you"}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:       "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "."}}},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Id:   "call_00_Sq0n5HVeFHkS1mBrIGbwBQcK",
							Name: "get_weather",
						},
					},
				},
			},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Arguments: `{"`,
						},
					},
				},
			},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Arguments: `location`,
						},
					},
				},
			},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Arguments: `":`,
						},
					},
				},
			},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Arguments: ` "`,
						},
					},
				},
			},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Arguments: `Tokyo`,
						},
					},
				},
			},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Arguments: `"}`,
						},
					},
				},
			},
		},
	},
	{
		Id:    "test-stream-tool-call-req-id",
		Model: "deepseek-chat",
		Message: &v1.Message{
			Id:   "1986f9e8-0b5d-4331-88da-d2d2d8dd7a70",
			Role: v1.Role_MODEL,
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:       161,
				CompletionTokens:   25,
				CachedPromptTokens: 128,
			},
		},
	},
}

func TestChatStreamWithToolCall(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client for streaming tool calls", t, func() {
		config := &conf.DeepSeekConfig{
			BaseUrl: "http://localhost",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := NewDeepSeekChatRepoWithClient(config, log.DefaultLogger, mockClient)
		So(err, ShouldBeNil)

		chatRepo, ok := repo.(*upstream)
		So(ok, ShouldBeTrue)

		req := &entity.ChatReq{
			Id:    "test-stream-tool-call-req-id",
			Model: "deepseek-chat",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "What is the weather in Tokyo?"}}}},
			},
			Tools: []*v1.Tool{
				{
					Tool: &v1.Tool_Function_{
						Function: &v1.Tool_Function{
							Name:        "get_weather",
							Description: "Get the current weather for a city",
							Parameters: &v1.Schema{
								Type: v1.Schema_TYPE_OBJECT,
								Properties: map[string]*v1.Schema{
									"location": {
										Type:        v1.Schema_TYPE_STRING,
										Description: "City name",
									},
								},
								Required: []string{"location"},
							},
						},
					},
				},
			},
		}

		Convey("When ChatStream is called with tools and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "http://localhost/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
				So(httpReq.Header.Get("Accept"), ShouldEqual, "text/event-stream")

				body, _ := io.ReadAll(httpReq.Body)
				So(gjson.Get(string(body), "model").String(), ShouldEqual, "deepseek-chat")
				So(gjson.Get(string(body), "messages.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(string(body), "messages.0.content").String(), ShouldEqual, "What is the weather in Tokyo?")
				So(gjson.Get(string(body), "tools.0.type").String(), ShouldEqual, "function")
				So(gjson.Get(string(body), "tools.0.function.name").String(), ShouldEqual, "get_weather")
				So(gjson.Get(string(body), "tools.0.function.description").String(), ShouldEqual, "Get the current weather for a city")
				So(gjson.Get(string(body), "tools.0.function.parameters.type").String(), ShouldEqual, "object")
				So(gjson.Get(string(body), "tools.0.function.parameters.properties.location.type").String(), ShouldEqual, "string")
				So(gjson.Get(string(body), "tools.0.function.parameters.properties.location.description").String(), ShouldEqual, "City name")
				So(gjson.Get(string(body), "tools.0.function.parameters.required.0").String(), ShouldEqual, "location")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(mockChatCompletionStreamWithToolResp)),
				}, nil
			}

			streamClient, err := chatRepo.ChatStream(context.Background(), req)

			Convey("Then it should return a stream client and no error", func() {
				So(err, ShouldBeNil)
				So(streamClient, ShouldNotBeNil)
				defer streamClient.Close()

				var responses []*entity.ChatResp
				for {
					resp, err := streamClient.Recv()
					if err == io.EOF {
						break
					}
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					responses = append(responses, resp)
				}

				So(len(responses), ShouldEqual, len(mockChatStreamWithToolCallResp))
				for i, resp := range responses {
					if !proto.Equal(resp, mockChatStreamWithToolCallResp[i]) {
						t.Errorf("Response %d does not match expected.\nGot: %+v\nExpected: %+v", i, resp, mockChatStreamWithToolCallResp[i])
					}
					So(proto.Equal(resp, mockChatStreamWithToolCallResp[i]), ShouldBeTrue)
				}
			})
		})
	})
}
