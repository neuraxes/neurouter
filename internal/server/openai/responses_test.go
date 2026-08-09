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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/data/upstream/openai/mock"
)

type responsesTestChatServer struct {
	v1.ChatServer
	chatFunc       func(context.Context, *v1.ChatReq) (*v1.ChatResp, error)
	chatStreamFunc func(*v1.ChatReq, v1.Chat_ChatStreamServer) error
}

func (s *responsesTestChatServer) Chat(ctx context.Context, req *v1.ChatReq) (*v1.ChatResp, error) {
	return s.chatFunc(ctx, req)
}

func (s *responsesTestChatServer) ChatStream(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
	return s.chatStreamFunc(req, stream)
}

func TestHandleResponses(t *testing.T) {
	Convey("Given a non-stream Responses request", t, func() {
		server := &Server{chatSvc: &responsesTestChatServer{
			chatFunc: func(_ context.Context, req *v1.ChatReq) (*v1.ChatResp, error) {
				expected := proto.Clone(mock.ResponsesText.ChatReq).(*v1.ChatReq)
				expected.Id = ""
				So(proto.Equal(req, expected), ShouldBeTrue)
				return mock.ResponsesText.ChatResp, nil
			},
		}}
		req, err := http.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(mock.ResponsesText.Request))
		So(err, ShouldBeNil)

		httpCtx := newResponsesTestHTTPContext(req)
		So(server.handleResponses(httpCtx), ShouldBeNil)

		Convey("Then the handler returns a Responses JSON object", func() {
			So(httpCtx.statusCode, ShouldEqual, http.StatusOK)
			So(httpCtx.headers.Get("Content-Type"), ShouldEqual, "application/json")
			var response responsesResponse
			So(json.Unmarshal(httpCtx.body.Bytes(), &response), ShouldBeNil)
			So(response.ID, ShouldEqual, mock.ResponsesText.ChatResp.Message.Id)
			So(response.Object, ShouldEqual, "response")
			So(response.Status, ShouldEqual, "completed")
			So(response.Output, ShouldHaveLength, 2)
		})
	})

	Convey("Given a streaming Responses request", t, func() {
		requestBody := []byte(`{"model":"gpt-5","input":"hello","stream":true}`)
		server := &Server{chatSvc: &responsesTestChatServer{
			chatStreamFunc: func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
				So(req.Model, ShouldEqual, "gpt-5")
				So(req.Messages, ShouldHaveLength, 1)
				for _, event := range []*v1.ChatEvent{
					v1.NewChatEvent("request-1", v1.NewMessageStartEvent("resp-1", "gpt-5")),
					v1.NewChatEvent("request-1", v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_NORMAL)),
					v1.NewChatEvent("request-1", v1.NewContentDeltaTextEvent(0, "hello")),
					v1.NewChatEvent("request-1", v1.NewContentStopEvent(0)),
					v1.NewChatEvent("request-1", v1.NewMessageStopEvent(v1.ChatStatus_CHAT_COMPLETED)),
				} {
					if err := stream.Send(event); err != nil {
						return err
					}
				}
				return nil
			},
		}}
		req, err := http.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(requestBody))
		So(err, ShouldBeNil)

		httpCtx := newResponsesTestHTTPContext(req)
		So(server.handleResponses(httpCtx), ShouldBeNil)

		Convey("Then the handler returns Responses SSE without a DONE sentinel", func() {
			So(httpCtx.headers.Get("Content-Type"), ShouldEqual, "text/event-stream")
			So(httpCtx.body.String(), ShouldContainSubstring, "event: response.created")
			So(httpCtx.body.String(), ShouldContainSubstring, "event: response.completed")
			So(httpCtx.body.String(), ShouldNotContainSubstring, "[DONE]")
		})
	})
}
