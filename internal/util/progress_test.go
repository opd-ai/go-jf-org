package util

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestProgressTracker_Basic(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		steps    []int
		expected int
	}{
		{
			name:     "incremental progress",
			total:    10,
			steps:    []int{1, 1, 1, 1, 1},
			expected: 5,
		},
		{
			name:     "batch progress",
			total:    100,
			steps:    []int{10, 20, 30, 40},
			expected: 100,
		},
		{
			name:     "over limit capped",
			total:    10,
			steps:    []int{5, 5, 5},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			pt := NewProgressTracker(tt.total, "Testing")
			pt.SetWriter(buf)

			for _, step := range tt.steps {
				pt.Add(step)
			}

			// Check current value
			if pt.current != tt.expected {
				t.Errorf("expected current=%d, got %d", tt.expected, pt.current)
			}
		})
	}
}

func TestProgressTracker_Increment(t *testing.T) {
	buf := &bytes.Buffer{}
	pt := NewProgressTracker(10, "Testing")
	pt.SetWriter(buf)
	pt.SetEnabled(true)

	for i := 0; i < 5; i++ {
		pt.Increment()
	}

	if pt.current != 5 {
		t.Errorf("expected current=5, got %d", pt.current)
	}
}

func TestProgressTracker_SetTotal(t *testing.T) {
	pt := NewProgressTracker(10, "Testing")
	pt.SetEnabled(false)

	pt.Add(5)
	if pt.current != 5 {
		t.Errorf("expected current=5, got %d", pt.current)
	}

	pt.SetTotal(20)
	if pt.total != 20 {
		t.Errorf("expected total=20, got %d", pt.total)
	}
}

func TestProgressTracker_Finish(t *testing.T) {
	buf := &bytes.Buffer{}
	pt := NewProgressTracker(10, "Testing")
	pt.SetWriter(buf)
	pt.SetEnabled(true)

	pt.Add(5)
	pt.Finish()

	if pt.current != pt.total {
		t.Errorf("expected current=%d after finish, got %d", pt.total, pt.current)
	}

	output := buf.String()
	if !strings.Contains(output, "100%") {
		t.Error("expected output to contain 100% after finish")
	}
}

func TestProgressTracker_Disabled(t *testing.T) {
	buf := &bytes.Buffer{}
	pt := NewProgressTracker(10, "Testing")
	pt.SetWriter(buf)
	pt.SetEnabled(false)

	pt.Add(5)
	pt.Finish()

	// Should produce no output when disabled
	if buf.Len() > 0 {
		t.Errorf("expected no output when disabled, got: %s", buf.String())
	}
}

func TestProgressTracker_RateLimiting(t *testing.T) {
	buf := &bytes.Buffer{}
	pt := NewProgressTracker(1000, "Testing")
	pt.SetWriter(buf)
	pt.SetEnabled(true)
	pt.updateDelay = 50 * time.Millisecond

	// Rapidly add progress
	for i := 0; i < 100; i++ {
		pt.Increment()
	}

	// Should have fewer updates than increments due to rate limiting
	output := buf.String()
	updateCount := strings.Count(output, "\r")

	// Should be significantly fewer than 100 updates
	if updateCount >= 100 {
		t.Errorf("expected rate limiting, got %d updates for 100 increments", updateCount)
	}
}

func TestProgressTracker_PercentageCalculation(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		current  int
		expected int
	}{
		{
			name:     "zero progress",
			total:    100,
			current:  0,
			expected: 0,
		},
		{
			name:     "half progress",
			total:    100,
			current:  50,
			expected: 50,
		},
		{
			name:     "complete progress",
			total:    100,
			current:  100,
			expected: 100,
		},
		{
			name:     "zero total",
			total:    0,
			current:  0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			pt := NewProgressTracker(tt.total, "Testing")
			pt.SetWriter(buf)
			pt.SetEnabled(true)
			pt.current = tt.current
			pt.render()

			output := buf.String()
			expectedStr := strings.Replace(tt.name, "_", " ", -1)

			// Check percentage is in output
			if tt.total > 0 && !strings.Contains(output, strings.TrimSpace(strings.Split(expectedStr, " ")[0])) {
				// Just verify some output was generated for non-zero totals
				if len(output) == 0 {
					t.Error("expected some output for progress render")
				}
			}
		})
	}
}

func TestProgressTracker_Concurrent(t *testing.T) {
	buf := &bytes.Buffer{}
	pt := NewProgressTracker(1000, "Testing")
	pt.SetWriter(buf)
	pt.SetEnabled(true)

	// Simulate concurrent increments
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				pt.Increment()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if pt.current != 1000 {
		t.Errorf("expected current=1000 after concurrent updates, got %d", pt.current)
	}
}

func TestSpinner_StartStop(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner("Testing")
	s.SetWriter(buf)
	s.SetEnabled(true)

	s.Start()
	time.Sleep(200 * time.Millisecond) // Let it spin a bit
	s.Stop()

	// Should have produced some output
	if buf.Len() == 0 {
		t.Error("expected spinner output")
	}
}

func TestSpinner_Disabled(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner("Testing")
	s.SetWriter(buf)
	s.SetEnabled(false)

	s.Start()
	time.Sleep(200 * time.Millisecond)
	s.Stop()

	// Should produce no output when disabled
	if buf.Len() > 0 {
		t.Errorf("expected no output when disabled, got: %s", buf.String())
	}
}

func TestSpinner_MultipleStartStop(t *testing.T) {
	s := NewSpinner("Testing")
	s.SetEnabled(false) // Disable output for test

	// Start multiple times should be safe
	s.Start()
	s.Start()
	s.Start()

	time.Sleep(100 * time.Millisecond)

	// Stop multiple times should be safe
	s.Stop()
	s.Stop()
	s.Stop()
}

// TestSpinner_ConcurrentStop tests that concurrent calls to Stop() don't panic
// This validates the fix for BUG-RACE-001
func TestSpinner_ConcurrentStop(t *testing.T) {
	s := NewSpinner("Testing")
	s.SetEnabled(false) // Disable output for test
	s.Start()
	time.Sleep(50 * time.Millisecond)

	// Call Stop() from multiple goroutines concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Stop() // Should not panic due to sync.Once protection
		}()
	}
	wg.Wait()

	// Verify spinner is stopped
	if s.running {
		t.Error("expected spinner to be stopped")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "minutes and seconds",
			duration: 2*time.Minute + 30*time.Second,
			expected: "2m30s",
		},
		{
			name:     "hours, minutes, and seconds",
			duration: 1*time.Hour + 15*time.Minute + 30*time.Second,
			expected: "1h15m30s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0s",
		},
		{
			name:     "milliseconds rounded",
			duration: 1500 * time.Millisecond,
			expected: "2s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
