package biz

import (
	"context"
	"errors"
	"io"

	"github.com/go-kratos/kratos/v2/log"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"git.xdea.xyz/Turing/router/internal/conf"
)

type ChatUseCase interface {
	Chat(ctx context.Context, req *ChatReq) (*ChatResp, error)
	ChatStream(ctx context.Context, req *ChatReq, stream ChatStreamServer) error
}

type chatUseCase struct {
	upstreams map[string]ChatCompletionRepo
	log       *log.Helper
}

func NewChatUseCase(
	c *conf.Upstream,
	openAIChatCompletionRepoFactory OpenAIChatCompletionRepoFactory,
	logger log.Logger,
) ChatUseCase {
	upstreams := map[string]ChatCompletionRepo{}

	if c != nil {
		for _, config := range c.Configs {
			switch config.GetConfig().(type) {
			case *conf.UpstreamConfig_Laas:
				panic("unimplemented")
			case *conf.UpstreamConfig_Openai:
				upstreams[config.GetName()] = openAIChatCompletionRepoFactory(config.GetOpenai())
			case *conf.UpstreamConfig_Google:
				panic("unimplemented")
			}
		}
	}

	return &chatUseCase{
		upstreams: upstreams,
		log:       log.NewHelper(logger),
	}
}

func (uc *chatUseCase) selectUpstream(req *ChatReq) (repo ChatCompletionRepo, err error) {
	for _, upstream := range uc.upstreams {
		// TODO: select upstream by req
		repo = upstream
		return
	}
	err = v1.ErrorNoUpstream("no upstream found")
	return
}

func (uc *chatUseCase) Chat(ctx context.Context, req *ChatReq) (resp *ChatResp, err error) {
	repo, err := uc.selectUpstream(req)
	if err != nil {
		return
	}
	return repo.Chat(ctx, req)
}

func (uc *chatUseCase) ChatStream(ctx context.Context, req *ChatReq, stream ChatStreamServer) error {
	repo, err := uc.selectUpstream(req)
	if err != nil {
		return err
	}

	var messages []*v1.Message
	for _, msg := range req.Messages {
		if msg.Role == v1.Role_SYSTEM {
			continue
		}
		messages = append(messages, msg)
	}
	req.Messages = messages

	upstream, err := repo.ChatStream(ctx, req)
	if err != nil {
		return err
	}
	
	for {
		resp, err := upstream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			// TODO: Close upstream
			return nil
		}

		err = stream.Send(resp)
		if err != nil {
			return err
		}
	}
}
