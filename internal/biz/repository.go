package biz

import (
	"context"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
)

type ChatReq v1.ChatReq
type ChatResp v1.ChatResp

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
