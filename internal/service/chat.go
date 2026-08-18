// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"

	"github.com/go-kratos/kratos/contrib/middleware/jwt/v3"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

// acceptChatRequest copies the inbound request so the caller's message is never
// mutated, and assigns the request id that every response and event of the turn
// echoes back.
func acceptChatRequest(req *v1.ChatRequest) *v1.ChatRequest {
	chatReq := proto.Clone(req).(*v1.ChatRequest)
	if chatReq.Id == "" {
		chatReq.Id = uuid.NewString()
	}
	return chatReq
}

func (s *RouterService) Chat(ctx context.Context, req *v1.ChatRequest) (resp *v1.ChatResponse, err error) {
	if claims, ok := jwt.FromContext(ctx); ok {
		sub, _ := claims.GetSubject()
		s.log.InfoContext(ctx, "authenticated with JWT", "subject", sub)
	}

	resp, err = s.chat.Chat(ctx, acceptChatRequest(req))
	return
}

type wrappedChatStreamServer struct {
	srv v1.Chat_ChatStreamServer
}

func (w *wrappedChatStreamServer) Send(event *entity.ChatEvent) error {
	return w.srv.Send(event)
}

func (s *RouterService) ChatStream(req *v1.ChatRequest, srv v1.Chat_ChatStreamServer) error {
	if claims, ok := jwt.FromContext(srv.Context()); ok {
		sub, _ := claims.GetSubject()
		s.log.InfoContext(srv.Context(), "authenticated with JWT", "subject", sub)
	}

	err := s.chat.ChatStream(srv.Context(), acceptChatRequest(req), &wrappedChatStreamServer{srv})
	return err
}
