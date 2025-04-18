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
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/chat"
	"github.com/neuraxes/neurouter/internal/biz/embedding"
	"github.com/neuraxes/neurouter/internal/biz/model"

	"github.com/go-kratos/kratos/v2/log"
)

type RouterService struct {
	v1.UnimplementedModelServer
	v1.UnimplementedChatServer
	v1.UnimplementedEmbeddingServer
	chat      chat.UseCase
	model     model.UseCase
	embedding embedding.UseCase
	log       *log.Helper
}

func NewRouterService(
	chat chat.UseCase,
	model model.UseCase,
	embedding embedding.UseCase,
	logger log.Logger,
) *RouterService {
	return &RouterService{
		chat:      chat,
		model:     model,
		embedding: embedding,
		log:       log.NewHelper(logger),
	}
}
