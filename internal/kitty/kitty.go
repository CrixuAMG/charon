package kitty

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ckaal/charon/internal/config"
)

// findKittySocket finds the kitty socket path
func findKittySocket() (string, error) {
	// First check environment variable
	if socket := os.Getenv("KITTY_LISTEN_ON"); socket != "" {
		return socket, nil
	}

	// Look for socket files in /tmp
	matches, err := filepath.Glob("/tmp/kitty-*")
	if err != nil {
		return "", fmt.Errorf("failed to search for kitty socket: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no kitty socket found in /tmp")
	}

	return "unix:" + matches[0], nil
}

// OpenProject opens a project in kitty with tabs for each task
func OpenProject(project config.Project, cfg *config.Config) error {
	dockerPath := project.DockerPath
	if dockerPath == "" {
		dockerPath = cfg.DockerPath
	}

	useDocker := dockerPath != ""

	if len(project.Tasks) == 0 {
		return fmt.Errorf("no tasks defined for project %s", project.Name)
	}

	socket, err := findKittySocket()
	if err != nil {
		return err
	}

	for _, task := range project.Tasks {
		title := getTabTitle(task)

		var launchErr error
		if useDocker {
			projectDir := fmt.Sprintf("%s/%s", strings.TrimSuffix(dockerPath, "/"), project.Name)
			launchErr = launchDockerTab(socket, title, cfg.Container, projectDir, task)
		} else {
			launchErr = launchLocalTab(socket, title, project.Path, task)
		}

		if launchErr != nil {
			return fmt.Errorf("failed to create tab for '%s': %w", task, launchErr)
		}

		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

func sendTextToWindow(socket string, windowID int, text string) error {
	args := []string{
		"@", "--to", socket,
		"send-text", "--match", fmt.Sprintf("id:%d", windowID),
		text,
	}
	cmd := exec.Command("kitty", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func launchLocalTab(socket, title, path, task string) error {
	launchArgs := []string{
		"@", "--to", socket,
		"launch", "--type=tab", "--tab-title", title,
		"--cwd", path,
	}

	cmd := exec.Command("kitty", launchArgs...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to launch tab: %w", err)
	}

	windowID, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse window ID: %w", err)
	}

	// Wait for shell to initialize
	time.Sleep(500 * time.Millisecond)

	// Send cd command
	if err := sendTextToWindow(socket, windowID, fmt.Sprintf("cd %s\n", path)); err != nil {
		return err
	}

	// Wait for cd to complete
	time.Sleep(300 * time.Millisecond)

	// Send task command
	return sendTextToWindow(socket, windowID, task+"\n")
}

func launchDockerTab(socket, title, container, projectDir, task string) error {
	launchArgs := []string{
		"@", "--to", socket,
		"launch", "--type=tab", "--tab-title", title,
		"docker", "exec", "-it", "-w", projectDir, container,
		"sh", "-c", "exec $SHELL",
	}

	cmd := exec.Command("kitty", launchArgs...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to launch tab: %w", err)
	}

	windowID, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse window ID: %w", err)
	}

	// Wait for docker exec and shell to initialize
	time.Sleep(2000 * time.Millisecond)

	// Send cd command
	if err := sendTextToWindow(socket, windowID, fmt.Sprintf("cd %s\n", projectDir)); err != nil {
		return err
	}

	// Wait for cd to complete
	time.Sleep(500 * time.Millisecond)

	// Send task command
	return sendTextToWindow(socket, windowID, task+"\n")
}

func getTabTitle(task string) string {
	task = strings.TrimSpace(task)

	switch {
	case strings.HasPrefix(task, "lazygit"):
		return "git"
	case strings.HasPrefix(task, "vim") || strings.HasPrefix(task, "nvim"):
		return "editor"
	case strings.Contains(task, "yarn dev") || strings.Contains(task, "npm run dev"):
		return "dev"
	case strings.HasPrefix(task, "db"):
		return "database"
	case strings.Contains(task, "yarn") || strings.Contains(task, "npm"):
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
