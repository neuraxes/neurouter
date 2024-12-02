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

type upstream struct {
	models []*conf.Model
	repo   ChatCompletionRepo
}

type chatUseCase struct {
	upstreams []*upstream
	log       *log.Helper
}

func NewChatUseCase(
	c *conf.Upstream,
	openAIChatCompletionRepoFactory OpenAIChatCompletionRepoFactory,
	logger log.Logger,
) ChatUseCase {
	var upstreams []*upstream

	if c != nil {
		for _, config := range c.Configs {
			var repo ChatCompletionRepo

			switch config.GetConfig().(type) {
			case *conf.UpstreamConfig_Laas:
				panic("unimplemented")
			case *conf.UpstreamConfig_Openai:
				repo = openAIChatCompletionRepoFactory(config.GetOpenai())
			case *conf.UpstreamConfig_Google:
				panic("unimplemented")
			}

			upstreams = append(upstreams, &upstream{models: config.Models, repo: repo})
		}
	}

	return &chatUseCase{
		upstreams: upstreams,
		log:       log.NewHelper(logger),
	}
}

// choose select the upstream and model by req
func (uc *chatUseCase) choose(req *ChatReq) (up *upstream, model *conf.Model, err error) {
	for _, u := range uc.upstreams {
		// TODO: select upstream by req
		up = u
		model = u.models[0]
		return
	}
	err = v1.ErrorNoUpstream("no upstream found")
	return
}

func (uc *chatUseCase) Chat(ctx context.Context, req *ChatReq) (resp *ChatResp, err error) {
	u, m, err := uc.choose(req)
	if err != nil {
		return
	}
	req.Model = m.Id
	return u.repo.Chat(ctx, req)
}

func (uc *chatUseCase) ChatStream(ctx context.Context, req *ChatReq, server ChatStreamServer) error {
	u, m, err := uc.choose(req)
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

	req.Model = m.Id
	client, err := u.repo.ChatStream(ctx, req)
	if err != nil {
		return err
	}

	for {
		resp, err := client.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			// TODO: Close upstream
			return nil
		}

		err = server.Send(resp)
		if err != nil {
			return err
		}
	}
}
