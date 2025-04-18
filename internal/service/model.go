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

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func (s *RouterService) ListModel(ctx context.Context, req *v1.ListModelReq) (resp *v1.ListModelResp, err error) {
	if claims, ok := jwt.FromContext(ctx); ok {
		sub, _ := claims.GetSubject()
		s.log.Infof("jwt authenticated for: %s", sub)
	}

	models, err := s.model.ListAvailableModels(ctx)
	if err != nil {
		return
	}

	respModels := make([]*v1.ModelSpec, len(models))
	for i, m := range models {
		respModels[i] = (*v1.ModelSpec)(m)
	}

	resp = &v1.ListModelResp{
		Models: respModels,
	}
	return
}
