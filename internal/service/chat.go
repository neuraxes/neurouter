package service

import (
	"context"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"git.xdea.xyz/Turing/router/internal/biz"
)

type RouterService struct {
	v1.UnimplementedChatServer
	chat biz.ChatUseCase
}

func NewRouterService(chat biz.ChatUseCase) *RouterService {
	return &RouterService{
		chat: chat,
	}
}

func (s *RouterService) Chat(ctx context.Context, req *v1.ChatReq) (resp *v1.ChatResp, err error) {
	r, err := s.chat.Chat(ctx, (*biz.ChatReq)(req))
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
	return s.chat.ChatStream(srv.Context(), (*biz.ChatReq)(req), &wrappedChatStreamServer{srv})
}
