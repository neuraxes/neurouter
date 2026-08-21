// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/go-kratos/kratos/v3/config"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type UseCase interface {
	ListAvailableModels(ctx context.Context) ([]*entity.ModelSpec, error)
}

type model struct {
	config            *conf.Model
	upstreamConfig    *conf.UpstreamConfig
	chatRepo          repository.ChatRepo
	embeddingRepo     repository.EmbeddingRepo
	inputTokens       atomic.Int64
	outputTokens      atomic.Int64
	cachedInputTokens atomic.Int64
	reasoningTokens   atomic.Int64
	upstreamLimiters  *limiterGroup // shared across models in same upstream
	modelLimiters     *limiterGroup // specific to this model
}

type alias struct {
	config *conf.AliasConfig
	models []*model
}

type UseCaseImpl struct {
	models  []*model
	aliases map[string]*alias
	log     *slog.Logger
}

func NewModelUseCase(
	c config.Config,
	anthropicFactory repository.UpstreamFactory[conf.AnthropicConfig],
	googleFactory repository.UpstreamFactory[conf.GoogleConfig],
	neurouterFactory repository.UpstreamFactory[conf.NeurouterConfig],
	openAIFactory repository.UpstreamFactory[conf.OpenAIConfig],
	logger *slog.Logger,
) *UseCaseImpl {
	var models []*model
	aliases := make(map[string]*alias)

	upstream, err := config.Get[conf.Upstream](c, "upstream")
	if err == nil {
		for _, upstreamConfig := range upstream.Configs {
			var (
				repo repository.Repo
				err  error
			)

			switch upstreamConfig.GetConfig().(type) {
			case *conf.UpstreamConfig_Neurouter:
				repo, err = neurouterFactory(upstreamConfig.GetNeurouter(), logger)
			case *conf.UpstreamConfig_OpenAi:
				repo, err = openAIFactory(upstreamConfig.GetOpenAi(), logger)
			case *conf.UpstreamConfig_Google:
				repo, err = googleFactory(upstreamConfig.GetGoogle(), logger)
			case *conf.UpstreamConfig_Anthropic:
				repo, err = anthropicFactory(upstreamConfig.GetAnthropic(), logger)
			}

			if err != nil {
				logger.Error("failed to create upstream repository", "error", err, "upstream", upstreamConfig.Name)
				continue
			}

			// Create upstream limiter group once (shared across all models in this upstream)
			us := upstreamConfig.GetScheduling()
			upstreamLimiters := newLimiterGroup(
				us.GetConcurrencyLimit(),
				us.GetRpmLimit(),
				us.GetRpdLimit(),
				us.GetTpmLimit(),
				us.GetTpdLimit(),
			)

			for _, modelConfig := range upstreamConfig.GetModels() {
				chatRepo, _ := repo.(repository.ChatRepo)
				embeddingRepo, _ := repo.(repository.EmbeddingRepo)

				// Create model limiter group (specific to this model)
				ms := modelConfig.GetScheduling()
				modelLimiters := newLimiterGroup(
					ms.GetConcurrencyLimit(),
					ms.GetRpmLimit(),
					ms.GetRpdLimit(),
					ms.GetTpmLimit(),
					ms.GetTpdLimit(),
				)

				models = append(models, &model{
					config:           modelConfig,
					upstreamConfig:   upstreamConfig,
					chatRepo:         chatRepo,
					embeddingRepo:    embeddingRepo,
					upstreamLimiters: upstreamLimiters,
					modelLimiters:    modelLimiters,
				})
			}
		}

		for _, ac := range upstream.GetAliases() {
			actual := ac.GetActual()
			if actual == nil {
				logger.Error("alias is missing actual config", "alias", ac.GetId())
				continue
			}
			var resolved []*model
			for _, m := range models {
				if m.config.Id != actual.GetModel() {
					continue
				}
				if upstream := actual.GetUpstream(); upstream != "" && m.upstreamConfig.Name != upstream {
					continue
				}
				resolved = append(resolved, m)
			}
			if len(resolved) == 0 {
				logger.Error(
					"alias target model not found",
					"alias", ac.GetId(),
					"upstream", actual.GetUpstream(),
					"model", actual.GetModel(),
				)
				continue
			}
			aliases[ac.GetId()] = &alias{config: ac, models: resolved}
		}
	}

	return &UseCaseImpl{
		models:  models,
		aliases: aliases,
		log:     logger,
	}
}

func (uc *UseCaseImpl) ListAvailableModels(ctx context.Context) ([]*entity.ModelSpec, error) {
	var specs []*entity.ModelSpec

	for _, m := range uc.models {
		specs = append(specs, convertModelConfigToSpec(m.config))
	}

	// Add virtual models from aliases
	for _, a := range uc.aliases {
		actual := a.models[0]
		spec := convertModelConfigToSpec(actual.config)
		spec.Id = a.config.Id
		if a.config.Name != "" {
			spec.Name = a.config.Name
		}
		specs = append(specs, spec)
	}

	return specs, nil
}

func convertModelConfigToSpec(cfg *conf.Model) *entity.ModelSpec {
	modalities := make([]v1.Modality, 0, len(cfg.Modalities))
	for _, modality := range cfg.Modalities {
		modalities = append(modalities, v1.Modality(modality))
	}
	capabilities := make([]v1.Capability, 0, len(cfg.Capabilities))
	for _, capability := range cfg.Capabilities {
		capabilities = append(capabilities, v1.Capability(capability))
	}
	return &entity.ModelSpec{
		Id:            cfg.Id,
		Name:          cfg.Name,
		Owner:         cfg.Owner,
		Provider:      cfg.Provider,
		Modalities:    modalities,
		Capabilities:  capabilities,
		ContextLength: cfg.ContextLength,
	}
}
