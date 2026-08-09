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

//go:embed responses_tool_result_request.json
var responsesToolResultRequest []byte

//go:embed responses_tool_result_response.json
var responsesToolResultResponse []byte

// ResponsesToolResult covers stateless replay of encrypted reasoning, a prior
// function call and its result before the final answer.
var ResponsesToolResult = &Fixture{
	Name:     "responses_tool_result",
	Request:  responsesToolResultRequest,
	Response: responsesToolResultResponse,
	ChatReq: &v1.ChatReq{
		Id:    "responses_tool_result",
		Model: "openai/gpt-5-mini",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(1024)),
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
			{
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
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{
						Content: &v1.Content_ToolResult{
							ToolResult: &v1.ToolResult{
								Id: "call_HuTyMtN6jtA1RB2cVXNI1obq",
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
		Id:     "responses_tool_result",
		Model:  "openai/gpt-5-mini",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "gen-1785502446-Gkxi8IHY2X1TDnhyveUP",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Id:    "rs_08b5de5c9cfca3ba016a6c9aef5cb48192b3489396e78e1c08",
					Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{
						Opaque: "gAAAAABqbJrx7o1iCxEaxKbipe0cFHtnNVthhyklgpvCaBuyJA-Tb-hb4xCeJLWG5D2Si9rRajjURjW5SSb39t8u5eKvOYLBZeUc0VORpb8Qs-UtD3kedJ1wlXYErT7n3rsk64CsmrgWI8ec8U8bUimnY6lzolPfW6aNeIbudWRe3wbFBE3rnFNxO261yuTC8DKZXZ1MR-shFvY6WzxiwHqDAd5Lqdrg5fB4K-ds5U2Ya-83xenuNqq9IkdYeFxrndpheyb-GItbnCFUf4JoNkn6-7GNVUrmZ8SWzZcmZjCVijYpOhaIgUkLV3c7e6GlQJtOeInXGCpT289_AxVa2eWsP90Y0nlx39Fnhq4Ihbx8VmovF2UeHGm3FwYDMHZX3OuxRPaHTnGtqH3E939FKahDwFrp05mICEUsVUbFHol5Z82Naoi3JutQlOQ6GMroV27KfiTsqE19mkAWyDW94LUvKGPWtl188NdFLFd5Gh_owx4zd4Uv9pNNQ6ZaqgLSHcsNNlqIxnFY_1QaXTMEFOBo2oLvhUc5e2yugvgFA1mOdUi-uNifX0T8HmTf5EZTRveXf_J5ksb9n_cxYmKRtONXrYYUlRIMlfWcsfz4pE_DLMTCmPhPmFcaX5SvO1gSVn-LEnH3WTwCNGRIWVyTjMtyCcExFAYmyZhwYYkMx2tEC26iF4PXx4rM_CZhFMRolJLmCUY6auZq01H0Z5yvTNUde8TNAw69l8ii2JNRe_F-Y6uhGZKngh7hufApEfW1W8ZLbbw1OxG88pBjP5Mr4YKALbP021FSrx8lonXnsglXDMbFAbMWDQUKKBvB03RlxGwaCMMue4wEEIwIqEQ4N3NdcrH5CeP1dlZi5WfPloDHyFobB43215_EnibSGjdC2rSegmiFM26kTjXzChXgN476x-oVf_2NWh2_HPj8kP3a96rhWiYVDoDUwhPq-Yxipa_3NrY5taBHSRpKDCw3URtuHNomXXuumuawnoEIQRi9Tp12iQVm5YQ-KN-pvR5tjyc3W-ArGy0UDaCPxKudtFr4BI9S0r75lIgmL-Iy0oVsT1hcAvbWYh6Hh3bH4FUWHs9v6hRcL7BLVdVHk0_KQBx2TcBhdVaS6QDmFafAxGDftu4AFa7UZGULUjzPLGQ3J4qIEgYEpi5423F5KSCjUlO4mGfIWZn-tQUs9IGaYdtW1G4nwGNhCB_AxeVizU59Q6v9m-F6O3P-VNaFSS4UexHpxEAkSpz8vss9t0X_J8ovx_HGQ38iKCC0MAO3Oh9FHX4rzS_Bm1SYje-Ull6xRKKMJmOX7zqWeu8v2_kNq4acEK9vlNWyuoE=",
					},
				},
				{Content: v1.NewTextContent("Shanghai on 2025-11-10: Cloudy, high 18°C, low 11°C, precipitation 2.3 mm.")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 253, OutputTokens: 115, ReasoningTokens: 64},
		},
	},
}
