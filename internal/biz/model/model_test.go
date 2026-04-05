package model

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel/metric/noop"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

func TestNewModelUseCase(t *testing.T) {
	Convey("Test NewModelUseCase", t, func() {
		// Factories that return mock repos
		openAIFactory := func(config *conf.OpenAIConfig, logger log.Logger) (repository.Repo, error) {
			return &mockChatEmbeddingRepo{}, nil
		}
		anthropicFactory := func(config *conf.AnthropicConfig, logger log.Logger) (repository.Repo, error) {
			return &mockChatRepo{}, nil
		}
		googleFactory := func(config *conf.GoogleConfig, logger log.Logger) (repository.Repo, error) {
			return &mockChatRepo{}, nil
		}
		neurouterFactory := func(config *conf.NeurouterConfig, logger log.Logger) (repository.Repo, error) {
			return &mockChatRepo{}, nil
		}

		Convey("with nil config should return empty use case", func() {
			uc := NewModelUseCase(nil, anthropicFactory, googleFactory, neurouterFactory, openAIFactory, noop.NewMeterProvider(), log.DefaultLogger)
			So(uc, ShouldNotBeNil)
			So(uc.models, ShouldBeEmpty)
			So(uc.aliases, ShouldBeEmpty)
		})

		Convey("with empty configs should return empty use case", func() {
			c := &conf.Upstream{
				Configs: []*conf.UpstreamConfig{},
			}
			uc := NewModelUseCase(c, anthropicFactory, googleFactory, neurouterFactory, openAIFactory, noop.NewMeterProvider(), log.DefaultLogger)
			So(uc, ShouldNotBeNil)
			So(uc.models, ShouldBeEmpty)
			So(uc.aliases, ShouldBeEmpty)
		})

		Convey("with OpenAI config should create models", func() {
			c := &conf.Upstream{
				Configs: []*conf.UpstreamConfig{
					{
						Name: "openai",
						Models: []*conf.Model{
							{
								Id:           "gpt-4",
								Capabilities: []conf.Capability{conf.Capability_CAPABILITY_CHAT},
							},
							{
								Id:           "text-embedding-ada",
								Capabilities: []conf.Capability{conf.Capability_CAPABILITY_EMBEDDING},
							},
						},
						Config: &conf.UpstreamConfig_OpenAi{
							OpenAi: &conf.OpenAIConfig{},
						},
					},
				},
			}

			uc := NewModelUseCase(c, anthropicFactory, googleFactory, neurouterFactory, openAIFactory, noop.NewMeterProvider(), log.DefaultLogger)
			So(len(uc.models), ShouldEqual, 2)
			So(uc.models[0].config.Id, ShouldEqual, "gpt-4")
			So(uc.models[1].config.Id, ShouldEqual, "text-embedding-ada")

			// ChatRepo should be set for both (since mockChatEmbeddingRepo implements both)
			So(uc.models[0].chatRepo, ShouldNotBeNil)
			So(uc.models[0].embeddingRepo, ShouldNotBeNil)
			So(uc.models[1].chatRepo, ShouldNotBeNil)
			So(uc.models[1].embeddingRepo, ShouldNotBeNil)
		})

		Convey("with Anthropic config should create models", func() {
			c := &conf.Upstream{
				Configs: []*conf.UpstreamConfig{
					{
						Name: "anthropic",
						Models: []*conf.Model{
							{
								Id:           "claude-3",
								Capabilities: []conf.Capability{conf.Capability_CAPABILITY_CHAT},
							},
						},
						Config: &conf.UpstreamConfig_Anthropic{
							Anthropic: &conf.AnthropicConfig{},
						},
					},
				},
			}

			uc := NewModelUseCase(c, anthropicFactory, googleFactory, neurouterFactory, openAIFactory, noop.NewMeterProvider(), log.DefaultLogger)
			So(len(uc.models), ShouldEqual, 1)
			So(uc.models[0].config.Id, ShouldEqual, "claude-3")
			So(uc.models[0].chatRepo, ShouldNotBeNil)
			// Anthropic factory returns mockChatRepo (not embedding-capable)
			So(uc.models[0].embeddingRepo, ShouldBeNil)
		})

		Convey("with scheduling should create limiter groups", func() {
			c := &conf.Upstream{
				Configs: []*conf.UpstreamConfig{
					{
						Name: "openai",
						Scheduling: &conf.UpstreamScheduling{
							ConcurrencyLimit: 10,
							RpmLimit:         100,
						},
						Models: []*conf.Model{
							{
								Id:           "gpt-4",
								Capabilities: []conf.Capability{conf.Capability_CAPABILITY_CHAT},
								Scheduling: &conf.ModelScheduling{
									TpmLimit: 50000,
								},
							},
						},
						Config: &conf.UpstreamConfig_OpenAi{
							OpenAi: &conf.OpenAIConfig{},
						},
					},
				},
			}

			uc := NewModelUseCase(c, anthropicFactory, googleFactory, neurouterFactory, openAIFactory, noop.NewMeterProvider(), log.DefaultLogger)
			So(len(uc.models), ShouldEqual, 1)
			// Upstream limiters should have concurrency + rpm
			So(len(uc.models[0].upstreamLimiters.requestLimiters), ShouldEqual, 2)
			// Model limiters should have tpm
			So(len(uc.models[0].modelLimiters.tokenLimiters), ShouldEqual, 1)
		})

		Convey("models in same upstream should share upstream limiters", func() {
			c := &conf.Upstream{
				Configs: []*conf.UpstreamConfig{
					{
						Name: "openai",
						Scheduling: &conf.UpstreamScheduling{
							ConcurrencyLimit: 5,
						},
						Models: []*conf.Model{
							{Id: "model-a", Capabilities: []conf.Capability{conf.Capability_CAPABILITY_CHAT}},
							{Id: "model-b", Capabilities: []conf.Capability{conf.Capability_CAPABILITY_CHAT}},
						},
						Config: &conf.UpstreamConfig_OpenAi{
							OpenAi: &conf.OpenAIConfig{},
						},
					},
				},
			}

			uc := NewModelUseCase(c, anthropicFactory, googleFactory, neurouterFactory, openAIFactory, noop.NewMeterProvider(), log.DefaultLogger)
			So(len(uc.models), ShouldEqual, 2)
			// Both models should share the same upstream limiter group pointer
			So(uc.models[0].upstreamLimiters, ShouldPointTo, uc.models[1].upstreamLimiters)
		})

		Convey("with factory error should skip that upstream", func() {
			failFactory := func(config *conf.OpenAIConfig, logger log.Logger) (repository.Repo, error) {
				return nil, v1.ErrorNoUpstream("factory error")
			}

			c := &conf.Upstream{
				Configs: []*conf.UpstreamConfig{
					{
						Name: "failing",
						Models: []*conf.Model{
							{Id: "model", Capabilities: []conf.Capability{conf.Capability_CAPABILITY_CHAT}},
						},
						Config: &conf.UpstreamConfig_OpenAi{
							OpenAi: &conf.OpenAIConfig{},
						},
					},
				},
			}

			uc := NewModelUseCase(c, anthropicFactory, googleFactory, neurouterFactory, failFactory, noop.NewMeterProvider(), log.DefaultLogger)
			So(uc.models, ShouldBeEmpty)
		})
	})
}

func TestListAvailableModels(t *testing.T) {
	Convey("Test ListAvailableModels", t, func() {
		Convey("with no models should return empty list", func() {
			uc := &UseCaseImpl{
				models: nil,
				log:    log.NewHelper(log.DefaultLogger),
			}
			models, err := uc.ListAvailableModels(context.Background())
			So(err, ShouldBeNil)
			So(models, ShouldBeEmpty)
		})

		Convey("should return all models with correct fields", func() {
			uc := &UseCaseImpl{
				models: []*model{
					{
						config: &conf.Model{
							Id:       "gpt-4",
							Name:     "GPT-4",
							Owner:    "openai",
							Provider: "openai",
							Modalities: []conf.Modality{
								conf.Modality_MODALITY_TEXT,
								conf.Modality_MODALITY_IMAGE,
							},
							Capabilities: []conf.Capability{
								conf.Capability_CAPABILITY_CHAT,
							},
						},
					},
					{
						config: &conf.Model{
							Id:       "text-embedding-ada",
							Name:     "Text Embedding Ada",
							Owner:    "openai",
							Provider: "openai",
							Modalities: []conf.Modality{
								conf.Modality_MODALITY_TEXT,
							},
							Capabilities: []conf.Capability{
								conf.Capability_CAPABILITY_EMBEDDING,
							},
						},
					},
				},
				log: log.NewHelper(log.DefaultLogger),
			}

			models, err := uc.ListAvailableModels(context.Background())
			So(err, ShouldBeNil)
			So(len(models), ShouldEqual, 2)

			So(models[0].Id, ShouldEqual, "gpt-4")
			So(models[0].Name, ShouldEqual, "GPT-4")
			So(models[0].Owner, ShouldEqual, "openai")
			So(models[0].Provider, ShouldEqual, "openai")
			So(models[0].Modalities, ShouldResemble, []v1.Modality{v1.Modality_MODALITY_TEXT, v1.Modality_MODALITY_IMAGE})
			So(models[0].Capabilities, ShouldResemble, []v1.Capability{v1.Capability_CAPABILITY_CHAT})

			So(models[1].Id, ShouldEqual, "text-embedding-ada")
			So(models[1].Name, ShouldEqual, "Text Embedding Ada")
			So(models[1].Owner, ShouldEqual, "openai")
			So(models[1].Provider, ShouldEqual, "openai")
			So(models[1].Modalities, ShouldResemble, []v1.Modality{v1.Modality_MODALITY_TEXT})
			So(models[1].Capabilities, ShouldResemble, []v1.Capability{v1.Capability_CAPABILITY_EMBEDDING})
		})
	})
}
