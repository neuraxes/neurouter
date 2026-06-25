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

//go:embed messages_non_stream_thinking_request.json
var nonStreamThinkingRequest []byte

//go:embed messages_non_stream_thinking_response.json
var nonStreamThinkingResponse []byte

const nonStreamThinkingText = "Let me calculate the expected total latency for each upstream, accounting for retries.\n\nFor each upstream, a request can:\n1. Succeed on first try: probability (1 - error_rate), latency = wait\n2. Fail on first try, succeed on retry: probability error_rate * (1 - error_rate), latency = 2 * wait\n3. Fail on both tries: probability error_rate^2, latency = 2 * wait (we still waited for both attempts)\nSo the expected latency simplifies to wait * (1 + error_rate). Let me recalculate this more carefully to make sure I'm getting it right.\n\nNow I'm comparing the three services with this formula: OpenAI comes to about 2016 ms, Anthropic to 121.2 ms, and Gemini to 416 ms. Anthropic is clearly the winner here."

const nonStreamThinkingSignature = "EsgLCmUIDxgCKkDNaCJqQ4igJANJ6Kt5qVZoclT/pJb5ttKtmDWwxitmrOF8dXaQgHi3Fgu0x550TjlK9Kh/jpOVdG3HmQNT/SrSMhFjbGF1ZGUtc29ubmV0LTQtNjgAQgh0aGlua2luZxIMNRdEc2aptoqXALYZGgxiRIzXcsEq7PBCpwAiMO7RdBaJPEtm3mU63b5FGY6e0vJM5vzn6/lyZ42fA/CJPSC7sPw/fr4X0f04z3dQJSqQCtVDzQms6h52wBJ2pprypxMffYt+PdXwXx0QZIUHDxo6SFeGkjOJCS4QsFcGV13A86MEi1AubsAs/NTqKsE2CR1LVFbqGOw2SO61c6tKYo33Tgd/1TELZ5sjbJEYYDab93Yr/Ww1DIyATIzeR+Y5Ns9UXjiwfMKs3wD/eysLc6ac8HYR9B6M+0zZ2ISZWk5APYhIFvB/gr65C1RO0QnMEJYHZWgYbXQuuTbZCaDkdInyY/zVLuN0evVg38b7jyNQFT68rWfmsguGGXUHGfyus180zj5MF3mMe/IkBH+kL4lVyr+9wWlAL2hSfkp6mMUcOG9Z7d/00spdouCbpXnXvmCVFNvx8B2PWlL4x4IpYtpmF4vRUE4BczYE0Qex9t5DYXh6OUCzn/Zmo0IEwrCOa/Mn5fbP8vljEo4H9/fy9MUmmbENbEhnn+PDY9h6+nVEaxz6iOHfsDCbR7s3AbmOVSb5MDxaUypP9MnP2I/iSHpdfpi6mnUGQOGLCH/jPM4X8yNdT1Pwn7cXjkzPyrhSBY46vCx68hCarHaVxmiIgdbP2SYDeIsuE4KhAHpF5/yzgM8myT+Zj5gGBDJLqVLohtZJ3tQVx+2bnzkTPVFiMK4QI9Kwv4b5TcUqCqMKaGuziyEbFycD9zzIn1QRzjNLuPwP8F/wBzkLZhu0a5kvtJydwsy5Nz11mJ4xUvqKYtl4h6YqnNk0EPN97mrrOPfBvpAywnyVsax+Xr67qlFr0pDpGql4CqIBFWfqOfUppL/mgqaME8JWbnJ0HlK5/PXTn3lGcDiGM9PVliNYnlhHGyVownIbhYfiV+rWRWuWt3GTJ2YCWM3iG/KPQYRpnZIKCCmXKObOcVGywxrOrmFPzuxp6SyuhRATZ0Ttn5QDTADXHyINAsj4kGy6pbVA0EzOviz97r2PSlCbNxofkO3NRjed6BmIjnmjUH28H74MsAxgU9/eLt98S7TVxEuvUJ1wMcHnOkdZwIQ+ROotGIveSI0jz0kZLZ8Bg82weMruJ1E81zYkAp2k//CYy0+/s0KAcQXnMeRk2Q2toWDCHkkAeuRZhUBzNcQBbEozF5i9wB0Kn+zwSsP6GuxaQc8aQ+aFd4nkggbfXhiF/AHqIvGvneXpo7ULpuoRqQcu5Xre/mk2L7O9nirkA+HCYwn3usdFawhOI19wtdjQxGvRXS31hutdKF7iuA0niq7lzABp52bp/IpUhpJ0kxeezSQaKOMdvTbnhey1JnVW/hnkJM0wjwN7xPxZRjpI24pCDV70TAgHy/t6xS05q7DhyoonYkjto5zDWiurq/VwduEC5ka1g0KiGcEp/TuODQsmqBHG4M4MKhyu0wEE9p9OzUn8t1exzgUSvHmcPOgZm98sUPeDsj4Y/c8/D8/ch6x/Y2Kqlb7ehtI1AsKSnM1IagMb8xEYK9ePEJOTm4FVY15f395N4K7u+vnDP+Hi1YKLMUlCnyaIZDtnmZ0cBT8MQwul4YifHggPzAaHoEqcHBGhnKyFVKw5SIVCkyUiJYF6AsJyKLnRi72pOvD85T98t31uQJKywRPB3O228rE6fHkc8vQs/xzORTZBBfulPD40Uu/+O5jV347pi9xkx4s9nAvRBMU6k+q310OGXZjU7F+5UKvSs/rwdDwQe0ZhKH3NZbGBKAGpX7f/vbLV5BNY6SsFpo9B0oCmsYwLgiBeckl9lH9mydlG4+MUlvi9oCQFepXz0X7yaRgB"

const nonStreamThinkingOutput = "## Step-by-Step Analysis\n\n### Setting Up the Model\n\nFor any upstream, a request always incurs at least one wait. If the first attempt fails (with probability = error_rate), a retry is made, incurring a second wait of the same duration.\n\n**Expected total latency = wait × (1 + error_rate)**\n\nThis follows from:\n- We *always* pay `wait` for attempt #1\n- With probability `error_rate`, we also pay `wait` for attempt #2\n- E[latency] = wait + (error_rate × wait) = wait × (1 + error_rate)\n\n*(Note: whether the retry itself succeeds or fails, the time cost is the same — we've already waited.)*\n\n---\n\n### Calculating Each Upstream\n\n| Upstream | Wait (ms) | Error Rate | Formula | Expected Latency |\n|----------|-----------|------------|---------|-----------------|\n| OpenAI | 1800 | 0.12 | 1800 × 1.12 | **2016.0 ms** |\n| Anthropic | 120 | 0.01 | 120 × 1.01 | **121.2 ms** |\n| Gemini | 400 | 0.04 | 400 × 1.04 | **416.0 ms** |\n\n---\n\n### Trade-off Commentary\n\n- **OpenAI** has the highest base latency by far (1800 ms), and the high error rate (12%) compounds this — worst by a large margin.\n- **Gemini** is middle-of-the-road: moderate latency and low error rate, but still ~3.4× slower than Anthropic.\n- **Anthropic** wins decisively. Its base latency (120 ms) is so low that even if the error rate were much higher, it would still outperform the others. Its 1% error rate barely adds anything (only +1.2 ms expected overhead).\n\n---\n\n**Recommendation:** Route requests to Anthropic, as its expected total latency of ~121.2 ms is roughly 3.4× lower than Gemini and nearly 17× lower than OpenAI when accounting for the retry penalty."

const nonStreamThinkingRedacted = "openrouter.reasoning:eyJzaWduYXR1cmUiOiJFc2dMQ21VSUR4Z0NLa0ROYUNKcVE0aWdKQU5KNkt0NXFWWm9jbFQvcEpiNXR0S3RtRFd3eGl0bXJPRjhkWGFRZ0hpM0ZndTB4NTUwVGpsSzlLaC9qcE9WZEczSG1RTlQvU3JTTWhGamJHRjFaR1V0YzI5dWJtVjBMVFF0TmpnQVFnaDBhR2x1YTJsdVp4SU1OUmRFYzJhcHRvcVhBTFlaR2d4aVJJelhjc0VxN1BCQ3B3QWlNTzdSZEJhSlBFdG0zbVU2M2I1RkdZNmUwdkpNNXZ6bjYvbHlaNDJmQS9DSlBTQzdzUHcvZnI0WDBmMDR6M2RRSlNxUUN0VkR6UW1zNmg1MndCSjJwcHJ5cHhNZmZZdCtQZFh3WHgwUVpJVUhEeG82U0ZlR2tqT0pDUzRRc0ZjR1YxM0E4Nk1FaTFBdWJzQXMvTlRxS3NFMkNSMUxWRmJxR093MlNPNjFjNnRLWW8zM1RnZC8xVEVMWjVzamJKRVlZRGFiOTNZci9XdzFESXlBVEl6ZVIrWTVOczlVWGppd2ZNS3Mzd0QvZXlzTGM2YWM4SFlSOUI2TSsweloySVNaV2s1QVBZaElGdkIvZ3I2NUMxUk8wUW5NRUpZSFpXZ1liWFF1dVRiWkNhRGtkSW55WS96Vkx1TjBldlZnMzhiN2p5TlFGVDY4cldmbXNndUdHWFVIR2Z5dXMxODB6ajVNRjNtTWUvSWtCSCtrTDRsVnlyKzl3V2xBTDJoU2ZrcDZtTVVjT0c5WjdkLzAwc3Bkb3VDYnBYblh2bUNWRk52eDhCMlBXbEw0eDRJcFl0cG1GNHZSVUU0QmN6WUUwUWV4OXQ1RFlYaDZPVUN6bi9abW8wSUV3ckNPYS9NbjVmYlA4dmxqRW80SDkvZnk5TVVtbWJFTmJFaG5uK1BEWTloNituVkVheHo2aU9IZnNEQ2JSN3MzQWJtT1ZTYjVNRHhhVXlwUDlNblAySS9pU0hwZGZwaTZtblVHUU9HTENIL2pQTTRYOHlOZFQxUHduN2NYamt6UHlyaFNCWTQ2dkN4NjhoQ2FySGFWeG1pSWdkYlAyU1lEZUlzdUU0S2hBSHBGNS95emdNOG15VCtaajVnR0JESkxxVkxvaHRaSjN0UVZ4KzJibnprVFBWRmlNSzRRSTlLd3Y0YjVUY1VxQ3FNS2FHdXppeUViRnljRDl6ekluMVFSempOTHVQd1A4Ri93QnprTFpodTBhNWt2dEp5ZHdzeTVOejExbUo0eFV2cUtZdGw0aDZZcW5OazBFUE45N21yck9QZkJ2cEF5d255VnNheCtYcjY3cWxGcjBwRHBHcWw0Q3FJQkZXZnFPZlVwcEwvbWdxYU1FOEpXYm5KMEhsSzUvUFhUbjNsR2NEaUdNOVBWbGlOWW5saEhHeVZvd25JYmhZZmlWK3JXUld1V3QzR1RKMllDV00zaUcvS1BRWVJwblpJS0NDbVhLT2JPY1ZHeXd4ck9ybUZQenV4cDZTeXVoUkFUWjBUdG41UURUQURYSHlJTkFzajRrR3k2cGJWQTBFek92aXo5N3IyUFNsQ2JOeG9ma08zTlJqZWQ2Qm1Jam5talVIMjhINzRNc0F4Z1U5L2VMdDk4UzdUVnhFdXZVSjF3TWNIbk9rZFp3SVErUk9vdEdJdmVTSTBqejBrWkxaOEJnODJ3ZU1ydUoxRTgxellrQXAyay8vQ1l5MCsvczBLQWNRWG5NZVJrMlEydG9XRENIa2tBZXVSWmhVQnpOY1FCYkVvekY1aTl3QjBLbit6d1NzUDZHdXhhUWM4YVErYUZkNG5rZ2diZlhoaUYvQUhxSXZHdm5lWHBvN1VMcHVvUnFRY3U1WHJlL21rMkw3TzluaXJrQStIQ1l3bjN1c2RGYXdoT0kxOXd0ZGpReEd2UlhTMzFodXRkS0Y3aXVBMG5pcTdsekFCcDUyYnAvSXBVaHBKMGt4ZWV6U1FhS09NZHZUYm5oZXkxSm5WVy9obmtKTTB3andON3hQeFpSanBJMjRwQ0RWNzBUQWdIeS90NnhTMDVxN0RoeW9vbllranRvNXpEV2l1cnEvVndkdUVDNWthMWcwS2lHY0VwL1R1T0RRc21xQkhHNE00TUtoeXUwd0VFOXA5T3pVbjh0MWV4emdVU3ZIbWNQT2dabTk4c1VQZURzajRZL2M4L0Q4L2NoNngvWTJLcWxiN2VodEkxQXNLU25NMUlhZ01iOHhFWUs5ZVBFSk9UbTRGVlkxNWYzOTVONEs3dSt2bkRQK0hpMVlLTE1VbENueWFJWkR0bm1aMGNCVDhNUXd1bDRZaWZIZ2dQekFhSG9FcWNIQkdobkt5RlZLdzVTSVZDa3lVaUpZRjZBc0p5S0xuUmk3MnBPdkQ4NVQ5OHQzMXVRSkt5d1JQQjNPMjI4ckU2ZkhrYzh2UXMveHpPUlRaQkJmdWxQRDQwVXUvK081alYzNDdwaTl4a3g0czluQXZSQk1VNmsrcTMxME9HWFpqVTdGKzVVS3ZTcy9yd2REd1FlMFpoS0gzTlpiR0JLQUdwWDdmL3ZiTFY1Qk5ZNlNzRnBvOUIwb0Ntc1l3TGdpQmVja2w5bEg5bXlkbEc0K01VbHZpOW9DUUZlcFh6MFg3eWFSZ0IiLCJ0ZXh0IjoiTGV0IG1lIGNhbGN1bGF0ZSB0aGUgZXhwZWN0ZWQgdG90YWwgbGF0ZW5jeSBmb3IgZWFjaCB1cHN0cmVhbSwgYWNjb3VudGluZyBmb3IgcmV0cmllcy5cblxuRm9yIGVhY2ggdXBzdHJlYW0sIGEgcmVxdWVzdCBjYW46XG4xLiBTdWNjZWVkIG9uIGZpcnN0IHRyeTogcHJvYmFiaWxpdHkgKDEgLSBlcnJvcl9yYXRlKSwgbGF0ZW5jeSA9IHdhaXRcbjIuIEZhaWwgb24gZmlyc3QgdHJ5LCBzdWNjZWVkIG9uIHJldHJ5OiBwcm9iYWJpbGl0eSBlcnJvcl9yYXRlICogKDEgLSBlcnJvcl9yYXRlKSwgbGF0ZW5jeSA9IDIgKiB3YWl0XG4zLiBGYWlsIG9uIGJvdGggdHJpZXM6IHByb2JhYmlsaXR5IGVycm9yX3JhdGVeMiwgbGF0ZW5jeSA9IDIgKiB3YWl0ICh3ZSBzdGlsbCB3YWl0ZWQgZm9yIGJvdGggYXR0ZW1wdHMpXG5TbyB0aGUgZXhwZWN0ZWQgbGF0ZW5jeSBzaW1wbGlmaWVzIHRvIHdhaXQgKiAoMSArIGVycm9yX3JhdGUpLiBMZXQgbWUgcmVjYWxjdWxhdGUgdGhpcyBtb3JlIGNhcmVmdWxseSB0byBtYWtlIHN1cmUgSSdtIGdldHRpbmcgaXQgcmlnaHQuXG5cbk5vdyBJJ20gY29tcGFyaW5nIHRoZSB0aHJlZSBzZXJ2aWNlcyB3aXRoIHRoaXMgZm9ybXVsYTogT3BlbkFJIGNvbWVzIHRvIGFib3V0IDIwMTYgbXMsIEFudGhyb3BpYyB0byAxMjEuMiBtcywgYW5kIEdlbWluaSB0byA0MTYgbXMuIEFudGhyb3BpYyBpcyBjbGVhcmx5IHRoZSB3aW5uZXIgaGVyZS4iLCJ0eXBlIjoicmVhc29uaW5nLnRleHQifQ=="

// NonStreamThinking covers a request that enables extended thinking with a token
// budget. The response carries a thinking block (text plus signature), the final
// text answer, and a redacted_thinking block, with stop reason end_turn.
var NonStreamThinking = &Fixture{
	Name:     "non_stream_thinking",
	Request:  nonStreamThinkingRequest,
	Response: nonStreamThinkingResponse,
	ChatReq: &v1.ChatReq{
		Id:    "non_stream_thinking",
		Model: "anthropic/claude-sonnet-4.6",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(4096)),
			ReasoningConfig: &v1.ReasoningConfig{TokenBudget: 2048},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Reason through the trade-offs before giving a final recommendation.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("A router can send a request to one of three upstreams. OpenAI: wait 1800 ms, error rate 0.12. Anthropic: wait 120 ms, error rate 0.01. Gemini: wait 400 ms, error rate 0.04. A failed request is retried once on the same upstream. Reason step by step about which upstream minimizes expected total latency including the retry, then end with a single sentence starting with 'Recommendation:'.")},
				},
			},
		},
		Metadata: map[string]string{"user_id": "anthropic-conversion-fixture-user"},
	},
	ChatResp: &v1.ChatResp{
		Id:     "non_stream_thinking",
		Model:  "anthropic/claude-4.6-sonnet-20260217",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "gen-1782639369-7ySC12UT9Uvu9FaKjjLU",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Phase:     v1.ContentPhase_CONTENT_PHASE_REASONING,
					Signature: nonStreamThinkingSignature,
					Content:   v1.NewTextContent(nonStreamThinkingText),
				},
				{Content: v1.NewTextContent(nonStreamThinkingOutput)},
				{
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{Opaque: nonStreamThinkingRedacted},
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 168, OutputTokens: 1013},
		},
	},
}
