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
	"slices"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
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
				repo repository.ChatRepo
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
		if !slices.Contains(m.config.Capabilities, conf.Capability_CAPABILITY_CHAT) {
			continue
		}
		repo = m.repo
		model = m.config
		if m.config.Id == uri {
			uc.log.Infof("using model: %s", m.config.Id)
			return
		}
	}

	if model != nil {
		uc.log.Infof("fallback to model: %s", model.Id)
		return
	}

	err = v1.ErrorNoUpstream("no upstream found")
	return
}

func (uc *UseCaseImpl) ElectForEmbedding(uri string) (repo repository.EmbeddingRepo, model *conf.Model, err error) {
	for _, m := range uc.models {
		if !slices.Contains(m.config.Capabilities, conf.Capability_CAPABILITY_EMBEDDING) {
			continue
		}
		if r, ok := m.repo.(repository.EmbeddingRepo); ok {
			repo = r
			model = m.config
		}
		if m.config.Id == uri {
			uc.log.Infof("using model: %s", m.config.Name)
			return
		}
	}

	if model != nil {
		uc.log.Infof("fallback to model: %s", model.Name)
		return
	}

	err = v1.ErrorNoUpstream("no upstream found")
	return
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
