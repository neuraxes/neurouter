package neurouter

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type ChatRepo struct {
	config *conf.NeurouterConfig
	client v1.ChatClient
	log    *log.Helper
}

func NewNeurouterChatRepoFactory() biz.NeurouterChatRepoFactory {
	return NewNeurouterChatRepo
}

func NewNeurouterChatRepo(config *conf.NeurouterConfig, logger log.Logger) (biz.ChatRepo, error) {
	conn, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint(config.Endpoint),
	)
	if err != nil {
		return nil, err
	}

	return &ChatRepo{
		config: config,
		client: v1.NewChatClient(conn),
		log:    log.NewHelper(logger),
	}, nil
}

func (r *ChatRepo) Chat(ctx context.Context, req *biz.ChatReq) (*biz.ChatResp, error) {
	resp, err := r.client.Chat(ctx, (*v1.ChatReq)(req))
	if err != nil {
		return nil, err
	}
	return (*biz.ChatResp)(resp), nil
}

type neurouterChatStreamClient struct {
	stream v1.Chat_ChatStreamClient
}

func (c *neurouterChatStreamClient) Recv() (*biz.ChatResp, error) {
	resp, err := c.stream.Recv()
	if err != nil {
		return nil, err
	}
	return (*biz.ChatResp)(resp), nil
}

func (c *neurouterChatStreamClient) Close() error {
	return nil
}

func (r *ChatRepo) ChatStream(ctx context.Context, req *biz.ChatReq) (biz.ChatStreamClient, error) {
	stream, err := r.client.ChatStream(ctx, (*v1.ChatReq)(req))
	if err != nil {
		return nil, err
	}

	return &neurouterChatStreamClient{
		stream: stream,
	}, nil
}
