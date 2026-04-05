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
	reservations    *reservationSet
	estimatedTokens int64
}

func (m *embeddingModel) EmbeddingRepo() repository.EmbeddingRepo { return m.embeddingRepo }

func (m *embeddingModel) RecordUsage(ctx context.Context, actualTokens int64) {
	// If upstream doesn't provide usage info, fall back to estimated tokens
	if actualTokens == 0 {
		actualTokens = m.estimatedTokens
	}

	m.metrics.recordTokenUsage(
		ctx,
		m.upstreamConfig.Name,
		m.config.Id,
		actualTokens, 0, 0, 0,
	)
	m.metrics.recordRequest(ctx, m.upstreamConfig.Name, m.config.Id)

	// Complete reservations with actual token usage
	m.reservations.complete(actualTokens)
}

func (m *embeddingModel) Close() {
	m.reservations.cancel()
}

// estimateEmbeddingTokens provides a rough token estimate for an embedding request.
// Uses ~4 characters per token heuristic.
func estimateEmbeddingTokens(req *v1.EmbedReq) int64 {
	totalChars := 0
	for _, c := range req.Contents {
		totalChars += len(c.GetText())
	}
	if totalChars == 0 {
		return 0
	}
	return int64(totalChars/4) + 1
}

func (uc *UseCaseImpl) ElectForEmbedding(ctx context.Context, req *v1.EmbedReq) (embedding.Model, error) {
	estimatedTokens := estimateEmbeddingTokens(req) // Estimate input tokens roughly: ~4 chars per token

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

	if a := uc.aliases[req.Model]; a != nil {
		for _, m := range a.models {
			if m.embeddingRepo == nil || !slices.Contains(m.config.Capabilities, conf.Capability_CAPABILITY_EMBEDDING) {
				continue
			}
			if !slices.Contains(matchingCandidates, m) {
				matchingCandidates = append(matchingCandidates, m)
			}
		}
	}

	var selected *model
	var rs *reservationSet
	var err error

	// If there are matching models, randomly select from them
	if len(matchingCandidates) > 0 {
		selected, rs, err = electFromCandidates(ctx, matchingCandidates, estimatedTokens)
		if err != nil {
			return nil, err
		}
		uc.log.Infof("using model: %s-%s", selected.upstreamConfig.Name, selected.config.Id)
	} else if len(allCandidates) > 0 {
		// No matching models, randomly select from all candidates
		selected, rs, err = electFromCandidates(ctx, allCandidates, estimatedTokens)
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

	return &embeddingModel{
		model:           selected,
		reservations:    rs,
		estimatedTokens: estimatedTokens,
	}, nil
}
