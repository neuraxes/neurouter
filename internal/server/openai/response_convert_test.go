// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openai

import (
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestConvertEffortFromOpenAIResponse(t *testing.T) {
	Convey("Given reasoning effort values", t, func() {
		So(convertEffortFromOpenAIResponse(shared.ReasoningEffortNone), ShouldEqual, v1.ReasoningEffort_REASONING_EFFORT_NONE)
		So(convertEffortFromOpenAIResponse(shared.ReasoningEffortMinimal), ShouldEqual, v1.ReasoningEffort_REASONING_EFFORT_MINIMAL)
		So(convertEffortFromOpenAIResponse(shared.ReasoningEffortLow), ShouldEqual, v1.ReasoningEffort_REASONING_EFFORT_LOW)
		So(convertEffortFromOpenAIResponse(shared.ReasoningEffortMedium), ShouldEqual, v1.ReasoningEffort_REASONING_EFFORT_MEDIUM)
		So(convertEffortFromOpenAIResponse(shared.ReasoningEffortHigh), ShouldEqual, v1.ReasoningEffort_REASONING_EFFORT_HIGH)
		So(convertEffortFromOpenAIResponse(shared.ReasoningEffortXhigh), ShouldEqual, v1.ReasoningEffort_REASONING_EFFORT_MAX)
		So(convertEffortFromOpenAIResponse("unknown"), ShouldEqual, v1.ReasoningEffort_REASONING_EFFORT_UNSPECIFIED)
	})
}

func TestConvertConfigFromResponseReq(t *testing.T) {
	Convey("Given a ResponseNewParams to extract config from", t, func() {

		Convey("When max_output_tokens is set", func() {
			req := &responses.ResponseNewParams{
				MaxOutputTokens: openai.Opt[int64](4096),
			}
			config := convertConfigFromResponseReq(req)

			Convey("Then MaxTokens should be set", func() {
				So(config.MaxTokens, ShouldNotBeNil)
				So(*config.MaxTokens, ShouldEqual, 4096)
			})
		})

		Convey("When temperature and top_p are set", func() {
			req := &responses.ResponseNewParams{
				Temperature: openai.Opt(0.7),
				TopP:        openai.Opt(0.9),
			}
			config := convertConfigFromResponseReq(req)

			Convey("Then Temperature and TopP should be set", func() {
				So(config.Temperature, ShouldNotBeNil)
				So(*config.Temperature, ShouldAlmostEqual, 0.7, 0.01)
				So(config.TopP, ShouldNotBeNil)
				So(*config.TopP, ShouldAlmostEqual, 0.9, 0.01)
			})
		})

		Convey("When reasoning effort is set", func() {
			req := &responses.ResponseNewParams{
				Reasoning: shared.ReasoningParam{
					Effort: shared.ReasoningEffortHigh,
				},
			}
			config := convertConfigFromResponseReq(req)

			Convey("Then ReasoningConfig should be populated", func() {
				So(config.ReasoningConfig, ShouldNotBeNil)
				So(config.ReasoningConfig.Effort, ShouldEqual, v1.ReasoningEffort_REASONING_EFFORT_HIGH)
			})
		})

		Convey("When text format is json_object", func() {
			req := &responses.ResponseNewParams{
				Text: responses.ResponseTextConfigParam{
					Format: responses.ResponseFormatTextConfigUnionParam{
						OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
					},
				},
			}
			config := convertConfigFromResponseReq(req)

			Convey("Then Grammar should be json_object preset", func() {
				preset, ok := config.Grammar.(*v1.GenerationConfig_PresetGrammar)
				So(ok, ShouldBeTrue)
				So(preset.PresetGrammar, ShouldEqual, "json_object")
			})
		})

		Convey("When text format is json_schema", func() {
			req := &responses.ResponseNewParams{
				Text: responses.ResponseTextConfigParam{
					Format: responses.ResponseFormatTextConfigUnionParam{
						OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
							Name: "person",
							Schema: map[string]any{
								"type": "object",
								"properties": map[string]any{
									"name": map[string]any{"type": "string"},
								},
							},
						},
					},
				},
			}
			config := convertConfigFromResponseReq(req)

			Convey("Then Grammar should be json_schema", func() {
				jsonSchema, ok := config.Grammar.(*v1.GenerationConfig_JsonSchema)
				So(ok, ShouldBeTrue)
				So(jsonSchema.JsonSchema, ShouldNotBeEmpty)
			})
		})

		Convey("When no config fields are set", func() {
			req := &responses.ResponseNewParams{}
			config := convertConfigFromResponseReq(req)

			Convey("Then all fields should be nil/zero", func() {
				So(config.MaxTokens, ShouldBeNil)
				So(config.Temperature, ShouldBeNil)
				So(config.TopP, ShouldBeNil)
				So(config.ReasoningConfig, ShouldBeNil)
				So(config.Grammar, ShouldBeNil)
			})
		})
	})
}

func TestConvertReqFromResponse(t *testing.T) {
	Convey("Given a ResponseNewParams to convert", t, func() {

		Convey("When input is a simple string", func() {
			req := &responses.ResponseNewParams{
				Model: "gpt-4o",
				Input: responses.ResponseNewParamsInputUnion{
					OfString: openai.Opt("Hello!"),
				},
			}
			result := convertReqFromResponse(req)

			Convey("Then it should create a single user message", func() {
				So(result.Model, ShouldEqual, "gpt-4o")
				So(result.Messages, ShouldHaveLength, 1)
				So(result.Messages[0].Role, ShouldEqual, v1.Role_USER)
				So(result.Messages[0].Contents[0].GetText(), ShouldEqual, "Hello!")
			})
		})

		Convey("When instructions are provided", func() {
			req := &responses.ResponseNewParams{
				Model:        "gpt-4o",
				Instructions: openai.Opt("You are a helpful assistant."),
				Input: responses.ResponseNewParamsInputUnion{
					OfString: openai.Opt("Hi"),
				},
			}
			result := convertReqFromResponse(req)

			Convey("Then instructions should be prepended as system message", func() {
				So(result.Messages, ShouldHaveLength, 2)
				So(result.Messages[0].Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Messages[0].Contents[0].GetText(), ShouldEqual, "You are a helpful assistant.")
				So(result.Messages[1].Role, ShouldEqual, v1.Role_USER)
				So(result.Messages[1].Contents[0].GetText(), ShouldEqual, "Hi")
			})
		})

		Convey("When input is an array of easy messages", func() {
			req := &responses.ResponseNewParams{
				Model: "gpt-4o",
				Input: responses.ResponseNewParamsInputUnion{
					OfInputItemList: responses.ResponseInputParam{
						{OfMessage: &responses.EasyInputMessageParam{
							Role:    responses.EasyInputMessageRoleSystem,
							Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt("Be helpful.")},
						}},
						{OfMessage: &responses.EasyInputMessageParam{
							Role:    responses.EasyInputMessageRoleUser,
							Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt("Hello")},
						}},
					},
				},
			}
			result := convertReqFromResponse(req)

			Convey("Then messages should be converted with correct roles", func() {
				So(result.Messages, ShouldHaveLength, 2)
				So(result.Messages[0].Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Messages[0].Contents[0].GetText(), ShouldEqual, "Be helpful.")
				So(result.Messages[1].Role, ShouldEqual, v1.Role_USER)
				So(result.Messages[1].Contents[0].GetText(), ShouldEqual, "Hello")
			})
		})

		Convey("When input contains function_call and function_call_output items", func() {
			req := &responses.ResponseNewParams{
				Model: "gpt-4o",
				Input: responses.ResponseNewParamsInputUnion{
					OfInputItemList: responses.ResponseInputParam{
						{OfMessage: &responses.EasyInputMessageParam{
							Role:    responses.EasyInputMessageRoleUser,
							Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt("What's the weather?")},
						}},
						{OfFunctionCall: &responses.ResponseFunctionToolCallParam{
							CallID:    "call_123",
							Name:      "get_weather",
							Arguments: `{"city":"Shanghai"}`,
						}},
						{OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
							CallID: "call_123",
							Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
								OfString: openai.Opt(`{"temp":22}`),
							},
						}},
					},
				},
			}
			result := convertReqFromResponse(req)

			Convey("Then function_call should be ToolUse and output should be ToolResult", func() {
				So(result.Messages, ShouldHaveLength, 3)
				So(result.Messages[0].Role, ShouldEqual, v1.Role_USER)

				So(result.Messages[1].Role, ShouldEqual, v1.Role_MODEL)
				tu := result.Messages[1].Contents[0].GetToolUse()
				So(tu, ShouldNotBeNil)
				So(tu.Id, ShouldEqual, "call_123")
				So(tu.Name, ShouldEqual, "get_weather")
				So(tu.GetTextualInput(), ShouldEqual, `{"city":"Shanghai"}`)

				So(result.Messages[2].Role, ShouldEqual, v1.Role_USER)
				tr := result.Messages[2].Contents[0].GetToolResult()
				So(tr, ShouldNotBeNil)
				So(tr.Id, ShouldEqual, "call_123")
				So(tr.Outputs[0].GetText(), ShouldEqual, `{"temp":22}`)
			})
		})

		Convey("When tools are provided", func() {
			req := &responses.ResponseNewParams{
				Model: "gpt-4o",
				Input: responses.ResponseNewParamsInputUnion{
					OfString: openai.Opt("Hello"),
				},
				Tools: []responses.ToolUnionParam{
					{OfFunction: &responses.FunctionToolParam{
						Name:        "get_weather",
						Description: openai.Opt("Get weather info"),
						Parameters: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"city": map[string]any{"type": "string"},
							},
						},
					}},
				},
			}
			result := convertReqFromResponse(req)

			Convey("Then tools should be converted", func() {
				So(result.Tools, ShouldHaveLength, 1)
				fn := result.Tools[0].GetFunction()
				So(fn, ShouldNotBeNil)
				So(fn.Name, ShouldEqual, "get_weather")
				So(fn.Description, ShouldEqual, "Get weather info")
			})
		})

		Convey("When developer role message is used", func() {
			req := &responses.ResponseNewParams{
				Model: "gpt-4o",
				Input: responses.ResponseNewParamsInputUnion{
					OfInputItemList: responses.ResponseInputParam{
						{OfMessage: &responses.EasyInputMessageParam{
							Role:    responses.EasyInputMessageRoleDeveloper,
							Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt("Developer instructions")},
						}},
					},
				},
			}
			result := convertReqFromResponse(req)

			Convey("Then developer should map to SYSTEM", func() {
				So(result.Messages[0].Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Messages[0].Contents[0].GetText(), ShouldEqual, "Developer instructions")
			})
		})

		Convey("When assistant role message is used", func() {
			req := &responses.ResponseNewParams{
				Model: "gpt-4o",
				Input: responses.ResponseNewParamsInputUnion{
					OfInputItemList: responses.ResponseInputParam{
						{OfMessage: &responses.EasyInputMessageParam{
							Role:    responses.EasyInputMessageRoleAssistant,
							Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt("Previous reply")},
						}},
					},
				},
			}
			result := convertReqFromResponse(req)

			Convey("Then assistant should map to MODEL", func() {
				So(result.Messages[0].Role, ShouldEqual, v1.Role_MODEL)
				So(result.Messages[0].Contents[0].GetText(), ShouldEqual, "Previous reply")
			})
		})
	})
}

func TestConvertStatusToResponse(t *testing.T) {
	Convey("Given various internal chat statuses", t, func() {
		So(convertStatusToResponse(v1.ChatStatus_CHAT_COMPLETED), ShouldEqual, "completed")
		So(convertStatusToResponse(v1.ChatStatus_CHAT_PENDING_TOOL_USE), ShouldEqual, "completed")
		So(convertStatusToResponse(v1.ChatStatus_CHAT_FAILED), ShouldEqual, "failed")
		So(convertStatusToResponse(v1.ChatStatus_CHAT_CANCELLED), ShouldEqual, "cancelled")
		So(convertStatusToResponse(v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT), ShouldEqual, "incomplete")
		So(convertStatusToResponse(v1.ChatStatus_CHAT_IN_PROGRESS), ShouldEqual, "in_progress")
	})
}

func TestConvertUsageToResponse(t *testing.T) {
	Convey("Given usage statistics to convert", t, func() {

		Convey("When usage is nil", func() {
			So(convertUsageToResponse(nil), ShouldBeNil)
		})

		Convey("When usage has all fields", func() {
			u := &v1.Statistics_Usage{
				InputTokens:       100,
				OutputTokens:      50,
				CachedInputTokens: 20,
				ReasoningTokens:   10,
			}
			result := convertUsageToResponse(u)

			Convey("Then all fields should be mapped", func() {
				So(result.InputTokens, ShouldEqual, 100)
				So(result.OutputTokens, ShouldEqual, 50)
				So(result.TotalTokens, ShouldEqual, 150)
				So(result.InputTokensDetails.CachedTokens, ShouldEqual, 20)
				So(result.OutputTokenDetails.ReasoningTokens, ShouldEqual, 10)
			})
		})
	})
}

func TestConvertRespToResponse(t *testing.T) {
	Convey("Given an internal ChatResp to convert to Responses API format", t, func() {

		Convey("When converting a text response", func() {
			resp := &v1.ChatResp{
				Id:     "test-id-123",
				Model:  "gpt-4o",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: "Hello!"}},
					},
				},
				Statistics: &v1.Statistics{
					Usage: &v1.Statistics_Usage{
						InputTokens:  10,
						OutputTokens: 5,
					},
				},
			}
			result := convertRespToResponse(resp)

			Convey("Then the response should be correctly structured", func() {
				So(result.Object, ShouldEqual, "response")
				So(result.Model, ShouldEqual, "gpt-4o")
				So(result.Status, ShouldEqual, "completed")
				So(result.Output, ShouldHaveLength, 1)

				msg, ok := result.Output[0].(responseOutputMessage)
				So(ok, ShouldBeTrue)
				So(msg.Type, ShouldEqual, "message")
				So(msg.Role, ShouldEqual, "assistant")
				So(msg.Status, ShouldEqual, "completed")
				So(msg.Content, ShouldHaveLength, 1)

				text, ok := msg.Content[0].(responseOutputText)
				So(ok, ShouldBeTrue)
				So(text.Type, ShouldEqual, "output_text")
				So(text.Text, ShouldEqual, "Hello!")
				So(text.Annotations, ShouldBeEmpty)

				So(result.Usage, ShouldNotBeNil)
				So(result.Usage.InputTokens, ShouldEqual, 10)
				So(result.Usage.OutputTokens, ShouldEqual, 5)
				So(result.Usage.TotalTokens, ShouldEqual, 15)
			})
		})

		Convey("When converting a response with function calls", func() {
			resp := &v1.ChatResp{
				Id:     "test-id-456",
				Model:  "gpt-4o",
				Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: "Let me check."}},
						{Content: &v1.Content_ToolUse{
							ToolUse: &v1.ToolUse{
								Id:   "call_abc",
								Name: "get_weather",
								Inputs: []*v1.ToolUse_Input{{
									Input: &v1.ToolUse_Input_Text{Text: `{"city":"Shanghai"}`},
								}},
							},
						}},
					},
				},
			}
			result := convertRespToResponse(resp)

			Convey("Then output should contain message and function_call", func() {
				So(result.Status, ShouldEqual, "completed")
				So(result.Output, ShouldHaveLength, 2)

				msg, ok := result.Output[0].(responseOutputMessage)
				So(ok, ShouldBeTrue)
				So(msg.Content, ShouldHaveLength, 1)

				fc, ok := result.Output[1].(responseFunctionCall)
				So(ok, ShouldBeTrue)
				So(fc.Type, ShouldEqual, "function_call")
				So(fc.CallID, ShouldEqual, "call_abc")
				So(fc.Name, ShouldEqual, "get_weather")
				So(fc.Arguments, ShouldEqual, `{"city":"Shanghai"}`)
				So(fc.Status, ShouldEqual, "completed")
			})
		})

		Convey("When converting a response with reasoning content", func() {
			resp := &v1.ChatResp{
				Id:     "test-id-789",
				Model:  "o3",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Reasoning: true,
							Metadata:  map[string]string{"summary": "Analyzing the problem."},
							Content:   &v1.Content_Text{Text: "Let me think..."},
						},
						{
							Content: &v1.Content_Text{Text: "The answer is 42."},
						},
					},
				},
			}
			result := convertRespToResponse(resp)

			Convey("Then output should contain reasoning and message items", func() {
				So(result.Output, ShouldHaveLength, 2)

				reasoning, ok := result.Output[0].(responseReasoning)
				So(ok, ShouldBeTrue)
				So(reasoning.Type, ShouldEqual, "reasoning")
				So(reasoning.Summary, ShouldHaveLength, 1)
				So(reasoning.Summary[0].Text, ShouldEqual, "Analyzing the problem.")
				So(reasoning.Content, ShouldHaveLength, 1)
				So(reasoning.Content[0].Type, ShouldEqual, "reasoning_text")
				So(reasoning.Content[0].Text, ShouldEqual, "Let me think...")

				msg, ok := result.Output[1].(responseOutputMessage)
				So(ok, ShouldBeTrue)
				So(msg.Content, ShouldHaveLength, 1)
				text, ok := msg.Content[0].(responseOutputText)
				So(ok, ShouldBeTrue)
				So(text.Text, ShouldEqual, "The answer is 42.")
			})
		})

		Convey("When converting a refused response", func() {
			resp := &v1.ChatResp{
				Id:     "test-id-ref",
				Model:  "gpt-4o",
				Status: v1.ChatStatus_CHAT_REFUSED,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: "I cannot do that."}},
					},
				},
			}
			result := convertRespToResponse(resp)

			Convey("Then content should be a refusal", func() {
				So(result.Status, ShouldEqual, "completed")
				msg, ok := result.Output[0].(responseOutputMessage)
				So(ok, ShouldBeTrue)
				refusal, ok := msg.Content[0].(responseRefusal)
				So(ok, ShouldBeTrue)
				So(refusal.Type, ShouldEqual, "refusal")
				So(refusal.Refusal, ShouldEqual, "I cannot do that.")
			})
		})

		Convey("When converting interleaved output items", func() {
			resp := &v1.ChatResp{
				Id:     "test-id-order",
				Model:  "o3",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Id: "msg_1", Content: &v1.Content_Text{Text: "First."}},
						{Id: "fc_1", Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{Id: "call_1", Name: "lookup", Inputs: []*v1.ToolUse_Input{{Input: &v1.ToolUse_Input_Text{Text: `{"q":"x"}`}}}}}},
						{Id: "rs_1", Reasoning: true, Metadata: map[string]string{"summary": "Thinking."}, Content: &v1.Content_Text{Text: "Hidden chain."}},
						{Id: "msg_2", Content: &v1.Content_Text{Text: "Second."}},
					},
				},
			}
			result := convertRespToResponse(resp)

			Convey("Then output order should match input order", func() {
				So(result.Output, ShouldHaveLength, 4)

				msg1, ok := result.Output[0].(responseOutputMessage)
				So(ok, ShouldBeTrue)
				So(msg1.ID, ShouldEqual, "msg_1")
				text1, ok := msg1.Content[0].(responseOutputText)
				So(ok, ShouldBeTrue)
				So(text1.Text, ShouldEqual, "First.")

				fc, ok := result.Output[1].(responseFunctionCall)
				So(ok, ShouldBeTrue)
				So(fc.ID, ShouldEqual, "fc_1")
				So(fc.CallID, ShouldEqual, "call_1")

				reasoning, ok := result.Output[2].(responseReasoning)
				So(ok, ShouldBeTrue)
				So(reasoning.ID, ShouldEqual, "rs_1")
				So(reasoning.Summary, ShouldHaveLength, 1)
				So(reasoning.Summary[0].Text, ShouldEqual, "Thinking.")
				So(reasoning.Content, ShouldHaveLength, 1)
				So(reasoning.Content[0].Text, ShouldEqual, "Hidden chain.")

				msg2, ok := result.Output[3].(responseOutputMessage)
				So(ok, ShouldBeTrue)
				So(msg2.ID, ShouldEqual, "msg_2")
			})
		})

		Convey("When message is nil", func() {
			resp := &v1.ChatResp{
				Id:    "test-id-nil",
				Model: "gpt-4o",
			}
			result := convertRespToResponse(resp)

			Convey("Then output should be empty", func() {
				So(result.Output, ShouldBeEmpty)
			})
		})
	})
}

func TestConvertEasyInputMessageWithImage(t *testing.T) {
	Convey("Given an EasyInputMessage with an image url part", t, func() {
		m := &responses.EasyInputMessageParam{
			Role: responses.EasyInputMessageRoleUser,
			Content: responses.EasyInputMessageContentUnionParam{
				OfInputItemContentList: responses.ResponseInputMessageContentListParam{
					{OfInputText: &responses.ResponseInputTextParam{Text: "what is in this image"}},
					{OfInputImage: &responses.ResponseInputImageParam{
						ImageURL: openai.Opt("https://example.com/cat.png"),
					}},
				},
			},
		}

		msg := convertEasyInputMessageFromResponse(m)

		Convey("Then both text and image contents should be preserved", func() {
			So(msg.Role, ShouldEqual, v1.Role_USER)
			So(msg.Contents, ShouldHaveLength, 2)
			So(msg.Contents[0].GetText(), ShouldEqual, "what is in this image")
			img := msg.Contents[1].GetImage()
			So(img, ShouldNotBeNil)
		})
	})
}

func TestConvertOutputMessageFromResponseWithRefusal(t *testing.T) {
	Convey("Given an OutputMessage with both text and refusal parts", t, func() {
		m := &responses.ResponseOutputMessageParam{
			ID: "msg_history_1",
			Content: []responses.ResponseOutputMessageContentUnionParam{
				{OfOutputText: &responses.ResponseOutputTextParam{Text: "partial answer"}},
				{OfRefusal: &responses.ResponseOutputRefusalParam{Refusal: "I cannot answer that."}},
			},
		}

		msg := convertOutputMessageFromResponse(m)

		Convey("Then both contents should be present and refusal flagged", func() {
			So(msg.Role, ShouldEqual, v1.Role_MODEL)
			So(msg.Contents, ShouldHaveLength, 2)
			So(msg.Contents[0].GetText(), ShouldEqual, "partial answer")
			So(msg.Contents[0].Meta("refusal"), ShouldBeEmpty)
			So(msg.Contents[1].GetText(), ShouldEqual, "I cannot answer that.")
			So(msg.Contents[1].Meta("refusal"), ShouldEqual, "true")
		})
	})
}

func TestConvertReasoningFromResponseAlignment(t *testing.T) {
	Convey("Given a ResponseReasoningItem from a previous turn", t, func() {
		Convey("With encrypted content, summary parts and reasoning text", func() {
			r := &responses.ResponseReasoningItemParam{
				ID:               "rs_history_1",
				EncryptedContent: openai.Opt("opaque-bytes"),
				Summary: []responses.ResponseReasoningItemSummaryParam{
					{Text: "first summary"},
					{Text: "second summary"},
				},
				Content: []responses.ResponseReasoningItemContentParam{
					{Text: "internal thought"},
				},
			}

			msg := convertReasoningFromResponse(r)

			Convey("Then layout matches the upstream incoming converter", func() {
				So(msg.Role, ShouldEqual, v1.Role_MODEL)
				So(msg.Contents, ShouldHaveLength, 2)

				// First content carries encrypted + first summary, both
				// metadata-only with an empty text slot.
				first := msg.Contents[0]
				So(first.Reasoning, ShouldBeTrue)
				So(first.Meta("encrypted"), ShouldEqual, "opaque-bytes")
				So(first.Meta("summary"), ShouldEqual, "first summary")
				So(first.Meta("summary_index"), ShouldEqual, "0")
				So(first.GetText(), ShouldEqual, "")

				// Second content holds the next summary and absorbs the
				// reasoning text into its empty Content_Text slot, mirroring
				// the upstream incoming converter exactly.
				second := msg.Contents[1]
				So(second.Reasoning, ShouldBeTrue)
				So(second.Meta("encrypted"), ShouldBeEmpty)
				So(second.Meta("summary"), ShouldEqual, "second summary")
				So(second.Meta("summary_index"), ShouldEqual, "1")
				So(second.GetText(), ShouldEqual, "internal thought")
			})
		})

		Convey("With only encrypted content", func() {
			r := &responses.ResponseReasoningItemParam{
				ID:               "rs_history_2",
				EncryptedContent: openai.Opt("opaque"),
			}

			msg := convertReasoningFromResponse(r)

			Convey("Then a single content carries the encrypted blob", func() {
				So(msg.Contents, ShouldHaveLength, 1)
				So(msg.Contents[0].Meta("encrypted"), ShouldEqual, "opaque")
			})
		})

		Convey("With only summary parts", func() {
			r := &responses.ResponseReasoningItemParam{
				ID: "rs_history_3",
				Summary: []responses.ResponseReasoningItemSummaryParam{
					{Text: "only summary"},
				},
			}

			msg := convertReasoningFromResponse(r)

			Convey("Then summary_index is recorded", func() {
				So(msg.Contents, ShouldHaveLength, 1)
				So(msg.Contents[0].Meta("summary"), ShouldEqual, "only summary")
				So(msg.Contents[0].Meta("summary_index"), ShouldEqual, "0")
			})
		})
	})
}

func TestConvertInputItemsGroupsAssistantTurn(t *testing.T) {
	Convey("Given an input history with reasoning, function call and output_message in the same assistant turn", t, func() {
		items := []responses.ResponseInputItemUnionParam{
			{OfMessage: &responses.EasyInputMessageParam{
				Role:    responses.EasyInputMessageRoleUser,
				Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt("hello")},
			}},
			{OfReasoning: &responses.ResponseReasoningItemParam{
				ID: "rs_1",
				Summary: []responses.ResponseReasoningItemSummaryParam{
					{Text: "thinking"},
				},
			}},
			{OfFunctionCall: &responses.ResponseFunctionToolCallParam{
				CallID:    "call_1",
				Name:      "lookup",
				Arguments: `{"q":"x"}`,
			}},
			{OfOutputMessage: &responses.ResponseOutputMessageParam{
				ID: "msg_1",
				Content: []responses.ResponseOutputMessageContentUnionParam{
					{OfOutputText: &responses.ResponseOutputTextParam{Text: "here you go"}},
				},
			}},
			{OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
				CallID: "call_1",
				Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
					OfString: openai.Opt(`{"ok":true}`),
				},
			}},
		}

		messages := convertInputItemsFromResponse(items)

		Convey("Then the assistant turn is collapsed into one MODEL message", func() {
			So(messages, ShouldHaveLength, 3)
			So(messages[0].Role, ShouldEqual, v1.Role_USER)

			model := messages[1]
			So(model.Role, ShouldEqual, v1.Role_MODEL)
			So(len(model.Contents), ShouldEqual, 3)

			So(model.Contents[0].Reasoning, ShouldBeTrue)
			So(model.Contents[0].Meta("summary"), ShouldEqual, "thinking")

			tu := model.Contents[1].GetToolUse()
			So(tu, ShouldNotBeNil)
			So(tu.Id, ShouldEqual, "call_1")

			So(model.Contents[2].GetText(), ShouldEqual, "here you go")

			So(messages[2].Role, ShouldEqual, v1.Role_USER)
			tr := messages[2].Contents[0].GetToolResult()
			So(tr, ShouldNotBeNil)
		})
	})
}
