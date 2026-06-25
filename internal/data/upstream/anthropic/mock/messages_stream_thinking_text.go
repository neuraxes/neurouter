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

//go:embed messages_stream_thinking_text_request.json
var streamThinkingTextRequest []byte

//go:embed messages_stream_thinking_text_response.txt
var streamThinkingTextResponse []byte

const streamThinkingTextSignature = "EokHCmUIDxgCKkBX0fNiSjOUrJmGbxV0gtSJSwQ6iA4GCxEd3KATwjOiKyL/kEjzcGo+7kk7r59j3iJ+dEGVxEXj23styoJlumN4MhFjbGF1ZGUtc29ubmV0LTQtNjgAQgh0aGlua2luZxIMpinNifR8Sh/XrEz/Ggwz+4BQnUaL8+jRLywiMO7Yi60rgGjTerXaUNE3ycpGVky/QwgSeB+e9IV2eUL4U1hQ2R5Ly2tqGNhUjcX3JirRBaFBIDtpe7VAyeyIXp1TDYIt4Poew+DVbX3DFoL9qTjCm2C1SLXZ5joeL0iYUqSul9eAWsCdLhjJ23LSofjuwhEr5LD1XfJtA+bzSCdlWkusvqCzte8g4+1WJi1ILCrLuqYwfxmSALCXCQiANrftVn+941cKPV1ZGUAootkX4Lj68Tkig+ngmM/WnZmVPoy3rCJcOyxwufpmpaQPQiL0SrPVYRlu+9Q2qX42ggdfJF9FebMpl76UFBqz1f0gFDnlkFc+MabKLwQPE+zsZwJFGdmvwgdE7aR/2RHA8JCnMF5FFh9RGk0nMAMjqbyv2DpPuazsxQqyJQMb36OeBY/1jn6B9GnusAiPHCPXOuq1jJLLyhfqnhY/NZhSxBBz6nZiZVVFxrzhadQrOdKfLfFxRw58vCD8pw2IAE90xsBjMWGsTTsK7/UH/JBrYwFd/MEE9LeGyvhWYl5aLIA+gqrtPxV/5oKnTOCgEkLu1xJe0RcaFeWqr2saQRm28JDXDv2xKQEfTs/POhFec8zrIDrIysKQhonqtrsmFej3bZu3dXoDcNwJHnMyGhcPjSxGKImtoJrs64pgZK/QOHr+Vlyy90qTjOhXFgtPUNWndHGocxfxqaNy/sA96rV1niYzCEtwioTeBtk8jxE+lA7Gso7W5yhna8CmdI04qCFatshj3/tTPPdgjmLNJjAqXyrmz8pY7fE4s6KsXomH48B3D+Fh/T/uPFscmP3Pnck+sNcSqW/lcEh/ONBXAjvwBhPskY5uitr4hLMVjs/EzuETMu0zWEtvndvvfucaQjiUqGx6f0rcWtrjQPxKcM8bsMLTqlZNb+k0vSpWSrMWPTMis/IcV3bs5SfLHYzy2LzPYItiibX1xnPBjjt+vjumWQ5UeyQ89nR8jIKXwjMcR8XmCAzr1cJKKlji62p6ZvIiBBJgOrnVh8H332qnTgXjy6TLvJAGAQIYAQ=="

const streamThinkingTextRedacted = "openrouter.reasoning:eyJzaWduYXR1cmUiOiJFb2tIQ21VSUR4Z0NLa0JYMGZOaVNqT1VySm1HYnhWMGd0U0pTd1E2aUE0R0N4RWQzS0FUd2pPaUt5TC9rRWp6Y0dvKzdrazdyNTlqM2lKK2RFR1Z4RVhqMjNzdHlvSmx1bU40TWhGamJHRjFaR1V0YzI5dWJtVjBMVFF0TmpnQVFnaDBhR2x1YTJsdVp4SU1waW5OaWZSOFNoL1hyRXovR2d3eis0QlFuVWFMOCtqUkx5d2lNTzdZaTYwcmdHalRlclhhVU5FM3ljcEdWa3kvUXdnU2VCK2U5SVYyZVVMNFUxaFEyUjVMeTJ0cUdOaFVqY1gzSmlyUkJhRkJJRHRwZTdWQXlleUlYcDFURFlJdDRQb2V3K0RWYlgzREZvTDlxVGpDbTJDMVNMWFo1am9lTDBpWVVxU3VsOWVBV3NDZExoakoyM0xTb2ZqdXdoRXI1TEQxWGZKdEErYnpTQ2RsV2t1c3ZxQ3p0ZThnNCsxV0ppMUlMQ3JMdXFZd2Z4bVNBTENYQ1FpQU5yZnRWbis5NDFjS1BWMVpHVUFvb3RrWDRMajY4VGtpZytuZ21NL1duWm1WUG95M3JDSmNPeXh3dWZwbXBhUVBRaUwwU3JQVllSbHUrOVEycVg0MmdnZGZKRjlGZWJNcGw3NlVGQnF6MWYwZ0ZEbmxrRmMrTWFiS0x3UVBFK3pzWndKRkdkbXZ3Z2RFN2FSLzJSSEE4SkNuTUY1RkZoOVJHazBuTUFNanFieXYyRHBQdWF6c3hRcXlKUU1iMzZPZUJZLzFqbjZCOUdudXNBaVBIQ1BYT3VxMWpKTEx5aGZxbmhZL05aaFN4QkJ6Nm5aaVpWVkZ4cnpoYWRRck9kS2ZMZkZ4Unc1OHZDRDhwdzJJQUU5MHhzQmpNV0dzVFRzSzcvVUgvSkJyWXdGZC9NRUU5TGVHeXZoV1lsNWFMSUErZ3FydFB4Vi81b0tuVE9DZ0VrTHUxeEplMFJjYUZlV3FyMnNhUVJtMjhKRFhEdjJ4S1FFZlRzL1BPaEZlYzh6cklEckl5c0tRaG9ucXRyc21GZWozYlp1M2RYb0RjTndKSG5NeUdoY1BqU3hHS0ltdG9KcnM2NHBnWksvUU9IcitWbHl5OTBxVGpPaFhGZ3RQVU5XbmRIR29jeGZ4cWFOeS9zQTk2clYxbmlZekNFdHdpb1RlQnRrOGp4RStsQTdHc283VzV5aG5hOENtZEkwNHFDRmF0c2hqMy90VFBQZGdqbUxOSmpBcVh5cm16OHBZN2ZFNHM2S3NYb21INDhCM0QrRmgvVC91UEZzY21QM1BuY2src05jU3FXL2xjRWgvT05CWEFqdndCaFBza1k1dWl0cjRoTE1WanMvRXp1RVRNdTB6V0V0dm5kdnZmdWNhUWppVXFHeDZmMHJjV3RyalFQeEtjTThic01MVHFsWk5iK2swdlNwV1NyTVdQVE1pcy9JY1YzYnM1U2ZMSFl6eTJMelBZSXRpaWJYMXhuUEJqanQrdmp1bVdRNVVleVE4OW5SOGpJS1h3ak1jUjhYbUNBenIxY0pLS2xqaTYycDZadklpQkJKZ09yblZoOEgzMzJxblRnWGp5NlRMdkpBR0FRSVlBUT09IiwidGV4dCI6IkxldCBtZSBjYWxjdWxhdGUgdGhlIGV4cGVjdGVkIGxhdGVuY3kgZm9yIGVhY2ggdXBzdHJlYW0gd2l0aCBvbmUgcmV0cnkgb24gZmFpbHVyZS5cblxuRm9yIGEgc2luZ2xlIGF0dGVtcHQ6XG4tIFN1Y2Nlc3Mgb24gZmlyc3QgdHJ5OiBsYXRlbmN5ID0gTFxuLSBGYWlsdXJlIG9uIGZpcnN0IHRyeSwgdGhlbiByZXRyeTpcbiAgLSBTdWNjZXNzIG9uIHJldHJ5OiBsYXRlbmN5ID0gTCArIEwgPSAyTFxuICAtIEZhaWx1cmUgb24gcmV0cnk6IGxhdGVuY3kgPSAyTCAod2Ugc3RvcCBhZnRlciBvbmUgcmV0cnkpXG5cbkV4cGVjdGVkIGxhdGVuY3kgPSBMICogKDEgLSBlKSArIDJMICogZSA9IEwgKiAoMSArIGUpXG5cbldoZXJlIGUgaXMgdGhlIGVycm9yIHJhdGUgYW5kIEwgaXMgbGF0ZW5jeS5cblxuV2FpdCwgYnV0IG9uIGZhaWx1cmUgd2Ugc3RpbGwgaW5jdXIgdGhlIGxhdGVuY3kgb2YgdGhlIGZhaWxlZCByZXF1ZXN0IGJlZm9yZSByZXRyeWluZy5cbkV4cGVjdGVkIGxhdGVuY3kgPSBMICsgZSAqIEwgPSBMKDEgKyBlKVxuXG5Ob3cgSSBjYW4gcGx1ZyBpbiB0aGUgbnVtYmVycyBmb3IgZWFjaCB1cHN0cmVhbS4gVXBzdHJlYW0gQSBoYXMgMjAwbXMgbGF0ZW5jeSB3aXRoIDUlIGVycm9yIHJhdGUsIGdpdmluZyAyMTBtcyBleHBlY3RlZCBsYXRlbmN5LiBVcHN0cmVhbSBCIGhhcyAxNTBtcyBsYXRlbmN5IHdpdGggOSUgZXJyb3IgcmF0ZSwgZ2l2aW5nIDE2My41bXMgZXhwZWN0ZWQgbGF0ZW5jeS4gU28gQiBpcyB0aGUgYmV0dGVyIGNob2ljZSBkZXNwaXRlIHRoZSBoaWdoZXIgZXJyb3IgcmF0ZS4iLCJ0eXBlIjoicmVhc29uaW5nLnRleHQifQ=="

func streamThinkingTextChatEvents() []*v1.ChatEvent {
	id := eventBuilder("stream_thinking_text")
	return []*v1.ChatEvent{
		id.of(v1.NewMessageStartEvent("gen-1782640358-8xpFd0HzF6SZnlURyaC9", "anthropic/claude-4.6-sonnet-20260217")),
		id.of(v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_REASONING)),
		id.of(v1.NewContentDeltaTextEvent(0, "Let me calculate")),
		id.of(v1.NewContentDeltaTextEvent(0, " the expected latency for each upstream with one retry on failure.\n\nFor a single attempt:\n- Success on")),
		id.of(v1.NewContentDeltaTextEvent(0, " first try: latency = L\n- Failure on first try, then retry:\n  - Success on retry: latency = L + L")),
		id.of(v1.NewContentDeltaTextEvent(0, " = 2L\n  - Failure on retry: latency = 2L (we stop after one")),
		id.of(v1.NewContentDeltaTextEvent(0, " retry)\n\nExpected latency = L * (1 - e) + 2L * e = L * (1 +")),
		id.of(v1.NewContentDeltaTextEvent(0, " e)\n\nWhere e is the error rate and L is latency.\n\nWait, but on")),
		id.of(v1.NewContentDeltaTextEvent(0, " failure we still incur the latency of the failed request")),
		id.of(v1.NewContentDeltaTextEvent(0, " before retrying.\n")),
		id.of(v1.NewContentDeltaTextEvent(0, "Expected")),
		id.of(v1.NewContentDeltaTextEvent(0, " latency = L + e * L = L(1 + e)\n\nNow I can plug in the numbers")),
		id.of(v1.NewContentDeltaTextEvent(0, " for each upstream. Upstream A has 200ms latency with")),
		id.of(v1.NewContentDeltaTextEvent(0, " 5% error rate, giving 210ms expected latency. Upstream B has 150ms latency with 9")),
		id.of(v1.NewContentDeltaTextEvent(0, "% error rate, giving 163.5ms expected latency. So B is the better choice despite the")),
		id.of(v1.NewContentDeltaTextEvent(0, " higher error rate.")),
		id.of(v1.NewContentDeltaSignatureEvent(0, streamThinkingTextSignature)),
		id.of(v1.NewContentStopEvent(0)),
		id.of(v1.NewContentStartTextEvent(1, v1.ContentPhase_CONTENT_PHASE_NORMAL)),
		id.of(v1.NewContentDeltaTextEvent(1, "## Reasoning\n\nWith")),
		id.of(v1.NewContentDeltaTextEvent(1, " one retry on failure, expected latency = L × (1 + e), since a failed attempt still cons")),
		id.of(v1.NewContentDeltaTextEvent(1, "umes full latency before the retry.\n\n- **A:** 200 × (1 + 0.05) = **210 ms**\n- **B")),
		id.of(v1.NewContentDeltaTextEvent(1, ":** 150 × (1 + 0.09) = **163.5 ms**\n\nDespite B's higher error rate, its lower base latency domin")),
		id.of(v1.NewContentDeltaTextEvent(1, "ates; the retry penalty (9% × 150 ms = 13.5 ms) is smaller")),
		id.of(v1.NewContentDeltaTextEvent(1, " than A's retry penalty (5% × 200 ms = 10 ms) in absolute terms, but B's base")),
		id.of(v1.NewContentDeltaTextEvent(1, " is so much faster that it wins comfortably.\n\n**Pick B**,")),
		id.of(v1.NewContentDeltaTextEvent(1, " as its expected latency of ~163.5 ms is well below A's ~")),
		id.of(v1.NewContentDeltaTextEvent(1, "210 ms when accounting for one retry on failure.")),
		id.of(v1.NewContentStopEvent(1)),
		id.of(v1.NewContentSnapshotEvent(&v1.Content{
			Index:   new(uint32(2)),
			Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
			Content: &v1.Content_Opaque{Opaque: streamThinkingTextRedacted},
		})),
		id.withUsage(
			v1.NewMessageStopEvent(v1.ChatStatus_CHAT_COMPLETED),
			&v1.Usage{InputTokens: 123, OutputTokens: 459},
		),
	}
}

// StreamThinkingText covers a streaming request that enables extended thinking
// with a token budget.
var StreamThinkingText = &Fixture{
	Name:     "stream_thinking_text",
	Request:  streamThinkingTextRequest,
	Response: streamThinkingTextResponse,
	Stream:   true,
	ChatReq: &v1.ChatReq{
		Id:    "stream_thinking_text",
		Model: "anthropic/claude-sonnet-4.6",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(4096)),
			ReasoningConfig: &v1.ReasoningConfig{TokenBudget: 2048},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Think briefly, then give a short final answer.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Two upstreams are available. A: 200 ms latency, 5% error rate. B: 150 ms latency, 9% error rate. Assuming one retry on failure, briefly reason which has the lower expected latency, then give your answer in a single sentence starting with 'Pick'.")},
				},
			},
		},
		Metadata: map[string]string{"user_id": "anthropic-conversion-fixture-user"},
	},
	ChatEvents: streamThinkingTextChatEvents(),
}
