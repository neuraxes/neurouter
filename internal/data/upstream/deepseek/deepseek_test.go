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
