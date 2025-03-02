package biz

import (
	"context"
	"errors"
	"io"

	"github.com/go-kratos/kratos/v2/log"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type ChatUseCase interface {
	Chat(ctx context.Context, req *ChatReq) (*ChatResp, error)
	ChatStream(ctx context.Context, req *ChatReq, stream ChatStreamServer) error
	ListAvailableModels(ctx context.Context) ([]*ModelSpec, error)
}

type upstream struct {
	models []*conf.Model
	repo   ChatRepo
}

type chatUseCase struct {
	upstreams []*upstream
	log       *log.Helper
}

func NewChatUseCase(
	c *conf.Upstream,
	neurouterChatRepoFactory NeurouterChatRepoFactory,
	openAIChatRepoFactory OpenAIChatRepoFactory,
	anthropicChatRepoFactory AnthropicChatRepoFactory,
	deepSeekChatRepoFactory DeepSeekChatRepoFactory,
	logger log.Logger,
) ChatUseCase {
	logHelper := log.NewHelper(logger)
	var upstreams []*upstream

	if c != nil {
		for _, config := range c.Configs {
			var (
				repo ChatRepo
				err  error
			)

			switch config.GetConfig().(type) {
			case *conf.UpstreamConfig_Neurouter:
				repo, err = neurouterChatRepoFactory(config.GetNeurouter(), logger)
			case *conf.UpstreamConfig_OpenAi:
				repo, err = openAIChatRepoFactory(config.GetOpenAi(), logger)
			case *conf.UpstreamConfig_Google:
				panic("unimplemented")
			case *conf.UpstreamConfig_Anthropic:
				repo, err = anthropicChatRepoFactory(config.GetAnthropic(), logger)
			case *conf.UpstreamConfig_DeepSeek:
				repo, err = deepSeekChatRepoFactory(config.GetDeepSeek(), logger)
			}

			if err != nil {
				logHelper.Errorf("failed to create chat repo: %v", err)
				continue
			}

			upstreams = append(upstreams, &upstream{models: config.Models, repo: repo})
		}
	}

	return &chatUseCase{
		upstreams: upstreams,
		log:       logHelper,
	}
}

// choose select the upstream and model by req
func (uc *chatUseCase) choose(req *ChatReq) (up *upstream, model *conf.Model, err error) {
	for _, u := range uc.upstreams {
		for _, m := range u.models {
			if m.Name == req.Model {
				up = u
				model = m
				uc.log.Infof("using model: %s", m.Name)
				return
			}
		}
	}

	for _, u := range uc.upstreams {
		up = u
		model = u.models[0]
		uc.log.Infof("fallback to model: %s", model.Name)
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

	req.Model = m.Id
	client, err := u.repo.ChatStream(ctx, req)
	if err != nil {
		return err
	}
	defer client.Close()

	for {
		resp, err := client.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}

		err = server.Send(resp)
		if err != nil {
			return err
		}
	}
}

func (uc *chatUseCase) ListAvailableModels(ctx context.Context) ([]*ModelSpec, error) {
	var models []*ModelSpec

	for _, u := range uc.upstreams {
		for _, m := range u.models {
			models = append(models, &ModelSpec{
				Id:       m.Id,
				Name:     m.Name,
				Provider: m.Provider,
			})
		}
	}

	return models, nil
}
