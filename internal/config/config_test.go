package config

import (
"os"
"path/filepath"
"testing"
)

func TestDefaultConfig(t *testing.T) {
cfg := DefaultConfig()

if cfg == nil {
t.Fatal("DefaultConfig returned nil")
}

// Check video extensions
if len(cfg.Filters.VideoExtensions) == 0 {
t.Error("Expected video extensions to be populated")
}

// Check audio extensions
if len(cfg.Filters.AudioExtensions) == 0 {
t.Error("Expected audio extensions to be populated")
}

// Check book extensions
if len(cfg.Filters.BookExtensions) == 0 {
t.Error("Expected book extensions to be populated")
}

// Check default organize settings
if !cfg.Organize.CreateNFO {
t.Error("Expected CreateNFO to be true by default")
}

// Check safety settings
if cfg.Safety.ConflictResolution != "skip" {
t.Errorf("Expected ConflictResolution to be 'skip', got '%s'", cfg.Safety.ConflictResolution)
}
}

func TestLoad_NoConfigFile(t *testing.T) {
// Load config when no config file exists
// Should use defaults
cfg, err := Load("/nonexistent/config.yaml")
if err != nil {
t.Fatalf("Load failed: %v", err)
}

// Check that defaults were applied
if len(cfg.Filters.VideoExtensions) == 0 {
t.Error("Expected default video extensions to be applied")
}

if len(cfg.Filters.AudioExtensions) == 0 {
t.Error("Expected default audio extensions to be applied")
}
}

func TestLoad_WithConfigFile(t *testing.T) {
// Create a temporary config file
tmpDir, err := os.MkdirTemp("", "config-test-*")
if err != nil {
t.Fatal(err)
}
defer os.RemoveAll(tmpDir)

configPath := filepath.Join(tmpDir, "config.yaml")
configContent := []byte(`
sources:
  - /test/source

destinations:
  movies: /test/movies
  tv: /test/tv
  music: /test/music
  books: /test/books

organize:
  create_nfo: false

safety:
  conflict_resolution: rename
`)

if err := os.WriteFile(configPath, configContent, 0644); err != nil {
t.Fatal(err)
}

cfg, err := Load(configPath)
if err != nil {
t.Fatalf("Load failed: %v", err)
}

// Check loaded values
if len(cfg.Sources) != 1 || cfg.Sources[0] != "/test/source" {
t.Error("Sources not loaded correctly")
}

if cfg.Destinations.Movies != "/test/movies" {
t.Error("Destinations.Movies not loaded correctly")
}

if cfg.Organize.CreateNFO != false {
t.Error("Organize.CreateNFO should be false")
}

if cfg.Safety.ConflictResolution != "rename" {
t.Errorf("Safety.ConflictResolution should be 'rename', got '%s'", cfg.Safety.ConflictResolution)
}

// Check that defaults were still applied for unspecified values
if len(cfg.Filters.VideoExtensions) == 0 {
t.Error("Default video extensions should still be applied")
}
}

func TestLoad_InvalidYAML(t *testing.T) {
tmpDir, err := os.MkdirTemp("", "config-test-*")
if err != nil {
t.Fatal(err)
}
defer os.RemoveAll(tmpDir)

configPath := filepath.Join(tmpDir, "config.yaml")
invalidContent := []byte(`
this is not: valid: yaml: content
  broken indentation
`)

if err := os.WriteFile(configPath, invalidContent, 0644); err != nil {
t.Fatal(err)
}

_, err = Load(configPath)
if err == nil {
t.Error("Expected error for invalid YAML, got nil")
}
}
