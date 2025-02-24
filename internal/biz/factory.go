package biz

import (
	"github.com/go-kratos/kratos/v2/log"

	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type NeurouterChatRepoFactory func(config *conf.NeurouterConfig, logger log.Logger) ChatRepo
type OpenAIChatRepoFactory func(config *conf.OpenAIConfig, logger log.Logger) ChatRepo
type GoogleChatRepoFactory func(config *conf.GoogleConfig, logger log.Logger) ChatRepo
type AnthropicChatRepoFactory func(config *conf.AnthropicConfig, logger log.Logger) ChatRepo
type DeepSeekChatRepoFactory func(config *conf.DeepSeekConfig, logger log.Logger) ChatRepo
