package model

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz/entity"
	"git.xdea.xyz/Turing/neurouter/internal/biz/repository"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type UseCase interface {
	ListAvailableModels(ctx context.Context) ([]*entity.ModelSpec, error)
}

type model struct {
	config *conf.Model
	repo   repository.ChatRepo
}

type UseCaseImpl struct {
	models []*model
	log    *log.Helper
}

func NewModelUseCase(
	c *conf.Upstream,
	neurouterChatRepoFactory repository.ChatRepoFactory[conf.NeurouterConfig],
	openAIChatRepoFactory repository.ChatRepoFactory[conf.OpenAIConfig],
	anthropicChatRepoFactory repository.ChatRepoFactory[conf.AnthropicConfig],
	deepSeekChatRepoFactory repository.ChatRepoFactory[conf.DeepSeekConfig],
	logger log.Logger,
) *UseCaseImpl {
	logHelper := log.NewHelper(logger)
	var models []*model

	if c != nil {
		for _, config := range c.Configs {
			var (
				repo repository.ChatRepo
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

			for _, m := range config.GetModels() {
				models = append(models, &model{
					config: m,
					repo:   repo,
				})
			}
		}
	}

	return &UseCaseImpl{
		models: models,
		log:    log.NewHelper(logger),
	}
}

func (uc *UseCaseImpl) ElectForChat(uri string) (repo repository.ChatRepo, model *conf.Model, err error) {
	for _, m := range uc.models {
		if m.config.Name == uri {
			repo = m.repo
			model = m.config
			uc.log.Infof("using model: %s", m.config.Name)
			return
		}
	}

	for _, m := range uc.models {
		repo = m.repo
		model = m.config
		uc.log.Infof("fallback to model: %s", m.config.Name)
		return
	}

	err = v1.ErrorNoUpstream("no upstream found")
	return
}

func (uc *UseCaseImpl) ListAvailableModels(ctx context.Context) ([]*entity.ModelSpec, error) {
	var models []*entity.ModelSpec

	for _, m := range uc.models {
		models = append(models, &entity.ModelSpec{
			Id:       m.config.Id,
			Name:     m.config.Name,
			Provider: m.config.Provider,
		})
	}

	return models, nil
}
