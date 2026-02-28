package model

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/data/limiter/local"
)

func TestElectFromCandidates(t *testing.T) {
	Convey("Test electFromCandidates", t, func() {
		Convey("with no candidates should return error", func() {
			_, _, err := electFromCandidates(context.Background(), nil, 0)
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("with empty candidate slice should return error", func() {
			_, _, err := electFromCandidates(context.Background(), []*model{}, 0)
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("with single available candidate should elect it", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(5),
					},
				},
				modelLimiters: &limiterGroup{},
			}
			selected, rs, err := electFromCandidates(context.Background(), []*model{m}, 0)
			So(err, ShouldBeNil)
			So(selected, ShouldEqual, m)
			So(rs, ShouldNotBeNil)
			rs.cancel()
		})

		Convey("with single candidate and no limiters should elect it", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters:    &limiterGroup{},
			}
			selected, rs, err := electFromCandidates(context.Background(), []*model{m}, 0)
			So(err, ShouldBeNil)
			So(selected, ShouldEqual, m)
			So(rs, ShouldNotBeNil)
		})

		Convey("should prefer available candidate over exhausted one", func() {
			// m1: concurrency exhausted
			m1Concurrency := local.NewConcurrencyLimiter(1)
			r, _ := m1Concurrency.Reserve()
			defer r.Cancel()

			m1 := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{m1Concurrency},
				},
			}

			// m2: available
			m2 := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(5),
					},
				},
			}

			// Run multiple times to verify m2 is always selected
			for range 10 {
				selected, rs, err := electFromCandidates(context.Background(), []*model{m1, m2}, 0)
				So(err, ShouldBeNil)
				So(selected, ShouldEqual, m2)
				rs.cancel()
			}
		})

		Convey("should wait and timeout when all candidates have concurrency exhausted", func() {
			m1Concurrency := local.NewConcurrencyLimiter(1)
			r1, _ := m1Concurrency.Reserve()
			defer r1.Cancel()

			m2Concurrency := local.NewConcurrencyLimiter(1)
			r2, _ := m2Concurrency.Reserve()
			defer r2.Cancel()

			m1 := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{m1Concurrency},
				},
			}
			m2 := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{m2Concurrency},
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			_, _, err := electFromCandidates(ctx, []*model{m1, m2}, 0)
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)
		})

		Convey("should wait and succeed when concurrency becomes available", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			res, _ := concurrency.Reserve()

			m := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{concurrency},
				},
			}

			var selected *model
			var rs *reservationSet
			var err error
			done := make(chan struct{})

			go func() {
				selected, rs, err = electFromCandidates(context.Background(), []*model{m}, 0)
				close(done)
			}()

			// Allow election to start waiting
			time.Sleep(10 * time.Millisecond)

			// Release the concurrency slot
			res.Cancel()

			<-done
			So(err, ShouldBeNil)
			So(selected, ShouldEqual, m)
			So(rs, ShouldNotBeNil)
			rs.cancel()
		})

		Convey("should distribute load among available candidates", func() {
			m1 := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(100),
					},
				},
			}
			m2 := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(100),
					},
				},
			}

			m1Count, m2Count := 0, 0
			for range 50 {
				selected, rs, err := electFromCandidates(context.Background(), []*model{m1, m2}, 0)
				So(err, ShouldBeNil)
				if selected == m1 {
					m1Count++
				} else {
					m2Count++
				}
				rs.cancel()
			}
			So(m1Count, ShouldBeGreaterThan, 10)
			So(m2Count, ShouldBeGreaterThan, 10)
		})

		Convey("should work with mixed request and token limiters", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(10),
					},
					tokenLimiters: []repository.TokenLimiter{
						local.NewTPMLimiter(100000),
					},
				},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewRPMLimiter(500),
					},
					tokenLimiters: []repository.TokenLimiter{
						local.NewDailyTokenLimiter(1000000),
					},
				},
			}
			selected, rs, err := electFromCandidates(context.Background(), []*model{m}, 1000)
			So(err, ShouldBeNil)
			So(selected, ShouldEqual, m)
			So(len(rs.requestReservations), ShouldEqual, 2) // concurrency + RPM
			So(len(rs.tokenReservations), ShouldEqual, 2)   // TPM + TPD
			rs.complete(500)
		})

		Convey("reservation lifecycle: complete releases concurrency", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			m := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{concurrency},
				},
			}

			_, rs, err := electFromCandidates(context.Background(), []*model{m}, 0)
			So(err, ShouldBeNil)
			So(concurrency.Probe(), ShouldBeGreaterThan, 0)

			rs.complete(0)
			So(concurrency.Probe(), ShouldEqual, 0)
		})

		Convey("reservation lifecycle: cancel releases concurrency", func() {
			concurrency := local.NewConcurrencyLimiter(1)
			m := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{concurrency},
				},
			}

			_, rs, err := electFromCandidates(context.Background(), []*model{m}, 0)
			So(err, ShouldBeNil)
			So(concurrency.Probe(), ShouldBeGreaterThan, 0)

			rs.cancel()
			So(concurrency.Probe(), ShouldEqual, 0)
		})

		Convey("with shared upstream limiters across multiple models", func() {
			// Simulate shared upstream limiter (concurrency=2)
			sharedConcurrency := local.NewConcurrencyLimiter(2)
			sharedGroup := &limiterGroup{
				requestLimiters: []repository.RequestLimiter{sharedConcurrency},
			}

			m1 := &model{
				upstreamLimiters: sharedGroup,
				modelLimiters:    &limiterGroup{},
			}
			m2 := &model{
				upstreamLimiters: sharedGroup,
				modelLimiters:    &limiterGroup{},
			}

			// First election
			_, rs1, err := electFromCandidates(context.Background(), []*model{m1, m2}, 0)
			So(err, ShouldBeNil)

			// Second election should also succeed (2 slots)
			_, rs2, err := electFromCandidates(context.Background(), []*model{m1, m2}, 0)
			So(err, ShouldBeNil)

			// Third election should wait and timeout (all slots taken)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			_, _, err = electFromCandidates(ctx, []*model{m1, m2}, 0)
			So(err, ShouldNotBeNil)
			So(v1.IsNoUpstream(err), ShouldBeTrue)

			rs1.cancel()
			rs2.cancel()
		})
	})
}
