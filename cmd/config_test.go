package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigInit(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".go-jf-org")
	configFile := filepath.Join(configDir, "config.yaml")

	tests := []struct {
		name        string
		setupFunc   func()
		force       bool
		expectError bool
		checkFunc   func(t *testing.T)
	}{
		{
			name:        "create config file successfully",
			setupFunc:   func() {},
			force:       false,
			expectError: false,
			checkFunc: func(t *testing.T) {
				// Check config file exists
				if _, err := os.Stat(configFile); os.IsNotExist(err) {
					t.Error("Config file was not created")
				}

				// Check cache directory exists
				cacheDir := filepath.Join(configDir, "cache")
				if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
					t.Error("Cache directory was not created")
				}

				// Check transaction directory exists
				txnDir := filepath.Join(configDir, "txn")
				if _, err := os.Stat(txnDir); os.IsNotExist(err) {
					t.Error("Transaction directory was not created")
				}

				// Check config file content
				content, err := os.ReadFile(configFile)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}

				// Verify key sections exist
				contentStr := string(content)
				requiredSections := []string{
					"sources:",
					"destinations:",
					"api_keys:",
					"organize:",
					"safety:",
					"conflict_resolution:",
				}

				for _, section := range requiredSections {
					if !strings.Contains(contentStr, section) {
						t.Errorf("Config file missing section: %s", section)
					}
				}
			},
		},
		{
			name: "fail when config exists without force",
			setupFunc: func() {
				// Create config file first
				os.MkdirAll(configDir, 0755)
				os.WriteFile(configFile, []byte("existing"), 0644)
			},
			force:       false,
			expectError: true,
			checkFunc: func(t *testing.T) {
				// Check that original file wasn't overwritten
				content, err := os.ReadFile(configFile)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}
				if string(content) != "existing" {
					t.Error("Config file was overwritten without --force")
				}
			},
		},
		{
			name: "overwrite config with force flag",
			setupFunc: func() {
				// Create config file first
				os.MkdirAll(configDir, 0755)
				os.WriteFile(configFile, []byte("existing"), 0644)
			},
			force:       true,
			expectError: false,
			checkFunc: func(t *testing.T) {
				// Check that file was overwritten
				content, err := os.ReadFile(configFile)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}
				if string(content) == "existing" {
					t.Error("Config file was not overwritten with --force")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.RemoveAll(configDir)

			// Run setup
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Set force flag
			configInitForce = tt.force

			// Run config init
			err := runConfigInit(configInitCmd, []string{})

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Run additional checks
			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}
