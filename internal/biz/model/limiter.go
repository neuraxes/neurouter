package model

import (
	"context"
	"time"

	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/data/limiter/local"
)

// limiterGroup groups request and token limiters for a single scope (upstream or model level).
type limiterGroup struct {
	requestLimiters []repository.RequestLimiter // concurrency, RPM, RPD
	tokenLimiters   []repository.TokenLimiter   // TPM, TPD
}

// newLimiterGroup creates a limiterGroup from scheduling configuration values.
// Limiters with zero or negative limits are automatically filtered out (nil return from factory).
func newLimiterGroup(concurrency, rpm, rpd, tpm, tpd uint64) *limiterGroup {
	g := &limiterGroup{}
	if l := local.NewConcurrencyLimiter(int64(concurrency)); l != nil {
		g.requestLimiters = append(g.requestLimiters, l)
	}
	if l := local.NewRPMLimiter(int64(rpm)); l != nil {
		g.requestLimiters = append(g.requestLimiters, l)
	}
	if l := local.NewDailyRequestLimiter(int64(rpd)); l != nil {
		g.requestLimiters = append(g.requestLimiters, l)
	}
	if l := local.NewTPMLimiter(int64(tpm)); l != nil {
		g.tokenLimiters = append(g.tokenLimiters, l)
	}
	if l := local.NewDailyTokenLimiter(int64(tpd)); l != nil {
		g.tokenLimiters = append(g.tokenLimiters, l)
	}
	return g
}

// probeDelay returns the maximum wait time across all limiters in the group.
func (g *limiterGroup) probeDelay(estimatedTokens int64) time.Duration {
	if g == nil {
		return 0
	}
	maxDelay := time.Duration(0)
	for _, rl := range g.requestLimiters {
		if d := rl.Probe(); d > maxDelay {
			maxDelay = d
		}
	}
	if estimatedTokens > 0 {
		for _, tl := range g.tokenLimiters {
			if d := tl.Probe(estimatedTokens); d > maxDelay {
				maxDelay = d
			}
		}
	}
	return maxDelay
}

// reservationSet holds all reservations acquired for a single model election.
type reservationSet struct {
	requestReservations []repository.Reservation
	tokenReservations   []repository.TokenReservation
}

// maxDelay returns the maximum delay across all reservations.
func (rs *reservationSet) maxDelay() time.Duration {
	maxD := time.Duration(0)
	for _, r := range rs.requestReservations {
		if d := r.Delay(); d > maxD {
			maxD = d
		}
	}
	for _, r := range rs.tokenReservations {
		if d := r.Delay(); d > maxD {
			maxD = d
		}
	}
	return maxD
}

// wait blocks until all reservations are ready or the context is cancelled.
// On failure, all reservations are cancelled automatically.
func (rs *reservationSet) wait(ctx context.Context) error {
	if rs.maxDelay() == 0 {
		return nil
	}
	for _, r := range rs.requestReservations {
		if err := r.Wait(ctx); err != nil {
			rs.cancel()
			return err
		}
	}
	for _, r := range rs.tokenReservations {
		if err := r.Wait(ctx); err != nil {
			rs.cancel()
			return err
		}
	}
	return nil
}

// cancel cancels all held reservations, returning quota.
func (rs *reservationSet) cancel() {
	for _, r := range rs.requestReservations {
		r.Cancel()
	}
	for _, r := range rs.tokenReservations {
		r.Cancel()
	}
	rs.requestReservations = nil
	rs.tokenReservations = nil
}

// complete finalizes all reservations after actual usage.
// Token reservations are completed with actual token count for accurate accounting.
func (rs *reservationSet) complete(actualTokens int64) {
	for _, r := range rs.requestReservations {
		r.Complete()
	}
	for _, r := range rs.tokenReservations {
		r.CompleteWithActual(actualTokens)
	}
	rs.requestReservations = nil
	rs.tokenReservations = nil
}

// probeModelDelay computes the maximum delay across upstream and model limiter groups.
func probeModelDelay(m *model, estimatedTokens int64) time.Duration {
	upstreamMaxDelay := m.upstreamLimiters.probeDelay(estimatedTokens)
	modelMaxDelay := m.modelLimiters.probeDelay(estimatedTokens)
	return max(upstreamMaxDelay, modelMaxDelay)
}

// tryReserveAll attempts to reserve all limiters for a model (non-blocking).
// On any failure, all previously acquired reservations are cancelled.
func tryReserveAll(m *model, estimatedTokens int64) (*reservationSet, error) {
	rs := &reservationSet{}

	for _, g := range []*limiterGroup{m.upstreamLimiters, m.modelLimiters} {
		if g == nil {
			continue
		}
		for _, rl := range g.requestLimiters {
			r, err := rl.Reserve()
			if err != nil {
				rs.cancel()
				return nil, err
			}
			rs.requestReservations = append(rs.requestReservations, r)
		}
		if estimatedTokens > 0 {
			for _, tl := range g.tokenLimiters {
				r, err := tl.Reserve(estimatedTokens)
				if err != nil {
					rs.cancel()
					return nil, err
				}
				rs.tokenReservations = append(rs.tokenReservations, r)
			}
		}
	}

	return rs, nil
}
