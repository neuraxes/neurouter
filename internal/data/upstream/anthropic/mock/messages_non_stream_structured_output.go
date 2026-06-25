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

//go:embed messages_non_stream_structured_output_request.json
var nonStreamStructuredOutputRequest []byte

//go:embed messages_non_stream_structured_output_response.json
var nonStreamStructuredOutputResponse []byte

// NonStreamStructuredOutput covers a request that constrains the reply with a
// JSON schema (grammar) and whose response is a single JSON text block with
// stop reason end_turn.
var NonStreamStructuredOutput = &Fixture{
	Name:     "non_stream_structured_output",
	Request:  nonStreamStructuredOutputRequest,
	Response: nonStreamStructuredOutputResponse,
	ChatReq: &v1.ChatReq{
		Id:    "non_stream_structured_output",
		Model: "anthropic/claude-sonnet-4.6",
		Config: &v1.GenerationConfig{
			MaxTokens:   new(int64(600)),
			Temperature: new(float32(0)),
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
		Metadata: map[string]string{"user_id": "anthropic-conversion-fixture-user"},
	},
	ChatResp: &v1.ChatResp{
		Id:     "non_stream_structured_output",
		Model:  "anthropic/claude-4.6-sonnet-20260217",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "gen-1782639574-Xno06ITSrNECtJCEjTXG",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent(`{"severity":"high","affected_region":"us-east","primary_cause":"OpenAI upstream rate limiting causing increased retry pressure and degraded customer traffic throughput","recommended_action":"Immediately shift us-east customer traffic from OpenAI upstream to the healthy Anthropic upstream, implement backoff strategies to reduce retry pressure, and monitor OpenAI rate limit recovery before rebalancing","signals":["Customer traffic slow in us-east","OpenAI upstream is rate limited","Anthropic upstream is healthy","Retry pressure is rising"]}`)},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 356, OutputTokens: 118},
		},
	},
}
