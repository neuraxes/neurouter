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

package repository

import (
	"context"

	"github.com/neuraxes/neurouter/internal/biz/entity"
)

// ChatStreamServer defines the server-side interface for sending chat responses.
type ChatStreamServer interface {
	Send(*entity.ChatResp) error
}

// ChatStreamClient defines the client-side interface for receiving chat responses.
type ChatStreamClient interface {
	Recv() (*entity.ChatResp, error)
	Close() error
}

// ChatRepo defines the interface for chat operations.
// It supports both synchronous chat and streaming chat interactions.
type ChatRepo interface {
	// Chat performs a synchronous chat interaction.
	Chat(context.Context, *entity.ChatReq) (*entity.ChatResp, error)
	// ChatStream initiates a streaming chat interaction.
	ChatStream(context.Context, *entity.ChatReq) (ChatStreamClient, error)
}
