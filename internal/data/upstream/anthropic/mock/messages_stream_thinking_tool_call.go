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

//go:embed messages_stream_thinking_tool_call_request.json
var streamThinkingToolCallRequest []byte

//go:embed messages_stream_thinking_tool_call_response.txt
var streamThinkingToolCallResponse []byte

const streamThinkingToolCallSignature = "EpMECmUIDxgCKkAMzSFbICKIRiy28yCmtpHD3W9xVqM7a1tdgOfnhh2zP4kJNR+mpQvy1AYiedPeMPVurbAPgpsmE0T9mUQRP90lMhFjbGF1ZGUtc29ubmV0LTQtNjgAQgh0aGlua2luZxIMbRhXZGm0zbkQLdU7GgyvCHsg1q7KPIAqmSsiMF9xvEpD7sZkb8t4vAYmEv2iYzDJ7n1vZJoMsPW/AkGYSkmPESFMHkfgSQEVZOJigSrbAmEtLfVMUEl8cv/vC6wEtvygtHjSqgj5kkgGD/ZBVEpAphcir6NrXzkl5LrkxSLoXBB9rjE1TrLdFOol57PRiNuedenRQTyIR94yBdGymzmTXNjOBAeCkNEF0VQgPA5zN982qnvDH5h0vfp8f+JVue02J3sCZOdRQ9fQGLIqK2zNxk5AKbFxYTZSOG1zdAth26ZhVdrElqUpw/4KKm0Z7bYHjVwa1nbBzKMFWOVgPhwSPUoyTwM96OP8vjsj3aoLtSfxyLebJQYn6RGz/IxjtHMOn6uhLTMQoPC4yRDUXpOYnpSdnjayU5uTAIIND22SNc3px1KPagoBbjq9MFmB3eCRsisiIw6GSqshd8IJEH/nn3BWRy1dmuc2T4wyYKdAiaDWwfZ2eY7ZhBL1t+Pq1JB7JJGICMLqz456kj3r1hqDCqMq04vHPFrPIttjY2RUIPSP4JhENa7emTkkGAE="

const streamThinkingToolCallRedacted = "openrouter.reasoning:eyJzaWduYXR1cmUiOiJFcE1FQ21VSUR4Z0NLa0FNelNGYklDS0lSaXkyOHlDbXRwSEQzVzl4VnFNN2ExdGRnT2ZuaGgyelA0a0pOUittcFF2eTFBWWllZFBlTVBWdXJiQVBncHNtRTBUOW1VUVJQOTBsTWhGamJHRjFaR1V0YzI5dWJtVjBMVFF0TmpnQVFnaDBhR2x1YTJsdVp4SU1iUmhYWkdtMHpia1FMZFU3R2d5dkNIc2cxcTdLUElBcW1Tc2lNRjl4dkVwRDdzWmtiOHQ0dkFZbUV2MmlZekRKN24xdlpKb01zUFcvQWtHWVNrbVBFU0ZNSGtmZ1NRRVZaT0ppZ1NyYkFtRXRMZlZNVUVsOGN2L3ZDNndFdHZ5Z3RIalNxZ2o1a2tnR0QvWkJWRXBBcGhjaXI2TnJYemtsNUxya3hTTG9YQkI5cmpFMVRyTGRGT29sNTdQUmlOdWVkZW5SUVR5SVI5NHlCZEd5bXptVFhOak9CQWVDa05FRjBWUWdQQTV6Tjk4MnFudkRINWgwdmZwOGYrSlZ1ZTAySjNzQ1pPZFJROWZRR0xJcUsyek54azVBS2JGeFlUWlNPRzF6ZEF0aDI2WmhWZHJFbHFVcHcvNEtLbTBaN2JZSGpWd2ExbmJCektNRldPVmdQaHdTUFVveVR3TTk2T1A4dmpzajNhb0x0U2Z4eUxlYkpRWW42Ukd6L0l4anRITU9uNnVoTFRNUW9QQzR5UkRVWHBPWW5wU2RuamF5VTV1VEFJSU5EMjJTTmMzcHgxS1BhZ29CYmpxOU1GbUIzZUNSc2lzaUl3NkdTcXNoZDhJSkVIL25uM0JXUnkxZG11YzJUNHd5WUtkQWlhRFd3ZloyZVk3WmhCTDF0K1BxMUpCN0pKR0lDTUxxejQ1NmtqM3IxaHFEQ3FNcTA0dkhQRnJQSXR0alkyUlVJUFNQNEpoRU5hN2VtVGtrR0FFPSIsInRleHQiOiJUaGUgdXNlciB3YW50cyBtZSB0byBjaG9vc2UgdGhlIGJlc3QgcHJvdmlkZXIgYmFzZWQgb24gd2FpdF9tcyBhbmQgZXJyb3JfcmF0ZSBtZXRyaWNzOlxuXG4tIE9wZW5BSTogd2FpdF9tcz0xODAwLCBlcnJvcl9yYXRlPTAuMTJcbi0gQW50aHJvcGljOiB3YWl0X21zPTEyMCwgZXJyb3JfcmF0ZT0wLjAxXG4tIEdlbWluaTogd2FpdF9tcz00MDAsIGVycm9yX3JhdGU9MC4wNFxuXG5BbnRocm9waWMgY2xlYXJseSB3aW5zIG9uIGJvdGggbWV0cmljcyAtIGxvd2VzdCB3YWl0IHRpbWUgKDEyMG1zKSBhbmQgbG93ZXN0IGVycm9yIHJhdGUgKDAuMDEpLiBJJ2xsIHJvdXRlIHRvIEFudGhyb3BpYy4iLCJ0eXBlIjoicmVhc29uaW5nLnRleHQifQ=="

func streamThinkingToolCallChatEvents() []*v1.ChatEvent {
	id := eventBuilder("stream_thinking_tool_call")
	return []*v1.ChatEvent{
		id.of(v1.NewMessageStartEvent("gen-1782639618-FBlJV7iR8SzgzHP8zhu2", "anthropic/claude-4.6-sonnet-20260217")),
		id.of(v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_REASONING)),
		id.of(v1.NewContentDeltaTextEvent(0, "The")),
		id.of(v1.NewContentDeltaTextEvent(0, " user wants me to choose the best provider based on wait_ms and error_rate metrics:\n\n- OpenAI: wait_ms=1800, error_rate=")),
		id.of(v1.NewContentDeltaTextEvent(0, "0.12\n- Anthropic: wait_ms=120, error_rate=0.01\n- Gemini: wait_ms=400, error_rate=0.04\n\nAnthropic clearly")),
		id.of(v1.NewContentDeltaTextEvent(0, " wins on both metrics - lowest wait time (120ms) and lowest error rate")),
		id.of(v1.NewContentDeltaTextEvent(0, " (0.01). I'll route to Anthropic.")),
		id.of(v1.NewContentDeltaSignatureEvent(0, streamThinkingToolCallSignature)),
		id.of(v1.NewContentStopEvent(0)),
		id.of(v1.NewContentStartTextEvent(1, v1.ContentPhase_CONTENT_PHASE_NORMAL)),
		id.of(v1.NewContentDeltaTextEvent(1, "Anthropic is the clear winner here —")),
		id.of(v1.NewContentDeltaTextEvent(1, " it has both the **lowest wait time** (120 ms vs. 400")),
		id.of(v1.NewContentDeltaTextEvent(1, " ms and 1800 ms) and the **lowest error rate** (0.01 vs. 0.04 and 0.12). No")),
		id.of(v1.NewContentDeltaTextEvent(1, " trade-off needed; I'll route straight to Anthropic.")),
		id.of(v1.NewContentStopEvent(1)),
		id.of(v1.NewContentStartToolUseEvent(2, "toolu_01GqmkpGoJRfAT2zBdva1bgV", "get_route_decision")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, `{"request_id": "req_fixture_stream_001`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, `", "provider": "anthropic`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, `", "observed_wait_ms": 120`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, `, "reason": "Anthropic has the lowest wait time (120 ms) and the lowest error rate (0.01),`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, ` outperforming both OpenAI and Gemini on every relevant metric.`)),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, `"}`)),
		id.of(v1.NewContentStopEvent(2)),
		id.of(v1.NewContentSnapshotEvent(&v1.Content{
			Index:   new(uint32(3)),
			Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
			Content: &v1.Content_Opaque{Opaque: streamThinkingToolCallRedacted},
		})),
		id.withUsage(
			v1.NewMessageStopEvent(v1.ChatStatus_CHAT_PENDING_TOOL_USE),
			&v1.Usage{InputTokens: 743, OutputTokens: 360},
		),
	}
}

// StreamThinkingToolCall covers a streaming request that enables adaptive
// thinking at high effort and asks for a tool call.
var StreamThinkingToolCall = &Fixture{
	Name:     "stream_thinking_tool_call",
	Request:  streamThinkingToolCallRequest,
	Response: streamThinkingToolCallResponse,
	Stream:   true,
	ChatReq: &v1.ChatReq{
		Id:    "stream_thinking_tool_call",
		Model: "anthropic/claude-sonnet-4.6",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(2048)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_HIGH},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Think briefly, then use the routing decision tool exactly once.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Choose a provider for chat request req_fixture_stream_001. OpenAI has wait_ms=1800 and error_rate=0.12. Anthropic has wait_ms=120 and error_rate=0.01. Gemini has wait_ms=400 and error_rate=0.04. Call get_route_decision with the best provider and a short reason.")},
				},
			},
		},
		Tools: []*v1.Tool{
			{
				Tool: &v1.Tool_Function_{
					Function: &v1.Tool_Function{
						Name:        "get_route_decision",
						Description: "Record the selected upstream provider for a chat request.",
						InputSchema: util.MustStructFromMap(map[string]any{
							"type": "object",
							"properties": map[string]any{
								"request_id": map[string]any{"type": "string"},
								"provider": map[string]any{
									"type": "string",
									"enum": []string{"openai", "anthropic", "gemini"},
								},
								"reason":           map[string]any{"type": "string"},
								"observed_wait_ms": map[string]any{"type": "integer", "minimum": 0},
							},
							"required":             []string{"request_id", "provider", "reason"},
							"additionalProperties": false,
						}),
					},
				},
			},
		},
		Metadata: map[string]string{"user_id": "anthropic-conversion-fixture-user"},
	},
	ChatEvents: streamThinkingToolCallChatEvents(),
}
