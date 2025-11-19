package model

import (
	"slices"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/chat"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type chatModel struct {
	*model
}

func (m *chatModel) ChatRepo() repository.ChatRepo { return m.chatRepo }
func (m *chatModel) RecordUsage(stats *v1.Statistics) {
	if stats == nil || stats.Usage == nil {
		return
	}
	m.inputTokens.Add(uint64(stats.Usage.InputTokens))
	m.outputTokens.Add(uint64(stats.Usage.OutputTokens))
	m.cachedInputTokens.Add(uint64(stats.Usage.CachedInputTokens))
}
func (m *chatModel) Close() {}

func (uc *UseCaseImpl) ElectForChat(req *v1.ChatReq) (chat.Model, error) {
	var selected *model
	for _, m := range uc.models {
		if m.chatRepo == nil || !slices.Contains(m.config.Capabilities, conf.Capability_CAPABILITY_CHAT) {
			continue
		}
		selected = m
		if m.config.Id == req.Model {
			uc.log.Infof("using model: %s", m.config.Id)
			if m.config.UpstreamId != "" {
				req.Model = m.config.UpstreamId
			} else {
				req.Model = m.config.Id
			}
			return &chatModel{m}, nil
		}
	}

	if selected != nil {
		uc.log.Infof("fallback to model: %s", selected.config.Id)
		if selected.config.UpstreamId != "" {
			req.Model = selected.config.UpstreamId
		} else {
			req.Model = selected.config.Id
		}
		return &chatModel{selected}, nil
	}

	return nil, v1.ErrorNoUpstream("no upstream found")
}
