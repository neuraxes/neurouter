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

//go:embed chat_completion_stop_sequence_request.json
var stopSequenceRequest []byte

//go:embed chat_completion_stop_sequence_response.json
var stopSequenceResponse []byte

// StopSequence covers a request carrying a stop sequence.
var StopSequence = &Fixture{
	Name:     "stop_sequence",
	Request:  stopSequenceRequest,
	Response: stopSequenceResponse,
	ChatReq: &v1.ChatReq{
		Id:    "stop_sequence",
		Model: "openai/gpt-4o",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(64)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_NONE},
			StopSequences:   []string{"5"},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Follow formatting instructions exactly and output nothing else.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Output the digits one through nine separated by single spaces, exactly like this: 1 2 3 4 5 6 7 8 9. Output only those digits and spaces, with no other words.")},
				},
			},
		},
	},
	ChatResp: &v1.ChatResp{
		Id:     "gen-1782736296-MRi1skRx6CDE9SAebYi6",
		Model:  "openai/gpt-4o",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "mock_message_id",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent("1 2 3 4 ")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 73, OutputTokens: 9},
		},
	},
}
