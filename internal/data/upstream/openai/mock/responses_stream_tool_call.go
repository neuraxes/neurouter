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

//go:embed responses_stream_tool_call_request.json
var responsesStreamToolCallRequest []byte

//go:embed responses_stream_tool_call_response.txt
var responsesStreamToolCallResponse []byte

func responsesStreamToolCallChatEvents() []*v1.ChatEvent {
	id := eventBuilder("responses_stream_tool_call")
	return []*v1.ChatEvent{
		id.of(v1.NewMessageStartEvent("gen-1785502380-2zHE7hADYroJvmN6sDQs", "openai/gpt-5-mini")),
		id.of(v1.NewContentSnapshotEvent(
			&v1.Content{
				Id:    "rs_0b770f1e999f2b3f016a6c9aad7b4c81a1a0377ba2bb1ead6e",
				Index: new(uint32(0)),
				Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
				Content: &v1.Content_Opaque{
					Opaque: "gAAAAABqbJqvEhbDwfUq9PzYA-SncpQEbM__I61f41hKOaFEkQXLuuGgYjthIg_8vxTzE24P7BVWLP0Y7yggTWS_UjppA9KRcK69OnDGM4JqFRiujpTZ9-0vFyJ-h_XpN1yDENysfFkafaXSKN1Ryc2q2MWa-GD2FnZOZq3LGV3en3QoOhfiNIzcvTiClDG2MOMZa6tI9iNAfzrA78tYtZ3OiNFd8SEd1VqnRR_DxarlOtXD5vKMsSRM508ibD4TapKwCh4R1MdMC0NvDGbxJKvo_WUSXuNlHnm1nVHLDalukA_tDWEWvc9cSX-PEfOUI7R7KdhDtw2pM_Inm0m8xioZrY3bOijY7GhNJZLJit-0KgJfefaXEMr9iP-6PK67QlZYltggl_RGcwaDs62zyox8zriIH88u151qJrjOQwtd2GryQmWNmr6FpMd_5AkcIl2NHxUMNiq6y_66nQPdmoBrPZpYjcBOwtUVwWa9idi7gTtWlmFlt7WExuqR5q_GjmOwcNHv_hivac_lVGjo_aUD5vu1w2hW5NBxpC0H7FPizk_nEbfThqwb8cR6A4zsRXQnZi0OSurM_YubRk-L-su18_3_noF9Ow1E9quVqVDQDKcKDA4zJfJFX319-qiR76KeGJmfHb3fLnlNdp4Y7Yv3MhFhAf5vW5j-7gjvSyy4SxuXuyza7h0_j21Jwq99o9cuqYoZe0uvTtk3MK_M2AGH-5M9g7BT8MOqoPCquphq6Kh5S3dXTALxcF0-BWtfGN1yaNt82U6_qM_vay6WI-q20wpz4KmzhIr1WSDUslrpGp8H7O6ll9PCldANWZMP2qwuDuMJeWcyagi5o0TnBByfKOn2LrNDxnwsqMZDJM_eR21vq6T604SOpOKo_j0iwnX-y1fX1dzA_tBLemUDGy3Cv4ZITdBQKc04_7UjugHpIWUEJOoKiSCEex7ve7oa0UasZa2jCLtC17xH52U9PRlTy9-_ucKyEQ0A0jjK8eQwVLntImSfxQy1AP3G05LtnUY0rMTP5Qxt1QqJym1WbK5-iuuPt0h9vk_AIKWjwN-b04lJ_Wv6nkDk3kII0bA7CyuvITWOUfZPbO95vU0MXU6bUKaZ0puQ2Mi1r2GgOPScrAA1lrY2RztgXwk-Wsl0HKBa8X9pnQSq0DIpRGgFdgZHJD2yjW7xy_UV7fAmUeOPg0THyKQg9QYQCsj1gStoMq-8MSoahpSVUF0XLzYk2HobT3EklJ0Viqx9FcN5UUpXGe_mSI6NXG2Q3NlR0WSY0VkGa3G6EBFMLcXnls6wkAD0ts84ym8wYd5PDTiENGgWsDQDnVdhwCSgAiSJkqOqVhpCDFxJckjRGhgnaJCGYkp2SOlhrPVt4j_11ItfyQ4Re59PFoNMipu6ZyCQpmneH5IVw3INNywqPISZnG6NEnlQ2mMlP2LhyA==",
				},
			},
		)),
		id.of(v1.NewIdentifiedContentStartTextEvent("msg_tmp_d0bmohz8m5h", 1, v1.ContentPhase_CONTENT_PHASE_NORMAL)),
		id.of(v1.NewContentDeltaTextEvent(1, "Preparing")),
		id.of(v1.NewContentDeltaTextEvent(1, " the")),
		id.of(v1.NewContentDeltaTextEvent(1, " weather")),
		id.of(v1.NewContentDeltaTextEvent(1, " lookup")),
		id.of(v1.NewContentDeltaTextEvent(1, ".")),
		id.of(v1.NewContentStopEvent(1)),
		id.of(v1.NewIdentifiedContentStartToolUseEvent("fc_tmp_eunwh2jc6lg", 2, "call_bAClRBNvIDQPvwzKp9A1Xv2c", "get_weather")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "{\"")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "city")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "\":\"")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "Shanghai")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "\",\"")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "date")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "\":\"")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "202")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "5")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "-")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "11")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "-")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "10")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "\",\"")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "units")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "\":\"")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "metric")),
		id.of(v1.NewContentDeltaToolInputTextEvent(2, "\"}")),
		id.of(v1.NewContentStopEvent(2)),
		id.withUsage(
			v1.NewMessageStopEvent(v1.ChatStatus_CHAT_STATUS_PENDING_TOOL_USE),
			&v1.Usage{
				InputTokens:     139,
				OutputTokens:    126,
				ReasoningTokens: 64,
			},
		),
	}
}

// ResponsesStreamToolCall covers streamed text followed by function-call
// argument deltas while preserving encrypted reasoning state.
var ResponsesStreamToolCall = &Fixture{
	Name:     "responses_stream_tool_call",
	Request:  responsesStreamToolCallRequest,
	Response: responsesStreamToolCallResponse,
	Stream:   true,
	ChatRequest: &v1.ChatRequest{
		Id:    "responses_stream_tool_call",
		Model: "openai/gpt-5-mini",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(2048)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_LOW},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_ROLE_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. First emit exactly one short sentence, then call the weather tool exactly once.")},
				},
			},
			{
				Role: v1.Role_ROLE_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("First say exactly: Preparing the weather lookup. Then call get_weather for Shanghai on 2025-11-10 using metric units. Do not provide the weather from memory.")},
				},
			},
		},
		Tools: []*v1.Tool{getWeatherTool()},
	},
	ChatEvents: responsesStreamToolCallChatEvents(),
}
