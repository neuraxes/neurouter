package biz

import (
	"github.com/go-kratos/kratos/v2/log"

	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type LaaSChatRepoFactory func(config *conf.LaaSConfig, logger log.Logger) ChatRepo
type OpenAIChatRepoFactory func(config *conf.OpenAIConfig, logger log.Logger) ChatRepo
type GoogleChatRepoFactory func(config *conf.GoogleConfig, logger log.Logger) ChatRepo
type AnthropicChatRepoFactory func(config *conf.AnthropicConfig, logger log.Logger) ChatRepo
