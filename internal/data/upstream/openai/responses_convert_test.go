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

package openai

import (
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/openai/openai-go/v3/responses"
	openaishared "github.com/openai/openai-go/v3/shared"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
)

func TestConvertRequestToOpenAIResponses(t *testing.T) {
	Convey("Given a native session ID", t, func() {
		repo := &upstream{config: &conf.OpenAIConfig{}, log: slog.Default()}

		Convey("When the session ID is present", func() {
			result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
				Session: "session-1",
				Model:   "gpt-5",
			})

			So(result.PromptCacheKey.Valid(), ShouldBeTrue)
			So(result.PromptCacheKey.Value, ShouldEqual, "session-1")
		})

		Convey("When the session ID is absent", func() {
			result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{Model: "gpt-5"})

			So(result.PromptCacheKey.Valid(), ShouldBeFalse)
		})
	})

	Convey("Given native reasoning configurations", t, func() {
		Convey("When effort is above none", func() {
			repo := &upstream{config: &conf.OpenAIConfig{}, log: slog.Default()}
			result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
				Model: "gpt-5",
				Config: &v1.GenerationConfig{ReasoningConfig: &v1.ReasoningConfig{
					Effort: v1.ReasoningEffort_REASONING_EFFORT_LOW,
				}},
			})
			So(result.Reasoning.Effort, ShouldEqual, openaishared.ReasoningEffortLow)
			So(result.Reasoning.Summary, ShouldEqual, openaishared.ReasoningSummaryAuto)
		})

		Convey("When raw reasoning is enabled", func() {
			repo := &upstream{
				config: &conf.OpenAIConfig{ResponsesUseRawReasoning: true},
				log:    slog.Default(),
			}
			result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
				Model: "gpt-5",
				Config: &v1.GenerationConfig{ReasoningConfig: &v1.ReasoningConfig{
					Effort: v1.ReasoningEffort_REASONING_EFFORT_HIGH,
				}},
			})
			So(result.Reasoning.Summary, ShouldBeEmpty)
		})
	})

	Convey("Given interleaved native content", t, func() {
		repo := &upstream{config: &conf.OpenAIConfig{}, log: slog.Default()}
		result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
			Model: "gpt-5",
			Messages: []*v1.Message{{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{
						Id:      "rs-1",
						Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
						Content: v1.NewTextContent("summary"),
					},
					{
						Id:      "rs-1",
						Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
						Content: &v1.Content_Opaque{Opaque: "encrypted"},
					},
					{
						Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
							Id:   "call-1",
							Name: "lookup",
							Inputs: []*v1.ToolUse_Input{{
								Input: &v1.ToolUse_Input_Text{Text: `{}`},
							}},
						}},
					},
					{
						Phase:   v1.ContentPhase_CONTENT_PHASE_OUTCOME,
						Content: v1.NewTextContent("answer"),
					},
				},
			}},
		})

		Convey("Then independent items preserve content order", func() {
			items := result.Input.OfInputItemList
			So(items, ShouldHaveLength, 3)
			So(items[0].OfReasoning, ShouldNotBeNil)
			So(items[0].OfReasoning.ID, ShouldEqual, "rs-1")
			So(items[0].OfReasoning.Summary, ShouldHaveLength, 1)
			So(items[0].OfReasoning.EncryptedContent.Value, ShouldEqual, "encrypted")
			So(items[1].OfFunctionCall, ShouldNotBeNil)
			So(items[1].OfFunctionCall.CallID, ShouldEqual, "call-1")
			So(items[2].OfMessage, ShouldNotBeNil)
			So(items[2].OfMessage.Phase, ShouldEqual, responses.EasyInputMessagePhaseFinalAnswer)
		})
	})

	Convey("Given native reasoning without item IDs", t, func() {
		repo := &upstream{config: &conf.OpenAIConfig{}, log: slog.Default()}

		Convey("When summary and encrypted contents are consecutive", func() {
			result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
				Model: "gpt-5",
				Messages: []*v1.Message{{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
							Content: v1.NewTextContent("summary"),
						},
						{
							Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
							Content: &v1.Content_Opaque{Opaque: "encrypted"},
						},
					},
				}},
			})

			items := result.Input.OfInputItemList
			So(items, ShouldHaveLength, 1)
			So(items[0].OfReasoning, ShouldNotBeNil)
			So(items[0].OfReasoning.ID, ShouldBeEmpty)
			So(items[0].OfReasoning.Summary, ShouldHaveLength, 1)
			So(items[0].OfReasoning.EncryptedContent.Value, ShouldEqual, "encrypted")
		})

		Convey("When plaintext reasoning has no encrypted content", func() {
			result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
				Model: "gpt-5",
				Messages: []*v1.Message{{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{{
						Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
						Content: v1.NewTextContent("summary"),
					}},
				}},
			})

			items := result.Input.OfInputItemList
			So(items, ShouldHaveLength, 1)
			So(items[0].OfReasoning, ShouldNotBeNil)
			So(items[0].OfReasoning.ID, ShouldBeEmpty)
			So(items[0].OfReasoning.Summary, ShouldHaveLength, 1)
			So(items[0].OfReasoning.EncryptedContent.Valid(), ShouldBeFalse)
		})

		Convey("When a tool call separates two unidentified reasoning segments", func() {
			result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
				Model: "gpt-5",
				Messages: []*v1.Message{{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
							Content: v1.NewTextContent("first"),
						},
						{
							Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Id:   "call-1",
								Name: "lookup",
								Inputs: []*v1.ToolUse_Input{{
									Input: &v1.ToolUse_Input_Text{Text: `{}`},
								}},
							}},
						},
						{
							Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
							Content: v1.NewTextContent("second"),
						},
					},
				}},
			})

			items := result.Input.OfInputItemList
			So(items, ShouldHaveLength, 3)
			So(items[0].OfReasoning, ShouldNotBeNil)
			So(items[0].OfReasoning.Summary[0].Text, ShouldEqual, "first")
			So(items[1].OfFunctionCall, ShouldNotBeNil)
			So(items[2].OfReasoning, ShouldNotBeNil)
			So(items[2].OfReasoning.Summary[0].Text, ShouldEqual, "second")
		})

		Convey("When raw reasoning is enabled", func() {
			repo.config.ResponsesUseRawReasoning = true
			result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
				Model: "gpt-5",
				Messages: []*v1.Message{{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{{
						Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
						Content: v1.NewTextContent("raw"),
					}},
				}},
			})

			items := result.Input.OfInputItemList
			So(items, ShouldHaveLength, 1)
			So(items[0].OfReasoning.ID, ShouldBeEmpty)
			So(items[0].OfReasoning.Summary, ShouldBeEmpty)
			So(items[0].OfReasoning.Content, ShouldHaveLength, 1)
			So(items[0].OfReasoning.Content[0].Text, ShouldEqual, "raw")
		})

	})
}

func TestConvertResponseFromOpenAIResponses(t *testing.T) {
	Convey("Given a Responses response with reasoning, text, and a tool call", t, func() {
		respBody := `{
			"id":"resp-1",
			"model":"gpt-5",
			"status":"completed",
			"output":[
				{"id":"rs-1","type":"reasoning","status":"completed","summary":[{"type":"summary_text","text":"summary"}],"encrypted_content":"encrypted"},
				{"id":"msg-1","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"answer","annotations":[],"logprobs":[]}]},
				{"id":"fc-1","type":"function_call","status":"completed","call_id":"call-1","name":"lookup","arguments":"{}"}
			],
			"usage":{"input_tokens":9,"input_tokens_details":{"cached_tokens":2},"output_tokens":6,"output_tokens_details":{"reasoning_tokens":3},"total_tokens":15}
		}`
		var openAIResp responses.Response
		So(json.Unmarshal([]byte(respBody), &openAIResp), ShouldBeNil)
		repo := &upstream{config: &conf.OpenAIConfig{}, log: slog.Default()}

		result := repo.convertResponseFromOpenAIResponses(&entity.ChatReq{Id: "request-1"}, &openAIResp)

		Convey("Then output items expand to native contents", func() {
			So(result.Id, ShouldEqual, "request-1")
			So(result.Message.Id, ShouldEqual, "resp-1")
			So(result.Status, ShouldEqual, v1.ChatStatus_CHAT_PENDING_TOOL_USE)
			So(result.Message.Contents, ShouldHaveLength, 4)
			So(result.Message.Contents[0].Id, ShouldEqual, "rs-1")
			So(result.Message.Contents[0].GetText().GetText(), ShouldEqual, "summary")
			So(result.Message.Contents[1].Id, ShouldEqual, "rs-1")
			So(result.Message.Contents[1].GetOpaque(), ShouldEqual, "encrypted")
			So(result.Message.Contents[2].Id, ShouldEqual, "msg-1")
			So(result.Message.Contents[2].GetText().GetText(), ShouldEqual, "answer")
			So(result.Message.Contents[3].Id, ShouldEqual, "fc-1")
			So(result.Message.Contents[3].GetToolUse().Id, ShouldEqual, "call-1")
			So(result.Message.Contents[3].GetToolUse().GetTextualInput(), ShouldEqual, "{}")
		})
	})

	Convey("Given a native model text message", t, func() {
		repo := &upstream{config: &conf.OpenAIConfig{}, log: slog.Default()}
		result := repo.convertRequestToOpenAIResponses(&entity.ChatReq{
			Model: "gpt-5",
			Messages: []*v1.Message{{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{{
					Phase:   v1.ContentPhase_CONTENT_PHASE_OUTCOME,
					Content: v1.NewTextContent("previous answer"),
				}},
			}},
		})

		items := result.Input.OfInputItemList
		So(items, ShouldHaveLength, 1)
		So(items[0].OfMessage, ShouldNotBeNil)
		So(items[0].OfMessage.Content.OfString.Valid(), ShouldBeTrue)
		So(items[0].OfMessage.Content.OfString.Value, ShouldEqual, "previous answer")
		So(items[0].OfMessage.Content.OfInputItemContentList, ShouldBeEmpty)
	})
}

func TestConvertStatusFromOpenAIResponses(t *testing.T) {
	Convey("Given Responses terminal states", t, func() {
		So(convertStatusFromOpenAIResponses(
			&responses.Response{Status: responses.ResponseStatusCompleted}),
			ShouldEqual, v1.ChatStatus_CHAT_COMPLETED)
		So(convertStatusFromOpenAIResponses(
			&responses.Response{Status: responses.ResponseStatusFailed}),
			ShouldEqual, v1.ChatStatus_CHAT_FAILED)
		So(convertStatusFromOpenAIResponses(
			&responses.Response{
				Status: responses.ResponseStatusFailed,
				Output: []responses.ResponseOutputItemUnion{{Type: "function_call"}},
			}),
			ShouldEqual, v1.ChatStatus_CHAT_FAILED)
		So(convertStatusFromOpenAIResponses(
			&responses.Response{
				Status:            responses.ResponseStatusIncomplete,
				IncompleteDetails: responses.ResponseIncompleteDetails{Reason: "content_filter"},
			}),
			ShouldEqual, v1.ChatStatus_CHAT_REFUSED)
		So(convertStatusFromOpenAIResponses(
			&responses.Response{Status: responses.ResponseStatusCancelled}),
			ShouldEqual, v1.ChatStatus_CHAT_CANCELLED)

		So(convertStatusFromOpenAIResponses(
			&responses.Response{
				Status: responses.ResponseStatusCompleted,
				Output: []responses.ResponseOutputItemUnion{{Type: "function_call"}},
			}),
			ShouldEqual, v1.ChatStatus_CHAT_PENDING_TOOL_USE)
		So(convertStatusFromOpenAIResponses(
			&responses.Response{
				Status:            responses.ResponseStatusIncomplete,
				IncompleteDetails: responses.ResponseIncompleteDetails{Reason: "max_output_tokens"},
			}),
			ShouldEqual, v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT)
		So(convertStatusFromOpenAIResponses(
			&responses.Response{
				Status:            responses.ResponseStatusIncomplete,
				IncompleteDetails: responses.ResponseIncompleteDetails{Reason: "max_output_tokens"},
				Output:            []responses.ResponseOutputItemUnion{{Type: "function_call"}},
			}),
			ShouldEqual, v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT)
	})
}

func TestConvertStatisticsFromOpenAIResponses(t *testing.T) {
	Convey("Given empty usage", t, func() {
		So(convertStatisticsFromOpenAIResponses(&responses.ResponseUsage{}), ShouldBeNil)
	})

	Convey("Given complete usage", t, func() {
		statistics := convertStatisticsFromOpenAIResponses(&responses.ResponseUsage{
			InputTokens:  12,
			OutputTokens: 7,
			InputTokensDetails: responses.ResponseUsageInputTokensDetails{
				CachedTokens: 3,
			},
			OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{
				ReasoningTokens: 2,
			},
		})
		So(statistics.Usage.InputTokens, ShouldEqual, 12)
		So(statistics.Usage.OutputTokens, ShouldEqual, 7)
		So(statistics.Usage.CachedInputTokens, ShouldEqual, 3)
		So(statistics.Usage.ReasoningTokens, ShouldEqual, 2)
	})
}

func TestConvertStreamEventFromOpenAIResponses(t *testing.T) {
	Convey("Given text deltas before their parts are added", t, func() {
		testCases := []struct {
			name      string
			rawEvent  string
			wantError string
		}{
			{
				name:      "reasoning text",
				rawEvent:  `{"type":"response.reasoning_summary_text.delta","output_index":0,"item_id":"rs-1","summary_index":0,"delta":"think"}`,
				wantError: "responses stream sent reasoning text before the reasoning part",
			},
			{
				name:      "output text",
				rawEvent:  `{"type":"response.output_text.delta","output_index":0,"item_id":"msg-1","content_index":0,"delta":"answer","logprobs":[]}`,
				wantError: "responses stream sent output text before the output text part",
			},
		}

		for _, testCase := range testCases {
			Convey(testCase.name, func() {
				client := &openAIResponseStreamClient{
					req:                           &entity.ChatReq{Id: "request-1"},
					openContentIndexByOutputIndex: make(map[int64]uint32),
				}
				var event responses.ResponseStreamEventUnion
				So(json.Unmarshal([]byte(testCase.rawEvent), &event), ShouldBeNil)

				converted, err := client.convertStreamEventFromOpenAIResponses(event)

				So(converted, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, testCase.wantError)
			})
		}
	})

	Convey("Given one reasoning item followed by text", t, func() {
		client := &openAIResponseStreamClient{
			req:                           &entity.ChatReq{Id: "request-1"},
			openContentIndexByOutputIndex: make(map[int64]uint32),
			outputPhase:                   make(map[int64]v1.ContentPhase),
		}
		rawEvents := []string{
			`{"type":"response.created","response":{"id":"resp-1","model":"gpt-5","status":"in_progress","output":[]}}`,
			`{"type":"response.reasoning_summary_part.added","output_index":0,"item_id":"rs-1","summary_index":0,"part":{"type":"summary_text","text":""}}`,
			`{"type":"response.reasoning_summary_text.delta","output_index":0,"item_id":"rs-1","summary_index":0,"delta":"think"}`,
			`{"type":"response.output_item.done","output_index":0,"item":{"id":"rs-1","type":"reasoning","status":"completed","summary":[{"type":"summary_text","text":"think"}],"encrypted_content":"encrypted"}}`,
			`{"type":"response.content_part.added","output_index":1,"item_id":"msg-1","content_index":0,"part":{"type":"output_text","text":"","annotations":[],"logprobs":[]}}`,
		}

		var nativeEvents []*entity.ChatEvent
		for _, rawEvent := range rawEvents {
			var event responses.ResponseStreamEventUnion
			So(json.Unmarshal([]byte(rawEvent), &event), ShouldBeNil)
			converted, err := client.convertStreamEventFromOpenAIResponses(event)
			So(err, ShouldBeNil)
			nativeEvents = append(nativeEvents, converted...)
		}

		Convey("Then the summary, encrypted snapshot, and text use consecutive native indexes", func() {
			So(nativeEvents, ShouldHaveLength, 6)
			So(nativeEvents[0].GetMessageStart().GetId(), ShouldEqual, "resp-1")
			So(nativeEvents[1].GetContentStart().GetIndex(), ShouldEqual, 0)
			So(nativeEvents[2].GetContentDelta().GetIndex(), ShouldEqual, 0)
			So(nativeEvents[3].GetContentStop().GetIndex(), ShouldEqual, 0)
			So(nativeEvents[4].GetContentSnapshot().GetIndex(), ShouldEqual, 1)
			So(nativeEvents[5].GetContentStart().GetId(), ShouldEqual, "msg-1")
			So(nativeEvents[5].GetContentStart().GetIndex(), ShouldEqual, 2)
		})
	})

	Convey("Given multiple reasoning summary parts in one output item", t, func() {
		client := &openAIResponseStreamClient{
			req:                           &entity.ChatReq{Id: "request-1"},
			openContentIndexByOutputIndex: make(map[int64]uint32),
			outputPhase:                   make(map[int64]v1.ContentPhase),
		}
		rawEvents := []string{
			`{"type":"response.created","response":{"id":"resp-1","model":"gpt-5","status":"in_progress","output":[]}}`,
			`{"type":"response.reasoning_summary_part.added","output_index":0,"item_id":"rs-1","summary_index":0,"part":{"type":"summary_text","text":""}}`,
			`{"type":"response.reasoning_summary_text.delta","output_index":0,"item_id":"rs-1","summary_index":0,"delta":"first"}`,
			`{"type":"response.reasoning_summary_part.done","output_index":0,"item_id":"rs-1","summary_index":0,"part":{"type":"summary_text","text":"first"}}`,
			`{"type":"response.reasoning_summary_part.added","output_index":0,"item_id":"rs-1","summary_index":1,"part":{"type":"summary_text","text":""}}`,
			`{"type":"response.reasoning_summary_text.delta","output_index":0,"item_id":"rs-1","summary_index":1,"delta":"second"}`,
			`{"type":"response.reasoning_summary_part.done","output_index":0,"item_id":"rs-1","summary_index":1,"part":{"type":"summary_text","text":"second"}}`,
			`{"type":"response.output_item.done","output_index":0,"item":{"id":"rs-1","type":"reasoning","status":"completed","summary":[{"type":"summary_text","text":"first"},{"type":"summary_text","text":"second"}]}}`,
		}

		var nativeEvents []*entity.ChatEvent
		for _, rawEvent := range rawEvents {
			var event responses.ResponseStreamEventUnion
			So(json.Unmarshal([]byte(rawEvent), &event), ShouldBeNil)
			converted, err := client.convertStreamEventFromOpenAIResponses(event)
			So(err, ShouldBeNil)
			nativeEvents = append(nativeEvents, converted...)
		}

		So(nativeEvents, ShouldHaveLength, 7)
		So(nativeEvents[0].GetMessageStart().GetId(), ShouldEqual, "resp-1")

		Convey("Then each summary part remains a separate native content block", func() {
			var starts []*v1.ContentStart
			var stops []*v1.ContentStop
			for _, event := range nativeEvents {
				if start := event.GetContentStart(); start != nil {
					starts = append(starts, start)
				}
				if stop := event.GetContentStop(); stop != nil {
					stops = append(stops, stop)
				}
			}
			So(starts, ShouldHaveLength, 2)
			So(stops, ShouldHaveLength, 2)
			So(starts[0].GetId(), ShouldEqual, "rs-1")
			So(starts[1].GetId(), ShouldEqual, "rs-1")
			So(starts[0].GetIndex(), ShouldEqual, 0)
			So(starts[1].GetIndex(), ShouldEqual, 1)
		})
	})

	Convey("Given multiple text parts in one output message", t, func() {
		client := &openAIResponseStreamClient{
			req:                           &entity.ChatReq{Id: "request-1"},
			openContentIndexByOutputIndex: make(map[int64]uint32),
			outputPhase:                   make(map[int64]v1.ContentPhase),
		}
		rawEvents := []string{
			`{"type":"response.created","response":{"id":"resp-1","model":"gpt-5","status":"in_progress","output":[]}}`,
			`{"type":"response.output_item.added","output_index":0,"item":{"id":"msg-1","type":"message","status":"in_progress","role":"assistant","phase":"final_answer","content":[]}}`,
			`{"type":"response.content_part.added","output_index":0,"item_id":"msg-1","content_index":0,"part":{"type":"output_text","text":"","annotations":[],"logprobs":[]}}`,
			`{"type":"response.output_text.delta","output_index":0,"item_id":"msg-1","content_index":0,"delta":"first","logprobs":[]}`,
			`{"type":"response.content_part.done","output_index":0,"item_id":"msg-1","content_index":0,"part":{"type":"output_text","text":"first","annotations":[],"logprobs":[]}}`,
			`{"type":"response.content_part.added","output_index":0,"item_id":"msg-1","content_index":1,"part":{"type":"output_text","text":"","annotations":[],"logprobs":[]}}`,
			`{"type":"response.output_text.delta","output_index":0,"item_id":"msg-1","content_index":1,"delta":"second","logprobs":[]}`,
			`{"type":"response.content_part.done","output_index":0,"item_id":"msg-1","content_index":1,"part":{"type":"output_text","text":"second","annotations":[],"logprobs":[]}}`,
			`{"type":"response.output_item.done","output_index":0,"item":{"id":"msg-1","type":"message","status":"completed","role":"assistant","content":[]}}`,
		}

		var nativeEvents []*entity.ChatEvent
		for _, rawEvent := range rawEvents {
			var event responses.ResponseStreamEventUnion
			So(json.Unmarshal([]byte(rawEvent), &event), ShouldBeNil)
			converted, err := client.convertStreamEventFromOpenAIResponses(event)
			So(err, ShouldBeNil)
			nativeEvents = append(nativeEvents, converted...)
		}

		So(nativeEvents, ShouldHaveLength, 7)
		So(nativeEvents[0].GetMessageStart().GetId(), ShouldEqual, "resp-1")

		Convey("Then each text part remains a separate outcome block", func() {
			var starts []*v1.ContentStart
			var stops []*v1.ContentStop
			for _, event := range nativeEvents {
				if start := event.GetContentStart(); start != nil {
					starts = append(starts, start)
				}
				if stop := event.GetContentStop(); stop != nil {
					stops = append(stops, stop)
				}
			}
			So(starts, ShouldHaveLength, 2)
			So(stops, ShouldHaveLength, 2)
			So(starts[0].GetId(), ShouldEqual, "msg-1")
			So(starts[1].GetId(), ShouldEqual, "msg-1")
			So(starts[0].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_OUTCOME)
			So(starts[1].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_OUTCOME)
		})
	})

	Convey("Given a Responses stream that ends before a terminal event", t, func() {
		client := &openAIResponseStreamClient{
			req:            &entity.ChatReq{Id: "request-1"},
			messageStarted: true,
		}

		events := client.finish()

		Convey("Then the synthetic terminal event reports a failure", func() {
			So(events, ShouldHaveLength, 1)
			So(events[0].GetMessageStop().GetStatus(), ShouldEqual, v1.ChatStatus_CHAT_FAILED)
		})
	})
}
