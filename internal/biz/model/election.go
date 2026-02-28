package model

import (
	"cmp"
	"context"
	"math/rand/v2"
	"slices"
	"time"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

// scoredModel pairs a model with its probed delay for sorting during election.
type scoredModel struct {
	model *model
	delay time.Duration
}

// electFromCandidates selects the best candidate using a Probe → Rank → Reserve strategy.
//
// Phase 1 (Probe): evaluate each candidate's delay across all limiters.
// Phase 2 (Rank): classify into available (delay=0) and waitable (0 < delay < Inf);
// available candidates are shuffled for load balancing, waitable are sorted by delay.
// Phase 3 (Reserve): try to reserve all limiters for the best candidate; if reservation
// fails or waiting is needed, fall back to the next candidate.
//
// estimatedTokens is the estimated token cost for token limiters (0 to skip token probing).
func electFromCandidates(ctx context.Context, candidates []*model, estimatedTokens int64) (*model, *reservationSet, error) {
	if len(candidates) == 0 {
		return nil, nil, v1.ErrorNoUpstream("no upstream found")
	}

	// Phase 1: Probe & classify
	var available, waitable []scoredModel

	for _, m := range candidates {
		d := probeModelDelay(m, estimatedTokens)
		switch {
		case d == 0:
			available = append(available, scoredModel{model: m, delay: d})
		case d < repository.InfDuration:
			waitable = append(waitable, scoredModel{model: m, delay: d})
			// InfDuration: skip (unwaitable or quota exhausted)
		}
	}

	// Shuffle available candidates for load balancing
	rand.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})
	// Sort waitable candidates by delay ascending
	slices.SortFunc(waitable, func(a, b scoredModel) int {
		return cmp.Compare(a.delay, b.delay)
	})

	// Phase 2: Try reserve from available first, then waitable
	ordered := append(available, waitable...)
	for _, s := range ordered {
		rs, err := tryReserveAll(s.model, estimatedTokens)
		if err != nil {
			continue // This candidate failed, try next
		}
		// Phase 3: Wait if needed
		if err := rs.wait(ctx); err != nil {
			continue
		}
		return s.model, rs, nil
	}

	return nil, nil, v1.ErrorNoUpstream("no available upstream found")
}
