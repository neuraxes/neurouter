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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
	"github.com/neuraxes/neurouter/internal/data/upstream/anthropic/mock"
)

var mockTestConfig = &conf.AnthropicConfig{
	BaseUrl: "https://api.anthropic.com/",
	ApiKey:  "test-key",
}

// jsonMap unmarshals a JSON document into a map for order-independent comparison.
func jsonMap(data []byte) map[string]any {
	var m map[string]any
	So(json.Unmarshal(data, &m), ShouldBeNil)
	return m
}

// mockResponder builds a DoFunc that asserts the outgoing request envelope
// (method, endpoint, auth and content-type headers), records the request body
// into captured, and replies with the given response content type and body.
func mockResponder(responseContentType string, responseBody []byte, captured *[]byte) func(*http.Request) (*http.Response, error) {
	return func(httpReq *http.Request) (*http.Response, error) {
		So(httpReq.Method, ShouldEqual, http.MethodPost)
		So(httpReq.URL.String(), ShouldEqual, "https://api.anthropic.com/v1/messages")
		So(httpReq.Header.Get("x-api-key"), ShouldEqual, "test-key")
		So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

		body, err := io.ReadAll(httpReq.Body)
		So(err, ShouldBeNil)
		*captured = body

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{responseContentType}},
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}, nil
	}
}

func TestChat(t *testing.T) {
	Convey("Given the anthropic upstream conversion fixtures", t, func() {
		for _, fixture := range mock.Fixtures {
			if fixture.Stream {
				continue
			}

			Convey("When Chat runs the "+fixture.Name+" fixture", func() {
				mockClient := &mockHTTPClient{}
				repo, err := newAnthropicUpstreamWithClient(mockTestConfig, mockClient, log.DefaultLogger)
				So(err, ShouldBeNil)

				var capturedBody []byte
				mockClient.DoFunc = mockResponder("application/json", fixture.Response, &capturedBody)

				resp, err := repo.Chat(context.Background(), fixture.ChatReq)
				So(err, ShouldBeNil)

				Convey("Then the request body matches the fixture request", func() {
					So(jsonMap(capturedBody), ShouldResemble, jsonMap(fixture.Request))
				})

				Convey("Then the response converts to the expected ChatResp", func() {
					So(proto.Equal(resp, fixture.ChatResp), ShouldBeTrue)
				})
			})
		}

		Convey("When the API call fails", func() {
			mockClient := &mockHTTPClient{
				DoFunc: func(*http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			}
			repo, err := newAnthropicUpstreamWithClient(mockTestConfig, mockClient, log.DefaultLogger)
			So(err, ShouldBeNil)

			_, err = repo.Chat(context.Background(), mock.NonStreamMaxTokens.ChatReq)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}

func TestChatStream(t *testing.T) {
	Convey("Given the anthropic upstream conversion fixtures", t, func() {
		for _, fixture := range mock.Fixtures {
			if !fixture.Stream {
				continue
			}

			Convey("When ChatStream runs the "+fixture.Name+" fixture", func() {
				mockClient := &mockHTTPClient{}
				repo, err := newAnthropicUpstreamWithClient(mockTestConfig, mockClient, log.DefaultLogger)
				So(err, ShouldBeNil)

				var capturedBody []byte
				mockClient.DoFunc = mockResponder("text/event-stream", fixture.Response, &capturedBody)

				seq := repo.ChatStream(context.Background(), fixture.ChatReq)
				So(seq, ShouldNotBeNil)

				var events []*entity.ChatEvent
				for event, err := range seq {
					So(err, ShouldBeNil)
					So(event, ShouldNotBeNil)
					events = append(events, event)
				}

				Convey("Then the request body matches the fixture request", func() {
					So(jsonMap(capturedBody), ShouldResemble, jsonMap(fixture.Request))
				})

				Convey("Then the stream converts to the expected ChatEvents", func() {
					So(len(events), ShouldEqual, len(fixture.ChatEvents))
					for i := range events {
						So(proto.Equal(events[i], fixture.ChatEvents[i]), ShouldBeTrue)
					}
				})
			})
		}

		Convey("When the API call fails", func() {
			mockClient := &mockHTTPClient{
				DoFunc: func(*http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			}
			repo, err := newAnthropicUpstreamWithClient(mockTestConfig, mockClient, log.DefaultLogger)
			So(err, ShouldBeNil)

			seq := repo.ChatStream(context.Background(), mock.StreamThinkingText.ChatReq)
			So(seq, ShouldNotBeNil)

			Convey("Then it should return an error in the iterator", func() {
				for resp, err := range seq {
					So(resp, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "network error")
				}
			})
		})
	})
}
