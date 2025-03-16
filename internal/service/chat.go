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

	"google.golang.org/protobuf/proto"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz/entity"
)

func (s *RouterService) Chat(ctx context.Context, req *v1.ChatReq) (resp *v1.ChatResp, err error) {
	chatReq := proto.Clone(req).(*v1.ChatReq)
	r, err := s.chat.Chat(ctx, (*entity.ChatReq)(chatReq))
	if err != nil {
		return
	}

	resp = (*v1.ChatResp)(r)
	return
}

type wrappedChatStreamServer struct {
	srv v1.Chat_ChatStreamServer
}

func (w *wrappedChatStreamServer) Send(resp *entity.ChatResp) error {
	return w.srv.Send((*v1.ChatResp)(resp))
}

func (s *RouterService) ChatStream(req *v1.ChatReq, srv v1.Chat_ChatStreamServer) error {
	m := s.chatStreamLog(func(ctx context.Context, req any) (_ any, err error) {
		chatReq := proto.Clone(req.(proto.Message)).(*v1.ChatReq)
		err = s.chat.ChatStream(ctx, (*entity.ChatReq)(chatReq), &wrappedChatStreamServer{srv})
		return
	})
	_, err := m(srv.Context(), req)
	return err
}
