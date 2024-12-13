package biz

import "git.xdea.xyz/Turing/router/internal/conf"

type LaaSChatRepoFactory func(config *conf.LaaSConfig) ChatRepo
type OpenAIChatRepoFactory func(config *conf.OpenAIConfig) ChatRepo
type GoogleChatRepoFactory func(config *conf.GoogleConfig) ChatRepo
type AnthropicChatRepoFactory func(config *conf.AnthropicConfig) ChatRepo
