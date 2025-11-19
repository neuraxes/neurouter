package model

import (
	"context"
	"math/rand/v2"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"golang.org/x/sync/semaphore"
)

// acquireSemaphores acquires both semaphores in order
func acquireSemaphores(ctx context.Context, s1, s2 *semaphore.Weighted) error {
	if s1 != nil {
		if err := s1.Acquire(ctx, 1); err != nil {
			return err
		}
	}

	if s2 != nil {
		if err := s2.Acquire(ctx, 1); err != nil {
			// Release first semaphore on failure
			if s1 != nil {
				s1.Release(1)
			}
			return err
		}
	}

	return nil
}

// tryAcquireSemaphores tries to acquire semaphores without blocking
func tryAcquireSemaphores(s1, s2 *semaphore.Weighted) bool {
	if s1 != nil {
		if !s1.TryAcquire(1) {
			return false
		}
	}

	if s2 != nil {
		if !s2.TryAcquire(1) {
			// Release first semaphore on failure
			if s1 != nil {
				s1.Release(1)
			}
			return false
		}
	}

	return true
}

// electFromCandidates randomly selects from candidates, trying to acquire semaphores.
// It tries all candidates randomly, falling back to the next when concurrency limit reached.
// If all models are busy, it blocks on the first candidate until it becomes available.
func electFromCandidates(ctx context.Context, candidates []*model) (*model, error) {
	if len(candidates) == 0 {
		return nil, v1.ErrorNoUpstream("no upstream found")
	}

	// Shuffle candidates for random selection
	shuffled := make([]*model, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Try to acquire semaphores for each candidate without blocking
	for _, m := range shuffled {
		if tryAcquireSemaphores(m.upstreamSem, m.modelSem) {
			return m, nil
		}
	}

	// All models are busy, block on the first candidate
	if err := acquireSemaphores(ctx, shuffled[0].upstreamSem, shuffled[0].modelSem); err != nil {
		return nil, err
	}

	return shuffled[0], nil
}
