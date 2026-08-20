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
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/observability"
)

func requestAttrs(req *v1.ChatRequest, streaming bool) []attribute.KeyValue {
	attrs := []attribute.KeyValue{semconv.GenAIRequestStream(streaming)}
	config := req.GetConfig()
	if config == nil {
		return attrs
	}
	if maxTokens := config.MaxTokens; maxTokens != nil {
		attrs = append(attrs, semconv.GenAIRequestMaxTokensKey.Int64(*maxTokens))
	}
	if temperature := config.Temperature; temperature != nil {
		attrs = append(attrs, semconv.GenAIRequestTemperature(float64(*temperature)))
	}
	if topP := config.TopP; topP != nil {
		attrs = append(attrs, semconv.GenAIRequestTopP(float64(*topP)))
	}
	return attrs
}

func genaiResult(resp *v1.ChatResponse) observability.GenAIResult {
	if resp == nil {
		return observability.GenAIResult{}
	}

	result := observability.GenAIResult{
		ResponseID:    resp.GetMessage().GetId(),
		ResponseModel: resp.GetModel(),
	}

	switch resp.GetStatus() {
	case v1.ChatStatus_CHAT_STATUS_COMPLETED:
		result.FinishReasons = []string{"completed"}
	case v1.ChatStatus_CHAT_STATUS_FAILED:
		result.FinishReasons = []string{"failed"}
	case v1.ChatStatus_CHAT_STATUS_REFUSED:
		result.FinishReasons = []string{"refused"}
	case v1.ChatStatus_CHAT_STATUS_CANCELLED:
		result.FinishReasons = []string{"cancelled"}
	case v1.ChatStatus_CHAT_STATUS_PENDING_TOOL_USE:
		result.FinishReasons = []string{"pending_tool_use"}
	case v1.ChatStatus_CHAT_STATUS_REACHED_TOKEN_LIMIT:
		result.FinishReasons = []string{"reached_token_limit"}
	}

	if usage := resp.GetStatistics().GetUsage(); usage != nil {
		result.Usage = &observability.GenAITokenUsage{
			Input:       int64(usage.GetInputTokens()),
			Output:      int64(usage.GetOutputTokens()),
			CachedInput: int64(usage.GetCachedInputTokens()),
			Reasoning:   int64(usage.GetReasoningTokens()),
		}
	}

	return result
}
