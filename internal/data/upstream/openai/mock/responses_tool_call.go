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

//go:embed responses_tool_call_request.json
var responsesToolCallRequest []byte

//go:embed responses_tool_call_response.json
var responsesToolCallResponse []byte

// ResponsesToolCall covers encrypted reasoning state followed by a function
// call.
var ResponsesToolCall = &Fixture{
	Name:     "responses_tool_call",
	Request:  responsesToolCallRequest,
	Response: responsesToolCallResponse,
	ChatReq: &v1.ChatReq{
		Id:    "responses_tool_call",
		Model: "openai/gpt-5-mini",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(2048)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_LOW},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Call the requested tool exactly once and return no final prose before receiving its result.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Call get_weather for Shanghai on 2025-11-10 using metric units. Do not answer the weather from memory.")},
				},
			},
		},
		Tools: []*v1.Tool{getWeatherTool()},
	},
	ChatResp: &v1.ChatResp{
		Id:     "responses_tool_call",
		Model:  "openai/gpt-5-mini",
		Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
		Message: &v1.Message{
			Id:   "gen-1785502387-oWD8DH4G5icXqqD2oDOf",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Id:    "rs_08b5de5c9cfca3ba016a6c9ab440b08192a79b2b9279b4dcfb",
					Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{
						Opaque: "gAAAAABqbJq1D9JzSZUSaZEqzoM0k6uayFu5t_lMusofFnNAt4tnMBS4kHfGgelEcGGlsl7q9JuVjBjDRGjmvK4VBRNCPIzDXuB7ZIu_smp8G6TsBe4QOFz-0WExHDY6Ok07s_nCYqOPJVTdYgOiKkffr2pxyjtZOs39nZjTJQJa9esagoH-I-Uf5hjncgX6rIs1oL_73ldzs0cw-hVZP1u-oPddDsK9dBJKefW6xwv6zjDqMLAXpwV5iRCvIcjrqlQc2Vdd1wwhx5Kf1ChiC-MhPJgOxiugO7igKgCwU61kQ65zGC6z3XU0ftWJDBSzyozzxji1uZVfnfjp5XNrWclEiAcVb6qk-7CPTzvJTpqxOOjW673rMdBp343BwYRT1FkXk1U3KsBBvojM3O7B6eyIlFu1xnFHn1qEPbOwaxkIhbM9TPfAQlIywQnxkGu1075R5Px6V_zH37NDP5YPn8Pn5v0HYFey1j2AXbe_qrxVXSzxjL2zD3NVzr1H_mA_ZVAUCA3F5eh-y0lFkjrojdva5j1xCV56jrDHeDlxmQ3N0eKrn4pW0c2OHptOH_FxBoGKK2FSdftre5EKnGrMFcWZ05DItJXDOcgFXxbxJdhj-9YgTtoNZlRvd4ukWp9amX2sVOsPj2lh7hP0k8H4n28Qs-iZwQ6Z3CEklXh4LJSSR6C21oCs6Jiu2CXzNRn2lk8EU35Jq8j2SHPRV3mxUvKFH7F8MwXl_IeW17t2QO_2y1x_WlN71EWPqREI_FnoFE3KfOyBgqYtGZHrNXhP1jD7XyDeD9y8h0rPMWXD8pBNBncS1rWl9YNxeso0FgHogV-tFK5lN6UfIm8u73s6IbzJIfmNEShdKzm6D0DPaK904yUozBGb5eZfPV3yvtrXeEWiXmeGNGo7MgIbQuwfksnJ3UQCQAlG--s0IQL8Ht9AUQXWfSG6_AcewlOAnQIkftSv7E8RxCRv-Dxda49YIDaPakpmMrKrzeoVn1cDAfTEF7hiDeOKjjRlpffFdczO33CFgbsOEYm3MXefm81ZY9Ov9KVilO0ow1VmU5_fZMOdkxNWwXzNmDVqy0c_bmlExVTtaQswAiKBz7_hXv6zAd3AmkBXg8qRNBmwv2B53ADiZRKRZ_P6nxWXecJPBL_rGxEX7HNN96oQwnwkem594laYaVxdC8vGrcE2T5mylvN8IthYRnPKBbT5W1E9i2dAT1BgDzdMsbUozYJrV2l480C1954qCgS3Oq00HJ4lG3rAiSRr0aEQdUtICeiuDnpgjWgdiMuSk93bSr9yDV5YqNgzO3bHNj18rZ9gzVoigO9qJBMx-GAR1a01O8mGjs4YjzyJRf0xfNd71hUvi32dU1ZEA7qUW2Bn_6he9qKq3jQDwNOmOlS8NJ-kQVIDnlch0HxmW1Gmb5hn6X_8uhQvF7gzegeu4ZsXU9ikfpvGuQCfzE2CQUPyfzx9dxKVMa_bsT54u1YeQb-E",
					},
				},
				{
					Id: "fc_tmp_e0qtpyr71w",
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "call_HuTyMtN6jtA1RB2cVXNI1obq",
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
			Usage: &v1.Usage{InputTokens: 130, OutputTokens: 118, ReasoningTokens: 64},
		},
	},
}
