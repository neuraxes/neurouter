package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"google.golang.org/protobuf/proto"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"git.xdea.xyz/Turing/router/internal/biz"
)

type RouterService struct {
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

func (s *RouterService) Chat(ctx context.Context, req *v1.ChatReq) (resp *v1.ChatResp, err error) {
	chatReq := proto.Clone(req).(*v1.ChatReq)
	r, err := s.chat.Chat(ctx, (*biz.ChatReq)(chatReq))
	if err != nil {
		return
	}

	resp = (*v1.ChatResp)(r)
	return
}

type wrappedChatStreamServer struct {
	srv v1.Chat_ChatStreamServer
}

func (w *wrappedChatStreamServer) Send(resp *biz.ChatResp) error {
	return w.srv.Send((*v1.ChatResp)(resp))
}

func (s *RouterService) ChatStream(req *v1.ChatReq, srv v1.Chat_ChatStreamServer) error {
	m := s.chatStreamLog(func(ctx context.Context, req any) (_ any, err error) {
		chatReq := proto.Clone(req.(proto.Message)).(*v1.ChatReq)
		err = s.chat.ChatStream(ctx, (*biz.ChatReq)(chatReq), &wrappedChatStreamServer{srv})
		return
	})
	_, err := m(srv.Context(), req)
	return err
}
