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

package deepseek

import (
	"encoding/json"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
)

func TestConvertMessageToDeepSeek(t *testing.T) {
	Convey("Given a upstream instance", t, func() {
		config := &conf.DeepSeekConfig{BaseUrl: "https://api.deepseek.com", ApiKey: "test-key"}
		repo := &upstream{
			config: config,
			log:    log.NewHelper(log.DefaultLogger),
		}

		Convey("When converting a system message", func() {
			message := &v1.Message{
				Id:   "msg-1",
				Role: v1.Role_SYSTEM,
				Name: "system_name",
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "You are a helpful assistant"}},
				},
			}

			result := repo.convertMessageToDeepSeek(message)

			Convey("Then the converted message should have correct role and content", func() {
				So(result.Role, ShouldEqual, "system")
				So(result.Content, ShouldEqual, "You are a helpful assistant")
				So(result.Name, ShouldEqual, "system_name")
			})
		})

		Convey("When converting a user message", func() {
			message := &v1.Message{
				Id:   "msg-2",
				Role: v1.Role_USER,
				Name: "user_name",
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Hello, how are you?"}},
				},
			}

			result := repo.convertMessageToDeepSeek(message)

			Convey("Then the converted message should have correct role and content", func() {
				So(result.Role, ShouldEqual, "user")
				So(result.Content, ShouldEqual, "Hello, how are you?")
				So(result.Name, ShouldEqual, "user_name")
			})
		})

		Convey("When converting a model message", func() {
			message := &v1.Message{
				Id:   "msg-3",
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "I'm doing well, thank you!"}},
				},
			}

			result := repo.convertMessageToDeepSeek(message)

			Convey("Then the converted message should have correct role and content", func() {
				So(result.Role, ShouldEqual, "assistant")
				So(result.Content, ShouldEqual, "I'm doing well, thank you!")
			})
		})

		Convey("When converting a tool message", func() {
			message := &v1.Message{
				Id:         "msg-4",
				Role:       v1.Role_TOOL,
				ToolCallId: "tool-call-1",
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Function result"}},
				},
			}

			result := repo.convertMessageToDeepSeek(message)

			Convey("Then the converted message should have correct role and tool call ID", func() {
				So(result.Role, ShouldEqual, "tool")
				So(result.Content, ShouldEqual, "Function result")
				So(result.ToolCallID, ShouldEqual, "tool-call-1")
			})
		})

		Convey("When converting a message with function calls", func() {
			message := &v1.Message{
				Id:   "msg-5",
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Let me call a function"}},
					{Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Id:        "call-1",
							Name:      "get_weather",
							Arguments: `{"location": "New York"}`,
						},
					}},
				},
			}

			result := repo.convertMessageToDeepSeek(message)

			Convey("Then the converted message should have both text content and tool calls", func() {
				So(result.Role, ShouldEqual, "assistant")
				So(result.Content, ShouldEqual, "Let me call a function")
				So(len(result.ToolCalls), ShouldEqual, 1)
				So(result.ToolCalls[0].ID, ShouldEqual, "call-1")
				So(result.ToolCalls[0].Type, ShouldEqual, "function")
				So(result.ToolCalls[0].Function.Name, ShouldEqual, "get_weather")
				So(result.ToolCalls[0].Function.Arguments, ShouldEqual, `{"location": "New York"}`)
			})
		})

		Convey("When converting a message with multiple text contents", func() {
			message := &v1.Message{
				Id:   "msg-6",
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "First part "}},
					{Content: &v1.Content_Text{Text: "second part"}},
				},
			}

			result := repo.convertMessageToDeepSeek(message)

			Convey("Then the text contents should be concatenated", func() {
				So(result.Content, ShouldEqual, "First part second part")
			})
		})
	})
}

func TestConvertRequestToDeepSeek(t *testing.T) {
	Convey("Given a upstream instance", t, func() {
		config := &conf.DeepSeekConfig{BaseUrl: "https://api.deepseek.com", ApiKey: "test-key"}
		repo := &upstream{
			config: config,
			log:    log.NewHelper(log.DefaultLogger),
		}

		Convey("When converting a basic chat request", func() {
			req := &entity.ChatReq{
				Id:    "req-1",
				Model: "deepseek-chat",
				Messages: []*v1.Message{
					{
						Id:   "msg-1",
						Role: v1.Role_USER,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Hello"}},
						},
					},
				},
			}

			result := repo.convertRequestToDeepSeek(req)

			Convey("Then the converted request should have correct model and messages", func() {
				So(result.Model, ShouldEqual, "deepseek-chat")
				So(len(result.Messages), ShouldEqual, 1)
				So(result.Messages[0].Role, ShouldEqual, "user")
				So(result.Messages[0].Content, ShouldEqual, "Hello")
			})
		})

		Convey("When converting a request with generation config", func() {
			req := &entity.ChatReq{
				Id:    "req-2",
				Model: "deepseek-chat",
				Config: &v1.GenerationConfig{
					MaxTokens:        1000,
					Temperature:      0.8,
					TopP:             0.9,
					FrequencyPenalty: 0.1,
					PresencePenalty:  0.2,
					Grammar:          &v1.GenerationConfig_PresetGrammar{PresetGrammar: "json_object"},
				},
				Messages: []*v1.Message{
					{
						Id:   "msg-1",
						Role: v1.Role_USER,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Generate JSON"}},
						},
					},
				},
			}

			result := repo.convertRequestToDeepSeek(req)

			Convey("Then the converted request should have correct generation config", func() {
				So(result.MaxTokens, ShouldEqual, 1000)
				So(result.Temperature, ShouldAlmostEqual, 0.8, 0.001)
				So(result.TopP, ShouldAlmostEqual, 0.9, 0.001)
				So(result.FrequencyPenalty, ShouldAlmostEqual, 0.1, 0.001)
				So(result.PresencePenalty, ShouldAlmostEqual, 0.2, 0.001)
				So(result.ResponseFormat, ShouldNotBeNil)
				So(result.ResponseFormat.Type, ShouldEqual, "json_object")
			})
		})

		Convey("When converting a request with tools", func() {
			req := &entity.ChatReq{
				Id:    "req-3",
				Model: "deepseek-chat",
				Tools: []*v1.Tool{
					{
						Tool: &v1.Tool_Function_{
							Function: &v1.Tool_Function{
								Name:        "get_weather",
								Description: "Get current weather",
								Parameters: &v1.Schema{
									Type: v1.Schema_TYPE_OBJECT,
									Properties: map[string]*v1.Schema{
										"location": {
											Type:        v1.Schema_TYPE_STRING,
											Description: "City name",
										},
									},
									Required: []string{"location"},
								},
							},
						},
					},
				},
				Messages: []*v1.Message{
					{
						Id:   "msg-1",
						Role: v1.Role_USER,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "What's the weather?"}},
						},
					},
				},
			}

			result := repo.convertRequestToDeepSeek(req)

			Convey("Then the converted request should have correct tools", func() {
				So(len(result.Tools), ShouldEqual, 1)
				So(result.Tools[0].Type, ShouldEqual, "function")
				So(result.Tools[0].Function.Name, ShouldEqual, "get_weather")
				So(result.Tools[0].Function.Description, ShouldEqual, "Get current weather")
				So(result.Tools[0].Function.Parameters, ShouldNotBeNil)
				j, err := json.Marshal(result.Tools[0].Function.Parameters)
				So(err, ShouldBeNil)
				So(string(j), ShouldContainSubstring, `"type":"object"`)
				So(string(j), ShouldContainSubstring, `"properties":{"location":{"type":"string","description":"City name"}}`)
				So(string(j), ShouldContainSubstring, `"required":["location"]`)
			})
		})
	})
}

func TestConvertMessageFromDeepSeek(t *testing.T) {
	Convey("Given a upstream instance", t, func() {
		config := &conf.DeepSeekConfig{BaseUrl: "https://api.deepseek.com", ApiKey: "test-key"}
		repo := &upstream{
			config: config,
			log:    log.NewHelper(log.DefaultLogger),
		}

		Convey("When converting a DeepSeek assistant message", func() {
			deepSeekMessage := &Message{
				Role:    "assistant",
				Content: "Hello! How can I help you today?",
			}

			result := repo.convertMessageFromDeepSeek("msg-1", deepSeekMessage)

			Convey("Then the converted message should have correct role and content", func() {
				So(result.Id, ShouldEqual, "msg-1")
				So(result.Role, ShouldEqual, v1.Role_MODEL)
				So(len(result.Contents), ShouldEqual, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "Hello! How can I help you today?")
			})
		})

		Convey("When converting a DeepSeek message with reasoning content", func() {
			deepSeekMessage := &Message{
				Role:             "assistant",
				Content:          "The answer is 42.",
				ReasoningContent: "Let me think about this step by step...",
			}

			result := repo.convertMessageFromDeepSeek("msg-2", deepSeekMessage)

			Convey("Then the converted message should have both thinking and text content", func() {
				So(result.Id, ShouldEqual, "msg-2")
				So(result.Role, ShouldEqual, v1.Role_MODEL)
				So(len(result.Contents), ShouldEqual, 2)
				So(result.Contents[0].GetThinking(), ShouldEqual, "Let me think about this step by step...")
				So(result.Contents[1].GetText(), ShouldEqual, "The answer is 42.")
			})
		})

		Convey("When converting a DeepSeek message with tool calls", func() {
			deepSeekMessage := &Message{
				Role: "assistant",
				ToolCalls: []*ToolCall{
					{
						ID:   "call-1",
						Type: "function",
						Function: &FunctionCall{
							Name:      "get_weather",
							Arguments: `{"location": "San Francisco"}`,
						},
					},
				},
			}

			result := repo.convertMessageFromDeepSeek("msg-3", deepSeekMessage)

			Convey("Then the converted message should have function call content", func() {
				So(result.Id, ShouldEqual, "msg-3")
				So(result.Role, ShouldEqual, v1.Role_MODEL)
				So(len(result.Contents), ShouldEqual, 1)
				funcCall := result.Contents[0].GetFunctionCall()
				So(funcCall, ShouldNotBeNil)
				So(funcCall.Id, ShouldEqual, "call-1")
				So(funcCall.Name, ShouldEqual, "get_weather")
				So(funcCall.Arguments, ShouldEqual, `{"location": "San Francisco"}`)
			})
		})

		Convey("When converting a DeepSeek tool message", func() {
			deepSeekMessage := &Message{
				Role:       "tool",
				Content:    "Weather in SF: sunny, 22°C",
				ToolCallID: "call-1",
			}

			result := repo.convertMessageFromDeepSeek("msg-4", deepSeekMessage)

			Convey("Then the converted message should have correct tool response", func() {
				So(result.Id, ShouldEqual, "msg-4")
				So(result.Role, ShouldEqual, v1.Role_TOOL)
				So(result.ToolCallId, ShouldEqual, "call-1")
				So(len(result.Contents), ShouldEqual, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "Weather in SF: sunny, 22°C")
			})
		})

		Convey("When converting a DeepSeek message with unsupported role", func() {
			deepSeekMessage := &Message{
				Role:    "unknown",
				Content: "This should not work",
			}

			result := repo.convertMessageFromDeepSeek("msg-5", deepSeekMessage)

			Convey("Then the result should be nil", func() {
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestConvertStreamRespFromDeepSeek(t *testing.T) {
	Convey("Given a stream response chunk from DeepSeek", t, func() {
		Convey("When converting a chunk with delta content", func() {
			chunk := &ChatStreamResponse{
				ID:    "chatcmpl-123",
				Model: "deepseek-chat",
				Choices: []*ChatStreamChoice{
					{
						Delta: &Message{
							Content: "Hello",
						},
					},
				},
			}

			result := convertStreamRespFromDeepSeek("req-1", chunk)

			Convey("Then the converted response should have correct structure", func() {
				So(result.Id, ShouldEqual, "req-1")
				So(result.Model, ShouldEqual, "deepseek-chat")
				So(result.Message, ShouldNotBeNil)
				So(result.Message.Id, ShouldEqual, "chatcmpl-123")
				So(result.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(len(result.Message.Contents), ShouldEqual, 1)
				So(result.Message.Contents[0].GetText(), ShouldEqual, "Hello")
			})
		})

		Convey("When converting a chunk with reasoning content", func() {
			chunk := &ChatStreamResponse{
				ID:    "chatcmpl-456",
				Model: "deepseek-reasoner",
				Choices: []*ChatStreamChoice{
					{
						Delta: &Message{
							ReasoningContent: "Thinking...",
							Content:          " world!",
						},
					},
				},
			}

			result := convertStreamRespFromDeepSeek("req-2", chunk)

			Convey("Then the converted response should have both thinking and text content", func() {
				So(result.Id, ShouldEqual, "req-2")
				So(result.Model, ShouldEqual, "deepseek-reasoner")
				So(result.Message, ShouldNotBeNil)
				So(len(result.Message.Contents), ShouldEqual, 2)
				So(result.Message.Contents[0].GetThinking(), ShouldEqual, "Thinking...")
				So(result.Message.Contents[1].GetText(), ShouldEqual, " world!")
			})
		})

		Convey("When converting a chunk with tool calls", func() {
			chunk := &ChatStreamResponse{
				ID:    "chatcmpl-tools",
				Model: "deepseek-chat",
				Choices: []*ChatStreamChoice{
					{
						Delta: &Message{
							ToolCalls: []*ToolCall{
								{
									ID:   "call-1",
									Type: "function",
									Function: &FunctionCall{
										Name:      "get_time",
										Arguments: `{"tz":"UTC"}`,
									},
								},
							},
						},
					},
				},
			}

			result := convertStreamRespFromDeepSeek("req-tools", chunk)

			Convey("Then the converted response should include function call content", func() {
				So(result.Id, ShouldEqual, "req-tools")
				So(result.Model, ShouldEqual, "deepseek-chat")
				So(result.Message, ShouldNotBeNil)
				So(result.Message.Id, ShouldEqual, "chatcmpl-tools")
				So(result.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(len(result.Message.Contents), ShouldEqual, 1)
				fc := result.Message.Contents[0].GetFunctionCall()
				So(fc, ShouldNotBeNil)
				So(fc.Id, ShouldEqual, "call-1")
				So(fc.Name, ShouldEqual, "get_time")
				So(fc.Arguments, ShouldEqual, `{"tz":"UTC"}`)
			})
		})

		Convey("When converting a chunk with usage statistics", func() {
			chunk := &ChatStreamResponse{
				ID:    "chatcmpl-789",
				Model: "deepseek-chat",
				Choices: []*ChatStreamChoice{
					{
						Delta: &Message{
							Content: "Done",
						},
					},
				},
				Usage: &Usage{
					PromptTokens:     10,
					CompletionTokens: 20,
				},
			}

			result := convertStreamRespFromDeepSeek("req-3", chunk)

			Convey("Then the converted response should have usage statistics", func() {
				So(result.Id, ShouldEqual, "req-3")
				So(result.Model, ShouldEqual, "deepseek-chat")
				So(result.Message, ShouldNotBeNil)
				So(len(result.Message.Contents), ShouldEqual, 1)
				So(result.Message.Contents[0].GetText(), ShouldEqual, "Done")
				So(result.Statistics, ShouldNotBeNil)
				So(result.Statistics.Usage.PromptTokens, ShouldEqual, 10)
				So(result.Statistics.Usage.CompletionTokens, ShouldEqual, 20)
			})
		})
	})
}

func TestConvertStatisticsFromDeepSeek(t *testing.T) {
	Convey("Given DeepSeek usage statistics", t, func() {
		Convey("When converting nil usage", func() {
			result := convertStatisticsFromDeepSeek(nil)

			Convey("Then the result should be nil", func() {
				So(result, ShouldBeNil)
			})
		})

		Convey("When converting valid usage statistics", func() {
			usage := &Usage{
				PromptTokens:         15,
				CompletionTokens:     25,
				PromptCacheHitTokens: 5,
			}

			result := convertStatisticsFromDeepSeek(usage)

			Convey("Then the converted statistics should have correct values", func() {
				So(result, ShouldNotBeNil)
				So(result.Usage, ShouldNotBeNil)
				So(result.Usage.PromptTokens, ShouldEqual, 15)
				So(result.Usage.CompletionTokens, ShouldEqual, 25)
				So(result.Usage.CachedPromptTokens, ShouldEqual, 5)
			})
		})
	})
}

func TestToolFunctionParametersToDeepSeek(t *testing.T) {
	Convey("Given a function parameter schema", t, func() {
		Convey("When converting to DeepSeek format", func() {
			params := &v1.Schema{
				Type: v1.Schema_TYPE_OBJECT,
				Properties: map[string]*v1.Schema{
					"name": {
						Type:        v1.Schema_TYPE_STRING,
						Description: "The name parameter",
					},
					"age": {
						Type:        v1.Schema_TYPE_INTEGER,
						Description: "The age parameter",
					},
				},
				Required: []string{"name"},
			}

			result := toolFunctionParametersToDeepSeek(params)

			Convey("Then the result should have correct structure", func() {
				So(result, ShouldNotBeNil)
				So(result["type"], ShouldEqual, params.Type)
				So(result["properties"], ShouldEqual, params.Properties)
				So(result["required"], ShouldResemble, params.Required)
			})
		})
	})
}
