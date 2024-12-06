package openai

import (
	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func RegisterOpenAIHTTPServer(s *http.Server, svc v1.ChatServer) {
	r := s.Route("/")
	r.POST("/chat/completions", func(ctx http.Context) error {
		return handleChatCompletion(ctx, svc)
	})
	r.POST("/v1/chat/completions", func(ctx http.Context) error {
		return handleChatCompletion(ctx, svc)
	})
}
