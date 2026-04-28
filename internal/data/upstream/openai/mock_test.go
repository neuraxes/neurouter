package openai

import (
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

var mockChatReq = &entity.ChatReq{
	Id:    "mock_chat_id",
	Model: "gpt-4o-mini",
	Config: &v1.GenerationConfig{
		Temperature: new(float32(0)),
		ReasoningConfig: &v1.ReasoningConfig{
			Effort: v1.ReasoningEffort_REASONING_EFFORT_HIGH,
		},
	},
	Messages: []*v1.Message{
		{
			Role: v1.Role_SYSTEM,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: "You are helpful assistant.",
					},
				},
			},
		},
		{
			Role: v1.Role_USER,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: "hi, how are you? and how is the weather yesterday in shanghai?",
					},
				},
			},
		},
		{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: "Hello! I'm doing well, thank you for asking. \n\nTo check the weather in Shanghai for yesterday, I'll need to know what date yesterday was. Let me get today's date first, and then I can look up the weather for the previous day.",
					},
				},
				{
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "call_xJAu30R2cdheI331NUxp6CqL",
							Name: "get_today_date",
							Inputs: []*v1.ToolUse_Input{
								{
									Input: &v1.ToolUse_Input_Text{
										Text: "{}",
									},
								},
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
							Id: "call_xJAu30R2cdheI331NUxp6CqL",
							Outputs: []*v1.ToolResult_Output{
								{
									Output: &v1.ToolResult_Output_Text{
										Text: `{"date":"2025-11-11"}`,
									},
								},
							},
						},
					},
				},
			},
		},
	},
	Tools: []*v1.Tool{
		{
			Tool: &v1.Tool_Function_{
				Function: &v1.Tool_Function{
					Name:        "get_today_date",
					Description: "Get today's date",
					Parameters: &v1.Schema{
						Type:       v1.Schema_TYPE_OBJECT,
						Properties: map[string]*v1.Schema{},
					},
				},
			},
		},
		{
			Tool: &v1.Tool_Function_{
				Function: &v1.Tool_Function{
					Name:        "get_weather",
					Description: "Get weather for specific date",
					Parameters: &v1.Schema{
						Type: v1.Schema_TYPE_OBJECT,
						Properties: map[string]*v1.Schema{
							"city": {
								Type:        v1.Schema_TYPE_STRING,
								Description: "The name of the city",
							},
							"date": {
								Type:        v1.Schema_TYPE_STRING,
								Description: "The date to get the weather for",
							},
						},
						Required: []string{"city", "date"},
					},
				},
			},
		},
	},
}

var mockChatCompletionRequestBody = `{
    "messages": [
        {
            "content": [
                {
                    "text": "You are helpful assistant.",
                    "type": "text"
                }
            ],
            "role": "system"
        },
        {
            "content": [
                {
                    "text": "hi, how are you? and how is the weather yesterday in shanghai?",
                    "type": "text"
                }
            ],
            "role": "user"
        },
        {
            "content": [
                {
                    "text": "Hello! I'm doing well, thank you for asking. \n\nTo check the weather in Shanghai for yesterday, I'll need to know what date yesterday was. Let me get today's date first, and then I can look up the weather for the previous day.",
                    "type": "text"
                }
            ],
            "tool_calls": [
                {
                    "id": "call_xJAu30R2cdheI331NUxp6CqL",
                    "function": {
                        "arguments": "{}",
                        "name": "get_today_date"
                    },
                    "type": "function"
                }
            ],
            "role": "assistant"
        },
        {
            "content": [
                {
                    "text": "{\"date\":\"2025-11-11\"}",
                    "type": "text"
                }
            ],
            "tool_call_id": "call_xJAu30R2cdheI331NUxp6CqL",
            "role": "tool"
        }
    ],
    "model": "gpt-4o-mini",
    "temperature": 0,
    "reasoning_effort": "high",
    "tools": [
        {
            "function": {
                "name": "get_today_date",
                "description": "Get today's date",
                "parameters": {
                    "properties": {},
                    "type": "object"
                }
            },
            "type": "function"
        },
        {
            "function": {
                "name": "get_weather",
                "description": "Get weather for specific date",
                "parameters": {
                    "properties": {
                        "city": {
                            "description": "The name of the city",
                            "type": "string"
                        },
                        "date": {
                            "description": "The date to get the weather for",
                            "type": "string"
                        }
                    },
                    "required": [
                        "city",
                        "date"
                    ],
                    "type": "object"
                }
            },
            "type": "function"
        }
    ]
}`

var mockResponsesRequestBody = `{
  "store": false,
  "temperature": 0,
  "include": ["reasoning.encrypted_content"],
  "input": [
    {
      "content": [
        { "text": "You are helpful assistant.", "type": "input_text" }
      ],
      "role": "system"
    },
    {
      "content": [
        {
          "text": "hi, how are you? and how is the weather yesterday in shanghai?",
          "type": "input_text"
        }
      ],
      "role": "user"
    },
    {
      "content": [
        {
          "text": "Hello! I'm doing well, thank you for asking. \n\nTo check the weather in Shanghai for yesterday, I'll need to know what date yesterday was. Let me get today's date first, and then I can look up the weather for the previous day.",
          "type": "output_text"
        }
      ],
      "role": "assistant",
      "type": "message"
    },
    {
      "arguments": "{}",
      "call_id": "call_xJAu30R2cdheI331NUxp6CqL",
      "name": "get_today_date",
      "type": "function_call"
    },
    {
      "call_id": "call_xJAu30R2cdheI331NUxp6CqL",
      "output": [{ "text": "{\"date\":\"2025-11-11\"}", "type": "input_text" }],
      "type": "function_call_output"
    }
  ],
  "model": "gpt-4o-mini",
  "reasoning": { "effort": "high", "summary": "auto" },
  "tools": [
    {
      "parameters": { "properties": {}, "type": "object" },
      "name": "get_today_date",
      "description": "Get today's date",
      "type": "function"
    },
    {
      "parameters": {
        "properties": {
          "city": { "description": "The name of the city", "type": "string" },
          "date": {
            "description": "The date to get the weather for",
            "type": "string"
          }
        },
        "required": ["city", "date"],
        "type": "object"
      },
      "name": "get_weather",
      "description": "Get weather for specific date",
      "type": "function"
    }
  ]
}`

var mockChatCompletionResponseBody = `{
    "id": "chatcmpl-vcH3X0yomLBqyz4Ox0he",
    "object": "chat.completion",
    "created": 1762970559,
    "model": "gpt-4o-mini",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "Now I'll check the weather for Shanghai for yesterday (November 10th, 2025):",
                "tool_calls": [
                    {
                        "index": 0,
                        "id": "call_ASrYNiguEOTadw8mGyP8cVwp",
                        "type": "function",
                        "function": {
                            "name": "get_weather",
                            "arguments": "{\"city\":\"Shanghai\",\"date\":\"2025-11-10\"}"
                        }
                    }
                ]
            },
            "logprobs": null,
            "finish_reason": "tool_calls"
        }
    ],
    "usage": {
        "prompt_tokens": 303,
        "completion_tokens": 46,
        "total_tokens": 349,
        "prompt_tokens_details": {
            "cached_tokens": 256
        }
    },
    "system_fingerprint": "fp_560af6e559"
}`

var mockChatResp = &entity.ChatResp{
	Id:     "chatcmpl-vcH3X0yomLBqyz4Ox0he",
	Model:  "gpt-4o-mini",
	Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
	Message: &v1.Message{
		Id:   "mock_message_id",
		Role: v1.Role_MODEL,
		Contents: []*v1.Content{
			{
				Content: &v1.Content_Text{
					Text: "Now I'll check the weather for Shanghai for yesterday (November 10th, 2025):",
				},
			},
			{
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   "call_ASrYNiguEOTadw8mGyP8cVwp",
						Name: "get_weather",
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: `{"city":"Shanghai","date":"2025-11-10"}`,
								},
							},
						},
					},
				},
			},
		},
	},
	Statistics: &v1.Statistics{
		Usage: &v1.Statistics_Usage{
			InputTokens:       303,
			OutputTokens:      46,
			CachedInputTokens: 256,
		},
	},
}

var mockResponsesResponseBody = `{
    "id": "resp_0dd30283291d2ea70169d0e9b95d488193a445876892202e22",
    "model": "gpt-4o-mini",
    "object": "response",
    "output": [
        {
            "type": "reasoning",
            "id": "rs_0dd30283291d2ea70169d0e9b9b3f08193a563e455f75eef7f",
            "summary": [
                {
                    "text": "**Looking up yesterday's weather**\n\nThe user greeted me with \"hi, how are you?\" and asked about the weather in Shanghai yesterday. Since today is November 11, 2025, I can easily calculate that yesterday's date was November 10, 2025. I need to use the weather tool to get that information, specifically checking for the weather conditions in Shanghai on that date. Let's proceed with that!",
                    "type": "summary_text"
                }
            ],
            "encrypted_content": "gAAAAABp0Om8ygrVlzOfI4jMt3tSNBZ-XBbU3uK_9f3ev0EGMZH2er5Pt1mpadlQpmXxYiK81yfuaPzZs2pgFeuTQMByY1qZd_lIeHQtX1m8Yt-Bkh49goZaUiXXPysITbkK9_8Z0tK7VzEfbE1wg6aVYN70pNVD5LM2K8TUD7rHX7VB9UpyVLJdNX6Pj0AdxbjPdpVZCieHB8wM4oNEUFmlLtOgw5CfOSHBAS3o3ewC2AFFVal-DXurw4Mu_c8_8zPBKNj_355wjT-1k-DdK6TrjRbfSUnkfKyIKCQHzPs0rVuC3kvLL2prLAsI1TqU1Al4JFFm3aTiViChdkf_bDtf-oo1NtDCRdSMYU0Vyj5qoHPnPnKuFCVzERJ3QE7Rz1-aYBb9DCDbvg-aWF-nqeLroclts6Fhu7ZBDYEJtLqzZvn67VrBcQEX0SUEO1rX5nwknxka0kxwo7wYmtKJAyK1F-TioSCBnfdorVIfMSDkiDTB6oc9Ddu82U6HLNX_oX4Ls1mfjo3sjNnRvXaS4vhtjqqH6xlwgEwWLg21ezSDPBHHKlNrqcUiU2Y9NXbZvv6mcZS8kUohNKSf8arGSZj60W58g77VGzMs292uqyTexCJSgFRk41mGmoVM8Nd757PEiaYw7AdYPQKrNzqbSeLk4ZOi2kpFPwsfGfpwVLQInZQ6iyCXCKLj7wR8qkHi4r4dwEjHPwfolyjJZMCjAlASZ5gxDYPztthffu8ju-4vBr1JtvO_f9dk7bys1vwG-zknbhWKz_w8hjeIqHq-KCLbCTvWmQI0SqWNOkDXbDmIdyfslQkFqW-A8wjABCOo1S271N6WzPcXMmqoT9qXwDhpWrZTkcUx7VSWKPRdiY5mH4_Y3EFUsHOFAmpPiuuuEwwBvULeNJhYuW4AyeWhGSs5b6iJapLfUqghsl1GpTqauOn8motlFUWQY0ynAtaUHmoS_jf0eAyLXymm6F3mdk1YjhnddOuGT5by-MoiQlj8XCksn3D11HFJOmeiNtCMZPggAUxNfHKjowt3C6a7fjQn0OmpSfoa5EL30hQiC3T51bzXohq8pBg="
        },
        {
            "type": "function_call",
            "id": "fc_0dd30283291d2ea70169d0e9bc4f948193a7ed1ab06c9f04f9",
            "status": "completed",
            "arguments": "{\"city\":\"Shanghai\",\"date\":\"2025-11-10\"}",
            "call_id": "call_MaEJJBhEY01WIMrgrTn4j32q",
            "name": "get_weather"
        }
    ],
    "reasoning": {
        "effort": "high",
        "summary": "detailed"
    },
    "status": "completed",
    "temperature": 1,
    "tool_choice": "auto",
    "tools": [
        {
            "type": "function",
            "name": "get_today_date",
            "description": "Get today's date",
            "parameters": {
                "properties": {},
                "type": "object",
                "additionalProperties": false,
                "required": []
            },
            "strict": true
        },
        {
            "type": "function",
            "name": "get_weather",
            "description": "Get weather for specific date",
            "parameters": {
                "properties": {
                    "city": {
                        "description": "The name of the city",
                        "type": "string"
                    },
                    "date": {
                        "description": "The date to get the weather for",
                        "type": "string"
                    }
                },
                "required": [
                    "city",
                    "date"
                ],
                "type": "object",
                "additionalProperties": false
            },
            "strict": true
        }
    ],
    "top_p": 0.98,
    "truncation": "disabled",
    "usage": {
        "total_tokens": 281,
        "input_tokens": 196,
        "output_tokens": 85,
        "input_tokens_details": {
            "cached_tokens": 0
        },
        "output_tokens_details": {
            "reasoning_tokens": 56
        }
    }
}`

var mockResponsesResp = &entity.ChatResp{
	Id:     "mock_chat_id",
	Model:  "gpt-4o-mini",
	Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
	Message: &v1.Message{
		Id:   "resp_0dd30283291d2ea70169d0e9b95d488193a445876892202e22",
		Role: v1.Role_MODEL,
		Contents: []*v1.Content{
			{
				Id:        "rs_0dd30283291d2ea70169d0e9b9b3f08193a563e455f75eef7f",
				Reasoning: true,
				Content:   &v1.Content_Text{},
				Metadata: map[string]string{
					"summary":       "**Looking up yesterday's weather**\n\nThe user greeted me with \"hi, how are you?\" and asked about the weather in Shanghai yesterday. Since today is November 11, 2025, I can easily calculate that yesterday's date was November 10, 2025. I need to use the weather tool to get that information, specifically checking for the weather conditions in Shanghai on that date. Let's proceed with that!",
					"summary_index": "0",
					"encrypted":     "gAAAAABp0Om8ygrVlzOfI4jMt3tSNBZ-XBbU3uK_9f3ev0EGMZH2er5Pt1mpadlQpmXxYiK81yfuaPzZs2pgFeuTQMByY1qZd_lIeHQtX1m8Yt-Bkh49goZaUiXXPysITbkK9_8Z0tK7VzEfbE1wg6aVYN70pNVD5LM2K8TUD7rHX7VB9UpyVLJdNX6Pj0AdxbjPdpVZCieHB8wM4oNEUFmlLtOgw5CfOSHBAS3o3ewC2AFFVal-DXurw4Mu_c8_8zPBKNj_355wjT-1k-DdK6TrjRbfSUnkfKyIKCQHzPs0rVuC3kvLL2prLAsI1TqU1Al4JFFm3aTiViChdkf_bDtf-oo1NtDCRdSMYU0Vyj5qoHPnPnKuFCVzERJ3QE7Rz1-aYBb9DCDbvg-aWF-nqeLroclts6Fhu7ZBDYEJtLqzZvn67VrBcQEX0SUEO1rX5nwknxka0kxwo7wYmtKJAyK1F-TioSCBnfdorVIfMSDkiDTB6oc9Ddu82U6HLNX_oX4Ls1mfjo3sjNnRvXaS4vhtjqqH6xlwgEwWLg21ezSDPBHHKlNrqcUiU2Y9NXbZvv6mcZS8kUohNKSf8arGSZj60W58g77VGzMs292uqyTexCJSgFRk41mGmoVM8Nd757PEiaYw7AdYPQKrNzqbSeLk4ZOi2kpFPwsfGfpwVLQInZQ6iyCXCKLj7wR8qkHi4r4dwEjHPwfolyjJZMCjAlASZ5gxDYPztthffu8ju-4vBr1JtvO_f9dk7bys1vwG-zknbhWKz_w8hjeIqHq-KCLbCTvWmQI0SqWNOkDXbDmIdyfslQkFqW-A8wjABCOo1S271N6WzPcXMmqoT9qXwDhpWrZTkcUx7VSWKPRdiY5mH4_Y3EFUsHOFAmpPiuuuEwwBvULeNJhYuW4AyeWhGSs5b6iJapLfUqghsl1GpTqauOn8motlFUWQY0ynAtaUHmoS_jf0eAyLXymm6F3mdk1YjhnddOuGT5by-MoiQlj8XCksn3D11HFJOmeiNtCMZPggAUxNfHKjowt3C6a7fjQn0OmpSfoa5EL30hQiC3T51bzXohq8pBg=",
				},
			},
			{
				Id: "fc_0dd30283291d2ea70169d0e9bc4f948193a7ed1ab06c9f04f9",
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   "call_MaEJJBhEY01WIMrgrTn4j32q",
						Name: "get_weather",
						Inputs: []*v1.ToolUse_Input{{
							Input: &v1.ToolUse_Input_Text{
								Text: "{\"city\":\"Shanghai\",\"date\":\"2025-11-10\"}",
							},
						}},
					},
				},
			},
		},
	},
	Statistics: &v1.Statistics{
		Usage: &v1.Statistics_Usage{
			InputTokens:       196,
			OutputTokens:      85,
			CachedInputTokens: 0,
			ReasoningTokens:   56,
		},
	},
}

var mockChatCompletionStreamRequestBody = `{
    "messages": [
        {
            "content": [
                {
                    "text": "You are helpful assistant.",
                    "type": "text"
                }
            ],
            "role": "system"
        },
        {
            "content": [
                {
                    "text": "hi, how are you? and how is the weather yesterday in shanghai?",
                    "type": "text"
                }
            ],
            "role": "user"
        },
        {
            "content": [
                {
                    "text": "Hello! I'm doing well, thank you for asking. \n\nTo check the weather in Shanghai for yesterday, I'll need to know what date yesterday was. Let me get today's date first, and then I can look up the weather for the previous day.",
                    "type": "text"
                }
            ],
            "tool_calls": [
                {
                    "id": "call_xJAu30R2cdheI331NUxp6CqL",
                    "function": {
                        "arguments": "{}",
                        "name": "get_today_date"
                    },
                    "type": "function"
                }
            ],
            "role": "assistant"
        },
        {
            "content": [
                {
                    "text": "{\"date\":\"2025-11-11\"}",
                    "type": "text"
                }
            ],
            "tool_call_id": "call_xJAu30R2cdheI331NUxp6CqL",
            "role": "tool"
        }
    ],
    "model": "gpt-4o-mini",
	"reasoning_effort": "high",
	"stream": true,
	"stream_options": {
		"include_usage": true
	},
    "temperature": 0,
    "tools": [
        {
            "function": {
                "name": "get_today_date",
                "description": "Get today's date",
                "parameters": {
                    "properties": {},
                    "type": "object"
                }
            },
            "type": "function"
        },
        {
            "function": {
                "name": "get_weather",
                "description": "Get weather for specific date",
                "parameters": {
                    "properties": {
                        "city": {
                            "type": "string",
                            "description": "The name of the city"
                        },
                        "date": {
                            "type": "string",
                            "description": "The date to get the weather for"
                        }
                    },
                    "required": [
                        "city",
                        "date"
                    ],
                    "type": "object"
                }
            },
            "type": "function"
        }
    ]
}`

var mockChatCompletionStreamResponseBody = `data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":"Today is November 11,"},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":" "},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":"2025. Therefore, yesterday was November 10, 2025."},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":" \n\n"},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":"Now, I will check the weather in Shanghai for November"},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":" "},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":"10, 2025"},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":"."},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":null,"tool_calls":[{"index":0,"id":"call_CzJFKEw26rJ6McvhRnMq1Izg","type":"function","function":{"name":"get_weather","arguments":""}}]},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":null,"tool_calls":[{"index":0,"function":{"arguments":"{\"city\":\"Shanghai\",\""},"type":"function"}]},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":null,"tool_calls":[{"index":0,"function":{"arguments":"date\":\"2025-11-10"},"type":"function"}]},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":null,"tool_calls":[{"index":0,"function":{"arguments":"\"}"},"type":"function"}]},"finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":"tool_calls","logprobs":null}],"system_fingerprint":"fp_560af6e559"}

data: {"id":"chatcmpl-tZZ2ljb9Bz4BoRIcS6cL","model":"gpt-4o-mini","object":"chat.completion.chunk","created":1763047385,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null,"logprobs":null}],"usage":{"prompt_tokens":192,"completion_tokens":65,"total_tokens":257,"prompt_tokens_details":{"cached_tokens":0}}}

data: [DONE]

`

var mockChatStreamResp = []*entity.ChatResp{
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:       "mock_message_id",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Today is November 11,"}}},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:       "mock_message_id",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " "}}},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:       "mock_message_id",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "2025. Therefore, yesterday was November 10, 2025."}}},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:       "mock_message_id",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " \n\n"}}},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:       "mock_message_id",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Now, I will check the weather in Shanghai for November"}}},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:       "mock_message_id",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: " "}}},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:       "mock_message_id",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "10, 2025"}}},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:       "mock_message_id",
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "."}}},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Index: new(uint32(0)),
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "call_CzJFKEw26rJ6McvhRnMq1Izg",
							Name: "get_weather",
							Inputs: []*v1.ToolUse_Input{
								{
									Input: &v1.ToolUse_Input_Text{},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Index: new(uint32(0)),
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Inputs: []*v1.ToolUse_Input{
								{
									Input: &v1.ToolUse_Input_Text{Text: "{\"city\":\"Shanghai\",\""},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Index: new(uint32(0)),
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Inputs: []*v1.ToolUse_Input{
								{
									Input: &v1.ToolUse_Input_Text{Text: "date\":\"2025-11-10"},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Index: new(uint32(0)),
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Inputs: []*v1.ToolUse_Input{
								{
									Input: &v1.ToolUse_Input_Text{Text: "\"}"},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		Id:     "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model:  "gpt-4o-mini",
		Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
		},
	},
	{
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens:  192,
				OutputTokens: 65,
			},
		},
	},
}
