package openlibrary

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opd-ai/go-jf-org/pkg/types"
	"github.com/rs/zerolog/log"
)

// Enricher enriches metadata using OpenLibrary API
type Enricher struct {
	client *Client
}

// NewEnricher creates a new metadata enricher
func NewEnricher(client *Client) *Enricher {
	return &Enricher{client: client}
}

// EnrichBook enriches book metadata with OpenLibrary data
func (e *Enricher) EnrichBook(metadata *types.Metadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}

	// Ensure BookMetadata exists
	if metadata.BookMetadata == nil {
		metadata.BookMetadata = &types.BookMetadata{}
	}

	// Try ISBN lookup first if available
	if metadata.BookMetadata.ISBN != "" {
		if isbnErr := e.enrichByISBN(metadata); isbnErr == nil {
			return nil
		} else {
			log.Debug().Err(isbnErr).Msg("ISBN lookup failed, falling back to search")
		}
	}

	// Fall back to title/author search
	title := metadata.Title
	author := metadata.BookMetadata.Author

	if title == "" {
		return fmt.Errorf("title is required for enrichment")
	}

	log.Debug().
		Str("title", title).
		Str("author", author).
		Msg("Enriching book metadata")

	// Search for book
	searchResp, err := e.client.Search(title, author)
	if err != nil {
		return fmt.Errorf("failed to search book: %w", err)
	}

	if searchResp.NumFound == 0 {
		log.Warn().
			Str("title", title).
			Str("author", author).
			Msg("No OpenLibrary results found for book")
		return nil // Not an error, just no results
	}

	// Use first result (best match)
	book := searchResp.Docs[0]

	// Apply search result metadata
	e.applyBookSearchResult(metadata, &book)

	// Try to get more details if we have a key
	if book.Key != "" {
		details, err := e.client.GetBookDetails(book.Key)
		if err != nil {
			log.Debug().Err(err).Str("key", book.Key).Msg("Failed to get book details")
		} else {
			e.applyBookDetails(metadata, details)
		}
	}

	log.Info().
		Str("title", metadata.Title).
		Str("author", metadata.BookMetadata.Author).
		Str("isbn", metadata.BookMetadata.ISBN).
		Msg("Book metadata enriched")

	return nil
}

// enrichByISBN enriches metadata using ISBN lookup
func (e *Enricher) enrichByISBN(metadata *types.Metadata) error {
	isbn := metadata.BookMetadata.ISBN
	log.Debug().Str("isbn", isbn).Msg("Looking up book by ISBN")

	response, err := e.client.GetBookByISBN(isbn)
	if err != nil {
		return err
	}

	// Set title
	if metadata.Title == "" {
		metadata.Title = response.Title
	}

	// Set author from authors reference
	if metadata.BookMetadata.Author == "" && len(response.Authors) > 0 {
		// Get author details
		authorDetails, err := e.client.GetAuthorDetails(response.Authors[0].Key)
		if err == nil {
			metadata.BookMetadata.Author = authorDetails.Name
		}
	}

	// Set year from publish date
	if metadata.Year == 0 && response.PublishDate != "" {
		year := e.extractYear(response.PublishDate)
		if year > 0 {
			metadata.Year = year
		}
	}

	// Set publisher
	if metadata.BookMetadata.Publisher == "" && len(response.Publishers) > 0 {
		metadata.BookMetadata.Publisher = response.Publishers[0]
	}

	// Set description from work if available
	if metadata.BookMetadata.Description == "" && len(response.Works) > 0 {
		workDetails, err := e.client.GetWorkDetails(response.Works[0].Key)
		if err == nil {
			metadata.BookMetadata.Description = e.extractDescription(workDetails.Description)
		}
	}

	log.Info().Str("isbn", isbn).Msg("Book metadata enriched by ISBN")
	return nil
}

// applyBookSearchResult applies metadata from a book search result
func (e *Enricher) applyBookSearchResult(metadata *types.Metadata, book *BookDoc) {
	// Set title
	if metadata.Title == "" {
		metadata.Title = book.Title
	}

	// Set author
	if metadata.BookMetadata.Author == "" && len(book.AuthorName) > 0 {
		metadata.BookMetadata.Author = book.AuthorName[0]
	}

	// Set year
	if metadata.Year == 0 && book.FirstPublishYear > 0 {
		metadata.Year = book.FirstPublishYear
	}

	// Set ISBN
	if metadata.BookMetadata.ISBN == "" && len(book.ISBN) > 0 {
		// Prefer ISBN-13, fall back to ISBN-10
		for _, isbn := range book.ISBN {
			if len(isbn) == 13 {
				metadata.BookMetadata.ISBN = isbn
				break
			}
		}
		if metadata.BookMetadata.ISBN == "" {
			metadata.BookMetadata.ISBN = book.ISBN[0]
		}
	}

	// Set publisher
	if metadata.BookMetadata.Publisher == "" && len(book.Publisher) > 0 {
		metadata.BookMetadata.Publisher = book.Publisher[0]
	}

	log.Debug().
		Str("title", metadata.Title).
		Str("author", metadata.BookMetadata.Author).
		Int("year", metadata.Year).
		Msg("Applied OpenLibrary search result")
}

// applyBookDetails applies metadata from detailed book information
func (e *Enricher) applyBookDetails(metadata *types.Metadata, details *BookDetails) {
	// Set title and subtitle
	if metadata.Title == "" {
		if details.Subtitle != "" {
			metadata.Title = fmt.Sprintf("%s: %s", details.Title, details.Subtitle)
		} else {
			metadata.Title = details.Title
		}
	}

	// Set description
	if metadata.BookMetadata.Description == "" {
		metadata.BookMetadata.Description = e.extractDescription(details.Description)
	}

	// Set year from publish date
	if metadata.Year == 0 && details.PublishDate != "" {
		year := e.extractYear(details.PublishDate)
		if year > 0 {
			metadata.Year = year
		}
	}

	// Set publisher
	if metadata.BookMetadata.Publisher == "" && len(details.Publishers) > 0 {
		metadata.BookMetadata.Publisher = details.Publishers[0]
	}

	// Set ISBN (prefer ISBN-13)
	if metadata.BookMetadata.ISBN == "" {
		if len(details.ISBN13) > 0 {
			metadata.BookMetadata.ISBN = details.ISBN13[0]
		} else if len(details.ISBN10) > 0 {
			metadata.BookMetadata.ISBN = details.ISBN10[0]
		}
	}

	log.Debug().
		Str("title", metadata.Title).
		Str("description", metadata.BookMetadata.Description[:min(50, len(metadata.BookMetadata.Description))]).
		Msg("Applied OpenLibrary book details")
}

// extractDescription extracts description string from interface{} (can be string or object)
func (e *Enricher) extractDescription(desc interface{}) string {
	if desc == nil {
		return ""
	}

	// Try as string
	if str, ok := desc.(string); ok {
		return str
	}

	// Try as object with "value" field
	if obj, ok := desc.(map[string]interface{}); ok {
		if value, exists := obj["value"]; exists {
			if str, ok := value.(string); ok {
				return str
			}
		}
	}

	return ""
}

// extractYear extracts year from date string (various formats)
func (e *Enricher) extractYear(dateStr string) int {
	if dateStr == "" {
		return 0
	}

	// Try to find a 4-digit year in the string
	parts := strings.Fields(dateStr)
	for _, part := range parts {
		// Remove common separators
		part = strings.Trim(part, ",-./")
		
		// Try to parse as integer
		year, err := strconv.Atoi(part)
		if err == nil && year >= 1450 && year <= 2100 {
			return year
		}
	}

	// Try YYYY-MM-DD format
	parts = strings.Split(dateStr, "-")
	if len(parts) > 0 {
		year, err := strconv.Atoi(parts[0])
		if err == nil && year >= 1450 && year <= 2100 {
			return year
		}
	}

	return 0
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
