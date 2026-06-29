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

//go:embed chat_completion_max_tokens_request.json
var maxTokensRequest []byte

//go:embed chat_completion_max_tokens_response.json
var maxTokensResponse []byte

// MaxTokens covers a request that caps output at 64 tokens, producing a reply
// truncated with finish reason length (which maps to a token-limit status).
var MaxTokens = &Fixture{
	Name:     "max_tokens",
	Request:  maxTokensRequest,
	Response: maxTokensResponse,
	ChatReq: &v1.ChatReq{
		Id:    "max_tokens",
		Model: "openai/gpt-4o",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(64)),
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
	},
	ChatResp: &v1.ChatResp{
		Id:     "gen-1782736293-pjqokdGS8Qo3GXjW4x9E",
		Model:  "openai/gpt-4o",
		Status: v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT,
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent("A Large Language Model (LLM) router balances the load across multiple upstream providers through a sophisticated process that involves probing, ranking, reservation, and rate limiting. These steps ensure optimized utilization of resources, reduced latency, and improved performance.\n\n**Probing:** The LLM router begins by continuously monitoring the health and performance of")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 56, OutputTokens: 64},
		},
	},
}
