package service

import (
	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
)

type RouterService struct {
	v1.UnimplementedModelServer
	v1.UnimplementedChatServer
	chat          biz.ChatUseCase
	chatStreamLog middleware.Middleware
}

func NewRouterService(chat biz.ChatUseCase, logger log.Logger) *RouterService {
	return &RouterService{
		chat:          chat,
		chatStreamLog: logging.Server(logger),
	}
}
