package detector

import (
	"testing"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     types.MediaType
	}{
		// Movie tests
		{
			name:     "movie with year and quality",
			filename: "The.Matrix.1999.1080p.BluRay.x264.mkv",
			want:     types.MediaTypeMovie,
		},
		{
			name:     "movie with year in parentheses",
			filename: "Inception (2010).mp4",
			want:     types.MediaTypeMovie,
		},
		{
			name:     "movie with year in brackets",
			filename: "The.Dark.Knight.[2008].1080p.mkv",
			want:     types.MediaTypeMovie,
		},
		{
			name:     "movie with quality tags",
			filename: "SomeMovie.2160p.4K.UHD.BluRay.x265.mkv",
			want:     types.MediaTypeMovie,
		},

		// TV show tests
		{
			name:     "tv show with S01E01 pattern",
			filename: "Breaking.Bad.S01E01.720p.mkv",
			want:     types.MediaTypeTV,
		},
		{
			name:     "tv show with s01e01 lowercase",
			filename: "game.of.thrones.s05e09.1080p.mp4",
			want:     types.MediaTypeTV,
		},
		{
			name:     "tv show with 1x01 pattern",
			filename: "ShowName.1x01.Episode.Title.mkv",
			want:     types.MediaTypeTV,
		},
		{
			name:     "tv show with leading zero in alt pattern",
			filename: "Series.Name.01x05.720p.mkv",
			want:     types.MediaTypeTV,
		},
		{
			name:     "tv show with uppercase SE pattern",
			filename: "The.Office.S02E15.HDTV.avi",
			want:     types.MediaTypeTV,
		},

		// Music tests
		{
			name:     "music mp3 file",
			filename: "Artist - Song Title.mp3",
			want:     types.MediaTypeMusic,
		},
		{
			name:     "music flac file",
			filename: "01 - Track Name.flac",
			want:     types.MediaTypeMusic,
		},

		// Book tests
		{
			name:     "epub book",
			filename: "The.Great.Gatsby.epub",
			want:     types.MediaTypeBook,
		},
		{
			name:     "mobi book",
			filename: "Book.Title.mobi",
			want:     types.MediaTypeBook,
		},

		// Unknown/ambiguous tests
		{
			name:     "unknown file type",
			filename: "document.txt",
			want:     types.MediaTypeUnknown,
		},
		{
			name:     "video without clear indicators defaults to movie",
			filename: "somefile.mkv",
			want:     types.MediaTypeMovie,
		},
	}

	detector := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.Detect(tt.filename)
			if got != tt.want {
				t.Errorf("Detect(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestMovieDetector_IsMovie(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "movie with year 1999",
			filename: "The.Matrix.1999.mkv",
			want:     true,
		},
		{
			name:     "movie with year 2023",
			filename: "Movie.2023.1080p.mkv",
			want:     true,
		},
		{
			name:     "movie with old year 1927",
			filename: "Metropolis.1927.Restored.mkv",
			want:     true,
		},
		{
			name:     "movie with future year 2099",
			filename: "SciFi.Movie.2099.mkv",
			want:     true,
		},
		{
			name:     "movie with year in brackets",
			filename: "Title [2015] 1080p.mkv",
			want:     true,
		},
		{
			name:     "movie with quality tag",
			filename: "Movie.BluRay.1080p.mkv",
			want:     true,
		},
		{
			name:     "movie with h265 codec",
			filename: "Film.x265.mkv",
			want:     true,
		},
		{
			name:     "file without year or tags",
			filename: "somefile.mkv",
			want:     false,
		},
		{
			name:     "file with invalid year",
			filename: "Movie.1799.mkv",
			want:     false,
		},
		// Test years beyond 2100 (BUG-EDGE-001 fix validation)
		{
			name:     "movie with year 2101",
			filename: "Future.2101.mkv",
			want:     true,
		},
		{
			name:     "movie with year 2150",
			filename: "SciFi.Movie.2150.1080p.mkv",
			want:     true,
		},
		{
			name:     "movie with year 2199",
			filename: "Final.Film.2199.mkv",
			want:     true,
		},
	}

	detector := NewMovieDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.IsMovie(tt.filename)
			if got != tt.want {
				t.Errorf("IsMovie(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestTVDetector_IsTV(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "standard S01E01 pattern",
			filename: "Show.S01E01.mkv",
			want:     true,
		},
		{
			name:     "lowercase s01e01",
			filename: "show.name.s01e01.720p.mkv",
			want:     true,
		},
		{
			name:     "single digit season and episode",
			filename: "Series.S1E5.mkv",
			want:     true,
		},
		{
			name:     "three digit episode",
			filename: "Anime.S01E123.mkv",
			want:     true,
		},
		{
			name:     "alternative 1x01 pattern",
			filename: "Show.1x01.mkv",
			want:     true,
		},
		{
			name:     "alternative with leading zero",
			filename: "Series.01x05.mkv",
			want:     true,
		},
		{
			name:     "uppercase SE pattern",
			filename: "THE.SHOW.S02E15.mkv",
			want:     true,
		},
		{
			name:     "movie with year should not match",
			filename: "Movie.2023.mkv",
			want:     false,
		},
		{
			name:     "no season/episode pattern",
			filename: "randomfile.mkv",
			want:     false,
		},
	}

	detector := NewTVDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.IsTV(tt.filename)
			if got != tt.want {
				t.Errorf("IsTV(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestExtensionDetection(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		fn   func(string) bool
		want bool
	}{
		// Video extensions
		{"mkv is video", ".mkv", isVideoExtension, true},
		{"MP4 uppercase is video", ".MP4", isVideoExtension, true},
		{"avi is video", ".avi", isVideoExtension, true},
		{"txt is not video", ".txt", isVideoExtension, false},

		// Audio extensions
		{"mp3 is audio", ".mp3", isAudioExtension, true},
		{"FLAC uppercase is audio", ".FLAC", isAudioExtension, true},
		{"mkv is not audio", ".mkv", isAudioExtension, false},

		// Book extensions
		{"epub is book", ".epub", isBookExtension, true},
		{"PDF uppercase is book", ".PDF", isBookExtension, true},
		{"mp3 is not book", ".mp3", isBookExtension, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(tt.ext)
			if got != tt.want {
				t.Errorf("%s(%q) = %v, want %v", tt.name, tt.ext, got, tt.want)
			}
		})
	}
}
