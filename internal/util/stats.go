package util

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Statistics tracks operation statistics and metrics
type Statistics struct {
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration_ms"`
	Counters  map[string]int         `json:"counters"`
	Sizes     map[string]int64       `json:"sizes_bytes"`
	Timings   map[string]time.Duration `json:"timings_ms"`
	mu        sync.RWMutex
}

// NewStatistics creates a new statistics tracker
func NewStatistics() *Statistics {
	return &Statistics{
		StartTime: time.Now(),
		Counters:  make(map[string]int),
		Sizes:     make(map[string]int64),
		Timings:   make(map[string]time.Duration),
	}
}

// Increment increments a counter by 1
func (s *Statistics) Increment(name string) {
	s.Add(name, 1)
}

// Add adds a value to a counter
func (s *Statistics) Add(name string, value int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Counters[name] += value
}

// Get returns the value of a counter
func (s *Statistics) Get(name string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Counters[name]
}

// AddSize adds to a size counter (in bytes)
func (s *Statistics) AddSize(name string, bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sizes[name] += bytes
}

// GetSize returns the value of a size counter
func (s *Statistics) GetSize(name string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Sizes[name]
}

// AddTiming adds a timing measurement
func (s *Statistics) AddTiming(name string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Timings[name] += duration
}

// GetTiming returns the value of a timing
func (s *Statistics) GetTiming(name string) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Timings[name]
}

// Finish marks the statistics collection as complete
func (s *Statistics) Finish() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.EndTime = time.Now()
	s.Duration = s.EndTime.Sub(s.StartTime)
}

// ToJSON converts statistics to JSON format
func (s *Statistics) ToJSON() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert to JSON-friendly format
	data := struct {
		StartTime string                    `json:"start_time"`
		EndTime   string                    `json:"end_time"`
		Duration  int64                     `json:"duration_ms"`
		Counters  map[string]int            `json:"counters"`
		Sizes     map[string]int64          `json:"sizes_bytes"`
		Timings   map[string]int64          `json:"timings_ms"`
	}{
		StartTime: s.StartTime.Format(time.RFC3339),
		EndTime:   s.EndTime.Format(time.RFC3339),
		Duration:  s.Duration.Milliseconds(),
		Counters:  s.Counters,
		Sizes:     s.Sizes,
		Timings:   make(map[string]int64),
	}

	for k, v := range s.Timings {
		data.Timings[k] = v.Milliseconds()
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Summary returns a human-readable summary
func (s *Statistics) Summary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := "Statistics Summary\n"
	summary += "==================\n\n"

	// Duration
	summary += fmt.Sprintf("Duration: %s\n\n", FormatDuration(s.Duration))

	// Counters
	if len(s.Counters) > 0 {
		summary += "Counters:\n"
		for name, value := range s.Counters {
			summary += fmt.Sprintf("  %s: %d\n", name, value)
		}
		summary += "\n"
	}

	// Sizes
	if len(s.Sizes) > 0 {
		summary += "Data Processed:\n"
		for name, bytes := range s.Sizes {
			summary += fmt.Sprintf("  %s: %s\n", name, FormatBytes(bytes))
		}
		summary += "\n"
	}

	// Timings
	if len(s.Timings) > 0 {
		summary += "Timings:\n"
		for name, duration := range s.Timings {
			summary += fmt.Sprintf("  %s: %s\n", name, FormatDuration(duration))
		}
		summary += "\n"
	}

	// Throughput calculations
	if s.Duration > 0 {
		filesProcessed := s.Counters["files_processed"]
		if filesProcessed > 0 {
			rate := float64(filesProcessed) / s.Duration.Seconds()
			summary += fmt.Sprintf("Throughput: %.2f files/second\n", rate)
		}

		bytesProcessed := s.Sizes["total_bytes"]
		if bytesProcessed > 0 {
			rate := float64(bytesProcessed) / s.Duration.Seconds()
			summary += fmt.Sprintf("Data Rate: %s/second\n", FormatBytes(int64(rate)))
		}
	}

	return summary
}

// Timer provides a convenient way to track operation timing
type Timer struct {
	stats *Statistics
	name  string
	start time.Time
}

// NewTimer creates a new timer for the given statistics tracker
func (s *Statistics) NewTimer(name string) *Timer {
	return &Timer{
		stats: s,
		name:  name,
		start: time.Now(),
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() {
	duration := time.Since(t.start)
	t.stats.AddTiming(t.name, duration)
}

// OperationStats tracks statistics for a specific operation
type OperationStats struct {
	Name      string
	Total     int
	Completed int
	Failed    int
	Skipped   int
	BytesProcessed int64
	StartTime time.Time
	EndTime   time.Time
	mu        sync.RWMutex
}

// NewOperationStats creates a new operation statistics tracker
func NewOperationStats(name string, total int) *OperationStats {
	return &OperationStats{
		Name:      name,
		Total:     total,
		StartTime: time.Now(),
	}
}

// IncrementCompleted increments the completed counter
func (os *OperationStats) IncrementCompleted() {
	os.mu.Lock()
	defer os.mu.Unlock()
	os.Completed++
}

// IncrementFailed increments the failed counter
func (os *OperationStats) IncrementFailed() {
	os.mu.Lock()
	defer os.mu.Unlock()
	os.Failed++
}

// IncrementSkipped increments the skipped counter
func (os *OperationStats) IncrementSkipped() {
	os.mu.Lock()
	defer os.mu.Unlock()
	os.Skipped++
}

// AddBytes adds to the bytes processed counter
func (os *OperationStats) AddBytes(bytes int64) {
	os.mu.Lock()
	defer os.mu.Unlock()
	os.BytesProcessed += bytes
}

// Finish marks the operation as complete
func (os *OperationStats) Finish() {
	os.mu.Lock()
	defer os.mu.Unlock()
	os.EndTime = time.Now()
}

// Duration returns the duration of the operation
func (os *OperationStats) Duration() time.Duration {
	os.mu.RLock()
	defer os.mu.RUnlock()
	if os.EndTime.IsZero() {
		return time.Since(os.StartTime)
	}
	return os.EndTime.Sub(os.StartTime)
}

// Summary returns a human-readable summary
func (os *OperationStats) Summary() string {
	os.mu.RLock()
	defer os.mu.RUnlock()
	
	// Calculate duration inline to avoid nested locking
	var duration time.Duration
	if os.EndTime.IsZero() {
		duration = time.Since(os.StartTime)
	} else {
		duration = os.EndTime.Sub(os.StartTime)
	}
	
	summary := fmt.Sprintf("%s Statistics:\n", os.Name)
	summary += fmt.Sprintf("  Total: %d\n", os.Total)
	summary += fmt.Sprintf("  Completed: %d\n", os.Completed)
	if os.Failed > 0 {
		summary += fmt.Sprintf("  Failed: %d\n", os.Failed)
	}
	if os.Skipped > 0 {
		summary += fmt.Sprintf("  Skipped: %d\n", os.Skipped)
	}
	summary += fmt.Sprintf("  Duration: %s\n", FormatDuration(duration))
	
	if os.BytesProcessed > 0 {
		summary += fmt.Sprintf("  Data Processed: %s\n", FormatBytes(os.BytesProcessed))
	}
	
	if duration > 0 && os.Completed > 0 {
		rate := float64(os.Completed) / duration.Seconds()
		summary += fmt.Sprintf("  Rate: %.2f items/sec\n", rate)
	}
	
	return summary
}

// FormatBytes formats bytes in human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	// Extended to support larger units (EB, ZB, YB)
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	// Bounds check to prevent panic - caps at YB (yottabyte)
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}
