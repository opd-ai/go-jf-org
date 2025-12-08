package metadata

import (
	"testing"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestMovieParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		wantTitle   string
		wantYear    int
		wantQuality string
		wantSource  string
		wantCodec   string
	}{
		{
			name:        "standard movie with year and quality",
			filename:    "The.Matrix.1999.1080p.BluRay.x264.mkv",
			wantTitle:   "The Matrix",
			wantYear:    1999,
			wantQuality: "1080P",
			wantSource:  "BluRay",
			wantCodec:   "x264",
		},
		{
			name:        "movie with year in parentheses",
			filename:    "Inception (2010) 1080p.mp4",
			wantTitle:   "Inception",
			wantYear:    2010,
			wantQuality: "1080P",
		},
		{
			name:        "movie with year in brackets",
			filename:    "The.Dark.Knight.[2008].720p.BDRip.mkv",
			wantTitle:   "The Dark Knight",
			wantYear:    2008,
			wantQuality: "720P",
			wantSource:  "BDRip",
		},
		{
			name:        "movie with 4K quality",
			filename:    "Blade.Runner.2049.2160p.4K.UHD.BluRay.x265.mkv",
			wantTitle:   "Blade Runner",
			wantYear:    2049,
			wantQuality: "2160P", // 2160P matches first in the regex
			wantSource:  "BluRay",
			wantCodec:   "x265",
		},
		{
			name:        "movie with WEB-DL source",
			filename:    "Movie.Title.2023.1080p.WEB-DL.h264.mkv",
			wantTitle:   "Movie Title",
			wantYear:    2023,
			wantQuality: "1080P",
			wantSource:  "WEB-DL",
			wantCodec:   "h264",
		},
		{
			name:        "movie with underscores",
			filename:    "Some_Movie_Title_2020_720p.mp4",
			wantTitle:   "Some Movie Title",
			wantYear:    2020,
			wantQuality: "720P",
		},
		{
			name:      "movie with old year",
			filename:  "Metropolis.1927.Restored.mkv",
			wantTitle: "Metropolis",
			wantYear:  1927,
		},
		{
			name:        "movie with HEVC codec",
			filename:    "Film.2022.2160p.HEVC.mkv",
			wantTitle:   "Film",
			wantYear:    2022,
			wantQuality: "2160P",
			wantCodec:   "hevc",
		},
		{
			name:      "movie without quality or source",
			filename:  "Simple.Movie.2021.mkv",
			wantTitle: "Simple Movie",
			wantYear:  2021,
		},
	}

	parser := NewMovieParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.filename)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
			}
			if got.Year != tt.wantYear {
				t.Errorf("Year = %d, want %d", got.Year, tt.wantYear)
			}
			if tt.wantQuality != "" && got.Quality != tt.wantQuality {
				t.Errorf("Quality = %q, want %q", got.Quality, tt.wantQuality)
			}
			if tt.wantSource != "" && got.Source != tt.wantSource {
				t.Errorf("Source = %q, want %q", got.Source, tt.wantSource)
			}
			if tt.wantCodec != "" && got.Codec != tt.wantCodec {
				t.Errorf("Codec = %q, want %q", got.Codec, tt.wantCodec)
			}
			if got.MovieMetadata == nil {
				t.Error("MovieMetadata should not be nil")
			}
		})
	}
}

func TestTVParser_Parse(t *testing.T) {
	tests := []struct {
		name             string
		filename         string
		wantShowTitle    string
		wantSeason       int
		wantEpisode      int
		wantEpisodeTitle string
	}{
		{
			name:          "standard S01E01 format",
			filename:      "Breaking.Bad.S01E01.720p.mkv",
			wantShowTitle: "Breaking Bad",
			wantSeason:    1,
			wantEpisode:   1,
		},
		{
			name:          "lowercase s01e01 format",
			filename:      "game.of.thrones.s05e09.1080p.mp4",
			wantShowTitle: "game of thrones",
			wantSeason:    5,
			wantEpisode:   9,
		},
		{
			name:             "with episode title",
			filename:         "The.Office.S02E15.The.Big.Job.720p.WEB-DL.mkv",
			wantShowTitle:    "The Office",
			wantSeason:       2,
			wantEpisode:      15,
			wantEpisodeTitle: "The Big Job",
		},
		{
			name:          "alternative 1x01 format",
			filename:      "Show.Name.1x01.Episode.mkv",
			wantShowTitle: "Show Name",
			wantSeason:    1,
			wantEpisode:   1,
		},
		{
			name:          "alternative format with leading zero",
			filename:      "Series.Title.01x05.720p.mkv",
			wantShowTitle: "Series Title",
			wantSeason:    1,
			wantEpisode:   5,
		},
		{
			name:          "single digit season/episode",
			filename:      "Anime.Series.S1E5.mkv",
			wantShowTitle: "Anime Series",
			wantSeason:    1,
			wantEpisode:   5,
		},
		{
			name:          "three digit episode",
			filename:      "Long.Series.S01E123.mkv",
			wantShowTitle: "Long Series",
			wantSeason:    1,
			wantEpisode:   123,
		},
		{
			name:          "show with year in title",
			filename:      "Doctor.Who.2005.S01E01.mkv",
			wantShowTitle: "Doctor Who 2005",
			wantSeason:    1,
			wantEpisode:   1,
		},
		{
			name:          "show with underscores",
			filename:      "My_Show_S02E10_720p.mkv",
			wantShowTitle: "My Show",
			wantSeason:    2,
			wantEpisode:   10,
		},
	}

	parser := NewTVParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.filename)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if got.TVMetadata == nil {
				t.Fatal("TVMetadata should not be nil")
			}

			if got.TVMetadata.ShowTitle != tt.wantShowTitle {
				t.Errorf("ShowTitle = %q, want %q", got.TVMetadata.ShowTitle, tt.wantShowTitle)
			}
			if got.TVMetadata.Season != tt.wantSeason {
				t.Errorf("Season = %d, want %d", got.TVMetadata.Season, tt.wantSeason)
			}
			if got.TVMetadata.Episode != tt.wantEpisode {
				t.Errorf("Episode = %d, want %d", got.TVMetadata.Episode, tt.wantEpisode)
			}
			if tt.wantEpisodeTitle != "" && got.TVMetadata.EpisodeTitle != tt.wantEpisodeTitle {
				t.Errorf("EpisodeTitle = %q, want %q", got.TVMetadata.EpisodeTitle, tt.wantEpisodeTitle)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		mediaType types.MediaType
		wantTitle string
		checkFunc func(*testing.T, *types.Metadata)
	}{
		{
			name:      "movie type",
			filename:  "The.Matrix.1999.1080p.mkv",
			mediaType: types.MediaTypeMovie,
			wantTitle: "The Matrix",
			checkFunc: func(t *testing.T, m *types.Metadata) {
				if m.Year != 1999 {
					t.Errorf("Year = %d, want 1999", m.Year)
				}
				if m.MovieMetadata == nil {
					t.Error("MovieMetadata should not be nil for movie")
				}
			},
		},
		{
			name:      "tv type",
			filename:  "Breaking.Bad.S01E01.mkv",
			mediaType: types.MediaTypeTV,
			wantTitle: "Breaking Bad",
			checkFunc: func(t *testing.T, m *types.Metadata) {
				if m.TVMetadata == nil {
					t.Error("TVMetadata should not be nil for TV show")
				}
				if m.TVMetadata.Season != 1 {
					t.Errorf("Season = %d, want 1", m.TVMetadata.Season)
				}
				if m.TVMetadata.Episode != 1 {
					t.Errorf("Episode = %d, want 1", m.TVMetadata.Episode)
				}
			},
		},
		{
			name:      "unknown type returns empty metadata",
			filename:  "file.txt",
			mediaType: types.MediaTypeUnknown,
			checkFunc: func(t *testing.T, m *types.Metadata) {
				// Should not crash and return valid (empty) metadata
				if m == nil {
					t.Error("Metadata should not be nil")
				}
			},
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.filename, tt.mediaType)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got == nil {
				t.Fatal("Parse() returned nil metadata")
			}

			if tt.wantTitle != "" && got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}
