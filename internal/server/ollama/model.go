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

package ollama

import (
	"context"
	"encoding/json"
	"io"

	"github.com/go-kratos/kratos/v2/transport/http"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func handleListModels(ctx http.Context, svc v1.ModelServer) error {
	m := ctx.Middleware(func(ctx context.Context, req any) (any, error) {
		return svc.ListModel(ctx, &v1.ListModelReq{})
	})
	r, err := m(ctx, nil)
	if err != nil {
		return err
	}

	resp := r.(*v1.ListModelResp)

	ollamaResp := &ListModelsResp{}
	for _, model := range resp.Models {
		ollamaResp.Models = append(ollamaResp.Models, Model{
			Name:  model.Name,
			Model: model.Id,
		})
	}

	return ctx.Result(200, ollamaResp)
}

func handleShowModel(ctx http.Context, svc v1.ModelServer) error {
	requestBody, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}

	showModelReq := ShowModelReq{}
	err = json.Unmarshal(requestBody, &showModelReq)
	if err != nil {
		return err
	}

	m := ctx.Middleware(func(ctx context.Context, req any) (any, error) {
		return svc.ListModel(ctx, &v1.ListModelReq{})
	})
	r, err := m(ctx, nil)
	if err != nil {
		return err
	}

	resp := r.(*v1.ListModelResp)

	for _, model := range resp.Models {
		if model.Id == showModelReq.Model {
			detail := &ModelDetail{
				ModelInfo: map[string]any{},
			}

			for _, c := range model.Capabilities {
				switch c {
				case v1.Capability_CAPABILITY_EMBEDDING:
					detail.Capabilities = append(detail.Capabilities, "embedding")
				case v1.Capability_CAPABILITY_TOOL_USE:
					detail.Capabilities = append(detail.Capabilities, "tools")
				}
			}

			return ctx.Result(200, detail)
		}
	}

	return ctx.Result(404, nil)
}
