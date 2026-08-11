package contextengine

import "fmt"

type CachePolicy struct {
	MaxEntries int
	MaxBytes   int64
}

func DefaultCachePolicy() CachePolicy {
	return CachePolicy{MaxEntries: 10_000, MaxBytes: 64 << 20}
}

func (policy CachePolicy) Validate() error {
	if policy.MaxEntries <= 0 || policy.MaxBytes <= 0 {
		return fmt.Errorf("cache limits must be positive: %w", ErrInvalidCachePolicy)
	}
	return nil
}

type CacheStats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	Entries   int
	Bytes     int64
}

type CacheInspector interface {
	CacheStats() CacheStats
}
