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

//go:embed chat_completion_structured_output_request.json
var structuredOutputRequest []byte

//go:embed chat_completion_structured_output_response.json
var structuredOutputResponse []byte

// StructuredOutput covers a request that constrains the reply with a JSON schema
// (grammar), which converts to a json_schema response format.
var StructuredOutput = &Fixture{
	Name:     "structured_output",
	Request:  structuredOutputRequest,
	Response: structuredOutputResponse,
	ChatReq: &v1.ChatReq{
		Id:    "structured_output",
		Model: "openai/gpt-4o",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(1024)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_NONE},
			Grammar: &v1.GenerationConfig_Schema{
				Schema: util.MustStructFromMap(map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"severity": map[string]any{
							"type": "string",
							"enum": []string{"low", "medium", "high"},
						},
						"affected_region":    map[string]any{"type": "string"},
						"primary_cause":      map[string]any{"type": "string"},
						"recommended_action": map[string]any{"type": "string"},
						"signals": map[string]any{
							"type":     "array",
							"items":    map[string]any{"type": "string"},
							"minItems": 1,
						},
					},
					"required": []string{"severity", "affected_region", "primary_cause", "recommended_action", "signals"},
				}),
			},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Return content that conforms to the requested output schema.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Classify this router incident: customer traffic is slow in us-east, the OpenAI upstream is rate limited, the Anthropic upstream is healthy, and retry pressure is rising.")},
				},
			},
		},
	},
	ChatResp: &v1.ChatResp{
		Id:     "gen-1782736289-Jbmm6j69oJCX7yvuuT7d",
		Model:  "openai/gpt-4o",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent(`{"severity":"medium","affected_region":"us-east","primary_cause":"Rate limiting of the OpenAI upstream","recommended_action":"Investigate and adjust rate limiting settings or policies for the OpenAI upstream to alleviate traffic congestion and reduce retry pressure.","signals":["Slow customer traffic in us-east","OpenAI upstream is rate limited","Anthropic upstream is healthy","Rising retry pressure"]}`)},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 135, OutputTokens: 77},
		},
	},
}
