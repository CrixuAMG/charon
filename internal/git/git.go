package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Info holds git metadata about a directory.
type Info struct {
	IsRepo     bool
	Branch     string
	HasChanges bool
}

// GetInfo returns git information for the given path.
// If the path is not a git repository, IsRepo will be false.
func GetInfo(path string) Info {
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return Info{}
	}

	branch := branch(path)
	hasChanges := hasUncommittedChanges(path)

	return Info{
		IsRepo:     true,
		Branch:     branch,
		HasChanges: hasChanges,
	}
}

// Branch returns the current git branch for the given path, or empty string
// if the path is not a git repository or git is unavailable.
func Branch(path string) string {
	return branch(path)
}

func branch(path string) string {
	out, err := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func hasUncommittedChanges(path string) bool {
	out, err := exec.Command("git", "-C", path, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}
