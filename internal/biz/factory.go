package biz

import "git.xdea.xyz/Turing/router/internal/conf"

type LaaSChatCompletionRepoFactory func(config *conf.LaaSConfig) ChatCompletionRepo
type OpenAIChatCompletionRepoFactory func(config *conf.OpenAIConfig) ChatCompletionRepo
type GoogleChatCompletionRepoFactory func(config *conf.GoogleConfig) ChatCompletionRepo
