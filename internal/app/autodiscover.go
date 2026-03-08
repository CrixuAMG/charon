package app

import (
	"os"
	"sync"

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

// UpdateProjectStatus checks whether each project path exists on disk.
// Checks run concurrently so startup stays fast with many projects.
func UpdateProjectStatus(cfg *config.Config) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	type result struct {
		index  int
		exists bool
	}

	results := make([]result, len(cfg.Projects))

	for i, p := range cfg.Projects {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			_, err := os.Stat(path)
			mu.Lock()
			results[idx] = result{index: idx, exists: err == nil}
			mu.Unlock()
		}(i, p.Path)
	}

	wg.Wait()

	for _, r := range results {
		cfg.Projects[r.index].Exists = r.exists
	}
}
