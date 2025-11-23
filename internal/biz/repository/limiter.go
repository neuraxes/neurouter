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
	// - InfDuration if the waiting duration is not predictable.
	Probe() time.Duration

	// Reserve tries to reserves 1 request.
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
	// - InfDuration if the waiting duration is not predictable.
	Probe(tokens int64) time.Duration

	// Reserve tries to reserves 1 request.
	Reserve(tokens int64) (TokenReservation, error)
}
