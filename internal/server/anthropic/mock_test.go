package anthropic

import (
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"k8s.io/utils/ptr"
)

var mockMessagesRequestBody = `{
    "max_tokens": 8192,
    "messages": [
        {
            "content": [
                {
                    "text": "hi, how are you?",
                    "type": "text"
                },
				{
                    "text": "and how is the weather yesterday in shanghai?",
                    "type": "text"
                }
            ],
            "role": "user"
        },
        {
            "content": [
                {
                    "signature": "Ep4ECkgICRABGAIqQPDbGoDv1NFNPOMhf8vnh3ThJJnizSYc3/qCq21j8CAwGCwTEcrY/ctoXRntgx/1cvl3mEfFEECC5LMfgJLVcQESDHKtOYoRTWm6f2jcRhoM9h+b/XKb6bxWHrScIjBD/7A9wL/wrdLHbMSM4YVrMXw7ZUQIMpTkYRWAhCMhY25FVp5KkkV2FoHQ29XQ7nAqgwMXVaVyxVPJJMQHhat2xNtBfsOafMu5TBeR+f1LjPMqdoz55nrTWGE5K2yO00BTDIFv4wf8jtbZHC1EskLkqWej1lt/wIL2fS3ZbcgPkaclKlPjGtWrGaCdgcjLYeK1BiiepbwepZWGSPfEduaLqQBkJSFB7ykbqbSk+gCKFfV1nQVRuBWQ5fJnK/59a9YjrBlizasV4d0QRA4Z1+NniaZh7Zh2s6/hOGFJHb3Aqypxiy/GFb34tkCojj6u8tF2tyBL0J/d09z+lZ/Sc4rCkfjya9/rx4QRKy42v2Cn+1fO5f90Fs5Dw8sL4czPVoD6bYNZE1AVHb5Vgu7tN22hYdxFzaR+vhhEtIwGs32IgWS5jRRR5LsZoEzaDFo3HyE5R1sZyE0E79tojMFmndvIvYQuybOEb/nqyJm1ua9jdmL+M1yNHBuO0NWB2Jh0c0IlsTre5enlLQrjTiwCmtMacdrsVJViUW2nkBEOUBudHu6bZkS1Fqe0Ro/7dSjYQyhBqUeJnvYYAQ==",
                    "thinking": "The user is asking two things:\n1. How am I? - This is a greeting\n2. What was the weather yesterday in Shanghai?\n\nFor the second part, I need to:\n1. First get today's date using get_date\n2. Then calculate yesterday's date\n3. Then get the weather for Shanghai on yesterday's date\n\nLet me start by getting today's date first, since I need that to determine yesterday's date.",
                    "type": "thinking"
                },
                {
                    "id": "toolu_011QvMN77rhma12jw2ETcneN",
                    "input": {},
                    "name": "get_date",
                    "type": "tool_use"
                }
            ],
            "role": "assistant"
        },
        {
            "content": [
                {
                    "tool_use_id": "toolu_011QvMN77rhma12jw2ETcneN",
                    "content": [
                        {
                            "text": "{\"date\":\"2025-11-11\"}",
                            "type": "text"
                        }
                    ],
                    "type": "tool_result"
                }
            ],
            "role": "user"
        }
    ],
    "model": "claude-haiku-4-5-20251001-thinking",
    "temperature": 0,
    "system": [
        {
            "text": "You are helpful assistant.",
            "type": "text"
        }
    ],
    "thinking": {
        "budget_tokens": 1024,
        "type": "enabled"
    },
    "tools": [
        {
            "input_schema": {
                "properties": {},
                "type": "object"
            },
            "name": "get_date",
            "description": "Get today's date"
        },
        {
            "input_schema": {
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
            },
            "name": "get_weather",
            "description": "Get weather for specific date"
        }
    ]
}`

var mockChatReq = &v1.ChatReq{
	Model: "claude-haiku-4-5-20251001-thinking",
	Config: &v1.GenerationConfig{
		MaxTokens:   ptr.To[int64](8192),
		Temperature: ptr.To[float32](0),
		ReasoningConfig: &v1.ReasoningConfig{
			TokenBudget: 1024,
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
						Text: "hi, how are you?",
					},
				},
				{
					Content: &v1.Content_Text{
						Text: "and how is the weather yesterday in shanghai?",
					},
				},
			},
		},
		{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Reasoning: true,
					Metadata: map[string]string{
						"signature": "Ep4ECkgICRABGAIqQPDbGoDv1NFNPOMhf8vnh3ThJJnizSYc3/qCq21j8CAwGCwTEcrY/ctoXRntgx/1cvl3mEfFEECC5LMfgJLVcQESDHKtOYoRTWm6f2jcRhoM9h+b/XKb6bxWHrScIjBD/7A9wL/wrdLHbMSM4YVrMXw7ZUQIMpTkYRWAhCMhY25FVp5KkkV2FoHQ29XQ7nAqgwMXVaVyxVPJJMQHhat2xNtBfsOafMu5TBeR+f1LjPMqdoz55nrTWGE5K2yO00BTDIFv4wf8jtbZHC1EskLkqWej1lt/wIL2fS3ZbcgPkaclKlPjGtWrGaCdgcjLYeK1BiiepbwepZWGSPfEduaLqQBkJSFB7ykbqbSk+gCKFfV1nQVRuBWQ5fJnK/59a9YjrBlizasV4d0QRA4Z1+NniaZh7Zh2s6/hOGFJHb3Aqypxiy/GFb34tkCojj6u8tF2tyBL0J/d09z+lZ/Sc4rCkfjya9/rx4QRKy42v2Cn+1fO5f90Fs5Dw8sL4czPVoD6bYNZE1AVHb5Vgu7tN22hYdxFzaR+vhhEtIwGs32IgWS5jRRR5LsZoEzaDFo3HyE5R1sZyE0E79tojMFmndvIvYQuybOEb/nqyJm1ua9jdmL+M1yNHBuO0NWB2Jh0c0IlsTre5enlLQrjTiwCmtMacdrsVJViUW2nkBEOUBudHu6bZkS1Fqe0Ro/7dSjYQyhBqUeJnvYYAQ==",
					},
					Content: &v1.Content_Text{
						Text: "The user is asking two things:\n1. How am I? - This is a greeting\n2. What was the weather yesterday in Shanghai?\n\nFor the second part, I need to:\n1. First get today's date using get_date\n2. Then calculate yesterday's date\n3. Then get the weather for Shanghai on yesterday's date\n\nLet me start by getting today's date first, since I need that to determine yesterday's date.",
					},
				},
				{
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "toolu_011QvMN77rhma12jw2ETcneN",
							Name: "get_date",
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
							Id: "toolu_011QvMN77rhma12jw2ETcneN",
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
					Name:        "get_date",
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

var mockChatResp = &v1.ChatResp{
	Id:     "mock_chat_id",
	Model:  "claude-haiku-4-5-20251001-thinking",
	Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
	Message: &v1.Message{
		Id:   "msg_01WoXFLH9R6UTE86iV53LFwk",
		Role: v1.Role_MODEL,
		Contents: []*v1.Content{
			{
				Content: &v1.Content_Text{
					Text: "Now let me get the weather for Shanghai yesterday:",
				},
			},
			{
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   "toolu_01RVxboZAP9EKN3ShcyL8tN6",
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
			InputTokens:  840,
			OutputTokens: 86,
		},
	},
}

var mockMessagesResponseBody = `{
    "model": "claude-haiku-4-5-20251001-thinking",
    "id": "msg_01WoXFLH9R6UTE86iV53LFwk",
    "type": "message",
    "role": "assistant",
    "content": [
        {
            "type": "text",
            "text": "Now let me get the weather for Shanghai yesterday:"
        },
        {
            "type": "tool_use",
            "id": "toolu_01RVxboZAP9EKN3ShcyL8tN6",
            "name": "get_weather",
            "input": {"city":"Shanghai","date":"2025-11-10"}
        }
    ],
    "stop_reason": "tool_use",
    "stop_sequence": null,
    "usage": {
        "input_tokens": 840,
        "cache_creation_input_tokens": 0,
        "cache_read_input_tokens": 0,
        "output_tokens": 86
    }
}`

var mockChatStreamResp = []*v1.ChatResp{
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index:     new(uint32(0)),
				Reasoning: true,
				Content:   &v1.Content_Text{Text: "The user wants weather info for Shanghai."},
			}},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens: 840,
			},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index:     new(uint32(0)),
				Reasoning: true,
				Metadata:  map[string]string{"signature": "sig-stream-abc"},
				Content:   &v1.Content_Text{Text: ""},
			}},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index:   new(uint32(1)),
				Content: &v1.Content_Text{Text: "Now let me get the weather for Shanghai yesterday"},
			}},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index:   new(uint32(1)),
				Content: &v1.Content_Text{Text: " (2025-11-10):"},
			}},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index: new(uint32(2)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   "toolu_016VE91YZYshFFPSevawmcDH",
						Name: "get_weather",
					},
				},
			}},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index: new(uint32(2)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: "",
								},
							},
						},
					},
				},
			}},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index: new(uint32(2)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: "{\"city\": \"Shanghai\"",
								},
							},
						},
					},
				},
			}},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index: new(uint32(2)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: ", \"date\": ",
								},
							},
						},
					},
				},
			}},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index: new(uint32(2)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: "\"2025-11-10",
								},
							},
						},
					},
				},
			}},
		},
	},
	{
		Model: "claude-haiku-4-5-20251001",
		Message: &v1.Message{
			Id:   "msg_016m3rsWB3U7eYBEKjTRSruv",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index: new(uint32(2)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: "\"}",
								},
							},
						},
					},
				},
			}},
		},
	},
	{
		Model:  "claude-haiku-4-5-20251001",
		Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				OutputTokens: 93,
			},
		},
	},
}
