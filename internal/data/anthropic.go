package data

import (
	"context"
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/google/uuid"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"git.xdea.xyz/Turing/router/internal/biz"
	"git.xdea.xyz/Turing/router/internal/conf"
)

type AnthropicChatRepo struct {
	config *conf.AnthropicConfig
	client *anthropic.Client
}

func NewAnthropicChatRepoFactory() biz.AnthropicChatRepoFactory {
	return func(config *conf.AnthropicConfig) biz.ChatRepo {
		return NewAnthropicChatRepo(config)
	}
}

func NewAnthropicChatRepo(config *conf.AnthropicConfig) biz.ChatRepo {
	options := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
	}

	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}

	return &AnthropicChatRepo{
		config: config,
		client: anthropic.NewClient(options...),
	}
}

// convertSystemToAnthropic converts system messages to a format that can be sent to the Anthropic API.
func (r *AnthropicChatRepo) convertSystemToAnthropic(messages []*v1.Message) []anthropic.TextBlockParam {
	var parts []anthropic.TextBlockParam
	for _, message := range messages {
		if message.Role != v1.Role_SYSTEM {
			continue
		}
		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_Text:
				parts = append(parts, anthropic.NewTextBlock(c.Text))
			}
		}
	}
	return parts
}

// convertMessageToAnthropic converts an internal message to a message that can be sent to the Anthropic API.
func (r *AnthropicChatRepo) convertMessageToAnthropic(message *v1.Message) anthropic.MessageParam {
	var parts []anthropic.ContentBlockParamUnion
	for _, content := range message.Contents {
		switch c := content.GetContent().(type) {
		case *v1.Content_Text:
			parts = append(parts, anthropic.NewTextBlock(c.Text))
		case *v1.Content_ImageUrl:
			// TODO: Implement image support
		}
	}
	if message.Role == v1.Role_USER || message.Role == v1.Role_SYSTEM {
		return anthropic.NewUserMessage(parts...)
	} else {
		return anthropic.NewAssistantMessage(anthropic.NewTextBlock(message.Contents[0].GetText()))
	}
}

// convertRequestToAnthropic converts an internal request to a request that can be sent to the Anthropic API.
func (r *AnthropicChatRepo) convertRequestToAnthropic(req *biz.ChatReq) anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model: anthropic.F(req.Model),
	}

	if !r.config.MergeSystem {
		params.System = anthropic.F(r.convertSystemToAnthropic(req.Messages))
	}

	var messages []anthropic.MessageParam
	for _, message := range req.Messages {
		if !r.config.MergeSystem && message.Role == v1.Role_SYSTEM {
			continue
		}
		messages = append(messages, r.convertMessageToAnthropic(message))
	}
	params.Messages = anthropic.F(messages)

	return params
}

func (r *AnthropicChatRepo) Chat(ctx context.Context, req *biz.ChatReq) (resp *biz.ChatResp, err error) {
	res, err := r.client.Messages.New(
		ctx,
		r.convertRequestToAnthropic(req),
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
						Text: res.Content[0].Text,
					},
				},
			},
		},
	}

	if res.Usage.InputTokens != 0 || res.Usage.OutputTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(res.Usage.InputTokens),
				CompletionTokens: int32(res.Usage.OutputTokens),
			},
		}
	}
	return
}

type anthropicChatStreamClient struct {
	id       string
	req      *biz.ChatReq
	upstream *ssestream.Stream[anthropic.MessageStreamEvent]
}

func (c anthropicChatStreamClient) Recv() (resp *biz.ChatResp, err error) {
next:
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}

	chunk := c.upstream.Current()
	if chunk.Type != anthropic.MessageStreamEventTypeContentBlockDelta {
		goto next
	}

	resp = &biz.ChatResp{
		Id: c.req.Id,
		Message: &v1.Message{
			Id:   c.id,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: chunk.Delta.(anthropic.ContentBlockDeltaEventDelta).Text,
					},
				},
			},
		},
	}

	if chunk.Usage.OutputTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				CompletionTokens: int32(chunk.Usage.OutputTokens),
			},
		}
	}
	return
}

func (r *AnthropicChatRepo) ChatStream(ctx context.Context, req *biz.ChatReq) (client biz.ChatStreamClient, err error) {
	stream := r.client.Messages.NewStreaming(
		ctx,
		r.convertRequestToAnthropic(req),
	)

	id, err := uuid.NewUUID()
	if err != nil {
		return
	}

	client = &anthropicChatStreamClient{
		id:       id.String(),
		req:      req,
		upstream: stream,
	}
	return
}
