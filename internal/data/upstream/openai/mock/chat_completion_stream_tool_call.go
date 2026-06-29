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

package mock

import (
	_ "embed"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

//go:embed chat_completion_stream_tool_call_request.json
var streamToolCallRequest []byte

//go:embed chat_completion_stream_tool_call_response.txt
var streamToolCallResponse []byte

func streamToolCallChatEvents() []*v1.ChatEvent {
	id := eventBuilder("stream_tool_call")
	return []*v1.ChatEvent{
		id.of(v1.NewMessageStartEvent("mock_message_id", "openai/gpt-4o")),
		id.of(v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_NORMAL)),
		id.of(v1.NewContentDeltaTextEvent(0, "Today's")),
		id.of(v1.NewContentDeltaTextEvent(0, " date")),
		id.of(v1.NewContentDeltaTextEvent(0, " is")),
		id.of(v1.NewContentDeltaTextEvent(0, " November")),
		id.of(v1.NewContentDeltaTextEvent(0, " ")),
		id.of(v1.NewContentDeltaTextEvent(0, "10")),
		id.of(v1.NewContentDeltaTextEvent(0, ",")),
		id.of(v1.NewContentDeltaTextEvent(0, " ")),
		id.of(v1.NewContentDeltaTextEvent(0, "202")),
		id.of(v1.NewContentDeltaTextEvent(0, "3")),
		id.of(v1.NewContentDeltaTextEvent(0, ".")),
		id.of(v1.NewContentDeltaTextEvent(0, " I")),
		id.of(v1.NewContentDeltaTextEvent(0, " will")),
		id.of(v1.NewContentDeltaTextEvent(0, " now")),
		id.of(v1.NewContentDeltaTextEvent(0, " look")),
		id.of(v1.NewContentDeltaTextEvent(0, " up")),
		id.of(v1.NewContentDeltaTextEvent(0, " the")),
		id.of(v1.NewContentDeltaTextEvent(0, " weather")),
		id.of(v1.NewContentDeltaTextEvent(0, " for")),
		id.of(v1.NewContentDeltaTextEvent(0, " Shanghai")),
		id.of(v1.NewContentDeltaTextEvent(0, " on")),
		id.of(v1.NewContentDeltaTextEvent(0, " November")),
		id.of(v1.NewContentDeltaTextEvent(0, " ")),
		id.of(v1.NewContentDeltaTextEvent(0, "10")),
		id.of(v1.NewContentDeltaTextEvent(0, ",")),
		id.of(v1.NewContentDeltaTextEvent(0, " ")),
		id.of(v1.NewContentDeltaTextEvent(0, "202")),
		id.of(v1.NewContentDeltaTextEvent(0, "5")),
		id.of(v1.NewContentDeltaTextEvent(0, ".")),
		id.of(v1.NewContentStopEvent(0)),
		id.of(v1.NewContentStartToolUseEvent(1, "call_6g00pJ6tnrsXQ0o9yILksX7j", "get_weather")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, `{"`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "city")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, `":"`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "Shanghai")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, `","`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "date")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, `":"`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "202")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "5")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "-")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "11")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "-")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, "10")),
		id.of(v1.NewContentDeltaToolInputTextEvent(1, `"}`)),
		id.of(v1.NewContentStopEvent(1)),
		id.withUsage(
			v1.NewMessageStopEvent(v1.ChatStatus_CHAT_PENDING_TOOL_USE),
			&v1.Usage{InputTokens: 132, OutputTokens: 53},
		),
	}
}

// StreamToolCall covers a streaming request offering a single function tool.
var StreamToolCall = &Fixture{
	Name:     "stream_tool_call",
	Request:  streamToolCallRequest,
	Response: streamToolCallResponse,
	Stream:   true,
	ChatReq: &v1.ChatReq{
		Id:    "stream_tool_call",
		Model: "openai/gpt-4o",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(1024)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_NONE},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Briefly note the date, then call the weather tool exactly once.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("First state today's date in one short sentence, then call get_weather for Shanghai on 2025-11-10. Do not answer the weather from memory.")},
				},
			},
		},
		Tools: []*v1.Tool{getWeatherTool()},
	},
	ChatEvents: streamToolCallChatEvents(),
}
