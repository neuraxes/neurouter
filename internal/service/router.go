package service

import (
	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz/chat"
	"git.xdea.xyz/Turing/neurouter/internal/biz/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
)

type RouterService struct {
	v1.UnimplementedModelServer
	v1.UnimplementedChatServer
	chat          chat.UseCase
	model         model.UseCase
	chatStreamLog middleware.Middleware
}

func NewRouterService(chat chat.UseCase, model model.UseCase, logger log.Logger) *RouterService {
	return &RouterService{
		chat:          chat,
		model:         model,
		chatStreamLog: logging.Server(logger),
	}
}
