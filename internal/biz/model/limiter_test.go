package model

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/data/limiter/local"
)

func TestNewLimiterGroup(t *testing.T) {
	Convey("Test newLimiterGroup", t, func() {
		Convey("all zeros should create empty group", func() {
			g := newLimiterGroup(0, 0, 0, 0, 0)
			So(g, ShouldNotBeNil)
			So(g.requestLimiters, ShouldBeEmpty)
			So(g.tokenLimiters, ShouldBeEmpty)
		})

		Convey("with concurrency only", func() {
			g := newLimiterGroup(10, 0, 0, 0, 0)
			So(len(g.requestLimiters), ShouldEqual, 1)
			So(g.tokenLimiters, ShouldBeEmpty)
		})

		Convey("with RPM only", func() {
			g := newLimiterGroup(0, 100, 0, 0, 0)
			So(len(g.requestLimiters), ShouldEqual, 1)
			So(g.tokenLimiters, ShouldBeEmpty)
		})

		Convey("with RPD only", func() {
			g := newLimiterGroup(0, 0, 1000, 0, 0)
			So(len(g.requestLimiters), ShouldEqual, 1)
			So(g.tokenLimiters, ShouldBeEmpty)
		})

		Convey("with TPM only", func() {
			g := newLimiterGroup(0, 0, 0, 50000, 0)
			So(g.requestLimiters, ShouldBeEmpty)
			So(len(g.tokenLimiters), ShouldEqual, 1)
		})

		Convey("with TPD only", func() {
			g := newLimiterGroup(0, 0, 0, 0, 500000)
			So(g.requestLimiters, ShouldBeEmpty)
			So(len(g.tokenLimiters), ShouldEqual, 1)
		})

		Convey("with all limits set", func() {
			g := newLimiterGroup(10, 100, 1000, 50000, 500000)
			So(len(g.requestLimiters), ShouldEqual, 3) // concurrency, RPM, RPD
			So(len(g.tokenLimiters), ShouldEqual, 2)   // TPM, TPD
		})

		Convey("with only token limits", func() {
			g := newLimiterGroup(0, 0, 0, 1000, 10000)
			So(g.requestLimiters, ShouldBeEmpty)
			So(len(g.tokenLimiters), ShouldEqual, 2)
		})

		Convey("with only request limits", func() {
			g := newLimiterGroup(5, 60, 500, 0, 0)
			So(len(g.requestLimiters), ShouldEqual, 3)
			So(g.tokenLimiters, ShouldBeEmpty)
		})
	})
}

func TestLimiterGroup_ProbeDelay(t *testing.T) {
	Convey("Test limiterGroup probeDelay", t, func() {
		Convey("nil group should return 0", func() {
			var g *limiterGroup
			So(g.probeDelay(100), ShouldEqual, 0)
		})

		Convey("empty group should return 0", func() {
			g := &limiterGroup{}
			So(g.probeDelay(100), ShouldEqual, 0)
		})

		Convey("with available request limiter should return 0", func() {
			g := &limiterGroup{
				requestLimiters: []repository.RequestLimiter{
					local.NewConcurrencyLimiter(10),
				},
			}
			So(g.probeDelay(0), ShouldEqual, 0)
		})

		Convey("with exhausted concurrency limiter should return estimated delay", func() {
			limiter := local.NewConcurrencyLimiter(1)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			defer res.Cancel()

			g := &limiterGroup{
				requestLimiters: []repository.RequestLimiter{limiter},
			}
			delay := g.probeDelay(0)
			So(delay, ShouldBeGreaterThan, 0)
			So(delay, ShouldBeLessThan, repository.InfDuration)
		})

		Convey("with token limiter and zero tokens should skip token probing", func() {
			limiter := local.NewTPMLimiter(100)
			res, err := limiter.Reserve(100)
			So(err, ShouldBeNil)
			defer res.Cancel()

			g := &limiterGroup{
				tokenLimiters: []repository.TokenLimiter{limiter},
			}
			So(g.probeDelay(0), ShouldEqual, 0)
		})

		Convey("with available token limiter should return 0", func() {
			g := &limiterGroup{
				tokenLimiters: []repository.TokenLimiter{
					local.NewTPMLimiter(1000),
				},
			}
			So(g.probeDelay(100), ShouldEqual, 0)
		})

		Convey("should return max delay across all limiters", func() {
			g := &limiterGroup{
				requestLimiters: []repository.RequestLimiter{
					local.NewConcurrencyLimiter(10),
					local.NewRPMLimiter(1000),
				},
			}
			So(g.probeDelay(0), ShouldEqual, 0)
		})

		Convey("should return estimated delay if any request limiter has exhausted concurrency", func() {
			available := local.NewConcurrencyLimiter(10)
			exhausted := local.NewConcurrencyLimiter(1)
			r, _ := exhausted.Reserve()
			defer r.Cancel()

			g := &limiterGroup{
				requestLimiters: []repository.RequestLimiter{available, exhausted},
			}
			delay := g.probeDelay(0)
			So(delay, ShouldBeGreaterThan, 0)
			So(delay, ShouldBeLessThan, repository.InfDuration)
		})

		Convey("with both request and token limiters available should return 0", func() {
			g := &limiterGroup{
				requestLimiters: []repository.RequestLimiter{
					local.NewConcurrencyLimiter(10),
				},
				tokenLimiters: []repository.TokenLimiter{
					local.NewTPMLimiter(10000),
				},
			}
			So(g.probeDelay(100), ShouldEqual, 0)
		})
	})
}

func TestReservationSet_MaxDelay(t *testing.T) {
	Convey("Test reservationSet maxDelay", t, func() {
		Convey("empty set should return 0", func() {
			rs := &reservationSet{}
			So(rs.maxDelay(), ShouldEqual, 0)
		})

		Convey("with immediately available reservation should return 0", func() {
			limiter := local.NewConcurrencyLimiter(1)
			r, err := limiter.Reserve()
			So(err, ShouldBeNil)

			rs := &reservationSet{
				requestReservations: []repository.Reservation{r},
			}
			So(rs.maxDelay(), ShouldEqual, 0)
			rs.cancel()
		})

		Convey("with delayed reservation should return positive delay", func() {
			limiter := local.NewConcurrencyLimiter(1)
			r1, _ := limiter.Reserve()

			// Second reservation will have estimated delay
			r2, _ := limiter.Reserve()
			rs := &reservationSet{
				requestReservations: []repository.Reservation{r2},
			}
			delay := rs.maxDelay()
			So(delay, ShouldBeGreaterThan, 0)
			So(delay, ShouldBeLessThan, repository.InfDuration)
			r1.Cancel()
			rs.cancel()
		})
	})
}

func TestReservationSet_WaitAll(t *testing.T) {
	Convey("Test reservationSet waitAll", t, func() {
		Convey("empty set should return nil", func() {
			rs := &reservationSet{}
			err := rs.wait(context.Background())
			So(err, ShouldBeNil)
		})

		Convey("with immediately available reservations should return nil", func() {
			limiter := local.NewConcurrencyLimiter(5)
			r, _ := limiter.Reserve()
			rs := &reservationSet{
				requestReservations: []repository.Reservation{r},
			}
			err := rs.wait(context.Background())
			So(err, ShouldBeNil)
			rs.cancel()
		})

		Convey("should cancel on context cancellation", func() {
			limiter := local.NewConcurrencyLimiter(1)
			r1, _ := limiter.Reserve()

			r2, _ := limiter.Reserve()
			rs := &reservationSet{
				requestReservations: []repository.Reservation{r2},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			err := rs.wait(ctx)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, context.DeadlineExceeded), ShouldBeTrue)
			r1.Cancel()
		})
	})
}

func TestReservationSet_CancelAll(t *testing.T) {
	Convey("Test reservationSet cancelAll", t, func() {
		Convey("should release all request reservations", func() {
			limiter := local.NewConcurrencyLimiter(2)
			r1, _ := limiter.Reserve()
			r2, _ := limiter.Reserve()

			rs := &reservationSet{
				requestReservations: []repository.Reservation{r1, r2},
			}

			So(limiter.Probe(), ShouldBeGreaterThan, 0)
			So(limiter.Probe(), ShouldBeLessThan, repository.InfDuration)
			rs.cancel()
			So(limiter.Probe(), ShouldEqual, 0)
			So(rs.requestReservations, ShouldBeNil)
		})

		Convey("should release token reservations", func() {
			limiter := local.NewTPMLimiter(100)
			r, _ := limiter.Reserve(50)

			rs := &reservationSet{
				tokenReservations: []repository.TokenReservation{r},
			}
			rs.cancel()
			So(rs.tokenReservations, ShouldBeNil)
			So(limiter.Probe(100), ShouldEqual, 0)
		})

		Convey("should be safe to call multiple times", func() {
			limiter := local.NewConcurrencyLimiter(1)
			r, _ := limiter.Reserve()
			rs := &reservationSet{
				requestReservations: []repository.Reservation{r},
			}
			rs.cancel()
			rs.cancel() // second call should be no-op
			So(rs.requestReservations, ShouldBeNil)
		})
	})
}

func TestReservationSet_Complete(t *testing.T) {
	Convey("Test reservationSet complete", t, func() {
		Convey("should finalize request reservations", func() {
			limiter := local.NewConcurrencyLimiter(2)
			r, _ := limiter.Reserve()

			rs := &reservationSet{
				requestReservations: []repository.Reservation{r},
			}
			rs.complete(0)
			So(limiter.Probe(), ShouldEqual, 0)
			So(rs.requestReservations, ShouldBeNil)
		})

		Convey("should finalize token reservations with actual tokens", func() {
			limiter := local.NewTPMLimiter(1000)
			r, _ := limiter.Reserve(500)

			rs := &reservationSet{
				tokenReservations: []repository.TokenReservation{r},
			}
			rs.complete(300) // Actual usage is less than estimated
			So(rs.tokenReservations, ShouldBeNil)
			So(limiter.Probe(700), ShouldEqual, 0)
		})

		Convey("should finalize both request and token reservations", func() {
			reqLimiter := local.NewConcurrencyLimiter(1)
			rReq, _ := reqLimiter.Reserve()

			tokLimiter := local.NewTPMLimiter(10000)
			rTok, _ := tokLimiter.Reserve(500)

			rs := &reservationSet{
				requestReservations: []repository.Reservation{rReq},
				tokenReservations:   []repository.TokenReservation{rTok},
			}
			rs.complete(400)
			So(rs.requestReservations, ShouldBeNil)
			So(rs.tokenReservations, ShouldBeNil)
			So(reqLimiter.Probe(), ShouldEqual, 0)
			So(tokLimiter.Probe(9600), ShouldEqual, 0)
		})
	})
}

func TestProbeModelDelay(t *testing.T) {
	Convey("Test probeModelDelay", t, func() {
		Convey("with no limiters should return 0", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters:    &limiterGroup{},
			}
			So(probeModelDelay(m, 100), ShouldEqual, 0)
		})

		Convey("with available upstream limiter should return 0", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(10),
					},
				},
				modelLimiters: &limiterGroup{},
			}
			So(probeModelDelay(m, 0), ShouldEqual, 0)
		})

		Convey("with available model limiter should return 0", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(10),
					},
				},
			}
			So(probeModelDelay(m, 0), ShouldEqual, 0)
		})

		Convey("should return max of upstream and model delays", func() {
			modelConcurrency := local.NewConcurrencyLimiter(1)
			r, _ := modelConcurrency.Reserve()
			defer r.Cancel()

			m := &model{
				upstreamLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(10),
					},
				},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{modelConcurrency},
				},
			}
			delay := probeModelDelay(m, 0)
			So(delay, ShouldBeGreaterThan, 0)
			So(delay, ShouldBeLessThan, repository.InfDuration)
		})

		Convey("with upstream exhausted and model available", func() {
			upstreamConcurrency := local.NewConcurrencyLimiter(1)
			r, _ := upstreamConcurrency.Reserve()
			defer r.Cancel()

			m := &model{
				upstreamLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{upstreamConcurrency},
				},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(10),
					},
				},
			}
			delay := probeModelDelay(m, 0)
			So(delay, ShouldBeGreaterThan, 0)
			So(delay, ShouldBeLessThan, repository.InfDuration)
		})

		Convey("with token estimation", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{
					tokenLimiters: []repository.TokenLimiter{
						local.NewTPMLimiter(10000),
					},
				},
				modelLimiters: &limiterGroup{},
			}
			So(probeModelDelay(m, 100), ShouldEqual, 0)
		})
	})
}

func TestTryReserveAll(t *testing.T) {
	Convey("Test tryReserveAll", t, func() {
		Convey("with no limiters should succeed", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{},
				modelLimiters:    &limiterGroup{},
			}
			rs, err := tryReserveAll(m, 0)
			So(err, ShouldBeNil)
			So(rs, ShouldNotBeNil)
			So(rs.requestReservations, ShouldBeEmpty)
			So(rs.tokenReservations, ShouldBeEmpty)
		})

		Convey("with available concurrency limiter should succeed", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(5),
					},
				},
				modelLimiters: &limiterGroup{},
			}
			rs, err := tryReserveAll(m, 0)
			So(err, ShouldBeNil)
			So(len(rs.requestReservations), ShouldEqual, 1)
			So(rs.requestReservations[0].Delay(), ShouldEqual, 0)
			rs.cancel()
		})

		Convey("with both upstream and model limiters should reserve all", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(5),
						local.NewRPMLimiter(100),
					},
				},
				modelLimiters: &limiterGroup{
					requestLimiters: []repository.RequestLimiter{
						local.NewConcurrencyLimiter(3),
					},
				},
			}
			rs, err := tryReserveAll(m, 0)
			So(err, ShouldBeNil)
			So(len(rs.requestReservations), ShouldEqual, 3)
			rs.cancel()
		})

		Convey("with token limiters and estimated tokens should reserve tokens", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{
					tokenLimiters: []repository.TokenLimiter{
						local.NewTPMLimiter(10000),
					},
				},
				modelLimiters: &limiterGroup{},
			}
			rs, err := tryReserveAll(m, 100)
			So(err, ShouldBeNil)
			So(len(rs.tokenReservations), ShouldEqual, 1)
			rs.cancel()
		})

		Convey("with zero estimated tokens should skip token limiters", func() {
			limiter := local.NewTPMLimiter(1)
			res, err := limiter.Reserve(1)
			So(err, ShouldBeNil)
			defer res.Cancel()

			m := &model{
				upstreamLimiters: &limiterGroup{
					tokenLimiters: []repository.TokenLimiter{limiter},
				},
				modelLimiters: &limiterGroup{},
			}
			rs, err := tryReserveAll(m, 0)
			So(err, ShouldBeNil)
			So(rs.tokenReservations, ShouldBeEmpty)
		})

		Convey("with nil limiter groups should succeed", func() {
			m := &model{
				upstreamLimiters: nil,
				modelLimiters:    nil,
			}
			rs, err := tryReserveAll(m, 100)
			So(err, ShouldBeNil)
			So(rs, ShouldNotBeNil)
		})

		Convey("with mixed upstream and model token limiters", func() {
			m := &model{
				upstreamLimiters: &limiterGroup{
					tokenLimiters: []repository.TokenLimiter{
						local.NewTPMLimiter(50000),
					},
				},
				modelLimiters: &limiterGroup{
					tokenLimiters: []repository.TokenLimiter{
						local.NewDailyTokenLimiter(1000000),
					},
				},
			}
			rs, err := tryReserveAll(m, 1000)
			So(err, ShouldBeNil)
			So(len(rs.tokenReservations), ShouldEqual, 2)
			rs.cancel()
		})
	})
}
