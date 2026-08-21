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

package embedding

import (
	"context"
	"log/slog"

	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/observability"
)

type UseCase interface {
	Embed(ctx context.Context, req *entity.EmbedRequest) (*entity.EmbedResponse, error)
}

type useCase struct {
	elector      Elector
	instrumenter *observability.GenAIInstrumenter
	log          *slog.Logger
}

// NewUseCase creates a new embedding use case instance.
func NewUseCase(
	elector Elector,
	instrumenter *observability.GenAIInstrumenter,
	logger *slog.Logger,
) UseCase {
	return &useCase{
		elector:      elector,
		instrumenter: instrumenter,
		log:          logger,
	}
}

// Embed creates embeddings for the given contents using the specified model.
func (uc *useCase) Embed(ctx context.Context, req *entity.EmbedRequest) (resp *entity.EmbedResponse, err error) {
	requestedModel := req.GetModel()
	model, err := uc.elector.ElectForEmbedding(ctx, req)
	if err != nil {
		return
	}
	defer model.Close()

	target := model.GenAITarget()
	target.RequestedModel = requestedModel
	ctx, invocation := uc.instrumenter.Start(ctx, genaiconv.OperationNameEmbeddings, target)
	defer func() {
		result := observability.GenAIResult{}
		if resp != nil {
			result.ResponseModel = resp.GetModel()
			result.Attributes = append(
				result.Attributes,
				semconv.GenAIEmbeddingsDimensionCount(len(resp.GetEmbedding())),
			)
		}
		invocation.End(result, err)
	}()

	resp, err = model.EmbeddingRepo().Embed(ctx, req)
	if err != nil {
		return
	}

	// TODO: record the actual token usage
	model.RecordUsage(ctx, 0)
	return
}
