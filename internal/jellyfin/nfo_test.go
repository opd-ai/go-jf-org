package jellyfin

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestGenerateMovieNFO(t *testing.T) {
	tests := []struct {
		name     string
		metadata *types.Metadata
		wantErr  bool
		validate func(t *testing.T, nfo string)
	}{
		{
			name: "basic movie with minimal metadata",
			metadata: &types.Metadata{
				Title: "The Matrix",
				Year:  1999,
			},
			wantErr: false,
			validate: func(t *testing.T, nfo string) {
				if !strings.Contains(nfo, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
					t.Error("NFO should contain XML declaration")
				}
				if !strings.Contains(nfo, "<movie>") {
					t.Error("NFO should contain movie root element")
				}
				if !strings.Contains(nfo, "<title>The Matrix</title>") {
					t.Error("NFO should contain movie title")
				}
				if !strings.Contains(nfo, "<year>1999</year>") {
					t.Error("NFO should contain year")
				}
			},
		},
		{
			name: "movie with full metadata",
			metadata: &types.Metadata{
				Title: "Inception",
				Year:  2010,
				MovieMetadata: &types.MovieMetadata{
					OriginalTitle: "Inception",
					Plot:          "A thief who steals corporate secrets through dream-sharing technology",
					Director:      []string{"Christopher Nolan"},
					Cast:          []string{"Leonardo DiCaprio", "Joseph Gordon-Levitt"},
					Genres:        []string{"Action", "Sci-Fi", "Thriller"},
					TMDBID:        27205,
					IMDBID:        "tt1375666",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, nfo string) {
				if !strings.Contains(nfo, "<title>Inception</title>") {
					t.Error("NFO should contain title")
				}
				if !strings.Contains(nfo, "<year>2010</year>") {
					t.Error("NFO should contain year")
				}
				if !strings.Contains(nfo, "<director>Christopher Nolan</director>") {
					t.Error("NFO should contain director")
				}
				if !strings.Contains(nfo, "<genre>Action</genre>") {
					t.Error("NFO should contain genres")
				}
				if !strings.Contains(nfo, "<tmdbid>27205</tmdbid>") {
					t.Error("NFO should contain TMDB ID")
				}
				if !strings.Contains(nfo, "<imdbid>tt1375666</imdbid>") {
					t.Error("NFO should contain IMDB ID")
				}
				
				// Validate XML is well-formed
				var movieNFO MovieNFO
				if err := xml.Unmarshal([]byte(nfo), &movieNFO); err != nil {
					t.Errorf("NFO should be valid XML: %v", err)
				}
			},
		},
		{
			name: "movie with special characters in title",
			metadata: &types.Metadata{
				Title: "Movie & Title: The <Special> Edition",
				Year:  2020,
			},
			wantErr: false,
			validate: func(t *testing.T, nfo string) {
				// Ensure it's valid XML - xml.Marshal handles escaping automatically
				var movieNFO MovieNFO
				if err := xml.Unmarshal([]byte(nfo), &movieNFO); err != nil {
					t.Errorf("NFO with special characters should be valid XML: %v", err)
				}
				
				// Verify the title was properly preserved after round-trip
				if movieNFO.Title != "Movie & Title: The <Special> Edition" {
					t.Errorf("Title should be preserved, got %q", movieNFO.Title)
				}
			},
		},
		{
			name:     "nil metadata should error",
			metadata: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewNFOGenerator()
			nfo, err := gen.GenerateMovieNFO(tt.metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMovieNFO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, nfo)
			}
		})
	}
}

func TestGenerateTVShowNFO(t *testing.T) {
	tests := []struct {
		name     string
		metadata *types.Metadata
		wantErr  bool
		validate func(t *testing.T, nfo string)
	}{
		{
			name: "basic TV show",
			metadata: &types.Metadata{
				Title: "Breaking Bad",
				TVMetadata: &types.TVMetadata{
					ShowTitle: "Breaking Bad",
					Plot:      "A chemistry teacher turns to cooking meth",
					AirDate:   "2008-01-20",
					TMDBID:    1396,
					TVDBID:    81189,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, nfo string) {
				if !strings.Contains(nfo, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
					t.Error("NFO should contain XML declaration")
				}
				if !strings.Contains(nfo, "<tvshow>") {
					t.Error("NFO should contain tvshow root element")
				}
				if !strings.Contains(nfo, "<title>Breaking Bad</title>") {
					t.Error("NFO should contain show title")
				}
				if !strings.Contains(nfo, "<premiered>2008-01-20</premiered>") {
					t.Error("NFO should contain premiere date")
				}
				if !strings.Contains(nfo, "<tmdbid>1396</tmdbid>") {
					t.Error("NFO should contain TMDB ID")
				}
				if !strings.Contains(nfo, "<tvdbid>81189</tvdbid>") {
					t.Error("NFO should contain TVDB ID")
				}
				
				// Validate XML
				var tvNFO TVShowNFO
				if err := xml.Unmarshal([]byte(nfo), &tvNFO); err != nil {
					t.Errorf("NFO should be valid XML: %v", err)
				}
			},
		},
		{
			name:     "nil metadata should error",
			metadata: nil,
			wantErr:  true,
		},
		{
			name: "missing TV metadata should error",
			metadata: &types.Metadata{
				Title: "Some Show",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewNFOGenerator()
			nfo, err := gen.GenerateTVShowNFO(tt.metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTVShowNFO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, nfo)
			}
		})
	}
}

func TestGenerateEpisodeNFO(t *testing.T) {
	tests := []struct {
		name     string
		metadata *types.Metadata
		wantErr  bool
		validate func(t *testing.T, nfo string)
	}{
		{
			name: "basic episode",
			metadata: &types.Metadata{
				TVMetadata: &types.TVMetadata{
					ShowTitle:    "Breaking Bad",
					Season:       1,
					Episode:      1,
					EpisodeTitle: "Pilot",
					Plot:         "Walter White gets a cancer diagnosis",
					AirDate:      "2008-01-20",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, nfo string) {
				if !strings.Contains(nfo, "<episodedetails>") {
					t.Error("NFO should contain episodedetails root element")
				}
				if !strings.Contains(nfo, "<title>Pilot</title>") {
					t.Error("NFO should contain episode title")
				}
				if !strings.Contains(nfo, "<season>1</season>") {
					t.Error("NFO should contain season number")
				}
				if !strings.Contains(nfo, "<episode>1</episode>") {
					t.Error("NFO should contain episode number")
				}
				if !strings.Contains(nfo, "<aired>2008-01-20</aired>") {
					t.Error("NFO should contain air date")
				}
				
				// Validate XML
				var episodeNFO EpisodeNFO
				if err := xml.Unmarshal([]byte(nfo), &episodeNFO); err != nil {
					t.Errorf("NFO should be valid XML: %v", err)
				}
			},
		},
		{
			name:     "nil metadata should error",
			metadata: nil,
			wantErr:  true,
		},
		{
			name: "missing TV metadata should error",
			metadata: &types.Metadata{
				Title: "Some Episode",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewNFOGenerator()
			nfo, err := gen.GenerateEpisodeNFO(tt.metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEpisodeNFO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, nfo)
			}
		})
	}
}

func TestGenerateSeasonNFO(t *testing.T) {
	tests := []struct {
		name         string
		seasonNumber int
		wantErr      bool
		validate     func(t *testing.T, nfo string)
	}{
		{
			name:         "season 1",
			seasonNumber: 1,
			wantErr:      false,
			validate: func(t *testing.T, nfo string) {
				if !strings.Contains(nfo, "<season>") {
					t.Error("NFO should contain season root element")
				}
				if !strings.Contains(nfo, "<seasonnumber>1</seasonnumber>") {
					t.Error("NFO should contain season number")
				}
				
				// Validate XML
				var seasonNFO SeasonNFO
				if err := xml.Unmarshal([]byte(nfo), &seasonNFO); err != nil {
					t.Errorf("NFO should be valid XML: %v", err)
				}
			},
		},
		{
			name:         "season 0 (specials)",
			seasonNumber: 0,
			wantErr:      false,
			validate: func(t *testing.T, nfo string) {
				// Season 0 is valid for specials - verify it appears in output
				var seasonNFO SeasonNFO
				if err := xml.Unmarshal([]byte(nfo), &seasonNFO); err != nil {
					t.Errorf("Season 0 NFO should be valid XML: %v", err)
				}
				if seasonNFO.SeasonNumber != 0 {
					t.Errorf("Expected season 0, got %d", seasonNFO.SeasonNumber)
				}
			},
		},
		{
			name:         "negative season should error",
			seasonNumber: -1,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewNFOGenerator()
			nfo, err := gen.GenerateSeasonNFO(tt.seasonNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSeasonNFO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, nfo)
			}
		})
	}
}

func TestMarshalNFO(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid movie NFO",
			input: MovieNFO{
				Title: "Test Movie",
				Year:  2023,
			},
			wantErr: false,
		},
		{
			name: "valid season NFO",
			input: SeasonNFO{
				SeasonNumber: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := marshalNFO(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("marshalNFO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !strings.HasPrefix(result, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
					t.Error("marshalNFO() should include XML declaration")
				}
				
				// Verify it's properly indented
				if !strings.Contains(result, "    ") {
					t.Error("marshalNFO() should indent with 4 spaces")
				}
			}
		})
	}
}

func TestGenerateMusicAlbumNFO(t *testing.T) {
tests := []struct {
name     string
metadata *types.Metadata
wantErr  bool
validate func(t *testing.T, nfo string)
}{
{
name: "basic album with minimal metadata",
metadata: &types.Metadata{
Title: "The Dark Side of the Moon",
Year:  1973,
MusicMetadata: &types.MusicMetadata{
Artist:      "Pink Floyd",
Album:       "The Dark Side of the Moon",
AlbumArtist: "Pink Floyd",
},
},
wantErr: false,
validate: func(t *testing.T, nfo string) {
if !strings.Contains(nfo, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
t.Error("NFO should contain XML declaration")
}
if !strings.Contains(nfo, "<album>") {
t.Error("NFO should contain album root element")
}
if !strings.Contains(nfo, "<title>The Dark Side of the Moon</title>") {
t.Error("NFO should contain album title")
}
if !strings.Contains(nfo, "<artist>Pink Floyd</artist>") {
t.Error("NFO should contain artist")
}
if !strings.Contains(nfo, "<year>1973</year>") {
t.Error("NFO should contain year")
}

// Validate XML is well-formed
var albumNFO MusicAlbumNFO
if err := xml.Unmarshal([]byte(nfo), &albumNFO); err != nil {
t.Errorf("NFO should be valid XML: %v", err)
}
},
},
{
name: "album with full metadata",
metadata: &types.Metadata{
Title: "Abbey Road",
Year:  1969,
MusicMetadata: &types.MusicMetadata{
Artist:         "The Beatles",
Album:          "Abbey Road",
AlbumArtist:    "The Beatles",
Genre:          "Rock",
MusicBrainzID:  "1234-5678",
MusicBrainzRID: "abcd-efgh",
},
},
wantErr: false,
validate: func(t *testing.T, nfo string) {
if !strings.Contains(nfo, "<title>Abbey Road</title>") {
t.Error("NFO should contain title")
}
if !strings.Contains(nfo, "<artist>The Beatles</artist>") {
t.Error("NFO should contain artist")
}
if !strings.Contains(nfo, "<albumartist>The Beatles</albumartist>") {
t.Error("NFO should contain album artist")
}
if !strings.Contains(nfo, "<genre>Rock</genre>") {
t.Error("NFO should contain genre")
}
if !strings.Contains(nfo, "<musicbrainzalbumid>1234-5678</musicbrainzalbumid>") {
t.Error("NFO should contain MusicBrainz album ID")
}
if !strings.Contains(nfo, "<musicbrainzreleasegroupid>abcd-efgh</musicbrainzreleasegroupid>") {
t.Error("NFO should contain MusicBrainz release ID")
}
},
},
{
name: "album without music metadata",
metadata: &types.Metadata{
Title: "Unknown Album",
Year:  2020,
},
wantErr: false,
validate: func(t *testing.T, nfo string) {
// Should still generate valid NFO with just title and year
if !strings.Contains(nfo, "<title>Unknown Album</title>") {
t.Error("NFO should contain title")
}
if !strings.Contains(nfo, "<year>2020</year>") {
t.Error("NFO should contain year")
}

var albumNFO MusicAlbumNFO
if err := xml.Unmarshal([]byte(nfo), &albumNFO); err != nil {
t.Errorf("NFO should be valid XML: %v", err)
}
},
},
{
name: "album with artist but no album artist",
metadata: &types.Metadata{
Title: "Greatest Hits",
Year:  2000,
MusicMetadata: &types.MusicMetadata{
Artist: "Various Artists",
Album:  "Greatest Hits",
// AlbumArtist is empty, should fall back to Artist
},
},
wantErr: false,
validate: func(t *testing.T, nfo string) {
if !strings.Contains(nfo, "<artist>Various Artists</artist>") {
t.Error("NFO should contain artist")
}
if !strings.Contains(nfo, "<albumartist>Various Artists</albumartist>") {
t.Error("NFO should fall back to artist for album artist")
}
},
},
{
name:     "nil metadata",
metadata: nil,
wantErr:  true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
gen := NewNFOGenerator()
result, err := gen.GenerateMusicAlbumNFO(tt.metadata)

if (err != nil) != tt.wantErr {
t.Errorf("GenerateMusicAlbumNFO() error = %v, wantErr %v", err, tt.wantErr)
return
}

if !tt.wantErr && tt.validate != nil {
tt.validate(t, result)
}
})
}
}

func TestGenerateBookNFO(t *testing.T) {
tests := []struct {
name     string
metadata *types.Metadata
wantErr  bool
validate func(t *testing.T, nfo string)
}{
{
name: "basic book with minimal metadata",
metadata: &types.Metadata{
Title: "The Great Gatsby",
Year:  1925,
BookMetadata: &types.BookMetadata{
Author: "F. Scott Fitzgerald",
},
},
wantErr: false,
validate: func(t *testing.T, nfo string) {
if !strings.Contains(nfo, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
t.Error("NFO should contain XML declaration")
}
if !strings.Contains(nfo, "<book>") {
t.Error("NFO should contain book root element")
}
if !strings.Contains(nfo, "<title>The Great Gatsby</title>") {
t.Error("NFO should contain book title")
}
if !strings.Contains(nfo, "<author>F. Scott Fitzgerald</author>") {
t.Error("NFO should contain author")
}
if !strings.Contains(nfo, "<year>1925</year>") {
t.Error("NFO should contain year")
}

// Validate XML is well-formed
var bookNFO BookNFO
if err := xml.Unmarshal([]byte(nfo), &bookNFO); err != nil {
t.Errorf("NFO should be valid XML: %v", err)
}
},
},
{
name: "book with full metadata",
metadata: &types.Metadata{
Title: "Harry Potter and the Philosopher's Stone",
Year:  1997,
BookMetadata: &types.BookMetadata{
Author:      "J.K. Rowling",
Publisher:   "Bloomsbury Publishing",
ISBN:        "978-0-7475-3269-9",
Series:      "Harry Potter",
SeriesIndex: 1,
Description: "A young wizard discovers his magical heritage on his 11th birthday.",
},
},
wantErr: false,
validate: func(t *testing.T, nfo string) {
// Validate it's proper XML first
var bookNFO BookNFO
if err := xml.Unmarshal([]byte(nfo), &bookNFO); err != nil {
t.Errorf("NFO should be valid XML: %v", err)
}

// Check content via unmarshaled struct to avoid XML escaping issues
if bookNFO.Title != "Harry Potter and the Philosopher's Stone" {
t.Errorf("NFO should contain correct title, got %q", bookNFO.Title)
}
if bookNFO.Author != "J.K. Rowling" {
t.Error("NFO should contain author")
}
if bookNFO.Publisher != "Bloomsbury Publishing" {
t.Error("NFO should contain publisher")
}
if bookNFO.ISBN != "978-0-7475-3269-9" {
t.Error("NFO should contain ISBN")
}
if bookNFO.Series != "Harry Potter" {
t.Error("NFO should contain series")
}
if bookNFO.SeriesIndex != 1 {
t.Error("NFO should contain series index")
}
if bookNFO.Description != "A young wizard discovers his magical heritage on his 11th birthday." {
t.Error("NFO should contain description")
}
},
},
{
name: "book without book metadata",
metadata: &types.Metadata{
Title: "Unknown Book",
Year:  2020,
},
wantErr: false,
validate: func(t *testing.T, nfo string) {
// Should still generate valid NFO with just title and year
if !strings.Contains(nfo, "<title>Unknown Book</title>") {
t.Error("NFO should contain title")
}
if !strings.Contains(nfo, "<year>2020</year>") {
t.Error("NFO should contain year")
}

var bookNFO BookNFO
if err := xml.Unmarshal([]byte(nfo), &bookNFO); err != nil {
t.Errorf("NFO should be valid XML: %v", err)
}
},
},
{
name: "book with special characters in metadata",
metadata: &types.Metadata{
Title: "Science & Technology: The <Modern> Age",
Year:  2021,
BookMetadata: &types.BookMetadata{
Author:      "Dr. Smith & Dr. Jones",
Description: "A comprehensive look at technology's impact on society.",
},
},
wantErr: false,
validate: func(t *testing.T, nfo string) {
// Ensure it's valid XML - xml.Marshal handles escaping automatically
var bookNFO BookNFO
if err := xml.Unmarshal([]byte(nfo), &bookNFO); err != nil {
t.Errorf("NFO with special characters should be valid XML: %v", err)
}

// Verify the title was properly preserved after round-trip
if bookNFO.Title != "Science & Technology: The <Modern> Age" {
t.Errorf("Title should be preserved, got %q", bookNFO.Title)
}
},
},
{
name:     "nil metadata",
metadata: nil,
wantErr:  true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
gen := NewNFOGenerator()
result, err := gen.GenerateBookNFO(tt.metadata)

if (err != nil) != tt.wantErr {
t.Errorf("GenerateBookNFO() error = %v, wantErr %v", err, tt.wantErr)
return
}

if !tt.wantErr && tt.validate != nil {
tt.validate(t, result)
}
})
}
}
