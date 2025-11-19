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

	"github.com/go-kratos/kratos/v2/log"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
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
	model, err := uc.elector.ElectForChat(ctx, req)
	if err != nil {
		return
	}
	defer model.Close()

	resp, err = model.ChatRepo().Chat(ctx, req)
	if err != nil {
		return
	}

	model.RecordUsage(resp.Statistics)
	uc.printChat(req, resp)
	return
}

func (uc *chatUseCase) ChatStream(ctx context.Context, req *entity.ChatReq, server repository.ChatStreamServer) error {
	model, err := uc.elector.ElectForChat(ctx, req)
	if err != nil {
		return err
	}
	defer model.Close()

	accumulator := NewChatRespAccumulator()
	for resp, err := range model.ChatRepo().ChatStream(ctx, req) {
		if err != nil {
			return err
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			break
		}

		accumulator.Accumulate(resp)
		err = server.Send(resp)
		if err != nil {
			return err
		}
	}

	finalResp := accumulator.Resp()
	model.RecordUsage(finalResp.Statistics)
	uc.printChat(req, finalResp)
	return nil
}
