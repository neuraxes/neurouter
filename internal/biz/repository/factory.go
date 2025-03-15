package repository

import (
	"github.com/go-kratos/kratos/v2/log"

	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

// ChatUpstreamConfig is a type constraint for LLM provider configurations.
// It allows for configuration of different upstream LLM providers like OpenAI, Google, Anthropic, etc.
type ChatUpstreamConfig interface {
	conf.NeurouterConfig | conf.OpenAIConfig | conf.GoogleConfig | conf.AnthropicConfig | conf.DeepSeekConfig
}

// ChatRepoFactory is a generic factory function type for creating ChatRepo instances.
type ChatRepoFactory[T ChatUpstreamConfig] func(config *T, logger log.Logger) (ChatRepo, error)
