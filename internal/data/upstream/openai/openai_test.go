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
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
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

const mockChatCompletionResp = `{
    "id": "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
    "object": "chat.completion",
    "created": 1698892410,
    "model": "gpt-4o",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "Hello! How can I help you today?"
            },
            "finish_reason": "stop"
        }
    ],
    "usage": {
        "prompt_tokens": 12,
        "completion_tokens": 9,
        "total_tokens": 21
    },
    "system_fingerprint": "fp_50a4261de5"
}`

const mockChatCompletionStreamResp = `data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" How"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" can"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" I"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" help"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" you"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" today"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":"?"},"finish_reason":null}]}

data: {"id":"chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":""},"finish_reason":"stop"}],"usage":{"prompt_tokens":12,"completion_tokens":9}}

data: [DONE]

`

const mockChatCompletionWithToolResp = `{
    "id": "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
    "object": "chat.completion",
    "created": 1698892410,
    "model": "gpt-4o",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "I'll check the current weather in Tokyo for you.",
                "tool_calls": [
                    {
                        "id": "call_abc123def456",
                        "type": "function",
                        "function": {
                            "name": "get_weather",
                            "arguments": "{\"location\": \"Tokyo\"}"
                        }
                    }
                ]
            },
            "logprobs": null,
            "finish_reason": "tool_calls"
        }
    ],
    "usage": {
        "prompt_tokens": 161,
        "completion_tokens": 25
    },
    "system_fingerprint": "fp_50a4261de5"
}`

const mockChatCompletionStreamWithToolResp = `data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":"I"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":"'ll"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" check"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" the"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" current"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" weather"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" in"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" Tokyo"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" for"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":" you"},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":"."},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_abc123def456","type":"function","function":{"name":"get_weather","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"location"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\":"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":" \""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"Tokyo"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9","object":"chat.completion.chunk","created":1698892410,"model":"gpt-4o","system_fingerprint":"fp_50a4261de5","choices":[{"index":0,"delta":{"content":""},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":161,"completion_tokens":25}}

data: [DONE]

`

func TestNewOpenAIUpstream(t *testing.T) {
	Convey("Given a configuration and logger", t, func() {
		config := &conf.OpenAIConfig{
			BaseUrl: "https://api.openai.com/v1/",
			ApiKey:  "test-key",
		}
		logger := log.DefaultLogger

		Convey("When newOpenAIUpstream is called", func() {
			repo, err := newOpenAIUpstream(config, logger)

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
			repo, err := newOpenAIUpstreamWithClient(config, logger, mockClient)

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
		repo, err := newOpenAIUpstreamWithClient(config, log.DefaultLogger, mockClient)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-req-id",
			Model: "gpt-4o",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Hello"}}}},
			},
		}

		Convey("When Chat is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "https://api.openai.com/v1/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatCompletionResp)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P")
				So(resp.Model, ShouldEqual, "gpt-4o")
				So(resp.Message, ShouldNotBeNil)
				So(len(resp.Message.Id), ShouldEqual, 36)
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

			_, err := repo.Chat(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}

var mockChatStreamResp = []*entity.ChatResp{
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "Hello"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "!"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " How"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " can"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " I"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " help"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " you"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " today"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "?"},
			}},
		},
	},
	{
		Id:    "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P",
		Model: "gpt-4o",
		Message: &v1.Message{
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
		config := &conf.OpenAIConfig{
			BaseUrl: "https://api.openai.com/v1/",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := newOpenAIUpstreamWithClient(config, log.DefaultLogger, mockClient)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-stream-req-id",
			Model: "gpt-4o",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Hello"}}}},
			},
		}

		Convey("When ChatStream is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "https://api.openai.com/v1/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/event-stream"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatCompletionStreamResp)),
				}, nil
			}

			streamClient, err := repo.ChatStream(context.Background(), req)

			Convey("Then it should return a stream client and no error", func() {
				So(err, ShouldBeNil)
				So(streamClient, ShouldNotBeNil)
				defer streamClient.Close()

				var messageID string
				var responses []*entity.ChatResp
				for {
					resp, err := streamClient.Recv()
					if err == io.EOF {
						break
					}
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					So(resp.Id, ShouldEqual, "chatcmpl-8GHoQAJ3zN2DJYqOFiVysrMQJfe1P")
					So(resp.Model, ShouldEqual, "gpt-4o")

					messageID = resp.Message.Id
					responses = append(responses, resp)
				}

				// Set the dynamic message ID on the mock responses for comparison
				for _, mockResp := range mockChatStreamResp {
					mockResp.Message.Id = messageID
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

			streamClient, err := repo.ChatStream(context.Background(), req)

			Convey("Then it should return an error", func() {
				// OpenAI client always returns a stream client even on error
				So(err, ShouldBeNil)
				// The error is returned on Recv()
				resp, err := streamClient.Recv()
				So(resp, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}

func TestChatWithToolCall(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client for tool calls", t, func() {
		config := &conf.OpenAIConfig{
			BaseUrl: "https://api.openai.com/v1/",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := newOpenAIUpstreamWithClient(config, log.DefaultLogger, mockClient)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-tool-call-req-id",
			Model: "gpt-4o",
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
				So(httpReq.URL.String(), ShouldEqual, "https://api.openai.com/v1/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatCompletionWithToolResp)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), req)

			Convey("Then it should return a valid response with tool calls and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9")
				So(resp.Model, ShouldEqual, "gpt-4o")
				So(resp.Message, ShouldNotBeNil)
				So(len(resp.Message.Id), ShouldEqual, 36)
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(len(resp.Message.Contents), ShouldEqual, 2)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "I'll check the current weather in Tokyo for you.")
				So(resp.Message.Contents[1].GetFunctionCall(), ShouldNotBeNil)
				fc := resp.Message.Contents[1].GetFunctionCall()
				So(fc.Id, ShouldEqual, "call_abc123def456")
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
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "I"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "'ll"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " check"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " the"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " current"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " weather"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " in"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " Tokyo"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " for"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " you"}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "."}}},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Id:   "call_abc123def456",
							Name: "get_weather",
						},
					},
				},
			},
		},
	},
	{
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
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
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
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
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
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
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
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
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
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
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
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
		Id:    "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9",
		Model: "gpt-4o",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     161,
				CompletionTokens: 25,
			},
		},
	},
}

func TestChatStreamWithToolCall(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client for streaming tool calls", t, func() {
		config := &conf.OpenAIConfig{
			BaseUrl: "https://api.openai.com/v1/",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := newOpenAIUpstreamWithClient(config, log.DefaultLogger, mockClient)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-stream-tool-call-req-id",
			Model: "gpt-4o",
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
				So(httpReq.URL.String(), ShouldEqual, "https://api.openai.com/v1/chat/completions")
				So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/event-stream"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatCompletionStreamWithToolResp)),
				}, nil
			}

			streamClient, err := repo.ChatStream(context.Background(), req)

			Convey("Then it should return a stream client and no error", func() {
				So(err, ShouldBeNil)
				So(streamClient, ShouldNotBeNil)
				defer streamClient.Close()

				var messageID string
				var responses []*entity.ChatResp
				for {
					resp, err := streamClient.Recv()
					if err == io.EOF {
						break
					}
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					So(resp.Id, ShouldEqual, "chatcmpl-8aX1c5c7d8e9f0g1h2i3j4k5l6m7n9")
					So(resp.Model, ShouldEqual, "gpt-4o")

					messageID = resp.Message.Id
					responses = append(responses, resp)
				}

				So(len(responses), ShouldEqual, len(mockChatStreamWithToolCallResp))

				// Set the dynamic message ID on the mock responses for comparison
				for _, mockResp := range mockChatStreamWithToolCallResp {
					mockResp.Message.Id = messageID
				}

				for i, resp := range responses {
					So(proto.Equal(resp, mockChatStreamWithToolCallResp[i]), ShouldBeTrue)
				}
			})
		})
	})
}
