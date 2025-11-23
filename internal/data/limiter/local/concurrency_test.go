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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/neuraxes/neurouter/internal/biz/repository"
)

func TestNewConcurrencyLimiter(t *testing.T) {
	Convey("Test NewConcurrencyLimiter", t, func() {
		Convey("with zero limit should return nil", func() {
			limiter := NewConcurrencyLimiter(0)
			So(limiter, ShouldBeNil)
		})

		Convey("with negative limit should return nil", func() {
			limiter := NewConcurrencyLimiter(-1)
			So(limiter, ShouldBeNil)
		})

		Convey("with positive limit should return limiter", func() {
			limiter := NewConcurrencyLimiter(5)
			So(limiter, ShouldNotBeNil)
			So(limiter, ShouldImplement, (*repository.RequestLimiter)(nil))
		})
	})
}

func TestConcurrencyLimiter_Probe(t *testing.T) {
	Convey("Test ConcurrencyLimiter Probe", t, func() {
		Convey("when quota is available", func() {
			limiter := NewConcurrencyLimiter(2)
			delay := limiter.Probe()
			So(delay, ShouldEqual, 0)
		})

		Convey("when quota is exhausted", func() {
			limiter := NewConcurrencyLimiter(1).(*ConcurrencyLimiter)

			// Acquire the only available slot
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldEqual, 0)

			// Probe should return InfDuration when no quota available
			delay := limiter.Probe()
			So(delay, ShouldEqual, repository.InfDuration)

			// Cleanup
			res.Cancel()
		})

		Convey("after releasing quota", func() {
			limiter := NewConcurrencyLimiter(1).(*ConcurrencyLimiter)

			// Acquire and then release
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			res.Cancel()

			// Probe should return 0 after release
			delay := limiter.Probe()
			So(delay, ShouldEqual, 0)
		})
	})
}

func TestConcurrencyLimiter_Reserve(t *testing.T) {
	Convey("Test ConcurrencyLimiter Reserve", t, func() {
		Convey("should succeed when quota is available", func() {
			limiter := NewConcurrencyLimiter(2)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.Delay(), ShouldEqual, 0)

			// Cleanup
			res.Cancel()
		})

		Convey("should return reservation with InfDuration when quota exhausted", func() {
			limiter := NewConcurrencyLimiter(1)

			// First reservation succeeds
			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res1.Delay(), ShouldEqual, 0)

			// Second reservation returns InfDuration delay
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, repository.InfDuration)

			// Cleanup
			res1.Cancel()
			res2.Cancel()
		})

		Convey("should allow multiple reservations within limit", func() {
			limiter := NewConcurrencyLimiter(3)
			var reservations []repository.Reservation

			// Reserve 3 slots
			for i := 0; i < 3; i++ {
				res, err := limiter.Reserve()
				So(err, ShouldBeNil)
				So(res.Delay(), ShouldEqual, 0)
				reservations = append(reservations, res)
			}

			// 4th reservation should have InfDuration delay
			res4, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res4.Delay(), ShouldEqual, repository.InfDuration)

			// Cleanup
			for _, res := range reservations {
				res.Cancel()
			}
			res4.Cancel()
		})
	})
}

func TestConcurrencyReservation_Delay(t *testing.T) {
	Convey("Test ConcurrencyReservation Delay", t, func() {
		Convey("should return 0 when acquired immediately", func() {
			limiter := NewConcurrencyLimiter(1)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res.Delay(), ShouldEqual, 0)

			res.Cancel()
		})

		Convey("should return InfDuration when not acquired", func() {
			limiter := NewConcurrencyLimiter(1)

			// Acquire the only slot
			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)

			// Second reservation cannot be acquired
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, repository.InfDuration)

			res1.Cancel()
			res2.Cancel()
		})
	})
}

func TestConcurrencyReservation_Wait(t *testing.T) {
	Convey("Test ConcurrencyReservation Wait", t, func() {
		Convey("should return immediately if already acquired", func() {
			limiter := NewConcurrencyLimiter(1)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			ctx := context.Background()
			err = res.Wait(ctx)
			So(err, ShouldBeNil)

			res.Cancel()
		})

		Convey("should block until quota becomes available", func() {
			limiter := NewConcurrencyLimiter(1)

			// First reservation acquires immediately
			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res1.Delay(), ShouldEqual, 0)

			// Second reservation needs to wait
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, repository.InfDuration)

			var wg sync.WaitGroup
			wg.Add(1)
			waitCompleted := false
			var waitErr error

			// Wait in background
			go func() {
				defer wg.Done()
				waitErr = res2.Wait(context.Background())
				waitCompleted = true
			}()

			// Give some time to ensure res2.Wait() is blocking
			time.Sleep(1 * time.Millisecond)
			So(waitCompleted, ShouldBeFalse)

			// Release res1, which should unblock res2
			res1.Cancel()

			// Wait for res2 to complete
			wg.Wait()
			So(waitCompleted, ShouldBeTrue)
			So(waitErr, ShouldBeNil)

			res2.Cancel()
		})

		Convey("should respect context cancellation", func() {
			limiter := NewConcurrencyLimiter(1)

			// Acquire the only slot
			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)

			// Second reservation needs to wait
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)

			ctx, cancel := context.WithCancel(context.Background())
			var waitErr error

			// Wait with cancellable context
			go func() {
				waitErr = res2.Wait(ctx)
			}()

			// Give some time to ensure Wait() is blocking
			time.Sleep(1 * time.Millisecond)

			// Cancel the context
			cancel()

			// Wait should return context error
			time.Sleep(1 * time.Millisecond)
			So(waitErr, ShouldEqual, context.Canceled)

			res1.Cancel()
			res2.Cancel()
		})

		Convey("should respect context timeout", func() {
			limiter := NewConcurrencyLimiter(1)

			// Acquire the only slot
			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)

			// Second reservation needs to wait
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			waitErr := res2.Wait(ctx)
			So(waitErr, ShouldEqual, context.DeadlineExceeded)

			res1.Cancel()
			res2.Cancel()
		})

		Convey("should be idempotent when already acquired", func() {
			limiter := NewConcurrencyLimiter(1)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			ctx := context.Background()

			// First Wait
			err = res.Wait(ctx)
			So(err, ShouldBeNil)

			// Second Wait should also succeed
			err = res.Wait(ctx)
			So(err, ShouldBeNil)

			res.Cancel()
		})
	})
}

func TestConcurrencyReservation_Cancel(t *testing.T) {
	Convey("Test ConcurrencyReservation Cancel", t, func() {
		Convey("should release acquired quota", func() {
			limiter := NewConcurrencyLimiter(1)

			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res1.Delay(), ShouldEqual, 0)

			// Cancel the reservation
			res1.Cancel()

			// Should be able to reserve again
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)

			res2.Cancel()
		})

		Convey("should be idempotent", func() {
			limiter := NewConcurrencyLimiter(1)

			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			// Cancel multiple times
			res.Cancel()
			res.Cancel()
			res.Cancel()

			// Should only release once
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)

			res2.Cancel()
		})

		Convey("should not release if not acquired", func() {
			limiter := NewConcurrencyLimiter(1)

			// First reservation acquires
			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)

			// Second reservation does not acquire
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, repository.InfDuration)

			// Cancel the non-acquired reservation
			res2.Cancel()

			// Should still not be able to acquire
			res3, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res3.Delay(), ShouldEqual, repository.InfDuration)

			res1.Cancel()
			res3.Cancel()
		})
	})
}

func TestConcurrencyReservation_Complete(t *testing.T) {
	Convey("Test ConcurrencyReservation Complete", t, func() {
		Convey("should release quota like Cancel", func() {
			limiter := NewConcurrencyLimiter(1)

			res1, err := limiter.Reserve()
			So(err, ShouldBeNil)

			// Complete the reservation
			res1.Complete()

			// Should be able to reserve again
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)

			res2.Cancel()
		})

		Convey("should be idempotent", func() {
			limiter := NewConcurrencyLimiter(1)

			res, err := limiter.Reserve()
			So(err, ShouldBeNil)

			// Complete multiple times
			res.Complete()
			res.Complete()
			res.Complete()

			// Should only release once
			res2, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res2.Delay(), ShouldEqual, 0)

			res2.Cancel()
		})
	})
}

func TestConcurrencyLimiter_ConcurrentAccess(t *testing.T) {
	Convey("Test ConcurrencyLimiter concurrent access", t, func() {
		Convey("should handle concurrent reservations correctly", func() {
			limiter := NewConcurrencyLimiter(10)
			concurrency := 50
			iterations := 100

			var wg sync.WaitGroup
			var successCount atomic.Int64

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < iterations; j++ {
						res, err := limiter.Reserve()
						if err != nil {
							continue
						}

						if res.Delay() == 0 {
							successCount.Add(1)
							res.Complete()
						} else {
							// Try to wait for quota
							ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
							err := res.Wait(ctx)
							cancel()

							if err == nil {
								successCount.Add(1)
								res.Complete()
							} else {
								res.Cancel()
							}
						}
					}
				}()
			}

			wg.Wait()

			// Should have processed requests successfully
			So(successCount.Load(), ShouldEqual, int64(concurrency*iterations))
		})
	})
}

func TestConcurrencyLimiter_InterfaceCompliance(t *testing.T) {
	Convey("Test interface compliance", t, func() {
		Convey("ConcurrencyLimiter should implement RequestLimiter", func() {
			limiter := NewConcurrencyLimiter(1)
			So(limiter, ShouldImplement, (*repository.RequestLimiter)(nil))
		})

		Convey("concurrencyReservation should implement Reservation", func() {
			limiter := NewConcurrencyLimiter(1)
			res, err := limiter.Reserve()
			So(err, ShouldBeNil)
			So(res, ShouldImplement, (*repository.Reservation)(nil))
			res.Cancel()
		})
	})
}
