package model

import (
	"context"
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

func (m *embeddingModel) Close() {
	// Release in reverse order
	if m.modelSem != nil {
		m.modelSem.Release(1)
	}
	if m.upstreamSem != nil {
		m.upstreamSem.Release(1)
	}
}

func (uc *UseCaseImpl) ElectForEmbedding(ctx context.Context, req *v1.EmbedReq) (embedding.Model, error) {
	// Collect all available candidates
	var allCandidates []*model
	var matchingCandidates []*model

	for _, m := range uc.models {
		if m.embeddingRepo == nil || !slices.Contains(m.config.Capabilities, conf.Capability_CAPABILITY_EMBEDDING) {
			continue
		}
		allCandidates = append(allCandidates, m)
		if m.config.Id == req.Model {
			matchingCandidates = append(matchingCandidates, m)
		}
	}

	var selected *model
	var err error

	// If there are matching models, randomly select from them
	if len(matchingCandidates) > 0 {
		selected, err = electFromCandidates(ctx, matchingCandidates)
		if err != nil {
			return nil, err
		}
		uc.log.Infof("using model: %s-%s", selected.upstreamConfig.Name, selected.config.Id)
	} else if len(allCandidates) > 0 {
		// No matching models, randomly select from all candidates
		selected, err = electFromCandidates(ctx, allCandidates)
		if err != nil {
			return nil, err
		}
		uc.log.Infof("fallback to model: %s-%s (requested: %s)", selected.upstreamConfig.Name, selected.config.Id, req.Model)
	} else {
		return nil, v1.ErrorNoUpstream("no upstream found")
	}

	// Update request model to upstream ID
	if selected.config.UpstreamId != "" {
		req.Model = selected.config.UpstreamId
	} else {
		req.Model = selected.config.Id
	}

	return &embeddingModel{selected}, nil
}
