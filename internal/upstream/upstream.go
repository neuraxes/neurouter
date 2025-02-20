package upstream

import (
	"github.com/google/wire"

	"git.xdea.xyz/Turing/neurouter/internal/upstream/anthropic"
	"git.xdea.xyz/Turing/neurouter/internal/upstream/deepseek"
	"git.xdea.xyz/Turing/neurouter/internal/upstream/openai"
)

var ProviderSet = wire.NewSet(
	deepseek.NewDeepSeekChatRepoFactory,
	openai.NewOpenAIChatRepoFactory,
	anthropic.NewAnthropicChatRepoFactory,
)
