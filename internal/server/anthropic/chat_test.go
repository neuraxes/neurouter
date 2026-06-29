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
	"bufio"
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/middleware"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/data/upstream/anthropic/mock"
)

type mockChatServer struct {
	v1.ChatServer
	chatFunc       func(ctx context.Context, req *v1.ChatReq) (*v1.ChatResp, error)
	chatStreamFunc func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error
}

func (m *mockChatServer) Chat(ctx context.Context, req *v1.ChatReq) (*v1.ChatResp, error) {
	if m.chatFunc != nil {
		return m.chatFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockChatServer) ChatStream(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
	if m.chatStreamFunc != nil {
		return m.chatStreamFunc(req, stream)
	}
	return nil
}

type mockResponseWriter struct {
	http.ResponseWriter
	ctx *mockHTTPContext
}

func (w *mockResponseWriter) Write(data []byte) (int, error) {
	return w.ctx.respBody.Write(data)
}

func (w *mockResponseWriter) Header() http.Header {
	return w.ctx.headers
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.ctx.statusCode = statusCode
}

func (w *mockResponseWriter) Flush() {}

// mockHTTPContext is a fake http.Context for testing handleMessageCompletion.
type mockHTTPContext struct {
	kratoshttp.Context
	req        *http.Request
	statusCode int
	headers    http.Header
	respBody   bytes.Buffer
}

func newMockHTTPContext(req *http.Request) *mockHTTPContext {
	return &mockHTTPContext{
		req:     req,
		headers: make(http.Header),
	}
}

func (t *mockHTTPContext) Request() *http.Request {
	return t.req
}

func (t *mockHTTPContext) Response() kratoshttp.ResponseWriter {
	return &mockResponseWriter{ctx: t}
}

func (t *mockHTTPContext) Middleware(handler middleware.Handler) middleware.Handler {
	return handler
}

func (t *mockHTTPContext) Blob(code int, contentType string, data []byte) error {
	t.statusCode = code
	t.headers.Set("Content-Type", contentType)
	t.respBody.Write(data)
	return nil
}

func TestChat(t *testing.T) {
	Convey("Given the Anthropic conversion fixtures", t, func() {
		for _, fixture := range mock.Fixtures {
			if fixture.Stream {
				continue
			}

			Convey("When handling the "+fixture.Name+" non-stream request", func() {
				srv := &Server{
					chatSvc: &mockChatServer{
						chatFunc: func(ctx context.Context, req *v1.ChatReq) (*v1.ChatResp, error) {
							assertChatReqEquality(req, fixture.ChatReq)
							return fixture.ChatResp, nil
						},
					},
				}

				req, err := http.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(fixture.Request))
				So(err, ShouldBeNil)
				ctx := newMockHTTPContext(req)

				err = srv.handleMessageCompletion(ctx)
				So(err, ShouldBeNil)

				Convey("Then it should return the fixture response as Anthropic JSON", func() {
					So(ctx.statusCode, ShouldEqual, http.StatusOK)
					So(ctx.headers.Get("Content-Type"), ShouldEqual, "application/json")
					// TODO: Compare the actual response body with the expected one.
				})
			})
		}
	})
}

type sseEvent struct {
	event string
	data  string
}

func parseSSEEvents(sseData string) []sseEvent {
	var events []sseEvent
	scanner := bufio.NewScanner(strings.NewReader(sseData))

	var currentEvent, currentData string
	for scanner.Scan() {
		line := scanner.Text()
		if after, ok := strings.CutPrefix(line, "event: "); ok {
			currentEvent = after
		} else if after, ok := strings.CutPrefix(line, "data: "); ok {
			currentData = after
		} else if line == "" && currentEvent != "" {
			events = append(events, sseEvent{event: currentEvent, data: currentData})
			currentEvent = ""
			currentData = ""
		}
	}
	if currentEvent != "" {
		events = append(events, sseEvent{event: currentEvent, data: currentData})
	}
	return events
}

func TestChatStream(t *testing.T) {
	Convey("Given the Anthropic conversion fixtures", t, func() {
		for _, fixture := range mock.Fixtures {
			if !fixture.Stream {
				continue
			}

			Convey("When handling the "+fixture.Name+" stream request", func() {
				srv := &Server{
					chatSvc: &mockChatServer{
						chatStreamFunc: func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
							assertChatReqEquality(req, fixture.ChatReq)
							for _, event := range fixture.ChatEvents {
								if err := stream.Send(event); err != nil {
									return err
								}
							}
							return nil
						},
					},
				}

				req, err := http.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(fixture.Request))
				So(err, ShouldBeNil)
				req.Header.Set("Accept", "text/event-stream")
				ctx := newMockHTTPContext(req)

				err = srv.handleMessageCompletion(ctx)
				So(err, ShouldBeNil)

				Convey("Then it should return the fixture events as Anthropic SSE", func() {
					So(ctx.headers.Get("Content-Type"), ShouldEqual, "text/event-stream")
					So(ctx.headers.Get("Cache-Control"), ShouldEqual, "no-cache")
					So(ctx.headers.Get("Connection"), ShouldEqual, "keep-alive")
					parseSSEEvents(ctx.respBody.String())
					// TODO: Compare the actual response body with the expected one.
				})
			})
		}
	})
}

func assertChatReqEquality(actual, expected *v1.ChatReq) {
	actualClone := proto.Clone(actual).(*v1.ChatReq)
	expectedClone := proto.Clone(expected).(*v1.ChatReq)
	actualClone.Id = ""
	expectedClone.Id = ""
	normalizeToolInputSchemas(actualClone)
	normalizeToolInputSchemas(expectedClone)
	So(proto.Equal(actualClone, expectedClone), ShouldBeTrue)
}

// normalizeToolInputSchemas removes the "additionalProperties" field from the
// input schema since it is not supported by the Anthropic SDK.
func normalizeToolInputSchemas(req *v1.ChatReq) {
	for _, tool := range req.GetTools() {
		function := tool.GetFunction()
		if function == nil || function.GetInputSchema() == nil {
			continue
		}
		delete(function.GetInputSchema().GetFields(), "additionalProperties")
	}
}
