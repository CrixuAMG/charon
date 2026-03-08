package tui

import (
	"testing"

	"github.com/crixuamg/charon/internal/config"
)

func projects(names ...string) []config.Project {
	ps := make([]config.Project, len(names))
	for i, n := range names {
		ps[i] = config.Project{Name: n, Path: "/tmp/" + n, Exists: true}
	}
	return ps
}

func cfg() *config.Config { return &config.Config{} }

// ── filterProjects ────────────────────────────────────────────────────────────

func TestFilterProjects_NoFilter(t *testing.T) {
	ps := projects("alpha", "beta", "gamma")
	got := filterProjects(ps, cfg(), filterNone, "", false, "")
	if len(got) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(got))
	}
}

func TestFilterProjects_HidesArchivedByDefault(t *testing.T) {
	ps := projects("alpha", "beta")
	ps[1].Archived = true

	got := filterProjects(ps, cfg(), filterNone, "", false, "")
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].project.Name != "alpha" {
		t.Errorf("expected alpha, got %s", got[0].project.Name)
	}
}

func TestFilterProjects_ShowArchived(t *testing.T) {
	ps := projects("alpha", "beta")
	ps[1].Archived = true

	got := filterProjects(ps, cfg(), filterNone, "", true, "")
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestFilterProjects_TagFilter(t *testing.T) {
	ps := projects("alpha", "beta", "gamma")
	ps[0].Tags = []string{"work"}
	ps[2].Tags = []string{"work", "infra"}

	got := filterProjects(ps, cfg(), filterNone, "work", false, "")
	if len(got) != 2 {
		t.Fatalf("expected 2 work projects, got %d", len(got))
	}
}

func TestFilterProjects_FuzzySearch(t *testing.T) {
	ps := projects("myapp", "api-service", "frontend")
	got := filterProjects(ps, cfg(), filterNone, "", false, "api")
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].project.Name != "api-service" {
		t.Errorf("expected api-service, got %s", got[0].project.Name)
	}
}

func TestFilterProjects_SearchByPath(t *testing.T) {
	ps := projects("alpha", "beta")
	ps[0].Path = "/home/user/work/alpha"
	ps[1].Path = "/home/user/hobby/beta"

	got := filterProjects(ps, cfg(), filterNone, "", false, "work")
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].project.Name != "alpha" {
		t.Errorf("expected alpha, got %s", got[0].project.Name)
	}
}

// ── sortProjects ──────────────────────────────────────────────────────────────

func withIndexes(ps []config.Project) []projectWithIndex {
	result := make([]projectWithIndex, len(ps))
	for i, p := range ps {
		result[i] = projectWithIndex{project: p, index: i}
	}
	return result
}

func TestSortByName(t *testing.T) {
	pws := withIndexes(projects("zebra", "alpha", "mango"))
	sortProjects(pws, sortByName, nil)
	want := []string{"alpha", "mango", "zebra"}
	for i, w := range want {
		if pws[i].project.Name != w {
			t.Errorf("pos %d: want %s, got %s", i, w, pws[i].project.Name)
		}
	}
}

func TestSortByCustomPreservesOrder(t *testing.T) {
	pws := withIndexes(projects("c", "a", "b"))
	sortProjects(pws, sortByCustom, nil)
	want := []string{"c", "a", "b"}
	for i, w := range want {
		if pws[i].project.Name != w {
			t.Errorf("pos %d: want %s, got %s", i, w, pws[i].project.Name)
		}
	}
}

func TestPinnedAlwaysFirst(t *testing.T) {
	pws := withIndexes(projects("regular", "pinned"))
	pws[1].project.Pinned = true

	sortProjects(pws, sortByName, nil)

	if pws[0].project.Name != "pinned" {
		t.Errorf("expected pinned to be first, got %s", pws[0].project.Name)
	}
}

func TestPinnedPreservesRelativeOrder(t *testing.T) {
	pws := withIndexes(projects("b-pinned", "a-pinned", "regular"))
	pws[0].project.Pinned = true
	pws[1].project.Pinned = true

	sortProjects(pws, sortByName, nil)

	// Both pinned first, sorted by name.
	if pws[0].project.Name != "a-pinned" {
		t.Errorf("expected a-pinned first, got %s", pws[0].project.Name)
	}
	if pws[1].project.Name != "b-pinned" {
		t.Errorf("expected b-pinned second, got %s", pws[1].project.Name)
	}
}

// ── allTags ───────────────────────────────────────────────────────────────────

func TestAllTags(t *testing.T) {
	ps := projects("a", "b", "c")
	ps[0].Tags = []string{"work", "frontend"}
	ps[1].Tags = []string{"work"}
	ps[2].Tags = []string{"infra"}

	tags := allTags(ps)
	seen := make(map[string]int)
	for _, t := range tags {
		seen[t]++
	}
	if seen["work"] != 1 {
		t.Error("expected work to appear exactly once (deduplication)")
	}
	if len(tags) != 3 {
		t.Errorf("expected 3 unique tags, got %d", len(tags))
	}
}

// ── getOriginalIndex ──────────────────────────────────────────────────────────

func TestGetOriginalIndex(t *testing.T) {
	pws := []projectWithIndex{
		{project: config.Project{Name: "a"}, index: 5},
		{project: config.Project{Name: "b"}, index: 2},
	}
	if got := getOriginalIndex(pws, 0); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
	if got := getOriginalIndex(pws, 1); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
	// Out of bounds should return 0.
	if got := getOriginalIndex(pws, 99); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}
