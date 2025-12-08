package util

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestConcurrentEnricher_EnrichBatch(t *testing.T) {
	tests := []struct {
		name       string
		numWorkers int
		items      int
		wantErrors bool
	}{
		{
			name:       "single worker",
			numWorkers: 1,
			items:      10,
			wantErrors: false,
		},
		{
			name:       "multiple workers",
			numWorkers: 4,
			items:      20,
			wantErrors: false,
		},
		{
			name:       "more workers than items",
			numWorkers: 10,
			items:      5,
			wantErrors: false,
		},
		{
			name:       "zero items",
			numWorkers: 4,
			items:      0,
			wantErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test metadata
			metadataList := make([]*types.Metadata, tt.items)
			for i := 0; i < tt.items; i++ {
				metadataList[i] = &types.Metadata{
					Title: "Test",
					Year:  2000 + i,
				}
			}

			// Create enricher
			enricher := NewConcurrentEnricher(tt.numWorkers)

			// Define enrichment function
			enrichFunc := func(m *types.Metadata) error {
				// Simulate enrichment by adding to year
				m.Year += 1000
				return nil
			}

			// Enrich
			ctx := context.Background()
			results, errs := enricher.EnrichBatch(ctx, metadataList, enrichFunc)

			// Verify
			if len(results) != tt.items {
				t.Errorf("Expected %d results, got %d", tt.items, len(results))
			}

			if len(errs) != tt.items {
				t.Errorf("Expected %d errors, got %d", tt.items, len(errs))
			}

			// Verify enrichment
			for i, result := range results {
				if result == nil {
					continue
				}
				expectedYear := 3000 + i
				if result.Year != expectedYear {
					t.Errorf("Item %d: expected year %d, got %d", i, expectedYear, result.Year)
				}
			}
		})
	}
}

func TestConcurrentEnricher_EnrichBatchWithErrors(t *testing.T) {
	metadataList := make([]*types.Metadata, 10)
	for i := 0; i < 10; i++ {
		metadataList[i] = &types.Metadata{
			Title: "Test",
			Year:  2000 + i,
		}
	}

	enricher := NewConcurrentEnricher(4)

	// Enrichment function that fails on even indices
	enrichFunc := func(m *types.Metadata) error {
		if m.Year%2 == 0 {
			return errors.New("even year error")
		}
		m.Year += 1000
		return nil
	}

	ctx := context.Background()
	results, errs := enricher.EnrichBatch(ctx, metadataList, enrichFunc)

	// Count errors
	errorCount := 0
	for _, err := range errs {
		if err != nil {
			errorCount++
		}
	}

	if errorCount != 5 {
		t.Errorf("Expected 5 errors, got %d", errorCount)
	}

	// Verify successful enrichments
	for i, result := range results {
		if errs[i] != nil {
			continue
		}
		expectedYear := 3000 + i
		if result.Year != expectedYear {
			t.Errorf("Item %d: expected year %d, got %d", i, expectedYear, result.Year)
		}
	}
}

func TestConcurrentEnricher_ContextCancellation(t *testing.T) {
	metadataList := make([]*types.Metadata, 100)
	for i := 0; i < 100; i++ {
		metadataList[i] = &types.Metadata{
			Title: "Test",
			Year:  2000 + i,
		}
	}

	enricher := NewConcurrentEnricher(4)

	// Enrichment function that takes time
	enrichFunc := func(m *types.Metadata) error {
		time.Sleep(10 * time.Millisecond)
		m.Year += 1000
		return nil
	}

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	results, _ := enricher.EnrichBatch(ctx, metadataList, enrichFunc)

	// Should have some results but not all
	completedCount := 0
	for _, result := range results {
		if result != nil && result.Year >= 3000 {
			completedCount++
		}
	}

	// We expect some but not all to complete
	if completedCount == 0 {
		t.Error("Expected some items to complete before cancellation")
	}
	if completedCount == 100 {
		t.Error("Expected cancellation to stop some items from completing")
	}
}

func TestConcurrentEnricher_WithProgress(t *testing.T) {
	metadataList := make([]*types.Metadata, 20)
	for i := 0; i < 20; i++ {
		metadataList[i] = &types.Metadata{
			Title: "Test",
			Year:  2000 + i,
		}
	}

	enricher := NewConcurrentEnricher(4)
	progress := NewProgressTracker(0, "Enriching")

	enrichFunc := func(m *types.Metadata) error {
		m.Year += 1000
		return nil
	}

	ctx := context.Background()
	results, errs := enricher.EnrichWithProgress(ctx, metadataList, enrichFunc, progress)

	// Verify results
	if len(results) != 20 {
		t.Errorf("Expected 20 results, got %d", len(results))
	}

	// Verify all completed
	for i, err := range errs {
		if err != nil {
			t.Errorf("Item %d: unexpected error: %v", i, err)
		}
	}

	// Verify progress completed
	if progress.current != len(metadataList) {
		t.Errorf("Expected progress to be %d, got %d", len(metadataList), progress.current)
	}
}

func TestConcurrentEnricher_ZeroWorkers(t *testing.T) {
	// Should default to 1 worker
	enricher := NewConcurrentEnricher(0)
	if enricher.numWorkers != 1 {
		t.Errorf("Expected 1 worker, got %d", enricher.numWorkers)
	}
}

func BenchmarkConcurrentEnricher_Sequential(b *testing.B) {
	benchmarkEnricher(b, 1)
}

func BenchmarkConcurrentEnricher_Parallel2(b *testing.B) {
	benchmarkEnricher(b, 2)
}

func BenchmarkConcurrentEnricher_Parallel4(b *testing.B) {
	benchmarkEnricher(b, 4)
}

func BenchmarkConcurrentEnricher_Parallel8(b *testing.B) {
	benchmarkEnricher(b, 8)
}

func benchmarkEnricher(b *testing.B, numWorkers int) {
	metadataList := make([]*types.Metadata, 100)
	for i := 0; i < 100; i++ {
		metadataList[i] = &types.Metadata{
			Title: "Test",
			Year:  2000 + i,
		}
	}

	enricher := NewConcurrentEnricher(numWorkers)
	enrichFunc := func(m *types.Metadata) error {
		// Simulate some work
		time.Sleep(100 * time.Microsecond)
		m.Year += 1000
		return nil
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = enricher.EnrichBatch(ctx, metadataList, enrichFunc)
	}
}
