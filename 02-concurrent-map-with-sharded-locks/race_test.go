package main

import (
	"sync"
	"testing"
)

func TestConcurrentReadWriteDelete_NoRace(t *testing.T) {
	for _, shards := range []int{1, 64} {
		t.Run("shards="+itoa(shards), func(t *testing.T) {
			cm := NewConcurrentMaps[int, int](shards)

			workers := 12
			opsPerWorker := 50_000

			var wg sync.WaitGroup
			wg.Add(workers)

			for w := 0; w < workers; w++ {
				workerID := w
				go func() {
					defer wg.Done()

					start := workerID * opsPerWorker
					for i := 0; i < opsPerWorker; i++ {
						k := start + i

						cm.Set(k, i)
						_, _ = cm.Get(k)

						if i%3 == 0 {
							cm.Delete(k)
						}
					}
				}()
			}

			wg.Wait()

			// Touch keys after the concurrent phase as a sanity pass.
			_ = len(cm.Keys())
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
