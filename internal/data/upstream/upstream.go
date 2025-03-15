package upstream

import (
	"github.com/google/wire"

	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/anthropic"
	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/deepseek"
	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/google"
	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/neurouter"
	"git.xdea.xyz/Turing/neurouter/internal/data/upstream/openai"
)

var ProviderSet = wire.NewSet(
	anthropic.NewAnthropicChatRepoFactory,
	deepseek.NewDeepSeekChatRepoFactory,
	google.NewGoogleChatRepoFactory,
	neurouter.NewNeurouterChatRepoFactory,
	openai.NewOpenAIChatRepoFactory,
)
