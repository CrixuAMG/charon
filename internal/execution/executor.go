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

func NewExecutorForProject(exec config.Execution, project config.Project, term Terminal) Executor {
	switch exec.Type {
	case "docker":
		return DockerExecutor{
			Container: exec.Container,
			Terminal:  term,
		}
	case "ssh":
		if project.SSH != nil {
			return SSHExecutor{
				Host:     project.SSH.Host,
				Path:     project.SSH.Path,
				Terminal: term,
			}
		}
		return LocalExecutor{Terminal: term}
	default:
		return LocalExecutor{
			Terminal: term,
		}
	}
}

// NewExecutor is retained for backwards compatibility.
func NewExecutor(exec config.Execution, term Terminal) Executor {
	return NewExecutorForProject(exec, config.Project{}, term)
}
