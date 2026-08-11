package contextengine

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"sync"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

type cacheEntry struct {
	value  any
	size   int64
	access uint64
}

type artifactCache struct {
	mu         sync.Mutex
	policy     pkgContext.CachePolicy
	entries    map[string]cacheEntry
	dimensions map[string]int
	sequence   uint64
	hits       uint64
	misses     uint64
	evictions  uint64
	bytes      int64
}

func newArtifactCache(policy pkgContext.CachePolicy) *artifactCache {
	return &artifactCache{
		policy: policy, entries: make(map[string]cacheEntry), dimensions: make(map[string]int),
	}
}

func (cache *artifactCache) get(key string) (any, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	entry, found := cache.entries[key]
	if !found {
		cache.misses++
		return nil, false
	}
	cache.hits++
	cache.sequence++
	entry.access = cache.sequence
	cache.entries[key] = entry
	return cloneCacheValue(entry.value), true
}

func (cache *artifactCache) put(key string, value any, size int64) {
	if size <= 0 || size > cache.policy.MaxBytes {
		return
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if previous, exists := cache.entries[key]; exists {
		cache.bytes -= previous.size
	}
	cache.sequence++
	cache.entries[key] = cacheEntry{value: cloneCacheValue(value), size: size, access: cache.sequence}
	cache.bytes += size
	cache.evictLocked()
}

func (cache *artifactCache) evictLocked() {
	for len(cache.entries) > cache.policy.MaxEntries || cache.bytes > cache.policy.MaxBytes {
		var victim string
		var oldest uint64
		first := true
		for key, entry := range cache.entries {
			if first || entry.access < oldest || (entry.access == oldest && key < victim) {
				victim, oldest, first = key, entry.access, false
			}
		}
		entry := cache.entries[victim]
		delete(cache.entries, victim)
		cache.bytes -= entry.size
		cache.evictions++
	}
}

func (cache *artifactCache) stats() pkgContext.CacheStats {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return pkgContext.CacheStats{
		Hits: cache.hits, Misses: cache.misses, Evictions: cache.evictions,
		Entries: len(cache.entries), Bytes: cache.bytes,
	}
}

func (cache *artifactCache) recordMisses(count int) {
	cache.mu.Lock()
	cache.misses += uint64(count)
	cache.mu.Unlock()
}

func (cache *artifactCache) embeddingDimension(target string) (int, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	dimension, found := cache.dimensions[target]
	return dimension, found
}

func (cache *artifactCache) setEmbeddingDimension(target string, dimension int) (changed bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	previous, found := cache.dimensions[target]
	if found && previous != dimension {
		prefix := "embedding\x00" + target + "\x00"
		for key, entry := range cache.entries {
			if strings.HasPrefix(key, prefix) {
				delete(cache.entries, key)
				cache.bytes -= entry.size
				cache.evictions++
			}
		}
		changed = true
	}
	cache.dimensions[target] = dimension
	return changed
}

func cloneCacheValue(value any) any {
	switch typed := value.(type) {
	case []float32:
		return slices.Clone(typed)
	default:
		return typed
	}
}

func analysisCacheKey(document pkgContext.Document, analyzer pkgContext.Analyzer) string {
	return "analysis\x00" + string(document.Digest()) + "\x00" + string(document.Path()) + "\x00" +
		document.MediaType() + "\x00" + string(document.Language()) + "\x00" +
		string(analyzer.ID()) + "\x00" + analyzer.Version()
}

func embeddingTarget(provider, model string) string { return provider + "\x00" + model }

func embeddingCacheKey(target string, dimension int, text string) string {
	return fmt.Sprintf("embedding\x00%s\x00%d\x00%x", target, dimension, sha256.Sum256([]byte(text)))
}

func estimatorCacheKey(estimator pkgContext.TokenEstimator, text string) string {
	return fmt.Sprintf("estimator\x00%s\x00%s\x00%x", estimator.ID(), estimator.Version(), sha256.Sum256([]byte(text)))
}

func analysisSize(analysis pkgContext.Analysis) int64 {
	size := int64(128 + len(analysis.Path()) + len(analysis.Digest()) + len(analysis.Analyzer()) + len(analysis.Version()))
	for _, symbol := range analysis.Symbols() {
		size += int64(64 + len(symbol.ID) + len(symbol.Name) + len(symbol.Container))
	}
	for _, relation := range analysis.Relations() {
		size += int64(32 + len(relation.From) + len(relation.To))
	}
	size += int64(len(analysis.Chunks())*48 + len(analysis.Diagnostics())*64)
	return size
}
