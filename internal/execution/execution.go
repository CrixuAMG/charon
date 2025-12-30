package execution

import "github.com/crixuamg/charon/internal/config"

func Resolve(cfg *config.Config, project config.Project) config.Execution {
	// if task.Execution != nil {
	// 	return *task.Execution
	// }

	if project.Execution != nil {
		return *project.Execution
	}

	if cfg.Execution != nil {
		return *cfg.Execution
	}

	return config.Execution{Type: "local"}
}
