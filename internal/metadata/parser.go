package metadata

import (
	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Parser extracts metadata from filenames and files
type Parser interface {
	// Parse extracts metadata from a filename
	Parse(filename string, mediaType types.MediaType) (*types.Metadata, error)
}

// parser is the main implementation
type parser struct {
	movieParser MovieParser
	tvParser    TVParser
}

// NewParser creates a new Parser instance
func NewParser() Parser {
	return &parser{
		movieParser: NewMovieParser(),
		tvParser:    NewTVParser(),
	}
}

// Parse extracts metadata based on the media type
func (p *parser) Parse(filename string, mediaType types.MediaType) (*types.Metadata, error) {
	switch mediaType {
	case types.MediaTypeMovie:
		return p.movieParser.Parse(filename)
	case types.MediaTypeTV:
		return p.tvParser.Parse(filename)
	default:
		// For music and books, we'll implement later
		return &types.Metadata{}, nil
	}
}
