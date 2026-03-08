package execution

import "github.com/crixuamg/charon/internal/config"

// Resolve returns the effective execution configuration for a project,
// falling back from project-level to global config, then to "local".
func Resolve(cfg *config.Config, project config.Project) config.Execution {
	if project.Execution != nil {
		return *project.Execution
	}

	if cfg.Execution != nil {
		return *cfg.Execution
	}

	return config.Execution{Type: "local"}
}
