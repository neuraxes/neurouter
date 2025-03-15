package openai

import (
	"context"
	"io"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz/entity"
	"git.xdea.xyz/Turing/neurouter/internal/biz/repository"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type ChatRepo struct {
	config *conf.OpenAIConfig
	client *openai.Client
	log    *log.Helper
}

func NewOpenAIChatRepoFactory() repository.ChatRepoFactory[conf.OpenAIConfig] {
	return NewOpenAIChatRepo
}

func NewOpenAIChatRepo(config *conf.OpenAIConfig, logger log.Logger) (repository.ChatRepo, error) {
	repo := &ChatRepo{
		config: config,
		log:    log.NewHelper(logger),
	}

	options := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
	}
	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}
	if config.PreferStringContentForSystem ||
		config.PreferStringContentForUser ||
		config.PreferStringContentForAssistant ||
		config.PreferStringContentForTool {
		options = append(options, option.WithMiddleware(repo.preferStringContent))
	}
	repo.client = openai.NewClient(options...)

	return repo, nil
}

func (r *ChatRepo) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	res, err := r.client.Chat.Completions.New(
		ctx,
		r.convertRequestToOpenAI(req),
	)
	if err != nil {
		return
	}

	resp = &entity.ChatResp{
		Id:      res.ID,
		Message: r.convertMessageFromOpenAI(&res.Choices[0].Message),
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

type openAIChatStreamClient struct {
	id       string
	req      *entity.ChatReq
	upstream *ssestream.Stream[openai.ChatCompletionChunk]
}

func (c openAIChatStreamClient) Recv() (resp *entity.ChatResp, err error) {
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}

	chunk := c.upstream.Current()
	resp = &entity.ChatResp{
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
		}
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

func (c openAIChatStreamClient) Close() error {
	return c.upstream.Close()
}

func (r *ChatRepo) ChatStream(ctx context.Context, req *entity.ChatReq) (client repository.ChatStreamClient, err error) {
	openAIReq := r.convertRequestToOpenAI(req)
	openAIReq.StreamOptions = openai.F(openai.ChatCompletionStreamOptionsParam{
		IncludeUsage: openai.F(true),
	})
	stream := r.client.Chat.Completions.NewStreaming(
		ctx,
		openAIReq,
	)

	id, err := uuid.NewUUID()
	if err != nil {
		return
	}

	client = &openAIChatStreamClient{
		id:       id.String(),
		req:      req,
		upstream: stream,
	}
	return
}
