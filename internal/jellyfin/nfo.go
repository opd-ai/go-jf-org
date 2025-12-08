package jellyfin

import (
	"encoding/xml"
	"fmt"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// NFOGenerator generates Kodi-compatible NFO files for Jellyfin
type NFOGenerator struct{}

// NewNFOGenerator creates a new NFO generator
func NewNFOGenerator() *NFOGenerator {
	return &NFOGenerator{}
}

// MovieNFO represents the XML structure for a movie NFO file
type MovieNFO struct {
	XMLName       xml.Name `xml:"movie"`
	Title         string   `xml:"title,omitempty"`
	OriginalTitle string   `xml:"originaltitle,omitempty"`
	Year          int      `xml:"year,omitempty"`
	Plot          string   `xml:"plot,omitempty"`
	Tagline       string   `xml:"tagline,omitempty"`
	Runtime       int      `xml:"runtime,omitempty"`
	MPAA          string   `xml:"mpaa,omitempty"`
	Genres        []string `xml:"genre,omitempty"`
	Studio        string   `xml:"studio,omitempty"`
	Directors     []string `xml:"director,omitempty"`
	Actors        []Actor  `xml:"actor,omitempty"`
	TMDBID        int      `xml:"tmdbid,omitempty"`
	IMDBID        string   `xml:"imdbid,omitempty"`
}

// TVShowNFO represents the XML structure for a TV show NFO file
type TVShowNFO struct {
	XMLName   xml.Name `xml:"tvshow"`
	Title     string   `xml:"title,omitempty"`
	Plot      string   `xml:"plot,omitempty"`
	Premiered string   `xml:"premiered,omitempty"`
	Genres    []string `xml:"genre,omitempty"`
	Studio    string   `xml:"studio,omitempty"`
	Actors    []Actor  `xml:"actor,omitempty"`
	TVDBID    int      `xml:"tvdbid,omitempty"`
	TMDBID    int      `xml:"tmdbid,omitempty"`
}

// EpisodeNFO represents the XML structure for a TV episode NFO file
type EpisodeNFO struct {
	XMLName xml.Name `xml:"episodedetails"`
	Title   string   `xml:"title,omitempty"`
	Season  int      `xml:"season,omitempty"`
	Episode int      `xml:"episode,omitempty"`
	Plot    string   `xml:"plot,omitempty"`
	Aired   string   `xml:"aired,omitempty"`
}

// SeasonNFO represents the XML structure for a season NFO file
type SeasonNFO struct {
	XMLName      xml.Name `xml:"season"`
	SeasonNumber int      `xml:"seasonnumber,omitempty"`
}

// Actor represents an actor in a movie or TV show
type Actor struct {
	Name string `xml:"name,omitempty"`
	Role string `xml:"role,omitempty"`
}

// GenerateMovieNFO generates a movie.nfo XML file content
func (g *NFOGenerator) GenerateMovieNFO(metadata *types.Metadata) (string, error) {
	if metadata == nil {
		return "", fmt.Errorf("metadata cannot be nil")
	}

	nfo := MovieNFO{
		Title:         metadata.Title,
		OriginalTitle: metadata.Title, // Default to same as title
		Year:          metadata.Year,
	}

	// Add movie-specific metadata if available
	if metadata.MovieMetadata != nil {
		mm := metadata.MovieMetadata
		
		if mm.OriginalTitle != "" {
			nfo.OriginalTitle = mm.OriginalTitle
		}
		
		nfo.Plot = mm.Plot
		nfo.TMDBID = mm.TMDBID
		nfo.IMDBID = mm.IMDBID
		
		for _, genre := range mm.Genres {
			nfo.Genres = append(nfo.Genres, genre)
		}
		
		for _, director := range mm.Director {
			nfo.Directors = append(nfo.Directors, director)
		}
		
		for _, cast := range mm.Cast {
			nfo.Actors = append(nfo.Actors, Actor{
				Name: cast,
			})
		}
	}

	return marshalNFO(nfo)
}

// GenerateTVShowNFO generates a tvshow.nfo XML file content
func (g *NFOGenerator) GenerateTVShowNFO(metadata *types.Metadata) (string, error) {
	if metadata == nil {
		return "", fmt.Errorf("metadata cannot be nil")
	}

	if metadata.TVMetadata == nil {
		return "", fmt.Errorf("TV metadata is required")
	}

	tm := metadata.TVMetadata
	
	nfo := TVShowNFO{
		Title: tm.ShowTitle,
		Plot:  tm.Plot,
	}

	if tm.AirDate != "" {
		nfo.Premiered = tm.AirDate
	}

	nfo.TMDBID = tm.TMDBID
	nfo.TVDBID = tm.TVDBID

	return marshalNFO(nfo)
}

// GenerateEpisodeNFO generates an episode NFO XML file content
// This is typically embedded in the video file or created as <filename>.nfo
func (g *NFOGenerator) GenerateEpisodeNFO(metadata *types.Metadata) (string, error) {
	if metadata == nil {
		return "", fmt.Errorf("metadata cannot be nil")
	}

	if metadata.TVMetadata == nil {
		return "", fmt.Errorf("TV metadata is required")
	}

	tm := metadata.TVMetadata
	
	nfo := EpisodeNFO{
		Title:   tm.EpisodeTitle,
		Season:  tm.Season,
		Episode: tm.Episode,
		Plot:    tm.Plot,
		Aired:   tm.AirDate,
	}

	return marshalNFO(nfo)
}

// GenerateSeasonNFO generates a season.nfo XML file content
func (g *NFOGenerator) GenerateSeasonNFO(seasonNumber int) (string, error) {
	if seasonNumber < 0 {
		return "", fmt.Errorf("season number cannot be negative")
	}

	nfo := SeasonNFO{
		SeasonNumber: seasonNumber,
	}

	return marshalNFO(nfo)
}

// marshalNFO marshals an NFO structure to XML with proper formatting
func marshalNFO(v interface{}) (string, error) {
	data, err := xml.MarshalIndent(v, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NFO: %w", err)
	}

	// Add XML declaration
	xmlHeader := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"
	return xmlHeader + string(data), nil
}
