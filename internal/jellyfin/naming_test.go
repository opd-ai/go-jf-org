package jellyfin

import (
	"path/filepath"
	"testing"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestGetMovieName(t *testing.T) {
	n := NewNaming()

	tests := []struct {
		name     string
		metadata *types.Metadata
		ext      string
		want     string
	}{
		{
			name: "movie with year",
			metadata: &types.Metadata{
				Title: "The Matrix",
				Year:  1999,
			},
			ext:  ".mkv",
			want: "The Matrix (1999).mkv",
		},
		{
			name: "movie without year",
			metadata: &types.Metadata{
				Title: "Unknown Movie",
			},
			ext:  ".mp4",
			want: "Unknown Movie.mp4",
		},
		{
			name: "movie with special characters",
			metadata: &types.Metadata{
				Title: "Movie: The \"Best\" Part",
				Year:  2023,
			},
			ext:  ".mkv",
			want: "Movie - The 'Best' Part (2023).mkv",
		},
		{
			name:     "nil metadata",
			metadata: nil,
			ext:      ".mkv",
			want:     "",
		},
		{
			name: "empty title",
			metadata: &types.Metadata{
				Title: "",
				Year:  2020,
			},
			ext:  ".mkv",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.GetMovieName(tt.metadata, tt.ext)
			if got != tt.want {
				t.Errorf("GetMovieName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetMovieDir(t *testing.T) {
	n := NewNaming()

	tests := []struct {
		name     string
		metadata *types.Metadata
		want     string
	}{
		{
			name: "with year",
			metadata: &types.Metadata{
				Title: "Inception",
				Year:  2010,
			},
			want: "Inception (2010)",
		},
		{
			name: "without year",
			metadata: &types.Metadata{
				Title: "Some Movie",
			},
			want: "Some Movie",
		},
		{
			name:     "nil metadata",
			metadata: nil,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.GetMovieDir(tt.metadata)
			if got != tt.want {
				t.Errorf("GetMovieDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTVShowName(t *testing.T) {
	n := NewNaming()

	tests := []struct {
		name     string
		metadata *types.Metadata
		ext      string
		want     string
	}{
		{
			name: "with episode title",
			metadata: &types.Metadata{
				TVMetadata: &types.TVMetadata{
					ShowTitle:    "Breaking Bad",
					Season:       1,
					Episode:      1,
					EpisodeTitle: "Pilot",
				},
			},
			ext:  ".mkv",
			want: "Breaking Bad - S01E01 - Pilot.mkv",
		},
		{
			name: "without episode title",
			metadata: &types.Metadata{
				TVMetadata: &types.TVMetadata{
					ShowTitle: "Game of Thrones",
					Season:    5,
					Episode:   9,
				},
			},
			ext:  ".mp4",
			want: "Game of Thrones - S05E09.mp4",
		},
		{
			name: "season 0 (specials)",
			metadata: &types.Metadata{
				TVMetadata: &types.TVMetadata{
					ShowTitle:    "Doctor Who",
					Season:       0,
					Episode:      1,
					EpisodeTitle: "Christmas Special",
				},
			},
			ext:  ".mkv",
			want: "Doctor Who - S00E01 - Christmas Special.mkv",
		},
		{
			name:     "nil TV metadata",
			metadata: &types.Metadata{},
			ext:      ".mkv",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.GetTVShowName(tt.metadata, tt.ext)
			if got != tt.want {
				t.Errorf("GetTVShowName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTVSeasonDir(t *testing.T) {
	n := NewNaming()

	tests := []struct {
		name   string
		season int
		want   string
	}{
		{
			name:   "regular season",
			season: 1,
			want:   "Season 01",
		},
		{
			name:   "double digit season",
			season: 12,
			want:   "Season 12",
		},
		{
			name:   "specials",
			season: 0,
			want:   "Specials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.GetTVSeasonDir(tt.season)
			if got != tt.want {
				t.Errorf("GetTVSeasonDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetMusicDir(t *testing.T) {
	n := NewNaming()

	tests := []struct {
		name        string
		metadata    *types.Metadata
		wantArtist  string
		wantAlbum   string
	}{
		{
			name: "with year",
			metadata: &types.Metadata{
				Year: 1973,
				MusicMetadata: &types.MusicMetadata{
					Artist: "Pink Floyd",
					Album:  "The Dark Side of the Moon",
				},
			},
			wantArtist: "Pink Floyd",
			wantAlbum:  "The Dark Side of the Moon (1973)",
		},
		{
			name: "without year",
			metadata: &types.Metadata{
				MusicMetadata: &types.MusicMetadata{
					Artist: "The Beatles",
					Album:  "Abbey Road",
				},
			},
			wantArtist: "The Beatles",
			wantAlbum:  "Abbey Road",
		},
		{
			name: "unknown artist",
			metadata: &types.Metadata{
				MusicMetadata: &types.MusicMetadata{
					Album: "Compilation",
				},
			},
			wantArtist: "Unknown Artist",
			wantAlbum:  "Compilation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArtist, gotAlbum := n.GetMusicDir(tt.metadata)
			if gotArtist != tt.wantArtist {
				t.Errorf("GetMusicDir() artist = %q, want %q", gotArtist, tt.wantArtist)
			}
			if gotAlbum != tt.wantAlbum {
				t.Errorf("GetMusicDir() album = %q, want %q", gotAlbum, tt.wantAlbum)
			}
		})
	}
}

func TestGetMusicTrackName(t *testing.T) {
	n := NewNaming()

	tests := []struct {
		name     string
		metadata *types.Metadata
		ext      string
		want     string
	}{
		{
			name: "with track number",
			metadata: &types.Metadata{
				Title: "Speak to Me",
				MusicMetadata: &types.MusicMetadata{
					TrackNumber: 1,
				},
			},
			ext:  ".mp3",
			want: "01 - Speak to Me.mp3",
		},
		{
			name: "without track number",
			metadata: &types.Metadata{
				Title: "Some Song",
				MusicMetadata: &types.MusicMetadata{},
			},
			ext:  ".flac",
			want: "Some Song.flac",
		},
		{
			name: "double digit track",
			metadata: &types.Metadata{
				Title: "Track Name",
				MusicMetadata: &types.MusicMetadata{
					TrackNumber: 12,
				},
			},
			ext:  ".m4a",
			want: "12 - Track Name.m4a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.GetMusicTrackName(tt.metadata, tt.ext)
			if got != tt.want {
				t.Errorf("GetMusicTrackName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetBookDir(t *testing.T) {
	n := NewNaming()

	tests := []struct {
		name       string
		metadata   *types.Metadata
		wantAuthor string
		wantBook   string
	}{
		{
			name: "with year and full name",
			metadata: &types.Metadata{
				Title: "The Great Gatsby",
				Year:  1925,
				BookMetadata: &types.BookMetadata{
					Author: "F. Scott Fitzgerald",
				},
			},
			wantAuthor: "Fitzgerald, F. Scott",
			wantBook:   "The Great Gatsby (1925)",
		},
		{
			name: "single name author",
			metadata: &types.Metadata{
				Title: "1984",
				Year:  1949,
				BookMetadata: &types.BookMetadata{
					Author: "Orwell",
				},
			},
			wantAuthor: "Orwell",
			wantBook:   "1984 (1949)",
		},
		{
			name: "already formatted author",
			metadata: &types.Metadata{
				Title: "To Kill a Mockingbird",
				BookMetadata: &types.BookMetadata{
					Author: "Lee, Harper",
				},
			},
			wantAuthor: "Lee, Harper",
			wantBook:   "To Kill a Mockingbird",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAuthor, gotBook := n.GetBookDir(tt.metadata)
			if gotAuthor != tt.wantAuthor {
				t.Errorf("GetBookDir() author = %q, want %q", gotAuthor, tt.wantAuthor)
			}
			if gotBook != tt.wantBook {
				t.Errorf("GetBookDir() book = %q, want %q", gotBook, tt.wantBook)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "with colon",
			input: "Movie: The Sequel",
			want:  "Movie - The Sequel",
		},
		{
			name:  "with quotes",
			input: `The "Best" Movie`,
			want:  "The 'Best' Movie",
		},
		{
			name:  "with slashes",
			input: "Movie/Part 1",
			want:  "Movie-Part 1",
		},
		{
			name:  "with multiple invalid chars",
			input: `Movie<>:"/\|?*`,
			want:  "Movie -'---",
		},
		{
			name:  "with leading/trailing dots",
			input: "...Movie...",
			want:  "Movie",
		},
		{
			name:  "with extra spaces",
			input: "Movie    With    Spaces",
			want:  "Movie With Spaces",
		},
		{
			name:  "normal filename",
			input: "The Matrix",
			want:  "The Matrix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatAuthorName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "first and last name",
			input: "F. Scott Fitzgerald",
			want:  "Fitzgerald, F. Scott",
		},
		{
			name:  "single name",
			input: "Orwell",
			want:  "Orwell",
		},
		{
			name:  "already formatted",
			input: "Lee, Harper",
			want:  "Lee, Harper",
		},
		{
			name:  "three part name",
			input: "J. R. R. Tolkien",
			want:  "Tolkien, J. R. R.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAuthorName(tt.input)
			if got != tt.want {
				t.Errorf("FormatAuthorName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildFullPath(t *testing.T) {
	n := NewNaming()

	tests := []struct {
		name      string
		destRoot  string
		mediaType types.MediaType
		metadata  *types.Metadata
		ext       string
		want      string
	}{
		{
			name:      "movie",
			destRoot:  "/media/movies",
			mediaType: types.MediaTypeMovie,
			metadata: &types.Metadata{
				Title: "The Matrix",
				Year:  1999,
			},
			ext:  ".mkv",
			want: filepath.Join("/media/movies", "The Matrix (1999)", "The Matrix (1999).mkv"),
		},
		{
			name:      "tv show",
			destRoot:  "/media/tv",
			mediaType: types.MediaTypeTV,
			metadata: &types.Metadata{
				TVMetadata: &types.TVMetadata{
					ShowTitle:    "Breaking Bad",
					Season:       1,
					Episode:      1,
					EpisodeTitle: "Pilot",
				},
			},
			ext:  ".mkv",
			want: filepath.Join("/media/tv", "Breaking Bad", "Season 01", "Breaking Bad - S01E01 - Pilot.mkv"),
		},
		{
			name:      "music",
			destRoot:  "/media/music",
			mediaType: types.MediaTypeMusic,
			metadata: &types.Metadata{
				Title: "Speak to Me",
				Year:  1973,
				MusicMetadata: &types.MusicMetadata{
					Artist:      "Pink Floyd",
					Album:       "The Dark Side of the Moon",
					TrackNumber: 1,
				},
			},
			ext:  ".mp3",
			want: filepath.Join("/media/music", "Pink Floyd", "The Dark Side of the Moon (1973)", "01 - Speak to Me.mp3"),
		},
		{
			name:      "book",
			destRoot:  "/media/books",
			mediaType: types.MediaTypeBook,
			metadata: &types.Metadata{
				Title: "The Great Gatsby",
				Year:  1925,
				BookMetadata: &types.BookMetadata{
					Author: "F. Scott Fitzgerald",
				},
			},
			ext:  ".epub",
			want: filepath.Join("/media/books", "Fitzgerald, F. Scott", "The Great Gatsby (1925)", "The Great Gatsby.epub"),
		},
		{
			name:      "unknown type",
			destRoot:  "/media",
			mediaType: types.MediaTypeUnknown,
			metadata: &types.Metadata{
				Title: "Something",
			},
			ext:  ".file",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.BuildFullPath(tt.destRoot, tt.mediaType, tt.metadata, tt.ext)
			if got != tt.want {
				t.Errorf("BuildFullPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
