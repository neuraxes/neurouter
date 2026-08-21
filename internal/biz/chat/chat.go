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

	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/observability"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

type UseCase interface {
	Chat(ctx context.Context, req *entity.ChatRequest) (*entity.ChatResponse, error)
	ChatStream(ctx context.Context, req *entity.ChatRequest, stream repository.ChatStreamServer) error
}

type chatUseCase struct {
	elector      Elector
	instrumenter *observability.GenAIInstrumenter
	log          *slog.Logger
}

func NewChatUseCase(
	elector Elector,
	instrumenter *observability.GenAIInstrumenter,
	logger *slog.Logger,
) UseCase {
	return &chatUseCase{
		elector:      elector,
		instrumenter: instrumenter,
		log:          logger,
	}
}

func (uc *chatUseCase) Chat(ctx context.Context, req *entity.ChatRequest) (resp *entity.ChatResponse, err error) {
	requestedModel := req.GetModel()
	model, err := uc.elector.ElectForChat(ctx, req)
	if err != nil {
		return
	}
	defer model.Close()

	target := model.GenAITarget()
	target.RequestedModel = requestedModel
	ctx, invocation := uc.instrumenter.Start(
		ctx,
		genaiconv.OperationNameChat,
		target,
		requestAttributes(req, false)...,
	)
	defer func() { invocation.End(invocationResult(resp), err) }()

	resp, err = model.ChatRepo().Chat(ctx, req)
	if err != nil {
		return
	}

	model.RecordUsage(ctx, resp.Statistics)
	uc.printChat(req, resp)
	return
}

func (uc *chatUseCase) ChatStream(
	ctx context.Context,
	req *entity.ChatRequest,
	server repository.ChatStreamServer,
) (err error) {
	requestedModel := req.GetModel()
	model, err := uc.elector.ElectForChat(ctx, req)
	if err != nil {
		return err
	}
	defer model.Close()

	target := model.GenAITarget()
	target.RequestedModel = requestedModel
	ctx, invocation := uc.instrumenter.Start(
		ctx,
		genaiconv.OperationNameChat,
		target,
		requestAttributes(req, true)...,
	)
	var finalResp *entity.ChatResponse
	var observationErr error
	defer func() {
		if observationErr == nil {
			observationErr = err
		}
		invocation.End(invocationResult(finalResp), observationErr)
	}()

	reducer := NewChatEventReducer(uc.log)
	firstChunk := true
	for event, streamErr := range model.ChatRepo().ChatStream(ctx, req) {
		if streamErr != nil {
			return streamErr
		}
		if firstChunk {
			invocation.FirstChunk()
			firstChunk = false
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			observationErr = ctx.Err()
			break
		}

		reducer.Reduce(event)
		if sendErr := server.Send(event); sendErr != nil {
			return sendErr
		}
	}

	finalResp = reducer.Resp()
	model.RecordUsage(ctx, finalResp.Statistics)
	uc.printChat(req, finalResp)
	return nil
}
