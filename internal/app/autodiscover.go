package app

import (
	"os"

	"github.com/crixuamg/charon/internal/autodiscover"
	"github.com/crixuamg/charon/internal/config"
)

func AutoDiscover(cfg *config.Config) error {
	paths, err := autodiscover.Scan(cfg.Scan.Paths)
	if err != nil {
		return err
	}

	added := false

	for _, path := range paths {
		if cfg.HasProjectPath(path) {
			continue
		}

		p := config.ProjectFromPath(path)
		p.Exists = true
		cfg.AddProject(p)

		added = true
	}

	if added {
		return config.Save(cfg)
	}

	return nil
}

func UpdateProjectStatus(cfg *config.Config) {
	for i := range cfg.Projects {
		_, err := os.Stat(cfg.Projects[i].Path)
		cfg.Projects[i].Exists = err == nil
	}
}

