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

// Package mock holds captured Anthropic Messages API request/response pairs
// together with the neurouter entity values that the anthropic upstream
// conversion must produce for them.
package mock

import (
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

// Fixture pairs captured Anthropic payloads with the neurouter entities that the
// upstream conversion must produce in both directions.
type Fixture struct {
	// Name identifies the fixture in test output.
	Name string
	// Request is the captured Anthropic Messages API request body. It is the
	// expected output of convertRequestToAnthropic for ChatReq.
	Request []byte
	// Response is the captured upstream reply: a JSON message for non-stream
	// fixtures, or a raw SSE event stream for stream fixtures.
	Response []byte
	// Stream reports whether Response is an SSE stream that converts into
	// ChatEvents rather than a single ChatResp.
	Stream bool
	// ChatReq is the neurouter request that must marshal into Request.
	ChatReq *v1.ChatReq
	// ChatResp is the expected conversion of Response for non-stream fixtures.
	ChatResp *v1.ChatResp
	// ChatEvents is the expected conversion of Response for stream fixtures.
	ChatEvents []*v1.ChatEvent
}

// Fixtures is the full conversion fixture set, aggregated from the per-fixture
// files in this package.
var Fixtures = []*Fixture{
	NonStreamToolCall,
	NonStreamToolResult,
	NonStreamStructuredOutput,
	NonStreamThinking,
	NonStreamVision,
	NonStreamMaxTokens,
	NonStreamStopSequence,
	StreamThinkingToolCall,
	StreamThinkingText,
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
