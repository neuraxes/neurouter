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

package repository

import (
	"context"
	"math"
	"time"
)

var InfDuration = time.Duration(math.MaxInt64)

// Reservation represents a reserved quota.
type Reservation interface {
	// Delay returns the time to wait before the reservation can be used.
	// - zero duration if quota is available immediately.
	// - non-zero duration if quota is not available immediately.
	// - InfDuration if the waiting duration is not predictable.
	Delay() time.Duration

	// Wait blocks until the resource is ready or the context is done.
	Wait(context.Context) error

	// Cancel returns the reserved quota without consuming it.
	Cancel()

	// Complete the reservation after actual usage.
	Complete()
}

// RequestLimiter handles request-based rate limiting (RPM, RPD, Concurrency).
// Each reservation reserves exactly 1 request.
type RequestLimiter interface {
	// Probe detects the waiting duration to reserve 1 request without blocking.
	// Returns
	// - zero duration if quota is available immediately.
	// - non-zero duration if quota is not available immediately.
	// - InfDuration if unwaitable or the waiting duration is not predictable.
	Probe() time.Duration

	// Reserve tries to acquire quota for 1 request without blocking.
	// Returns
	// - Reservation with zero delay if quota is available immediately.
	// - Reservation with non-zero delay if waiting is needed.
	// - error if reservation fails.
	Reserve() (Reservation, error)
}

// TokenReservation represents a reserved token quota.
type TokenReservation interface {
	Reservation

	// CompleteWithActual completes the reservation after actual usage with token adjustment.
	// - If actual < estimated: refunds the difference.
	// - If actual > estimated: charges the overdraft (if supported).
	CompleteWithActual(actualTokens int64)
}

// TokenLimiter handles token-based rate limiting (TPM, TPD).
// Requires token estimation and supports adjustment with actual usage.
type TokenLimiter interface {
	// Probe detects the waiting duration to reserve 1 request without blocking.
	// Returns
	// - zero duration if quota is available immediately.
	// - non-zero duration if quota is not available immediately.
	// - InfDuration if unwaitable or the waiting duration is not predictable.
	Probe(tokens int64) time.Duration

	// Reserve tries to acquire quota for given tokens without blocking.
	// Returns
	// - Reservation with zero delay if quota is available immediately.
	// - Reservation with non-zero delay if waiting is needed.
	// - error if reservation fails.
	Reserve(tokens int64) (TokenReservation, error)
}
