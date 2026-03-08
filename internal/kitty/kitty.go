package kitty

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/execution"
	"github.com/crixuamg/charon/internal/tasks"
)

func findKittySocket() (string, error) {
	if socket := os.Getenv("KITTY_LISTEN_ON"); socket != "" {
		return socket, nil
	}

	matches, err := filepath.Glob("/tmp/kitty-*")
	if err != nil {
		return "", fmt.Errorf("failed to search kitty socket: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no kitty socket found")
	}

	return "unix:" + matches[0], nil
}

func OpenProject(project config.Project, cfg *config.Config) error {
	socket, err := findKittySocket()
	if err != nil {
		return err
	}

	taskList := tasks.EffectiveTasks(project, cfg)
	if len(taskList) == 0 {
		return fmt.Errorf("no tasks defined for project %q", project.Name)
	}

	execCfg := execution.Resolve(cfg, project)
	executor := execution.NewExecutor(execCfg, Terminal{})

	for _, task := range taskList {
		if err := executor.OpenTab(socket, getTabTitle(task), project, task); err != nil {
			return fmt.Errorf("open tab for %q: %w", task, err)
		}
	}

	return nil
}

func getTabTitle(task string) string {
	task = strings.TrimSpace(task)

	switch {
	case task == "db":
		return "database"
	case strings.HasPrefix(task, "lazygit"):
		return "git"
	case strings.HasPrefix(task, "vim"), strings.HasPrefix(task, "nvim"):
		return "editor"
	case strings.Contains(task, "yarn dev"), strings.Contains(task, "npm run dev"):
		return "dev"
	case strings.Contains(task, "yarn install"), strings.Contains(task, "npm install"):
		return "node"
	case strings.Contains(task, "go "):
		return "go"
	case strings.Contains(task, "php"):
		return "php"
	default:
		parts := strings.Fields(task)
		if len(parts) > 0 {
			return parts[0]
		}
		return "shell"
	}
}
