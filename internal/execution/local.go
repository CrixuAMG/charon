package execution

import (
	"time"

	"github.com/crixuamg/charon/internal/config"
)

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

	time.Sleep(300 * time.Millisecond)

	return e.Terminal.SendText(socket, id, task+"\n")
}
