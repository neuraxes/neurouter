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

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
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
	})
}
