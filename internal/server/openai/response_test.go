// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/middleware"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/utils/ptr"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
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

var mockResponseAPIRequestBody = `{
	"model": "gpt-4o",
	"instructions": "You are a helpful assistant.",
	"input": "What is 2+2?",
	"temperature": 0.7,
	"max_output_tokens": 1024,
	"tools": [
		{
			"type": "function",
			"name": "calculate",
			"description": "Perform a calculation",
			"parameters": {
				"type": "object",
				"properties": {
					"expression": { "type": "string" }
				}
			}
		}
	]
}`

var mockResponseAPIChatReq = &v1.ChatReq{
	Model: "gpt-4o",
	Config: &v1.GenerationConfig{
		MaxTokens:   ptr.To[int64](1024),
		Temperature: ptr.To[float32](0.7),
	},
	Messages: []*v1.Message{
		{
			Role: v1.Role_SYSTEM,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "You are a helpful assistant."},
			}},
		},
		{
			Role: v1.Role_USER,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "What is 2+2?"},
			}},
		},
	},
	Tools: []*v1.Tool{{
		Tool: &v1.Tool_Function_{
			Function: &v1.Tool_Function{
				Name:        "calculate",
				Description: "Perform a calculation",
				Parameters: &v1.Schema{
					Type: v1.Schema_TYPE_OBJECT,
					Properties: map[string]*v1.Schema{
						"expression": {Type: v1.Schema_TYPE_STRING},
					},
				},
			},
		},
	}},
}

var mockResponseAPIChatResp = &v1.ChatResp{
	Id:     "test-response-id",
	Model:  "gpt-4o",
	Status: v1.ChatStatus_CHAT_COMPLETED,
	Message: &v1.Message{
		Id:   "msg_test",
		Role: v1.Role_MODEL,
		Contents: []*v1.Content{
			{Content: &v1.Content_Text{Text: "2+2 equals 4."}},
		},
	},
	Statistics: &v1.Statistics{
		Usage: &v1.Statistics_Usage{
			InputTokens:  25,
			OutputTokens: 10,
		},
	},
}

var mockResponseAPIStreamResp = []*v1.ChatResp{
	{
		Model: "gpt-4o",
		Message: &v1.Message{
			Id:   "msg_stream",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index:   new(uint32(0)),
				Content: &v1.Content_Text{Text: "2+2"},
			}},
		},
	},
	{
		Model: "gpt-4o",
		Message: &v1.Message{
			Id:   "msg_stream",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index:   new(uint32(0)),
				Content: &v1.Content_Text{Text: " equals 4."},
			}},
		},
	},
	{
		Model:  "gpt-4o",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens:  25,
				OutputTokens: 10,
			},
		},
	},
}

func TestHandleCreateResponse(t *testing.T) {
	Convey("Given a mock ChatServer for non-streaming Responses API requests", t, func() {
		mock := &mockChatServer{}
		srv := &Server{chatSvc: mock}

		Convey("When a non-streaming response request is sent", func() {
			mock.chatFunc = func(ctx context.Context, req *v1.ChatReq) (*v1.ChatResp, error) {
				So(req.Model, ShouldEqual, "gpt-4o")
				So(req.Messages, ShouldHaveLength, 2)
				So(req.Messages[0].Role, ShouldEqual, v1.Role_SYSTEM)
				So(req.Messages[0].Contents[0].GetText(), ShouldEqual, "You are a helpful assistant.")
				So(req.Messages[1].Role, ShouldEqual, v1.Role_USER)
				So(req.Messages[1].Contents[0].GetText(), ShouldEqual, "What is 2+2?")
				So(req.Tools, ShouldHaveLength, 1)
				So(req.Tools[0].GetFunction().Name, ShouldEqual, "calculate")
				return mockResponseAPIChatResp, nil
			}

			httpReq, _ := http.NewRequest("POST", "/v1/responses", strings.NewReader(mockResponseAPIRequestBody))
			ctx := newMockHTTPContext(httpReq)

			err := srv.handleCreateResponse(ctx)
			So(err, ShouldBeNil)

			Convey("Then the response should be a valid Responses API object", func() {
				var resp responseObject
				err := json.Unmarshal(ctx.respBody.Bytes(), &resp)
				So(err, ShouldBeNil)
				So(resp.Object, ShouldEqual, "response")
				So(resp.Model, ShouldEqual, "gpt-4o")
				So(resp.Status, ShouldEqual, "completed")
				So(resp.Output, ShouldHaveLength, 1)

				outputJSON, _ := json.Marshal(resp.Output[0])
				var msg responseOutputMessage
				json.Unmarshal(outputJSON, &msg)
				So(msg.Type, ShouldEqual, "message")
				So(msg.Role, ShouldEqual, "assistant")
				So(msg.Status, ShouldEqual, "completed")

				So(resp.Usage, ShouldNotBeNil)
				So(resp.Usage.InputTokens, ShouldEqual, 25)
				So(resp.Usage.OutputTokens, ShouldEqual, 10)
				So(resp.Usage.TotalTokens, ShouldEqual, 35)
			})
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

func eventNames(events []sseEvent) []string {
	names := make([]string, 0, len(events))
	for _, ev := range events {
		names = append(names, ev.event)
	}
	return names
}

func containsEvent(events []sseEvent, name string) bool {
	for _, ev := range events {
		if ev.event == name {
			return true
		}
	}
	return false
}

func countEvent(eventNames []string, name string) int {
	c := 0
	for _, n := range eventNames {
		if n == name {
			c++
		}
	}
	return c
}

func collectByEvent(events []sseEvent, name string) []string {
	var out []string
	for _, ev := range events {
		if ev.event == name {
			out = append(out, ev.data)
		}
	}
	return out
}

func strconvI(i int) string { return strconv.Itoa(i) }

var errBoom = errors.New("boom")

func lastEventByName(events []sseEvent, name string) *sseEvent {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].event == name {
			return &events[i]
		}
	}
	return nil
}

func TestHandleCreateResponseStream(t *testing.T) {
	Convey("Given a mock ChatServer for streaming Responses API requests", t, func() {
		mock := &mockChatServer{}
		srv := &Server{chatSvc: mock}

		Convey("When a streaming response request is sent", func() {
			mock.chatStreamFunc = func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
				for _, resp := range mockResponseAPIStreamResp {
					if err := stream.Send(resp); err != nil {
						return err
					}
				}
				return nil
			}

			body := `{"model":"gpt-4o","input":"What is 2+2?","stream":true}`
			httpReq, _ := http.NewRequest("POST", "/v1/responses", strings.NewReader(body))
			ctx := newMockHTTPContext(httpReq)

			err := srv.handleCreateResponse(ctx)
			So(err, ShouldBeNil)

			events := parseSSEEvents(ctx.respBody.String())

			Convey("Then the correct SSE events should be emitted", func() {
				So(len(events), ShouldBeGreaterThanOrEqualTo, 5)

				So(events[0].event, ShouldEqual, "response.created")
				So(events[1].event, ShouldEqual, "response.in_progress")
			})

			Convey("Then response.output_item.added should be emitted for the message", func() {
				var found bool
				for _, ev := range events {
					if ev.event == "response.output_item.added" {
						found = true
						var data responseStreamEvent
						json.Unmarshal([]byte(ev.data), &data)
						So(data.Type, ShouldEqual, "response.output_item.added")
						break
					}
				}
				So(found, ShouldBeTrue)
			})

			Convey("Then response.output_text.delta should be emitted for text chunks", func() {
				var deltas []string
				for _, ev := range events {
					if ev.event == "response.output_text.delta" {
						var data responseStreamEvent
						json.Unmarshal([]byte(ev.data), &data)
						deltas = append(deltas, data.Delta)
					}
				}
				So(len(deltas), ShouldEqual, 2)
				So(deltas[0], ShouldEqual, "2+2")
				So(deltas[1], ShouldEqual, " equals 4.")
			})

			Convey("Then response.completed should be the final event", func() {
				lastEvent := events[len(events)-1]
				So(lastEvent.event, ShouldEqual, "response.completed")
				var data responseStreamEvent
				json.Unmarshal([]byte(lastEvent.data), &data)
				So(data.Type, ShouldEqual, "response.completed")
			})
		})
	})
}

func TestHandleCreateResponseStreamWithToolUse(t *testing.T) {
	Convey("Given a mock ChatServer for streaming tool use", t, func() {
		mock := &mockChatServer{}
		srv := &Server{chatSvc: mock}

		Convey("When a streaming response with tool calls is sent", func() {
			mock.chatStreamFunc = func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
				responses := []*v1.ChatResp{
					{
						Model: "gpt-4o",
						Message: &v1.Message{
							Id:   "msg_tool",
							Role: v1.Role_MODEL,
							Contents: []*v1.Content{{
								Index: new(uint32(0)),
								Content: &v1.Content_ToolUse{
									ToolUse: &v1.ToolUse{
										Id:   "call_xyz",
										Name: "get_weather",
									},
								},
							}},
						},
					},
					{
						Model: "gpt-4o",
						Message: &v1.Message{
							Id:   "msg_tool",
							Role: v1.Role_MODEL,
							Contents: []*v1.Content{{
								Index: new(uint32(0)),
								Content: &v1.Content_ToolUse{
									ToolUse: &v1.ToolUse{
										Inputs: []*v1.ToolUse_Input{{
											Input: &v1.ToolUse_Input_Text{Text: `{"city":"Shanghai"}`},
										}},
									},
								},
							}},
						},
					},
					{
						Model:  "gpt-4o",
						Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
						Statistics: &v1.Statistics{
							Usage: &v1.Statistics_Usage{
								InputTokens:  50,
								OutputTokens: 20,
							},
						},
					},
				}
				for _, resp := range responses {
					if err := stream.Send(resp); err != nil {
						return err
					}
				}
				return nil
			}

			body := `{"model":"gpt-4o","input":"Weather in Shanghai?","stream":true}`
			httpReq, _ := http.NewRequest("POST", "/v1/responses", strings.NewReader(body))
			ctx := newMockHTTPContext(httpReq)

			err := srv.handleCreateResponse(ctx)
			So(err, ShouldBeNil)

			events := parseSSEEvents(ctx.respBody.String())

			Convey("Then function_call events should be emitted", func() {
				var foundAdded, foundArgsDelta bool
				for _, ev := range events {
					if ev.event == "response.output_item.added" {
						var data map[string]any
						json.Unmarshal([]byte(ev.data), &data)
						if item, ok := data["item"].(map[string]any); ok {
							if item["type"] == "function_call" {
								foundAdded = true
								So(item["name"], ShouldEqual, "get_weather")
								So(item["call_id"], ShouldEqual, "call_xyz")
							}
						}
					}
					if ev.event == "response.function_call_arguments.delta" {
						foundArgsDelta = true
						var data responseStreamEvent
						json.Unmarshal([]byte(ev.data), &data)
						So(data.Delta, ShouldEqual, `{"city":"Shanghai"}`)
					}
				}
				So(foundAdded, ShouldBeTrue)
				So(foundArgsDelta, ShouldBeTrue)
			})

			Convey("Then response.completed should have usage", func() {
				lastEvent := events[len(events)-1]
				So(lastEvent.event, ShouldEqual, "response.completed")
			})
		})
	})
}

func TestHandleCreateResponseStreamRefusal(t *testing.T) {
	Convey("Given a mock ChatServer for streaming refusals", t, func() {
		mock := &mockChatServer{}
		srv := &Server{chatSvc: mock}

		Convey("When a refusal is streamed", func() {
			mock.chatStreamFunc = func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
				refusalContent := func(text string) *v1.Content {
					c := &v1.Content{
						Id:      "msg_refused",
						Index:   new(uint32(0)),
						Content: &v1.Content_Text{Text: text},
					}
					c.SetMeta("refusal", "true")
					return c
				}
				responses := []*v1.ChatResp{
					// Two delta chunks then the terminal stats chunk; only the
					// final response carries the CHAT_REFUSED status, exactly
					// like the upstream incoming converter does.
					{
						Model: "gpt-4o",
						Message: &v1.Message{
							Role:     v1.Role_MODEL,
							Contents: []*v1.Content{refusalContent("I cannot ")},
						},
					},
					{
						Model: "gpt-4o",
						Message: &v1.Message{
							Role:     v1.Role_MODEL,
							Contents: []*v1.Content{refusalContent("help with that.")},
						},
					},
					{
						Model:      "gpt-4o",
						Status:     v1.ChatStatus_CHAT_REFUSED,
						Statistics: &v1.Statistics{Usage: &v1.Statistics_Usage{InputTokens: 5, OutputTokens: 3}},
					},
				}
				for _, resp := range responses {
					if err := stream.Send(resp); err != nil {
						return err
					}
				}
				return nil
			}

			body := `{"model":"gpt-4o","input":"forbidden","stream":true}`
			httpReq, _ := http.NewRequest("POST", "/v1/responses", strings.NewReader(body))
			ctx := newMockHTTPContext(httpReq)

			err := srv.handleCreateResponse(ctx)
			So(err, ShouldBeNil)

			events := parseSSEEvents(ctx.respBody.String())
			Convey("Then the SSE stream emits refusal-shaped events", func() {
				So(containsEvent(events, "response.refusal.delta"), ShouldBeTrue)
				So(containsEvent(events, "response.refusal.done"), ShouldBeTrue)
				So(containsEvent(events, "response.content_part.added"), ShouldBeTrue)
				So(containsEvent(events, "response.content_part.done"), ShouldBeTrue)
			})
			Convey("And the terminal event reports the cumulative refusal text", func() {
				last := events[len(events)-1]
				So(last.event, ShouldEqual, "response.completed")
				var data responseStreamEvent
				So(json.Unmarshal([]byte(last.data), &data), ShouldBeNil)
				So(data.Response.Status, ShouldEqual, "completed")
				output := data.Response.Output
				So(len(output), ShouldEqual, 1)
			})
			Convey("And every event carries a sequence_number", func() {
				for _, ev := range events {
					var raw map[string]any
					So(json.Unmarshal([]byte(ev.data), &raw), ShouldBeNil)
					_, ok := raw["sequence_number"]
					So(ok, ShouldBeTrue)
				}
			})
		})
	})
}

func TestHandleCreateResponseStreamReasoningAndIncomplete(t *testing.T) {
	Convey("Given a mock ChatServer for reasoning streams", t, func() {
		mock := &mockChatServer{}
		srv := &Server{chatSvc: mock}

		Convey("When reasoning text is streamed and the response is incomplete", func() {
			mock.chatStreamFunc = func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
				responses := []*v1.ChatResp{
					{
						Model: "o3",
						Message: &v1.Message{
							Role: v1.Role_MODEL,
							Contents: []*v1.Content{{
								Id:        "rs_1",
								Index:     new(uint32(0)),
								Reasoning: true,
								Metadata:  map[string]string{"summary": "Planning."},
								Content:   &v1.Content_Text{Text: "Long internal thought."},
							}},
						},
					},
					{
						Model:  "o3",
						Status: v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT,
					},
				}
				for _, resp := range responses {
					if err := stream.Send(resp); err != nil {
						return err
					}
				}
				return nil
			}

			body := `{"model":"o3","input":"think","stream":true}`
			httpReq, _ := http.NewRequest("POST", "/v1/responses", strings.NewReader(body))
			ctx := newMockHTTPContext(httpReq)

			err := srv.handleCreateResponse(ctx)
			So(err, ShouldBeNil)

			events := parseSSEEvents(ctx.respBody.String())
			So(containsEvent(events, "response.reasoning_summary_text.delta"), ShouldBeTrue)
			So(containsEvent(events, "response.reasoning_text.delta"), ShouldBeTrue)
			last := events[len(events)-1]
			So(last.event, ShouldEqual, "response.incomplete")
		})
	})
}

func TestHandleCreateResponseStreamFailedWithoutUsageChunk(t *testing.T) {
	Convey("Given a mock ChatServer that ends without a stats chunk", t, func() {
		mock := &mockChatServer{}
		srv := &Server{chatSvc: mock}

		Convey("When the final chunk only carries failed status", func() {
			mock.chatStreamFunc = func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
				responses := []*v1.ChatResp{
					{
						Model: "gpt-4o",
						Message: &v1.Message{
							Role: v1.Role_MODEL,
							Contents: []*v1.Content{{
								Index:   new(uint32(0)),
								Content: &v1.Content_Text{Text: "partial"},
							}},
						},
					},
					{
						Model:  "gpt-4o",
						Status: v1.ChatStatus_CHAT_FAILED,
					},
				}
				for _, resp := range responses {
					if err := stream.Send(resp); err != nil {
						return err
					}
				}
				return nil
			}

			body := `{"model":"gpt-4o","input":"hi","stream":true}`
			httpReq, _ := http.NewRequest("POST", "/v1/responses", strings.NewReader(body))
			ctx := newMockHTTPContext(httpReq)

			err := srv.handleCreateResponse(ctx)
			So(err, ShouldBeNil)

			events := parseSSEEvents(ctx.respBody.String())
			So(containsEvent(events, "response.output_item.done"), ShouldBeTrue)
			last := events[len(events)-1]
			So(last.event, ShouldEqual, "response.failed")
		})
	})
}

func TestHandleCreateResponseStreamReasoningSummaryParts(t *testing.T) {
	Convey("Given a mock ChatServer that emits two summary parts followed by reasoning text", t, func() {
		mock := &mockChatServer{}
		srv := &Server{chatSvc: mock}

		mock.chatStreamFunc = func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
			summaryDelta := func(idx int, text string) *v1.Content {
				c := &v1.Content{
					Id:        "rs_upstream_id",
					Index:     new(uint32(0)),
					Reasoning: true,
					Content:   &v1.Content_Text{},
				}
				c.SetMeta("summary", text)
				c.SetMeta("summary_index", strconvI(idx))
				return c
			}
			reasoningTextDelta := func(text string) *v1.Content {
				return &v1.Content{
					Id:        "rs_upstream_id",
					Index:     new(uint32(0)),
					Reasoning: true,
					Content:   &v1.Content_Text{Text: text},
				}
			}
			responses := []*v1.ChatResp{
				{Id: "abc", Model: "o3", Message: &v1.Message{Role: v1.Role_MODEL, Contents: []*v1.Content{summaryDelta(0, "Looking ")}}},
				{Model: "o3", Message: &v1.Message{Role: v1.Role_MODEL, Contents: []*v1.Content{summaryDelta(0, "things up.")}}},
				{Model: "o3", Message: &v1.Message{Role: v1.Role_MODEL, Contents: []*v1.Content{summaryDelta(1, "Then deciding.")}}},
				{Model: "o3", Message: &v1.Message{Role: v1.Role_MODEL, Contents: []*v1.Content{reasoningTextDelta("Internal analysis.")}}},
				{Model: "o3", Status: v1.ChatStatus_CHAT_COMPLETED, Statistics: &v1.Statistics{Usage: &v1.Statistics_Usage{InputTokens: 5, OutputTokens: 10, ReasoningTokens: 7}}},
			}
			for _, resp := range responses {
				if err := stream.Send(resp); err != nil {
					return err
				}
			}
			return nil
		}

		body := `{"model":"o3","input":"think","stream":true}`
		httpReq, _ := http.NewRequest("POST", "/v1/responses", strings.NewReader(body))
		ctx := newMockHTTPContext(httpReq)

		err := srv.handleCreateResponse(ctx)
		So(err, ShouldBeNil)

		events := parseSSEEvents(ctx.respBody.String())

		Convey("Then summary part lifecycle events are emitted in order", func() {
			seq := []string{}
			for _, e := range events {
				seq = append(seq, e.event)
			}
			// Must contain part.added/done bracketing each part and a final
			// summary_text.done flushing each accumulated summary.
			added := countEvent(seq, "response.reasoning_summary_part.added")
			done := countEvent(seq, "response.reasoning_summary_part.done")
			textDone := countEvent(seq, "response.reasoning_summary_text.done")
			So(added, ShouldEqual, 2)
			So(done, ShouldEqual, 2)
			So(textDone, ShouldEqual, 2)
			So(containsEvent(events, "response.reasoning_text.done"), ShouldBeTrue)
		})

		Convey("Then summary deltas for the same part are merged on the wire", func() {
			deltas := collectByEvent(events, "response.reasoning_summary_text.delta")
			So(len(deltas), ShouldEqual, 3) // "Looking ", "things up.", "Then deciding."

			var first, second responseStreamEvent
			So(json.Unmarshal([]byte(deltas[0]), &first), ShouldBeNil)
			So(json.Unmarshal([]byte(deltas[1]), &second), ShouldBeNil)
			// First two deltas belong to summary_index 0.
			So(*first.SummaryIndex, ShouldEqual, 0)
			So(*second.SummaryIndex, ShouldEqual, 0)
			var third responseStreamEvent
			So(json.Unmarshal([]byte(deltas[2]), &third), ShouldBeNil)
			So(*third.SummaryIndex, ShouldEqual, 1)
		})

		Convey("Then the streamed item_id reuses the upstream content id", func() {
			added := collectByEvent(events, "response.output_item.added")
			So(len(added), ShouldEqual, 1)
			var ev responseStreamEvent
			So(json.Unmarshal([]byte(added[0]), &ev), ShouldBeNil)
			b, _ := json.Marshal(ev.Item)
			So(string(b), ShouldContainSubstring, `"id":"rs_upstream_id"`)
		})

		Convey("Then sequence numbers are strictly increasing from zero", func() {
			var prev int64 = -1
			for _, e := range events {
				var raw responseStreamEvent
				So(json.Unmarshal([]byte(e.data), &raw), ShouldBeNil)
				So(raw.SequenceNumber, ShouldEqual, prev+1)
				prev = raw.SequenceNumber
			}
		})

		Convey("Then the terminal completed event echoes the reasoning item in output", func() {
			last := events[len(events)-1]
			So(last.event, ShouldEqual, "response.completed")
			var data responseStreamEvent
			So(json.Unmarshal([]byte(last.data), &data), ShouldBeNil)
			So(data.Response.Output, ShouldHaveLength, 1)
			b, _ := json.Marshal(data.Response.Output[0])
			So(string(b), ShouldContainSubstring, `"type":"reasoning"`)
			So(string(b), ShouldContainSubstring, `"id":"rs_upstream_id"`)
		})

		Convey("Then the response id is derived from the upstream resp.Id", func() {
			created := collectByEvent(events, "response.created")
			So(len(created), ShouldEqual, 1)
			var data responseStreamEvent
			So(json.Unmarshal([]byte(created[0]), &data), ShouldBeNil)
			So(data.Response.ID, ShouldEqual, "resp_abc")
		})
	})
}

func TestHandleCreateResponseStreamUpstreamError(t *testing.T) {
	Convey("Given a ChatStream that returns an error after partial output", t, func() {
		mock := &mockChatServer{}
		srv := &Server{chatSvc: mock}

		mock.chatStreamFunc = func(req *v1.ChatReq, stream v1.Chat_ChatStreamServer) error {
			_ = stream.Send(&v1.ChatResp{
				Id:    "abc",
				Model: "gpt-4o",
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{{
						Index:   new(uint32(0)),
						Content: &v1.Content_Text{Text: "partial"},
					}},
				},
			})
			return errBoom
		}

		body := `{"model":"gpt-4o","input":"hi","stream":true}`
		httpReq, _ := http.NewRequest("POST", "/v1/responses", strings.NewReader(body))
		ctx := newMockHTTPContext(httpReq)

		_ = srv.handleCreateResponse(ctx)
		events := parseSSEEvents(ctx.respBody.String())

		Convey("Then the client still sees a terminal failed event", func() {
			So(len(events), ShouldBeGreaterThan, 0)
			last := events[len(events)-1]
			So(last.event, ShouldEqual, "response.failed")
		})
	})
}
