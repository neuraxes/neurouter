package openai

import (
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"k8s.io/utils/ptr"
)

var mockChatReq = &entity.ChatReq{
	Id:    "mock_chat_id",
	Model: "gpt-4o-mini",
	Config: &v1.GenerationConfig{
		Temperature: ptr.To[float32](0),
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
	Id:    "chatcmpl-vcH3X0yomLBqyz4Ox0he",
	Model: "gpt-4o-mini",
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
		Id:    "chatcmpl-tZZ2ljb9Bz4BoRIcS6cL",
		Model: "gpt-4o-mini",
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Metadata: map[string]string{
				"finish_reason": "tool_calls",
			},
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
