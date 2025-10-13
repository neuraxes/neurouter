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

package openai

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/http"
	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/sashabaranov/go-openai"
)

func handleListModels(httpCtx http.Context, svc v1.ModelServer) error {
	m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
		return svc.ListModel(ctx, &v1.ListModelReq{})
	})
	r, err := m(httpCtx, nil)
	if err != nil {
		return err
	}

	resp := r.(*v1.ListModelResp)

	openaiResp := &openai.ModelsList{}
	for _, model := range resp.Models {
		openaiResp.Models = append(openaiResp.Models, openai.Model{
			ID: model.Id,
		})
	}

	return httpCtx.Result(200, openaiResp)
}
