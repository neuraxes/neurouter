package repository

import (
	"context"

	"git.xdea.xyz/Turing/neurouter/internal/biz/entity"
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
