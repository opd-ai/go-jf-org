package util

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressTracker tracks progress of operations and displays real-time updates
type ProgressTracker struct {
	total       int
	current     int
	description string
	startTime   time.Time
	mu          sync.Mutex
	writer      io.Writer
	enabled     bool
	lastUpdate  time.Time
	updateDelay time.Duration
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int, description string) *ProgressTracker {
	return &ProgressTracker{
		total:       total,
		description: description,
		startTime:   time.Now(),
		writer:      os.Stderr,
		enabled:     true,
		updateDelay: 100 * time.Millisecond, // Update at most every 100ms
	}
}

// SetEnabled enables or disables progress display
func (p *ProgressTracker) SetEnabled(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = enabled
}

// SetWriter sets the output writer for progress updates
func (p *ProgressTracker) SetWriter(w io.Writer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writer = w
}

// Increment increments the current count by 1
func (p *ProgressTracker) Increment() {
	p.Add(1)
}

// Add adds to the current count
func (p *ProgressTracker) Add(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current += n
	if p.current > p.total {
		p.current = p.total
	}

	// Rate limit updates
	now := time.Now()
	if now.Sub(p.lastUpdate) < p.updateDelay && p.current < p.total {
		return
	}
	p.lastUpdate = now

	p.render()
}

// SetTotal updates the total count (useful when total is discovered incrementally)
func (p *ProgressTracker) SetTotal(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = total
}

// Finish marks the progress as complete
func (p *ProgressTracker) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = p.total
	p.render()
	if p.enabled {
		fmt.Fprintln(p.writer) // New line after progress bar
	}
}

// render displays the current progress (must be called with lock held)
func (p *ProgressTracker) render() {
	if !p.enabled {
		return
	}

	elapsed := time.Since(p.startTime)
	percentage := 0
	if p.total > 0 {
		percentage = (p.current * 100) / p.total
	}

	// Calculate rate
	rate := 0.0
	if elapsed.Seconds() > 0 {
		rate = float64(p.current) / elapsed.Seconds()
	}

	// Calculate ETA
	eta := time.Duration(0)
	if rate > 0 && p.current < p.total {
		remaining := p.total - p.current
		eta = time.Duration(float64(remaining)/rate) * time.Second
	}

	// Build progress bar
	barWidth := 40
	filledWidth := (percentage * barWidth) / 100
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	// Format output
	output := fmt.Sprintf("\r%s [%s] %d%% (%d/%d) | %.1f/s",
		p.description,
		bar,
		percentage,
		p.current,
		p.total,
		rate,
	)

	if eta > 0 {
		output += fmt.Sprintf(" | ETA: %s", FormatDuration(eta))
	}

	// Clear to end of line and write
	fmt.Fprintf(p.writer, "\r%s\033[K", output)
}

// Spinner provides a simple spinner for operations without known progress
type Spinner struct {
	description string
	chars       []string
	index       int
	running     bool
	mu          sync.Mutex
	writer      io.Writer
	enabled     bool
	stopChan    chan struct{}
	stopOnce    sync.Once // Ensures channel is closed only once
}

// NewSpinner creates a new spinner
func NewSpinner(description string) *Spinner {
	return &Spinner{
		description: description,
		chars:       []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		writer:      os.Stderr,
		enabled:     true,
		stopChan:    make(chan struct{}),
	}
}

// SetEnabled enables or disables spinner display
func (s *Spinner) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// SetWriter sets the output writer for spinner
func (s *Spinner) SetWriter(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writer = w
}

// Start starts the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{}) // Recreate channel for reuse
	s.stopOnce = sync.Once{}         // Reset sync.Once for this start/stop cycle
	s.mu.Unlock()

	go s.animate()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false

	// Use sync.Once to ensure channel is closed exactly once
	// This prevents race conditions when Stop() is called concurrently
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})

	// Clear spinner line
	if s.enabled {
		fmt.Fprintf(s.writer, "\r\033[K")
	}
}

// animate runs the spinner animation loop
func (s *Spinner) animate() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.mu.Lock()
			if s.enabled {
				fmt.Fprintf(s.writer, "\r%s %s", s.chars[s.index], s.description)
			}
			s.index = (s.index + 1) % len(s.chars)
			s.mu.Unlock()
		}
	}
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
