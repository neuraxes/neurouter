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

//go:embed messages_non_stream_max_tokens_request.json
var nonStreamMaxTokensRequest []byte

//go:embed messages_non_stream_max_tokens_response.json
var nonStreamMaxTokensResponse []byte

// NonStreamMaxTokens covers a request that disables thinking and caps output at
// 16 tokens, producing a response truncated with stop reason max_tokens.
var NonStreamMaxTokens = &Fixture{
	Name:     "non_stream_max_tokens",
	Request:  nonStreamMaxTokensRequest,
	Response: nonStreamMaxTokensResponse,
	ChatReq: &v1.ChatReq{
		Id:    "non_stream_max_tokens",
		Model: "anthropic/claude-sonnet-4.6",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(16)),
			Temperature:     new(float32(0)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_NONE},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Answer in long, detailed prose.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Write a detailed 300-word explanation of how an LLM router balances load across multiple upstream providers, covering probing, ranking, reservation, and rate limiting.")},
				},
			},
		},
		Metadata: map[string]string{"user_id": "anthropic-conversion-fixture-user"},
	},
	ChatResp: &v1.ChatResp{
		Id:     "non_stream_max_tokens",
		Model:  "anthropic/claude-4.6-sonnet-20260217",
		Status: v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT,
		Message: &v1.Message{
			Id:   "gen-1782639366-sxvOUZF4maUOUI6sgtQy",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent("# How an LLM Router Balances Load Across Multiple Upstream")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 58, OutputTokens: 16},
		},
	},
}
