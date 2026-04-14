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

import (
	"github.com/openai/openai-go/v3"
)

// Response DTOs matching the OpenAI REST API JSON schema.
// Defined locally instead of reusing the official SDK's response types because
// those are designed for parsing (not construction) and lack fields like
// reasoning_content that deepseek and openrouter support.

type functionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function functionCall `json:"function"`
}

type chatCompletionMessage struct {
	Role             string     `json:"role"`
	Content          string     `json:"content"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	ToolCalls        []toolCall `json:"tool_calls,omitempty"`
}

type chatCompletionChoice struct {
	Message      chatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

type chatCompletionResponse struct {
	ID      string                  `json:"id"`
	Object  string                  `json:"object"`
	Model   string                  `json:"model"`
	Choices []chatCompletionChoice  `json:"choices"`
	Usage   *openai.CompletionUsage `json:"usage,omitempty"`
}

type chatCompletionChunkDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
}

type chatCompletionChunkChoice struct {
	Delta        chatCompletionChunkDelta `json:"delta"`
	FinishReason string                   `json:"finish_reason"`
}

type chatCompletionChunk struct {
	ID      string                      `json:"id"`
	Object  string                      `json:"object"`
	Model   string                      `json:"model"`
	Choices []chatCompletionChunkChoice `json:"choices"`
	Usage   *openai.CompletionUsage     `json:"usage,omitempty"`
}

type embeddingResponse struct {
	Object string             `json:"object"`
	Model  string             `json:"model"`
	Data   []openai.Embedding `json:"data"`
}

type modelsList struct {
	Object string         `json:"object"`
	Models []openai.Model `json:"data"`
}
