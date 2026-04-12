package model

import (
	"context"
	"errors"
	"iter"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"google.golang.org/protobuf/proto"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type mockKratosConfig struct {
	upstream *conf.Upstream
}

func (m *mockKratosConfig) Load() error      { return nil }
func (m *mockKratosConfig) Scan(v any) error { return nil }
func (m *mockKratosConfig) Value(key string) config.Value {
	return &mockConfigValue{upstream: m.upstream}
}
func (m *mockKratosConfig) Watch(key string, o config.Observer) error { return nil }
func (m *mockKratosConfig) Close() error                              { return nil }

type mockConfigValue struct {
	upstream *conf.Upstream
}

func (v *mockConfigValue) Bool() (bool, error)                   { return false, nil }
func (v *mockConfigValue) Int() (int64, error)                   { return 0, nil }
func (v *mockConfigValue) Float() (float64, error)               { return 0, nil }
func (v *mockConfigValue) String() (string, error)               { return "", nil }
func (v *mockConfigValue) Duration() (time.Duration, error)      { return 0, nil }
func (v *mockConfigValue) Slice() ([]config.Value, error)        { return nil, nil }
func (v *mockConfigValue) Map() (map[string]config.Value, error) { return nil, nil }
func (v *mockConfigValue) Load() any                             { return nil }
func (v *mockConfigValue) Store(any)                             {}
func (v *mockConfigValue) Scan(dst any) error {
	if v.upstream == nil {
		return errors.New("no upstream config")
	}
	proto.Merge(dst.(proto.Message), v.upstream)
	return nil
}

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
