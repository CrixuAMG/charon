package execution

import (
	"time"

	"github.com/crixuamg/charon/internal/config"
)

// dockerStartupDelay gives the docker exec session time to be ready
// before sending commands.
const dockerStartupDelay = 1 * time.Second

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

	time.Sleep(dockerStartupDelay)

	return e.Terminal.SendText(socket, id, task+"\n")
}
