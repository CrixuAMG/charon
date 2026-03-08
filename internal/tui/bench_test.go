package tui

import (
	"fmt"
	"testing"

	"github.com/crixuamg/charon/internal/config"
)

func largeProjectSet(n int) []config.Project {
	ps := make([]config.Project, n)
	tags := []string{"work", "hobby", "infra", "frontend", "backend"}
	for i := range ps {
		ps[i] = config.Project{
			Name:   fmt.Sprintf("project-%04d", i),
			Path:   fmt.Sprintf("/home/user/code/project-%04d", i),
			Exists: true,
			Tags:   []string{tags[i%len(tags)]},
			Pinned: i%50 == 0,
		}
	}
	return ps
}

func BenchmarkFilterProjects(b *testing.B) {
	ps := largeProjectSet(1000)
	cfg := &config.Config{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterProjects(ps, cfg, filterNone, "", false, "")
	}
}

func BenchmarkFilterProjectsWithSearch(b *testing.B) {
	ps := largeProjectSet(1000)
	cfg := &config.Config{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterProjects(ps, cfg, filterNone, "", false, "proj")
	}
}

func BenchmarkSortProjectsByName(b *testing.B) {
	ps := largeProjectSet(1000)
	cfg := &config.Config{}
	pws := filterProjects(ps, cfg, filterNone, "", false, "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Re-create the slice to avoid sorting an already-sorted slice.
		clone := make([]projectWithIndex, len(pws))
		copy(clone, pws)
		sortProjects(clone, sortByName, nil)
	}
}

func BenchmarkSortProjectsByCustom(b *testing.B) {
	ps := largeProjectSet(1000)
	cfg := &config.Config{}
	pws := filterProjects(ps, cfg, filterNone, "", false, "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clone := make([]projectWithIndex, len(pws))
		copy(clone, pws)
		sortProjects(clone, sortByCustom, nil)
	}
}
