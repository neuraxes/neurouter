package model

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/embedding"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
	"github.com/neuraxes/neurouter/internal/data/limiter/local"
)

func TestEstimateEmbeddingTokens(t *testing.T) {
	Convey("Test estimateEmbeddingTokens", t, func() {
		Convey("with nil contents should return 0", func() {
			req := &v1.EmbedReq{}
			So(estimateEmbeddingTokens(req), ShouldEqual, 0)
		})

		Convey("with empty contents should return 0", func() {
			req := &v1.EmbedReq{
				Contents: []*v1.Content{},
			}
			So(estimateEmbeddingTokens(req), ShouldEqual, 0)
		})

		Convey("with text content should estimate tokens", func() {
			req := &v1.EmbedReq{
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Hello, world!"}},
				},
			}
			So(estimateEmbeddingTokens(req), ShouldEqual, 13/4+1)
		})

		Convey("with multiple text contents should sum all text", func() {
			req := &v1.EmbedReq{
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "Hello"}},    // 5 chars
					{Content: &v1.Content_Text{Text: "World!!!"}}, // 8 chars
				},
			}
			So(estimateEmbeddingTokens(req), ShouldEqual, (5+8)/4+1)
		})

		Convey("with non-text content should ignore it", func() {
			req := &v1.EmbedReq{
				Contents: []*v1.Content{
					{Content: &v1.Content_Image{Image: &v1.Image{}}},
				},
			}
			So(estimateEmbeddingTokens(req), ShouldEqual, 0)
		})
	})
}

func TestEmbeddingModel_EmbeddingRepo(t *testing.T) {
	Convey("Test embeddingModel EmbeddingRepo", t, func() {
		Convey("should return the embedding repo", func() {
			m := &embeddingModel{
				model:        &model{embeddingRepo: &mockEmbeddingRepo{}},
				reservations: &reservationSet{},
			}
			So(m.EmbeddingRepo(), ShouldNotBeNil)
		})

		Convey("should return nil if no embedding repo", func() {
			m := &embeddingModel{
				model:        &model{},
				reservations: &reservationSet{},
			}
			So(m.EmbeddingRepo(), ShouldBeNil)
		})
	})
}

func TestEmbeddingModel_RecordUsage(t *testing.T) {
	Convey("Test embeddingModel RecordUsage", t, func() {
		Convey("should complete reservations with actual tokens", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			r, _ := concurrency.Reserve()

			m := &embeddingModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					requestReservations: []repository.Reservation{r},
				},
			}
			So(concurrency.Probe(), ShouldBeGreaterThan, 0)

			m.RecordUsage(context.Background(), 100)

			// Reservation should be completed
			So(concurrency.Probe(), ShouldEqual, 0)
		})

		Convey("should complete token reservations with actual usage", func() {
			tpmLimiter := local.NewTPMLimiter(10000)
			r, _ := tpmLimiter.Reserve(500)

			m := &embeddingModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					tokenReservations: []repository.TokenReservation{r},
				},
			}

			m.RecordUsage(context.Background(), 300)

			So(m.reservations.tokenReservations, ShouldBeNil)
			So(tpmLimiter.Probe(9700), ShouldEqual, 0)
		})

		Convey("with zero tokens should still complete reservations", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			r, _ := concurrency.Reserve()

			m := &embeddingModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					requestReservations: []repository.Reservation{r},
				},
			}

			m.RecordUsage(context.Background(), 0)

			So(concurrency.Probe(), ShouldEqual, 0)
		})

		Convey("should record OTel metrics with actual input tokens", func() {
			metrics, reader := newTestMetrics()

			m := &embeddingModel{
				model: &model{
					config:         &conf.Model{Id: "text-embedding-ada"},
					upstreamConfig: &conf.UpstreamConfig{Name: "openai"},
					metrics:        metrics,
				},
				reservations: &reservationSet{},
			}

			m.RecordUsage(context.Background(), 300)

			data := collectMetrics(reader)
			So(data["neurouter_input_tokens_total"], ShouldHaveLength, 1)
			So(data["neurouter_input_tokens_total"][0].Value, ShouldEqual, 300)
			So(data["neurouter_requests_total"], ShouldHaveLength, 1)

			m.Close()
		})

		Convey("should use estimated tokens for OTel metrics when actual is zero", func() {
			metrics, reader := newTestMetrics()

			m := &embeddingModel{
				model: &model{
					config:         &conf.Model{Id: "text-embedding-ada"},
					upstreamConfig: &conf.UpstreamConfig{Name: "openai"},
					metrics:        metrics,
				},
				reservations:    &reservationSet{},
				estimatedTokens: 150,
			}

			m.RecordUsage(context.Background(), 0)

			data := collectMetrics(reader)
			So(data["neurouter_input_tokens_total"], ShouldHaveLength, 1)
			So(data["neurouter_input_tokens_total"][0].Value, ShouldEqual, 150)

			m.Close()
		})
	})
}

func TestEmbeddingModel_Close(t *testing.T) {
	Convey("Test embeddingModel Close", t, func() {
		Convey("should cancel all unreleased reservations", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			r, _ := concurrency.Reserve()

			m := &embeddingModel{
				model: &model{},
				reservations: &reservationSet{
					requestReservations: []repository.Reservation{r},
				},
			}
			So(concurrency.Probe(), ShouldBeGreaterThan, 0)

			m.Close()

			So(concurrency.Probe(), ShouldEqual, 0)
		})

		Convey("should be safe to call after RecordUsage", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			r, _ := concurrency.Reserve()

			m := &embeddingModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					requestReservations: []repository.Reservation{r},
				},
			}

			m.RecordUsage(context.Background(), 100)
			So(func() { m.Close() }, ShouldNotPanic)
		})

		Convey("should be safe to call multiple times", func() {
			m := &embeddingModel{
				model:        &model{},
				reservations: &reservationSet{},
			}
			So(func() {
				m.Close()
				m.Close()
			}, ShouldNotPanic)
		})
	})
}

func TestElectForEmbedding(t *testing.T) {
	Convey("Test ElectForEmbedding", t, func() {
		Convey("with no models should return error", func() {
			uc := &UseCaseImpl{
				models: nil,
				log:    log.NewHelper(log.DefaultLogger),
			}
			_, err := uc.ElectForEmbedding(context.Background(), &v1.EmbedReq{Model: "test"})
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("should match model by ID", func() {
			m := &model{
				config: &conf.Model{
					Id:           "text-embedding-ada",
					Capabilities: []conf.Capability{conf.Capability_CAPABILITY_EMBEDDING},
				},
				upstreamConfig:   &conf.UpstreamConfig{Name: "openai"},
				embeddingRepo:    &mockEmbeddingRepo{},
				upstreamLimiters: &limiterGroup{},
				modelLimiters:    &limiterGroup{},
			}
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			result, err := uc.ElectForEmbedding(context.Background(), &v1.EmbedReq{Model: "text-embedding-ada"})
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			defer result.Close()
		})

		Convey("should fallback when requested model not found", func() {
			m := &model{
				config: &conf.Model{
					Id:           "text-embedding-ada",
					Capabilities: []conf.Capability{conf.Capability_CAPABILITY_EMBEDDING},
				},
				upstreamConfig:   &conf.UpstreamConfig{Name: "openai"},
				embeddingRepo:    &mockEmbeddingRepo{},
				upstreamLimiters: &limiterGroup{},
				modelLimiters:    &limiterGroup{},
			}
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			req := &v1.EmbedReq{Model: "nonexistent"}
			result, err := uc.ElectForEmbedding(context.Background(), req)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			defer result.Close()
		})

		Convey("should skip models without embedding capability", func() {
			m := &model{
				config: &conf.Model{
					Id:           "chat-model",
					Capabilities: []conf.Capability{conf.Capability_CAPABILITY_CHAT},
				},
				upstreamConfig:   &conf.UpstreamConfig{Name: "openai"},
				embeddingRepo:    &mockEmbeddingRepo{},
				upstreamLimiters: &limiterGroup{},
				modelLimiters:    &limiterGroup{},
			}
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			_, err := uc.ElectForEmbedding(context.Background(), &v1.EmbedReq{Model: "chat-model"})
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("should skip models without embedding repo", func() {
			m := &model{
				config: &conf.Model{
					Id:           "no-repo",
					Capabilities: []conf.Capability{conf.Capability_CAPABILITY_EMBEDDING},
				},
				upstreamConfig:   &conf.UpstreamConfig{Name: "test"},
				embeddingRepo:    nil,
				upstreamLimiters: &limiterGroup{},
				modelLimiters:    &limiterGroup{},
			}
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			_, err := uc.ElectForEmbedding(context.Background(), &v1.EmbedReq{Model: "no-repo"})
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("should update model ID to upstream ID", func() {
			m := &model{
				config: &conf.Model{
					Id:           "my-embed",
					UpstreamId:   "text-embedding-3-large",
					Capabilities: []conf.Capability{conf.Capability_CAPABILITY_EMBEDDING},
				},
				upstreamConfig:   &conf.UpstreamConfig{Name: "openai"},
				embeddingRepo:    &mockEmbeddingRepo{},
				upstreamLimiters: &limiterGroup{},
				modelLimiters:    &limiterGroup{},
			}
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			req := &v1.EmbedReq{Model: "my-embed"}
			result, err := uc.ElectForEmbedding(context.Background(), req)
			So(err, ShouldBeNil)
			defer result.Close()
			So(req.Model, ShouldEqual, "text-embedding-3-large")
		})

		Convey("should respect limiter during election", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			m := &model{
				config: &conf.Model{
					Id:           "ada",
					Capabilities: []conf.Capability{conf.Capability_CAPABILITY_EMBEDDING},
				},
				upstreamConfig: &conf.UpstreamConfig{Name: "openai"},
				embeddingRepo:  &mockEmbeddingRepo{},
				upstreamLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{concurrency},
				},
				modelLimiters: &limiterGroup{},
			}
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			result1, err := uc.ElectForEmbedding(context.Background(), &v1.EmbedReq{Model: "ada"})
			So(err, ShouldBeNil)

			// Second election should wait (concurrency exhausted but waitable)
			done := make(chan struct{})
			var model embedding.Model
			go func() {
				model, err = uc.ElectForEmbedding(context.Background(), &v1.EmbedReq{Model: "ada"})
				close(done)
			}()

			// Allow election to start waiting
			time.Sleep(10 * time.Millisecond)

			// Release first, which should unblock second
			result1.Close()

			<-done
			So(err, ShouldBeNil)
			So(model, ShouldNotBeNil)
		})
	})
}
