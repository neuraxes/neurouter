// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package local

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/neuraxes/neurouter/internal/biz/repository"
)

func TestNewDailyRequestLimiter(t *testing.T) {
	Convey("DailyRequestLimiter constructor", t, func() {
		Convey("zero limit returns nil", func() {
			l := NewDailyRequestLimiter(0)
			So(l, ShouldBeNil)
		})

		Convey("negative limit returns nil", func() {
			l := NewDailyRequestLimiter(-5)
			So(l, ShouldBeNil)
		})

		Convey("positive limit returns limiter", func() {
			l := NewDailyRequestLimiter(10)
			So(l, ShouldNotBeNil)
			So(l, ShouldImplement, (*repository.RequestLimiter)(nil))
		})
	})
}

func TestDailyRequestLimiter_ProbeAndReserve(t *testing.T) {
	Convey("DailyRequestLimiter probe and reserve behavior", t, func() {
		Convey("probe returns 0 when under limit", func() {
			l := NewDailyRequestLimiterWithTimeZone(2, time.UTC).(*DailyRequestLimiter)
			So(l.Probe(), ShouldEqual, 0)
		})

		Convey("probe waits until reset when exhausted", func() {
			l := NewDailyRequestLimiterWithTimeZone(1, time.UTC).(*DailyRequestLimiter)
			r, err := l.Reserve()
			So(err, ShouldBeNil)
			So(r.Delay(), ShouldEqual, 0)

			// Exhausted now
			delay := l.Probe()
			So(delay, ShouldBeGreaterThan, time.Duration(0))
			r.Cancel()
		})
	})
}

func TestDailyRequestReservation_WaitReset(t *testing.T) {
	Convey("DailyRequestReservation waits until midnight reset", t, func() {
		l := NewDailyRequestLimiterWithTimeZone(1, time.UTC).(*DailyRequestLimiter)

		// Take the only slot
		r1, err := l.Reserve()
		So(err, ShouldBeNil)
		So(r1.Delay(), ShouldEqual, 0)

		// Next reservation should wait until reset
		r2, err := l.Reserve()
		So(err, ShouldBeNil)
		So(r2.Delay(), ShouldBeGreaterThan, time.Duration(0))

		// Force a near-future reset to avoid long waits
		l.state.mu.Lock()
		l.state.resetTime = time.Now().Add(5 * time.Millisecond)
		l.state.mu.Unlock()

		ctx := context.Background()
		start := time.Now()
		err = r2.Wait(ctx)
		elapsed := time.Since(start)
		So(err, ShouldBeNil)
		So(elapsed, ShouldBeLessThanOrEqualTo, 7*time.Millisecond)
		So(elapsed, ShouldBeGreaterThanOrEqualTo, 3*time.Millisecond)

		r1.Cancel()
		r2.Cancel()
	})
}

func TestDailyRequestReservation_CancelComplete(t *testing.T) {
	Convey("DailyRequestReservation cancel/complete semantics", t, func() {
		Convey("cancel releases only if acquired", func() {
			l := NewDailyRequestLimiterWithTimeZone(1, time.UTC).(*DailyRequestLimiter)
			r1, _ := l.Reserve()
			So(r1.Delay(), ShouldEqual, 0)

			r2, _ := l.Reserve() // not acquired
			So(r2.Delay(), ShouldBeGreaterThan, time.Duration(0))

			r2.Cancel() // should not change used

			// Still exhausted
			r3, _ := l.Reserve()
			So(r3.Delay(), ShouldBeGreaterThan, time.Duration(0))

			r1.Cancel()
			r3.Cancel()
		})

		Convey("complete is a no-op for daily quotas", func() {
			l := NewDailyRequestLimiterWithTimeZone(1, time.UTC).(*DailyRequestLimiter)
			r, _ := l.Reserve()
			So(r.Delay(), ShouldEqual, 0)
			r.Complete()
			r.Complete()

			// Daily quotas are consumed; next reservation should be delayed until reset
			r2, _ := l.Reserve()
			So(r2.Delay(), ShouldBeGreaterThan, time.Duration(0))
			r2.Cancel()
		})
	})
}

func TestNewDailyTokenLimiter(t *testing.T) {
	Convey("DailyTokenLimiter constructor", t, func() {
		Convey("zero/negative limit returns nil", func() {
			So(NewDailyTokenLimiter(0), ShouldBeNil)
			So(NewDailyTokenLimiter(-1), ShouldBeNil)
		})

		Convey("positive limit returns limiter", func() {
			l := NewDailyTokenLimiter(100)
			So(l, ShouldNotBeNil)
			So(l, ShouldImplement, (*repository.TokenLimiter)(nil))
		})
	})
}

func TestDailyTokenLimiter_ProbeReserve(t *testing.T) {
	Convey("DailyTokenLimiter probe and reserve", t, func() {
		Convey("probe returns Inf when request exceeds limit", func() {
			l := NewDailyTokenLimiterWithTimeZone(100, time.UTC).(*DailyTokenLimiter)
			So(l.Probe(101), ShouldEqual, repository.InfDuration)
		})

		Convey("reserve returns error when request exceeds limit", func() {
			l := NewDailyTokenLimiterWithTimeZone(100, time.UTC).(*DailyTokenLimiter)
			res, err := l.Reserve(150)
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("reserve succeeds within limit and consumes tokens", func() {
			l := NewDailyTokenLimiterWithTimeZone(100, time.UTC).(*DailyTokenLimiter)
			res, err := l.Reserve(60)
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldEqual, 0)
			res.Cancel()
		})

		Convey("probe/reserve delay when exhausted until reset", func() {
			l := NewDailyTokenLimiterWithTimeZone(50, time.UTC).(*DailyTokenLimiter)
			r1, err := l.Reserve(50)
			So(err, ShouldBeNil)
			So(r1.Delay(), ShouldEqual, 0)

			// Now exhausted
			So(l.Probe(1), ShouldBeGreaterThan, time.Duration(0))
			r2, err := l.Reserve(1)
			So(err, ShouldBeNil)
			So(r2.Delay(), ShouldBeGreaterThan, time.Duration(0))

			r1.Cancel()
			r2.Cancel()
		})
	})
}

func TestDailyTokenReservation_WaitAndAdjust(t *testing.T) {
	Convey("DailyTokenReservation wait and CompleteWithActual", t, func() {
		Convey("wait blocks until reset then acquires", func() {
			l := NewDailyTokenLimiterWithTimeZone(10, time.UTC).(*DailyTokenLimiter)
			r1, _ := l.Reserve(10)
			So(r1.Delay(), ShouldEqual, 0)

			r2, _ := l.Reserve(5)
			So(r2.Delay(), ShouldBeGreaterThan, time.Duration(0))

			// Force a near reset
			l.state.mu.Lock()
			l.state.resetTime = time.Now().Add(5 * time.Millisecond)
			l.state.mu.Unlock()

			start := time.Now()
			err := r2.Wait(context.Background())
			elapsed := time.Since(start)

			So(err, ShouldBeNil)
			So(elapsed, ShouldBeGreaterThanOrEqualTo, 3*time.Millisecond)
			So(elapsed, ShouldBeLessThanOrEqualTo, 7*time.Millisecond)

			r1.Cancel()
			r2.Cancel()
		})

		Convey("CompleteWithActual adjusts consumed tokens", func() {
			l := NewDailyTokenLimiterWithTimeZone(200, time.UTC).(*DailyTokenLimiter)

			// Reserve 100 tokens
			r, err := l.Reserve(100)
			So(err, ShouldBeNil)
			So(r.Delay(), ShouldEqual, 0)

			// Simulate actual usage of 60 tokens
			tr := r.(*dailyTokenReservation)
			// Before complete, used should be 100
			So(l.state.used, ShouldEqual, int64(100))
			tr.CompleteWithActual(60)

			// After adjustment, used should be max(0, used - reserved + actual) = 60
			So(l.state.used, ShouldEqual, int64(60))

			// Further reservations within remaining quota should succeed
			r2, err := l.Reserve(140)
			So(err, ShouldBeNil)
			So(r2.Delay(), ShouldEqual, 0)
			r2.Cancel()
		})

		Convey("CompleteWithActual is idempotent when already released", func() {
			l := NewDailyTokenLimiterWithTimeZone(50, time.UTC).(*DailyTokenLimiter)
			r, _ := l.Reserve(30)
			tr := r.(*dailyTokenReservation)
			tr.CompleteWithActual(20)
			// Call again should have no effect
			tr.CompleteWithActual(10)

			// Remaining quota should allow further reservations
			r2, err := l.Reserve(30)
			So(err, ShouldBeNil)
			So(r2.Delay(), ShouldEqual, 0)
			r2.Cancel()
		})
	})
}

func TestDailyQuota_InterfaceCompliance(t *testing.T) {
	Convey("Daily quota types implement interfaces", t, func() {
		Convey("DailyRequestLimiter implements RequestLimiter", func() {
			l := NewDailyRequestLimiter(1)
			So(l, ShouldImplement, (*repository.RequestLimiter)(nil))
		})

		Convey("dailyRequestReservation implements Reservation", func() {
			l := NewDailyRequestLimiter(1).(*DailyRequestLimiter)
			r, _ := l.Reserve()
			So(r, ShouldImplement, (*repository.Reservation)(nil))
			r.Cancel()
		})

		Convey("DailyTokenLimiter implements TokenLimiter", func() {
			l := NewDailyTokenLimiter(1)
			So(l, ShouldImplement, (*repository.TokenLimiter)(nil))
		})

		Convey("dailyTokenReservation implements TokenReservation", func() {
			l := NewDailyTokenLimiter(2).(*DailyTokenLimiter)
			r, _ := l.Reserve(1)
			So(r, ShouldImplement, (*repository.TokenReservation)(nil))
			r.Cancel()
		})
	})
}
