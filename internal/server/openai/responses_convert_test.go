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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/data/upstream/openai/mock"
)

func TestConvertChatReqFromOpenAIResponses(t *testing.T) {
	Convey("Given an OpenAI prompt cache key", t, func() {
		actual, err := convertChatReqFromOpenAIResponses([]byte(`{
			"model": "gpt-5",
			"prompt_cache_key": "session-1",
			"input": "hello"
		}`))
		So(err, ShouldBeNil)

		Convey("Then it becomes the native session ID", func() {
			So(actual.Session, ShouldEqual, "session-1")
		})
	})

	Convey("Given the Responses API request fixtures", t, func() {
		for _, fixture := range mock.ResponsesFixtures {
			Convey("When converting the "+fixture.Name+" request", func() {
				actual, err := convertChatReqFromOpenAIResponses(fixture.Request)
				So(err, ShouldBeNil)

				expected := proto.Clone(fixture.ChatReq).(*v1.ChatReq)
				expected.Id = ""
				So(proto.Equal(actual, expected), ShouldBeTrue)
			})
		}
	})
}

func TestConvertChatReqFromOpenAIResponsesReplayedOutput(t *testing.T) {
	Convey("Given a history that replays the output items of a previous response", t, func() {
		req, err := convertChatReqFromOpenAIResponses([]byte(`{
			"model": "gpt-5",
			"input": [
				{"type":"message","role":"user","content":[{"type":"input_text","text":"first question"}]},
				{"type":"reasoning","id":"rs-1","summary":[{"type":"summary_text","text":"thought"}],"encrypted_content":"encrypted"},
				{"type":"message","role":"assistant","status":"completed","id":"msg-1","phase":"final_answer","content":[{"type":"output_text","text":"previous answer","annotations":[]}]},
				{"role":"user","content":[{"type":"input_text","text":"second question"}]}
			]
		}`))
		So(err, ShouldBeNil)

		Convey("Then every turn survives, including the assistant output_text", func() {
			So(proto.Equal(req, &v1.ChatReq{
				Model:  "gpt-5",
				Config: &v1.GenerationConfig{},
				Messages: []*v1.Message{
					{
						Role:     v1.Role_USER,
						Contents: []*v1.Content{{Content: v1.NewTextContent("first question")}},
					},
					{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{
								Id:      "rs-1",
								Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
								Content: v1.NewTextContent("thought"),
							},
							{
								Id:      "rs-1",
								Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
								Content: &v1.Content_Opaque{Opaque: "encrypted"},
							},
							{
								Phase:   v1.ContentPhase_CONTENT_PHASE_OUTCOME,
								Content: v1.NewTextContent("previous answer"),
							},
						},
					},
					{
						Role:     v1.Role_USER,
						Contents: []*v1.Content{{Content: v1.NewTextContent("second question")}},
					},
				},
			}), ShouldBeTrue)
		})
	})

	Convey("Given a tool result sent as a list of content parts", t, func() {
		req, err := convertChatReqFromOpenAIResponses([]byte(`{
			"model": "gpt-5",
			"input": [
				{"type":"function_call_output","call_id":"call-1","output":[{"type":"output_text","text":"tool result"}]}
			]
		}`))
		So(err, ShouldBeNil)

		Convey("Then the parts are flattened into the tool result text", func() {
			So(req.Messages, ShouldHaveLength, 1)
			So(req.Messages[0].Contents[0].GetToolResult().Id, ShouldEqual, "call-1")
			So(req.Messages[0].Contents[0].GetToolResult().GetTextualOutput(), ShouldEqual, "tool result")
		})
	})

	Convey("Given items and content parts that are not supported", t, func() {
		req, err := convertChatReqFromOpenAIResponses([]byte(`{
			"model": "gpt-5",
			"input": [
				{"type":"web_search_call","id":"ws-1","status":"completed"},
				{"type":"message","role":"assistant","content":[{"type":"refusal","refusal":"no"},{"type":"output_text","text":"answer"}]}
			]
		}`))
		So(err, ShouldBeNil)

		Convey("Then they are skipped without discarding the rest of the history", func() {
			So(req.Messages, ShouldHaveLength, 1)
			So(req.Messages[0].Role, ShouldEqual, v1.Role_MODEL)
			So(req.Messages[0].Contents[0].GetText().GetText(), ShouldEqual, "answer")
		})
	})

	Convey("Given a string input", t, func() {
		req, err := convertChatReqFromOpenAIResponses([]byte(`{"model":"gpt-5","input":"hello"}`))
		So(err, ShouldBeNil)

		Convey("Then it becomes a single user message", func() {
			So(req.Messages, ShouldHaveLength, 1)
			So(req.Messages[0].Role, ShouldEqual, v1.Role_USER)
			So(req.Messages[0].Contents[0].GetText().GetText(), ShouldEqual, "hello")
		})
	})
}

func TestConvertChatRespToOpenAIResponses(t *testing.T) {
	Convey("Given a native response containing every supported output item", t, func() {
		resp := &v1.ChatResp{
			Id:     "request-1",
			Model:  "gpt-5",
			Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
			Message: &v1.Message{
				Id:   "resp-1",
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{
						Id:      "rs-1",
						Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
						Content: v1.NewTextContent("Reasoning summary"),
					},
					{
						Id:      "rs-1",
						Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
						Content: &v1.Content_Opaque{Opaque: "encrypted"},
					},
					{
						Id:      "msg-1",
						Phase:   v1.ContentPhase_CONTENT_PHASE_OUTCOME,
						Content: v1.NewTextContent("Final answer"),
					},
					{
						Id: "fc-1",
						Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
							Id:   "call-1",
							Name: "lookup",
							Inputs: []*v1.ToolUse_Input{{
								Input: &v1.ToolUse_Input_Text{Text: `{"query":"test"}`},
							}},
						}},
					},
				},
			},
			Statistics: &v1.Statistics{Usage: &v1.Usage{
				InputTokens:       20,
				OutputTokens:      10,
				CachedInputTokens: 5,
				ReasoningTokens:   4,
			}},
		}

		result := convertChatRespToOpenAIResponses(resp)

		Convey("Then the response and output items retain their order and identities", func() {
			So(result.ID, ShouldEqual, "resp-1")
			So(result.Status, ShouldEqual, "completed")
			So(result.Output, ShouldHaveLength, 3)

			reasoning := result.Output[0]
			So(reasoning.ID, ShouldEqual, "rs-1")
			So(reasoning.Type, ShouldEqual, "reasoning")
			So(reasoning.EncryptedContent, ShouldEqual, "encrypted")
			So(reasoning.Summary, ShouldNotBeNil)
			So(reasoning.Summary, ShouldHaveLength, 1)
			So(reasoning.Summary[0].Text, ShouldEqual, "Reasoning summary")

			message := result.Output[1]
			So(message.ID, ShouldEqual, "msg-1")
			So(message.Type, ShouldEqual, "message")
			So(message.Phase, ShouldEqual, "final_answer")
			So(message.Content, ShouldNotBeNil)
			So(message.Content[0].Text, ShouldEqual, "Final answer")

			call := result.Output[2]
			So(call.ID, ShouldEqual, "fc-1")
			So(call.Type, ShouldEqual, "function_call")
			So(call.CallID, ShouldEqual, "call-1")
			So(call.Name, ShouldEqual, "lookup")
			So(call.Arguments, ShouldNotBeNil)
			So(*call.Arguments, ShouldEqual, `{"query":"test"}`)
		})

		Convey("Then usage is mapped", func() {
			So(result.Usage.InputTokens, ShouldEqual, 20)
			So(result.Usage.OutputTokens, ShouldEqual, 10)
			So(result.Usage.TotalTokens, ShouldEqual, 30)
			So(result.Usage.InputTokensDetails.CachedTokens, ShouldEqual, 5)
			So(result.Usage.OutputTokensDetails.ReasoningTokens, ShouldEqual, 4)
		})
	})
}

func TestConvertUsageToOpenAIResponses(t *testing.T) {
	Convey("Given no native usage", t, func() {
		So(convertUsageToOpenAIResponses(nil), ShouldBeNil)
	})

	Convey("Given complete native usage", t, func() {
		result := convertUsageToOpenAIResponses(&v1.Usage{
			InputTokens:       12,
			OutputTokens:      7,
			CachedInputTokens: 3,
			ReasoningTokens:   2,
		})
		So(result.InputTokens, ShouldEqual, 12)
		So(result.OutputTokens, ShouldEqual, 7)
		So(result.TotalTokens, ShouldEqual, 19)
		So(result.InputTokensDetails.CachedTokens, ShouldEqual, 3)
		So(result.OutputTokensDetails.ReasoningTokens, ShouldEqual, 2)
	})
}

func TestConvertContentsToOpenAIResponses(t *testing.T) {
	Convey("Given native contents with upstream item ids", t, func() {
		output := convertContentsToOpenAIResponses("resp-1", []*v1.Content{
			{Id: "msg-1", Content: v1.NewTextContent("first")},
			{Id: "msg-1", Content: v1.NewTextContent("second")},
			{Id: "msg-2", Content: v1.NewTextContent("third")},
			{
				Id: "fc-1",
				Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
					Id:   "call-1",
					Name: "lookup",
				}},
			},
		})

		Convey("Then parts retain their item boundaries and identities", func() {
			So(output, ShouldHaveLength, 3)
			So(output[0].ID, ShouldEqual, "msg-1")
			So(output[0].Content, ShouldHaveLength, 2)
			So(output[1].ID, ShouldEqual, "msg-2")
			So(output[1].Content, ShouldHaveLength, 1)
			So(output[2].ID, ShouldEqual, "fc-1")
			So(output[2].CallID, ShouldEqual, "call-1")
		})
	})

	Convey("Given native message and tool contents without item ids", t, func() {
		output := convertContentsToOpenAIResponses("resp-1", []*v1.Content{
			{Content: v1.NewTextContent("answer")},
			{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{Id: "call-1", Name: "lookup"}}},
		})

		Convey("Then compatible item ids are synthesized", func() {
			So(output, ShouldHaveLength, 2)
			So(output[0].ID, ShouldEqual, "msg_resp-1_0")
			So(output[1].ID, ShouldEqual, "fc_resp-1_1")
		})
	})

	Convey("Given reasoning contents that carry no item id", t, func() {
		Convey("When a summary is followed by its encrypted snapshot", func() {
			output := convertContentsToOpenAIResponses("resp-1", []*v1.Content{
				{
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: v1.NewTextContent("thought"),
				},
				{
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{Opaque: "encrypted"},
				},
			})

			Convey("Then both land in a single reasoning item", func() {
				So(output, ShouldHaveLength, 1)
				So(output[0].Type, ShouldEqual, "reasoning")
				So(output[0].EncryptedContent, ShouldEqual, "encrypted")
				So(output[0].Summary, ShouldHaveLength, 1)
			})
		})

		Convey("When an answer separates two reasoning runs", func() {
			output := convertContentsToOpenAIResponses("resp-1", []*v1.Content{
				{
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: v1.NewTextContent("thought"),
				},
				{Content: v1.NewTextContent("answer")},
				{
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{Opaque: "encrypted"},
				},
			})

			Convey("Then the runs stay in distinct reasoning items", func() {
				So(output, ShouldHaveLength, 3)
				So(output[0].Type, ShouldEqual, "reasoning")
				So(output[1].Type, ShouldEqual, "message")
				So(output[2].Type, ShouldEqual, "reasoning")
				So(output[0].ID, ShouldNotEqual, output[2].ID)
			})
		})
	})
}

func TestConvertStatusToOpenAIResponses(t *testing.T) {
	Convey("Given native terminal statuses", t, func() {
		status, incomplete, responseError := convertStatusToOpenAIResponses(v1.ChatStatus_CHAT_IN_PROGRESS)
		So(status, ShouldEqual, "failed")
		So(incomplete, ShouldBeNil)
		So(responseError, ShouldNotBeNil)

		status, incomplete, responseError = convertStatusToOpenAIResponses(v1.ChatStatus_CHAT_COMPLETED)
		So(status, ShouldEqual, "completed")
		So(incomplete, ShouldBeNil)
		So(responseError, ShouldBeNil)

		status, incomplete, responseError = convertStatusToOpenAIResponses(v1.ChatStatus_CHAT_FAILED)
		So(status, ShouldEqual, "failed")
		So(incomplete, ShouldBeNil)
		So(responseError, ShouldNotBeNil)

		status, incomplete, responseError = convertStatusToOpenAIResponses(v1.ChatStatus_CHAT_REFUSED)
		So(status, ShouldEqual, "incomplete")
		So(incomplete.Reason, ShouldEqual, "content_filter")
		So(responseError, ShouldBeNil)

		status, incomplete, responseError = convertStatusToOpenAIResponses(v1.ChatStatus_CHAT_CANCELLED)
		So(status, ShouldEqual, "cancelled")
		So(incomplete, ShouldBeNil)
		So(responseError, ShouldNotBeNil)

		status, incomplete, responseError = convertStatusToOpenAIResponses(v1.ChatStatus_CHAT_PENDING_TOOL_USE)
		So(status, ShouldEqual, "completed")
		So(incomplete, ShouldBeNil)
		So(responseError, ShouldBeNil)

		status, incomplete, responseError = convertStatusToOpenAIResponses(v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT)
		So(status, ShouldEqual, "incomplete")
		So(incomplete.Reason, ShouldEqual, "max_output_tokens")
		So(responseError, ShouldBeNil)
	})
}

func TestSyntheticResponsesItemID(t *testing.T) {
	Convey("Given the same response and output index", t, func() {
		first := syntheticResponsesItemID("msg", "resp-1", 2)
		second := syntheticResponsesItemID("msg", "resp-1", 2)
		So(first, ShouldEqual, "msg_resp-1_2")
		So(second, ShouldEqual, first)
	})
}
