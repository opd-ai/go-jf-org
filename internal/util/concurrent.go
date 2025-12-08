package util

import (
	"context"
	"sync"

	"github.com/opd-ai/go-jf-org/pkg/types"
	"github.com/rs/zerolog/log"
)

// EnricherFunc is a function that enriches metadata for a single file
type EnricherFunc func(*types.Metadata) error

// EnrichmentResult represents the result of enriching a single file
type EnrichmentResult struct {
	Index    int
	Metadata *types.Metadata
	Error    error
}

// ConcurrentEnricher manages concurrent metadata enrichment
type ConcurrentEnricher struct {
	numWorkers int
}

// NewConcurrentEnricher creates a new concurrent enricher
func NewConcurrentEnricher(numWorkers int) *ConcurrentEnricher {
	if numWorkers < 1 {
		numWorkers = 1
	}
	return &ConcurrentEnricher{
		numWorkers: numWorkers,
	}
}

// EnrichBatch enriches a batch of metadata using concurrent workers
// The enricher function is called for each metadata item
// Results are returned in the same order as the input
func (ce *ConcurrentEnricher) EnrichBatch(ctx context.Context, metadataList []*types.Metadata, enricher EnricherFunc) ([]*types.Metadata, []error) {
	if len(metadataList) == 0 {
		return metadataList, nil
	}

	// Create channels
	jobChan := make(chan int, len(metadataList))
	resultChan := make(chan EnrichmentResult, len(metadataList))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < ce.numWorkers; i++ {
		wg.Add(1)
		go ce.worker(ctx, &wg, jobChan, resultChan, metadataList, enricher)
	}

	// Send jobs
	go func() {
		for i := range metadataList {
			select {
			case jobChan <- i:
			case <-ctx.Done():
				break
			}
		}
		close(jobChan)
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]*types.Metadata, len(metadataList))
	errors := make([]error, len(metadataList))

	for result := range resultChan {
		results[result.Index] = result.Metadata
		errors[result.Index] = result.Error
	}

	return results, errors
}

// worker processes enrichment jobs
func (ce *ConcurrentEnricher) worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan int, resultChan chan<- EnrichmentResult, metadataList []*types.Metadata, enricher EnricherFunc) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case idx, ok := <-jobChan:
			if !ok {
				return
			}

			metadata := metadataList[idx]
			err := enricher(metadata)

			result := EnrichmentResult{
				Index:    idx,
				Metadata: metadata,
				Error:    err,
			}

			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// EnrichWithProgress enriches a batch of metadata with progress tracking
func (ce *ConcurrentEnricher) EnrichWithProgress(ctx context.Context, metadataList []*types.Metadata, enricher EnricherFunc, progress *ProgressTracker) ([]*types.Metadata, []error) {
	if len(metadataList) == 0 {
		return metadataList, nil
	}

	// Set total if progress tracker provided
	if progress != nil {
		progress.SetTotal(len(metadataList))
	}

	// Create channels
	jobChan := make(chan int, len(metadataList))
	resultChan := make(chan EnrichmentResult, len(metadataList))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < ce.numWorkers; i++ {
		wg.Add(1)
		go ce.workerWithProgress(ctx, &wg, jobChan, resultChan, metadataList, enricher, progress)
	}

	// Send jobs
	go func() {
		for i := range metadataList {
			select {
			case jobChan <- i:
			case <-ctx.Done():
				break
			}
		}
		close(jobChan)
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]*types.Metadata, len(metadataList))
	errors := make([]error, len(metadataList))

	for result := range resultChan {
		results[result.Index] = result.Metadata
		errors[result.Index] = result.Error
	}

	// Finish progress
	if progress != nil {
		progress.Finish()
	}

	return results, errors
}

// workerWithProgress processes enrichment jobs and updates progress
func (ce *ConcurrentEnricher) workerWithProgress(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan int, resultChan chan<- EnrichmentResult, metadataList []*types.Metadata, enricher EnricherFunc, progress *ProgressTracker) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case idx, ok := <-jobChan:
			if !ok {
				return
			}

			metadata := metadataList[idx]
			err := enricher(metadata)

			if err != nil {
				log.Debug().Err(err).Int("index", idx).Msg("Enrichment error")
			}

			// Update progress
			if progress != nil {
				progress.Increment()
			}

			result := EnrichmentResult{
				Index:    idx,
				Metadata: metadata,
				Error:    err,
			}

			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
