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

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

// Embed creates embeddings for the given contents using the specified model.
func (s *RouterService) Embed(ctx context.Context, req *v1.EmbedReq) (resp *v1.EmbedResp, err error) {
	// Convert API request to entity
	embedReq := &entity.EmbedReq{
		Id:       req.Id,
		Model:    req.Model,
		Contents: req.Contents,
	}

	// Call use case
	r, err := s.embedding.Embed(ctx, embedReq)
	if err != nil {
		return
	}

	// Convert entity response to API response
	resp = &v1.EmbedResp{
		Id:        r.Id,
		Embedding: r.Embedding,
	}
	return
}
