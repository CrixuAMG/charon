package app

import (
	"github.com/crixuamg/charon/internal/autodiscover"
	"github.com/crixuamg/charon/internal/config"
)

func AutoDiscover(cfg *config.Config) error {
	paths, err := autodiscover.Scan(cfg.Scan.Paths)
	if err != nil {
		return err
	}

	for _, path := range paths {
		if cfg.HasProjectPath(path) {
			continue
		}

		p := config.ProjectFromPath(path)
		cfg.AddProject(p)
	}

	return config.Save(cfg)
}

