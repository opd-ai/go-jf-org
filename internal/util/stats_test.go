package util

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestStatistics_Counters(t *testing.T) {
	stats := NewStatistics()

	stats.Increment("files")
	stats.Increment("files")
	stats.Add("files", 3)

	if got := stats.Get("files"); got != 5 {
		t.Errorf("expected files=5, got %d", got)
	}
}

func TestStatistics_Sizes(t *testing.T) {
	stats := NewStatistics()

	stats.AddSize("total", 1024)
	stats.AddSize("total", 2048)

	if got := stats.GetSize("total"); got != 3072 {
		t.Errorf("expected total=3072, got %d", got)
	}
}

func TestStatistics_Timings(t *testing.T) {
	stats := NewStatistics()

	stats.AddTiming("operation", 100*time.Millisecond)
	stats.AddTiming("operation", 200*time.Millisecond)

	expected := 300 * time.Millisecond
	if got := stats.GetTiming("operation"); got != expected {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestStatistics_Finish(t *testing.T) {
	stats := NewStatistics()
	time.Sleep(10 * time.Millisecond)
	stats.Finish()

	if stats.Duration < 10*time.Millisecond {
		t.Error("expected duration >= 10ms")
	}

	if stats.EndTime.IsZero() {
		t.Error("expected EndTime to be set")
	}
}

func TestStatistics_Timer(t *testing.T) {
	stats := NewStatistics()

	timer := stats.NewTimer("test_operation")
	time.Sleep(50 * time.Millisecond)
	timer.Stop()

	timing := stats.GetTiming("test_operation")
	if timing < 50*time.Millisecond {
		t.Errorf("expected timing >= 50ms, got %v", timing)
	}
}

func TestStatistics_ToJSON(t *testing.T) {
	stats := NewStatistics()
	stats.Increment("files")
	stats.AddSize("total", 1024)
	stats.AddTiming("operation", 100*time.Millisecond)
	stats.Finish()

	jsonStr, err := stats.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}

	// Check for expected fields
	if _, ok := data["counters"]; !ok {
		t.Error("expected 'counters' in JSON output")
	}
	if _, ok := data["sizes_bytes"]; !ok {
		t.Error("expected 'sizes_bytes' in JSON output")
	}
	if _, ok := data["timings_ms"]; !ok {
		t.Error("expected 'timings_ms' in JSON output")
	}
}

func TestStatistics_Summary(t *testing.T) {
	stats := NewStatistics()
	stats.Increment("files_processed")
	stats.Add("files_processed", 9) // Total 10
	stats.AddSize("total_bytes", 1024*1024)
	stats.AddTiming("scan", 100*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	stats.Finish()

	summary := stats.Summary()

	// Check that summary contains expected sections
	expectedParts := []string{
		"Statistics Summary",
		"Duration:",
		"Counters:",
		"files_processed: 10",
		"Data Processed:",
		"Timings:",
		"Throughput:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(summary, part) {
			t.Errorf("expected summary to contain %q", part)
		}
	}
}

func TestStatistics_Concurrent(t *testing.T) {
	stats := NewStatistics()
	var wg sync.WaitGroup

	// Concurrent increments
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				stats.Increment("files")
				stats.AddSize("bytes", 1024)
			}
		}()
	}

	wg.Wait()

	if got := stats.Get("files"); got != 1000 {
		t.Errorf("expected files=1000, got %d", got)
	}

	if got := stats.GetSize("bytes"); got != 1024*1000 {
		t.Errorf("expected bytes=%d, got %d", 1024*1000, got)
	}
}

func TestOperationStats_Basic(t *testing.T) {
	ops := NewOperationStats("Test Operation", 100)

	ops.IncrementCompleted()
	ops.IncrementCompleted()
	ops.IncrementFailed()
	ops.IncrementSkipped()
	ops.AddBytes(1024)

	if ops.Completed != 2 {
		t.Errorf("expected Completed=2, got %d", ops.Completed)
	}
	if ops.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", ops.Failed)
	}
	if ops.Skipped != 1 {
		t.Errorf("expected Skipped=1, got %d", ops.Skipped)
	}
	if ops.BytesProcessed != 1024 {
		t.Errorf("expected BytesProcessed=1024, got %d", ops.BytesProcessed)
	}
}

func TestOperationStats_Duration(t *testing.T) {
	ops := NewOperationStats("Test", 10)
	time.Sleep(50 * time.Millisecond)

	// Before finish, duration should be time since start
	duration := ops.Duration()
	if duration < 50*time.Millisecond {
		t.Errorf("expected duration >= 50ms, got %v", duration)
	}

	ops.Finish()

	// After finish, duration should be fixed
	duration1 := ops.Duration()
	time.Sleep(10 * time.Millisecond)
	duration2 := ops.Duration()

	if duration1 != duration2 {
		t.Error("expected duration to be fixed after Finish()")
	}
}

func TestOperationStats_Summary(t *testing.T) {
	ops := NewOperationStats("File Processing", 100)
	ops.Completed = 90
	ops.Failed = 5
	ops.Skipped = 5
	ops.BytesProcessed = 1024 * 1024 * 10 // 10 MB
	time.Sleep(10 * time.Millisecond)
	ops.Finish()

	summary := ops.Summary()

	expectedParts := []string{
		"File Processing Statistics:",
		"Total: 100",
		"Completed: 90",
		"Failed: 5",
		"Skipped: 5",
		"Duration:",
		"Data Processed:",
		"Rate:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(summary, part) {
			t.Errorf("expected summary to contain %q, got:\n%s", part, summary)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			bytes:    1024,
			expected: "1.00 KB",
		},
		{
			name:     "megabytes",
			bytes:    1024 * 1024,
			expected: "1.00 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1024 * 1024 * 1024,
			expected: "1.00 GB",
		},
		{
			name:     "terabytes",
			bytes:    1024 * 1024 * 1024 * 1024,
			expected: "1.00 TB",
		},
		{
			name:     "mixed MB",
			bytes:    1536 * 1024, // 1.5 MB
			expected: "1.50 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStatistics_MultipleTimers(t *testing.T) {
	stats := NewStatistics()

	// Start multiple timers
	timer1 := stats.NewTimer("operation1")
	time.Sleep(30 * time.Millisecond)
	timer1.Stop()

	timer2 := stats.NewTimer("operation2")
	time.Sleep(20 * time.Millisecond)
	timer2.Stop()

	// Verify both timings are recorded
	timing1 := stats.GetTiming("operation1")
	if timing1 < 30*time.Millisecond {
		t.Errorf("expected operation1 >= 30ms, got %v", timing1)
	}

	timing2 := stats.GetTiming("operation2")
	if timing2 < 20*time.Millisecond {
		t.Errorf("expected operation2 >= 20ms, got %v", timing2)
	}
}

func BenchmarkStatistics_Increment(b *testing.B) {
	stats := NewStatistics()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stats.Increment("counter")
	}
}

func BenchmarkStatistics_AddSize(b *testing.B) {
	stats := NewStatistics()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stats.AddSize("bytes", 1024)
	}
}

func BenchmarkStatistics_Timer(b *testing.B) {
	stats := NewStatistics()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timer := stats.NewTimer("operation")
		timer.Stop()
	}
}
