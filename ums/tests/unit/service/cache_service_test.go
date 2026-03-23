package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Regression tests for BUG-001 ─────────────────────────────────────────────
// CacheService.Incr must propagate pipeline errors to the caller.
// Previously the error was swallowed and 0 was returned silently.

func TestCacheIncr_PropagatesError(t *testing.T) {
	cache := new(mockCache)
	// Simulate a Redis connection failure
	cache.On("Incr", mock.Anything, "test:key", time.Minute).
		Return(int64(0), assert.AnError)

	_, err := cache.Incr(context.Background(), "test:key", time.Minute)
	require.Error(t, err, "Incr must propagate pipeline errors — BUG-001 regression")
}

func TestCacheIncr_SuccessReturnsCount(t *testing.T) {
	cache := new(mockCache)
	cache.On("Incr", mock.Anything, "rate:alice@x.com", 5*time.Minute).
		Return(int64(3), nil)

	n, err := cache.Incr(context.Background(), "rate:alice@x.com", 5*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

func TestCacheIncr_FirstCallReturnsOne(t *testing.T) {
	cache := new(mockCache)
	cache.On("Incr", mock.Anything, "rate:new@x.com", 5*time.Minute).
		Return(int64(1), nil)

	n, err := cache.Incr(context.Background(), "rate:new@x.com", 5*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n, "first Incr should return 1")
}
