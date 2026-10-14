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
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestChatCompletionStreamServer(t *testing.T) {
	builtinToolUseStart := &v1.ChatEvent_ContentStart{ContentStart: &v1.ContentStart{
		Index: 1,
		Content: &v1.ContentStart_ToolUse{ToolUse: &v1.ToolUseStart{
			Id:   "call-2",
			Name: "web_search",
			Type: v1.ToolType_TOOL_TYPE_BUILTIN,
		}},
	}}

	streamChunks := func(events []*v1.ChatEvent) []string {
		httpCtx := newResponsesTestHTTPContext()
		server := &chatCompletionStreamServer{ctx: context.Background(), httpCtx: httpCtx}
		for _, event := range events {
			So(server.Send(event), ShouldBeNil)
		}
		var chunks []string
		for _, event := range parseResponsesSSEEvents(httpCtx.body.String()) {
			raw, err := json.Marshal(event.data)
			So(err, ShouldBeNil)
			chunks = append(chunks, string(raw))
		}
		return chunks
	}

	Convey("Given custom and builtin tool calls followed by a function call", t, func() {
		chunks := streamChunks([]*v1.ChatEvent{
			v1.NewChatEvent("request-1", v1.NewMessageStartEvent("resp-1", "gpt-5")),
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartCustomToolUseEvent("ctc-1", 0, "call-1", "calculate")),
			v1.NewChatEvent("request-1", v1.NewContentDeltaToolInputTextEvent(0, "1+2")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(0)),
			v1.NewChatEvent("request-1", builtinToolUseStart),
			v1.NewChatEvent("request-1", v1.NewContentDeltaToolInputTextEvent(1, `{"q":"news"}`)),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(1)),
			v1.NewChatEvent("request-1", v1.NewContentStartToolUseEvent(2, "call-3", "search")),
			v1.NewChatEvent("request-1", v1.NewContentDeltaToolInputTextEvent(2, `{"q":"test"}`)),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(2)),
			v1.NewChatEvent("request-1", v1.NewMessageStopEvent(v1.ChatStatus_CHAT_STATUS_PENDING_TOOL_USE)),
		})

		Convey("Then only the function call is streamed, as the first tool call", func() {
			var toolCallChunks []string
			for _, chunk := range chunks {
				if gjson.Get(chunk, "choices.0.delta.tool_calls").Exists() {
					toolCallChunks = append(toolCallChunks, chunk)
				}
			}
			So(toolCallChunks, ShouldHaveLength, 2)
			So(gjson.Get(toolCallChunks[0], "choices.0.delta.tool_calls.0.index").Int(), ShouldEqual, 0)
			So(gjson.Get(toolCallChunks[0], "choices.0.delta.tool_calls.0.id").String(), ShouldEqual, "call-3")
			So(gjson.Get(toolCallChunks[1], "choices.0.delta.tool_calls.0.function.arguments").String(), ShouldEqual, `{"q":"test"}`)
			So(gjson.Get(chunks[len(chunks)-1], "choices.0.finish_reason").String(), ShouldEqual, "tool_calls")
		})
	})

	Convey("Given a stream whose only tool call is custom", t, func() {
		chunks := streamChunks([]*v1.ChatEvent{
			v1.NewChatEvent("request-1", v1.NewMessageStartEvent("resp-1", "gpt-5")),
			v1.NewChatEvent("request-1", v1.NewIdentifiedContentStartCustomToolUseEvent("ctc-1", 0, "call-1", "calculate")),
			v1.NewChatEvent("request-1", v1.NewContentDeltaToolInputTextEvent(0, "1+2")),
			v1.NewChatEvent("request-1", v1.NewContentStopEvent(0)),
			v1.NewChatEvent("request-1", v1.NewMessageStopEvent(v1.ChatStatus_CHAT_STATUS_PENDING_TOOL_USE)),
		})

		Convey("Then no tool call is streamed and the stream finishes as a plain stop", func() {
			for _, chunk := range chunks {
				So(gjson.Get(chunk, "choices.0.delta.tool_calls").Exists(), ShouldBeFalse)
			}
			So(gjson.Get(chunks[len(chunks)-1], "choices.0.finish_reason").String(), ShouldEqual, "stop")
		})
	})
}
