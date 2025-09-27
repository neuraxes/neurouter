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
    "model": "deepseek-chat",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "Hello! How can I help you today?"
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

const mockChatCompletionError = `{
    "error": {
        "message": "Model Not Exist",
        "type": "invalid_request_error",
        "param": null,
        "code": "invalid_request_error"
    }
}`

const mockChatCompletionStreamResp = `data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"role":"assistant","content":""},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"Hello"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"!"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" How"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" can"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" I"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" help"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" you"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":" today"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":"?"},"logprobs":null,"finish_reason":null}]}

data: {"id":"71b67039-9a15-4b3d-be53-2d1ce5847f2f","object":"chat.completion.chunk","created":1758989659,"model":"deepseek-chat","system_fingerprint":"fp_f253fc19d1_prod0820_fp8_kvcache","choices":[{"index":0,"delta":{"content":""},"logprobs":null,"finish_reason":"stop"}],"usage":{"prompt_tokens":12,"completion_tokens":9,"total_tokens":21,"prompt_tokens_details":{"cached_tokens":0},"prompt_cache_hit_tokens":0,"prompt_cache_miss_tokens":12}}

data: [DONE]

`

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
				So(resp.Model, ShouldEqual, "deepseek-chat")
				So(resp.Choices, ShouldHaveLength, 1)
				So(resp.Choices[0].Index, ShouldEqual, 0)
				So(resp.Choices[0].Message.Role, ShouldEqual, "assistant")
				So(resp.Choices[0].Message.Content, ShouldEqual, "Hello! How can I help you today?")
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
