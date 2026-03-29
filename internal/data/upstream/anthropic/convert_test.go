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
	"encoding/json"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
)

func TestConvertStatusFromAnthropic(t *testing.T) {
	Convey("Given various anthropic stop reasons", t, func() {
		So(convertStatusFromAnthropic(anthropic.StopReasonToolUse), ShouldEqual, v1.ChatStatus_CHAT_PENDING_TOOL_USE)
		So(convertStatusFromAnthropic(anthropic.StopReasonEndTurn), ShouldEqual, v1.ChatStatus_CHAT_COMPLETED)
		So(convertStatusFromAnthropic(anthropic.StopReasonStopSequence), ShouldEqual, v1.ChatStatus_CHAT_COMPLETED)
		So(convertStatusFromAnthropic(anthropic.StopReasonMaxTokens), ShouldEqual, v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT)
		So(convertStatusFromAnthropic(anthropic.StopReasonRefusal), ShouldEqual, v1.ChatStatus_CHAT_REFUSED)
		So(convertStatusFromAnthropic(anthropic.StopReasonPauseTurn), ShouldEqual, v1.ChatStatus_CHAT_IN_PROGRESS)
		So(convertStatusFromAnthropic(anthropic.StopReason("")), ShouldEqual, v1.ChatStatus_CHAT_IN_PROGRESS)
		So(convertStatusFromAnthropic(anthropic.StopReason("unknown_reason")), ShouldEqual, v1.ChatStatus_CHAT_IN_PROGRESS)
	})
}

func TestConvertGenerationConfigToAnthropic(t *testing.T) {
	Convey("Given a generation config converter", t, func() {
		repo := &upstream{
			config: &conf.AnthropicConfig{},
			log:    log.NewHelper(log.DefaultLogger),
		}

		Convey("When config is nil", func() {
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(nil, req)

			Convey("Then maxTokens should not be set", func() {
				So(req.MaxTokens, ShouldEqual, 0)
			})
		})

		Convey("When config has maxTokens set", func() {
			config := &v1.GenerationConfig{MaxTokens: new(int64(4096))}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then maxTokens should be applied", func() {
				So(req.MaxTokens, ShouldEqual, 4096)
			})
		})

		Convey("When config has no maxTokens or zero maxTokens", func() {
			config := &v1.GenerationConfig{}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then default maxTokens 8192 should be used", func() {
				So(req.MaxTokens, ShouldEqual, 8192)
			})
		})

		Convey("When config has temperature", func() {
			config := &v1.GenerationConfig{Temperature: new(float32(0.756))}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then temperature should be rounded to 2 decimal places", func() {
				So(req.Temperature.Value, ShouldEqual, 0.76)
			})
		})

		Convey("When config has topP", func() {
			config := &v1.GenerationConfig{TopP: new(float32(0.956))}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then topP should be rounded to 2 decimal places", func() {
				So(req.TopP.Value, ShouldEqual, 0.96)
			})
		})

		Convey("When config has topK", func() {
			config := &v1.GenerationConfig{TopK: new(int64(50))}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then topK should be applied", func() {
				So(req.TopK.Value, ShouldEqual, 50)
			})
		})

		Convey("When config has reasoning enabled with default budget", func() {
			config := &v1.GenerationConfig{
				ReasoningConfig: &v1.ReasoningConfig{
					Enabled: true,
				},
			}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then thinking should be enabled with default budget 1024", func() {
				So(req.Thinking.OfEnabled, ShouldNotBeNil)
				So(req.Thinking.OfEnabled.BudgetTokens, ShouldEqual, 1024)
			})
		})

		Convey("When config has reasoning enabled with custom budget", func() {
			config := &v1.GenerationConfig{
				ReasoningConfig: &v1.ReasoningConfig{
					Enabled:     true,
					TokenBudget: 2048,
				},
			}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then thinking should be enabled with custom budget", func() {
				So(req.Thinking.OfEnabled, ShouldNotBeNil)
				So(req.Thinking.OfEnabled.BudgetTokens, ShouldEqual, 2048)
			})
		})

		Convey("When config has reasoning enabled with budget below minimum", func() {
			config := &v1.GenerationConfig{
				ReasoningConfig: &v1.ReasoningConfig{
					Enabled:     true,
					TokenBudget: 512,
				},
			}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then thinking should use minimum budget 1024", func() {
				So(req.Thinking.OfEnabled, ShouldNotBeNil)
				So(req.Thinking.OfEnabled.BudgetTokens, ShouldEqual, 1024)
			})
		})

		Convey("When config has all parameters set", func() {
			config := &v1.GenerationConfig{
				MaxTokens:   new(int64(4096)),
				Temperature: new(float32(0.75)),
				TopP:        new(float32(0.95)),
				TopK:        new(int64(40)),
				ReasoningConfig: &v1.ReasoningConfig{
					Enabled:     true,
					TokenBudget: 2048,
				},
			}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then all parameters should be applied", func() {
				So(req.MaxTokens, ShouldEqual, 4096)
				So(req.Temperature.Value, ShouldEqual, 0.75)
				So(req.TopP.Value, ShouldEqual, 0.95)
				So(req.TopK.Value, ShouldEqual, 40)
				So(req.Thinking.OfEnabled, ShouldNotBeNil)
				So(req.Thinking.OfEnabled.BudgetTokens, ShouldEqual, 2048)
			})
		})

		Convey("When config has json_schema grammar", func() {
			config := &v1.GenerationConfig{
				Grammar: &v1.GenerationConfig_JsonSchema{
					JsonSchema: `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}},"required":["name"]}`,
				},
			}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then OutputConfig.Format.Schema should be set", func() {
				expectedSchema := map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"age":  map[string]any{"type": "integer"},
					},
					"required": []any{"name"},
				}
				So(req.OutputConfig.Format.Schema, ShouldResemble, expectedSchema)
			})
		})

		Convey("When config has proto schema grammar", func() {
			config := &v1.GenerationConfig{
				Grammar: &v1.GenerationConfig_Schema{
					Schema: &v1.Schema{
						Type: v1.Schema_TYPE_OBJECT,
						Properties: map[string]*v1.Schema{
							"city": {Type: v1.Schema_TYPE_STRING, Description: "City name"},
							"temp": {Type: v1.Schema_TYPE_NUMBER},
						},
						Required: []string{"city"},
					},
				},
			}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then OutputConfig.Format.Schema should be set from proto schema", func() {
				So(req.OutputConfig.Format.Schema, ShouldNotBeNil)
				So(req.OutputConfig.Format.Schema["type"], ShouldEqual, "object")
				props, ok := req.OutputConfig.Format.Schema["properties"].(map[string]any)
				So(ok, ShouldBeTrue)
				cityProp, ok := props["city"].(map[string]any)
				So(ok, ShouldBeTrue)
				So(cityProp["type"], ShouldEqual, "string")
				So(cityProp["description"], ShouldEqual, "City name")
				required, ok := req.OutputConfig.Format.Schema["required"].([]any)
				So(ok, ShouldBeTrue)
				So(required, ShouldContain, "city")
			})
		})

		Convey("When config has invalid json_schema grammar", func() {
			config := &v1.GenerationConfig{
				Grammar: &v1.GenerationConfig_JsonSchema{
					JsonSchema: `{invalid json`,
				},
			}
			req := &anthropic.MessageNewParams{}
			repo.convertGenerationConfigToAnthropic(config, req)

			Convey("Then OutputConfig.Format.Schema should remain nil", func() {
				So(req.OutputConfig.Format.Schema, ShouldBeNil)
			})
		})
	})
}

func TestConvertSystemToAnthropic(t *testing.T) {
	Convey("Given messages with system/user roles", t, func() {
		repo := &upstream{
			config: &conf.AnthropicConfig{SystemAsUser: false},
			log:    log.NewHelper(log.DefaultLogger),
		}

		msgs := []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "sys prompt"}},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "hi"}},
				},
			},
		}

		parts := repo.convertSystemToAnthropic(msgs)

		Convey("Then only system text should be included", func() {
			So(parts, ShouldHaveLength, 1)
			So(parts[0].Text, ShouldEqual, "sys prompt")
		})
	})
}

func TestConvertMessageToAnthropic(t *testing.T) {
	Convey("Given a message converter", t, func() {
		repo := &upstream{
			config: &conf.AnthropicConfig{},
			log:    log.NewHelper(log.DefaultLogger),
		}

		Convey("When converting a USER message with text content", func() {
			msg := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Hello, world!"}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create a user message with text block", func() {
				So(string(result.Role), ShouldEqual, "user")
				So(result.Content, ShouldHaveLength, 1)
				textBlock := result.Content[0].OfText
				So(textBlock, ShouldNotBeNil)
				So(textBlock.Text, ShouldEqual, "Hello, world!")
			})
		})

		Convey("When converting a SYSTEM message with text content", func() {
			msg := &v1.Message{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "You are a helpful assistant"}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create a user message (system treated as user)", func() {
				So(string(result.Role), ShouldEqual, "user")
				So(result.Content, ShouldHaveLength, 1)
				textBlock := result.Content[0].OfText
				So(textBlock, ShouldNotBeNil)
				So(textBlock.Text, ShouldEqual, "You are a helpful assistant")
			})
		})

		Convey("When converting a MODEL message with text content", func() {
			msg := &v1.Message{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "I'm here to help!"}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create an assistant message", func() {
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.Content, ShouldHaveLength, 1)
				textBlock := result.Content[0].OfText
				So(textBlock, ShouldNotBeNil)
				So(textBlock.Text, ShouldEqual, "I'm here to help!")
			})
		})

		Convey("When converting a message with redacted thinking content", func() {
			msg := &v1.Message{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{
						Reasoning: true,
						Metadata:  map[string]string{"redacted_thinking": "opaque-data"},
						Content:   &v1.Content_Text{Text: ""},
					},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create an assistant message with redacted thinking block", func() {
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.Content, ShouldHaveLength, 1)
				redacted := result.Content[0].OfRedactedThinking
				So(redacted, ShouldNotBeNil)
				So(redacted.Data, ShouldEqual, "opaque-data")
			})
		})

		Convey("When converting a message with image content", func() {
			msg := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Image{
						Image: &v1.Image{
							Source: &v1.Image_Url{Url: "https://example.com/image.jpg"},
						},
					}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create a message with image block", func() {
				So(string(result.Role), ShouldEqual, "user")
				So(result.Content, ShouldHaveLength, 1)
				imageBlock := result.Content[0].OfImage
				So(imageBlock, ShouldNotBeNil)
				So(imageBlock.Source.OfURL, ShouldNotBeNil)
				So(imageBlock.Source.OfURL.URL, ShouldEqual, "https://example.com/image.jpg")
			})
		})

		Convey("When converting a message with base64 image content", func() {
			msg := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Image{
						Image: &v1.Image{
							MimeType: "image/png",
							Source:   &v1.Image_Data{Data: []byte("aW1hZ2VkYXRh")},
						},
					}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create a message with base64 image block", func() {
				So(string(result.Role), ShouldEqual, "user")
				So(result.Content, ShouldHaveLength, 1)
				imageBlock := result.Content[0].OfImage
				So(imageBlock, ShouldNotBeNil)
				So(imageBlock.Source.OfBase64, ShouldNotBeNil)
				So(imageBlock.Source.OfBase64.MediaType, ShouldEqual, anthropic.Base64ImageSourceMediaType("image/png"))
				So(imageBlock.Source.OfBase64.Data, ShouldEqual, "YVcxaFoyVmtZWFJo")
			})
		})

		Convey("When converting a message with multiple content types", func() {
			msg := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "What's in this image?"}},
					{Content: &v1.Content_Image{
						Image: &v1.Image{
							Source: &v1.Image_Url{Url: "https://example.com/photo.png"},
						},
					}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create a message with both text and image blocks", func() {
				So(string(result.Role), ShouldEqual, "user")
				So(result.Content, ShouldHaveLength, 2)

				// First content should be text
				textBlock := result.Content[0].OfText
				So(textBlock, ShouldNotBeNil)
				So(textBlock.Text, ShouldEqual, "What's in this image?")

				// Second content should be image
				imageBlock := result.Content[1].OfImage
				So(imageBlock, ShouldNotBeNil)
				So(imageBlock.Source.OfURL, ShouldNotBeNil)
				So(imageBlock.Source.OfURL.URL, ShouldEqual, "https://example.com/photo.png")
			})
		})

		Convey("When converting a message with empty contents", func() {
			msg := &v1.Message{
				Role:     v1.Role_USER,
				Contents: []*v1.Content{},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create a message with no content blocks", func() {
				So(string(result.Role), ShouldEqual, "user")
				So(result.Content, ShouldHaveLength, 0)
			})
		})

		Convey("When converting a message with tool_use that has invalid JSON input", func() {
			msg := &v1.Message{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "tool-1",
							Name: "my_tool",
							Inputs: []*v1.ToolUse_Input{{
								Input: &v1.ToolUse_Input_Text{Text: "not-valid-json"},
							}},
						},
					}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should still create a tool_use block with string fallback", func() {
				So(string(result.Role), ShouldEqual, "assistant")
				So(result.Content, ShouldHaveLength, 1)
				toolUse := result.Content[0].OfToolUse
				So(toolUse, ShouldNotBeNil)
				So(toolUse.ID, ShouldEqual, "tool-1")
				So(toolUse.Name, ShouldEqual, "my_tool")
				So(toolUse.Input, ShouldEqual, "not-valid-json")
			})
		})

		Convey("When converting a message with tool_result image URL", func() {
			msg := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_ToolResult{ToolResult: &v1.ToolResult{
						Id: "tool-1",
						Outputs: []*v1.ToolResult_Output{
							{Output: &v1.ToolResult_Output_Image{
								Image: &v1.Image{
									Source: &v1.Image_Url{Url: "https://example.com/tool-result.png"},
								},
							}},
						},
					}}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create a tool_result block with image URL", func() {
				So(string(result.Role), ShouldEqual, "user")
				So(result.Content, ShouldHaveLength, 1)
				toolResult := result.Content[0].OfToolResult
				So(toolResult, ShouldNotBeNil)
				So(toolResult.ToolUseID, ShouldEqual, "tool-1")
				So(toolResult.Content, ShouldHaveLength, 1)
				imageBlock := toolResult.Content[0].OfImage
				So(imageBlock, ShouldNotBeNil)
				So(imageBlock.Source.OfURL, ShouldNotBeNil)
				So(imageBlock.Source.OfURL.URL, ShouldEqual, "https://example.com/tool-result.png")
			})
		})

		Convey("When converting a message with tool_result base64 image", func() {
			msg := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_ToolResult{ToolResult: &v1.ToolResult{
						Id: "tool-1",
						Outputs: []*v1.ToolResult_Output{
							{Output: &v1.ToolResult_Output_Image{
								Image: &v1.Image{
									MimeType: "image/png",
									Source:   &v1.Image_Data{Data: []byte("imagedata")},
								},
							}},
						},
					}}},
				},
			}
			result := repo.convertMessageToAnthropic(msg)

			Convey("Then it should create a tool_result block with base64 image", func() {
				So(string(result.Role), ShouldEqual, "user")
				So(result.Content, ShouldHaveLength, 1)
				toolResult := result.Content[0].OfToolResult
				So(toolResult, ShouldNotBeNil)
				So(toolResult.ToolUseID, ShouldEqual, "tool-1")
				So(toolResult.Content, ShouldHaveLength, 1)
				imageBlock := toolResult.Content[0].OfImage
				So(imageBlock, ShouldNotBeNil)
				So(imageBlock.Source.OfBase64, ShouldNotBeNil)
				So(imageBlock.Source.OfBase64.MediaType, ShouldEqual, anthropic.Base64ImageSourceMediaType("image/png"))
				So(imageBlock.Source.OfBase64.Data, ShouldEqual, "aW1hZ2VkYXRh")
			})
		})
	})
}

func TestConvertInputSchemaToAnthropic(t *testing.T) {
	Convey("Given tool input schema", t, func() {
		repo := &upstream{config: &conf.AnthropicConfig{}, log: log.NewHelper(log.DefaultLogger)}

		Convey("When params is nil", func() {
			sch := repo.convertInputSchemaToAnthropic(nil)
			Convey("Then result is nil", func() {
				So(sch, ShouldBeZeroValue)
			})
		})

		Convey("When params has properties/required", func() {
			params := &v1.Schema{
				Type: v1.Schema_TYPE_OBJECT,
				Properties: map[string]*v1.Schema{
					"location": {Type: v1.Schema_TYPE_STRING, Description: "City name"},
				},
				Required: []string{"location"},
			}
			sch := repo.convertInputSchemaToAnthropic(params)
			Convey("Then fields are copied over", func() {
				So(sch.Properties, ShouldResemble, params.Properties)
				So(sch.Required, ShouldResemble, params.Required)
			})
		})
	})
}

func TestConvertRequestToAnthropic(t *testing.T) {
	Convey("Given a chat request with system, user, model messages and tools", t, func() {
		tools := []*v1.Tool{
			{
				Tool: &v1.Tool_Function_{
					Function: &v1.Tool_Function{
						Name:        "get_weather",
						Description: "Get weather",
						Parameters: &v1.Schema{
							Type: v1.Schema_TYPE_OBJECT,
							Properties: map[string]*v1.Schema{
								"city": {Type: v1.Schema_TYPE_STRING, Description: "City name"},
							},
							Required: []string{"city"},
						},
					},
				},
			},
		}

		msgs := []*v1.Message{
			{Role: v1.Role_SYSTEM, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "sys"}}}},
			{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "hi"}}}},
			{Role: v1.Role_MODEL, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "resp"}}}},
			{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Image{Image: &v1.Image{Source: &v1.Image_Url{Url: "https://a.b/c.png"}}}}}},
		}

		req := &entity.ChatReq{Id: "req-1", Model: "claude-3", Messages: msgs, Tools: tools}

		Convey("When SystemAsUser=false", func() {
			repo := &upstream{config: &conf.AnthropicConfig{SystemAsUser: false}, log: log.NewHelper(log.DefaultLogger)}
			out := repo.convertRequestToAnthropic(req)

			Convey("Then system is populated and system message is excluded from messages", func() {
				So(string(out.Model), ShouldEqual, "claude-3")
				So(out.System, ShouldHaveLength, 1)
				So(out.System[0].Text, ShouldEqual, "sys")
				So(out.Messages, ShouldHaveLength, 3) // all except system
			})

			Convey("And tools are converted", func() {
				So(len(out.Tools), ShouldEqual, 1)
				So(out.Tools[0].OfTool, ShouldNotBeNil)

				tool := out.Tools[0].OfTool
				So(tool.Name, ShouldEqual, "get_weather")
				So(tool.Description.Value, ShouldEqual, "Get weather")
				So(tool.InputSchema.Properties, ShouldNotBeNil)
				So(tool.InputSchema.Required, ShouldResemble, []string{"city"})

				// Validate nested properties structure
				properties, err := json.Marshal(tool.InputSchema.Properties)
				So(err, ShouldBeNil)
				So(properties, ShouldNotBeNil)
				So(gjson.Get(string(properties), "city").Exists(), ShouldBeTrue)
				So(gjson.Get(string(properties), "city").Get("description").String(), ShouldEqual, "City name")
			})
		})

		Convey("When SystemAsUser=true", func() {
			repo := &upstream{config: &conf.AnthropicConfig{SystemAsUser: true}, log: log.NewHelper(log.DefaultLogger)}
			out := repo.convertRequestToAnthropic(req)

			Convey("Then system is empty and messages include system entry", func() {
				So(out.System, ShouldBeNil)
				So(out.Messages, ShouldHaveLength, 4)
			})
		})
	})
}

func TestConvertContentsFromAnthropic(t *testing.T) {
	Convey("Given anthropic content blocks", t, func() {
		anthropicMessage := &anthropic.Message{
			ID: "msg-123",
			Content: []anthropic.ContentBlockUnion{
				{
					Type:     "thinking",
					Thinking: "think",
				},
				{
					Type: "text",
					Text: "answer",
				},
			},
		}

		msg := convertMessageFromAnthropic(anthropicMessage)

		Convey("Then they are mapped to message with thinking and text", func() {
			So(msg, ShouldNotBeNil)
			So(msg.Id, ShouldEqual, "msg-123")
			So(msg.Role, ShouldEqual, v1.Role_MODEL)
			So(len(msg.Contents), ShouldEqual, 2)
			So(msg.Contents[0].Reasoning, ShouldBeTrue)
			So(msg.Contents[0].GetText(), ShouldEqual, "think")
			So(msg.Contents[1].GetText(), ShouldEqual, "answer")
		})
	})

	Convey("Given anthropic content blocks with redacted_thinking", t, func() {
		anthropicMessage := &anthropic.Message{
			ID: "msg-456",
			Content: []anthropic.ContentBlockUnion{
				{
					Type:      "thinking",
					Thinking:  "visible thought",
					Signature: "sig-abc",
				},
				{
					Type: "redacted_thinking",
					Data: "opaque-encrypted-data",
				},
				{
					Type: "text",
					Text: "final answer",
				},
			},
		}

		msg := convertMessageFromAnthropic(anthropicMessage)

		Convey("Then thinking, redacted_thinking, and text are all mapped", func() {
			So(msg, ShouldNotBeNil)
			So(len(msg.Contents), ShouldEqual, 3)

			// thinking block
			So(msg.Contents[0].Reasoning, ShouldBeTrue)
			So(msg.Contents[0].GetText(), ShouldEqual, "visible thought")
			So(msg.Contents[0].Metadata["signature"], ShouldEqual, "sig-abc")

			// redacted_thinking block
			So(msg.Contents[1].Reasoning, ShouldBeTrue)
			So(msg.Contents[1].GetText(), ShouldEqual, "")
			So(msg.Contents[1].Metadata["redacted_thinking"], ShouldEqual, "opaque-encrypted-data")

			// text block
			So(msg.Contents[2].GetText(), ShouldEqual, "final answer")
		})
	})
}

func TestConvertChunkFromAnthropic(t *testing.T) {
	Convey("Given a stream client and various event chunks", t, func() {
		client := &anthropicChatStreamClient{req: &entity.ChatReq{Id: "req-1"}}

		Convey("When receiving message_start", func() {
			chunk := &anthropic.MessageStreamEventUnion{
				Type: "message_start",
				Message: anthropic.Message{
					ID:    "msg-1",
					Model: anthropic.Model("claude-3-sonnet"),
					Usage: anthropic.Usage{InputTokens: 12},
				},
			}
			resp := client.convertChunkFromAnthropic(chunk)
			Convey("Then no response is emitted yet", func() {
				So(resp, ShouldBeNil)
			})
		})

		Convey("When receiving content_block_start with text", func() {
			chunk := &anthropic.MessageStreamEventUnion{
				Type: "content_block_start",
				ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
					Type: "text",
					Text: "Hello",
				},
				Index: 0,
			}
			resp := client.convertChunkFromAnthropic(chunk)
			Convey("Then a text content is emitted", func() {
				So(resp, ShouldNotBeNil)
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Contents[0], ShouldNotBeNil)
				So(resp.Message.Contents[0].Index, ShouldNotBeNil)
				So(*resp.Message.Contents[0].Index, ShouldEqual, 0)
				So(resp.Message.Contents[0].Reasoning, ShouldBeFalse)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "Hello")
			})
		})

		Convey("When receiving content_block_start with thinking", func() {
			chunk := &anthropic.MessageStreamEventUnion{
				Type: "content_block_start",
				ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
					Type:      "thinking",
					Thinking:  "Let me analyze this...",
					Signature: "sig-abc",
				},
				Index: 1,
			}
			resp := client.convertChunkFromAnthropic(chunk)
			Convey("Then a thinking content is emitted", func() {
				So(resp, ShouldNotBeNil)
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Contents[0], ShouldNotBeNil)
				So(resp.Message.Contents[0].Index, ShouldNotBeNil)
				So(*resp.Message.Contents[0].Index, ShouldEqual, 1)
				So(resp.Message.Contents[0].Reasoning, ShouldBeTrue)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "Let me analyze this...")
				So(resp.Message.Contents[0].Metadata["signature"], ShouldEqual, "sig-abc")
			})
		})

		Convey("When receiving content_block_start with tool_use", func() {
			chunk := &anthropic.MessageStreamEventUnion{
				Type: "content_block_start",
				ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
					Type: "tool_use",
					ID:   "tool-1",
					Name: "get_weather",
				},
			}
			resp := client.convertChunkFromAnthropic(chunk)
			Convey("Then a function_call content is emitted", func() {
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "req-1")
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(resp.Message.Contents[0].GetToolUse(), ShouldNotBeNil)
				So(resp.Message.Contents[0].GetToolUse().Id, ShouldEqual, "tool-1")
				So(resp.Message.Contents[0].GetToolUse().Name, ShouldEqual, "get_weather")
			})
		})

		Convey("When receiving content_block_start with redacted_thinking", func() {
			chunk := &anthropic.MessageStreamEventUnion{
				Type: "content_block_start",
				ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
					Type: "redacted_thinking",
					Data: "opaque-data-123",
				},
				Index: 2,
			}
			resp := client.convertChunkFromAnthropic(chunk)
			Convey("Then a redacted thinking content is emitted", func() {
				So(resp, ShouldNotBeNil)
				So(resp.Message.Contents[0], ShouldNotBeNil)
				So(resp.Message.Contents[0].Index, ShouldNotBeNil)
				So(*resp.Message.Contents[0].Index, ShouldEqual, 2)
				So(resp.Message.Contents[0].Reasoning, ShouldBeTrue)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "")
				So(resp.Message.Contents[0].Metadata["redacted_thinking"], ShouldEqual, "opaque-data-123")
			})
		})

		Convey("When receiving content_block_delta variants", func() {
			// thinking_delta
			d1 := &anthropic.MessageStreamEventUnion{Type: "content_block_delta", Delta: anthropic.MessageStreamEventUnionDelta{Type: "thinking_delta", Thinking: "let me think"}}
			r1 := client.convertChunkFromAnthropic(d1)
			So(r1, ShouldNotBeNil)
			So(r1.Message.Contents[0].Reasoning, ShouldBeTrue)
			So(r1.Message.Contents[0].GetText(), ShouldEqual, "let me think")
			So(r1.Message.Contents[0].Metadata, ShouldBeNil)

			// signature_delta
			d1s := &anthropic.MessageStreamEventUnion{Type: "content_block_delta", Delta: anthropic.MessageStreamEventUnionDelta{Type: "signature_delta", Signature: "sig-xyz"}}
			r1s := client.convertChunkFromAnthropic(d1s)
			So(r1s, ShouldNotBeNil)
			So(r1s.Message.Contents[0].Reasoning, ShouldBeTrue)
			So(r1s.Message.Contents[0].Metadata["signature"], ShouldEqual, "sig-xyz")
			So(r1s.Message.Contents[0].GetText(), ShouldEqual, "")

			// text_delta
			d2 := &anthropic.MessageStreamEventUnion{Type: "content_block_delta", Delta: anthropic.MessageStreamEventUnionDelta{Type: "text_delta", Text: "hello"}}
			r2 := client.convertChunkFromAnthropic(d2)
			So(r2, ShouldNotBeNil)
			So(r2.Message.Contents[0].GetText(), ShouldEqual, "hello")

			// input_json_delta
			d3 := &anthropic.MessageStreamEventUnion{Type: "content_block_delta", Delta: anthropic.MessageStreamEventUnionDelta{Type: "input_json_delta", PartialJSON: "{\"x\":1}"}}
			r3 := client.convertChunkFromAnthropic(d3)
			So(r3, ShouldNotBeNil)
			So(r3.Message.Contents[0].GetToolUse(), ShouldNotBeNil)
			So(r3.Message.Contents[0].GetToolUse().GetTextualInput(), ShouldEqual, "{\"x\":1}")
		})

		Convey("When receiving message_delta with usage", func() {
			// For input tokens
			chunk := &anthropic.MessageStreamEventUnion{
				Type: "message_start",
				Message: anthropic.Message{
					ID:    "msg-1",
					Model: anthropic.Model("claude-3-sonnet"),
					Usage: anthropic.Usage{InputTokens: 12},
				},
			}
			resp := client.convertChunkFromAnthropic(chunk)
			So(resp, ShouldBeNil)

			chunk = &anthropic.MessageStreamEventUnion{
				Type:  "message_delta",
				Usage: anthropic.MessageDeltaUsage{OutputTokens: 34},
			}
			resp = client.convertChunkFromAnthropic(chunk)
			Convey("Then statistics are emitted", func() {
				So(resp, ShouldNotBeNil)
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage.InputTokens, ShouldEqual, 12)
				So(resp.Statistics.Usage.OutputTokens, ShouldEqual, 34)
			})
		})
	})
}
