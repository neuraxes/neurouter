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
	"github.com/neuraxes/neurouter/internal/util"
)

//go:embed messages_non_stream_tool_call_request.json
var nonStreamToolCallRequest []byte

//go:embed messages_non_stream_tool_call_response.json
var nonStreamToolCallResponse []byte

// getWeatherTool is the shared tool definition used by the tool-call and
// tool-result fixtures.
func getWeatherTool() *v1.Tool {
	return &v1.Tool{
		Tool: &v1.Tool_Function_{
			Function: &v1.Tool_Function{
				Name:        "get_weather",
				Description: "Look up historical weather for a city on a specific date.",
				InputSchema: util.MustStructFromMap(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"city": map[string]any{
							"type":        "string",
							"description": "City name in English.",
						},
						"date": map[string]any{
							"type":        "string",
							"description": "Date in YYYY-MM-DD format.",
						},
						"units": map[string]any{
							"type":        "string",
							"description": "Temperature unit.",
							"enum":        []string{"metric", "imperial"},
						},
					},
					"required":             []string{"city", "date"},
					"additionalProperties": false,
				}),
			},
		},
	}
}

// NonStreamToolCall covers a non-stream request whose generation config sets
// temperature, top_p and top_k, and whose response is a tool call (text block
// followed by a tool_use block, stop reason tool_use).
var NonStreamToolCall = &Fixture{
	Name:     "non_stream_tool_call",
	Request:  nonStreamToolCallRequest,
	Response: nonStreamToolCallResponse,
	ChatReq: &v1.ChatReq{
		Id:    "non_stream_tool_call",
		Model: "anthropic/claude-sonnet-4.6",
		Config: &v1.GenerationConfig{
			MaxTokens:   new(int64(512)),
			Temperature: new(float32(0)),
			TopP:        new(float32(0.8)),
			TopK:        new(int64(40)),
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
		Tools:    []*v1.Tool{getWeatherTool()},
		Metadata: map[string]string{"user_id": "anthropic-conversion-fixture-user"},
	},
	ChatResp: &v1.ChatResp{
		Id:     "non_stream_tool_call",
		Model:  "anthropic/claude-4.6-sonnet-20260217",
		Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
		Message: &v1.Message{
			Id:   "gen-1782639386-Li6vCZE4duB99RWBlwEB",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent("Sure! Let me fetch the weather data for Shanghai on 2025-11-10 right away.")},
				{
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "toolu_01EQKr5bPHvy92UvtzNbXFDu",
							Name: "get_weather",
							Inputs: []*v1.ToolUse_Input{
								{Input: &v1.ToolUse_Input_Text{Text: `{"city":"Shanghai","date":"2025-11-10"}`}},
							},
						},
					},
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 700, OutputTokens: 98},
		},
	},
}
