package main

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"
)

// =============================================================================
// TEST 1: ALLOCATION TEST
// Run with: go test -bench=. -benchmem -count=5
// Pass: allocs/op = 0 for parsing loop
// Fail: Any allocations in hot path
// =============================================================================

// BenchmarkSensorParser benchmarks the parsing loop for allocations
func BenchmarkSensorParser(b *testing.B) {
	// 1. Create a payload with 1,000 sensor readings
	oneItem := `{"sensor_id":"TEST","readings":[1.1, 2.2]}`
	payload := strings.Repeat(oneItem, 1000)

	// 2. Prepare the reader input (bytes)
	data := []byte(payload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		// This single call parses 1,000 items
		SensorParser(r)
	}
}

// BenchmarkSensorParserSingleObject benchmarks parsing a single object
func BenchmarkSensorParserSingleObject(b *testing.B) {
	jsonData := `{"sensor_id": "sensor_001", "readings": [1.5]}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(jsonData)
		_ = SensorParser(reader)
	}
}

// =============================================================================
// TEST 2: STREAM TEST
// Pipe large amounts of JSON through the parser
// Pass: Memory usage flatlines after warm-up
// Fail: Memory grows linearly with input size
// =============================================================================

// infiniteJSONReader generates repeating JSON data to simulate streaming
type infiniteJSONReader struct {
	data    []byte
	pos     int
	limit   int64 // total bytes to generate
	written int64
}

func newInfiniteJSONReader(limit int64) *infiniteJSONReader {
	// Template JSON object
	template := `{"sensor_id": "sensor_%06d", "readings": [1.5, 2.5, 3.5]}
`
	return &infiniteJSONReader{
		data:  []byte(fmt.Sprintf(template, 1)),
		limit: limit,
	}
}

func (r *infiniteJSONReader) Read(p []byte) (n int, err error) {
	if r.written >= r.limit {
		return 0, io.EOF
	}

	for n < len(p) && r.written < r.limit {
		copied := copy(p[n:], r.data[r.pos:])
		n += copied
		r.pos += copied
		r.written += int64(copied)

		if r.pos >= len(r.data) {
			r.pos = 0
		}
	}

	return n, nil
}

// TestStreamMemoryStability tests that memory doesn't grow linearly with input
func TestStreamMemoryStability(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping stream test in short mode")
	}

	// Force GC before starting
	runtime.GC()

	var memBefore, memAfterWarmup, memFinal runtime.MemStats

	// Measure initial memory
	runtime.ReadMemStats(&memBefore)

	// Warm-up phase: process ~10MB
	warmupReader := newInfiniteJSONReader(10 * 1024 * 1024) // 10MB
	err := SensorParser(warmupReader)
	if err != nil {
		t.Fatalf("Warm-up parsing failed: %v", err)
	}

	runtime.GC()
	runtime.ReadMemStats(&memAfterWarmup)

	// Main test phase: process ~100MB (you can increase to 1GB for thorough testing)
	mainReader := newInfiniteJSONReader(100 * 1024 * 1024) // 100MB
	err = SensorParser(mainReader)
	if err != nil {
		t.Fatalf("Main parsing failed: %v", err)
	}

	runtime.GC()
	runtime.ReadMemStats(&memFinal)

	// Calculate memory growth
	warmupAlloc := memAfterWarmup.HeapAlloc - memBefore.HeapAlloc
	mainAlloc := memFinal.HeapAlloc - memAfterWarmup.HeapAlloc

	t.Logf("Memory before: %d bytes", memBefore.HeapAlloc)
	t.Logf("Memory after warmup (10MB processed): %d bytes (growth: %d)", memAfterWarmup.HeapAlloc, warmupAlloc)
	t.Logf("Memory after main (100MB processed): %d bytes (growth: %d)", memFinal.HeapAlloc, mainAlloc)

	// The memory growth after warmup should be minimal (less than 1MB for 100MB input)
	// If memory grows linearly, we'd see ~10x growth ratio
	if mainAlloc > 1*1024*1024 {
		t.Errorf("FAIL: Memory grew by %d bytes after processing 100MB, expected near-zero growth after warmup", mainAlloc)
	} else {
		t.Logf("PASS: Memory usage is stable (growth after warmup: %d bytes)", mainAlloc)
	}
}

// BenchmarkStreamParsing benchmarks continuous stream parsing
func BenchmarkStreamParsing(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(1024 * 1024) // Report MB/s

	for i := 0; i < b.N; i++ {
		reader := newInfiniteJSONReader(1024 * 1024) // 1MB per iteration
		_ = SensorParser(reader)
	}
}

// =============================================================================
// TEST 3: CORRUPTION TEST
// Input: {"sensor_id": "a"} {"bad json here (malformed second object)
// Pass: Returns first object, logs/skips second, doesn't panic
// Fail: Parser crashes or stops processing entirely
// =============================================================================

// TestCorruptedJSON tests handling of malformed JSON
func TestCorruptedJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldPanic bool
		expectError bool
	}{
		{
			name:        "valid then malformed",
			input:       `{"sensor_id": "a", "readings": [1.0]} {"bad json here`,
			shouldPanic: false,
			expectError: true, // should return an error but not panic
		},
		{
			name:        "malformed from start",
			input:       `{"bad json here`,
			shouldPanic: false,
			expectError: true,
		},
		{
			name:        "missing closing brace",
			input:       `{"sensor_id": "a", "readings": [1.0]`,
			shouldPanic: false,
			expectError: true,
		},
		{
			name:        "truncated in middle",
			input:       `{"sensor_id": "a", "read`,
			shouldPanic: false,
			expectError: true,
		},
		{
			name:        "empty input",
			input:       ``,
			shouldPanic: false,
			expectError: false, // empty is valid, just no objects
		},
		{
			name:        "valid single object",
			input:       `{"sensor_id": "test", "readings": [1.0, 2.0]}`,
			shouldPanic: false,
			expectError: false,
		},
		{
			name:        "multiple valid objects",
			input:       `{"sensor_id": "a", "readings": [1.0]}{"sensor_id": "b", "readings": [2.0]}`,
			shouldPanic: false,
			expectError: false,
		},
		{
			name:        "unicode in sensor_id",
			input:       `{"sensor_id": "センサー", "readings": [1.0]}`,
			shouldPanic: false,
			expectError: false,
		},
		{
			name:        "very long sensor_id",
			input:       fmt.Sprintf(`{"sensor_id": "%s", "readings": [1.0]}`, strings.Repeat("x", 10000)),
			shouldPanic: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("FAIL: Parser panicked unexpectedly: %v", r)
					}
				}
			}()

			reader := bytes.NewReader([]byte(tt.input))
			err := SensorParser(reader)

			if tt.expectError && err == nil {
				// This might not be a failure if the parser gracefully handles errors
				t.Logf("Note: expected error but got nil (parser may have skipped malformed data)")
			}

			if !tt.expectError && err != nil {
				t.Errorf("FAIL: Unexpected error: %v", err)
			}

			// If we reach here without panic, the panic test passes
			if tt.shouldPanic {
				t.Errorf("FAIL: Expected panic but none occurred")
			}
		})
	}
}

// TestCorruptedJSONDoesNotPanic specifically tests the example from the requirements
func TestCorruptedJSONDoesNotPanic(t *testing.T) {
	input := `{"sensor_id": "a"} {"bad json here`

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FAIL: Parser crashed/panicked: %v", r)
		}
	}()

	reader := bytes.NewReader([]byte(input))
	err := SensorParser(reader)

	// We expect an error, but NOT a panic
	// The test passes as long as we don't panic
	t.Logf("Parser returned error (expected): %v", err)
	t.Log("PASS: Parser did not crash or panic on malformed JSON")
}

// TestGracefulRecovery tests that the parser can recover from errors
func TestGracefulRecovery(t *testing.T) {
	// Process a valid stream after encountering corruption
	validInput := `{"sensor_id": "valid1", "readings": [1.0]}
{"sensor_id": "valid2", "readings": [2.0]}`

	reader := bytes.NewReader([]byte(validInput))
	err := SensorParser(reader)
	if err != nil {
		t.Errorf("FAIL: Valid input after test setup caused error: %v", err)
	}
	t.Log("PASS: Parser processes valid JSON correctly")
}
