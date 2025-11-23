package local_limiter

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

// TryReserve attempts to acquire 1 slot without blocking.
func (c *ConcurrencyLimiter) TryReserve() (repository.RequestReservation, time.Duration) {
	if c.sem.TryAcquire(1) {
		return &concurrencyReservation{sem: c.sem}, 0
	}
	return nil, 0 // Concurrency has no predictable wait time
}

// Reserve acquires 1 slot, blocking until available or context is done.
func (c *ConcurrencyLimiter) Reserve(ctx context.Context) (repository.RequestReservation, error) {
	if err := c.sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	return &concurrencyReservation{sem: c.sem}, nil
}

// concurrencyReservation implements repository.RequestReservation for concurrency limits.
type concurrencyReservation struct {
	sem      *semaphore.Weighted
	released bool
}

// Cancel returns the semaphore slot (election failure).
func (r *concurrencyReservation) Cancel() {
	if !r.released {
		r.sem.Release(1)
		r.released = true
	}
}

// Release returns the semaphore slot (normal completion).
func (r *concurrencyReservation) Release() {
	if !r.released {
		r.sem.Release(1)
		r.released = true
	}
}

var _ repository.RequestLimiter = (*ConcurrencyLimiter)(nil)
var _ repository.RequestReservation = (*concurrencyReservation)(nil)
