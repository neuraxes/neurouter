package anthropic

import (
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/util"
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
		MaxTokens:   new(int64(8192)),
		Temperature: new(float32(0)),
		ReasoningConfig: &v1.ReasoningConfig{
			TokenBudget: 1024,
		},
	},
	Messages: []*v1.Message{
		{
			Role: v1.Role_SYSTEM,
			Contents: []*v1.Content{
				{
					Content: v1.NewTextContent("You are helpful assistant."),
				},
			},
		},
		{
			Role: v1.Role_USER,
			Contents: []*v1.Content{
				{
					Content: v1.NewTextContent("hi, how are you?"),
				},
				{
					Content: v1.NewTextContent("and how is the weather yesterday in shanghai?"),
				},
			},
		},
		{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Phase:     v1.ContentPhase_CONTENT_PHASE_REASONING,
					Signature: "Ep4ECkgICRABGAIqQPDbGoDv1NFNPOMhf8vnh3ThJJnizSYc3/qCq21j8CAwGCwTEcrY/ctoXRntgx/1cvl3mEfFEECC5LMfgJLVcQESDHKtOYoRTWm6f2jcRhoM9h+b/XKb6bxWHrScIjBD/7A9wL/wrdLHbMSM4YVrMXw7ZUQIMpTkYRWAhCMhY25FVp5KkkV2FoHQ29XQ7nAqgwMXVaVyxVPJJMQHhat2xNtBfsOafMu5TBeR+f1LjPMqdoz55nrTWGE5K2yO00BTDIFv4wf8jtbZHC1EskLkqWej1lt/wIL2fS3ZbcgPkaclKlPjGtWrGaCdgcjLYeK1BiiepbwepZWGSPfEduaLqQBkJSFB7ykbqbSk+gCKFfV1nQVRuBWQ5fJnK/59a9YjrBlizasV4d0QRA4Z1+NniaZh7Zh2s6/hOGFJHb3Aqypxiy/GFb34tkCojj6u8tF2tyBL0J/d09z+lZ/Sc4rCkfjya9/rx4QRKy42v2Cn+1fO5f90Fs5Dw8sL4czPVoD6bYNZE1AVHb5Vgu7tN22hYdxFzaR+vhhEtIwGs32IgWS5jRRR5LsZoEzaDFo3HyE5R1sZyE0E79tojMFmndvIvYQuybOEb/nqyJm1ua9jdmL+M1yNHBuO0NWB2Jh0c0IlsTre5enlLQrjTiwCmtMacdrsVJViUW2nkBEOUBudHu6bZkS1Fqe0Ro/7dSjYQyhBqUeJnvYYAQ==",
					Content:   v1.NewTextContent("The user is asking two things:\n1. How am I? - This is a greeting\n2. What was the weather yesterday in Shanghai?\n\nFor the second part, I need to:\n1. First get today's date using get_date\n2. Then calculate yesterday's date\n3. Then get the weather for Shanghai on yesterday's date\n\nLet me start by getting today's date first, since I need that to determine yesterday's date."),
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
					InputSchema: util.MustStructFromMap(map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					}),
				},
			},
		},
		{
			Tool: &v1.Tool_Function_{
				Function: &v1.Tool_Function{
					Name:        "get_weather",
					Description: "Get weather for specific date",
					InputSchema: util.MustStructFromMap(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"city": map[string]any{
								"type":        "string",
								"description": "The name of the city",
							},
							"date": map[string]any{
								"type":        "string",
								"description": "The date to get the weather for",
							},
						},
						"required": []string{"city", "date"},
					}),
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
				Content: v1.NewTextContent("Now let me get the weather for Shanghai yesterday:"),
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
		Usage: &v1.Usage{
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

func mockStreamEvent(event v1.ChatEventPayload) *v1.ChatEvent {
	return v1.NewChatEvent("", event)
}

func mockStreamEventWithUsage(event v1.ChatEventPayload, usage *v1.Usage) *v1.ChatEvent {
	e := mockStreamEvent(event)
	e.Usage = usage
	return e
}

var mockChatStreamEvents = []*v1.ChatEvent{
	mockStreamEventWithUsage(
		v1.NewMessageStartEvent("msg_016m3rsWB3U7eYBEKjTRSruv", "claude-haiku-4-5-20251001"),
		&v1.Usage{InputTokens: 840},
	),
	mockStreamEvent(v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_REASONING)),
	mockStreamEvent(v1.NewContentDeltaTextEvent(0, "The user wants weather info for Shanghai.")),
	mockStreamEvent(v1.NewContentDeltaSignatureEvent(0, "sig-stream-abc")),
	mockStreamEvent(v1.NewContentStopEvent(0)),
	mockStreamEvent(v1.NewContentStartTextEvent(1, v1.ContentPhase_CONTENT_PHASE_NORMAL)),
	mockStreamEvent(v1.NewContentDeltaTextEvent(1, "Now let me get the weather for Shanghai yesterday")),
	mockStreamEvent(v1.NewContentDeltaTextEvent(1, " (2025-11-10):")),
	mockStreamEvent(v1.NewContentStopEvent(1)),
	mockStreamEvent(v1.NewContentStartToolUseEvent(2, "toolu_016VE91YZYshFFPSevawmcDH", "get_weather")),
	mockStreamEvent(v1.NewContentDeltaToolInputTextEvent(2, "")),
	mockStreamEvent(v1.NewContentDeltaToolInputTextEvent(2, "{\"city\": \"Shanghai\"")),
	mockStreamEvent(v1.NewContentDeltaToolInputTextEvent(2, ", \"date\": ")),
	mockStreamEvent(v1.NewContentDeltaToolInputTextEvent(2, "\"2025-11-10")),
	mockStreamEvent(v1.NewContentDeltaToolInputTextEvent(2, "\"}")),
	mockStreamEvent(v1.NewContentStopEvent(2)),
	mockStreamEventWithUsage(
		v1.NewMessageStopEvent(v1.ChatStatus_CHAT_PENDING_TOOL_USE),
		&v1.Usage{OutputTokens: 93},
	),
}
