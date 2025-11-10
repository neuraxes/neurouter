package google

import (
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"k8s.io/utils/ptr"
)

var mockChatReq = &entity.ChatReq{
	Id:    "mock_chat_id",
	Model: "gemini-2.5-flash",
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
					Reasoning: true,
					Content: &v1.Content_Text{
						Text: "**Navigating the User's Queries**\n\nOkay, so I've got two things to address here. First, a simple \"hello\" - easy, I can handle that right away. But then there's the weather in Shanghai *yesterday*. That's where things get interesting. I can't just throw \"yesterday\" at the 'get_weather' tool; it needs a specific date. So, here's my plan:\n\n1.  Acknowledge the initial greeting. Done.\n2.  Use the 'get_date' tool to find out what *today's* date is. Gotta know where we are to figure out yesterday.\n3.  Calculate *yesterday's* date based on the information I get from 'get_date'. Simple math, but it needs to be done.\n4.  Then, and only then, can I use the 'get_weather' tool with \"Shanghai\" and the calculated date for yesterday.\n5.  Finally, I'll package up the response to the greeting and the weather report in a coherent way for the user.\n\nSince I have to actually *do* a calculation (subtracting a day from today), I need to do this in a very systematic way. Get today's date first, then use that as the starting point to get the weather for yesterday. Piece of cake.\n",
					},
				},
				{
					Metadata: map[string]string{
						"thoughtSignature": "CpkHAdHtim8DJwCo3wMevPPWRB6wqQ+we/V4TAgfYvSHEKtu1gdeqx6NJu7qzJCSItuKMXyaYfFqQO6/C5AHyWv1+sP1at3ZpftpdcduLw3lwTO8Yh/PtglzbVUT+LvNdo4RDVethWSfv3RSl+7fAX/2zTY0pzxFGTrcIX9Qwtyv1RmZ+8uJ3+l8dC28mEQIyw9nUrwAe6K6QXCtiHaBGNfXbgh/BlVAaTOR8gKfIkg5XcN9a9foovFodz4Gr1SvIm6DN6K/Gvt2cPkIfGizL1MJyqEjnYkNIzLZmx43uPvaNRoCeMTe9lTpG+a+T9piA/XZn1vAP9XKQEvUUdCOPisdc4u2WAWK5OqotEJ1U12nT/TxULl8PNsx6fM2pdfcrdCgZVyoFZvraAmeGinDel7xuTkwb9PODYTTDgotXqP1V/knbxHG6laUIqioiSA3FHN8c4n+k0dbVf3fILJrJo6hh9vKTkUvlo0w66FxHm0vPYF+fGl8aIFXJeNIUJMvO5pKK8IdKKonZcgQrKCtBXb5/1gHKsYh93yYABs5GJwtz7a3kYwOAZLQe4WP0ehnJ995JQxvVSMgRHz8DBY8zsu6Xl7+LZc1cRrxnX8Fovgeq8o2Badi/mrIZXMYgS7IA77lbah/HCBm/+LR7K6Ty4Svh4TsCPD0PxM7hNRXykAua3eeBpzPVScA6snST4Fts22ZiX3MF7qcjvq4JrgYbpBc6q78kXmtLOx10y6//O34jXeW9KSkZ0CI3Lmi4j+a+ptchftPimk2QeX+Xx54/sBCCXZ7UhcMckJq69fVQf+dX+ueNlbnu0iS4Y3ZhR0c+1+1L4m26k/0FnrWQhqJhk1XzyTsHQurrIA2kv2IsLpZv+aVMEXVe0oTXsLG0GJpEIZ+FHFYG/T1wSEM9CxeSsumLf/90mg8BNGC6C9Sw05Ka2cOduv6g6khhzB6giINw5cm4owss18VtgO1dFD5hCYzbvnxeYDbyfN0fdvzQobu5Ch9gNe5+HjD5Ij+Rt0YtG6bLbKWBYWl+2LA/mM5yNRz3Hk4YWj9JnFNUE6zPFvh7oPan3D6cJy+VU//YP6CnHArBDUimql/rFhcaa7IBE7BUJTZ0RtJWlAV2Bw6QwomLW8CfHo38MUjA5bb29a5LOHt+Nokqlb3JdnKCuCTaaZ+WBZWqL47jGT/Mft39e+ikLHWhP+tSr/M8+06budDTtaAYkcTdI94cAP2",
					},
					Content: &v1.Content_Text{
						Text: "I am doing well, thank you!\n",
					},
				},
				{
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "get_date",
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
							Id: "get_date",
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

const mockGenerateContentRequestBody = `{
    "contents": [
        {
            "parts": [
                {
                    "text": "hi, how are you? and how is the weather yesterday in shanghai?"
                }
            ],
            "role": "user"
        },
        {
            "parts": [
                {
                    "text": "I am doing well, thank you!\n",
                    "thoughtSignature": "CpkHAdHtim8DJwCo3wMevPPWRB6wqQ+we/V4TAgfYvSHEKtu1gdeqx6NJu7qzJCSItuKMXyaYfFqQO6/C5AHyWv1+sP1at3ZpftpdcduLw3lwTO8Yh/PtglzbVUT+LvNdo4RDVethWSfv3RSl+7fAX/2zTY0pzxFGTrcIX9Qwtyv1RmZ+8uJ3+l8dC28mEQIyw9nUrwAe6K6QXCtiHaBGNfXbgh/BlVAaTOR8gKfIkg5XcN9a9foovFodz4Gr1SvIm6DN6K/Gvt2cPkIfGizL1MJyqEjnYkNIzLZmx43uPvaNRoCeMTe9lTpG+a+T9piA/XZn1vAP9XKQEvUUdCOPisdc4u2WAWK5OqotEJ1U12nT/TxULl8PNsx6fM2pdfcrdCgZVyoFZvraAmeGinDel7xuTkwb9PODYTTDgotXqP1V/knbxHG6laUIqioiSA3FHN8c4n+k0dbVf3fILJrJo6hh9vKTkUvlo0w66FxHm0vPYF+fGl8aIFXJeNIUJMvO5pKK8IdKKonZcgQrKCtBXb5/1gHKsYh93yYABs5GJwtz7a3kYwOAZLQe4WP0ehnJ995JQxvVSMgRHz8DBY8zsu6Xl7+LZc1cRrxnX8Fovgeq8o2Badi/mrIZXMYgS7IA77lbah/HCBm/+LR7K6Ty4Svh4TsCPD0PxM7hNRXykAua3eeBpzPVScA6snST4Fts22ZiX3MF7qcjvq4JrgYbpBc6q78kXmtLOx10y6//O34jXeW9KSkZ0CI3Lmi4j+a+ptchftPimk2QeX+Xx54/sBCCXZ7UhcMckJq69fVQf+dX+ueNlbnu0iS4Y3ZhR0c+1+1L4m26k/0FnrWQhqJhk1XzyTsHQurrIA2kv2IsLpZv+aVMEXVe0oTXsLG0GJpEIZ+FHFYG/T1wSEM9CxeSsumLf/90mg8BNGC6C9Sw05Ka2cOduv6g6khhzB6giINw5cm4owss18VtgO1dFD5hCYzbvnxeYDbyfN0fdvzQobu5Ch9gNe5+HjD5Ij+Rt0YtG6bLbKWBYWl+2LA/mM5yNRz3Hk4YWj9JnFNUE6zPFvh7oPan3D6cJy+VU//YP6CnHArBDUimql/rFhcaa7IBE7BUJTZ0RtJWlAV2Bw6QwomLW8CfHo38MUjA5bb29a5LOHt+Nokqlb3JdnKCuCTaaZ+WBZWqL47jGT/Mft39e+ikLHWhP+tSr/M8+06budDTtaAYkcTdI94cAP2"
                },
                {
                    "functionCall": {
                        "name": "get_date"
                    }
                }
            ],
            "role": "model"
        },
        {
            "parts": [
                {
                    "functionResponse": {
                        "name": "get_date",
                        "response": {
                            "date": "2025-11-11"
                        }
                    }
                }
            ],
            "role": "user"
        }
    ],
    "generationConfig": {
        "temperature": 0,
        "thinkingConfig": {
            "includeThoughts": true
        }
    },
	"systemInstruction": {
		"parts": [
			{
				"text": "You are helpful assistant."
			}
		],
		"role": "user"
	},
    "tools": [
        {
            "functionDeclarations": [
                {
                    "name": "get_date",
                    "description": "Get today's date",
                    "parameters": {
                        "type": "OBJECT"
                    }
                },
                {
                    "name": "get_weather",
                    "description": "Get weather for specific date",
                    "parameters": {
                        "properties": {
                            "city": {
                                "description": "The name of the city",
                                "type": "STRING"
                            },
                            "date": {
                                "description": "The date to get the weather for",
                                "type": "STRING"
                            }
                        },
                        "required": [
                            "city",
                            "date"
                        ],
                        "type": "OBJECT"
                    }
                }
            ]
        }
    ]
}`

const mockGenerateContentResponseBody = `{
    "candidates": [
        {
            "content": {
                "parts": [
                    {
                        "text": "**Yesterday's Shanghai Weather**\n\nOkay, so the user asked about the weather in Shanghai yesterday. I've already handled the initial greeting, so I'm past that. Right now, I have today's date, which is \"2025-11-11\". That's a good start.\n\nMy next step is simple: I need to figure out what date \"yesterday\" actually was. With today being the 11th, that means yesterday was the 10th. Easy enough, that's \"2025-11-10\". \n\nNow that I have the date I need, I can use the 'get_weather' tool with the location \"shanghai\" and the date \"2025-11-10\". That should give me the information the user is looking for.\n",
                        "thought": true
                    },
                    {
                        "text": "Yesterday was 2025-11-10.\n",
                        "thoughtSignature": "CtcCAdHtim9//a+avd6Mdp2nfaFDy9UN0XRKL5s7OpASpj4EYl4F3YcytuWj2af37z/RF2Wu4wabZG8dj9X5w5alnFCjBrepYwCZwmjbDWmeDcWfygAo6gtbThCoD7I8k1cZD4SzAMR5tlXSQJdqLJE/D2kT54WsTHLm95UHQ5s68/mYf0n6yKLvj+le92wXvkxFctZ6Gsu/W5ihhj9rlADwWWy4fhhJEvPgHm7z+t9+FmfltfX+/Yumzk2GUXSxxhosRXFX4WXlpW096MEFALnkURWeJ+owj6ppNyqNx6i7Hbz70gH3Y5odjvpGVyk8iaDM6SAWV81q95bcGjqso1LF/AyqnQS26XRFtoRcpdPMfDCrFEOPUcD3MlswapgGmnFK2DdKgpwvc3TPCLna9JdltoEVohVaimT7nS4EkSkrDsHiKsp0Omb7cmOCMCe9WXOHhUsx7qeLaA=="
                    },
                    {
                        "functionCall": {
                            "name": "get_weather",
                            "args": {
                                "date": "2025-11-10",
                                "city": "shanghai"
                            }
                        }
                    }
                ],
                "role": "model"
            },
            "finishReason": "STOP",
            "index": 0,
            "finishMessage": "Model generated function call(s)."
        }
    ],
    "usageMetadata": {
        "promptTokenCount": 144,
        "candidatesTokenCount": 46,
        "totalTokenCount": 292,
        "promptTokensDetails": [
            {
                "modality": "TEXT",
                "tokenCount": 144
            }
        ],
        "thoughtsTokenCount": 102
    },
    "modelVersion": "gemini-2.5-flash",
    "responseId": "gC8Saaa0BOmL2roP5LPX4Q4"
}`

var mockChatResp = &entity.ChatResp{
	Id:    "mock_chat_id",
	Model: "gemini-2.5-flash",
	Message: &v1.Message{
		Id:   "gC8Saaa0BOmL2roP5LPX4Q4",
		Role: v1.Role_MODEL,
		Contents: []*v1.Content{
			{
				Reasoning: true,
				Content: &v1.Content_Text{
					Text: "**Yesterday's Shanghai Weather**\n\nOkay, so the user asked about the weather in Shanghai yesterday. I've already handled the initial greeting, so I'm past that. Right now, I have today's date, which is \"2025-11-11\". That's a good start.\n\nMy next step is simple: I need to figure out what date \"yesterday\" actually was. With today being the 11th, that means yesterday was the 10th. Easy enough, that's \"2025-11-10\". \n\nNow that I have the date I need, I can use the 'get_weather' tool with the location \"shanghai\" and the date \"2025-11-10\". That should give me the information the user is looking for.\n",
				},
			},
			{
				Metadata: map[string]string{
					"thoughtSignature": "CtcCAdHtim9//a+avd6Mdp2nfaFDy9UN0XRKL5s7OpASpj4EYl4F3YcytuWj2af37z/RF2Wu4wabZG8dj9X5w5alnFCjBrepYwCZwmjbDWmeDcWfygAo6gtbThCoD7I8k1cZD4SzAMR5tlXSQJdqLJE/D2kT54WsTHLm95UHQ5s68/mYf0n6yKLvj+le92wXvkxFctZ6Gsu/W5ihhj9rlADwWWy4fhhJEvPgHm7z+t9+FmfltfX+/Yumzk2GUXSxxhosRXFX4WXlpW096MEFALnkURWeJ+owj6ppNyqNx6i7Hbz70gH3Y5odjvpGVyk8iaDM6SAWV81q95bcGjqso1LF/AyqnQS26XRFtoRcpdPMfDCrFEOPUcD3MlswapgGmnFK2DdKgpwvc3TPCLna9JdltoEVohVaimT7nS4EkSkrDsHiKsp0Omb7cmOCMCe9WXOHhUsx7qeLaA==",
				},
				Content: &v1.Content_Text{
					Text: "Yesterday was 2025-11-10.\n",
				},
			},
			{
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   "get_weather",
						Name: "get_weather",
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: `{"city":"shanghai","date":"2025-11-10"}`,
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
			InputTokens:  144,
			OutputTokens: 148,
		},
	},
}

const mockStreamGenerateContentResponseBody = `data: {"candidates": [{"content": {"parts": [{"text": "**Calculating Yesterday's Date**\n\nI've successfully retrieved today's date, 2025-11-11. My current focus is to determine yesterday's date, which I've now calculated as 2025-11-10. This is a crucial step towards providing the requested weather information, and I'm on track to deliver a complete and accurate response.\n\n\n","thought": true}],"role": "model"},"index": 0}],"usageMetadata": {"promptTokenCount": 391,"totalTokenCount": 463,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 391}],"thoughtsTokenCount": 72},"modelVersion": "gemini-2.5-flash","responseId": "HTESab3lIJWe0-kP6_Wh4Ao"}

data: {"candidates": [{"content": {"parts": [{"text": "**Refining Date Parameters**\n\nI've determined yesterday's date, 2025-11-10, crucial for getting the weather. I'm building on that success. My task now is to integrate this date into the 'get_weather' tool with \"shanghai\" as the location. This will allow me to finally provide the requested information, which I will then report back.\n\n\n","thought": true}],"role": "model"},"index": 0}],"usageMetadata": {"promptTokenCount": 391,"totalTokenCount": 493,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 391}],"thoughtsTokenCount": 102},"modelVersion": "gemini-2.5-flash","responseId": "HTESab3lIJWe0-kP6_Wh4Ao"}

data: {"candidates": [{"content": {"parts": [{"text": "Yesterday was 2025-11-10.\n","thoughtSignature": "CikB0e2KbwvEqAUe/Jbf3zx5lg6fKQe382RFpFzHXfaI7x59tkpEUrWQ0Qp6AdHtim93jg0+fEbEV+4yvK/XAUKtsxzs/NjSKNdVB9bE6QiZZgun3CCrEMtOnvI0c0YPeSh7cD7pbCYrEdJAedfO0qLEkLB0Txf7vP8CCXFdVRfid8HX5vATXDgmJBvsb+oNBJMtbYKmvT89CvcceFYWtWUxQPVE7p0KsAEB0e2Kb7CsrMats3mXW9aOqkSbuS3kM3a6YTOQymYEfIsWmRNpM/1wvZsg8rfBSQ4rNtnKoXsRLhYdmpR4T3h5xV/UpclCduXabEjl4BV4lhln1Rp0CdAzW4j60NUv85NKR9Z0rt1sPZwwJ9B+XAgLnqz3aHWGImJG5ZXMa9FRmTIdV3ko8bAgup1nLYrl7UeOb/+QFSEMxqgZ0a9IPAJN/gB1BSBCRAzvZ/xvP3F3tgppAdHtim9mSHTlqsHDzXB6eIXcG+ciJayNMWdfVwrJ1RLK5aXwNgaeQb7p8QANO3gH9GpY5bIIazL+w20wnKM8xFP5rFD8T/x4LNhe+0sXh4Y7aJewaR8C6DN2ocob8zKi1qxqXyEaJL9I"}],"role": "model"},"index": 0}],"usageMetadata": {"promptTokenCount": 391,"candidatesTokenCount": 12,"totalTokenCount": 505,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 391}],"thoughtsTokenCount": 102},"modelVersion": "gemini-2.5-flash","responseId": "HTESab3lIJWe0-kP6_Wh4Ao"}

data: {"candidates": [{"content": {"parts": [{"functionCall": {"name": "get_weather","args": {"city": "shanghai","date": "2025-11-10"}}}],"role": "model"},"finishReason": "STOP","index": 0,"finishMessage": "Model generated function call(s)."}],"usageMetadata": {"promptTokenCount": 391,"candidatesTokenCount": 43,"totalTokenCount": 536,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 391}],"thoughtsTokenCount": 102},"modelVersion": "gemini-2.5-flash","responseId": "HTESab3lIJWe0-kP6_Wh4Ao"}
`

var mockStreamChatResp = []*entity.ChatResp{
	{
		Id:    "mock_chat_id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "HTESab3lIJWe0-kP6_Wh4Ao",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Reasoning: true,
					Content: &v1.Content_Text{
						Text: "**Calculating Yesterday's Date**\n\nI've successfully retrieved today's date, 2025-11-11. My current focus is to determine yesterday's date, which I've now calculated as 2025-11-10. This is a crucial step towards providing the requested weather information, and I'm on track to deliver a complete and accurate response.\n\n\n",
					},
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens:  391,
				OutputTokens: 72,
			},
		},
	},
	{
		Id:    "mock_chat_id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "HTESab3lIJWe0-kP6_Wh4Ao",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Reasoning: true,
					Content: &v1.Content_Text{
						Text: "**Refining Date Parameters**\n\nI've determined yesterday's date, 2025-11-10, crucial for getting the weather. I'm building on that success. My task now is to integrate this date into the 'get_weather' tool with \"shanghai\" as the location. This will allow me to finally provide the requested information, which I will then report back.\n\n\n",
					},
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens:  391,
				OutputTokens: 102,
			},
		},
	},
	{
		Id:    "mock_chat_id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "HTESab3lIJWe0-kP6_Wh4Ao",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Metadata: map[string]string{
						"thoughtSignature": "CikB0e2KbwvEqAUe/Jbf3zx5lg6fKQe382RFpFzHXfaI7x59tkpEUrWQ0Qp6AdHtim93jg0+fEbEV+4yvK/XAUKtsxzs/NjSKNdVB9bE6QiZZgun3CCrEMtOnvI0c0YPeSh7cD7pbCYrEdJAedfO0qLEkLB0Txf7vP8CCXFdVRfid8HX5vATXDgmJBvsb+oNBJMtbYKmvT89CvcceFYWtWUxQPVE7p0KsAEB0e2Kb7CsrMats3mXW9aOqkSbuS3kM3a6YTOQymYEfIsWmRNpM/1wvZsg8rfBSQ4rNtnKoXsRLhYdmpR4T3h5xV/UpclCduXabEjl4BV4lhln1Rp0CdAzW4j60NUv85NKR9Z0rt1sPZwwJ9B+XAgLnqz3aHWGImJG5ZXMa9FRmTIdV3ko8bAgup1nLYrl7UeOb/+QFSEMxqgZ0a9IPAJN/gB1BSBCRAzvZ/xvP3F3tgppAdHtim9mSHTlqsHDzXB6eIXcG+ciJayNMWdfVwrJ1RLK5aXwNgaeQb7p8QANO3gH9GpY5bIIazL+w20wnKM8xFP5rFD8T/x4LNhe+0sXh4Y7aJewaR8C6DN2ocob8zKi1qxqXyEaJL9I",
					},
					Content: &v1.Content_Text{
						Text: "Yesterday was 2025-11-10.\n",
					},
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens:  391,
				OutputTokens: 114,
			},
		},
	},
	{
		Id:    "mock_chat_id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "HTESab3lIJWe0-kP6_Wh4Ao",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "get_weather",
							Name: "get_weather",
							Inputs: []*v1.ToolUse_Input{
								{
									Input: &v1.ToolUse_Input_Text{
										Text: `{"city":"shanghai","date":"2025-11-10"}`,
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
				InputTokens:  391,
				OutputTokens: 145,
			},
		},
	},
}
