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

package entity

import v1 "github.com/neuraxes/neurouter/api/neurouter/v1"

// ChatRequest represents a chat request, aliased from the API proto definition.
type ChatRequest = v1.ChatRequest

// ChatResponse represents a chat response, aliased from the API proto definition.
type ChatResponse = v1.ChatResponse

// ChatEvent represents a streaming chat event, aliased from the API proto definition.
type ChatEvent = v1.ChatEvent

// ModelSpec represents a model specification, aliased from the API proto definition.
type ModelSpec = v1.ModelSpec

// EmbedRequest represents an embedding request, aliased from the API proto definition.
type EmbedRequest = v1.EmbedRequest

// EmbedResponse represents an embedding response, aliased from the API proto definition.
type EmbedResponse = v1.EmbedResponse
