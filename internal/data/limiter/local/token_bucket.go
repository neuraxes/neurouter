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

// tokenBucket contains the basic state and algorithm of the token bucket, reused by TPM and RPM
type tokenBucket struct {
	mu         sync.Mutex
	rate       float64
	burst      float64
	tokens     float64
	lastUpdate time.Time
}

func newTokenBucket(rate, burst float64) *tokenBucket {
	return &tokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     burst,
		lastUpdate: time.Now(),
	}
}

// fill updates the current time and replenishes tokens
func (b *tokenBucket) fill() {
	now := time.Now()
	elapsed := now.Sub(b.lastUpdate).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.lastUpdate = now
}

// probe probes the wait time (does not modify state)
func (b *tokenBucket) probe(cost float64) time.Duration {
	if cost >= b.burst {
		return repository.InfDuration
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Simulate sync
	elapsed := time.Since(b.lastUpdate).Seconds()
	estimatedTokens := b.tokens + elapsed*b.rate
	if estimatedTokens > b.burst {
		estimatedTokens = b.burst
	}

	// Simulate deduction
	remaining := estimatedTokens - cost
	if remaining >= 0 {
		return 0
	}

	// Calculate wait time: deficit amount / generation rate
	return time.Duration(-remaining / b.rate * float64(time.Second))
}

// reserve performs actual deduction and returns ready time
func (b *tokenBucket) reserve(cost float64) time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Sync token bucket state to current time
	b.fill()

	// Deduct estimated cost
	b.tokens -= cost

	// If tokens remain non-negative, quota is immediately available
	if b.tokens >= 0 {
		return time.Now()
	}

	// Otherwise compute the time point when deficit will be replenished
	wait := -b.tokens / b.rate * float64(time.Second)
	return time.Now().Add(time.Duration(wait))
}

// adjust corrects the balance (refund excess, deduct shortage)
func (b *tokenBucket) adjust(diff float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.fill()

	b.tokens += diff
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
}

// rpmLimiter implements the repository.RequestLimiter for RPM
type rpmLimiter struct {
	bucket *tokenBucket
}

// NewRPMLimiter creates an RPM limiter
func NewRPMLimiter(limit int64) repository.RequestLimiter {
	return &rpmLimiter{
		bucket: newTokenBucket(float64(limit)/60.0, float64(limit)),
	}
}

func (l *rpmLimiter) Probe() time.Duration {
	return l.bucket.probe(1)
}

func (l *rpmLimiter) Reserve() (repository.Reservation, error) {
	return &requestReservation{
		limiter: l,
		readyAt: l.bucket.reserve(1),
	}, nil
}

// requestReservation implements the Reservation interface
type requestReservation struct {
	limiter *rpmLimiter
	readyAt time.Time
}

func (r *requestReservation) Delay() time.Duration {
	return max(0, time.Until(r.readyAt))
}

func (r *requestReservation) Wait(ctx context.Context) error {
	select {
	case <-time.After(r.Delay()):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *requestReservation) Cancel() {
	// Cancel means the request was not executed, return 1 unit
	r.limiter.bucket.adjust(1)
}

func (r *requestReservation) Complete() {
	// Complete means the request was executed, consuming exactly 1 unit
}

type tpmLimiter struct {
	bucket *tokenBucket
}

// NewTPMLimiter implements the repository.TokenLimiter for TPM
func NewTPMLimiter(limit int64) repository.TokenLimiter {
	return &tpmLimiter{
		bucket: newTokenBucket(float64(limit)/60.0, float64(limit)),
	}
}

func (l *tpmLimiter) Probe(tokens int64) time.Duration {
	return l.bucket.probe(float64(tokens))
}

func (l *tpmLimiter) Reserve(tokens int64) (repository.TokenReservation, error) {
	if tokens > int64(l.bucket.burst) {
		return nil, v1.ErrorTokenQuotaExhausted("token quota exhausted")
	}
	return &tokenReservation{
		limiter: l,
		tokens:  tokens,
		readyAt: l.bucket.reserve(float64(tokens)),
	}, nil
}

// tokenReservation implements the TokenReservation interface
type tokenReservation struct {
	limiter *tpmLimiter
	tokens  int64
	readyAt time.Time
}

func (r *tokenReservation) Delay() time.Duration {
	return max(0, time.Until(r.readyAt))
}

func (r *tokenReservation) Wait(ctx context.Context) error {
	select {
	case <-time.After(r.Delay()):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *tokenReservation) Cancel() {
	r.CompleteWithActual(0)
}

func (r *tokenReservation) Complete() {
	// Settle by default using estimated value (no difference)
}

func (r *tokenReservation) CompleteWithActual(actualTokens int64) {
	// Calculate difference: estimated - actual
	// diff > 0: overestimated, refund
	// diff < 0: underestimated, deduct more
	diff := float64(r.tokens - actualTokens)
	if diff == 0 {
		return
	}
	r.limiter.bucket.adjust(diff)
}
