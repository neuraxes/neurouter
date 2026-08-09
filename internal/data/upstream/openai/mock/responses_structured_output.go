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

//go:embed responses_structured_output_request.json
var responsesStructuredOutputRequest []byte

//go:embed responses_structured_output_response.json
var responsesStructuredOutputResponse []byte

// ResponsesStructuredOutput covers a strict JSON schema response format.
var ResponsesStructuredOutput = &Fixture{
	Name:     "responses_structured_output",
	Request:  responsesStructuredOutputRequest,
	Response: responsesStructuredOutputResponse,
	ChatReq: &v1.ChatReq{
		Id:    "responses_structured_output",
		Model: "openai/gpt-5-mini",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(1024)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_MINIMAL},
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
		Id:     "responses_structured_output",
		Model:  "openai/gpt-5-mini",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "gen-1785502384-ednipl8xASaFIwjUjxAV",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Id:    "rs_062432a4b1a36741016a6c9ab122d081a197d277b7596be463",
					Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{
						Opaque: "gAAAAABqbJqxhA1H89QMf-XNMMf_VGn2BFxpi-QXvWHWHcFnJuR88furZacDDJ9ffC5bZPYq6vrIjM6aDsaSQAhCzKi-eLsSZlnLg-nWxW3YZe5q3p5WOwHlWiDQAKwUFTKcxMGX3NibK4FHgxbWuzxZbt-siV6Tn3IGG87ht5IKCmMwsWz9rf6Cl5G5mAdB_F8DuNPBoLwvmo-vJsjC2-sl7VRser1rUU68tgb213r5MZveLY5__HoijHxeM1lVmAOKeCyZlG2HxY_zWB37_92nvd72L9lUgETNXyH_WGKwi9ti-WpIS7KJpAxa5-G-qXjm3OZ8kZy3d9XNDiGB7DQM6ChFVTjLAEKEgt2z7Sjx4iuxGTlVOy0UNsfGFdAQK9qqTvtEbM0l4HHEAxdNn0VBvqZbLPGMsZoT4sFfQ7ZUXzzVUdtLnUrmpbhyFD6BmSCS7pEgkLet6y8jIQNkVO-xtZEAjqPG1cirQWcEpynUen54FgBQEQOLlNMX5EZEebp9HFvELUnb-X_STWUIys1Rrt5oveAX4WsrFriDsXbD18v6BFXjodaKaAIniDhqbR8EFteQ3qN1daa4Az4D2jLgHPN3mNhtAfVHzNyaLn2foOwhn0Nlv1uc0W6QQnImiolEWebYsCmIxKxjwnF_grbjzTL1WrbXk5leEoFdyDLtKy0nS3hUDFi7OZpFLWPVHd-kPkFxlqiXCJbdNMi7S6XTnaRGaESwI05AsOxKFSbnaX7IF_GWWuEShWOrq-hb367quElBF3U1VyoUGqT9jkf2PURmIuOFDn9aT4l8Q6H8TQE0gJnAEJuI2CNkwCaL1n_rUgYp-J5szMGNJDzn7tfTkokEwwcw6vuts7fGZIs934FNnNwKUDu7w-7odFOAOzT62T9u2NFrHwE2bYI3mitr0-g8stkutmeW5_LOY44gRzefSUxAf7d9GCPZERVLncgdyqIgvAiEAt1DmrscXFgUyRPkMaL3atcAVdhsZWa_M9wEoZdhFZU=",
					},
				},
				{Content: v1.NewTextContent(`{"severity":"high","affected_region":"us-east","primary_cause":"OpenAI upstream rate limiting causing increased retries and elevated retry pressure","recommended_action":"Investigate and mitigate OpenAI upstream rate limits (coordinate with upstream provider, adjust client rate limits or backoff settings). Throttle nonessential traffic, increase capacity or failover to healthy upstreams where possible, and monitor retry queue depth and latency until retry pressure subsides.","signals":["customer traffic slow in us-east","OpenAI upstream is rate limited","Anthropic upstream healthy","retry pressure rising"]}`)},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 130, OutputTokens: 134},
		},
	},
}
