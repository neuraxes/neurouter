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
	"iter"
	"slices"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

func (uc *chatUseCase) recoverChatRequest(ctx context.Context, req *entity.ChatRequest, err error) bool {
	if ctx.Err() != nil || !errors.Is(err, entity.ErrInvalidReasoningState) {
		return false
	}
	if !removeReasoningContents(req) {
		return false
	}
	uc.log.WarnContext(ctx, "retrying chat request without reasoning contents", "error", err)
	return true
}

func (uc *chatUseCase) chatWithRecovery(
	ctx context.Context, repo repository.ChatRepo, req *entity.ChatRequest,
) (*entity.ChatResponse, error) {
	resp, err := repo.Chat(ctx, req)
	if uc.recoverChatRequest(ctx, req, err) {
		return repo.Chat(ctx, req)
	}
	return resp, err
}

func (uc *chatUseCase) chatStreamWithRecovery(
	ctx context.Context, repo repository.ChatRepo, req *entity.ChatRequest,
) iter.Seq2[*entity.ChatEvent, error] {
	return func(yield func(*entity.ChatEvent, error) bool) {
		started := false
		for attempt := 0; attempt < 2; attempt++ {
			var attemptErr error
			for event, err := range repo.ChatStream(ctx, req) {
				if err != nil {
					attemptErr = err
					break
				}
				if event == nil {
					continue
				}
				started = true
				if !yield(event, nil) {
					return
				}
			}
			if attemptErr == nil {
				return
			}
			if !started && attempt == 0 {
				if uc.recoverChatRequest(ctx, req, attemptErr) {
					continue
				}
			}
			yield(nil, attemptErr)
			return
		}
	}
}

func removeReasoningContents(req *entity.ChatRequest) bool {
	stripped := false
	for _, message := range req.Messages {
		if message == nil {
			continue
		}
		count := len(message.Contents)
		message.Contents = slices.DeleteFunc(message.Contents, func(content *v1.Content) bool {
			return content.GetPhase() == v1.ContentPhase_CONTENT_PHASE_REASONING
		})
		stripped = stripped || len(message.Contents) < count
	}
	return stripped
}
