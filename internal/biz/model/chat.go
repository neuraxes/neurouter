package model

import (
	"context"
	"slices"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/chat"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type chatModel struct {
	*model
	reservations    *reservationSet
	estimatedTokens int64
}

func (m *chatModel) ChatRepo() repository.ChatRepo { return m.chatRepo }
func (m *chatModel) RecordUsage(ctx context.Context, stats *v1.Statistics) {
	actualTokens := m.estimatedTokens // Default to estimated tokens

	if stats != nil && stats.Usage != nil {
		inputTokens := int64(stats.Usage.InputTokens)
		outputTokens := int64(stats.Usage.OutputTokens)
		cachedInputTokens := int64(stats.Usage.CachedInputTokens)

		m.inputTokens.Add(inputTokens)
		m.outputTokens.Add(outputTokens)
		m.cachedInputTokens.Add(cachedInputTokens)

		m.metrics.recordTokenUsage(
			ctx,
			m.upstreamConfig.Name,
			m.config.Id,
			inputTokens,
			outputTokens,
			cachedInputTokens,
		)

		tokenUsage := int64(stats.Usage.InputTokens + stats.Usage.OutputTokens)
		// If upstream provides usage info, use actual tokens
		if tokenUsage > 0 {
			actualTokens = tokenUsage
		}
	}

	m.metrics.recordRequest(ctx, m.upstreamConfig.Name, m.config.Id)

	// Complete reservations with actual or estimated token usage
	m.reservations.complete(actualTokens)
}

func (m *chatModel) Close() {
	m.reservations.cancel()
}

// estimateTokens provides a rough token estimate for a chat request.
// Uses ~4 characters per token heuristic for text.
// Images are assigned a fixed token count of 768 tokens per image.
func estimateTokens(req *v1.ChatReq) int64 {
	totalChars := 0
	imageCount := 0

	for _, msg := range req.Messages {
		for _, c := range msg.Contents {
			switch content := c.Content.(type) {
			case *v1.Content_Image:
				imageCount++
			case *v1.Content_Text:
				totalChars += len(content.Text)
			case *v1.Content_ToolUse:
				totalChars += len(content.ToolUse.Name)
				for _, input := range content.ToolUse.Inputs {
					totalChars += len(input.GetText())
				}
			case *v1.Content_ToolResult:
				for _, output := range content.ToolResult.Outputs {
					totalChars += len(output.GetText())
				}
			}
		}
	}

	textTokens := int64(totalChars / 4)
	if totalChars > 0 {
		textTokens++
	}
	imageTokens := int64(imageCount * 768)

	return textTokens + imageTokens
}

func (uc *UseCaseImpl) ElectForChat(ctx context.Context, req *v1.ChatReq) (chat.Model, error) {
	estimatedTokens := estimateTokens(req) // Estimate input tokens roughly: ~4 chars per token
	estimatedTokens += 512                 // Add some buffer for output tokens

	// Collect all available candidates
	var allCandidates []*model
	var matchingCandidates []*model

	for _, m := range uc.models {
		if m.chatRepo == nil || !slices.Contains(m.config.Capabilities, conf.Capability_CAPABILITY_CHAT) {
			continue
		}
		allCandidates = append(allCandidates, m)
		if m.config.Id == req.Model {
			matchingCandidates = append(matchingCandidates, m)
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
		uc.log.Infof("using model: %s:%s", selected.upstreamConfig.Name, selected.config.Id)
	} else if len(allCandidates) > 0 {
		// No matching models, randomly select from all candidates
		selected, rs, err = electFromCandidates(ctx, allCandidates, estimatedTokens)
		if err != nil {
			return nil, err
		}
		uc.log.Infof("fallback to model: %s:%s (requested: %s)", selected.upstreamConfig.Name, selected.config.Id, req.Model)
	} else {
		return nil, v1.ErrorNoUpstream("no upstream found")
	}

	// Update request model to upstream ID
	if selected.config.UpstreamId != "" {
		req.Model = selected.config.UpstreamId
	} else {
		req.Model = selected.config.Id
	}

	return &chatModel{
		model:           selected,
		reservations:    rs,
		estimatedTokens: estimatedTokens,
	}, nil
}
