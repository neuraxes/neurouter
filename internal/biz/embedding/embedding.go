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

	"github.com/go-kratos/kratos/v2/log"

	"github.com/neuraxes/neurouter/internal/biz/entity"
)

type UseCase interface {
	Embed(ctx context.Context, req *entity.EmbedReq) (*entity.EmbedResp, error)
}

type useCase struct {
	elector Elector
	log     *log.Helper
}

// NewUseCase creates a new embedding use case instance.
func NewUseCase(elector Elector, logger log.Logger) UseCase {
	return &useCase{
		elector: elector,
		log:     log.NewHelper(logger),
	}
}

// Embed creates embeddings for the given contents using the specified model.
func (uc *useCase) Embed(ctx context.Context, req *entity.EmbedReq) (resp *entity.EmbedResp, err error) {
	model, err := uc.elector.ElectForEmbedding(ctx, req)
	if err != nil {
		return
	}
	defer model.Close()

	return model.EmbeddingRepo().Embed(ctx, req)
}
