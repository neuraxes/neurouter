package model

import (
	"context"
	"iter"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

// mockChatRepo implements repository.ChatRepo for testing.
type mockChatRepo struct{}

func (m *mockChatRepo) Chat(context.Context, *entity.ChatReq) (*entity.ChatResp, error) {
	return nil, nil
}

func (m *mockChatRepo) ChatStream(context.Context, *entity.ChatReq) iter.Seq2[*entity.ChatResp, error] {
	return nil
}

var _ repository.ChatRepo = (*mockChatRepo)(nil)

// mockEmbeddingRepo implements repository.EmbeddingRepo for testing.
type mockEmbeddingRepo struct{}

func (m *mockEmbeddingRepo) Embed(context.Context, *entity.EmbedReq) (*entity.EmbedResp, error) {
	return nil, nil
}

var _ repository.EmbeddingRepo = (*mockEmbeddingRepo)(nil)

// mockChatEmbeddingRepo implements both ChatRepo and EmbeddingRepo for testing.
type mockChatEmbeddingRepo struct {
	mockChatRepo
	mockEmbeddingRepo
}
