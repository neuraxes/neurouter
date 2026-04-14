// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openai

type responseOutputText struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Annotations []any  `json:"annotations"`
}

type responseRefusal struct {
	Type    string `json:"type"`
	Refusal string `json:"refusal"`
}

type responseOutputContent any

type responseReasoningSummary struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responseOutputMessage struct {
	Type    string                  `json:"type"`
	ID      string                  `json:"id"`
	Role    string                  `json:"role"`
	Status  string                  `json:"status"`
	Content []responseOutputContent `json:"content"`
}

type responseFunctionCall struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	Status    string `json:"status"`
}

type responseReasoning struct {
	Type             string                    `json:"type"`
	ID               string                    `json:"id"`
	Summary          []responseReasoningSummary `json:"summary"`
	EncryptedContent string                     `json:"encrypted_content,omitempty"`
}

type responseOutputItem any

type responseUsage struct {
	InputTokens        int64                       `json:"input_tokens"`
	OutputTokens       int64                       `json:"output_tokens"`
	TotalTokens        int64                       `json:"total_tokens"`
	InputTokensDetails responseInputTokensDetails  `json:"input_tokens_details"`
	OutputTokenDetails responseOutputTokensDetails `json:"output_tokens_details"`
}

type responseInputTokensDetails struct {
	CachedTokens int64 `json:"cached_tokens"`
}

type responseOutputTokensDetails struct {
	ReasoningTokens int64 `json:"reasoning_tokens"`
}

type responseObject struct {
	ID     string             `json:"id"`
	Object string             `json:"object"`
	Model  string             `json:"model"`
	Status string             `json:"status"`
	Output []responseOutputItem `json:"output"`
	Usage  *responseUsage     `json:"usage,omitempty"`
}

// Streaming event types.

type responseStreamEvent struct {
	Type         string          `json:"type"`
	Response     *responseObject `json:"response,omitempty"`
	OutputIndex  *int64          `json:"output_index,omitempty"`
	ContentIndex *int64          `json:"content_index,omitempty"`
	ItemID       string          `json:"item_id,omitempty"`
	SummaryIndex *int64          `json:"summary_index,omitempty"`
	Item         any             `json:"item,omitempty"`
	Part         any             `json:"part,omitempty"`
	Delta        string          `json:"delta,omitempty"`
	Text         string          `json:"text,omitempty"`
	Arguments    string          `json:"arguments,omitempty"`
}
