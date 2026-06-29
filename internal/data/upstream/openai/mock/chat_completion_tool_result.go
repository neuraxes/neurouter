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

//go:embed chat_completion_tool_result_request.json
var toolResultRequest []byte

//go:embed chat_completion_tool_result_response.json
var toolResultResponse []byte

// ToolResult covers a multi-turn request that feeds a prior tool call (an
// assistant message carrying only a tool_use) and its tool result back to the
// model.
var ToolResult = &Fixture{
	Name:     "tool_result",
	Request:  toolResultRequest,
	Response: toolResultResponse,
	ChatReq: &v1.ChatReq{
		Id:    "tool_result",
		Model: "openai/gpt-4o",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(1024)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_NONE},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Use prior tool results exactly and keep the final answer concise.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Use the get_weather tool for Shanghai on 2025-11-10. Do not answer from memory. Return no final prose until the tool has been called.")},
				},
			},
			{
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
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{
						Content: &v1.Content_ToolResult{
							ToolResult: &v1.ToolResult{
								Id: "call_3ErPbH5DIz3E8M4hevb9tMg7",
								Outputs: []*v1.ToolResult_Output{
									{Output: &v1.ToolResult_Output_Text{Text: `{"city":"Shanghai","date":"2025-11-10","condition":"Cloudy","high_c":18,"low_c":11,"precip_mm":2.3,"humidity":0.74,"units":"metric"}`}},
								},
							},
						},
					},
					{Content: v1.NewTextContent("Using only the tool result, give the city, date, condition, high and low temperature in Celsius, and precipitation in one sentence. Do not call another tool.")},
				},
			},
		},
		Tools: []*v1.Tool{getWeatherTool()},
	},
	ChatResp: &v1.ChatResp{
		Id:     "gen-1782736825-52jMHSQnqyXGTxw7F3CD",
		Model:  "openai/gpt-4o",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent("On November 10, 2025, in Shanghai, the weather was cloudy with a high of 18°C, a low of 11°C, and 2.3 mm of precipitation.")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 250, OutputTokens: 41},
		},
	},
}
