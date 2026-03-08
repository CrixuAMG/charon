package execution

import (
	"fmt"
	"time"

	"github.com/crixuamg/charon/internal/config"
)

type SSHExecutor struct {
	Host     string
	Path     string
	Terminal Terminal
}

func (e SSHExecutor) OpenTab(
	socket string,
	title string,
	project config.Project,
	task string,
) error {
	id, err := e.Terminal.LaunchSSHTab(socket, title, e.Host, e.Path)
	if err != nil {
		return fmt.Errorf("ssh tab: %w", err)
	}

	time.Sleep(shellInitDelay)

	return e.Terminal.SendText(socket, id, task+"\n")
}
