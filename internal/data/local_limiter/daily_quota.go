package local_limiter

import (
	"context"
	"sync"
	"time"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

// getNextMidnightUTC returns the next midnight UTC time.
func getNextMidnightUTC() time.Time {
	now := time.Now().UTC()
	tomorrow := now.Add(24 * time.Hour)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)
}

// dailyQuotaState holds shared state for daily quota tracking.
type dailyQuotaState struct {
	mu sync.Mutex

	limit     int64     // Maximum units per day
	used      int64     // Units used in current day
	resetTime time.Time // Next reset time (midnight UTC)
}

// flush resets the quota if we've passed midnight UTC.
// Must be called with lock held.
func (s *dailyQuotaState) flush() {
	if time.Now().UTC().After(s.resetTime) {
		s.used = 0
		s.resetTime = getNextMidnightUTC()
	}
}

// dailyRequestReservation implements repository.RequestReservation for daily request limits.
type dailyRequestReservation struct {
	state    *dailyQuotaState
	reserved int64
	released bool
}

// Cancel returns the reserved quota (election failure - request not made).
func (r *dailyRequestReservation) Cancel() {
	if r.released {
		return
	}

	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	r.state.used = max(0, r.state.used-r.reserved)
	r.released = true
}

// Release is a no-op for time-based limiters (quota consumed when request was made).
func (r *dailyRequestReservation) Release() {
	r.released = true
}

// DailyRequestLimiter implements daily quota tracking for RPD.
// Resets at UTC midnight each day.
type DailyRequestLimiter struct {
	state *dailyQuotaState
}

// NewDailyRequestLimiter creates a new daily quota limiter for RPD.
// If limit is 0 or negative, returns nil (unlimited).
func NewDailyRequestLimiter(RPDLimit int64) repository.RequestLimiter {
	if RPDLimit <= 0 {
		return nil
	}

	return &DailyRequestLimiter{
		state: &dailyQuotaState{
			limit:     RPDLimit,
			used:      0,
			resetTime: getNextMidnightUTC(),
		},
	}
}

// TryReserve attempts to reserve 1 request without blocking.
func (d *DailyRequestLimiter) TryReserve() (repository.RequestReservation, time.Duration) {
	d.state.mu.Lock()
	defer d.state.mu.Unlock()

	d.state.flush()

	if d.state.used < d.state.limit {
		d.state.used++
		return &dailyRequestReservation{
			state:    d.state,
			reserved: 1,
		}, 0
	}

	// Return wait time until next reset
	return nil, time.Until(d.state.resetTime)
}

// Reserve reserves 1 request, blocking until available or context is done.
func (d *DailyRequestLimiter) Reserve(ctx context.Context) (repository.RequestReservation, error) {
	for {
		d.state.mu.Lock()
		d.state.flush()

		if d.state.used < d.state.limit {
			d.state.used++
			d.state.mu.Unlock()
			return &dailyRequestReservation{
				state:    d.state,
				reserved: 1,
			}, nil
		}

		// Wait until next reset
		dur := time.Until(d.state.resetTime)
		d.state.mu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(dur):
			// Retry after reset
			continue
		}
	}
}

// dailyTokenReservation implements repository.TokenReservation for daily token limits.
type dailyTokenReservation struct {
	state    *dailyQuotaState
	reserved int64
	released bool
}

// Cancel returns the reserved quota (election failure - tokens not consumed).
func (r *dailyTokenReservation) Cancel() {
	if r.released {
		return
	}

	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	r.state.used = max(0, r.state.used-r.reserved)
	r.released = true
}

// Release adjusts the quota to actual usage.
func (r *dailyTokenReservation) Release(actualTokens int64) {
	if r.released {
		return
	}

	// Only adjust if actual usage is known
	if actualTokens > 0 {
		r.state.mu.Lock()
		r.state.used = max(0, r.state.used-r.reserved+actualTokens)
		r.state.mu.Unlock()
	}
	// Otherwise keep the estimate (no adjustment needed)

	r.released = true
}

// DailyTokenLimiter implements daily quota tracking for TPD.
// Resets at UTC midnight each day.
type DailyTokenLimiter struct {
	state *dailyQuotaState
}

// NewDailyTokenLimiter creates a new daily quota limiter for TPD.
// If limit is 0 or negative, returns nil (unlimited).
func NewDailyTokenLimiter(TPDLimit int64) repository.TokenLimiter {
	if TPDLimit <= 0 {
		return nil
	}

	return &DailyTokenLimiter{
		state: &dailyQuotaState{
			limit:     TPDLimit,
			used:      0,
			resetTime: getNextMidnightUTC(),
		},
	}
}

// TryReserve attempts to reserve tokens without blocking.
func (d *DailyTokenLimiter) TryReserve(tokens int64) (repository.TokenReservation, time.Duration) {
	// Check if request exceeds limit
	if tokens > d.state.limit {
		return nil, 0
	}

	d.state.mu.Lock()
	defer d.state.mu.Unlock()

	d.state.flush()

	if d.state.used+tokens <= d.state.limit {
		d.state.used += tokens
		return &dailyTokenReservation{
			state:    d.state,
			reserved: tokens,
		}, 0
	}

	// Return wait time until next reset
	return nil, time.Until(d.state.resetTime)
}

// Reserve reserves tokens, blocking until available or context is done.
func (d *DailyTokenLimiter) Reserve(ctx context.Context, tokens int64) (repository.TokenReservation, error) {
	// Check if request exceeds limit
	if tokens > d.state.limit {
		return nil, v1.ErrorTokenQuotaExhausted("request exceeds daily token quota")
	}

	for {
		d.state.mu.Lock()
		d.state.flush()

		if d.state.used+tokens <= d.state.limit {
			d.state.used += tokens
			d.state.mu.Unlock()
			return &dailyTokenReservation{
				state:    d.state,
				reserved: tokens,
			}, nil
		}

		// Wait until next reset
		dur := time.Until(d.state.resetTime)
		d.state.mu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(dur):
			// Retry after reset
			continue
		}
	}
}

var _ repository.RequestReservation = (*dailyRequestReservation)(nil)
var _ repository.RequestLimiter = (*DailyRequestLimiter)(nil)
var _ repository.TokenReservation = (*dailyTokenReservation)(nil)
var _ repository.TokenLimiter = (*DailyTokenLimiter)(nil)
