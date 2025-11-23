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
	"time"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

// getNextMidnight returns the next midnight in the specified timezone.
func getNextMidnight(loc *time.Location) time.Time {
	tomorrow := time.Now().In(loc).Add(24 * time.Hour)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)
}

// dailyQuotaState holds shared state for daily quota limiters.
type dailyQuotaState struct {
	mu sync.Mutex

	limit     int64          // Maximum units per day
	used      int64          // Units used in current day
	resetTime time.Time      // Next reset time (midnight in specified timezone)
	location  *time.Location // Timezone for daily resets
}

// flush resets the quota if we've passed midnight in the configured timezone.
// Must be called with lock held.
func (s *dailyQuotaState) flush() {
	if time.Now().In(s.location).After(s.resetTime) {
		s.used = 0
		s.resetTime = getNextMidnight(s.location)
	}
}

// DailyRequestLimiter implements limiter for RPD.
type DailyRequestLimiter struct {
	state *dailyQuotaState
}

// NewDailyRequestLimiter creates a new limiter for RPD.
// Resets at UTC midnight. If limit is 0 or negative, returns nil (unlimited).
func NewDailyRequestLimiter(RPDLimit int64) repository.RequestLimiter {
	return NewDailyRequestLimiterWithTimeZone(RPDLimit, time.UTC)
}

// NewDailyRequestLimiterWithTimeZone creates a new limiter for RPD with custom timezone.
// Resets at midnight in the specified timezone. If limit is 0 or negative, returns nil (unlimited).
func NewDailyRequestLimiterWithTimeZone(RPDLimit int64, loc *time.Location) repository.RequestLimiter {
	if RPDLimit <= 0 {
		return nil
	}

	return &DailyRequestLimiter{
		state: &dailyQuotaState{
			limit:     RPDLimit,
			used:      0,
			resetTime: getNextMidnight(loc),
			location:  loc,
		},
	}
}

// Probe detects the waiting duration to reserve 1 request without blocking.
func (d *DailyRequestLimiter) Probe() time.Duration {
	d.state.mu.Lock()
	defer d.state.mu.Unlock()

	d.state.flush()

	if d.state.used < d.state.limit {
		return 0
	}

	// Return wait time until next reset
	return time.Until(d.state.resetTime)
}

// Reserve tries to acquire quota for 1 request without blocking.
func (d *DailyRequestLimiter) Reserve() (repository.Reservation, error) {
	d.state.mu.Lock()
	defer d.state.mu.Unlock()

	d.state.flush()

	available := d.state.used < d.state.limit
	if available {
		d.state.used++
	}

	return &dailyRequestReservation{
		state:    d.state,
		reserved: 1,
		acquired: available,
	}, nil
}

// dailyRequestReservation implements repository.Reservation for daily request limits.
type dailyRequestReservation struct {
	state    *dailyQuotaState
	reserved int64
	acquired bool
	released bool
}

// Delay returns the time to wait before the reservation can be used.
func (r *dailyRequestReservation) Delay() time.Duration {
	if r.acquired {
		return 0
	}

	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	r.state.flush()

	if r.state.used < r.state.limit {
		return 0
	}

	return time.Until(r.state.resetTime)
}

// Wait blocks until the resource is ready or the context is done.
func (r *dailyRequestReservation) Wait(ctx context.Context) error {
	if r.acquired || r.released {
		return nil
	}

	// Wait until quota becomes available
	for {
		r.state.mu.Lock()
		r.state.flush()

		if r.state.used < r.state.limit {
			r.state.used++
			r.acquired = true
			r.state.mu.Unlock()
			return nil
		}

		// Wait until next reset
		dur := time.Until(r.state.resetTime)
		r.state.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			// Retry after reset
			continue
		}
	}
}

// Cancel returns the reserved quota without consuming it.
func (r *dailyRequestReservation) Cancel() {
	if r.released {
		return
	}

	if r.acquired {
		r.state.mu.Lock()
		r.state.used = max(0, r.state.used-r.reserved)
		r.state.mu.Unlock()
	}

	r.released = true
}

// Complete the reservation after actual usage.
func (r *dailyRequestReservation) Complete() {
	r.released = true
}

// DailyTokenLimiter implements limiter for TPD.
type DailyTokenLimiter struct {
	state *dailyQuotaState
}

// NewDailyTokenLimiter creates a new limiter for TPD.
// Resets at UTC midnight. If limit is 0 or negative, returns nil (unlimited).
func NewDailyTokenLimiter(TPDLimit int64) repository.TokenLimiter {
	return NewDailyTokenLimiterWithTimeZone(TPDLimit, time.UTC)
}

// NewDailyTokenLimiterWithTimeZone creates a new limiter for TPD with custom timezone.
// Resets at midnight in the specified timezone. If limit is 0 or negative, returns nil (unlimited).
func NewDailyTokenLimiterWithTimeZone(TPDLimit int64, loc *time.Location) repository.TokenLimiter {
	if TPDLimit <= 0 {
		return nil
	}

	return &DailyTokenLimiter{
		state: &dailyQuotaState{
			limit:     TPDLimit,
			used:      0,
			resetTime: getNextMidnight(loc),
			location:  loc,
		},
	}
}

// Probe detects the waiting duration to reserve tokens without blocking.
func (d *DailyTokenLimiter) Probe(tokens int64) time.Duration {
	// Check if request exceeds limit
	if tokens > d.state.limit {
		return repository.InfDuration
	}

	d.state.mu.Lock()
	defer d.state.mu.Unlock()

	d.state.flush()

	if d.state.used+tokens <= d.state.limit {
		return 0
	}

	// Return wait time until next reset
	return time.Until(d.state.resetTime)
}

// Reserve tries to acquire quota for given tokens without blocking.
func (d *DailyTokenLimiter) Reserve(tokens int64) (repository.TokenReservation, error) {
	// Check if request exceeds limit
	if tokens > d.state.limit {
		return nil, v1.ErrorTokenQuotaExhausted("request exceeds daily token quota")
	}

	d.state.mu.Lock()
	defer d.state.mu.Unlock()

	d.state.flush()

	available := d.state.used+tokens <= d.state.limit
	if available {
		d.state.used += tokens
	}

	return &dailyTokenReservation{
		state:    d.state,
		reserved: tokens,
		acquired: available,
	}, nil
}

// dailyTokenReservation implements repository.TokenReservation for daily token limits.
type dailyTokenReservation struct {
	state    *dailyQuotaState
	reserved int64
	acquired bool
	released bool
}

// Delay returns the time to wait before the reservation can be used.
func (r *dailyTokenReservation) Delay() time.Duration {
	if r.acquired {
		return 0
	}

	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	r.state.flush()

	if r.state.used+r.reserved <= r.state.limit {
		return 0
	}

	return time.Until(r.state.resetTime)
}

// Wait blocks until the resource is ready or the context is done.
func (r *dailyTokenReservation) Wait(ctx context.Context) error {
	if r.acquired || r.released {
		return nil
	}

	// Wait until quota becomes available
	for {
		r.state.mu.Lock()
		r.state.flush()

		if r.state.used+r.reserved <= r.state.limit {
			r.state.used += r.reserved
			r.acquired = true
			r.state.mu.Unlock()
			return nil
		}

		// Wait until next reset
		dur := time.Until(r.state.resetTime)
		r.state.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			// Retry after reset
			continue
		}
	}
}

// Cancel returns the reserved quota without consuming it.
func (r *dailyTokenReservation) Cancel() {
	if r.released {
		return
	}

	if r.acquired {
		r.state.mu.Lock()
		r.state.used = max(0, r.state.used-r.reserved)
		r.state.mu.Unlock()
	}

	r.released = true
}

// Complete the reservation after actual usage.
func (r *dailyTokenReservation) Complete() {
	r.released = true
}

// CompleteWithActual completes the reservation after actual usage with token adjustment.
func (r *dailyTokenReservation) CompleteWithActual(actualTokens int64) {
	if r.released {
		return
	}

	if r.acquired {
		r.state.mu.Lock()
		r.state.used = max(0, r.state.used-r.reserved+actualTokens)
		r.state.mu.Unlock()
	}

	r.released = true
}

var _ repository.RequestLimiter = (*DailyRequestLimiter)(nil)
var _ repository.Reservation = (*dailyRequestReservation)(nil)
var _ repository.TokenLimiter = (*DailyTokenLimiter)(nil)
var _ repository.TokenReservation = (*dailyTokenReservation)(nil)
