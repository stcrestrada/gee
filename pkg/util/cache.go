package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gee/pkg/types"
)

// CachedRepo is one entry in the JSON cache file.
type CachedRepo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`          // Full absolute path to the repo root (contains .git)
	Remote       string    `json:"remote"`         // origin URL, may be ""
	Pinned       bool      `json:"pinned"`         // true = user-curated, false = auto-discovered
	DiscoveredAt time.Time `json:"discovered_at"`
}

// RepoCache manages reading/writing ~/.config/gee/cache.json.
// Safe for concurrent use.
type RepoCache struct {
	path  string
	mu    sync.Mutex
	repos map[string]CachedRepo // keyed by absolute path for dedup
}

// NewRepoCache creates a RepoCache using the default cache path.
func NewRepoCache() *RepoCache {
	return &RepoCache{
		path:  DefaultCachePath(),
		repos: make(map[string]CachedRepo),
	}
}

// DefaultCachePath returns ~/.config/gee/cache.json.
func DefaultCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "gee", "cache.json")
}

// Load reads the cache from disk. Creates the directory and file if missing.
// Returns the repos that were loaded.
func (c *RepoCache) Load() ([]CachedRepo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var repos []CachedRepo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}

	c.repos = make(map[string]CachedRepo, len(repos))
	for _, r := range repos {
		c.repos[r.Path] = r
	}

	return repos, nil
}

// Add inserts or updates a repo in memory. Returns true if it was genuinely new.
// If the repo already exists by path, it updates Remote if previously empty
// but does NOT overwrite Pinned (pinned is sticky).
func (c *RepoCache) Add(repo CachedRepo) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	existing, exists := c.repos[repo.Path]
	if exists {
		// Update remote if it was empty and we now have one
		if existing.Remote == "" && repo.Remote != "" {
			existing.Remote = repo.Remote
			c.repos[repo.Path] = existing
		}
		return false
	}

	c.repos[repo.Path] = repo
	return true
}

// Pin sets pinned=true for the repo at the given path. Returns false if not found.
func (c *RepoCache) Pin(path string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	r, ok := c.repos[path]
	if !ok {
		return false
	}
	r.Pinned = true
	c.repos[path] = r
	return true
}

// Unpin sets pinned=false for the repo at the given path. Returns false if not found.
func (c *RepoCache) Unpin(path string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	r, ok := c.repos[path]
	if !ok {
		return false
	}
	r.Pinned = false
	c.repos[path] = r
	return true
}

// Remove deletes a repo from the cache entirely.
func (c *RepoCache) Remove(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.repos, path)
}

// Save writes the full cache to disk atomically (write .tmp, then os.Rename).
func (c *RepoCache) Save() error {
	c.mu.Lock()
	repos := c.allLocked()
	c.mu.Unlock()

	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpPath := c.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, c.path)
}

// All returns a snapshot sorted: pinned first (alphabetical), then discovered (alphabetical).
func (c *RepoCache) All() []CachedRepo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.allLocked()
}

func (c *RepoCache) allLocked() []CachedRepo {
	repos := make([]CachedRepo, 0, len(c.repos))
	for _, r := range c.repos {
		repos = append(repos, r)
	}
	sort.Slice(repos, func(i, j int) bool {
		if repos[i].Pinned != repos[j].Pinned {
			return repos[i].Pinned // pinned first
		}
		return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
	})
	return repos
}

// Pinned returns only pinned repos, sorted alphabetically.
func (c *RepoCache) Pinned() []CachedRepo {
	c.mu.Lock()
	defer c.mu.Unlock()

	var repos []CachedRepo
	for _, r := range c.repos {
		if r.Pinned {
			repos = append(repos, r)
		}
	}
	sort.Slice(repos, func(i, j int) bool {
		return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
	})
	return repos
}

// FindByPath returns the cached repo whose Path matches or is a parent of dir.
// Used by CLI commands to detect "am I inside a known repo?"
func (c *RepoCache) FindByPath(dir string) (CachedRepo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Walk up from dir checking each level against the cache.
	current := dir
	for {
		if r, ok := c.repos[current]; ok {
			return r, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return CachedRepo{}, false
}

// LoadReposForCLI returns the repo list that a CLI command should operate on.
// If cwd is inside a known repo, returns just that repo.
// If all is true, returns everything. Otherwise returns only pinned.
func (c *RepoCache) LoadReposForCLI(cwd string, all bool) []CachedRepo {
	if !all {
		if repo, found := c.FindByPath(cwd); found {
			return []CachedRepo{repo}
		}
	}
	if all {
		return c.All()
	}
	return c.Pinned()
}

// ToRepoSlice converts []CachedRepo to []types.Repo for compatibility
// with existing command internals that use types.Repo.
// types.Repo.Path is the parent directory; types.Repo.Name is the basename.
func ToRepoSlice(cached []CachedRepo) []types.Repo {
	repos := make([]types.Repo, len(cached))
	for i, c := range cached {
		repos[i] = types.Repo{
			Name:   c.Name,
			Path:   filepath.Dir(c.Path),
			Remote: c.Remote,
		}
	}
	return repos
}
