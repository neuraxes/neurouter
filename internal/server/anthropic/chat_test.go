package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-kratos/kratos/v2/middleware"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

// mockChatServer implements v1.ChatServer for testing.
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

func (t *mockHTTPContext) Result(code int, v any) error {
	t.statusCode = code
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	t.respBody.Write(data)
	return nil
}

func TestChat(t *testing.T) {
	Convey("Given a mock ChatServer for non-streaming requests", t, func() {
		mock := &mockChatServer{}

		Convey("When a non-streaming Anthropic request is sent", func() {
			mock.chatFunc = func(ctx context.Context, req *v1.ChatReq) (*v1.ChatResp, error) {
				So(proto.Equal(req, mockChatReq), ShouldBeTrue)
				return mockChatResp, nil
			}

			// Create HTTP context with mock request
			req, err := http.NewRequest("POST", "/v1/messages", strings.NewReader(mockMessagesRequestBody))
			So(err, ShouldBeNil)
			ctx := newMockHTTPContext(req)
			So(ctx, ShouldNotBeNil)

			// Call handleMessageCompletion
			err = handleMessageCompletion(ctx, mock)
			So(err, ShouldBeNil)

			// Verify the response
			var anthropicResp anthropic.Message
			err = json.Unmarshal(ctx.respBody.Bytes(), &anthropicResp)
			So(err, ShouldBeNil)
			anthropicRespJson, err := json.Marshal(anthropicResp)
			So(err, ShouldBeNil)

			expectedAnthropicResp := &anthropic.Message{}
			err = json.Unmarshal([]byte(mockMessagesResponseBody), expectedAnthropicResp)
			So(err, ShouldBeNil)
			expectedAnthropicRespJson, err := json.Marshal(expectedAnthropicResp)
			So(err, ShouldBeNil)

			So(anthropicRespJson, ShouldEqual, expectedAnthropicRespJson)
		})
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
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			currentData = strings.TrimPrefix(line, "data: ")
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
	Convey("Given a mock ChatServer for streaming requests", t, func() {
		mock := &mockChatServer{}

		Convey("When ChatStream is called and responses are sent", func() {
			mock.chatStreamFunc = func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
				So(proto.Equal(req, mockChatReq), ShouldBeTrue)
				for _, resp := range mockChatStreamResp {
					if err := stream.Send(resp); err != nil {
						return err
					}
				}
				return nil
			}

			// Create HTTP context with streaming request
			req, err := http.NewRequest("POST", "/v1/messages", strings.NewReader(mockMessagesRequestBody))
			So(err, ShouldBeNil)
			req.Header.Set("Accept", "text/event-stream")
			ctx := newMockHTTPContext(req)

			// Call handleMessageCompletion for streaming
			err = handleMessageCompletion(ctx, mock)
			So(err, ShouldBeNil)

			// Parse SSE events from the buffer
			events := parseSSEEvents(ctx.respBody.String())

			Convey("Then the correct number of SSE events should be emitted", func() {
				// Expected events:
				// 0: message_start
				// 1: content_block_start (thinking)
				// 2: content_block_delta (thinking_delta)
				// 3: content_block_delta (signature_delta)
				// 4: content_block_stop
				// 5: content_block_start (text)
				// 6: content_block_delta (text_delta: first chunk)
				// 7: content_block_delta (text_delta: second chunk)
				// 8: content_block_stop
				// 9: content_block_start (tool_use)
				// 10: content_block_delta (input_json_delta: empty)
				// 11: content_block_delta (input_json_delta: 1st chunk)
				// 12: content_block_delta (input_json_delta: 2nd chunk)
				// 13: content_block_delta (input_json_delta: 3rd chunk)
				// 14: content_block_delta (input_json_delta: 4th chunk)
				// 15: content_block_stop
				// 16: message_delta
				// 17: message_stop
				So(len(events), ShouldEqual, 18)
			})

			Convey("Then message_start should contain the message ID and role", func() {
				So(events[0].event, ShouldEqual, "message_start")
				var ev anthropic.MessageStartEvent
				json.Unmarshal([]byte(events[0].data), &ev)
				So(ev.Message.ID, ShouldEqual, "msg_016m3rsWB3U7eYBEKjTRSruv")
				So(string(ev.Message.Role), ShouldEqual, "assistant")
				So(string(ev.Message.Model), ShouldEqual, "claude-haiku-4-5-20251001")
				So(ev.Message.Usage.InputTokens, ShouldEqual, 840)
			})

			Convey("Then thinking events should be emitted correctly", func() {
				// content_block_start for thinking
				So(events[1].event, ShouldEqual, "content_block_start")
				var cbStart anthropic.ContentBlockStartEvent
				json.Unmarshal([]byte(events[1].data), &cbStart)
				So(cbStart.Index, ShouldEqual, 0)
				So(cbStart.ContentBlock.Type, ShouldEqual, "thinking")

				// content_block_delta for thinking text
				So(events[2].event, ShouldEqual, "content_block_delta")
				var cbDelta anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[2].data), &cbDelta)
				So(cbDelta.Index, ShouldEqual, 0)
				So(cbDelta.Delta.Type, ShouldEqual, "thinking_delta")
				So(cbDelta.Delta.Thinking, ShouldEqual, "The user wants weather info for Shanghai.")

				// content_block_delta for signature
				So(events[3].event, ShouldEqual, "content_block_delta")
				var sigDelta anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[3].data), &sigDelta)
				So(sigDelta.Index, ShouldEqual, 0)
				So(sigDelta.Delta.Type, ShouldEqual, "signature_delta")
				So(sigDelta.Delta.Signature, ShouldEqual, "sig-stream-abc")

				// content_block_stop
				So(events[4].event, ShouldEqual, "content_block_stop")
			})

			Convey("Then text events should be emitted correctly", func() {
				// content_block_start for text
				So(events[5].event, ShouldEqual, "content_block_start")
				var cbStart anthropic.ContentBlockStartEvent
				json.Unmarshal([]byte(events[5].data), &cbStart)
				So(cbStart.Index, ShouldEqual, 1)
				So(cbStart.ContentBlock.Type, ShouldEqual, "text")

				// first text delta
				So(events[6].event, ShouldEqual, "content_block_delta")
				var delta1 anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[6].data), &delta1)
				So(delta1.Index, ShouldEqual, 1)
				So(delta1.Delta.Type, ShouldEqual, "text_delta")
				So(delta1.Delta.Text, ShouldEqual, "Now let me get the weather for Shanghai yesterday")

				// second text delta
				So(events[7].event, ShouldEqual, "content_block_delta")
				var delta2 anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[7].data), &delta2)
				So(delta2.Index, ShouldEqual, 1)
				So(delta2.Delta.Type, ShouldEqual, "text_delta")
				So(delta2.Delta.Text, ShouldEqual, " (2025-11-10):")

				// content_block_stop
				So(events[8].event, ShouldEqual, "content_block_stop")
			})

			Convey("Then tool_use events should be emitted correctly", func() {
				// content_block_start for tool_use
				So(events[9].event, ShouldEqual, "content_block_start")
				var cbStart anthropic.ContentBlockStartEvent
				json.Unmarshal([]byte(events[9].data), &cbStart)
				So(cbStart.Index, ShouldEqual, 2)
				So(cbStart.ContentBlock.Type, ShouldEqual, "tool_use")
				So(cbStart.ContentBlock.ID, ShouldEqual, "toolu_016VE91YZYshFFPSevawmcDH")
				So(cbStart.ContentBlock.Name, ShouldEqual, "get_weather")

				// empty input delta
				So(events[10].event, ShouldEqual, "content_block_delta")
				var delta0 anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[10].data), &delta0)
				So(delta0.Index, ShouldEqual, 2)
				So(delta0.Delta.Type, ShouldEqual, "input_json_delta")
				So(delta0.Delta.PartialJSON, ShouldEqual, "")

				// first input delta
				So(events[11].event, ShouldEqual, "content_block_delta")
				var delta1 anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[11].data), &delta1)
				So(delta1.Index, ShouldEqual, 2)
				So(delta1.Delta.Type, ShouldEqual, "input_json_delta")
				So(delta1.Delta.PartialJSON, ShouldEqual, `{"city": "Shanghai"`)

				// second input delta
				So(events[12].event, ShouldEqual, "content_block_delta")
				var delta2 anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[12].data), &delta2)
				So(delta2.Index, ShouldEqual, 2)
				So(delta2.Delta.Type, ShouldEqual, "input_json_delta")
				So(delta2.Delta.PartialJSON, ShouldEqual, `, "date": `)

				// third input delta
				So(events[13].event, ShouldEqual, "content_block_delta")
				var delta3 anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[13].data), &delta3)
				So(delta3.Index, ShouldEqual, 2)
				So(delta3.Delta.Type, ShouldEqual, "input_json_delta")
				So(delta3.Delta.PartialJSON, ShouldEqual, `"2025-11-10`)

				// fourth input delta
				So(events[14].event, ShouldEqual, "content_block_delta")
				var delta4 anthropic.ContentBlockDeltaEvent
				json.Unmarshal([]byte(events[14].data), &delta4)
				So(delta4.Index, ShouldEqual, 2)
				So(delta4.Delta.Type, ShouldEqual, "input_json_delta")
				So(delta4.Delta.PartialJSON, ShouldEqual, `"}`)

				// content_block_stop
				So(events[15].event, ShouldEqual, "content_block_stop")
			})

			Convey("Then message_delta should have stop reason and usage", func() {
				So(events[16].event, ShouldEqual, "message_delta")
				var msgDelta anthropic.MessageDeltaEvent
				json.Unmarshal([]byte(events[16].data), &msgDelta)
				So(msgDelta.Delta.StopReason, ShouldEqual, anthropic.StopReasonToolUse)
				So(msgDelta.Usage.OutputTokens, ShouldEqual, 93)
			})

			Convey("Then message_stop should be the last event", func() {
				So(events[17].event, ShouldEqual, "message_stop")
			})
		})
	})
}
