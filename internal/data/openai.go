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

type OpenAIChatRepo struct {
	client *openai.Client
}

func NewOpenAIChatRepoFactory() biz.OpenAIChatRepoFactory {
	return func(config *conf.OpenAIConfig) biz.ChatRepo {
		return NewOpenAIChatRepo(config.ApiKey, config.BaseUrl)
	}
}

func NewOpenAIChatRepo(apiKey, baseUrl string) biz.ChatRepo {
	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if baseUrl != "" {
		options = append(options, option.WithBaseURL(baseUrl))
	}

	return &OpenAIChatRepo{
		client: openai.NewClient(options...),
	}
}

// convertMessageToOpenAI converts an internal message to a message that can be sent to the OpenAI API.
func (r *OpenAIChatRepo) convertMessageToOpenAI(message *v1.Message) openai.ChatCompletionMessageParamUnion {
	if message.Role == v1.Role_SYSTEM {
		return openai.SystemMessage(message.Contents[0].GetText())
	} else if message.Role == v1.Role_USER {
		var parts []openai.ChatCompletionContentPartUnionParam
		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_Text:
				parts = append(parts, openai.TextPart(c.Text))
			case *v1.Content_ImageUrl:
				parts = append(parts, openai.ImagePart(c.ImageUrl))
			}
		}
		return openai.UserMessageParts(parts...)
	} else {
		return openai.AssistantMessage(message.Contents[0].GetText())
	}
}

// convertRequestToOpenAI converts an internal request to a request that can be sent to the OpenAI API.
func (r *OpenAIChatRepo) convertRequestToOpenAI(req *biz.ChatReq) openai.ChatCompletionNewParams {
	var messages []openai.ChatCompletionMessageParamUnion
	for _, message := range req.Messages {
		messages = append(messages, r.convertMessageToOpenAI(message))
	}
	return openai.ChatCompletionNewParams{
		Model:    openai.F(req.Model),
		Messages: openai.F(messages),
	}
}

func (r *OpenAIChatRepo) Chat(ctx context.Context, req *biz.ChatReq) (resp *biz.ChatResp, err error) {
	res, err := r.client.Chat.Completions.New(
		ctx,
		r.convertRequestToOpenAI(req),
	)
	if err != nil {
		return
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return
	}

	resp = &biz.ChatResp{
		Id: req.Id,
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

	if res.Usage.PromptTokens != 0 || res.Usage.CompletionTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(res.Usage.PromptTokens),
				CompletionTokens: int32(res.Usage.CompletionTokens),
			},
		}
	}
	return
}

type openaiChatStreamClient struct {
	id       string
	req      *biz.ChatReq
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
		Id: c.req.Id,
		Message: &v1.Message{
			Id:   c.id,
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

	if chunk.Usage.PromptTokens != 0 || chunk.Usage.CompletionTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(chunk.Usage.PromptTokens),
				CompletionTokens: int32(chunk.Usage.CompletionTokens),
			},
		}
	}
	return
}

func (r *OpenAIChatRepo) ChatStream(ctx context.Context, req *biz.ChatReq) (client biz.ChatStreamClient, err error) {
	stream := r.client.Chat.Completions.NewStreaming(
		ctx,
		r.convertRequestToOpenAI(req),
	)

	id, err := uuid.NewUUID()
	if err != nil {
		return
	}

	client = &openaiChatStreamClient{
		id:       id.String(),
		req:      req,
		upstream: stream,
	}
	return
}
