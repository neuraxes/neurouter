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

// Package mock holds captured OpenAI Chat Completions API request/response
// pairs together with the neurouter entity values that the openai upstream
// conversion must produce for them.
package mock

import (
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/util"
)

// Fixture pairs captured OpenAI payloads with the neurouter entities that the
// upstream conversion must produce in both directions.
type Fixture struct {
	// Name identifies the fixture in test output.
	Name string
	// Request is the hand-authored ground-truth Chat Completions request body.
	// It is the expected output of convertRequestToOpenAIChat for ChatReq.
	Request []byte
	// Response is the captured upstream reply: a JSON completion for non-stream
	// fixtures, or a raw SSE event stream for stream fixtures.
	Response []byte
	// Stream reports whether Response is an SSE stream that converts into
	// ChatEvents rather than a single ChatResp.
	Stream bool
	// ChatReq is the neurouter request that must convert into Request.
	ChatReq *v1.ChatReq
	// ChatResp is the expected conversion of Response for non-stream fixtures.
	ChatResp *v1.ChatResp
	// ChatEvents is the expected conversion of Response for stream fixtures.
	ChatEvents []*v1.ChatEvent
}

// Fixtures is the full conversion fixture set, aggregated from the per-fixture
// files in this package.
var Fixtures = []*Fixture{
	ToolCall,
	ToolResult,
	StructuredOutput,
	Vision,
	MaxTokens,
	StopSequence,
	StreamToolCall,
}

// eventBuilder constructs ChatEvents that all carry the same request id.
type eventBuilder string

func (id eventBuilder) of(payload v1.ChatEventPayload) *v1.ChatEvent {
	return v1.NewChatEvent(string(id), payload)
}

func (id eventBuilder) withUsage(payload v1.ChatEventPayload, usage *v1.Usage) *v1.ChatEvent {
	event := v1.NewChatEvent(string(id), payload)
	event.Usage = usage
	return event
}

// getWeatherTool is the shared tool definition used by the tool-call,
// tool-result and stream tool-call fixtures.
func getWeatherTool() *v1.Tool {
	return &v1.Tool{
		Tool: &v1.Tool_Function_{
			Function: &v1.Tool_Function{
				Name:        "get_weather",
				Description: "Look up historical weather for a city on a specific date.",
				InputSchema: util.MustStructFromMap(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"city": map[string]any{
							"type":        "string",
							"description": "City name in English.",
						},
						"date": map[string]any{
							"type":        "string",
							"description": "Date in YYYY-MM-DD format.",
						},
						"units": map[string]any{
							"type":        "string",
							"description": "Temperature unit.",
							"enum":        []string{"metric", "imperial"},
						},
					},
					"required":             []string{"city", "date"},
					"additionalProperties": false,
				}),
			},
		},
	}
}
