package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expands tilde to home directory",
			input:    "~/projects",
			expected: filepath.Join(homeDir, "projects"),
		},
		{
			name:     "leaves absolute path unchanged",
			input:    "/var/www/html",
			expected: "/var/www/html",
		},
		{
			name:     "leaves relative path unchanged",
			input:    "projects/myapp",
			expected: "projects/myapp",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContractPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "contracts home directory to tilde",
			input:    filepath.Join(homeDir, "projects"),
			expected: "~/projects",
		},
		{
			name:     "leaves non-home path unchanged",
			input:    "/var/www/html",
			expected: "/var/www/html",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contractPath(tt.input)
			if result != tt.expected {
				t.Errorf("contractPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadNonExistentConfig(t *testing.T) {
	// Temporarily change home to a temp directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load() returned error for non-existent config: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if len(cfg.Projects) != 0 {
		t.Errorf("Expected empty projects, got %d", len(cfg.Projects))
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use temp directory as home
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create a config to save
	cfg := &Config{
		DockerPath: "/var/www/html",
		Container:  "dev-container",
		Projects: []Project{
			{
				Name:       "test-project",
				Path:       "/home/user/projects/test",
				DockerPath: "/var/www/html",
				Tasks:      []string{"vim", "lazygit"},
			},
		},
	}

	// Save the config
	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Load it back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify the loaded config
	if loaded.DockerPath != cfg.DockerPath {
		t.Errorf("DockerPath = %q, want %q", loaded.DockerPath, cfg.DockerPath)
	}

	if loaded.Container != cfg.Container {
		t.Errorf("Container = %q, want %q", loaded.Container, cfg.Container)
	}

	if len(loaded.Projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(loaded.Projects))
	}

	if loaded.Projects[0].Name != "test-project" {
		t.Errorf("Project name = %q, want %q", loaded.Projects[0].Name, "test-project")
	}

	if len(loaded.Projects[0].Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(loaded.Projects[0].Tasks))
	}
}
