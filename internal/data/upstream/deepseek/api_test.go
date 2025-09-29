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
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"

	"github.com/neuraxes/neurouter/internal/conf"
)

// mockHTTPClient is a mock implementation of the httpClient interface for testing.
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
    "id": "1c416330-2dca-4478-a1ac-1257d6512c7d",
    "object": "chat.completion",
    "created": 1758977545,
    "model": "deepseek-reasoner",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "Hello! How can I help you today?",
                "reasoning_content": "Hmm, the user just said \"Hello!\" with an exclamation mark, so they seem cheerful and friendly. This is a simple greeting, so no complex analysis needed. \n\nI should mirror their friendly tone while keeping it warm and professional. A straightforward welcoming response would work best here - acknowledge the greeting, express readiness to help, and leave the conversation open-ended for them to steer. \n\nNo need to overthink this. A simple \"Hello!\" in return, followed by a standard offer of assistance, covers all bases. The exclamation mark matches their energy level appropriately."
            },
            "logprobs": null,
            "finish_reason": "stop"
        }
    ],
    "usage": {
        "prompt_tokens": 12,
        "completion_tokens": 9,
        "total_tokens": 21,
        "prompt_tokens_details": {
            "cached_tokens": 0
        },
        "prompt_cache_hit_tokens": 0,
        "prompt_cache_miss_tokens": 12
    },
    "system_fingerprint": "fp_f253fc19d1_prod0820_fp8_kvcache"
}`

const mockChatCompletionStreamResp = `data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"role":"assistant","content":null,"reasoning_content":""},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":null,"reasoning_content":"Hmm, the user just said \"Hello!\" with an exclamation mark, which suggests a friendly and enthusiastic tone."},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":null,"reasoning_content":" This is a simple greeting, so no complex analysis is needed."},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":null,"reasoning_content":" \n\n"},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":null,"reasoning_content":"The best response would be to mirror their energy with an equally warm and welcoming reply"},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":null,"reasoning_content":"."},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"Hello","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"!","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" 😊","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" It","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"'s","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" wonderful","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" to","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" meet","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" you","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"!","reasoning_content":null},"logprobs":null,"finish_reason":null}]}

data: {"id":"f802c271-d4da-4773-b8b3-b1e4dfbf14f3","object":"chat.completion.chunk","created":1759299469,"model":"deepseek-reasoner","system_fingerprint":"fp_ffc7281d48_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"","reasoning_content":null},"logprobs":null,"finish_reason":"stop"}],"usage":{"prompt_tokens":6,"completion_tokens":169,"total_tokens":175,"prompt_tokens_details":{"cached_tokens":0},"completion_tokens_details":{"reasoning_tokens":117},"prompt_cache_hit_tokens":1,"prompt_cache_miss_tokens":5}}

data: [DONE]

`

const mockChatCompletionWithToolResp = `{
    "id": "fd8b3b63-8112-403f-aa52-57a46f320424",
    "object": "chat.completion",
    "created": 1759074553,
    "model": "deepseek-chat",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "I'll check the current weather in Tokyo for you.",
                "tool_calls": [
                    {
                        "index": 0,
                        "id": "call_00_wVp0FIPEgzSN4qfP502y9zG8",
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
        "completion_tokens": 25,
        "total_tokens": 186,
        "prompt_tokens_details": {
            "cached_tokens": 0
        },
        "prompt_cache_hit_tokens": 0,
        "prompt_cache_miss_tokens": 161
    },
    "system_fingerprint": "fp_8333852bec_prod0820_fp8_kvcache"
}`

const mockChatCompletionStreamWithToolResp = `data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"role":"assistant","content":""},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"I"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"'ll"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" check"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" the"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" current"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" weather"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" in"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" Tokyo"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" for"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" you"},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"."},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_00_Sq0n5HVeFHkS1mBrIGbwBQcK","type":"function","function":{"name":"get_weather","arguments":""}}]},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\""}}]},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"location"}}]},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\":"}}]},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":" \""}}]},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"Tokyo"}}]},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"}"}}]},"logprobs":null,"finish_reason":null}]}

data: {"id":"1986f9e8-0b5d-4331-88da-d2d2d8dd7a70","object":"chat.completion.chunk","created":1759075956,"model":"deepseek-chat","system_fingerprint":"fp_8333852bec_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":""},"logprobs":null,"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":161,"completion_tokens":25,"total_tokens":186,"prompt_tokens_details":{"cached_tokens":128},"prompt_cache_hit_tokens":128,"prompt_cache_miss_tokens":33}}

data: [DONE]

`

const mockChatCompletionError = `{
    "error": {
        "message": "Model Not Exist",
        "type": "invalid_request_error",
        "param": null,
        "code": "invalid_request_error"
    }
}`

func TestCreateChatCompletion(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client", t, func() {
		repo := &upstream{
			config: &conf.DeepSeekConfig{
				BaseUrl: "http://localhost",
				ApiKey:  "test-key",
			},
			log: log.NewHelper(log.DefaultLogger),
		}
		req := &ChatRequest{
			Model: "deepseek-chat",
			Messages: []*Message{
				{Role: "user", Content: "Hello!"},
			},
			Stream: false,
		}

		Convey("When CreateChatCompletion is called and the request is successful", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					So(httpReq.URL.String(), ShouldEqual, "http://localhost/chat/completions")
					So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
					So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

					body, _ := io.ReadAll(httpReq.Body)
					So(gjson.Get(string(body), "model").String(), ShouldEqual, "deepseek-chat")
					So(gjson.Get(string(body), "messages.0.role").String(), ShouldEqual, "user")
					So(gjson.Get(string(body), "messages.0.content").String(), ShouldEqual, "Hello!")
					So(gjson.Get(string(body), "stream").Bool(), ShouldBeFalse)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(mockChatCompletionResp)),
					}, nil
				},
			}

			resp, err := repo.CreateChatCompletion(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.ID, ShouldEqual, "1c416330-2dca-4478-a1ac-1257d6512c7d")
				So(resp.Object, ShouldEqual, "chat.completion")
				So(resp.Created, ShouldEqual, 1758977545)
				So(resp.Model, ShouldEqual, "deepseek-reasoner")
				So(resp.Choices, ShouldHaveLength, 1)
				So(resp.Choices[0].Index, ShouldEqual, 0)
				So(resp.Choices[0].Message.Role, ShouldEqual, "assistant")
				So(resp.Choices[0].Message.Content, ShouldEqual, "Hello! How can I help you today?")
				So(resp.Choices[0].Message.ReasoningContent, ShouldEqual, "Hmm, the user just said \"Hello!\" with an exclamation mark, so they seem cheerful and friendly. This is a simple greeting, so no complex analysis needed. \n\nI should mirror their friendly tone while keeping it warm and professional. A straightforward welcoming response would work best here - acknowledge the greeting, express readiness to help, and leave the conversation open-ended for them to steer. \n\nNo need to overthink this. A simple \"Hello!\" in return, followed by a standard offer of assistance, covers all bases. The exclamation mark matches their energy level appropriately.")
				So(resp.Choices[0].LogProbs, ShouldBeNil)
				So(resp.Choices[0].FinishReason, ShouldEqual, "stop")
				So(resp.Usage.PromptTokens, ShouldEqual, 12)
				So(resp.Usage.CompletionTokens, ShouldEqual, 9)
				So(resp.Usage.TotalTokens, ShouldEqual, 21)
				So(resp.Usage.PromptCacheHitTokens, ShouldEqual, 0)
				So(resp.Usage.PromptCacheMissTokens, ShouldEqual, 12)
				So(resp.SystemFingerprint, ShouldEqual, "fp_f253fc19d1_prod0820_fp8_kvcache")
			})
		})

		Convey("When the API returns an error", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(strings.NewReader(mockChatCompletionError)),
					}, nil
				},
			}

			_, err := repo.CreateChatCompletion(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "Model Not Exist")
			})
		})

		Convey("When the HTTP request fails", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			}

			_, err := repo.CreateChatCompletion(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})

		Convey("When decoding the response fails", func() {
			mockClient := &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
					}, nil
				},
			}
			repo.client = mockClient

			_, err := repo.CreateChatCompletion(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to decode response")
			})
		})
	})
}

func TestCreateChatCompletionStream(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client for streaming", t, func() {
		repo := &upstream{
			config: &conf.DeepSeekConfig{
				BaseUrl: "http://localhost",
				ApiKey:  "test-key",
			},
			log: log.NewHelper(log.DefaultLogger),
		}
		req := &ChatRequest{
			Model: "deepseek-chat",
			Messages: []*Message{
				{Role: "user", Content: "Hello!"},
			},
			Stream: true,
		}

		Convey("When CreateChatCompletionStream is called and the request is successful", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					So(httpReq.URL.String(), ShouldEqual, "http://localhost/chat/completions")
					So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
					So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")
					So(httpReq.Header.Get("Accept"), ShouldEqual, "text/event-stream")

					body, _ := io.ReadAll(httpReq.Body)
					So(gjson.Get(string(body), "model").String(), ShouldEqual, "deepseek-chat")
					So(gjson.Get(string(body), "messages.0.role").String(), ShouldEqual, "user")
					So(gjson.Get(string(body), "messages.0.content").String(), ShouldEqual, "Hello!")
					So(gjson.Get(string(body), "stream").Bool(), ShouldBeTrue)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(mockChatCompletionStreamResp)),
					}, nil
				},
			}

			resp, err := repo.CreateChatCompletionStream(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				body, readErr := io.ReadAll(resp.Body)
				So(readErr, ShouldBeNil)
				So(string(body), ShouldEqual, mockChatCompletionStreamResp)
				So(resp.Body.Close(), ShouldBeNil)
			})
		})

		Convey("When the API returns an error", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(strings.NewReader(mockChatCompletionError)),
					}, nil
				},
			}

			_, err := repo.CreateChatCompletionStream(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "Model Not Exist")
			})
		})

		Convey("When the HTTP request fails", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			}

			_, err := repo.CreateChatCompletionStream(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}

func TestCreateChatCompletionWithToolCalls(t *testing.T) {
	Convey("Given an upstream with a mock HTTP client and a tool-enabled request", t, func() {
		repo := &upstream{
			config: &conf.DeepSeekConfig{
				BaseUrl: "http://localhost",
				ApiKey:  "test-key",
			},
			log: log.NewHelper(log.DefaultLogger),
		}

		// Build a request that includes tools
		req := &ChatRequest{
			Model: "deepseek-chat",
			Messages: []*Message{
				{Role: "user", Content: "What is the weather in Tokyo?"},
			},
			Tools: []*Tool{
				{
					Type: "function",
					Function: &FunctionDefinition{
						Name:        "get_weather",
						Description: "Get the current weather for a city",
						Parameters: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"location": map[string]any{
									"type":        "string",
									"description": "City name",
								},
							},
							"required": []string{"location"},
						},
					},
				},
			},
		}

		Convey("When CreateChatCompletion is called and the API returns a tool_calls response", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
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
				},
			}

			resp, err := repo.CreateChatCompletion(context.Background(), req)

			Convey("Then it should decode tool_calls correctly", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.ID, ShouldEqual, "fd8b3b63-8112-403f-aa52-57a46f320424")
				So(resp.Object, ShouldEqual, "chat.completion")
				So(resp.Created, ShouldEqual, 1759074553)
				So(resp.Model, ShouldEqual, "deepseek-chat")
				So(resp.Choices, ShouldHaveLength, 1)
				choice := resp.Choices[0]
				So(choice.Index, ShouldEqual, 0)
				So(choice.Message.Role, ShouldEqual, "assistant")
				So(choice.Message.Content, ShouldEqual, "I'll check the current weather in Tokyo for you.")
				So(choice.LogProbs, ShouldBeNil)
				So(choice.FinishReason, ShouldEqual, "tool_calls")
				So(choice.Message.ToolCalls, ShouldHaveLength, 1)
				tc := choice.Message.ToolCalls[0]
				So(tc.ID, ShouldStartWith, "call_00_wVp0FIPEgzSN4qfP502y9zG8")
				So(tc.Type, ShouldEqual, "function")
				So(tc.Function, ShouldNotBeNil)
				So(tc.Function.Name, ShouldEqual, "get_weather")
				So(tc.Function.Arguments, ShouldEqual, "{\"location\": \"Tokyo\"}")
			})
		})
	})
}

func TestCreateChatCompletionStreamWithToolCalls(t *testing.T) {
	Convey("Given an upstream with a mock HTTP client for streaming with tool calls", t, func() {
		repo := &upstream{
			config: &conf.DeepSeekConfig{
				BaseUrl: "http://localhost",
				ApiKey:  "test-key",
			},
			log: log.NewHelper(log.DefaultLogger),
		}

		// Build a request that includes tools and enables streaming
		req := &ChatRequest{
			Model: "deepseek-chat",
			Messages: []*Message{
				{Role: "user", Content: "What is the weather in Tokyo?"},
			},
			Tools: []*Tool{
				{
					Type: "function",
					Function: &FunctionDefinition{
						Name:        "get_weather",
						Description: "Get the current weather for a city",
						Parameters: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"location": map[string]any{
									"type":        "string",
									"description": "City name",
								},
							},
							"required": []string{"location"},
						},
					},
				},
			},
			Stream: true,
		}

		Convey("When CreateChatCompletionStream is called and the API returns a streaming response with tool_calls", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					So(httpReq.URL.String(), ShouldEqual, "http://localhost/chat/completions")
					So(httpReq.Header.Get("Authorization"), ShouldEqual, "Bearer test-key")
					So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")
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
				},
			}

			resp, err := repo.CreateChatCompletionStream(context.Background(), req)

			Convey("Then it should return a valid streaming response with tool calls", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				body, readErr := io.ReadAll(resp.Body)
				So(readErr, ShouldBeNil)
				So(string(body), ShouldEqual, mockChatCompletionStreamWithToolResp)
				So(resp.Body.Close(), ShouldBeNil)
			})
		})

		Convey("When the API returns an error for streaming with tools", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(strings.NewReader(mockChatCompletionError)),
					}, nil
				},
			}

			_, err := repo.CreateChatCompletionStream(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "Model Not Exist")
			})
		})

		Convey("When the HTTP request fails for streaming with tools", func() {
			repo.client = &mockHTTPClient{
				DoFunc: func(httpReq *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			}

			_, err := repo.CreateChatCompletionStream(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}
