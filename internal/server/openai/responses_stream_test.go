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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v3/middleware"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

type responsesTestResponseWriter struct {
	http.ResponseWriter
	ctx *responsesTestHTTPContext
}

func (w *responsesTestResponseWriter) Header() http.Header {
	return w.ctx.headers
}

func (w *responsesTestResponseWriter) Write(data []byte) (int, error) {
	return w.ctx.body.Write(data)
}

func (w *responsesTestResponseWriter) WriteHeader(statusCode int) {
	w.ctx.statusCode = statusCode
}

func (w *responsesTestResponseWriter) Flush() {}

type responsesTestHTTPContext struct {
	kratoshttp.Context
	req        *http.Request
	statusCode int
	headers    http.Header
	body       bytes.Buffer
	writer     *responsesTestResponseWriter
}

func newResponsesTestHTTPContext(request ...*http.Request) *responsesTestHTTPContext {
	ctx := &responsesTestHTTPContext{headers: make(http.Header)}
	if len(request) > 0 {
		ctx.req = request[0]
	}
	ctx.writer = &responsesTestResponseWriter{ctx: ctx}
	return ctx
}

func (c *responsesTestHTTPContext) Request() *http.Request {
	return c.req
}

func (c *responsesTestHTTPContext) Response() kratoshttp.ResponseWriter {
	return c.writer
}

func (c *responsesTestHTTPContext) Middleware(handler middleware.Handler) middleware.Handler {
	return handler
}

func (c *responsesTestHTTPContext) Blob(statusCode int, contentType string, data []byte) error {
	c.statusCode = statusCode
	c.headers.Set("Content-Type", contentType)
	_, _ = c.body.Write(data)
	return nil
}

type parsedResponsesSSEEvent struct {
	typeName string
	data     map[string]any
}

func parseResponsesSSEEvents(body string) []parsedResponsesSSEEvent {
	var events []parsedResponsesSSEEvent
	var typeName string
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if value, ok := strings.CutPrefix(line, "event: "); ok {
			typeName = value
			continue
		}
		if value, ok := strings.CutPrefix(line, "data: "); ok {
			var data map[string]any
			if json.Unmarshal([]byte(value), &data) == nil {
				events = append(events, parsedResponsesSSEEvent{typeName: typeName, data: data})
			}
		}
	}
	return events
}

func TestResponsesStreamServer(t *testing.T) {
	Convey("Given native reasoning, text, and tool stream events", t, func() {
		httpCtx := newResponsesTestHTTPContext()
		server := &responsesStreamServer{
			ctx:     context.Background(),
			httpCtx: httpCtx,
		}

		for _, event := range []*v1.ChatEvent{
			v1.NewChatEvent("request-1", v1.NewMessageStartEvent("resp-1", "gpt-5")),
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartTextEvent(
				"rs-1",
				0,
				v1.ContentPhase_CONTENT_PHASE_REASONING,
			)),
			v1.NewChatEvent("request-1", v1.NewContentDeltaTextEvent(0, "think")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(0)),
			v1.NewChatEvent("request-1", v1.NewContentSnapshotEvent(&v1.Content{
				Id:      "rs-1",
				Index:   new(uint32(1)),
				Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
				Content: &v1.Content_Opaque{Opaque: "encrypted"},
			})),
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartTextEvent(
				"msg-1",
				2,
				v1.ContentPhase_CONTENT_PHASE_NORMAL,
			)),
			v1.NewChatEvent("request-1", v1.NewContentDeltaTextEvent(2, "answer")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(2)),
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartToolUseEvent(
				"fc-1",
				3,
				"call-1",
				"lookup",
			)),
			v1.NewChatEvent("request-1", v1.NewContentDeltaToolInputTextEvent(3, `{"q":"test"}`)),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(3)),
			{
				Id:    "request-1",
				Usage: &v1.Usage{InputTokens: 10, OutputTokens: 8, ReasoningTokens: 3},
				Event: v1.NewMessageStopEvent(v1.ChatStatus_CHAT_PENDING_TOOL_USE),
			},
		} {
			So(server.Send(event), ShouldBeNil)
		}

		parsed := parseResponsesSSEEvents(httpCtx.body.String())

		expectedTypes := []string{
			"response.created",
			"response.in_progress",
			"response.output_item.added",
			"response.reasoning_summary_part.added",
			"response.reasoning_summary_text.delta",
			"response.reasoning_summary_text.done",
			"response.reasoning_summary_part.done",
			"response.output_item.done",
			"response.output_item.added",
			"response.content_part.added",
			"response.output_text.delta",
			"response.output_text.done",
			"response.content_part.done",
			"response.output_item.done",
			"response.output_item.added",
			"response.function_call_arguments.delta",
			"response.function_call_arguments.done",
			"response.output_item.done",
			"response.completed",
		}

		Convey("Then event types and sequence numbers follow the Responses protocol", func() {
			So(parsed, ShouldHaveLength, len(expectedTypes))
			for index, expectedType := range expectedTypes {
				So(parsed[index].typeName, ShouldEqual, expectedType)
				So(parsed[index].data["type"], ShouldEqual, expectedType)
				So(parsed[index].data["sequence_number"], ShouldEqual, float64(index))
			}
			So(httpCtx.body.String(), ShouldNotContainSubstring, "[DONE]")
		})

		Convey("Then reasoning text and encrypted state share one output item", func() {
			doneItem := parsed[7].data["item"].(map[string]any)
			So(doneItem["id"], ShouldEqual, "rs-1")
			So(doneItem["encrypted_content"], ShouldEqual, "encrypted")
			summary := doneItem["summary"].([]any)
			So(summary, ShouldHaveLength, 1)
			So(summary[0].(map[string]any)["text"], ShouldEqual, "think")

			addedReasoningItems := 0
			for _, event := range parsed {
				if event.typeName != "response.output_item.added" {
					continue
				}
				item := event.data["item"].(map[string]any)
				if item["type"] == "reasoning" {
					addedReasoningItems++
				}
			}
			So(addedReasoningItems, ShouldEqual, 1)
		})

		Convey("Then the response carries a creation timestamp", func() {
			created := parsed[0].data["response"].(map[string]any)["created_at"]
			So(created, ShouldNotBeNil)
			So(created.(float64), ShouldBeGreaterThan, 0)
		})

		Convey("Then message and function-call item identities are preserved", func() {
			terminal := parsed[len(parsed)-1].data["response"].(map[string]any)
			output := terminal["output"].([]any)
			So(output, ShouldHaveLength, 3)
			So(output[1].(map[string]any)["id"], ShouldEqual, "msg-1")
			So(output[2].(map[string]any)["id"], ShouldEqual, "fc-1")
			So(output[2].(map[string]any)["call_id"], ShouldEqual, "call-1")
		})

		Convey("Then the terminal response contains the complete output and usage", func() {
			terminal := parsed[len(parsed)-1].data["response"].(map[string]any)
			So(terminal["id"], ShouldEqual, "resp-1")
			So(terminal["status"], ShouldEqual, "completed")
			So(terminal["output"], ShouldHaveLength, 3)
			usage := terminal["usage"].(map[string]any)
			So(usage["input_tokens"], ShouldEqual, float64(10))
			So(usage["output_tokens"], ShouldEqual, float64(8))
			So(usage["total_tokens"], ShouldEqual, float64(18))
			So(usage["output_tokens_details"].(map[string]any)["reasoning_tokens"], ShouldEqual, float64(3))
		})
	})

	Convey("Given two text blocks sharing one item id", t, func() {
		httpCtx := newResponsesTestHTTPContext()
		server := &responsesStreamServer{
			ctx:     context.Background(),
			httpCtx: httpCtx,
		}

		for _, event := range []*v1.ChatEvent{
			v1.NewChatEvent("request-1", v1.NewMessageStartEvent("resp-1", "gpt-5")),
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartTextEvent(
				"msg-1",
				0,
				v1.ContentPhase_CONTENT_PHASE_NORMAL,
			)),
			v1.NewChatEvent("request-1", v1.NewContentDeltaTextEvent(0, "first")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(0)),
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartTextEvent(
				"msg-1",
				1,
				v1.ContentPhase_CONTENT_PHASE_NORMAL,
			)),
			v1.NewChatEvent("request-1", v1.NewContentDeltaTextEvent(1, "second")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(1)),
			v1.NewChatEvent("request-1", v1.NewMessageStopEvent(v1.ChatStatus_CHAT_COMPLETED)),
		} {
			So(server.Send(event), ShouldBeNil)
		}

		parsed := parseResponsesSSEEvents(httpCtx.body.String())

		Convey("Then both parts remain in one output message", func() {
			var addedItems, doneItems int
			var contentIndexes []float64
			for _, event := range parsed {
				switch event.typeName {
				case "response.output_item.added":
					addedItems++
				case "response.output_item.done":
					doneItems++
				case "response.content_part.added":
					contentIndexes = append(contentIndexes, event.data["content_index"].(float64))
				}
			}
			So(addedItems, ShouldEqual, 1)
			So(doneItems, ShouldEqual, 1)
			So(contentIndexes, ShouldResemble, []float64{0, 1})

			terminal := parsed[len(parsed)-1].data["response"].(map[string]any)
			output := terminal["output"].([]any)
			So(output, ShouldHaveLength, 1)
			item := output[0].(map[string]any)
			So(item["id"], ShouldEqual, "msg-1")
			So(item["status"], ShouldEqual, "completed")
			content := item["content"].([]any)
			So(content, ShouldHaveLength, 2)
			So(content[0].(map[string]any)["text"], ShouldEqual, "first")
			So(content[1].(map[string]any)["text"], ShouldEqual, "second")
		})
	})

	Convey("Given two reasoning blocks sharing one item id", t, func() {
		httpCtx := newResponsesTestHTTPContext()
		server := &responsesStreamServer{
			ctx:     context.Background(),
			httpCtx: httpCtx,
		}

		for _, event := range []*v1.ChatEvent{
			v1.NewChatEvent("request-1", v1.NewMessageStartEvent("resp-1", "gpt-5")),
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartTextEvent("rs-1", 0, v1.ContentPhase_CONTENT_PHASE_REASONING)),
			v1.NewChatEvent("request-1", v1.NewContentDeltaTextEvent(0, "first")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(0)),
		} {
			So(server.Send(event), ShouldBeNil)
		}

		So(server.pendingReasoning, ShouldNotBeNil)
		So(server.pendingReasoning.item.Status, ShouldEqual, "in_progress")

		for _, event := range []*v1.ChatEvent{
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartTextEvent("rs-1", 1, v1.ContentPhase_CONTENT_PHASE_REASONING)),
			v1.NewChatEvent("request-1", v1.NewContentDeltaTextEvent(1, "second")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(1)),
			v1.NewChatEvent("request-1", v1.NewMessageStopEvent(v1.ChatStatus_CHAT_COMPLETED)),
		} {
			So(server.Send(event), ShouldBeNil)
		}

		parsed := parseResponsesSSEEvents(httpCtx.body.String())

		Convey("Then one item carries both summary parts", func() {
			var added, done []parsedResponsesSSEEvent
			for _, event := range parsed {
				switch event.typeName {
				case "response.output_item.added":
					added = append(added, event)
				case "response.output_item.done":
					done = append(done, event)
				}
			}
			So(added, ShouldHaveLength, 1)
			So(done, ShouldHaveLength, 1)

			item := done[0].data["item"].(map[string]any)
			So(item["id"], ShouldEqual, "rs-1")
			summary := item["summary"].([]any)
			So(summary, ShouldHaveLength, 2)
			So(summary[0].(map[string]any)["text"], ShouldEqual, "first")
			So(summary[1].(map[string]any)["text"], ShouldEqual, "second")
		})

		Convey("Then the second part is announced under its own summary index", func() {
			var summaryIndexes []float64
			for _, event := range parsed {
				if event.typeName == "response.reasoning_summary_part.added" {
					summaryIndexes = append(summaryIndexes, event.data["summary_index"].(float64))
				}
			}
			So(summaryIndexes, ShouldResemble, []float64{0, 1})
		})
	})

	Convey("Given a reasoning snapshot without an item id", t, func() {
		httpCtx := newResponsesTestHTTPContext()
		server := &responsesStreamServer{
			ctx:     context.Background(),
			httpCtx: httpCtx,
		}

		for _, event := range []*v1.ChatEvent{
			v1.NewChatEvent("request-1", v1.NewMessageStartEvent("resp-1", "claude")),
			v1.NewChatEvent("request-1", v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_REASONING)),
			v1.NewChatEvent("request-1", v1.NewContentDeltaTextEvent(0, "think")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(0)),
			v1.NewChatEvent("request-1", v1.NewContentSnapshotEvent(&v1.Content{
				Index:   new(uint32(1)),
				Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
				Content: &v1.Content_Opaque{Opaque: "encrypted"},
			})),
			v1.NewChatEvent("request-1", v1.NewMessageStopEvent(v1.ChatStatus_CHAT_COMPLETED)),
		} {
			So(server.Send(event), ShouldBeNil)
		}

		parsed := parseResponsesSSEEvents(httpCtx.body.String())

		Convey("Then it merges into the reasoning item that is still open", func() {
			terminal := parsed[len(parsed)-1].data["response"].(map[string]any)
			output := terminal["output"].([]any)
			So(output, ShouldHaveLength, 1)
			item := output[0].(map[string]any)
			So(item["encrypted_content"], ShouldEqual, "encrypted")
			So(item["summary"].([]any), ShouldHaveLength, 1)
		})
	})
}
