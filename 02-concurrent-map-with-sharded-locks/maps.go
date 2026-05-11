package main

import (
	"hash/fnv"
	"sync"
	"unsafe"
)

// ConcurrentMapShard represents a single shard of the concurrent map, containing a standard map and a mutex for synchronization.
type ConcurrentMapShard[K comparable, V any] struct {
	item map[K]V
	mu   sync.RWMutex
}

// ConcurrentMap is a thread-safe map that uses sharding to reduce lock contention.
// It consists of multiple shards, each responsible for a portion of the key space, allowing for concurrent access to different shards without blocking each other.
type ConcurrentMaps[K comparable, V any] struct {
	shards []*ConcurrentMapShard[K, V]
}

// This creates shards of the concurrent map. As many as desired.
func NewConcurrentMaps[K comparable, V any](numShards int) ConcurrentMaps[K, V] {
	m := ConcurrentMaps[K, V]{
		shards: make([]*ConcurrentMapShard[K, V], numShards),
	}
	for i := 0; i < numShards; i++ {
		m.shards[i] = &ConcurrentMapShard[K, V]{
			item: make(map[K]V),
		}
	}
	return m
}

func (cm *ConcurrentMaps[K, V]) getSharedHashFunction(key K) *ConcurrentMapShard[K, V] {
	h := fnv.New64a()
	h.Write(unsafe.Slice((*byte)(unsafe.Pointer(&key)), unsafe.Sizeof(key)))
	i := h.Sum64() % uint64(len(cm.shards))
	return cm.shards[i]
}

func (cm *ConcurrentMaps[K, V]) Get(key K) (V, bool) {
	shard := cm.getSharedHashFunction(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	v, ok := shard.item[key]
	return v, ok
}

func (cm *ConcurrentMaps[K, V]) Set(key K, value V) {
	shard := cm.getSharedHashFunction(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.item[key] = value
}

func (cm *ConcurrentMaps[K, V]) Delete(Key K) {
	shard := cm.getSharedHashFunction(Key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.item, Key)
}

func (cm *ConcurrentMaps[K, V]) Keys() []K {
	keys := make([]K, 0)
	for _, shard := range cm.shards {
		shard.mu.RLock()
		for k := range shard.item {
			keys = append(keys, k)
		}
		shard.mu.RUnlock()
	}
	return keys
}
