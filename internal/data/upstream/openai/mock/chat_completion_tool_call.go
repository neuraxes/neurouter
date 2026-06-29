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

//go:embed chat_completion_tool_call_request.json
var toolCallRequest []byte

//go:embed chat_completion_tool_call_response.json
var toolCallResponse []byte

// ToolCall covers a request offering a single function tool.
var ToolCall = &Fixture{
	Name:     "tool_call",
	Request:  toolCallRequest,
	Response: toolCallResponse,
	ChatReq: &v1.ChatReq{
		Id:    "tool_call",
		Model: "openai/gpt-4o",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(4096)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_NONE},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. When the user asks for fresh or external facts and a matching tool is available, call exactly one tool before producing final prose.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Use the get_weather tool for Shanghai on 2025-11-10. Do not answer from memory. Return no final prose until the tool has been called.")},
				},
			},
		},
		Tools: []*v1.Tool{getWeatherTool()},
	},
	ChatResp: &v1.ChatResp{
		Id:     "gen-1782736288-lXdveO1kqMfsE8T6qkFS",
		Model:  "openai/gpt-4o",
		Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "call_3ErPbH5DIz3E8M4hevb9tMg7",
							Name: "get_weather",
							Inputs: []*v1.ToolUse_Input{
								{Input: &v1.ToolUse_Input_Text{Text: `{"city":"Shanghai","date":"2025-11-10","units":"metric"}`}},
							},
						},
					},
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 144, OutputTokens: 27},
		},
	},
}
