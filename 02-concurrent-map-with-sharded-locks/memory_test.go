package main

import (
	"fmt"
	"runtime"
	"testing"
)

const (
	oneMillion         = 1_000_000
	maxExtraMemoryMB   = 50
	maxExtraMemoryByte = maxExtraMemoryMB * 1024 * 1024
)

var (
	baselineSink   map[int]interface{}
	concurrentSink ConcurrentMaps[int, interface{}]
)

func heapAllocAfterGC() uint64 {
	runtime.GC()
	runtime.GC()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapAlloc
}

func TestMemoryOverheadVsBaselineMap(t *testing.T) {
	baselineSink = nil
	concurrentSink = ConcurrentMaps[int, interface{}]{}

	beforeBaseline := heapAllocAfterGC()
	baselineSink = make(map[int]interface{}, oneMillion)
	for i := 0; i < oneMillion; i++ {
		baselineSink[i] = i
	}
	afterBaseline := heapAllocAfterGC()
	baselineDelta := afterBaseline - beforeBaseline

	beforeConcurrent := heapAllocAfterGC()
	concurrentSink = NewConcurrentMaps[int, interface{}](64)
	for i := 0; i < oneMillion; i++ {
		concurrentSink.Set(i, i)
	}
	afterConcurrent := heapAllocAfterGC()
	concurrentDelta := afterConcurrent - beforeConcurrent

	extra := concurrentDelta - baselineDelta
	if concurrentDelta < baselineDelta {
		extra = 0
	}

	t.Logf("baseline retained heap:   %.2f MB", float64(baselineDelta)/(1024*1024))
	t.Logf("concurrent retained heap: %.2f MB", float64(concurrentDelta)/(1024*1024))
	t.Logf("extra over baseline:      %.2f MB", float64(extra)/(1024*1024))

	if extra > maxExtraMemoryByte {
		t.Fatalf("memory overhead too high: extra=%s, limit=%dMB", formatBytes(extra), maxExtraMemoryMB)
	}
}

func formatBytes(b uint64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	if b < kb {
		return fmt.Sprintf("%d B", b)
	}
	if b < mb {
		return fmt.Sprintf("%.2f KB", float64(b)/kb)
	}
	return fmt.Sprintf("%.2f MB", float64(b)/mb)
}
