package execution

import (
	"time"

	"github.com/crixuamg/charon/internal/config"
)

type DockerExecutor struct {
	Container string
	Terminal  Terminal
}

func (e DockerExecutor) OpenTab(
	socket string,
	title string,
	project config.Project,
	task string,
) error {
	id, err := e.Terminal.LaunchDockerTab(socket, title, e.Container, project.Path)
	if err != nil {
		return err
	}

	time.Sleep(300 * time.Millisecond)

	return e.Terminal.SendText(socket, id, task+"\n")
}
