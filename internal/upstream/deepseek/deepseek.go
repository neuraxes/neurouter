package deepseek

import (
	"context"
	"io"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/openai/openai-go/packages/ssestream"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type ChatRepo struct {
	config *conf.DeepSeekConfig
	log    *log.Helper
}

func NewDeepSeekChatRepoFactory() biz.DeepSeekChatRepoFactory {
	return NewDeepSeekChatRepo
}

func NewDeepSeekChatRepo(config *conf.DeepSeekConfig, logger log.Logger) (biz.ChatRepo, error) {
	// Trim the trailing slash from the base URL to avoid double slashes
	config.BaseUrl = strings.TrimSuffix(config.BaseUrl, "/")

	return &ChatRepo{
		config: config,
		log:    log.NewHelper(logger),
	}, nil
}

func (r *ChatRepo) Chat(ctx context.Context, req *biz.ChatReq) (resp *biz.ChatResp, err error) {
	res, err := r.CreateChatCompletion(
		ctx,
		r.convertRequestToDeepSeek(req),
	)
	if err != nil {
		return
	}

	resp = &biz.ChatResp{
		Id:      res.ID,
		Message: r.convertMessageFromDeepSeek(res.Choices[0].Message),
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

type deepSeekChatStreamClient struct {
	id       string
	req      *biz.ChatReq
	upstream *ssestream.Stream[ChatStreamResponse] // Reuse SSE Stream from OpenAI
}

func (c *deepSeekChatStreamClient) Recv() (resp *biz.ChatResp, err error) {
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
	}

	if len(chunk.Choices) > 0 {
		resp.Message = &v1.Message{
			Id:   c.id,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: chunk.Choices[0].Delta.Content,
					},
				},
			},
			ReasoningContent: chunk.Choices[0].Delta.ReasoningContent,
		}
	}

	if chunk.Usage != nil && (chunk.Usage.PromptTokens != 0 || chunk.Usage.CompletionTokens != 0) {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(chunk.Usage.PromptTokens),
				CompletionTokens: int32(chunk.Usage.CompletionTokens),
			},
		}
	}

	// Clear due to the reuse of the same message struct
	chunk.Choices[0].Delta = nil
	return
}

func (c *deepSeekChatStreamClient) Close() error {
	return c.upstream.Close()
}

func (r *ChatRepo) ChatStream(ctx context.Context, req *biz.ChatReq) (client biz.ChatStreamClient, err error) {
	deepSeekReq := r.convertRequestToDeepSeek(req)
	deepSeekReq.Stream = true
	deepSeekReq.StreamOptions = &StreamOptions{
		IncludeUsage: true,
	}

	resp, err := r.CreateChatCompletionStream(ctx, deepSeekReq)
	if err != nil {
		return
	}

	id, err := uuid.NewUUID()
	if err != nil {
		_ = resp.Body.Close()
		return
	}

	client = &deepSeekChatStreamClient{
		id:       id.String(),
		req:      req,
		upstream: ssestream.NewStream[ChatStreamResponse](ssestream.NewDecoder(resp), err),
	}
	return
}
