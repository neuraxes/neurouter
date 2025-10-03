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
		blocks := []anthropic.ContentBlockUnion{
			{Thinking: "think"},
			{Text: "answer"},
		}

		msg := convertContentsFromAnthropic(blocks)

		Convey("Then they are mapped to message with thinking and text", func() {
			So(msg, ShouldNotBeNil)
			So(msg.Id, ShouldHaveLength, 36)
			So(msg.Role, ShouldEqual, v1.Role_MODEL)
			So(len(msg.Contents), ShouldEqual, 2)
			So(msg.Contents[0].GetThinking(), ShouldEqual, "think")
			So(msg.Contents[1].GetText(), ShouldEqual, "answer")
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

		Convey("When receiving content_block_start with tool_use", func() {
			chunk := &anthropic.MessageStreamEventUnion{
				Type: "content_block_start",
				ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
					Type:      "tool_use",
					ToolUseID: "tool-1",
					Name:      "get_weather",
				},
			}
			resp := client.convertChunkFromAnthropic(chunk)
			Convey("Then a function_call content is emitted", func() {
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "req-1")
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				fc := resp.Message.Contents[0].GetFunctionCall()
				So(fc, ShouldNotBeNil)
				So(fc.Id, ShouldEqual, "tool-1")
				So(fc.Name, ShouldEqual, "get_weather")
			})
		})

		Convey("When receiving content_block_delta variants", func() {
			// thinking_delta
			d1 := &anthropic.MessageStreamEventUnion{Type: "content_block_delta", Delta: anthropic.MessageStreamEventUnionDelta{Type: "thinking_delta", Thinking: "let me think"}}
			r1 := client.convertChunkFromAnthropic(d1)
			So(r1, ShouldNotBeNil)
			So(r1.Message.Contents[0].GetThinking(), ShouldEqual, "let me think")

			// text_delta
			d2 := &anthropic.MessageStreamEventUnion{Type: "content_block_delta", Delta: anthropic.MessageStreamEventUnionDelta{Type: "text_delta", Text: "hello"}}
			r2 := client.convertChunkFromAnthropic(d2)
			So(r2, ShouldNotBeNil)
			So(r2.Message.Contents[0].GetText(), ShouldEqual, "hello")

			// input_json_delta
			d3 := &anthropic.MessageStreamEventUnion{Type: "content_block_delta", Delta: anthropic.MessageStreamEventUnionDelta{Type: "input_json_delta", PartialJSON: "{\"x\":1}"}}
			r3 := client.convertChunkFromAnthropic(d3)
			So(r3, ShouldNotBeNil)
			So(r3.Message.Contents[0].GetFunctionCall(), ShouldNotBeNil)
			So(r3.Message.Contents[0].GetFunctionCall().Arguments, ShouldEqual, "{\"x\":1}")
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
				So(resp.Statistics.Usage.PromptTokens, ShouldEqual, 12)
				So(resp.Statistics.Usage.CompletionTokens, ShouldEqual, 34)
			})
		})
	})
}
