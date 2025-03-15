package upstream

import (
	"github.com/google/wire"

	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/anthropic"
	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/deepseek"
	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/neurouter"
	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/openai"
)

var ProviderSet = wire.NewSet(
	neurouter.NewNeurouterChatRepoFactory,
	openai.NewOpenAIChatRepoFactory,
	anthropic.NewAnthropicChatRepoFactory,
	deepseek.NewDeepSeekChatRepoFactory,
)
