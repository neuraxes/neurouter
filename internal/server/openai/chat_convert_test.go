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

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestConvertImageFromOpenAIURL(t *testing.T) {
	Convey("Given a URL to convert to an image", t, func() {

		Convey("When the URL is a regular HTTP URL", func() {
			result := convertImageFromOpenAIURL("https://example.com/image.jpg")

			Convey("Then it should return an image with URL source", func() {
				So(result, ShouldNotBeNil)
				urlSrc, ok := result.Source.(*v1.Image_Url)
				So(ok, ShouldBeTrue)
				So(urlSrc.Url, ShouldEqual, "https://example.com/image.jpg")
			})
		})

		Convey("When the URL is a base64 data URI", func() {
			result := convertImageFromOpenAIURL("data:image/png;base64,aW1hZ2VkYXRh")

			Convey("Then it should return an image with decoded data", func() {
				So(result, ShouldNotBeNil)
				So(result.MimeType, ShouldEqual, "image/png")
				dataSrc, ok := result.Source.(*v1.Image_Data)
				So(ok, ShouldBeTrue)
				So(string(dataSrc.Data), ShouldEqual, "imagedata")
			})
		})

		Convey("When the URL is a data URI with invalid base64", func() {
			result := convertImageFromOpenAIURL("data:image/png;base64,!!!invalid!!!")

			Convey("Then it should fall back to URL source", func() {
				So(result, ShouldNotBeNil)
				urlSrc, ok := result.Source.(*v1.Image_Url)
				So(ok, ShouldBeTrue)
				So(urlSrc.Url, ShouldEqual, "data:image/png;base64,!!!invalid!!!")
			})
		})

		Convey("When the URL is a data URI without base64 encoding", func() {
			result := convertImageFromOpenAIURL("data:text/plain,hello")

			Convey("Then it should fall back to URL source", func() {
				So(result, ShouldNotBeNil)
				urlSrc, ok := result.Source.(*v1.Image_Url)
				So(ok, ShouldBeTrue)
				So(urlSrc.Url, ShouldEqual, "data:text/plain,hello")
			})
		})
	})
}

func TestConvertDeveloperMessageFromOpenAIChat(t *testing.T) {
	Convey("Given a developer message to convert", t, func() {

		Convey("When content is a string", func() {
			m := &openai.ChatCompletionDeveloperMessageParam{
				Content: openai.ChatCompletionDeveloperMessageParamContentUnion{
					OfString: openai.Opt("You are a helpful assistant."),
				},
			}
			result := convertDeveloperMessageFromOpenAIChat(m)

			Convey("Then it should create a SYSTEM message with text content", func() {
				So(result.Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "You are a helpful assistant.")
			})
		})

		Convey("When content is an array of text parts", func() {
			m := &openai.ChatCompletionDeveloperMessageParam{
				Content: openai.ChatCompletionDeveloperMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartTextParam{
						{Text: "Part 1"},
						{Text: "Part 2"},
					},
				},
			}
			result := convertDeveloperMessageFromOpenAIChat(m)

			Convey("Then it should create a SYSTEM message with multiple text contents", func() {
				So(result.Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Contents, ShouldHaveLength, 2)
				So(result.Contents[0].GetText(), ShouldEqual, "Part 1")
				So(result.Contents[1].GetText(), ShouldEqual, "Part 2")
			})
		})

		Convey("When name is set", func() {
			m := &openai.ChatCompletionDeveloperMessageParam{
				Content: openai.ChatCompletionDeveloperMessageParamContentUnion{
					OfString: openai.Opt("Hello"),
				},
				Name: openai.Opt("dev"),
			}
			result := convertDeveloperMessageFromOpenAIChat(m)

			Convey("Then the name should be preserved", func() {
				So(result.Name, ShouldEqual, "dev")
			})
		})

		Convey("When name is not set", func() {
			m := &openai.ChatCompletionDeveloperMessageParam{
				Content: openai.ChatCompletionDeveloperMessageParamContentUnion{
					OfString: openai.Opt("Hello"),
				},
			}
			result := convertDeveloperMessageFromOpenAIChat(m)

			Convey("Then the name should be empty", func() {
				So(result.Name, ShouldBeEmpty)
			})
		})
	})
}

func TestConvertSystemMessageFromOpenAIChat(t *testing.T) {
	Convey("Given a system message to convert", t, func() {

		Convey("When content is a string", func() {
			m := &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.Opt("System prompt"),
				},
			}
			result := convertSystemMessageFromOpenAIChat(m)

			Convey("Then it should create a SYSTEM message with text content", func() {
				So(result.Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "System prompt")
			})
		})

		Convey("When content is an array of text parts", func() {
			m := &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartTextParam{
						{Text: "Be helpful."},
						{Text: "Be concise."},
					},
				},
			}
			result := convertSystemMessageFromOpenAIChat(m)

			Convey("Then it should create a SYSTEM message with multiple text contents", func() {
				So(result.Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Contents, ShouldHaveLength, 2)
				So(result.Contents[0].GetText(), ShouldEqual, "Be helpful.")
				So(result.Contents[1].GetText(), ShouldEqual, "Be concise.")
			})
		})

		Convey("When name is set", func() {
			m := &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.Opt("Hello"),
				},
				Name: openai.Opt("sys"),
			}
			result := convertSystemMessageFromOpenAIChat(m)

			Convey("Then the name should be preserved", func() {
				So(result.Name, ShouldEqual, "sys")
			})
		})
	})
}

func TestConvertUserMessageFromOpenAIChat(t *testing.T) {
	Convey("Given a user message to convert", t, func() {

		Convey("When content is a string", func() {
			m := &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.Opt("Hello!"),
				},
			}
			result := convertUserMessageFromOpenAIChat(m)

			Convey("Then it should create a USER message with text content", func() {
				So(result.Role, ShouldEqual, v1.Role_USER)
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "Hello!")
			})
		})

		Convey("When content is an array with text parts", func() {
			m := &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						{OfText: &openai.ChatCompletionContentPartTextParam{Text: "Part A"}},
						{OfText: &openai.ChatCompletionContentPartTextParam{Text: "Part B"}},
					},
				},
			}
			result := convertUserMessageFromOpenAIChat(m)

			Convey("Then it should create a USER message with multiple text contents", func() {
				So(result.Role, ShouldEqual, v1.Role_USER)
				So(result.Contents, ShouldHaveLength, 2)
				So(result.Contents[0].GetText(), ShouldEqual, "Part A")
				So(result.Contents[1].GetText(), ShouldEqual, "Part B")
			})
		})

		Convey("When content is an array with an image URL", func() {
			m := &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						{OfText: &openai.ChatCompletionContentPartTextParam{Text: "Look at this:"}},
						{OfImageURL: &openai.ChatCompletionContentPartImageParam{
							ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
								URL: "https://example.com/photo.png",
							},
						}},
					},
				},
			}
			result := convertUserMessageFromOpenAIChat(m)

			Convey("Then it should create text and image contents", func() {
				So(result.Contents, ShouldHaveLength, 2)
				So(result.Contents[0].GetText(), ShouldEqual, "Look at this:")

				img := result.Contents[1].GetImage()
				So(img, ShouldNotBeNil)
				urlSrc, ok := img.Source.(*v1.Image_Url)
				So(ok, ShouldBeTrue)
				So(urlSrc.Url, ShouldEqual, "https://example.com/photo.png")
			})
		})

		Convey("When content is an array with a base64 data URI image", func() {
			m := &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						{OfImageURL: &openai.ChatCompletionContentPartImageParam{
							ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
								URL: "data:image/jpeg;base64,aW1hZ2VkYXRh",
							},
						}},
					},
				},
			}
			result := convertUserMessageFromOpenAIChat(m)

			Convey("Then it should decode the base64 data", func() {
				So(result.Contents, ShouldHaveLength, 1)
				img := result.Contents[0].GetImage()
				So(img, ShouldNotBeNil)
				So(img.MimeType, ShouldEqual, "image/jpeg")
				dataSrc, ok := img.Source.(*v1.Image_Data)
				So(ok, ShouldBeTrue)
				So(string(dataSrc.Data), ShouldEqual, "imagedata")
			})
		})

		Convey("When content array has unsupported part types", func() {
			m := &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						{},
					},
				},
			}
			result := convertUserMessageFromOpenAIChat(m)

			Convey("Then unsupported parts should be skipped", func() {
				So(result.Contents, ShouldBeEmpty)
			})
		})

		Convey("When name is set", func() {
			m := &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.Opt("Hi"),
				},
				Name: openai.Opt("Alice"),
			}
			result := convertUserMessageFromOpenAIChat(m)

			Convey("Then the name should be preserved", func() {
				So(result.Name, ShouldEqual, "Alice")
			})
		})
	})
}

func TestConvertAssistantMessageFromOpenAIChat(t *testing.T) {
	Convey("Given an assistant message to convert", t, func() {

		Convey("When content is a string", func() {
			m := &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.Opt("I can help!"),
				},
			}
			result := convertAssistantMessageFromOpenAIChat(m)

			Convey("Then it should create a MODEL message with text content", func() {
				So(result.Role, ShouldEqual, v1.Role_MODEL)
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "I can help!")
			})
		})

		Convey("When content is an array of text parts", func() {
			m := &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion{
						{OfText: &openai.ChatCompletionContentPartTextParam{Text: "Line 1"}},
						{OfText: &openai.ChatCompletionContentPartTextParam{Text: "Line 2"}},
					},
				},
			}
			result := convertAssistantMessageFromOpenAIChat(m)

			Convey("Then it should create a MODEL message with multiple text contents", func() {
				So(result.Role, ShouldEqual, v1.Role_MODEL)
				So(result.Contents, ShouldHaveLength, 2)
				So(result.Contents[0].GetText(), ShouldEqual, "Line 1")
				So(result.Contents[1].GetText(), ShouldEqual, "Line 2")
			})
		})

		Convey("When content array has nil OfText parts", func() {
			m := &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion{
						{},
					},
				},
			}
			result := convertAssistantMessageFromOpenAIChat(m)

			Convey("Then nil parts should be skipped", func() {
				So(result.Contents, ShouldBeEmpty)
			})
		})

		Convey("When tool calls are present", func() {
			m := &openai.ChatCompletionAssistantMessageParam{
				ToolCalls: []openai.ChatCompletionMessageToolCallUnionParam{
					{OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: "call-1",
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      "get_weather",
							Arguments: `{"city":"Shanghai"}`,
						},
					}},
					{OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: "call-2",
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      "search",
							Arguments: `{"q":"test"}`,
						},
					}},
				},
			}
			result := convertAssistantMessageFromOpenAIChat(m)

			Convey("Then tool calls should be converted to ToolUse contents", func() {
				So(result.Contents, ShouldHaveLength, 2)
				tu1 := result.Contents[0].GetToolUse()
				So(tu1, ShouldNotBeNil)
				So(tu1.Id, ShouldEqual, "call-1")
				So(tu1.Name, ShouldEqual, "get_weather")
				So(tu1.GetTextualInput(), ShouldEqual, `{"city":"Shanghai"}`)

				tu2 := result.Contents[1].GetToolUse()
				So(tu2, ShouldNotBeNil)
				So(tu2.Id, ShouldEqual, "call-2")
				So(tu2.Name, ShouldEqual, "search")
				So(tu2.GetTextualInput(), ShouldEqual, `{"q":"test"}`)
			})
		})

		Convey("When both text and tool calls are present", func() {
			m := &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.Opt("Let me check."),
				},
				ToolCalls: []openai.ChatCompletionMessageToolCallUnionParam{
					{OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: "call-1",
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      "lookup",
							Arguments: `{}`,
						},
					}},
				},
			}
			result := convertAssistantMessageFromOpenAIChat(m)

			Convey("Then both text and tool call should appear in contents", func() {
				So(result.Contents, ShouldHaveLength, 2)
				So(result.Contents[0].GetText(), ShouldEqual, "Let me check.")
				So(result.Contents[1].GetToolUse().Name, ShouldEqual, "lookup")
			})
		})

		Convey("When tool call has nil OfFunction", func() {
			m := &openai.ChatCompletionAssistantMessageParam{
				ToolCalls: []openai.ChatCompletionMessageToolCallUnionParam{
					{},
				},
			}
			result := convertAssistantMessageFromOpenAIChat(m)

			Convey("Then non-function tool calls should be skipped", func() {
				So(result.Contents, ShouldBeEmpty)
			})
		})

		Convey("When name is set", func() {
			m := &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.Opt("Reply"),
				},
				Name: openai.Opt("Claude"),
			}
			result := convertAssistantMessageFromOpenAIChat(m)

			Convey("Then the name should be preserved", func() {
				So(result.Name, ShouldEqual, "Claude")
			})
		})
	})
}

func TestConvertToolMessageFromOpenAIChat(t *testing.T) {
	Convey("Given a tool message to convert", t, func() {

		Convey("When content is a string", func() {
			m := &openai.ChatCompletionToolMessageParam{
				ToolCallID: "call-1",
				Content: openai.ChatCompletionToolMessageParamContentUnion{
					OfString: openai.Opt("Tool result text"),
				},
			}
			result := convertToolMessageFromOpenAIChat(m)

			Convey("Then it should create a USER message with ToolResult content", func() {
				So(result.Role, ShouldEqual, v1.Role_USER)
				So(result.Contents, ShouldHaveLength, 1)
				tr := result.Contents[0].GetToolResult()
				So(tr, ShouldNotBeNil)
				So(tr.Id, ShouldEqual, "call-1")
				So(tr.Outputs, ShouldHaveLength, 1)
				So(tr.Outputs[0].GetText(), ShouldEqual, "Tool result text")
			})
		})

		Convey("When content is an array of text parts", func() {
			m := &openai.ChatCompletionToolMessageParam{
				ToolCallID: "call-2",
				Content: openai.ChatCompletionToolMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartTextParam{
						{Text: "Output line 1"},
						{Text: "Output line 2"},
					},
				},
			}
			result := convertToolMessageFromOpenAIChat(m)

			Convey("Then it should create a ToolResult with multiple outputs", func() {
				So(result.Role, ShouldEqual, v1.Role_USER)
				tr := result.Contents[0].GetToolResult()
				So(tr, ShouldNotBeNil)
				So(tr.Id, ShouldEqual, "call-2")
				So(tr.Outputs, ShouldHaveLength, 2)
				So(tr.Outputs[0].GetText(), ShouldEqual, "Output line 1")
				So(tr.Outputs[1].GetText(), ShouldEqual, "Output line 2")
			})
		})

		Convey("When content is empty", func() {
			m := &openai.ChatCompletionToolMessageParam{
				ToolCallID: "call-3",
			}
			result := convertToolMessageFromOpenAIChat(m)

			Convey("Then it should create a ToolResult with empty outputs", func() {
				tr := result.Contents[0].GetToolResult()
				So(tr, ShouldNotBeNil)
				So(tr.Id, ShouldEqual, "call-3")
				So(tr.Outputs, ShouldBeEmpty)
			})
		})
	})
}

func TestConvertChatMessageFromOpenAIChat(t *testing.T) {
	Convey("Given a ChatCompletionMessageParamUnion to dispatch", t, func() {

		Convey("When it is a developer message", func() {
			msg := openai.ChatCompletionMessageParamUnion{
				OfDeveloper: &openai.ChatCompletionDeveloperMessageParam{
					Content: openai.ChatCompletionDeveloperMessageParamContentUnion{
						OfString: openai.Opt("dev msg"),
					},
				},
			}
			result := convertChatMessageFromOpenAIChat(msg)

			Convey("Then it should dispatch to developer converter", func() {
				So(result, ShouldNotBeNil)
				So(result.Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Contents[0].GetText(), ShouldEqual, "dev msg")
			})
		})

		Convey("When it is a system message", func() {
			msg := openai.ChatCompletionMessageParamUnion{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: openai.Opt("sys msg"),
					},
				},
			}
			result := convertChatMessageFromOpenAIChat(msg)

			Convey("Then it should dispatch to system converter", func() {
				So(result, ShouldNotBeNil)
				So(result.Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Contents[0].GetText(), ShouldEqual, "sys msg")
			})
		})

		Convey("When it is a user message", func() {
			msg := openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.Opt("user msg"),
					},
				},
			}
			result := convertChatMessageFromOpenAIChat(msg)

			Convey("Then it should dispatch to user converter", func() {
				So(result, ShouldNotBeNil)
				So(result.Role, ShouldEqual, v1.Role_USER)
				So(result.Contents[0].GetText(), ShouldEqual, "user msg")
			})
		})

		Convey("When it is an assistant message", func() {
			msg := openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: openai.Opt("assistant msg"),
					},
				},
			}
			result := convertChatMessageFromOpenAIChat(msg)

			Convey("Then it should dispatch to assistant converter", func() {
				So(result, ShouldNotBeNil)
				So(result.Role, ShouldEqual, v1.Role_MODEL)
				So(result.Contents[0].GetText(), ShouldEqual, "assistant msg")
			})
		})

		Convey("When it is a tool message", func() {
			msg := openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					ToolCallID: "call-1",
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.Opt("tool output"),
					},
				},
			}
			result := convertChatMessageFromOpenAIChat(msg)

			Convey("Then it should dispatch to tool converter", func() {
				So(result, ShouldNotBeNil)
				So(result.Role, ShouldEqual, v1.Role_USER)
				tr := result.Contents[0].GetToolResult()
				So(tr, ShouldNotBeNil)
				So(tr.Id, ShouldEqual, "call-1")
			})
		})

		Convey("When all fields are nil", func() {
			msg := openai.ChatCompletionMessageParamUnion{}
			result := convertChatMessageFromOpenAIChat(msg)

			Convey("Then it should return nil", func() {
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestConvertChatReqFromOpenAIChat(t *testing.T) {
	Convey("Given an OpenAI ChatCompletionNewParams to convert", t, func() {

		Convey("When MaxCompletionTokens is set", func() {
			req := &openai.ChatCompletionNewParams{
				Model:               "gpt-4o",
				MaxCompletionTokens: openai.Opt[int64](4096),
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then MaxTokens should be set in config", func() {
				So(result.Config.MaxTokens, ShouldNotBeNil)
				So(*result.Config.MaxTokens, ShouldEqual, 4096)
			})
		})

		Convey("When only MaxTokens (deprecated) is set", func() {
			req := &openai.ChatCompletionNewParams{
				Model:     "gpt-4",
				MaxTokens: openai.Opt[int64](2048),
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then MaxTokens should be set from the deprecated field", func() {
				So(result.Config.MaxTokens, ShouldNotBeNil)
				So(*result.Config.MaxTokens, ShouldEqual, 2048)
			})
		})

		Convey("When both MaxCompletionTokens and MaxTokens are set", func() {
			req := &openai.ChatCompletionNewParams{
				Model:               "gpt-4",
				MaxCompletionTokens: openai.Opt[int64](8192),
				MaxTokens:           openai.Opt[int64](1024),
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then MaxCompletionTokens should take precedence", func() {
				So(*result.Config.MaxTokens, ShouldEqual, 8192)
			})
		})

		Convey("When Temperature is set", func() {
			req := &openai.ChatCompletionNewParams{
				Model:       "gpt-4",
				Temperature: openai.Opt(0.7),
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then Temperature should be applied", func() {
				So(result.Config.Temperature, ShouldNotBeNil)
				So(*result.Config.Temperature, ShouldAlmostEqual, 0.7, 0.01)
			})
		})

		Convey("When TopP is set", func() {
			req := &openai.ChatCompletionNewParams{
				Model: "gpt-4",
				TopP:  openai.Opt(0.95),
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then TopP should be applied", func() {
				So(result.Config.TopP, ShouldNotBeNil)
				So(*result.Config.TopP, ShouldAlmostEqual, 0.95, 0.01)
			})
		})

		Convey("When FrequencyPenalty is set", func() {
			req := &openai.ChatCompletionNewParams{
				Model:            "gpt-4",
				FrequencyPenalty: openai.Opt(1.5),
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then FrequencyPenalty should be applied", func() {
				So(result.Config.FrequencyPenalty, ShouldNotBeNil)
				So(*result.Config.FrequencyPenalty, ShouldAlmostEqual, 1.5, 0.01)
			})
		})

		Convey("When PresencePenalty is set", func() {
			req := &openai.ChatCompletionNewParams{
				Model:           "gpt-4",
				PresencePenalty: openai.Opt(0.5),
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then PresencePenalty should be applied", func() {
				So(result.Config.PresencePenalty, ShouldNotBeNil)
				So(*result.Config.PresencePenalty, ShouldAlmostEqual, 0.5, 0.01)
			})
		})

		Convey("When ResponseFormat is JSON object", func() {
			req := &openai.ChatCompletionNewParams{
				Model: "gpt-4",
				ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
					OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
				},
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then Grammar should be set as json_object preset", func() {
				So(result.Config.Grammar, ShouldNotBeNil)
				preset, ok := result.Config.Grammar.(*v1.GenerationConfig_PresetGrammar)
				So(ok, ShouldBeTrue)
				So(preset.PresetGrammar, ShouldEqual, "json_object")
			})
		})

		Convey("When no optional config fields are set", func() {
			req := &openai.ChatCompletionNewParams{
				Model: "gpt-4",
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then config optional fields should be nil", func() {
				So(result.Config.MaxTokens, ShouldBeNil)
				So(result.Config.Temperature, ShouldBeNil)
				So(result.Config.TopP, ShouldBeNil)
				So(result.Config.FrequencyPenalty, ShouldBeNil)
				So(result.Config.PresencePenalty, ShouldBeNil)
				So(result.Config.Grammar, ShouldBeNil)
			})
		})

		Convey("When messages are provided", func() {
			req := &openai.ChatCompletionNewParams{
				Model: "gpt-4",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{OfSystem: &openai.ChatCompletionSystemMessageParam{
						Content: openai.ChatCompletionSystemMessageParamContentUnion{
							OfString: openai.Opt("Be helpful."),
						},
					}},
					{OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.Opt("Hello"),
						},
					}},
					{OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						Content: openai.ChatCompletionAssistantMessageParamContentUnion{
							OfString: openai.Opt("Hi there"),
						},
					}},
				},
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then all messages should be converted", func() {
				So(result.Model, ShouldEqual, "gpt-4")
				So(result.Messages, ShouldHaveLength, 3)
				So(result.Messages[0].Role, ShouldEqual, v1.Role_SYSTEM)
				So(result.Messages[1].Role, ShouldEqual, v1.Role_USER)
				So(result.Messages[2].Role, ShouldEqual, v1.Role_MODEL)
			})
		})

		Convey("When nil messages are in the list", func() {
			req := &openai.ChatCompletionNewParams{
				Model: "gpt-4",
				Messages: []openai.ChatCompletionMessageParamUnion{
					{OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.Opt("Hello"),
						},
					}},
					{},
				},
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then nil messages should be filtered out", func() {
				So(result.Messages, ShouldHaveLength, 1)
			})
		})

		Convey("When tools are provided", func() {
			req := &openai.ChatCompletionNewParams{
				Model: "gpt-4",
				Tools: []openai.ChatCompletionToolUnionParam{
					{OfFunction: &openai.ChatCompletionFunctionToolParam{
						Function: shared.FunctionDefinitionParam{
							Name:        "get_weather",
							Description: openai.Opt("Get weather info"),
							Parameters: shared.FunctionParameters{
								"type": "object",
								"properties": map[string]any{
									"city": map[string]any{
										"type": "string",
									},
								},
							},
						},
					}},
				},
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then tools should be converted", func() {
				So(result.Tools, ShouldHaveLength, 1)
				fn := result.Tools[0].GetFunction()
				So(fn, ShouldNotBeNil)
				So(fn.Name, ShouldEqual, "get_weather")
				So(fn.Description, ShouldEqual, "Get weather info")
				So(fn.Parameters, ShouldNotBeNil)
				So(fn.Parameters.Type, ShouldEqual, v1.Schema_TYPE_OBJECT)
				So(fn.Parameters.Properties, ShouldContainKey, "city")
			})
		})

		Convey("When tool has nil OfFunction", func() {
			req := &openai.ChatCompletionNewParams{
				Model: "gpt-4",
				Tools: []openai.ChatCompletionToolUnionParam{
					{},
				},
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then nil tools should be skipped", func() {
				So(result.Tools, ShouldBeEmpty)
			})
		})

		Convey("When all parameters are set", func() {
			req := &openai.ChatCompletionNewParams{
				Model:               "gpt-4o",
				MaxCompletionTokens: openai.Opt[int64](4096),
				Temperature:         openai.Opt(0.8),
				TopP:                openai.Opt(0.9),
				FrequencyPenalty:    openai.Opt(0.5),
				PresencePenalty:     openai.Opt(0.3),
				ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
					OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
				},
				Messages: []openai.ChatCompletionMessageParamUnion{
					{OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.Opt("Hello"),
						},
					}},
				},
			}
			result := convertChatReqFromOpenAIChat(req)

			Convey("Then all fields should be populated", func() {
				So(result.Model, ShouldEqual, "gpt-4o")
				So(*result.Config.MaxTokens, ShouldEqual, 4096)
				So(*result.Config.Temperature, ShouldAlmostEqual, 0.8, 0.01)
				So(*result.Config.TopP, ShouldAlmostEqual, 0.9, 0.01)
				So(*result.Config.FrequencyPenalty, ShouldAlmostEqual, 0.5, 0.01)
				So(*result.Config.PresencePenalty, ShouldAlmostEqual, 0.3, 0.01)
				So(result.Config.Grammar, ShouldNotBeNil)
				preset, ok := result.Config.Grammar.(*v1.GenerationConfig_PresetGrammar)
				So(ok, ShouldBeTrue)
				So(preset.PresetGrammar, ShouldEqual, "json_object")
				So(result.Messages, ShouldHaveLength, 1)
			})
		})
	})
}

func TestConvertStatusToOpenAIChat(t *testing.T) {
	Convey("Given various internal chat statuses", t, func() {
		So(convertStatusToOpenAIChat(v1.ChatStatus_CHAT_COMPLETED), ShouldEqual, "stop")
		So(convertStatusToOpenAIChat(v1.ChatStatus_CHAT_REFUSED), ShouldEqual, "content_filter")
		So(convertStatusToOpenAIChat(v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT), ShouldEqual, "length")
		So(convertStatusToOpenAIChat(v1.ChatStatus_CHAT_PENDING_TOOL_USE), ShouldEqual, "tool_calls")
		So(convertStatusToOpenAIChat(v1.ChatStatus_CHAT_IN_PROGRESS), ShouldEqual, "")
		So(convertStatusToOpenAIChat(v1.ChatStatus_CHAT_CANCELLED), ShouldEqual, "")
	})
}

func TestConvertUsageToOpenAIChat(t *testing.T) {
	Convey("Given usage statistics to convert", t, func() {

		Convey("When usage is nil", func() {
			result := convertUsageToOpenAIChat(nil)

			Convey("Then result should be nil", func() {
				So(result, ShouldBeNil)
			})
		})

		Convey("When usage has all fields", func() {
			u := &v1.Statistics_Usage{
				InputTokens:       100,
				OutputTokens:      50,
				CachedInputTokens: 20,
				ReasoningTokens:   10,
			}
			result := convertUsageToOpenAIChat(u)

			Convey("Then all fields should be mapped", func() {
				So(result, ShouldNotBeNil)
				So(result.PromptTokens, ShouldEqual, 100)
				So(result.CompletionTokens, ShouldEqual, 50)
				So(result.TotalTokens, ShouldEqual, 150)
				So(result.PromptTokensDetails.CachedTokens, ShouldEqual, 20)
				So(result.CompletionTokensDetails.ReasoningTokens, ShouldEqual, 10)
			})
		})

		Convey("When usage has zero values", func() {
			u := &v1.Statistics_Usage{}
			result := convertUsageToOpenAIChat(u)

			Convey("Then all fields should be zero", func() {
				So(result, ShouldNotBeNil)
				So(result.PromptTokens, ShouldEqual, 0)
				So(result.CompletionTokens, ShouldEqual, 0)
				So(result.TotalTokens, ShouldEqual, 0)
			})
		})
	})
}

func TestConvertChatRespToOpenAIChat(t *testing.T) {
	Convey("Given an internal ChatResp to convert to OpenAI format", t, func() {

		Convey("When converting a response with text content", func() {
			resp := &v1.ChatResp{
				Id:     "chatcmpl-1",
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
			result := convertChatRespToOpenAIChat(resp)

			Convey("Then the response should have correct fields", func() {
				So(result.ID, ShouldEqual, "chatcmpl-1")
				So(result.Object, ShouldEqual, "chat.completion")
				So(result.Model, ShouldEqual, "gpt-4o")
				So(result.Choices, ShouldHaveLength, 1)
				So(result.Choices[0].Message.Role, ShouldEqual, "assistant")
				So(result.Choices[0].Message.Content, ShouldEqual, "Hello!")
				So(result.Choices[0].FinishReason, ShouldEqual, "stop")
				So(result.Usage, ShouldNotBeNil)
				So(result.Usage.PromptTokens, ShouldEqual, 10)
				So(result.Usage.CompletionTokens, ShouldEqual, 5)
			})
		})

		Convey("When converting a response with reasoning content", func() {
			resp := &v1.ChatResp{
				Id:     "chatcmpl-2",
				Model:  "o3",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Reasoning: true,
							Content:   &v1.Content_Text{Text: "Let me think..."},
						},
						{
							Content: &v1.Content_Text{Text: "The answer is 42."},
						},
					},
				},
			}
			result := convertChatRespToOpenAIChat(resp)

			Convey("Then reasoning content should be in reasoning_content field", func() {
				So(result.Choices, ShouldHaveLength, 1)
				So(result.Choices[0].Message.ReasoningContent, ShouldEqual, "Let me think...")
				So(result.Choices[0].Message.Content, ShouldEqual, "The answer is 42.")
			})
		})

		Convey("When converting a response with tool calls", func() {
			resp := &v1.ChatResp{
				Id:     "chatcmpl-3",
				Model:  "gpt-4o",
				Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_ToolUse{
							ToolUse: &v1.ToolUse{
								Id:   "call-1",
								Name: "get_weather",
								Inputs: []*v1.ToolUse_Input{{
									Input: &v1.ToolUse_Input_Text{Text: `{"city":"Shanghai"}`},
								}},
							},
						}},
					},
				},
			}
			result := convertChatRespToOpenAIChat(resp)

			Convey("Then tool calls should be in the response", func() {
				So(result.Choices, ShouldHaveLength, 1)
				So(result.Choices[0].FinishReason, ShouldEqual, "tool_calls")
				So(result.Choices[0].Message.ToolCalls, ShouldHaveLength, 1)
				So(result.Choices[0].Message.ToolCalls[0].ID, ShouldEqual, "call-1")
				So(result.Choices[0].Message.ToolCalls[0].Type, ShouldEqual, "function")
				So(result.Choices[0].Message.ToolCalls[0].Function.Name, ShouldEqual, "get_weather")
				So(result.Choices[0].Message.ToolCalls[0].Function.Arguments, ShouldEqual, `{"city":"Shanghai"}`)
			})
		})

		Convey("When converting a response with mixed text and tool calls", func() {
			resp := &v1.ChatResp{
				Id:     "chatcmpl-4",
				Model:  "gpt-4o",
				Status: v1.ChatStatus_CHAT_PENDING_TOOL_USE,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: "I'll check."}},
						{Content: &v1.Content_ToolUse{
							ToolUse: &v1.ToolUse{
								Id:   "call-1",
								Name: "search",
								Inputs: []*v1.ToolUse_Input{{
									Input: &v1.ToolUse_Input_Text{Text: `{"q":"test"}`},
								}},
							},
						}},
					},
				},
			}
			result := convertChatRespToOpenAIChat(resp)

			Convey("Then both text and tool calls should be present", func() {
				So(result.Choices[0].Message.Content, ShouldEqual, "I'll check.")
				So(result.Choices[0].Message.ToolCalls, ShouldHaveLength, 1)
			})
		})

		Convey("When message is nil", func() {
			resp := &v1.ChatResp{
				Id:    "chatcmpl-5",
				Model: "gpt-4",
			}
			result := convertChatRespToOpenAIChat(resp)

			Convey("Then choices should be empty", func() {
				So(result.Choices, ShouldBeEmpty)
			})
		})

		Convey("When statistics are nil", func() {
			resp := &v1.ChatResp{
				Id:     "chatcmpl-6",
				Model:  "gpt-4",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: "Ok"}},
					},
				},
			}
			result := convertChatRespToOpenAIChat(resp)

			Convey("Then usage should be nil", func() {
				So(result.Usage, ShouldBeNil)
			})
		})

		Convey("When text content is concatenated from multiple parts", func() {
			resp := &v1.ChatResp{
				Id:     "chatcmpl-7",
				Model:  "gpt-4",
				Status: v1.ChatStatus_CHAT_COMPLETED,
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: "Part 1 "}},
						{Content: &v1.Content_Text{Text: "Part 2"}},
					},
				},
			}
			result := convertChatRespToOpenAIChat(resp)

			Convey("Then text parts should be concatenated", func() {
				So(result.Choices[0].Message.Content, ShouldEqual, "Part 1 Part 2")
			})
		})
	})
}

func TestConvertEmbeddingReqFromOpenAIChat(t *testing.T) {
	Convey("Given an OpenAI EmbeddingNewParams to convert", t, func() {

		Convey("When input is a single string", func() {
			req := &openai.EmbeddingNewParams{
				Model: "text-embedding-3-small",
				Input: openai.EmbeddingNewParamsInputUnion{
					OfString: openai.Opt("Hello world"),
				},
			}
			result := convertEmbeddingReqFromOpenAIChat(req)

			Convey("Then it should create an EmbedReq with text content", func() {
				So(result.Model, ShouldEqual, "text-embedding-3-small")
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "Hello world")
			})
		})

		Convey("When input is an array of strings", func() {
			req := &openai.EmbeddingNewParams{
				Model: "text-embedding-3-large",
				Input: openai.EmbeddingNewParamsInputUnion{
					OfArrayOfStrings: []string{"first", "second"},
				},
			}
			result := convertEmbeddingReqFromOpenAIChat(req)

			Convey("Then it should use the first string", func() {
				So(result.Model, ShouldEqual, "text-embedding-3-large")
				So(result.Contents, ShouldHaveLength, 1)
				So(result.Contents[0].GetText(), ShouldEqual, "first")
			})
		})

		Convey("When input is an empty array of strings", func() {
			req := &openai.EmbeddingNewParams{
				Model: "text-embedding-3-small",
				Input: openai.EmbeddingNewParamsInputUnion{
					OfArrayOfStrings: []string{},
				},
			}
			result := convertEmbeddingReqFromOpenAIChat(req)

			Convey("Then contents should be empty", func() {
				So(result.Contents, ShouldBeEmpty)
			})
		})

		Convey("When neither string nor array is set", func() {
			req := &openai.EmbeddingNewParams{
				Model: "text-embedding-3-small",
			}
			result := convertEmbeddingReqFromOpenAIChat(req)

			Convey("Then contents should be empty", func() {
				So(result.Contents, ShouldBeEmpty)
			})
		})
	})
}

func TestConvertEmbeddingRespToOpenAIChat(t *testing.T) {
	Convey("Given an internal EmbedResp to convert", t, func() {

		Convey("When embedding has values", func() {
			resp := &v1.EmbedResp{
				Model:     "text-embedding-3-small",
				Embedding: []float32{0.1, 0.2, 0.3, -0.5},
			}
			result := convertEmbeddingRespToOpenAIChat(resp)

			Convey("Then the response should have correct fields", func() {
				So(result.Object, ShouldEqual, "list")
				So(result.Model, ShouldEqual, "text-embedding-3-small")
				So(result.Data, ShouldHaveLength, 1)
				So(result.Data[0].Index, ShouldEqual, 0)
				So(result.Data[0].Embedding, ShouldHaveLength, 4)
				So(result.Data[0].Embedding[0], ShouldAlmostEqual, 0.1, 0.001)
				So(result.Data[0].Embedding[1], ShouldAlmostEqual, 0.2, 0.001)
				So(result.Data[0].Embedding[2], ShouldAlmostEqual, 0.3, 0.001)
				So(result.Data[0].Embedding[3], ShouldAlmostEqual, -0.5, 0.001)
			})
		})

		Convey("When embedding is empty", func() {
			resp := &v1.EmbedResp{
				Model:     "text-embedding-3-small",
				Embedding: []float32{},
			}
			result := convertEmbeddingRespToOpenAIChat(resp)

			Convey("Then the embedding should be empty", func() {
				So(result.Data, ShouldHaveLength, 1)
				So(result.Data[0].Embedding, ShouldBeEmpty)
			})
		})
	})
}
