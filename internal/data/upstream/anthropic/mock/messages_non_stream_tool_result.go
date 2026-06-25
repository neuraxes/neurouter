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

//go:embed messages_non_stream_tool_result_request.json
var nonStreamToolResultRequest []byte

//go:embed messages_non_stream_tool_result_response.json
var nonStreamToolResultResponse []byte

// NonStreamToolResult covers a multi-turn request that feeds a prior tool call
// and its tool result back to the model, and whose response is a plain text
// answer with stop reason end_turn.
var NonStreamToolResult = &Fixture{
	Name:     "non_stream_tool_result",
	Request:  nonStreamToolResultRequest,
	Response: nonStreamToolResultResponse,
	ChatReq: &v1.ChatReq{
		Id:    "non_stream_tool_result",
		Model: "anthropic/claude-sonnet-4.6",
		Config: &v1.GenerationConfig{
			MaxTokens:   new(int64(512)),
			Temperature: new(float32(0)),
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
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{
						Content: &v1.Content_ToolResult{
							ToolResult: &v1.ToolResult{
								Id: "toolu_01EQKr5bPHvy92UvtzNbXFDu",
								Outputs: []*v1.ToolResult_Output{
									{Output: &v1.ToolResult_Output_Text{Text: `{"city":"Shanghai","date":"2025-11-10","condition":"Cloudy","high_c":18,"low_c":11,"precip_mm":2.3,"humidity":0.74}`}},
								},
							},
						},
					},
					{Content: v1.NewTextContent("Using only the tool result, give the city, date, condition, high and low temperature in Celsius, and precipitation in one sentence. Do not call another tool.")},
				},
			},
		},
		Tools:    []*v1.Tool{getWeatherTool()},
		Metadata: map[string]string{"user_id": "anthropic-conversion-fixture-user"},
	},
	ChatResp: &v1.ChatResp{
		Id:     "non_stream_tool_result",
		Model:  "anthropic/claude-4.6-sonnet-20260217",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "gen-1782639842-oi4uhCV6FgqoD7s1IGrX",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent("On **2025-11-10**, **Shanghai** experienced **cloudy** conditions with a high of **18°C** and a low of **11°C**, and **2.3 mm** of precipitation.")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 883, OutputTokens: 50},
		},
	},
}
