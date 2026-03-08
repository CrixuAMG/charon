package browser

import (
	"os/exec"
	"runtime"
)

// Open opens path in the OS file browser (Finder on macOS, xdg-open on Linux).
// The call returns immediately; the file manager opens in the background.
func Open(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
