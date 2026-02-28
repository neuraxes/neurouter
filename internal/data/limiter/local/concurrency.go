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
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/neuraxes/neurouter/internal/biz/repository"
)

const (
	// defaultConcurrencyDelay is the initial estimated wait time for concurrency-limited requests.
	// It serves as the starting point before any actual observations are recorded.
	defaultConcurrencyDelay = 10 * time.Second

	// ewmaAlpha is the smoothing factor for the EWMA of observed wait times.
	// Higher values give more weight to recent observations.
	ewmaAlpha = 0.3
)

// ConcurrencyLimiter wraps semaphore.Weighted to implement repository.RequestLimiter.
// It enforces a maximum number of concurrent requests and dynamically estimates
// wait times based on observed delays using EWMA (Exponentially Weighted Moving Average).
type ConcurrencyLimiter struct {
	sem            *semaphore.Weighted
	estimatedDelay atomic.Int64 // nanoseconds, EWMA of observed wait times
}

// NewConcurrencyLimiter creates a new concurrency limiter.
// If limit is 0 or negative, returns nil (unlimited).
func NewConcurrencyLimiter(limit int64) repository.RequestLimiter {
	if limit <= 0 {
		return nil
	}
	l := &ConcurrencyLimiter{
		sem: semaphore.NewWeighted(limit),
	}
	l.estimatedDelay.Store(int64(defaultConcurrencyDelay))
	return l
}

// Probe detects the waiting duration to reserve 1 request without blocking.
// For concurrency limits, returns 0 if available, or the estimated wait time
// (based on EWMA of observed delays) if all slots are occupied.
func (c *ConcurrencyLimiter) Probe() time.Duration {
	if c.sem.TryAcquire(1) {
		c.sem.Release(1)
		return 0
	}
	return time.Duration(c.estimatedDelay.Load())
}

// Reserve tries to acquire quota for 1 request without blocking.
// Returns a reservation that may require waiting.
func (c *ConcurrencyLimiter) Reserve() (repository.Reservation, error) {
	return &concurrencyReservation{
		limiter:  c,
		acquired: c.sem.TryAcquire(1),
	}, nil
}

// recordWaitTime updates the estimated delay using EWMA based on actual observed wait time.
func (c *ConcurrencyLimiter) recordWaitTime(d time.Duration) {
	o := c.estimatedDelay.Load()
	n := int64(ewmaAlpha*float64(d) + (1-ewmaAlpha)*float64(o))
	c.estimatedDelay.Store(n)
}

// concurrencyReservation implements repository.Reservation for concurrency limits.
type concurrencyReservation struct {
	limiter  *ConcurrencyLimiter
	acquired bool
	released bool
}

// Delay returns the time to wait before the reservation can be used.
// Returns 0 if quota was acquired immediately, or the estimated wait time otherwise.
func (r *concurrencyReservation) Delay() time.Duration {
	if r.acquired {
		return 0
	}
	return time.Duration(r.limiter.estimatedDelay.Load())
}

// Wait blocks until the resource is ready or the context is done.
// On success, records the actual wait time to improve future delay estimates.
func (r *concurrencyReservation) Wait(ctx context.Context) error {
	if r.acquired || r.released {
		return nil
	}

	start := time.Now()

	// Acquire the semaphore (blocking)
	if err := r.limiter.sem.Acquire(ctx, 1); err != nil {
		return err
	}

	// Record actual wait time for future estimates
	r.limiter.recordWaitTime(time.Since(start))

	r.acquired = true
	return nil
}

// Cancel returns the reserved quota without consuming it.
func (r *concurrencyReservation) Cancel() {
	if r.acquired && !r.released {
		r.limiter.sem.Release(1)
		r.released = true
	}
}

// Complete the reservation after actual usage.
func (r *concurrencyReservation) Complete() {
	r.Cancel()
}

var _ repository.RequestLimiter = (*ConcurrencyLimiter)(nil)
var _ repository.Reservation = (*concurrencyReservation)(nil)
