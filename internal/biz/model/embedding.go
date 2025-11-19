package model

import (
	"slices"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/embedding"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type embeddingModel struct {
	*model
}

func (m *embeddingModel) EmbeddingRepo() repository.EmbeddingRepo { return m.embeddingRepo }
func (m *embeddingModel) Close()                                  {}

func (uc *UseCaseImpl) ElectForEmbedding(req *v1.EmbedReq) (embedding.Model, error) {
	var selected *model
	for _, m := range uc.models {
		if m.embeddingRepo == nil || !slices.Contains(m.config.Capabilities, conf.Capability_CAPABILITY_EMBEDDING) {
			continue
		}
		selected = m
		if m.config.Id == req.Model {
			uc.log.Infof("using model: %s", m.config.Name)
			if m.config.UpstreamId != "" {
				req.Model = m.config.UpstreamId
			} else {
				req.Model = m.config.Id
			}
			return &embeddingModel{m}, nil
		}
	}

	if selected != nil {
		uc.log.Infof("fallback to model: %s", selected.config.Name)
		if selected.config.UpstreamId != "" {
			req.Model = selected.config.UpstreamId
		} else {
			req.Model = selected.config.Id
		}
		return &embeddingModel{selected}, nil
	}

	return nil, v1.ErrorNoUpstream("no upstream found")
}
