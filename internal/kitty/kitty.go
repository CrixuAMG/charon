package kitty

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ckaal/charon/internal/config"
)

// OpenProject opens a project in kitty with tabs for each task
func OpenProject(project config.Project, cfg *config.Config) error {
	dockerPath := project.DockerPath
	if dockerPath == "" {
		dockerPath = cfg.DockerPath
	}

	useDocker := dockerPath != ""

	// Build all commands
	var commands []struct {
		title string
		cmd   string
	}

	for _, task := range project.Tasks {
		var cmd string
		if useDocker {
			projectDir := fmt.Sprintf("%s/%s", strings.TrimSuffix(dockerPath, "/"), project.Name)
			// Use docker exec to run commands inside the container
			containerCmd := fmt.Sprintf("cd %s && %s", projectDir, task)
			cmd = fmt.Sprintf("docker exec -it -w %s %s bash -ic '%s'",
				projectDir,
				cfg.Container,
				strings.ReplaceAll(containerCmd, "'", "'\\''"))
		} else {
			cmd = fmt.Sprintf("cd %s && %s", project.Path, task)
		}

		commands = append(commands, struct {
			title string
			cmd   string
		}{
			title: getTabTitle(task),
			cmd:   cmd,
		})
	}

	if len(commands) == 0 {
		return fmt.Errorf("no tasks defined for project %s", project.Name)
	}

	// Try to find kitty socket
	socket := os.Getenv("KITTY_LISTEN_ON")
	if socket == "" {
		// Fallback to common socket location
		socket = "unix:/tmp/kitty"
	}

	// Try remote control first
	if err := openWithRemoteControl(socket, commands); err == nil {
		return nil
	}

	// Fallback: launch new kitty window with session
	return openNewKittyWindow(project.Name, commands)
}

func openWithRemoteControl(socket string, commands []struct{ title, cmd string }) error {
	for _, c := range commands {
		// All tabs: add to current window
		args := []string{
			"@", "--to", socket,
			"launch", "--type=tab", "--title", c.title,
			"bash", "-ic", c.cmd,
		}

		cmd := exec.Command("kitty", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create tab '%s': %w", c.title, err)
		}

		// Small delay to ensure tabs are created in order
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func openNewKittyWindow(projectName string, commands []struct{ title, cmd string }) error {
	if len(commands) == 0 {
		return nil
	}

	// Create a temporary session file
	var sessionContent strings.Builder
	for _, c := range commands {
		sessionContent.WriteString(fmt.Sprintf("new_tab %s\n", c.title))
		escapedCmd := strings.ReplaceAll(c.cmd, "'", "'\\''")
		sessionContent.WriteString(fmt.Sprintf("launch bash -ic '%s'\n", escapedCmd))
	}

	// Write session file
	tmpFile := fmt.Sprintf("/tmp/charon-session-%d.conf", time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, []byte(sessionContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	// Launch kitty with session
	args := []string{
		"--session", tmpFile,
		"--title", projectName,
	}

	cmd := exec.Command("kitty", args...)
	if err := cmd.Start(); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to launch kitty: %w", err)
	}

	// Clean up session file after a delay
	go func() {
		time.Sleep(2 * time.Second)
		os.Remove(tmpFile)
	}()

	return nil
}

func getTabTitle(task string) string {
	task = strings.TrimSpace(task)

	switch {
	case strings.HasPrefix(task, "git"):
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
