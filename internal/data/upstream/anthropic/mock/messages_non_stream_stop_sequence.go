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

//go:embed messages_non_stream_stop_sequence_request.json
var nonStreamStopSequenceRequest []byte

//go:embed messages_non_stream_stop_sequence_response.json
var nonStreamStopSequenceResponse []byte

// NonStreamStopSequence covers a request carrying a stop sequence; the model
// halts right before emitting "5", yielding stop reason stop_sequence (which
// maps to a completed status).
var NonStreamStopSequence = &Fixture{
	Name:     "non_stream_stop_sequence",
	Request:  nonStreamStopSequenceRequest,
	Response: nonStreamStopSequenceResponse,
	ChatReq: &v1.ChatReq{
		Id:    "non_stream_stop_sequence",
		Model: "anthropic/claude-sonnet-4.6",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(64)),
			Temperature:     new(float32(0)),
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
		Metadata: map[string]string{"user_id": "anthropic-conversion-fixture-user"},
	},
	ChatResp: &v1.ChatResp{
		Id:     "non_stream_stop_sequence",
		Model:  "anthropic/claude-4.6-sonnet-20260217",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "gen-1782639367-LWXTlYSwUf9TmicH9TYr",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: v1.NewTextContent("1 2 3 4 ")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 71, OutputTokens: 10},
		},
	},
}
