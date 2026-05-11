package main

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func benchmarkSetContention(b *testing.B, shards int) {
	cm := NewConcurrentMaps[int, int](shards)
	workers := 10

	// Keep worker count fixed for this contention comparison.
	runtime.GOMAXPROCS(workers)

	b.ReportAllocs()
	b.ResetTimer()

	opsPerWorker := b.N / workers
	extra := b.N % workers

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		workerID := w
		go func() {
			defer wg.Done()

			n := opsPerWorker
			if workerID < extra {
				n++
			}

			start := workerID * (opsPerWorker + 1)
			for i := 0; i < n; i++ {
				// Sequential key stream per worker.
				k := start + i
				cm.Set(k, k)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkSetContention(b *testing.B) {
	for _, shards := range []int{1, 64} {
		b.Run(fmt.Sprintf("shards=%d", shards), func(b *testing.B) {
			benchmarkSetContention(b, shards)
		})
	}
}
