package tui

import (
	"sort"
	"strings"
	"time"

	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/db"
	"github.com/crixuamg/charon/internal/execution"
	"github.com/sahilm/fuzzy"
)

type projectWithIndex struct {
	project config.Project
	index   int
}

func filterProjects(projects []config.Project, cfg *config.Config, filter filterMode, activeTag string, searchQuery string) []projectWithIndex {
	var filtered []projectWithIndex

	for i, p := range projects {
		if !matchesFilter(p, cfg, filter) {
			continue
		}
		if activeTag != "" && !hasTag(p, activeTag) {
			continue
		}
		filtered = append(filtered, projectWithIndex{project: p, index: i})
	}

	if searchQuery == "" {
		return filtered
	}

	return fuzzyFilterProjects(filtered, searchQuery)
}

func matchesFilter(p config.Project, cfg *config.Config, filter filterMode) bool {
	switch filter {
	case filterDocker:
		return execution.Resolve(cfg, p).Type == "docker"
	case filterLocal:
		return execution.Resolve(cfg, p).Type == "local"
	default:
		return true
	}
}

func hasTag(p config.Project, tag string) bool {
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AllTags returns a sorted, deduplicated list of all tags across all projects.
func allTags(projects []config.Project) []string {
	seen := make(map[string]struct{})
	var tags []string
	for _, p := range projects {
		for _, t := range p.Tags {
			if _, ok := seen[t]; !ok {
				seen[t] = struct{}{}
				tags = append(tags, t)
			}
		}
	}
	return tags
}

func fuzzyFilterProjects(projects []projectWithIndex, query string) []projectWithIndex {
	// Search across name and path so users can find projects by location too.
	targets := make([]string, len(projects))
	for i, p := range projects {
		targets[i] = p.project.Name + " " + p.project.Path
	}

	matches := fuzzy.Find(query, targets)
	if len(matches) == 0 {
		return nil
	}

	result := make([]projectWithIndex, len(matches))
	for i, match := range matches {
		result[i] = projects[match.Index]
	}

	return result
}

func sortProjects(projects []projectWithIndex, mode sortMode, database *db.DB) {
	// Cache DB data upfront to avoid O(n log n) queries during sort.
	var accessTimes map[string]*time.Time
	var accessCounts map[string]int

	if database != nil {
		switch mode {
		case sortByRecent:
			accessTimes = make(map[string]*time.Time, len(projects))
			for _, p := range projects {
				t, _ := database.GetProjectAccessTime(p.project.Name)
				accessTimes[p.project.Name] = t
			}
		case sortByFrequent:
			accessCounts = make(map[string]int, len(projects))
			for _, p := range projects {
				c, _ := database.GetProjectAccessCount(p.project.Name)
				accessCounts[p.project.Name] = c
			}
		}
	}

	sort.SliceStable(projects, func(i, j int) bool {
		pi, pj := projects[i].project, projects[j].project

		// Pinned projects always sort before unpinned.
		if pi.Pinned != pj.Pinned {
			return pi.Pinned
		}

		switch mode {
		case sortByName:
			return strings.ToLower(pi.Name) < strings.ToLower(pj.Name)

		case sortByRecent:
			ti := accessTimes[pi.Name]
			tj := accessTimes[pj.Name]
			if ti == nil && tj == nil {
				return false
			}
			if ti == nil {
				return false
			}
			if tj == nil {
				return true
			}
			return ti.After(*tj)

		case sortByFrequent:
			return accessCounts[pi.Name] > accessCounts[pj.Name]

		default: // sortByCustom — preserve original config order
			return projects[i].index < projects[j].index
		}
	})
}

func getOriginalIndex(projects []projectWithIndex, cursor int) int {
	if cursor < 0 || cursor >= len(projects) {
		return 0
	}
	return projects[cursor].index
}
