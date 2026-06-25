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
	"github.com/openai/openai-go/v3"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func convertEmbeddingReqFromOpenAI(req *openai.EmbeddingNewParams) *v1.EmbedReq {
	var contents []*v1.Content

	if req.Input.OfString.Valid() {
		contents = append(contents, &v1.Content{
			Content: v1.NewTextContent(req.Input.OfString.Value),
		})
	} else if len(req.Input.OfArrayOfStrings) > 0 {
		contents = append(contents, &v1.Content{
			Content: v1.NewTextContent(req.Input.OfArrayOfStrings[0]),
		})
	}

	return &v1.EmbedReq{
		Model:    string(req.Model),
		Contents: contents,
	}
}

func convertEmbeddingRespToOpenAI(resp *v1.EmbedResp) *embeddingResponse {
	embedding := make([]float64, len(resp.Embedding))
	for i, v := range resp.Embedding {
		embedding[i] = float64(v)
	}
	return &embeddingResponse{
		Object: "list",
		Model:  resp.Model,
		Data: []openai.Embedding{
			{
				Index:     0,
				Embedding: embedding,
			},
		},
	}
}
