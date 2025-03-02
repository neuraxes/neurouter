package biz

import (
	"context"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
)

type ChatReq v1.ChatReq
type ChatResp v1.ChatResp
type ModelSpec v1.ModelSpec

type ChatStreamServer interface {
	Send(*ChatResp) error
}

type ChatStreamClient interface {
	Recv() (*ChatResp, error)
	Close() error
}

type ChatRepo interface {
	Chat(context.Context, *ChatReq) (*ChatResp, error)
	ChatStream(context.Context, *ChatReq) (ChatStreamClient, error)
}
