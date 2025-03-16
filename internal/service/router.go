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
	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz/chat"
	"git.xdea.xyz/Turing/neurouter/internal/biz/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
)

type RouterService struct {
	v1.UnimplementedModelServer
	v1.UnimplementedChatServer
	chat          chat.UseCase
	model         model.UseCase
	chatStreamLog middleware.Middleware
}

func NewRouterService(chat chat.UseCase, model model.UseCase, logger log.Logger) *RouterService {
	return &RouterService{
		chat:          chat,
		model:         model,
		chatStreamLog: logging.Server(logger),
	}
}
