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

package chat

import (
	"context"
	"errors"
	"io"

	"github.com/go-kratos/kratos/v2/log"

	"git.xdea.xyz/Turing/neurouter/internal/biz/entity"
	"git.xdea.xyz/Turing/neurouter/internal/biz/repository"
)

type UseCase interface {
	Chat(ctx context.Context, req *entity.ChatReq) (*entity.ChatResp, error)
	ChatStream(ctx context.Context, req *entity.ChatReq, stream repository.ChatStreamServer) error
}

type chatUseCase struct {
	elector Elector
	log     *log.Helper
}

func NewChatUseCase(elector Elector, logger log.Logger) UseCase {
	return &chatUseCase{
		elector: elector,
		log:     log.NewHelper(logger),
	}
}

func (uc *chatUseCase) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	repo, m, err := uc.elector.ElectForChat(req.Model)
	if err != nil {
		return
	}
	req.Model = m.Id
	return repo.Chat(ctx, req)
}

func (uc *chatUseCase) ChatStream(ctx context.Context, req *entity.ChatReq, server repository.ChatStreamServer) error {
	repo, m, err := uc.elector.ElectForChat(req.Model)
	if err != nil {
		return err
	}

	req.Model = m.Id
	client, err := repo.ChatStream(ctx, req)
	if err != nil {
		return err
	}
	defer client.Close()

	for {
		resp, err := client.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}

		err = server.Send(resp)
		if err != nil {
			return err
		}
	}
}
