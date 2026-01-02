package execution

import "github.com/crixuamg/charon/internal/config"

type Executor interface {
	OpenTab(
		socket string,
		title string,
		project config.Project,
		task string,
	) error
}

func NewExecutor(exec config.Execution, term Terminal) Executor {
	switch exec.Type {
	case "docker":
		return DockerExecutor{
			Container: exec.Container,
			Terminal:  term,
		}
	default:
		return LocalExecutor{
			Terminal: term,
		}
	}
}
