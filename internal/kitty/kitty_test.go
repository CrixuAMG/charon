package kitty

import (
	"testing"
)

func TestGetTabTitle(t *testing.T) {
	tests := []struct {
		name     string
		task     string
		expected string
	}{
		{
			name:     "lazygit returns git",
			task:     "lazygit",
			expected: "git",
		},
		{
			name:     "lazygit with args returns git",
			task:     "lazygit --all",
			expected: "git",
		},
		{
			name:     "vim returns editor",
			task:     "vim",
			expected: "editor",
		},
		{
			name:     "nvim returns editor",
			task:     "nvim .",
			expected: "editor",
		},
		{
			name:     "yarn dev returns dev",
			task:     "yarn dev",
			expected: "dev",
		},
		{
			name:     "npm run dev returns dev",
			task:     "npm run dev",
			expected: "dev",
		},
		{
			name:     "db returns database",
			task:     "db",
			expected: "database",
		},
		{
			name:     "yarn install returns node",
			task:     "yarn install",
			expected: "node",
		},
		{
			name:     "npm install returns node",
			task:     "npm install",
			expected: "node",
		},
		{
			name:     "go run returns go",
			task:     "go run main.go",
			expected: "go",
		},
		{
			name:     "php artisan returns php",
			task:     "php artisan serve",
			expected: "php",
		},
		{
			name:     "unknown command returns first word",
			task:     "make build",
			expected: "make",
		},
		{
			name:     "handles whitespace",
			task:     "  lazygit  ",
			expected: "git",
		},
		{
			name:     "empty string returns shell",
			task:     "",
			expected: "shell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTabTitle(tt.task)
			if result != tt.expected {
				t.Errorf("getTabTitle(%q) = %q, want %q", tt.task, result, tt.expected)
			}
		})
	}
}
