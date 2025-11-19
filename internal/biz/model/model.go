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
	"sync/atomic"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/semaphore"

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
	inputTokens       atomic.Uint64
	outputTokens      atomic.Uint64
	cachedInputTokens atomic.Uint64
	upstreamSem       *semaphore.Weighted // nil means unlimited
	modelSem          *semaphore.Weighted // nil means unlimited
}

type UseCaseImpl struct {
	models []*model
	log    *log.Helper
}

func newSem(limit uint64) *semaphore.Weighted {
	if limit > 0 {
		return semaphore.NewWeighted(int64(limit))
	}
	return nil
}

func NewModelUseCase(
	c *conf.Upstream,
	anthropicFactory repository.UpstreamFactory[conf.AnthropicConfig],
	googleFactory repository.UpstreamFactory[conf.GoogleConfig],
	neurouterFactory repository.UpstreamFactory[conf.NeurouterConfig],
	openAIFactory repository.UpstreamFactory[conf.OpenAIConfig],
	logger log.Logger,
) *UseCaseImpl {
	logHelper := log.NewHelper(logger)
	var models []*model

	if c != nil {
		for _, config := range c.Configs {
			var (
				repo repository.Repo
				err  error
			)

			switch config.GetConfig().(type) {
			case *conf.UpstreamConfig_Neurouter:
				repo, err = neurouterFactory(config.GetNeurouter(), logger)
			case *conf.UpstreamConfig_OpenAi:
				repo, err = openAIFactory(config.GetOpenAi(), logger)
			case *conf.UpstreamConfig_Google:
				repo, err = googleFactory(config.GetGoogle(), logger)
			case *conf.UpstreamConfig_Anthropic:
				repo, err = anthropicFactory(config.GetAnthropic(), logger)
			}

			if err != nil {
				logHelper.Errorf("failed to create chat repo: %v", err)
				continue
			}

			// Create upstream semaphore once (shared across all models in this upstream)
			upstreamSem := newSem(config.GetScheduling().GetConcurrencyLimit())

			for _, modelConfig := range config.GetModels() {
				chatRepo, _ := repo.(repository.ChatRepo)
				embeddingRepo, _ := repo.(repository.EmbeddingRepo)

				// Create model semaphore (specific to this model)
				modelSem := newSem(modelConfig.GetScheduling().GetConcurrencyLimit())

				models = append(models, &model{
					config:         modelConfig,
					upstreamConfig: config,
					chatRepo:       chatRepo,
					embeddingRepo:  embeddingRepo,
					upstreamSem:    upstreamSem,
					modelSem:       modelSem,
				})
			}
		}
	}

	return &UseCaseImpl{
		models: models,
		log:    log.NewHelper(logger),
	}
}

func (uc *UseCaseImpl) ListAvailableModels(ctx context.Context) ([]*entity.ModelSpec, error) {
	var models []*entity.ModelSpec

	for _, m := range uc.models {
		modalities := make([]v1.Modality, 0, len(m.config.Modalities))
		for _, modality := range m.config.Modalities {
			modalities = append(modalities, v1.Modality(modality))
		}

		capabilities := make([]v1.Capability, 0, len(m.config.Capabilities))
		for _, capability := range m.config.Capabilities {
			capabilities = append(capabilities, v1.Capability(capability))
		}

		models = append(models, &entity.ModelSpec{
			Id:           m.config.Id,
			Name:         m.config.Name,
			From:         m.config.From,
			Provider:     m.config.Provider,
			Modalities:   modalities,
			Capabilities: capabilities,
		})
	}

	return models, nil
}
