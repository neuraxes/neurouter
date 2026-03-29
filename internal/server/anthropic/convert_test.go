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

package anthropic

import (
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestConvertGenerationConfigFromAnthropic(t *testing.T) {
	Convey("Given an Anthropic request to convert generation config from", t, func() {

		Convey("When MaxTokens is set", func() {
			req := &anthropic.MessageNewParams{MaxTokens: 4096}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then MaxTokens should be applied", func() {
				So(config.MaxTokens, ShouldNotBeNil)
				So(*config.MaxTokens, ShouldEqual, 4096)
			})
		})

		Convey("When MaxTokens is zero", func() {
			req := &anthropic.MessageNewParams{MaxTokens: 0}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then MaxTokens should be nil", func() {
				So(config.MaxTokens, ShouldBeNil)
			})
		})

		Convey("When Temperature is set", func() {
			req := &anthropic.MessageNewParams{Temperature: anthropic.Opt(0.75)}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then Temperature should be applied", func() {
				So(config.Temperature, ShouldNotBeNil)
				So(*config.Temperature, ShouldAlmostEqual, 0.75, 0.01)
			})
		})

		Convey("When TopP is set", func() {
			req := &anthropic.MessageNewParams{TopP: anthropic.Opt(0.95)}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then TopP should be applied", func() {
				So(config.TopP, ShouldNotBeNil)
				So(*config.TopP, ShouldAlmostEqual, 0.95, 0.01)
			})
		})

		Convey("When TopK is set", func() {
			req := &anthropic.MessageNewParams{TopK: anthropic.Opt[int64](50)}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then TopK should be applied", func() {
				So(config.TopK, ShouldNotBeNil)
				So(*config.TopK, ShouldEqual, 50)
			})
		})

		Convey("When Thinking is enabled", func() {
			req := &anthropic.MessageNewParams{
				Thinking: anthropic.ThinkingConfigParamUnion{
					OfEnabled: &anthropic.ThinkingConfigEnabledParam{
						BudgetTokens: 2048,
					},
				},
			}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then ReasoningConfig should be set", func() {
				So(config.ReasoningConfig, ShouldNotBeNil)
				So(config.ReasoningConfig.Enabled, ShouldBeTrue)
				So(config.ReasoningConfig.TokenBudget, ShouldEqual, 2048)
			})
		})

		Convey("When Thinking is not enabled", func() {
			req := &anthropic.MessageNewParams{}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then ReasoningConfig should be nil", func() {
				So(config.ReasoningConfig, ShouldBeNil)
			})
		})

		Convey("When Thinking is disabled", func() {
			req := &anthropic.MessageNewParams{
				Thinking: anthropic.ThinkingConfigParamUnion{
					OfDisabled: &anthropic.ThinkingConfigDisabledParam{},
				},
			}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then ReasoningConfig should be set with Enabled false", func() {
				So(config.ReasoningConfig, ShouldNotBeNil)
				So(config.ReasoningConfig.Enabled, ShouldBeFalse)
			})
		})

		Convey("When Thinking is adaptive", func() {
			req := &anthropic.MessageNewParams{
				Thinking: anthropic.ThinkingConfigParamUnion{
					OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
				},
			}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then ReasoningConfig should be set with Enabled true", func() {
				So(config.ReasoningConfig, ShouldNotBeNil)
				So(config.ReasoningConfig.Enabled, ShouldBeTrue)
			})
		})

		Convey("When all parameters are set", func() {
			req := &anthropic.MessageNewParams{
				MaxTokens:   8192,
				Temperature: anthropic.Opt(0.5),
				TopP:        anthropic.Opt(0.9),
				TopK:        anthropic.Opt[int64](40),
				Thinking: anthropic.ThinkingConfigParamUnion{
					OfEnabled: &anthropic.ThinkingConfigEnabledParam{
						BudgetTokens: 1024,
					},
				},
			}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then all fields should be populated", func() {
				So(*config.MaxTokens, ShouldEqual, 8192)
				So(*config.Temperature, ShouldAlmostEqual, 0.5, 0.01)
				So(*config.TopP, ShouldAlmostEqual, 0.9, 0.01)
				So(*config.TopK, ShouldEqual, 40)
				So(config.ReasoningConfig.Enabled, ShouldBeTrue)
				So(config.ReasoningConfig.TokenBudget, ShouldEqual, 1024)
			})
		})

		Convey("When OutputConfig has JSON schema format", func() {
			req := &anthropic.MessageNewParams{
				OutputConfig: anthropic.OutputConfigParam{
					Format: anthropic.JSONOutputFormatParam{
						Schema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{"type": "string"},
								"age":  map[string]any{"type": "integer"},
							},
							"required": []string{"name"},
						},
					},
				},
			}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then Grammar should be set as JsonSchema", func() {
				So(config.Grammar, ShouldNotBeNil)
				jsonSchema, ok := config.Grammar.(*v1.GenerationConfig_JsonSchema)
				So(ok, ShouldBeTrue)
				So(jsonSchema.JsonSchema, ShouldNotBeEmpty)
				So(jsonSchema.JsonSchema, ShouldContainSubstring, `"type":"object"`)
				So(jsonSchema.JsonSchema, ShouldContainSubstring, `"name":{"type":"string"}`)
			})
		})

		Convey("When OutputConfig has empty schema", func() {
			req := &anthropic.MessageNewParams{
				OutputConfig: anthropic.OutputConfigParam{
					Format: anthropic.JSONOutputFormatParam{
						Schema: map[string]any{},
					},
				},
			}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then Grammar should remain nil", func() {
				So(config.Grammar, ShouldBeNil)
			})
		})

		Convey("When OutputConfig has no schema", func() {
			req := &anthropic.MessageNewParams{}
			config := convertGenerationConfigFromAnthropic(req)

			Convey("Then Grammar should be nil", func() {
				So(config.Grammar, ShouldBeNil)
			})
		})
	})
}

func TestConvertSystemFromAnthropic(t *testing.T) {
	Convey("Given Anthropic system blocks", t, func() {

		Convey("When system blocks are empty", func() {
			result := convertSystemFromAnthropic(nil)

			Convey("Then result should be nil", func() {
				So(result, ShouldBeNil)
			})
		})

		Convey("When system blocks contain text", func() {
			system := []anthropic.TextBlockParam{
				{Text: "You are a helpful assistant."},
				{Text: "Be concise."},
			}
			result := convertSystemFromAnthropic(system)

			Convey("Then a system message should be created with all texts", func() {
				So(result, ShouldNotBeNil)
				So(result.Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Contents, ShouldHaveLength, 2)
				So(result.Contents[0].GetText(), ShouldEqual, "You are a helpful assistant.")
				So(result.Contents[1].GetText(), ShouldEqual, "Be concise.")
			})
		})
	})
}

func TestConvertMessageFromAnthropic(t *testing.T) {
	Convey("Given an Anthropic message to convert", t, func() {

		Convey("When converting a user message with text", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewTextBlock("Hello!"),
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create a USER message with text", func() {
				So(result.Role, ShouldEqual, v1.Role_USER)
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "Hello!")
			})
		})

		Convey("When converting an assistant message with text", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleAssistant,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewTextBlock("I'm here to help!"),
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create a MODEL message", func() {
				So(result.Role, ShouldEqual, v1.Role_MODEL)
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "I'm here to help!")
			})
		})

		Convey("When converting a message with image URL", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewImageBlock(anthropic.URLImageSourceParam{
						URL: "https://example.com/image.jpg",
					}),
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create an image content with URL", func() {
				So(result.Contents, ShouldHaveLength, 1)
				img := result.Contents[0].GetImage()
				So(img, ShouldNotBeNil)
				urlSrc, ok := img.Source.(*v1.Image_Url)
				So(ok, ShouldBeTrue)
				So(urlSrc.Url, ShouldEqual, "https://example.com/image.jpg")
			})
		})

		Convey("When converting a message with base64 image", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewImageBlockBase64("image/png", "aW1hZ2VkYXRh"),
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create an image content with decoded data", func() {
				So(result.Contents, ShouldHaveLength, 1)
				img := result.Contents[0].GetImage()
				So(img, ShouldNotBeNil)
				So(img.MimeType, ShouldEqual, "image/png")
				dataSrc, ok := img.Source.(*v1.Image_Data)
				So(ok, ShouldBeTrue)
				So(string(dataSrc.Data), ShouldEqual, "imagedata")
			})
		})

		Convey("When converting a message with thinking content", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleAssistant,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewThinkingBlock("sig-abc", "Let me think about this..."),
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create a reasoning content with signature", func() {
				So(result.Role, ShouldEqual, v1.Role_MODEL)
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].Reasoning, ShouldBeTrue)
				So(result.Contents[0].GetText(), ShouldEqual, "Let me think about this...")
				So(result.Contents[0].Metadata["signature"], ShouldEqual, "sig-abc")
			})
		})

		Convey("When converting a message with redacted thinking content", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleAssistant,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewRedactedThinkingBlock("opaque-encrypted-data"),
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create a redacted thinking content", func() {
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].Reasoning, ShouldBeTrue)
				So(result.Contents[0].GetText(), ShouldEqual, "")
				So(result.Contents[0].Metadata["redacted_thinking"], ShouldEqual, "opaque-encrypted-data")
			})
		})

		Convey("When converting a message with tool_use", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleAssistant,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewToolUseBlock("tool-1", map[string]any{"city": "Shanghai"}, "get_weather"),
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create a ToolUse content", func() {
				So(result.Contents, ShouldHaveLength, 1)
				tu := result.Contents[0].GetToolUse()
				So(tu, ShouldNotBeNil)
				So(tu.Id, ShouldEqual, "tool-1")
				So(tu.Name, ShouldEqual, "get_weather")
				So(tu.GetTextualInput(), ShouldEqual, `{"city":"Shanghai"}`)
			})
		})

		Convey("When converting a message with tool_result", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					{
						OfToolResult: &anthropic.ToolResultBlockParam{
							ToolUseID: "tool-1",
							Content: []anthropic.ToolResultBlockParamContentUnion{
								{
									OfText: &anthropic.TextBlockParam{
										Text: `Result: `,
									},
								},
								{
									OfText: &anthropic.TextBlockParam{
										Text: `{"result": "sunny"}`,
									},
								},
							},
						},
					},
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create a ToolResult content", func() {
				So(result.Contents, ShouldHaveLength, 1)
				tr := result.Contents[0].GetToolResult()
				So(tr, ShouldNotBeNil)
				So(tr.Id, ShouldEqual, "tool-1")
				So(tr.Outputs, ShouldHaveLength, 2)
				So(tr.Outputs[0].GetText(), ShouldEqual, `Result: `)
				So(tr.Outputs[1].GetText(), ShouldEqual, `{"result": "sunny"}`)
			})
		})

		Convey("When converting a message with tool_result image URL", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					{
						OfToolResult: &anthropic.ToolResultBlockParam{
							ToolUseID: "tool-1",
							Content: []anthropic.ToolResultBlockParamContentUnion{
								{
									OfImage: &anthropic.ImageBlockParam{
										Source: anthropic.ImageBlockParamSourceUnion{
											OfURL: &anthropic.URLImageSourceParam{
												URL: "https://example.com/tool-result.png",
											},
										},
									},
								},
							},
						},
					},
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create a ToolResult image URL output", func() {
				So(result.Contents, ShouldHaveLength, 1)
				tr := result.Contents[0].GetToolResult()
				So(tr, ShouldNotBeNil)
				So(tr.Outputs, ShouldHaveLength, 1)
				img := tr.Outputs[0].GetImage()
				So(img, ShouldNotBeNil)
				urlSrc, ok := img.Source.(*v1.Image_Url)
				So(ok, ShouldBeTrue)
				So(urlSrc.Url, ShouldEqual, "https://example.com/tool-result.png")
			})
		})

		Convey("When converting a message with tool_result image", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					{
						OfToolResult: &anthropic.ToolResultBlockParam{
							ToolUseID: "tool-1",
							Content: []anthropic.ToolResultBlockParamContentUnion{
								{
									OfImage: &anthropic.ImageBlockParam{
										Source: anthropic.ImageBlockParamSourceUnion{
											OfBase64: &anthropic.Base64ImageSourceParam{
												Data:      "aW1hZ2VkYXRh",
												MediaType: anthropic.Base64ImageSourceMediaTypeImagePNG,
											},
										},
									},
								},
							},
						},
					},
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then it should create a ToolResult image output", func() {
				So(result.Contents, ShouldHaveLength, 1)
				tr := result.Contents[0].GetToolResult()
				So(tr, ShouldNotBeNil)
				So(tr.Outputs, ShouldHaveLength, 1)
				img := tr.Outputs[0].GetImage()
				So(img, ShouldNotBeNil)
				So(img.MimeType, ShouldEqual, "image/png")
				dataSrc, ok := img.Source.(*v1.Image_Data)
				So(ok, ShouldBeTrue)
				So(string(dataSrc.Data), ShouldEqual, "imagedata")
			})
		})

		Convey("When converting a message with mixed thinking, text, and tool_use", func() {
			msg := &anthropic.MessageParam{
				Role: anthropic.MessageParamRoleAssistant,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewThinkingBlock("sig-1", "thinking..."),
					anthropic.NewRedactedThinkingBlock("opaque-data"),
					anthropic.NewTextBlock("Here is the answer"),
					anthropic.NewToolUseBlock("tool-2", map[string]any{"k": "v"}, "search"),
				},
			}
			result := convertMessageFromAnthropic(msg)

			Convey("Then all content types should be correctly converted", func() {
				So(result.Contents, ShouldHaveLength, 4)

				// thinking
				So(result.Contents[0].Reasoning, ShouldBeTrue)
				So(result.Contents[0].GetText(), ShouldEqual, "thinking...")
				So(result.Contents[0].Metadata["signature"], ShouldEqual, "sig-1")

				// redacted thinking
				So(result.Contents[1].Reasoning, ShouldBeTrue)
				So(result.Contents[1].GetText(), ShouldEqual, "")
				So(result.Contents[1].Metadata["redacted_thinking"], ShouldEqual, "opaque-data")

				// text
				So(result.Contents[2].Reasoning, ShouldBeFalse)
				So(result.Contents[2].GetText(), ShouldEqual, "Here is the answer")

				// tool_use
				tu := result.Contents[3].GetToolUse()
				So(tu, ShouldNotBeNil)
				So(tu.Id, ShouldEqual, "tool-2")
				So(tu.Name, ShouldEqual, "search")
				So(tu.GetTextualInput(), ShouldEqual, `{"k":"v"}`)
			})
		})
	})
}

func TestConvertChatReqFromAnthropic(t *testing.T) {
	Convey("Given an Anthropic MessageNewParams request", t, func() {

		Convey("When converting a complete request with system, messages, tools", func() {
			req := &anthropic.MessageNewParams{
				Model:     anthropic.Model("claude-3-sonnet"),
				MaxTokens: 4096,
				System: []anthropic.TextBlockParam{
					{Text: "You are helpful."},
				},
				Messages: []anthropic.MessageParam{
					anthropic.NewUserMessage(anthropic.NewTextBlock("Hello")),
					anthropic.NewAssistantMessage(anthropic.NewTextBlock("Hi there")),
				},
				Tools: []anthropic.ToolUnionParam{
					{
						OfTool: &anthropic.ToolParam{
							Name:        "get_weather",
							Description: anthropic.Opt("Get weather info"),
							InputSchema: anthropic.ToolInputSchemaParam{
								Properties: map[string]any{
									"city": map[string]any{
										"type":        "string",
										"description": "The name of city",
									},
								},
								Required: []string{"city"},
							},
						},
					},
				},
				Thinking: anthropic.ThinkingConfigParamUnion{
					OfEnabled: &anthropic.ThinkingConfigEnabledParam{
						BudgetTokens: 1024,
					},
				},
			}

			result := convertChatReqFromAnthropic(req)

			Convey("Then the model should be set", func() {
				So(result.Model, ShouldEqual, "claude-3-sonnet")
			})

			Convey("Then config should include reasoning", func() {
				So(result.Config, ShouldNotBeNil)
				So(*result.Config.MaxTokens, ShouldEqual, 4096)
				So(result.Config.ReasoningConfig, ShouldNotBeNil)
				So(result.Config.ReasoningConfig.Enabled, ShouldBeTrue)
				So(result.Config.ReasoningConfig.TokenBudget, ShouldEqual, 1024)
			})

			Convey("Then system message should be included", func() {
				So(result.Messages, ShouldHaveLength, 3) // system + 2 messages
				So(result.Messages[0].Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Messages[0].Contents[0].GetText(), ShouldEqual, "You are helpful.")
			})

			Convey("Then user and assistant messages should be converted", func() {
				So(result.Messages[1].Role, ShouldEqual, v1.Role_USER)
				So(result.Messages[1].Contents[0].GetText(), ShouldEqual, "Hello")
				So(result.Messages[2].Role, ShouldEqual, v1.Role_MODEL)
				So(result.Messages[2].Contents[0].GetText(), ShouldEqual, "Hi there")
			})

			Convey("Then tools should be converted", func() {
				So(result.Tools, ShouldHaveLength, 1)
				fn := result.Tools[0].GetFunction()
				So(fn, ShouldNotBeNil)
				So(fn.Name, ShouldEqual, "get_weather")
				So(fn.Description, ShouldEqual, "Get weather info")
				So(fn.Parameters, ShouldNotBeNil)
				So(fn.Parameters.Properties, ShouldContainKey, "city")
				So(fn.Parameters.Properties["city"].Type, ShouldEqual, v1.Schema_TYPE_STRING)
				So(fn.Parameters.Properties["city"].Description, ShouldEqual, "The name of city")
				So(fn.Parameters.Required, ShouldResemble, []string{"city"})
			})
		})

		Convey("When converting a request without system messages", func() {
			req := &anthropic.MessageNewParams{
				Model:     anthropic.Model("claude-3"),
				MaxTokens: 1024,
				Messages: []anthropic.MessageParam{
					anthropic.NewUserMessage(anthropic.NewTextBlock("Hello")),
				},
			}

			result := convertChatReqFromAnthropic(req)

			Convey("Then messages should not include a system message", func() {
				So(result.Messages, ShouldHaveLength, 1)
				So(result.Messages[0].Role, ShouldEqual, v1.Role_USER)
				So(result.Messages[0].Contents[0].GetText(), ShouldEqual, "Hello")
			})
		})

		Convey("When converting a request without tools", func() {
			req := &anthropic.MessageNewParams{
				Model:     anthropic.Model("claude-3"),
				MaxTokens: 1024,
				Messages: []anthropic.MessageParam{
					anthropic.NewUserMessage(anthropic.NewTextBlock("Hi")),
				},
			}

			result := convertChatReqFromAnthropic(req)

			Convey("Then tools should be nil", func() {
				So(result.Tools, ShouldBeNil)
			})
		})
	})
}

func TestConvertStatusToAnthropic(t *testing.T) {
	Convey("Given various internal chat statuses", t, func() {
		So(convertStatusToAnthropic(v1.ChatStatus_CHAT_IN_PROGRESS), ShouldEqual, anthropic.StopReasonEndTurn)
		So(convertStatusToAnthropic(v1.ChatStatus_CHAT_COMPLETED), ShouldEqual, anthropic.StopReasonEndTurn)
		So(convertStatusToAnthropic(v1.ChatStatus_CHAT_REFUSED), ShouldEqual, anthropic.StopReasonRefusal)
		So(convertStatusToAnthropic(v1.ChatStatus_CHAT_CANCELLED), ShouldEqual, anthropic.StopReasonEndTurn)
		So(convertStatusToAnthropic(v1.ChatStatus_CHAT_PENDING_TOOL_USE), ShouldEqual, anthropic.StopReasonToolUse)
		So(convertStatusToAnthropic(v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT), ShouldEqual, anthropic.StopReasonMaxTokens)
	})
}

func TestConvertChatRespToAnthropic(t *testing.T) {
	Convey("Given an internal ChatResp to convert to Anthropic format", t, func() {

		Convey("When converting a response with text content", func() {
			resp := &v1.ChatResp{
				Model:  "claude-3-sonnet",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Id:   "msg-1",
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: "Hello!"}},
					},
				},
				Statistics: &v1.Statistics{
					Usage: &v1.Statistics_Usage{
						InputTokens:       10,
						OutputTokens:      5,
						CachedInputTokens: 3,
					},
				},
			}

			result := convertChatRespToAnthropic(resp)

			Convey("Then the response should have correct fields", func() {
				So(string(result.Type), ShouldEqual, "message")
				So(result.ID, ShouldEqual, "msg-1")
				So(string(result.Model), ShouldEqual, "claude-3-sonnet")
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.StopReason, ShouldEqual, anthropic.StopReasonEndTurn)
				So(result.Content, ShouldHaveLength, 1)
				So(result.Content[0].Type, ShouldEqual, "text")
				So(result.Content[0].Text, ShouldEqual, "Hello!")
				So(result.Usage.InputTokens, ShouldEqual, 10)
				So(result.Usage.OutputTokens, ShouldEqual, 5)
				So(result.Usage.CacheReadInputTokens, ShouldEqual, 3)
			})
		})

		Convey("When converting a response with thinking content and signature", func() {
			resp := &v1.ChatResp{
				Model:  "claude-3-sonnet",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Id:   "msg-2",
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Reasoning: true,
							Metadata:  map[string]string{"signature": "sig-abc"},
							Content:   &v1.Content_Text{Text: "Let me think..."},
						},
						{
							Content: &v1.Content_Text{Text: "The answer is 42."},
						},
					},
				},
			}

			result := convertChatRespToAnthropic(resp)

			Convey("Then thinking content should have signature", func() {
				So(string(result.Type), ShouldEqual, "message")
				So(result.ID, ShouldEqual, "msg-2")
				So(string(result.Model), ShouldEqual, "claude-3-sonnet")
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.StopReason, ShouldEqual, anthropic.StopReasonEndTurn)
				So(result.Content, ShouldHaveLength, 2)
				So(result.Content[0].Type, ShouldEqual, "thinking")
				So(result.Content[0].Thinking, ShouldEqual, "Let me think...")
				So(result.Content[0].Signature, ShouldEqual, "sig-abc")
				So(result.Content[1].Type, ShouldEqual, "text")
				So(result.Content[1].Text, ShouldEqual, "The answer is 42.")
			})
		})

		Convey("When converting a response with redacted thinking", func() {
			resp := &v1.ChatResp{
				Model:  "claude-3-sonnet",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Id:   "msg-3",
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Reasoning: true,
							Metadata:  map[string]string{"redacted_thinking": "opaque-data"},
							Content:   &v1.Content_Text{Text: ""},
						},
						{
							Content: &v1.Content_Text{Text: "Result here."},
						},
					},
				},
			}

			result := convertChatRespToAnthropic(resp)

			Convey("Then redacted thinking should be emitted correctly", func() {
				So(string(result.Type), ShouldEqual, "message")
				So(result.ID, ShouldEqual, "msg-3")
				So(string(result.Model), ShouldEqual, "claude-3-sonnet")
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.StopReason, ShouldEqual, anthropic.StopReasonEndTurn)
				So(result.Content, ShouldHaveLength, 2)
				So(result.Content[0].Type, ShouldEqual, "redacted_thinking")
				So(result.Content[0].Data, ShouldEqual, "opaque-data")
				So(result.Content[1].Type, ShouldEqual, "text")
				So(result.Content[1].Text, ShouldEqual, "Result here.")
			})
		})

		Convey("When converting a response with tool_use", func() {
			resp := &v1.ChatResp{
				Model:  "claude-3-sonnet",
				Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
				Message: &v1.Message{
					Id:   "msg-4",
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Content: &v1.Content_ToolUse{
								ToolUse: &v1.ToolUse{
									Id:   "tool-1",
									Name: "get_weather",
									Inputs: []*v1.ToolUse_Input{{
										Input: &v1.ToolUse_Input_Text{Text: `{"city":"Shanghai"}`},
									}},
								},
							},
						},
					},
				},
			}

			result := convertChatRespToAnthropic(resp)

			Convey("Then tool_use should be correctly converted", func() {
				So(string(result.Type), ShouldEqual, "message")
				So(result.ID, ShouldEqual, "msg-4")
				So(string(result.Model), ShouldEqual, "claude-3-sonnet")
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.StopReason, ShouldEqual, anthropic.StopReasonToolUse)
				So(result.Content, ShouldHaveLength, 1)
				So(result.Content[0].Type, ShouldEqual, "tool_use")
				So(result.Content[0].ID, ShouldEqual, "tool-1")
				So(result.Content[0].Name, ShouldEqual, "get_weather")
				So(string(result.Content[0].Input), ShouldEqual, `{"city":"Shanghai"}`)
			})
		})

		Convey("When converting a response with mixed thinking, redacted_thinking, text, and tool_use", func() {
			resp := &v1.ChatResp{
				Model:  "claude-3-sonnet",
				Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
				Message: &v1.Message{
					Id:   "msg-5",
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Reasoning: true,
							Metadata:  map[string]string{"signature": "sig-1"},
							Content:   &v1.Content_Text{Text: "thinking content"},
						},
						{
							Reasoning: true,
							Metadata:  map[string]string{"redacted_thinking": "secret-data"},
							Content:   &v1.Content_Text{Text: ""},
						},
						{
							Content: &v1.Content_Text{Text: "Let me help you."},
						},
						{
							Content: &v1.Content_ToolUse{
								ToolUse: &v1.ToolUse{
									Id:   "tool-2",
									Name: "search",
									Inputs: []*v1.ToolUse_Input{{
										Input: &v1.ToolUse_Input_Text{Text: `{"k":"v"}`},
									}},
								},
							},
						},
					},
				},
				Statistics: &v1.Statistics{
					Usage: &v1.Statistics_Usage{
						InputTokens:  100,
						OutputTokens: 50,
					},
				},
			}

			result := convertChatRespToAnthropic(resp)

			Convey("Then all content types should be correctly converted", func() {
				So(string(result.Type), ShouldEqual, "message")
				So(result.ID, ShouldEqual, "msg-5")
				So(string(result.Model), ShouldEqual, "claude-3-sonnet")
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.StopReason, ShouldEqual, anthropic.StopReasonToolUse)
				So(result.Content, ShouldHaveLength, 4)

				// thinking
				So(result.Content[0].Type, ShouldEqual, "thinking")
				So(result.Content[0].Thinking, ShouldEqual, "thinking content")
				So(result.Content[0].Signature, ShouldEqual, "sig-1")

				// redacted thinking
				So(result.Content[1].Type, ShouldEqual, "redacted_thinking")
				So(result.Content[1].Data, ShouldEqual, "secret-data")

				// text
				So(result.Content[2].Type, ShouldEqual, "text")
				So(result.Content[2].Text, ShouldEqual, "Let me help you.")

				// tool_use
				So(result.Content[3].Type, ShouldEqual, "tool_use")
				So(result.Content[3].ID, ShouldEqual, "tool-2")
				So(result.Content[3].Name, ShouldEqual, "search")
				So(string(result.Content[3].Input), ShouldEqual, `{"k":"v"}`)
			})
		})

		Convey("When empty text content should be skipped for non-reasoning", func() {
			resp := &v1.ChatResp{
				Model:  "claude-3",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Id:   "msg-6",
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: ""}},
						{Content: &v1.Content_Text{Text: "actual content"}},
					},
				},
			}

			result := convertChatRespToAnthropic(resp)

			Convey("Then empty text blocks should be skipped", func() {
				So(string(result.Type), ShouldEqual, "message")
				So(result.ID, ShouldEqual, "msg-6")
				So(string(result.Model), ShouldEqual, "claude-3")
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.StopReason, ShouldEqual, anthropic.StopReasonEndTurn)
				So(result.Content, ShouldHaveLength, 1)
				So(result.Content[0].Text, ShouldEqual, "actual content")
			})
		})
	})
}

func TestConvertStatisticsToAnthropic(t *testing.T) {
	Convey("Given statistics to convert", t, func() {
		stats := &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens:       100,
				OutputTokens:      50,
				CachedInputTokens: 20,
			},
		}

		result := convertStatisticsToAnthropic(stats)

		Convey("Then all usage fields should be mapped", func() {
			So(result.InputTokens, ShouldEqual, 100)
			So(result.OutputTokens, ShouldEqual, 50)
			So(result.CacheReadInputTokens, ShouldEqual, 20)
		})
	})
}
