package data

import (
	"testing"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"git.xdea.xyz/Turing/router/internal/conf"
	"github.com/openai/openai-go"
	. "github.com/smartystreets/goconvey/convey"
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
	})
}
