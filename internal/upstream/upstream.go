package upstream

import (
	"github.com/google/wire"

	"git.xdea.xyz/Turing/router/internal/upstream/openai"
)

var ProviderSet = wire.NewSet(
	openai.NewOpenAIChatRepoFactory,
	NewAnthropicChatRepoFactory,
)
