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

func TestNewRPMLimiter(t *testing.T) {
	Convey("Test NewRPMLimiter", t, func() {
		Convey("should create limiter with positive limit", func() {
			limiter := NewRPMLimiter(60)
			So(limiter, ShouldNotBeNil)
			So(limiter, ShouldImplement, (*repository.RequestLimiter)(nil))
		})

		Convey("should create limiter with zero limit", func() {
			limiter := NewRPMLimiter(0)
			So(limiter, ShouldNotBeNil)
		})
	})
}

func TestRPMLimiter_Probe(t *testing.T) {
	Convey("Test RPMLimiter Probe", t, func() {
		Convey("should return 0 when tokens are available", func() {
			limiter := NewRPMLimiter(60)
			delay := limiter.Probe()
			So(delay, ShouldEqual, 0)
		})

		Convey("should return positive delay when tokens exhausted", func() {
			limiter := NewRPMLimiter(60)

			// Reserve all burst capacity
			for i := 0; i < 60; i++ {
				res, err := limiter.Reserve()
				So(err, ShouldBeNil)
				So(res.Delay(), ShouldEqual, 0)
			}

			// Probe should indicate waiting time
			delay := limiter.Probe()
			So(delay, ShouldBeGreaterThan, 0)
		})

		Convey("should return approximately correct wait time", func() {
			limiter := NewRPMLimiter(60) // 1 token per second

			// Exhaust tokens
			for i := 0; i < 60; i++ {
				limiter.Reserve()
			}

			// Probe for next request
			delay := limiter.Probe()
			So(delay, ShouldBeGreaterThanOrEqualTo, 980*time.Millisecond)
			So(delay, ShouldBeLessThanOrEqualTo, 1020*time.Millisecond)
		})
	})
}

func TestRPMLimiter_Reserve(t *testing.T) {
	Convey("Test RPMLimiter Reserve", t, func() {
		Convey("should succeed immediately when tokens available", func() {
			limiter := NewRPMLimiter(60)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.Delay(), ShouldEqual, 0)
		})

		Convey("should allow burst up to limit", func() {
			limiter := NewRPMLimiter(60)

			// Reserve entire burst capacity
			for i := 0; i < 60; i++ {
				res, err := limiter.Reserve()
				So(err, ShouldBeNil)
				So(res.Delay(), ShouldEqual, 0)
			}

			// Next reservation should have delay
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldBeGreaterThan, 0)
		})

		Convey("should generate tokens over time", func() {
			limiter := NewRPMLimiter(60) // 1 token/second

			// Exhaust initial burst
			for i := 0; i < 60; i++ {
				limiter.Reserve()
			}

			// Check initial delay for next request
			res1, _ := limiter.Reserve()
			initialDelay := res1.Delay()
			So(initialDelay, ShouldBeGreaterThan, 0)

			// Wait for some tokens to regenerate
			time.Sleep(1 * time.Millisecond)

			// Delay for the same reservation should decrease
			laterDelay := res1.Delay()
			So(laterDelay, ShouldBeLessThan, initialDelay)
		})
	})
}

func TestRequestReservation_Delay(t *testing.T) {
	Convey("Test requestReservation Delay", t, func() {
		Convey("should return 0 for immediate reservation", func() {
			limiter := NewRPMLimiter(60)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldEqual, 0)
		})

		Convey("should return positive delay when quota exceeded", func() {
			limiter := NewRPMLimiter(60)

			// Exhaust burst
			for i := 0; i < 60; i++ {
				limiter.Reserve()
			}

			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldBeGreaterThan, 0)
		})

		Convey("should decrease over time", func() {
			limiter := NewRPMLimiter(60)

			// Exhaust burst
			for i := 0; i < 60; i++ {
				limiter.Reserve()
			}

			res, _ := limiter.Reserve()
			initialDelay := res.Delay()

			time.Sleep(1 * time.Microsecond)
			laterDelay := res.Delay()

			So(laterDelay, ShouldBeLessThan, initialDelay)
		})
	})
}

func TestRequestReservation_Wait(t *testing.T) {
	Convey("Test requestReservation Wait", t, func() {
		Convey("should return immediately when no delay", func() {
			limiter := NewRPMLimiter(60)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			start := time.Now()
			err = res.Wait(context.Background())
			elapsed := time.Since(start)

			So(err, ShouldBeNil)
			So(elapsed, ShouldBeLessThan, 1*time.Millisecond)
		})

		Convey("should block for appropriate duration", func() {
			limiter := NewRPMLimiter(1200) // 10 tokens/second

			// Exhaust burst
			for i := 0; i < 1200; i++ {
				limiter.Reserve()
			}

			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			start := time.Now()
			err = res.Wait(context.Background())
			elapsed := time.Since(start)

			So(err, ShouldBeNil)
			So(elapsed, ShouldBeGreaterThanOrEqualTo, 45*time.Millisecond)
			So(elapsed, ShouldBeLessThan, 55*time.Millisecond)
		})

		Convey("should respect context cancellation", func() {
			limiter := NewRPMLimiter(60)

			// Exhaust burst
			for i := 0; i < 60; i++ {
				limiter.Reserve()
			}

			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			err = res.Wait(ctx)
			So(err, ShouldEqual, context.Canceled)
		})

		Convey("should respect context timeout", func() {
			limiter := NewRPMLimiter(60)

			// Exhaust burst
			for i := 0; i < 60; i++ {
				limiter.Reserve()
			}

			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			start := time.Now()
			err = res.Wait(ctx)
			elapsed := time.Since(start)

			So(err, ShouldEqual, context.DeadlineExceeded)
			So(elapsed, ShouldBeGreaterThanOrEqualTo, 8*time.Millisecond)
			So(elapsed, ShouldBeLessThan, 12*time.Millisecond)
		})
	})
}

func TestRequestReservation_Cancel(t *testing.T) {
	Convey("Test requestReservation Cancel", t, func() {
		Convey("should refund token", func() {
			limiter := NewRPMLimiter(1)

			// Reserve and cancel
			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)
			res1.Cancel()

			// Should still have full burst available
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)
		})
	})
}

func TestRequestReservation_Complete(t *testing.T) {
	Convey("Test requestReservation Complete", t, func() {
		Convey("should consume token", func() {
			limiter := NewRPMLimiter(2)

			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)
			res1.Complete()

			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			res2.Complete()

			// Should have consumed 2 tokens
			probe := limiter.Probe()
			So(probe, ShouldBeGreaterThan, 0)
		})

		Convey("should be idempotent", func() {
			limiter := NewRPMLimiter(2)

			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			// Complete multiple times
			res.Complete()
			res.Complete()
			res.Complete()

			// Should have consumed only once
			probe := limiter.Probe()
			So(probe, ShouldEqual, 0)
		})
	})
}

func TestNewTPMLimiter(t *testing.T) {
	Convey("Test NewTPMLimiter", t, func() {
		Convey("should create limiter with positive limit", func() {
			limiter := NewTPMLimiter(10000)
			So(limiter, ShouldNotBeNil)
			So(limiter, ShouldImplement, (*repository.TokenLimiter)(nil))
		})

		Convey("should create limiter with zero limit", func() {
			limiter := NewTPMLimiter(0)
			So(limiter, ShouldNotBeNil)
		})
	})
}

func TestTPMLimiter_Probe(t *testing.T) {
	Convey("Test TPMLimiter Probe", t, func() {
		Convey("should return 0 when tokens available", func() {
			limiter := NewTPMLimiter(1000)
			delay := limiter.Probe(100)
			So(delay, ShouldEqual, 0)
		})

		Convey("should return InfDuration when requested tokens exceed burst", func() {
			limiter := NewTPMLimiter(1000)
			delay := limiter.Probe(1001)
			So(delay, ShouldEqual, repository.InfDuration)
		})

		Convey("should return positive delay when tokens exhausted", func() {
			limiter := NewTPMLimiter(6000)

			// Exhaust tokens
			res, err := limiter.Reserve(6000)
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldEqual, 0)

			// Probe should indicate waiting time
			delay := limiter.Probe(100)
			So(delay, ShouldBeGreaterThan, 0)
			So(delay, ShouldBeLessThan, 1001*time.Millisecond)
		})

		Convey("should return approximately correct wait time", func() {
			limiter := NewTPMLimiter(6000) // 100 tokens/second

			// Exhaust tokens
			limiter.Reserve(6000)

			// Probe for 200 tokens should wait ~2 seconds
			delay := limiter.Probe(200)
			So(delay, ShouldBeGreaterThanOrEqualTo, 1950*time.Millisecond)
			So(delay, ShouldBeLessThanOrEqualTo, 2050*time.Millisecond)
		})
	})
}

func TestTPMLimiter_Reserve(t *testing.T) {
	Convey("Test TPMLimiter Reserve", t, func() {
		Convey("should succeed immediately when tokens available", func() {
			limiter := NewTPMLimiter(1000)
			res, err := limiter.Reserve(500)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.Delay(), ShouldEqual, 0)
		})

		Convey("should return error when requested tokens exceed burst", func() {
			limiter := NewTPMLimiter(1000)
			res, err := limiter.Reserve(1001)
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("should allow burst up to limit", func() {
			limiter := NewTPMLimiter(1000)

			// Reserve entire burst capacity
			res, err := limiter.Reserve(1000)
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldEqual, 0)

			// Next reservation should have delay
			res2, err := limiter.Reserve(1)
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldBeGreaterThan, 0)
		})

		Convey("should generate tokens over time", func() {
			limiter := NewTPMLimiter(6000) // 100 tokens/second

			// Exhaust initial burst
			limiter.Reserve(6000)

			// Check initial delay for next request
			res1, _ := limiter.Reserve(100)
			initialDelay := res1.Delay()
			So(initialDelay, ShouldBeGreaterThan, 0)

			// Wait for some tokens to regenerate
			time.Sleep(1 * time.Microsecond)

			// Delay for the same reservation should decrease
			laterDelay := res1.Delay()
			So(laterDelay, ShouldBeLessThan, initialDelay)
		})

		Convey("should handle multiple partial reservations", func() {
			limiter := NewTPMLimiter(1000)

			res1, err := limiter.Reserve(300)
			So(err, ShouldBeNil)
			So(res1.Delay(), ShouldEqual, 0)

			res2, err := limiter.Reserve(400)
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)

			res3, err := limiter.Reserve(300)
			So(err, ShouldBeNil)
			So(res3.Delay(), ShouldEqual, 0)

			// Next reservation should have delay
			res4, err := limiter.Reserve(1)
			So(err, ShouldBeNil)
			So(res4.Delay(), ShouldBeGreaterThan, 0)
		})
	})
}

func TestTokenReservation_Delay(t *testing.T) {
	Convey("Test tokenReservation Delay", t, func() {
		Convey("should return 0 for immediate reservation", func() {
			limiter := NewTPMLimiter(1000)
			res, err := limiter.Reserve(500)
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldEqual, 0)
		})

		Convey("should return positive delay when quota exceeded", func() {
			limiter := NewTPMLimiter(1000)

			// Exhaust burst
			limiter.Reserve(1000)

			res, err := limiter.Reserve(100)
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldBeGreaterThan, 0)
		})

		Convey("should decrease over time", func() {
			limiter := NewTPMLimiter(1000)

			// Exhaust burst
			limiter.Reserve(1000)

			res, _ := limiter.Reserve(100)
			initialDelay := res.Delay()

			time.Sleep(1 * time.Microsecond)
			laterDelay := res.Delay()

			So(laterDelay, ShouldBeLessThan, initialDelay)
		})
	})
}

func TestTokenReservation_Wait(t *testing.T) {
	Convey("Test tokenReservation Wait", t, func() {
		Convey("should return immediately when no delay", func() {
			limiter := NewTPMLimiter(1000)
			res, err := limiter.Reserve(500)
			So(err, ShouldBeNil)

			start := time.Now()
			err = res.Wait(context.Background())
			elapsed := time.Since(start)

			So(err, ShouldBeNil)
			So(elapsed, ShouldBeLessThan, 1*time.Millisecond)
		})

		Convey("should block for appropriate duration", func() {
			limiter := NewTPMLimiter(6000) // 100 tokens/second

			// Exhaust burst
			limiter.Reserve(6000)

			res, err := limiter.Reserve(200)
			So(err, ShouldBeNil)

			start := time.Now()
			err = res.Wait(context.Background())
			elapsed := time.Since(start)

			So(err, ShouldBeNil)
			So(elapsed, ShouldBeGreaterThanOrEqualTo, 1950*time.Millisecond)
			So(elapsed, ShouldBeLessThan, 2050*time.Millisecond)
		})

		Convey("should respect context cancellation", func() {
			limiter := NewTPMLimiter(1000)

			// Exhaust burst
			limiter.Reserve(1000)

			res, err := limiter.Reserve(100)
			So(err, ShouldBeNil)

			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			err = res.Wait(ctx)
			So(err, ShouldEqual, context.Canceled)
		})

		Convey("should respect context timeout", func() {
			limiter := NewTPMLimiter(1000)

			// Exhaust burst
			limiter.Reserve(1000)

			res, err := limiter.Reserve(100)
			So(err, ShouldBeNil)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			start := time.Now()
			err = res.Wait(ctx)
			elapsed := time.Since(start)

			So(err, ShouldEqual, context.DeadlineExceeded)
			So(elapsed, ShouldBeGreaterThanOrEqualTo, 5*time.Millisecond)
			So(elapsed, ShouldBeLessThan, 10*time.Millisecond)
		})
	})
}

func TestTokenReservation_Cancel(t *testing.T) {
	Convey("Test tokenReservation Cancel", t, func() {
		Convey("should refund all tokens", func() {
			limiter := NewTPMLimiter(1000)

			// Reserve and cancel
			res, err := limiter.Reserve(500)
			So(err, ShouldBeNil)
			res.Cancel()

			// Should still have full burst available
			res2, err := limiter.Reserve(1000)
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)
		})
	})
}

func TestTokenReservation_Complete(t *testing.T) {
	Convey("Test tokenReservation Complete", t, func() {
		Convey("should consume tokens", func() {
			limiter := NewTPMLimiter(1000)

			res, err := limiter.Reserve(500)
			So(err, ShouldBeNil)
			res.Complete()

			// Should have consumed 500 tokens
			res2, err := limiter.Reserve(500)
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)

			res3, err := limiter.Reserve(1)
			So(err, ShouldBeNil)
			So(res3.Delay(), ShouldBeGreaterThan, 0)
		})

		Convey("should be idempotent", func() {
			limiter := NewTPMLimiter(1000)

			res, err := limiter.Reserve(500)
			So(err, ShouldBeNil)

			// Complete multiple times
			res.Complete()
			res.Complete()
			res.Complete()

			// Should have consumed only 500
			probe := limiter.Probe(500)
			So(probe, ShouldEqual, 0)

			probe = limiter.Probe(501)
			So(probe, ShouldBeGreaterThan, 0)
		})
	})
}

func TestTokenReservation_CompleteWithActual(t *testing.T) {
	Convey("Test tokenReservation CompleteWithActual", t, func() {
		Convey("should refund excess when actual < estimated", func() {
			limiter := NewTPMLimiter(1000)

			// Reserve 500 but actually use 300
			res, err := limiter.Reserve(500)
			So(err, ShouldBeNil)
			res.CompleteWithActual(300)

			// Should have 700 tokens remaining (1000 - 300)
			probe := limiter.Probe(700)
			So(probe, ShouldEqual, 0)

			probe = limiter.Probe(701)
			So(probe, ShouldBeGreaterThan, 0)
		})

		Convey("should deduct more when actual > estimated", func() {
			limiter := NewTPMLimiter(1000)

			// Reserve 300 but actually use 500
			res, err := limiter.Reserve(300)
			So(err, ShouldBeNil)
			res.CompleteWithActual(500)

			// Should have 500 tokens remaining (1000 - 500)
			probe := limiter.Probe(500)
			So(probe, ShouldEqual, 0)

			probe = limiter.Probe(501)
			So(probe, ShouldBeGreaterThan, 0)
		})

		Convey("should handle exact match", func() {
			limiter := NewTPMLimiter(1000)

			// Reserve and use exactly the same
			res, err := limiter.Reserve(400)
			So(err, ShouldBeNil)
			res.CompleteWithActual(400)

			// Should have 600 tokens remaining
			probe := limiter.Probe(600)
			So(probe, ShouldEqual, 0)
		})

		Convey("should handle zero actual usage", func() {
			limiter := NewTPMLimiter(1000)

			// Reserve but don't use
			res, err := limiter.Reserve(500)
			So(err, ShouldBeNil)
			res.CompleteWithActual(0)

			// Should refund all 500 tokens, giving us full capacity
			res2, err := limiter.Reserve(1000)
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)
		})

		Convey("should cap tokens at burst limit", func() {
			limiter := NewTPMLimiter(1000)

			// Reserve and refund more than burst
			res1, _ := limiter.Reserve(500)
			res1.CompleteWithActual(0) // Refund 500

			res2, _ := limiter.Reserve(600)
			res2.CompleteWithActual(0) // Try to refund another 600

			// Should be capped at burst (1000), not 1100
			// We can reserve the full burst
			res3, err := limiter.Reserve(1000)
			So(err, ShouldBeNil)
			So(res3.Delay(), ShouldEqual, 0)

			// But not more than burst
			res4, err := limiter.Reserve(1)
			So(err, ShouldBeNil)
			So(res4.Delay(), ShouldBeGreaterThan, 0)
		})
	})
}

func TestTokenBucket_InterfaceCompliance(t *testing.T) {
	Convey("Test interface compliance", t, func() {
		Convey("rpmLimiter should implement RequestLimiter", func() {
			limiter := NewRPMLimiter(60)
			So(limiter, ShouldImplement, (*repository.RequestLimiter)(nil))
		})

		Convey("requestReservation should implement Reservation", func() {
			limiter := NewRPMLimiter(60)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res, ShouldImplement, (*repository.Reservation)(nil))
		})

		Convey("tpmLimiter should implement TokenLimiter", func() {
			limiter := NewTPMLimiter(1000)
			So(limiter, ShouldImplement, (*repository.TokenLimiter)(nil))
		})

		Convey("tokenReservation should implement TokenReservation", func() {
			limiter := NewTPMLimiter(1000)
			res, err := limiter.Reserve(100)
			So(err, ShouldBeNil)
			So(res, ShouldImplement, (*repository.TokenReservation)(nil))
		})
	})
}
