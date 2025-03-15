package biz

import (
	"github.com/google/wire"

	"git.xdea.xyz/Turing/neurouter/internal/biz/chat"
	"git.xdea.xyz/Turing/neurouter/internal/biz/model"
)

var ProviderSet = wire.NewSet(
	chat.NewChatUseCase,
	model.NewModelUseCase,
	wire.Bind(new(model.UseCase), new(*model.UseCaseImpl)),
	wire.Bind(new(chat.Elector), new(*model.UseCaseImpl)),
)
