package execution

import (
	"time"

	"github.com/crixuamg/charon/internal/config"
)

// shellInitDelay gives the shell time to initialize before sending commands.
const shellInitDelay = 300 * time.Millisecond

type LocalExecutor struct {
	Terminal Terminal
}

func (e LocalExecutor) OpenTab(
	socket string,
	title string,
	project config.Project,
	task string,
) error {
	id, err := e.Terminal.LaunchLocalTab(socket, title, project.Path)
	if err != nil {
		return err
	}

	time.Sleep(shellInitDelay)

	return e.Terminal.SendText(socket, id, task+"\n")
}
