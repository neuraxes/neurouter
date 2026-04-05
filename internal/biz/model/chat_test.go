package model

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/chat"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
	"github.com/neuraxes/neurouter/internal/data/limiter/local"
)

func TestEstimateTokens(t *testing.T) {
	Convey("Test estimateTokens", t, func() {
		Convey("with nil request should return 0", func() {
			req := &v1.ChatReq{}
			So(estimateTokens(req), ShouldEqual, 0)
		})

		Convey("with empty messages should return 0", func() {
			req := &v1.ChatReq{
				Messages: []*v1.Message{},
			}
			So(estimateTokens(req), ShouldEqual, 0)
		})

		Convey("with text content should estimate tokens", func() {
			req := &v1.ChatReq{
				Messages: []*v1.Message{
					{
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Hello, world!"}}, // 13 chars -> 13/4+1 = 4
						},
					},
				},
			}
			So(estimateTokens(req), ShouldEqual, 4) // 13/4 + 1
		})

		Convey("with multiple messages should sum all text", func() {
			req := &v1.ChatReq{
				Messages: []*v1.Message{
					{
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Hello"}}, // 5 chars
						},
					},
					{
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "World!!!"}}, // 8 chars
						},
					},
				},
			}
			// (5+8)/4 + 1 = 4
			So(estimateTokens(req), ShouldEqual, 4)
		})

		Convey("with image content should assign fixed tokens", func() {
			req := &v1.ChatReq{
				Messages: []*v1.Message{
					{
						Contents: []*v1.Content{
							{Content: &v1.Content_Image{Image: &v1.Image{}}},
						},
					},
				},
			}
			So(estimateTokens(req), ShouldEqual, 768)
		})

		Convey("with mixed text and image content", func() {
			req := &v1.ChatReq{
				Messages: []*v1.Message{
					{
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Describe this"}}, // 13 chars -> 4 tokens
							{Content: &v1.Content_Image{Image: &v1.Image{}}},
						},
					},
				},
			}
			So(estimateTokens(req), ShouldEqual, 4+768)
		})

		Convey("with tool use content should count name and inputs", func() {
			req := &v1.ChatReq{
				Messages: []*v1.Message{
					{
						Contents: []*v1.Content{
							{
								Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
									Name: "search", // 6 chars
									Inputs: []*v1.ToolUse_Input{
										{Input: &v1.ToolUse_Input_Text{Text: "test"}}, // 4 chars
									},
								}},
							},
						},
					},
				},
			}
			So(estimateTokens(req), ShouldEqual, (6+4)/4+1)
		})

		Convey("with tool result content should count output text", func() {
			req := &v1.ChatReq{
				Messages: []*v1.Message{
					{
						Contents: []*v1.Content{
							{
								Content: &v1.Content_ToolResult{ToolResult: &v1.ToolResult{
									Outputs: []*v1.ToolResult_Output{
										{Output: &v1.ToolResult_Output_Text{Text: "result data"}}, // 11 chars
									},
								}},
							},
						},
					},
				},
			}
			So(estimateTokens(req), ShouldEqual, (11/4 + 1))
		})

		Convey("with tool result image should count image tokens", func() {
			req := &v1.ChatReq{
				Messages: []*v1.Message{
					{
						Contents: []*v1.Content{
							{
								Content: &v1.Content_ToolResult{ToolResult: &v1.ToolResult{
									Outputs: []*v1.ToolResult_Output{
										{Output: &v1.ToolResult_Output_Image{Image: &v1.Image{}}},
									},
								}},
							},
						},
					},
				},
			}
			So(estimateTokens(req), ShouldEqual, 768)
		})
	})
}

func TestChatModel_ChatRepo(t *testing.T) {
	Convey("Test chatModel ChatRepo", t, func() {
		Convey("should return the chat repo", func() {
			m := &chatModel{
				model:        &model{chatRepo: &mockChatRepo{}},
				reservations: &reservationSet{},
			}
			So(m.ChatRepo(), ShouldNotBeNil)
		})

		Convey("should return nil if no chat repo", func() {
			m := &chatModel{
				model:        &model{},
				reservations: &reservationSet{},
			}
			So(m.ChatRepo(), ShouldBeNil)
		})
	})
}

func TestChatModel_RecordUsage(t *testing.T) {
	Convey("Test chatModel RecordUsage", t, func() {
		Convey("with nil stats should complete with estimated tokens", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			r, _ := concurrency.Reserve()

			m := &chatModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					requestReservations: []repository.Reservation{r},
				},
				estimatedTokens: 100,
			}
			So(concurrency.Probe(), ShouldBeGreaterThan, 0)

			m.RecordUsage(context.Background(), nil)

			So(concurrency.Probe(), ShouldEqual, 0)
			m.Close()
		})

		Convey("with nil usage should complete with estimated tokens", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			r, _ := concurrency.Reserve()

			m := &chatModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					requestReservations: []repository.Reservation{r},
				},
				estimatedTokens: 50,
			}
			So(concurrency.Probe(), ShouldBeGreaterThan, 0)

			m.RecordUsage(context.Background(), &v1.Statistics{})

			So(concurrency.Probe(), ShouldEqual, 0)
			m.Close()
		})

		Convey("should record token counts and complete reservations", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			r, _ := concurrency.Reserve()

			m := &chatModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					requestReservations: []repository.Reservation{r},
				},
			}
			So(concurrency.Probe(), ShouldBeGreaterThan, 0)

			m.RecordUsage(context.Background(), &v1.Statistics{
				Usage: &v1.Statistics_Usage{
					InputTokens:       100,
					OutputTokens:      50,
					CachedInputTokens: 10,
					ReasoningTokens:   7,
				},
			})

			So(m.inputTokens.Load(), ShouldEqual, 100)
			So(m.outputTokens.Load(), ShouldEqual, 50)
			So(m.cachedInputTokens.Load(), ShouldEqual, 10)
			So(m.reasoningTokens.Load(), ShouldEqual, 7)
			So(concurrency.Probe(), ShouldEqual, 0)
		})

		Convey("should complete token reservations with actual usage", func() {
			tpmLimiter := local.NewDailyTokenLimiter(10000)
			r, _ := tpmLimiter.Reserve(500)

			m := &chatModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					tokenReservations: []repository.TokenReservation{r},
				},
			}

			m.RecordUsage(context.Background(), &v1.Statistics{
				Usage: &v1.Statistics_Usage{
					InputTokens:  200,
					OutputTokens: 100,
				},
			})

			So(m.reservations.tokenReservations, ShouldBeNil)
			So(tpmLimiter.Probe(9700), ShouldEqual, time.Duration(0))
		})

		Convey("should record OTel token and request metrics when usage exists", func() {
			metrics, reader := newTestMetrics()

			m := &chatModel{
				model: &model{
					config:         &conf.Model{Id: "gpt-4"},
					upstreamConfig: &conf.UpstreamConfig{Name: "openai"},
					metrics:        metrics,
				},
				reservations: &reservationSet{},
			}

			m.RecordUsage(context.Background(), &v1.Statistics{
				Usage: &v1.Statistics_Usage{
					InputTokens:       100,
					OutputTokens:      50,
					CachedInputTokens: 10,
					ReasoningTokens:   20,
				},
			})

			data := collectMetrics(reader)
			So(data["neurouter_input_tokens_total"], ShouldHaveLength, 1)
			So(data["neurouter_input_tokens_total"][0].Value, ShouldEqual, 100)
			So(data["neurouter_output_tokens_total"][0].Value, ShouldEqual, 50)
			So(data["neurouter_cached_input_tokens_total"][0].Value, ShouldEqual, 10)
			So(data["neurouter_reasoning_tokens_total"][0].Value, ShouldEqual, 20)
			So(data["neurouter_requests_total"], ShouldHaveLength, 1)
			So(data["neurouter_requests_total"][0].Value, ShouldEqual, 1)

			m.Close()
		})

		Convey("should record only request metric when stats are nil", func() {
			metrics, reader := newTestMetrics()

			m := &chatModel{
				model: &model{
					config:         &conf.Model{Id: "gpt-4"},
					upstreamConfig: &conf.UpstreamConfig{Name: "openai"},
					metrics:        metrics,
				},
				reservations: &reservationSet{},
			}

			m.RecordUsage(context.Background(), nil)

			data := collectMetrics(reader)
			So(data["neurouter_input_tokens_total"], ShouldBeEmpty)
			So(data["neurouter_requests_total"], ShouldHaveLength, 1)
			So(data["neurouter_requests_total"][0].Value, ShouldEqual, 1)

			m.Close()
		})
	})
}

func TestChatModel_Close(t *testing.T) {
	Convey("Test chatModel Close", t, func() {
		Convey("should cancel all unreleased reservations", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			r, _ := concurrency.Reserve()

			m := &chatModel{
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

			m := &chatModel{
				model: &model{
					config:         &conf.Model{Id: "test"},
					upstreamConfig: &conf.UpstreamConfig{Name: "test"},
				},
				reservations: &reservationSet{
					requestReservations: []repository.Reservation{r},
				},
			}

			m.RecordUsage(context.Background(), &v1.Statistics{
				Usage: &v1.Statistics_Usage{
					InputTokens:  10,
					OutputTokens: 5,
				},
			})

			// Close after RecordUsage should be a no-op (reservations already completed)
			So(func() { m.Close() }, ShouldNotPanic)
		})

		Convey("should be safe to call multiple times", func() {
			m := &chatModel{
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

func makeModel(id string, upstreamID string, capabilities []conf.Capability) *model {
	return &model{
		config: &conf.Model{
			Id:           id,
			UpstreamId:   upstreamID,
			Capabilities: capabilities,
		},
		upstreamConfig:   &conf.UpstreamConfig{Name: "openai"},
		chatRepo:         &mockChatRepo{},
		upstreamLimiters: &limiterGroup{},
		modelLimiters:    &limiterGroup{},
	}
}

func TestElectForChat(t *testing.T) {
	Convey("Test ElectForChat", t, func() {
		Convey("with no models should return error", func() {
			uc := &UseCaseImpl{
				models: nil,
				log:    log.NewHelper(log.DefaultLogger),
			}
			_, err := uc.ElectForChat(context.Background(), &v1.ChatReq{Model: "test"})
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("should match model by ID", func() {
			m := makeModel("gpt", "gpt-4", []conf.Capability{conf.Capability_CAPABILITY_CHAT})
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			result, err := uc.ElectForChat(context.Background(), &v1.ChatReq{Model: "gpt"})
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			defer result.Close()
		})

		Convey("should fallback when requested model not found", func() {
			m := makeModel("gpt", "gpt-4", []conf.Capability{conf.Capability_CAPABILITY_CHAT})
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			req := &v1.ChatReq{Model: "nonexistent"}
			result, err := uc.ElectForChat(context.Background(), req)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			defer result.Close()
		})

		Convey("should skip models without chat capability", func() {
			m := makeModel("embed-model", "text-embedding-ada", []conf.Capability{conf.Capability_CAPABILITY_EMBEDDING})
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			_, err := uc.ElectForChat(context.Background(), &v1.ChatReq{Model: "embed-model"})
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("should skip models without chat repo", func() {
			m := makeModel("no-repo", "gpt-4", []conf.Capability{conf.Capability_CAPABILITY_CHAT})
			m.chatRepo = nil // No chat repo
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			_, err := uc.ElectForChat(context.Background(), &v1.ChatReq{Model: "no-repo"})
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("should update model ID to upstream ID", func() {
			m := makeModel("my-gpt", "gpt-4-turbo", []conf.Capability{conf.Capability_CAPABILITY_CHAT})
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			req := &v1.ChatReq{Model: "my-gpt"}
			result, err := uc.ElectForChat(context.Background(), req)
			So(err, ShouldBeNil)
			defer result.Close()
			So(req.Model, ShouldEqual, "gpt-4-turbo")
		})

		Convey("should keep model ID when no upstream ID", func() {
			m := makeModel("gpt-4", "", []conf.Capability{conf.Capability_CAPABILITY_CHAT})
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			req := &v1.ChatReq{Model: "gpt-4"}
			result, err := uc.ElectForChat(context.Background(), req)
			So(err, ShouldBeNil)
			defer result.Close()
			So(req.Model, ShouldEqual, "gpt-4")
		})

		Convey("should respect limiter during election", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			m := makeModel("gpt-4", "gpt-4", []conf.Capability{conf.Capability_CAPABILITY_CHAT})
			m.upstreamLimiters = &limiterGroup{
				requestLimiters: []repository.RequestLimiter{concurrency},
			}
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			// First election should succeed
			result1, err := uc.ElectForChat(context.Background(), &v1.ChatReq{Model: "gpt-4"})
			So(err, ShouldBeNil)

			// Second election should wait (concurrency exhausted but waitable)
			done := make(chan struct{})
			var model chat.Model
			go func() {
				model, err = uc.ElectForChat(context.Background(), &v1.ChatReq{Model: "gpt-4"})
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

		Convey("should timeout when concurrency exhausted with deadline", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			m := makeModel("gpt-4", "gpt-4", []conf.Capability{conf.Capability_CAPABILITY_CHAT})
			m.upstreamLimiters = &limiterGroup{
				requestLimiters: []repository.RequestLimiter{concurrency},
			}
			uc := &UseCaseImpl{
				models: []*model{m},
				log:    log.NewHelper(log.DefaultLogger),
			}

			// First election should succeed
			result1, err := uc.ElectForChat(context.Background(), &v1.ChatReq{Model: "gpt-4"})
			So(err, ShouldBeNil)

			// Second election should timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			_, err = uc.ElectForChat(ctx, &v1.ChatReq{Model: "gpt-4"})
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)

			// Release and try again
			result1.Close()
			result2, err := uc.ElectForChat(context.Background(), &v1.ChatReq{Model: "gpt-4"})
			So(err, ShouldBeNil)
			result2.Close()
		})
	})
}
