package kitty

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type Terminal struct{}

func (Terminal) LaunchLocalTab(
	socket string,
	title string,
	path string,
) (int, error) {
	cmd := exec.Command(
		"kitty", "@", "--to", socket,
		"launch", "--type=tab",
		"--tab-title", title,
		"--cwd", path,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("launch local tab: %w", err)
	}

	id, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("parse window id from output %q: %w", string(output), err)
	}

	return id, nil
}

func (Terminal) LaunchDockerTab(
	socket string,
	title string,
	container string,
	workdir string,
) (int, error) {
	args := []string{
		"@", "--to", socket,
		"launch", "--type=tab",
		"--tab-title", title,
		"docker", "exec", "-it",
		"-w", workdir,
		container,
		"/bin/bash", "--login",
	}

	cmd := exec.Command("kitty", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("launch docker tab: %w (output: %s)", err, string(output))
	}

	id, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("parse window id from output %q: %w", string(output), err)
	}

	return id, nil
}

func (Terminal) SendText(
	socket string,
	windowID int,
	text string,
) error {
	cmd := exec.Command(
		"kitty", "@", "--to", socket,
		"send-text",
		"--match", fmt.Sprintf("id:%d", windowID),
		text,
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
