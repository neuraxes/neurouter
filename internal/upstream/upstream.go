package upstream

import (
	"github.com/google/wire"

	"git.xdea.xyz/Turing/neurouter/internal/upstream/anthropic"
	"git.xdea.xyz/Turing/neurouter/internal/upstream/deepseek"
	"git.xdea.xyz/Turing/neurouter/internal/upstream/neurouter"
	"git.xdea.xyz/Turing/neurouter/internal/upstream/openai"
)

var ProviderSet = wire.NewSet(
	neurouter.NewNeurouterChatRepoFactory,
	openai.NewOpenAIChatRepoFactory,
	anthropic.NewAnthropicChatRepoFactory,
	deepseek.NewDeepSeekChatRepoFactory,
)
