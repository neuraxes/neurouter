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
	"log/slog"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

type UseCase interface {
	Chat(ctx context.Context, req *entity.ChatRequest) (*entity.ChatResponse, error)
	ChatStream(ctx context.Context, req *entity.ChatRequest, stream repository.ChatStreamServer) error
}

type chatUseCase struct {
	elector Elector
	log     *slog.Logger
}

func NewChatUseCase(elector Elector, logger *slog.Logger) UseCase {
	return &chatUseCase{
		elector: elector,
		log:     logger,
	}
}

func (uc *chatUseCase) Chat(ctx context.Context, req *entity.ChatRequest) (resp *entity.ChatResponse, err error) {
	model, err := uc.elector.ElectForChat(ctx, req)
	if err != nil {
		return
	}
	defer model.Close()

	resp, err = model.ChatRepo().Chat(ctx, req)
	if err != nil {
		return
	}

	model.RecordUsage(ctx, resp.Statistics)
	uc.printChat(req, resp)
	return
}

func (uc *chatUseCase) ChatStream(ctx context.Context, req *entity.ChatRequest, server repository.ChatStreamServer) error {
	model, err := uc.elector.ElectForChat(ctx, req)
	if err != nil {
		return err
	}
	defer model.Close()

	reducer := NewChatEventReducer(uc.log)
	for event, err := range model.ChatRepo().ChatStream(ctx, req) {
		if err != nil {
			return err
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			break
		}

		reducer.Reduce(event)
		err = server.Send(event)
		if err != nil {
			return err
		}
	}

	finalResp := reducer.Resp()
	model.RecordUsage(ctx, finalResp.Statistics)
	uc.printChat(req, finalResp)
	return nil
}
