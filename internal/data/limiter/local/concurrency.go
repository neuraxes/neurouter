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
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/neuraxes/neurouter/internal/biz/repository"
)

// ConcurrencyLimiter wraps semaphore.Weighted to implement repository.RequestLimiter.
// It enforces a maximum number of concurrent requests.
type ConcurrencyLimiter struct {
	sem *semaphore.Weighted
}

// NewConcurrencyLimiter creates a new concurrency limiter.
// If limit is 0 or negative, returns nil (unlimited).
func NewConcurrencyLimiter(limit int64) repository.RequestLimiter {
	if limit <= 0 {
		return nil
	}
	return &ConcurrencyLimiter{
		sem: semaphore.NewWeighted(limit),
	}
}

// Probe detects the waiting duration to reserve 1 request without blocking.
// For concurrency limits, returns 0 if available, InfDuration if not (unpredictable wait time).
func (c *ConcurrencyLimiter) Probe() time.Duration {
	if c.sem.TryAcquire(1) {
		c.sem.Release(1)
		return 0
	}
	return repository.InfDuration // Concurrency has no predictable wait time
}

// Reserve tries to acquire quota for 1 request without blocking.
// Returns a reservation that may require waiting.
func (c *ConcurrencyLimiter) Reserve() (repository.Reservation, error) {
	return &concurrencyReservation{
		sem:      c.sem,
		acquired: c.sem.TryAcquire(1),
	}, nil
}

// concurrencyReservation implements repository.Reservation for concurrency limits.
type concurrencyReservation struct {
	sem      *semaphore.Weighted
	acquired bool
	released bool
}

// Delay returns the time to wait before the reservation can be used.
func (r *concurrencyReservation) Delay() time.Duration {
	if r.acquired {
		return 0
	}
	return repository.InfDuration
}

// Wait blocks until the resource is ready or the context is done.
func (r *concurrencyReservation) Wait(ctx context.Context) error {
	if r.acquired || r.released {
		return nil
	}

	// Otherwise, acquire the semaphore (blocking)
	if err := r.sem.Acquire(ctx, 1); err != nil {
		return err
	}

	r.acquired = true
	return nil
}

// Cancel returns the reserved quota without consuming it.
func (r *concurrencyReservation) Cancel() {
	if r.acquired && !r.released {
		r.sem.Release(1)
		r.released = true
	}
}

// Complete the reservation after actual usage.
func (r *concurrencyReservation) Complete() {
	r.Cancel()
}

var _ repository.RequestLimiter = (*ConcurrencyLimiter)(nil)
var _ repository.Reservation = (*concurrencyReservation)(nil)
