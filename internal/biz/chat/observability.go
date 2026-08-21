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

func requestAttributes(req *v1.ChatRequest, streaming bool) []attribute.KeyValue {
	attrs := []attribute.KeyValue{semconv.GenAIRequestStream(streaming)}
	config := req.GetConfig()
	if config == nil {
		return attrs
	}
	if config.MaxTokens != nil {
		attrs = append(attrs, semconv.GenAIRequestMaxTokensKey.Int64(config.GetMaxTokens()))
	}
	if config.Temperature != nil {
		attrs = append(attrs, semconv.GenAIRequestTemperature(float64(config.GetTemperature())))
	}
	if config.TopP != nil {
		attrs = append(attrs, semconv.GenAIRequestTopP(float64(config.GetTopP())))
	}
	return attrs
}

func invocationResult(resp *v1.ChatResponse) observability.GenAIResult {
	if resp == nil {
		return observability.GenAIResult{}
	}

	result := observability.GenAIResult{
		ResponseModel: resp.GetModel(),
		FinishReasons: finishReasons(resp.GetStatus()),
	}
	if resp.GetMessage() != nil {
		result.ResponseID = resp.GetMessage().GetId()
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

func finishReasons(status v1.ChatStatus) []string {
	switch status {
	case v1.ChatStatus_CHAT_STATUS_COMPLETED:
		return []string{"stop"}
	case v1.ChatStatus_CHAT_STATUS_FAILED:
		return []string{"error"}
	case v1.ChatStatus_CHAT_STATUS_REFUSED:
		return []string{"content_filter"}
	case v1.ChatStatus_CHAT_STATUS_CANCELLED:
		return []string{"cancelled"}
	case v1.ChatStatus_CHAT_STATUS_PENDING_TOOL_USE:
		return []string{"tool_calls"}
	case v1.ChatStatus_CHAT_STATUS_REACHED_TOKEN_LIMIT:
		return []string{"length"}
	default:
		return nil
	}
}
