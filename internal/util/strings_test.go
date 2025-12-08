package util

import "testing"

func TestRemoveExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"file with extension", "movie.mkv", "movie"},
		{"file with multiple dots", "The.Matrix.1999.mkv", "The.Matrix.1999"},
		{"file without extension", "noext", "noext"},
		{"hidden file", ".gitignore", ".gitignore"},
		{"file with complex name", "Show.S01E01.720p.mkv", "Show.S01E01.720p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveExtension(tt.filename)
			if got != tt.want {
				t.Errorf("RemoveExtension(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"title with dots", "The.Matrix", "The Matrix"},
		{"title with underscores", "Breaking_Bad", "Breaking Bad"},
		{"title with both", "The.Office_US", "The Office US"},
		{"title with spaces", "Already Clean", "Already Clean"},
		{"title with leading/trailing spaces", "  Padded  ", "Padded"},
		{"mixed format", "Game.of_Thrones", "Game of Thrones"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanTitle(tt.title)
			if got != tt.want {
				t.Errorf("CleanTitle(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestContainsExtension(t *testing.T) {
	extensions := []string{".mkv", ".mp4", ".avi"}

	tests := []struct {
		name string
		ext  string
		want bool
	}{
		{"exact match lowercase", ".mkv", true},
		{"uppercase extension", ".MKV", true},
		{"mixed case", ".Mp4", true},
		{"not in list", ".txt", false},
		{"empty extension", "", false},
		{"without dot", "mkv", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsExtension(extensions, tt.ext)
			if got != tt.want {
				t.Errorf("ContainsExtension(extensions, %q) = %v, want %v", tt.ext, got, tt.want)
			}
		})
	}
}
