package data

import (
	"context"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"git.xdea.xyz/Turing/router/internal/biz"
	"git.xdea.xyz/Turing/router/internal/conf"
)

type OpenAIChatCompletionRepo struct {
	client *openai.Client
}

func NewOpenAIChatCompletionRepoFactory() biz.OpenAIChatCompletionRepoFactory {
	return func(config *conf.OpenAIConfig) biz.ChatCompletionRepo {
		return NewOpenAIChatCompletionRepo(config.ApiKey, config.BaseUrl)
	}
}

func NewOpenAIChatCompletionRepo(apiKey, baseUrl string) biz.ChatCompletionRepo {
	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if baseUrl != "" {
		options = append(options, option.WithBaseURL(baseUrl))
	}

	return &OpenAIChatCompletionRepo{
		client: openai.NewClient(options...),
	}
}

func (r *OpenAIChatCompletionRepo) Chat(
	ctx context.Context,
	req *biz.ChatReq,
) (resp *biz.ChatResp, err error) {
	var messages []openai.ChatCompletionMessageParamUnion
	for _, message := range req.Messages {
		messages = append(messages, r.convertMessageToOpenAI(message))
	}

	res, err := r.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model:    openai.F(req.Model),
			Messages: openai.F(messages),
		},
	)
	if err != nil {
		return
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return
	}

	resp = &biz.ChatResp{
		Message: &v1.Message{
			Id:   id.String(),
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						// The result contains a leading space, so we need to trim it
						Text: strings.TrimSpace(res.Choices[0].Message.Content),
					},
				},
			},
		},
	}
	return
}

type openaiChatStreamClient struct {
	upstream *ssestream.Stream[openai.ChatCompletionChunk]
}

func (c openaiChatStreamClient) Recv() (resp *biz.ChatResp, err error) {
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}
	chunk := c.upstream.Current()
	resp = &biz.ChatResp{
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: chunk.Choices[0].Delta.Content,
					},
				},
			},
		},
	}
	return
}

func (r *OpenAIChatCompletionRepo) ChatStream(
	ctx context.Context,
	req *biz.ChatReq,
) (client biz.ChatStreamClient, err error) {
	var messages []openai.ChatCompletionMessageParamUnion
	for _, message := range req.Messages {
		messages = append(messages, r.convertMessageToOpenAI(message))
	}

	stream := r.client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Model:    openai.F(req.Model),
			Messages: openai.F(messages),
		},
	)

	client = &openaiChatStreamClient{
		upstream: stream,
	}
	return
}

func (r *OpenAIChatCompletionRepo) convertMessageToOpenAI(message *v1.Message) openai.ChatCompletionMessageParamUnion {
	if message.Role == v1.Role_SYSTEM {
		return openai.SystemMessage(message.Contents[0].GetText())
	} else if message.Role == v1.Role_USER {
		var parts []openai.ChatCompletionContentPartUnionParam
		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_Text:
				parts = append(parts, openai.TextPart(c.Text))
			}
		}
		return openai.UserMessageParts(parts...)
	} else {
		return openai.AssistantMessage(message.Contents[0].GetText())
	}
}
