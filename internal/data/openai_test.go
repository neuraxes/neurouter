package data

import (
	"testing"

	"github.com/openai/openai-go"
	. "github.com/smartystreets/goconvey/convey"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"git.xdea.xyz/Turing/router/internal/conf"
)

func TestConvertMessageToOpenAI(t *testing.T) {
	Convey("Given a OpenAIChatRepo with merge content enabled", t, func() {
		repo := &OpenAIChatRepo{
			config: &conf.OpenAIConfig{
				MergeContent: true,
			},
		}

		Convey("When the message role is USER and all contents are text", func() {
			message := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Hello"}},
					{Content: &v1.Content_Text{Text: " World"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.UserMessage("Hello World"))
		})

		Convey("When the message role is USER and contents are mixed", func() {
			message := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Hello"}},
					{Content: &v1.Content_ImageUrl{ImageUrl: "http://example.com/image.png"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.UserMessageParts(
				openai.TextPart("Hello"),
				openai.ImagePart("http://example.com/image.png"),
			))
		})
	})

	Convey("Given a OpenAIChatRepo with merge content disabled", t, func() {
		repo := &OpenAIChatRepo{
			config: &conf.OpenAIConfig{
				MergeContent: false,
			},
		}

		Convey("When the message role is USER and all contents are text", func() {
			message := &v1.Message{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Hello"}},
					{Content: &v1.Content_Text{Text: "World"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.UserMessageParts(
				openai.TextPart("Hello"),
				openai.TextPart("World"),
			))
		})
	})

	Convey("Given a OpenAIChatRepo", t, func() {
		repo := &OpenAIChatRepo{
			config: &conf.OpenAIConfig{},
		}

		Convey("When the message role is SYSTEM", func() {
			message := &v1.Message{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "System message"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.SystemMessage("System message"))
		})

		Convey("When the message role is SYSTEM with multiple text contents", func() {
			message := &v1.Message{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "System message part 1"}},
					{Content: &v1.Content_Text{Text: " and part 2"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.SystemMessage("System message part 1 and part 2"))
		})

		Convey("When the message role is MODEL", func() {
			message := &v1.Message{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Model message"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.AssistantMessage("Model message"))
		})

		Convey("When the message role is MODEL with multiple text contents", func() {
			message := &v1.Message{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Model message part 1"}},
					{Content: &v1.Content_Text{Text: " and part 2"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.AssistantMessage("Model message part 1 and part 2"))
		})

		Convey("When the message role is MODEL with tool calls", func() {
			message := &v1.Message{
				Role: v1.Role_MODEL,
				ToolCalls: []*v1.ToolCall{
					{
						Id: "tool-call-id",
						Tool: &v1.ToolCall_Function_{
							Function: &v1.ToolCall_Function{
								Name:      "function-name",
								Arguments: "function-arguments",
							},
						},
					},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.ChatCompletionAssistantMessageParam{
				Role: openai.F(openai.ChatCompletionAssistantMessageParamRoleAssistant),
				ToolCalls: openai.F([]openai.ChatCompletionMessageToolCallParam{
					{
						ID:   openai.F("tool-call-id"),
						Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
						Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      openai.F("function-name"),
							Arguments: openai.F("function-arguments"),
						}),
					},
				}),
			})
		})

		Convey("When the message role is MODEL with multiple tool calls", func() {
			message := &v1.Message{
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Model message"}},
				},
				ToolCalls: []*v1.ToolCall{
					{
						Id: "tool-call-id-1",
						Tool: &v1.ToolCall_Function_{
							Function: &v1.ToolCall_Function{
								Name:      "function-name-1",
								Arguments: "function-arguments-1",
							},
						},
					},
					{
						Id: "tool-call-id-2",
						Tool: &v1.ToolCall_Function_{
							Function: &v1.ToolCall_Function{
								Name:      "function-name-2",
								Arguments: "function-arguments-2",
							},
						},
					},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.ChatCompletionAssistantMessageParam{
				Role: openai.F(openai.ChatCompletionAssistantMessageParamRoleAssistant),
				Content: openai.F([]openai.ChatCompletionAssistantMessageParamContentUnion{
					openai.TextPart("Model message"),
				}),
				ToolCalls: openai.F([]openai.ChatCompletionMessageToolCallParam{
					{
						ID:   openai.F("tool-call-id-1"),
						Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
						Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      openai.F("function-name-1"),
							Arguments: openai.F("function-arguments-1"),
						}),
					},
					{
						ID:   openai.F("tool-call-id-2"),
						Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
						Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      openai.F("function-name-2"),
							Arguments: openai.F("function-arguments-2"),
						}),
					},
				}),
			})
		})

		Convey("When the message role is TOOL", func() {
			message := &v1.Message{
				Role:       v1.Role_TOOL,
				ToolCallId: "tool-call-id",
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Tool message"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.ToolMessage("tool-call-id", "Tool message"))
		})

		Convey("When the message role is TOOL with multiple text contents", func() {
			message := &v1.Message{
				Role:       v1.Role_TOOL,
				ToolCallId: "tool-call-id",
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Tool message part 1"}},
					{Content: &v1.Content_Text{Text: " and part 2"}},
				},
			}
			result := repo.convertMessageToOpenAI(message)
			So(result, ShouldResemble, openai.ToolMessage("tool-call-id", "Tool message part 1 and part 2"))
		})
	})
}
