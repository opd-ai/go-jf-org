package scanner

import (
"os"
"path/filepath"
"testing"

"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestNewScanner(t *testing.T) {
videoExts := []string{".mkv", ".mp4"}
audioExts := []string{".mp3", ".flac"}
bookExts := []string{".epub", ".pdf"}
minSize := int64(1024)

s := NewScanner(videoExts, audioExts, bookExts, minSize)

if s == nil {
t.Fatal("NewScanner returned nil")
}

if len(s.videoExtensions) != 2 {
t.Errorf("Expected 2 video extensions, got %d", len(s.videoExtensions))
}

if s.minFileSize != minSize {
t.Errorf("Expected minFileSize %d, got %d", minSize, s.minFileSize)
}
}

func TestNormalizeExtensions(t *testing.T) {
tests := []struct {
name     string
input    []string
expected []string
}{
{
name:     "lowercase with dot",
input:    []string{".mkv", ".mp4"},
expected: []string{".mkv", ".mp4"},
},
{
name:     "uppercase without dot",
input:    []string{"MKV", "MP4"},
expected: []string{".mkv", ".mp4"},
},
{
name:     "mixed case with and without dot",
input:    []string{".MKV", "mp4", ".MP3"},
expected: []string{".mkv", ".mp4", ".mp3"},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := normalizeExtensions(tt.input)
if len(result) != len(tt.expected) {
t.Fatalf("Expected %d extensions, got %d", len(tt.expected), len(result))
}
for i, ext := range result {
if ext != tt.expected[i] {
t.Errorf("Expected extension %s, got %s", tt.expected[i], ext)
}
}
})
}
}

func TestIsMediaFile(t *testing.T) {
s := NewScanner(
[]string{".mkv", ".mp4"},
[]string{".mp3", ".flac"},
[]string{".epub", ".pdf"},
1024,
)

tests := []struct {
name     string
path     string
expected bool
}{
{"video mkv", "/path/to/movie.mkv", true},
{"video mp4", "/path/to/video.mp4", true},
{"audio mp3", "/path/to/song.mp3", true},
{"book epub", "/path/to/book.epub", true},
{"unknown txt", "/path/to/file.txt", false},
{"no extension", "/path/to/file", false},
{"uppercase extension", "/path/to/FILE.MKV", true},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := s.isMediaFile(tt.path)
if result != tt.expected {
t.Errorf("isMediaFile(%s) = %v, expected %v", tt.path, result, tt.expected)
}
})
}
}

func TestGetMediaType(t *testing.T) {
s := NewScanner(
[]string{".mkv", ".mp4"},
[]string{".mp3", ".flac"},
[]string{".epub", ".pdf"},
1024,
)

tests := []struct {
name     string
path     string
expected types.MediaType
}{
{"video file", "/path/to/movie.mkv", types.MediaTypeUnknown}, // Unknown because we can't distinguish movie vs TV yet
{"audio file", "/path/to/song.mp3", types.MediaTypeMusic},
{"book file", "/path/to/book.epub", types.MediaTypeBook},
{"unknown file", "/path/to/file.txt", types.MediaTypeUnknown},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := s.GetMediaType(tt.path)
if result != tt.expected {
t.Errorf("GetMediaType(%s) = %v, expected %v", tt.path, result, tt.expected)
}
})
}
}

func TestScan(t *testing.T) {
// Create a temporary test directory
tmpDir, err := os.MkdirTemp("", "scanner-test-*")
if err != nil {
t.Fatal(err)
}
defer os.RemoveAll(tmpDir)

// Create test files
testFiles := map[string]int64{
"movie.mkv":  15 * 1024 * 1024, // 15MB
"song.mp3":   5 * 1024 * 1024,  // 5MB
"book.epub":  2 * 1024 * 1024,  // 2MB
"readme.txt": 1024,              // 1KB
}

for filename, size := range testFiles {
path := filepath.Join(tmpDir, filename)
f, err := os.Create(path)
if err != nil {
t.Fatal(err)
}
if err := f.Truncate(size); err != nil {
t.Fatal(err)
}
f.Close()
}

// Create scanner with 10MB minimum size
s := NewScanner(
[]string{".mkv", ".mp4"},
[]string{".mp3", ".flac"},
[]string{".epub", ".pdf"},
10*1024*1024,
)

result, err := s.Scan(tmpDir)
if err != nil {
t.Fatalf("Scan failed: %v", err)
}

// Should find only movie.mkv (15MB > 10MB)
if len(result.Files) != 1 {
t.Errorf("Expected 1 file, got %d", len(result.Files))
}

if len(result.Files) > 0 && filepath.Base(result.Files[0]) != "movie.mkv" {
t.Errorf("Expected movie.mkv, got %s", filepath.Base(result.Files[0]))
}
}

func TestScanNonExistentDirectory(t *testing.T) {
s := NewScanner(
[]string{".mkv"},
[]string{".mp3"},
[]string{".epub"},
1024,
)

_, err := s.Scan("/non/existent/path")
if err == nil {
t.Error("Expected error for non-existent directory, got nil")
}
}
