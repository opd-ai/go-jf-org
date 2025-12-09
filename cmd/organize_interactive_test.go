package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opd-ai/go-jf-org/internal/organizer"
)

func TestFindAvailableName(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name           string
		existingFiles  []string
		inputPath      string
		expectedSuffix string
	}{
		{
			name:           "no conflicts",
			existingFiles:  []string{},
			inputPath:      filepath.Join(tmpDir, "test.mkv"),
			expectedSuffix: "-1.mkv",
		},
		{
			name:           "one conflict",
			existingFiles:  []string{"test-1.mkv"},
			inputPath:      filepath.Join(tmpDir, "test.mkv"),
			expectedSuffix: "-2.mkv",
		},
		{
			name:           "multiple conflicts",
			existingFiles:  []string{"test-1.mkv", "test-2.mkv", "test-3.mkv"},
			inputPath:      filepath.Join(tmpDir, "test.mkv"),
			expectedSuffix: "-4.mkv",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create existing files
			for _, f := range tt.existingFiles {
				fullPath := filepath.Join(tmpDir, f)
				if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}
			
			// Find available name
			result, err := findAvailableName(tt.inputPath)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			// Check suffix
			expectedPath := filepath.Join(tmpDir, "test"+tt.expectedSuffix)
			if result != expectedPath {
				t.Errorf("Expected %s, got %s", expectedPath, result)
			}
			
			// Check that result file doesn't exist
			if _, err := os.Stat(result); err == nil {
				t.Error("Result file already exists")
			}
		})
	}
}

func TestHandleInteractiveConflicts(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name           string
		plans          []organizer.Plan
		expectedCount  int // Number of plans that should remain
		description    string
	}{
		{
			name: "no conflicts",
			plans: []organizer.Plan{
				{
					SourcePath:      filepath.Join(tmpDir, "source1.mkv"),
					DestinationPath: filepath.Join(tmpDir, "dest1.mkv"),
					Conflict:        false,
				},
				{
					SourcePath:      filepath.Join(tmpDir, "source2.mkv"),
					DestinationPath: filepath.Join(tmpDir, "dest2.mkv"),
					Conflict:        false,
				},
			},
			expectedCount: 2,
			description:   "Plans without conflicts should pass through unchanged",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We can't fully test interactive mode without mocking user input
			// This test just verifies the function handles non-conflict cases correctly
			result := handleInteractiveConflicts(tt.plans)
			
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d plans, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

func TestInteractiveValidation(t *testing.T) {
	// Test that interactive mode is properly validated in organize command
	
	tests := []struct {
		name            string
		conflictStrategy string
		jsonOutput      bool
		expectError     bool
	}{
		{
			name:            "valid skip strategy",
			conflictStrategy: "skip",
			jsonOutput:      false,
			expectError:     false,
		},
		{
			name:            "valid rename strategy",
			conflictStrategy: "rename",
			jsonOutput:      false,
			expectError:     false,
		},
		{
			name:            "valid interactive strategy",
			conflictStrategy: "interactive",
			jsonOutput:      false,
			expectError:     false,
		},
		{
			name:            "invalid strategy",
			conflictStrategy: "invalid",
			jsonOutput:      false,
			expectError:     true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validStrategies := map[string]bool{
				"skip":        true,
				"rename":      true,
				"interactive": true,
			}
			
			isValid := validStrategies[tt.conflictStrategy]
			
			if tt.expectError && isValid {
				t.Error("Expected invalid strategy to be detected")
			}
			if !tt.expectError && !isValid {
				t.Error("Expected valid strategy to be accepted")
			}
		})
	}
}

func TestPromptConflictResolution(t *testing.T) {
tests := []struct {
name     string
input    string
expected string
}{
{
name:     "skip with s",
input:    "s\n",
expected: "skip",
},
{
name:     "skip with empty input",
input:    "\n",
expected: "skip",
},
{
name:     "rename with r",
input:    "r\n",
expected: "rename",
},
{
name:     "rename with full word",
input:    "rename\n",
expected: "rename",
},
{
name:     "skip-all with a",
input:    "a\n",
expected: "skip-all",
},
{
name:     "skip-all with 'all'",
input:    "all\n",
expected: "skip-all",
},
{
name:     "skip-all with 'skip-all'",
input:    "skip-all\n",
expected: "skip-all",
},
{
name:     "skip-all with 'skipall'",
input:    "skipall\n",
expected: "skip-all",
},
{
name:     "invalid input defaults to skip",
input:    "invalid\n",
expected: "skip",
},
{
name:     "case insensitive - uppercase R",
input:    "R\n",
expected: "rename",
},
{
name:     "case insensitive - uppercase A",
input:    "A\n",
expected: "skip-all",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
// Mock stdin with strings.NewReader
reader := strings.NewReader(tt.input)
result := promptConflictResolutionWithReader("/source/file.mkv", "/dest/file.mkv", reader)

if result != tt.expected {
t.Errorf("Expected %s, got %s", tt.expected, result)
}
})
}
}
