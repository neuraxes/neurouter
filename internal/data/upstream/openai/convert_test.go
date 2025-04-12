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

	"github.com/go-kratos/kratos/v2/log"
	"github.com/openai/openai-go"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
)

func TestConvertMessageToOpenAI(t *testing.T) {
	repo := &ChatRepo{
		config: &conf.OpenAIConfig{},
		log:    log.NewHelper(log.DefaultLogger),
	}

	singlePartTextualMessage := &v1.Message{
		Contents: []*v1.Content{
			{
				Content: &v1.Content_Text{
					Text: "You are helpful assistant.",
				},
			},
		},
	}

	multiPartTextualMessage := &v1.Message{
		Contents: []*v1.Content{
			{
				Content: &v1.Content_Text{
					Text: "You are helpful",
				},
			},
			{
				Content: &v1.Content_Text{
					Text: " assistant.",
				},
			},
		},
	}

	multiPartRichMessage := &v1.Message{
		Contents: []*v1.Content{
			{
				Content: &v1.Content_Text{
					Text: "Here is a image:",
				},
			},
			{
				Content: &v1.Content_Image_{
					Image: &v1.Content_Image{
						Url: "https://example.com/image.jpg",
					},
				},
			},
		},
	}

	Convey("Test for SYSTEM role", t, func() {
		repo.config = &conf.OpenAIConfig{}
		singlePartTextualMessage.Role = v1.Role_SYSTEM
		multiPartTextualMessage.Role = v1.Role_SYSTEM
		multiPartRichMessage.Role = v1.Role_SYSTEM

		Convey("with single part textual content", func() {
			param := repo.convertMessageToOpenAI(singlePartTextualMessage)
			So(param, ShouldEqual, openai.SystemMessage("You are helpful assistant."))
		})

		Convey("with multi part textual content", func() {
			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.ChatCompletionSystemMessageParam{
				Role: openai.F(openai.ChatCompletionSystemMessageParamRoleSystem),
				Content: openai.F([]openai.ChatCompletionContentPartTextParam{
					{
						Type: openai.F(openai.ChatCompletionContentPartTextTypeText),
						Text: openai.F("You are helpful"),
					},
					{
						Type: openai.F(openai.ChatCompletionContentPartTextTypeText),
						Text: openai.F(" assistant."),
					},
				}),
			})
		})

		Convey("with multi part rich content", func() {
			param := repo.convertMessageToOpenAI(multiPartRichMessage)
			So(param, ShouldEqual, openai.SystemMessage("Here is a image:"))
		})

		Convey("with name", func() {
			message := &v1.Message{
				Role: v1.Role_SYSTEM,
				Name: "System",
				Contents: []*v1.Content{
					{
						Content: &v1.Content_Text{
							Text: "Hello",
						},
					},
				},
			}
			param := repo.convertMessageToOpenAI(message)
			result := param.(openai.ChatCompletionSystemMessageParam)
			So(result.Name.Value, ShouldEqual, "System")
		})

		Convey("with PreferStringContentForSystem enabled", func() {
			repo.config = &conf.OpenAIConfig{
				PreferStringContentForSystem: true,
			}

			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.SystemMessage("You are helpful assistant."))
		})

		Convey("with PreferSinglePartContent enabled", func() {
			repo.config = &conf.OpenAIConfig{
				PreferSinglePartContent: true,
			}

			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.SystemMessage("You are helpful assistant."))
		})
	})

	Convey("Test for USER role", t, func() {
		repo.config = &conf.OpenAIConfig{}
		singlePartTextualMessage.Role = v1.Role_USER
		multiPartTextualMessage.Role = v1.Role_USER
		multiPartRichMessage.Role = v1.Role_USER

		Convey("with single part textual content", func() {
			param := repo.convertMessageToOpenAI(singlePartTextualMessage)
			So(param, ShouldEqual, openai.UserMessage("You are helpful assistant."))
		})

		Convey("with multi part textual content", func() {
			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.UserMessageParts(
				openai.TextPart("You are helpful"),
				openai.TextPart(" assistant."),
			))
		})

		Convey("with multi part rich content", func() {
			param := repo.convertMessageToOpenAI(multiPartRichMessage)
			So(param, ShouldEqual, openai.UserMessageParts(
				openai.TextPart("Here is a image:"),
				openai.ImagePart("https://example.com/image.jpg"),
			))
		})

		Convey("with name", func() {
			message := &v1.Message{
				Role: v1.Role_USER,
				Name: "User",
				Contents: []*v1.Content{
					{
						Content: &v1.Content_Text{
							Text: "Hello",
						},
					},
				},
			}
			param := repo.convertMessageToOpenAI(message)
			result := param.(openai.ChatCompletionUserMessageParam)
			So(result.Name.Value, ShouldEqual, "User")
		})

		Convey("with PreferStringContentForUser enabled", func() {
			repo.config = &conf.OpenAIConfig{
				PreferStringContentForUser: true,
			}

			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.UserMessage("You are helpful assistant."))
		})

		Convey("with PreferSinglePartContent enabled", func() {
			repo.config = &conf.OpenAIConfig{
				PreferSinglePartContent: true,
			}

			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.UserMessage("You are helpful assistant."))
		})

		Convey("with nil content", func() {
			message := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{
						Content: nil,
					},
				},
			}
			param := repo.convertMessageToOpenAI(message)
			result := param.(openai.ChatCompletionUserMessageParam)
			So(result.Role.Value, ShouldEqual, openai.ChatCompletionUserMessageParamRoleUser)
			So(result.Content.Value, ShouldBeEmpty)
		})

		Convey("with empty contents array", func() {
			message := &v1.Message{
				Role:     v1.Role_USER,
				Contents: []*v1.Content{},
			}
			param := repo.convertMessageToOpenAI(message)
			result := param.(openai.ChatCompletionUserMessageParam)
			So(result.Role.Value, ShouldEqual, openai.ChatCompletionUserMessageParamRoleUser)
			So(result.Content.Value, ShouldBeEmpty)
		})
	})

	Convey("Test for MODEL role", t, func() {
		repo.config = &conf.OpenAIConfig{}
		singlePartTextualMessage.Role = v1.Role_MODEL
		multiPartTextualMessage.Role = v1.Role_MODEL
		multiPartRichMessage.Role = v1.Role_MODEL

		Convey("with single part textual content", func() {
			param := repo.convertMessageToOpenAI(singlePartTextualMessage)
			So(param, ShouldEqual, openai.AssistantMessage("You are helpful assistant."))
		})

		Convey("with multi part textual content", func() {
			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.ChatCompletionAssistantMessageParam{
				Role: openai.F(openai.ChatCompletionAssistantMessageParamRoleAssistant),
				Content: openai.F([]openai.ChatCompletionAssistantMessageParamContentUnion{
					openai.TextPart("You are helpful"),
					openai.TextPart(" assistant."),
				}),
			})
		})

		Convey("with multi part rich content", func() {
			param := repo.convertMessageToOpenAI(multiPartRichMessage)
			So(param, ShouldEqual, openai.AssistantMessage("Here is a image:"))
		})

		Convey("with PreferStringContentForAssistant enabled", func() {
			repo.config = &conf.OpenAIConfig{
				PreferStringContentForAssistant: true,
			}

			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.AssistantMessage("You are helpful assistant."))
		})

		Convey("with PreferSinglePartContent enabled", func() {
			repo.config = &conf.OpenAIConfig{
				PreferSinglePartContent: true,
			}

			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.AssistantMessage("You are helpful assistant."))
		})

		Convey("with name", func() {
			message := &v1.Message{
				Role: v1.Role_MODEL,
				Name: "Claude",
				Contents: []*v1.Content{
					{
						Content: &v1.Content_Text{
							Text: "Hello",
						},
					},
				},
			}
			param := repo.convertMessageToOpenAI(message)
			result := param.(openai.ChatCompletionAssistantMessageParam)
			So(result.Name.Value, ShouldEqual, "Claude")
		})

		Convey("with tool calls", func() {
			message := &v1.Message{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{
						Content: &v1.Content_Text{
							Text: "Let me help you",
						},
					},
				},
				ToolCalls: []*v1.ToolCall{
					{
						Id: "call-1",
						Tool: &v1.ToolCall_Function{
							Function: &v1.ToolCall_FunctionCall{
								Name:      "search",
								Arguments: `{"query":"weather"}`,
							},
						},
					},
					{
						Id: "call-2",
						Tool: &v1.ToolCall_Function{
							Function: &v1.ToolCall_FunctionCall{
								Name:      "calculate",
								Arguments: `{"expression":"1+1"}`,
							},
						},
					},
				},
			}
			param := repo.convertMessageToOpenAI(message)
			result := param.(openai.ChatCompletionAssistantMessageParam)
			calls := result.ToolCalls.Value
			So(len(calls), ShouldEqual, 2)
			So(calls[0].ID.Value, ShouldEqual, "call-1")
			So(calls[0].Type.Value, ShouldEqual, openai.ChatCompletionMessageToolCallTypeFunction)
			So(calls[0].Function.Value.Name.Value, ShouldEqual, "search")
			So(calls[0].Function.Value.Arguments.Value, ShouldEqual, `{"query":"weather"}`)
			So(calls[1].ID.Value, ShouldEqual, "call-2")
			So(calls[1].Function.Value.Name.Value, ShouldEqual, "calculate")
			So(calls[1].Function.Value.Arguments.Value, ShouldEqual, `{"expression":"1+1"}`)
		})

		Convey("with unsupported tool call type", func() {
			message := &v1.Message{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{
						Content: &v1.Content_Text{
							Text: "Hello",
						},
					},
				},
				ToolCalls: []*v1.ToolCall{
					{
						Id:   "call-1",
						Tool: nil,
					},
				},
			}
			param := repo.convertMessageToOpenAI(message)
			result := param.(openai.ChatCompletionAssistantMessageParam)
			So(result.ToolCalls.Value, ShouldBeEmpty)
		})
	})

	Convey("Test for TOOL role", t, func() {
		repo.config = &conf.OpenAIConfig{}
		singlePartTextualMessage.Role = v1.Role_TOOL
		singlePartTextualMessage.ToolCallId = "tool-call-id-1"
		multiPartTextualMessage.Role = v1.Role_TOOL
		multiPartTextualMessage.ToolCallId = "tool-call-id-2"
		multiPartRichMessage.Role = v1.Role_TOOL
		multiPartRichMessage.ToolCallId = "tool-call-id-3"

		Convey("with single part textual content", func() {
			param := repo.convertMessageToOpenAI(singlePartTextualMessage)
			So(param, ShouldEqual, openai.ToolMessage("tool-call-id-1", "You are helpful assistant."))
		})

		Convey("with multi part textual content", func() {
			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.ChatCompletionToolMessageParam{
				Role:       openai.F(openai.ChatCompletionToolMessageParamRoleTool),
				ToolCallID: openai.F("tool-call-id-2"),
				Content: openai.F([]openai.ChatCompletionContentPartTextParam{
					{
						Type: openai.F(openai.ChatCompletionContentPartTextTypeText),
						Text: openai.F("You are helpful"),
					},
					{
						Type: openai.F(openai.ChatCompletionContentPartTextTypeText),
						Text: openai.F(" assistant."),
					},
				}),
			})
		})

		Convey("with multi part rich content", func() {
			param := repo.convertMessageToOpenAI(multiPartRichMessage)
			So(param, ShouldEqual, openai.ToolMessage("tool-call-id-3", "Here is a image:"))
		})

		Convey("with PreferStringContentForTool enabled", func() {
			repo.config = &conf.OpenAIConfig{
				PreferStringContentForTool: true,
			}

			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.ToolMessage("tool-call-id-2", "You are helpful assistant."))
		})

		Convey("with PreferSinglePartContent enabled", func() {
			repo.config = &conf.OpenAIConfig{
				PreferSinglePartContent: true,
			}

			param := repo.convertMessageToOpenAI(multiPartTextualMessage)
			So(param, ShouldEqual, openai.ToolMessage("tool-call-id-2", "You are helpful assistant."))
		})

		Convey("with unsupported role", func() {
			invalidMessage := &v1.Message{
				Role: v1.Role(-1),
				Contents: []*v1.Content{
					{
						Content: &v1.Content_Text{
							Text: "Test",
						},
					},
				},
			}
			param := repo.convertMessageToOpenAI(invalidMessage)
			So(param, ShouldBeNil)
		})
	})
}

func TestConvertRequestToOpenAI(t *testing.T) {
	repo := &ChatRepo{
		config: &conf.OpenAIConfig{},
		log:    log.NewHelper(log.DefaultLogger),
	}

	Convey("Test convertRequestToOpenAI", t, func() {
		Convey("with basic request", func() {
			req := &entity.ChatReq{
				Model: "gpt-4",
				Messages: []*v1.Message{
					{
						Role: v1.Role_USER,
						Contents: []*v1.Content{
							{
								Content: &v1.Content_Text{
									Text: "Hello",
								},
							},
						},
					},
				},
			}

			param := repo.convertRequestToOpenAI(req)
			So(param.Model.Value, ShouldEqual, "gpt-4")
			So(len(param.Messages.Value), ShouldEqual, 1)
		})

		Convey("with configuration", func() {
			req := &entity.ChatReq{
				Model: "gpt-4",
				Config: &v1.GenerationConfig{
					MaxTokens:        100,
					Temperature:      0.7,
					TopP:             0.9,
					FrequencyPenalty: 1.0,
					PresencePenalty:  1.0,
					Grammar: &v1.GenerationConfig_PresetGrammar{
						PresetGrammar: "json_object",
					},
				},
			}

			param := repo.convertRequestToOpenAI(req)
			So(param.MaxCompletionTokens.Value, ShouldEqual, 100)
			So(param.Temperature.Value, ShouldAlmostEqual, 0.7, 0.000001)
			So(param.TopP.Value, ShouldAlmostEqual, 0.9, 0.000001)
			So(param.FrequencyPenalty.Value, ShouldEqual, 1.0)
			So(param.PresencePenalty.Value, ShouldEqual, 1.0)
			So(param.ResponseFormat, ShouldNotBeNil)
		})

		Convey("with tools", func() {
			req := &entity.ChatReq{
				Model: "gpt-4",
				Tools: []*v1.Tool{
					{
						Tool: &v1.Tool_Function_{
							Function: &v1.Tool_Function{
								Name:        "test_function",
								Description: "Test function description",
								Parameters: &v1.Tool_Function_Parameters{
									Type: "object",
									Properties: map[string]*v1.Tool_Function_Parameters_Property{
										"prop1": {
											Type: "string",
										},
									},
									Required: []string{"prop1"},
								},
							},
						},
					},
				},
			}

			param := repo.convertRequestToOpenAI(req)
			tools := param.Tools.Value
			So(len(tools), ShouldEqual, 1)
			So(tools[0].Type.Value, ShouldEqual, openai.ChatCompletionToolTypeFunction)
			So(tools[0].Function.Value.Name.Value, ShouldEqual, "test_function")
			So(tools[0].Function.Value.Description.Value, ShouldEqual, "Test function description")
			So(tools[0].Function.Value.Parameters.Value, ShouldResemble, openai.FunctionParameters{
				"type": "object",
				"properties": map[string]*v1.Tool_Function_Parameters_Property{
					"prop1": {
						Type: "string",
					},
				},
				"required": []string{"prop1"},
			})
		})

		Convey("with nil config", func() {
			req := &entity.ChatReq{
				Model:    "gpt-4",
				Messages: []*v1.Message{},
				Config:   nil,
			}

			param := repo.convertRequestToOpenAI(req)
			So(param.Model.Value, ShouldEqual, "gpt-4")
			So(param.Temperature.Value, ShouldEqual, 0)
			So(param.MaxCompletionTokens.Present, ShouldBeFalse)
		})

		Convey("with unsupported tool type", func() {
			req := &entity.ChatReq{
				Model: "gpt-4",
				Tools: []*v1.Tool{
					{
						Tool: nil,
					},
				},
			}

			param := repo.convertRequestToOpenAI(req)
			So(param.Tools, ShouldNotBeNil)
			So(len(param.Tools.Value), ShouldEqual, 0)
		})
	})
}

func TestConvertMessageFromOpenAI(t *testing.T) {
	repo := &ChatRepo{
		config: &conf.OpenAIConfig{},
		log:    log.NewHelper(log.DefaultLogger),
	}

	Convey("Test convertMessageFromOpenAI", t, func() {
		Convey("with text content", func() {
			openAIMsg := &openai.ChatCompletionMessage{
				Content: "Hello world",
			}

			msg := repo.convertMessageFromOpenAI(openAIMsg)
			So(msg.Role, ShouldEqual, v1.Role_MODEL)
			So(msg.Contents[0].GetText(), ShouldEqual, "Hello world")
		})

		Convey("with tool calls", func() {
			openAIMsg := &openai.ChatCompletionMessage{
				Content: "",
				ToolCalls: []openai.ChatCompletionMessageToolCall{
					{
						ID:   "call-1",
						Type: openai.ChatCompletionMessageToolCallTypeFunction,
						Function: openai.ChatCompletionMessageToolCallFunction{
							Name:      "test_function",
							Arguments: `{"arg1":"value1"}`,
						},
					},
				},
			}

			msg := repo.convertMessageFromOpenAI(openAIMsg)
			So(msg.Role, ShouldEqual, v1.Role_MODEL)
			So(len(msg.ToolCalls), ShouldEqual, 1)
			So(msg.ToolCalls[0].Id, ShouldEqual, "call-1")
			So(msg.ToolCalls[0].GetFunction().Name, ShouldEqual, "test_function")
			So(msg.ToolCalls[0].GetFunction().Arguments, ShouldEqual, `{"arg1":"value1"}`)
		})

		Convey("with empty content and no tool calls", func() {
			openAIMsg := &openai.ChatCompletionMessage{
				Content:   "",
				ToolCalls: nil,
			}

			msg := repo.convertMessageFromOpenAI(openAIMsg)
			So(msg.Role, ShouldEqual, v1.Role_MODEL)
			So(msg.Contents, ShouldBeNil)
			So(msg.ToolCalls, ShouldBeNil)
		})

		Convey("with unsupported tool call type", func() {
			openAIMsg := &openai.ChatCompletionMessage{
				Content: "",
				ToolCalls: []openai.ChatCompletionMessageToolCall{
					{
						ID:   "call-1",
						Type: "unknown",
					},
				},
			}

			msg := repo.convertMessageFromOpenAI(openAIMsg)
			So(msg.Role, ShouldEqual, v1.Role_MODEL)
			So(msg.ToolCalls, ShouldBeNil)
		})
	})
}

func TestConvertResponseFromOpenAI(t *testing.T) {
	repo := &ChatRepo{
		config: &conf.OpenAIConfig{},
		log:    log.NewHelper(log.DefaultLogger),
	}

	Convey("Test convertResponseFromOpenAI", t, func() {
		Convey("with basic response", func() {
			openAIResp := &openai.ChatCompletion{
				ID: "resp-1",
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Hello world",
						},
					},
				},
			}

			resp := repo.convertResponseFromOpenAI(openAIResp)
			So(resp.Id, ShouldEqual, "resp-1")
			So(resp.Message.Contents[0].GetText(), ShouldEqual, "Hello world")
		})

		Convey("with usage statistics", func() {
			openAIResp := &openai.ChatCompletion{
				ID: "resp-1",
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Hello world",
						},
					},
				},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 20,
				},
			}

			resp := repo.convertResponseFromOpenAI(openAIResp)
			So(resp.Statistics.Usage.PromptTokens, ShouldEqual, 10)
			So(resp.Statistics.Usage.CompletionTokens, ShouldEqual, 20)
		})
	})
}

func TestConvertChunkFromOpenAI(t *testing.T) {
	Convey("Test convertChunkFromOpenAI", t, func() {
		Convey("with content", func() {
			chunk := &openai.ChatCompletionChunk{
				Choices: []openai.ChatCompletionChunkChoice{
					{
						Delta: openai.ChatCompletionChunkChoicesDelta{
							Content: "Hello",
						},
					},
				},
			}

			resp := convertChunkFromOpenAI(chunk, "req-1", "msg-1")
			So(resp.Id, ShouldEqual, "req-1")
			So(resp.Message.Id, ShouldEqual, "msg-1")
			So(resp.Message.Contents[0].GetText(), ShouldEqual, "Hello")
		})

		Convey("with usage statistics", func() {
			chunk := &openai.ChatCompletionChunk{
				Choices: []openai.ChatCompletionChunkChoice{
					{
						Delta: openai.ChatCompletionChunkChoicesDelta{
							Content: "Hello",
						},
					},
				},
				Usage: openai.CompletionUsage{
					PromptTokens:     5,
					CompletionTokens: 10,
				},
			}

			resp := convertChunkFromOpenAI(chunk, "req-1", "msg-1")
			So(resp.Statistics.Usage.PromptTokens, ShouldEqual, 5)
			So(resp.Statistics.Usage.CompletionTokens, ShouldEqual, 10)
		})
	})
}
