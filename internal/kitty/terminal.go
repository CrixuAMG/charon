package kitty

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
		return 0, err
	}

	// optional delay for shell init
	time.Sleep(300 * time.Millisecond)

	return id, nil
}

func (Terminal) LaunchDockerTab(
	socket string,
	title string,
	container string,
	workdir string,
) (int, error) {

	cmd := exec.Command(
		"kitty", "@", "--to", socket,
		"launch", "--type=tab",
		"--tab-title", title,
		"docker", "exec", "-it",
		"-w", workdir,
		container,
		"sh", "-c", "exec $SHELL",
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("launch docker tab: %w", err)
	}

	id, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}

	// longer wait for docker init
	time.Sleep(1 * time.Second)

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
