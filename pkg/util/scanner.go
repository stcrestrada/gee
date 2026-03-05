package util

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/stcrestrada/gogo/v3"
)

// ScanResult represents a single discovered git repo.
type ScanResult struct {
	Name   string // directory name (basename)
	Path   string // absolute path to the repo root
	Remote string // origin URL if detectable, else ""
}

// ScannerConfig controls the filesystem scan.
type ScannerConfig struct {
	Root     string // starting directory (default: ~)
	MaxDepth int    // maximum recursion depth (default: 5)
}

// skipDirs are directory basenames that are never git repos and are expensive to walk.
var skipDirs = map[string]bool{
	"node_modules": true,
	".cache":       true,
	"vendor":       true,
	"Library":      true,
	".Trash":       true,
	".local":       true,
	".npm":         true,
	".cargo":       true,
	".rustup":      true,
	".pyenv":       true,
	".nvm":         true,
	"Caches":       true,
	".git":         true,
	"go":           true,
}

// scanTask is the internal payload submitted to the StreamPool.
type scanTask struct {
	dir   string
	depth int
}

// ScanForRepos walks the filesystem from cfg.Root and streams discovered repos.
// The returned channel closes when the scan completes. Cancel ctx to abort early.
func ScanForRepos(ctx context.Context, cfg ScannerConfig) <-chan ScanResult {
	if cfg.Root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		cfg.Root = home
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 5
	}

	outCh := make(chan ScanResult, 32)

	// StreamPool with bounded concurrency — we don't want to thrash the filesystem.
	pool := gogo.NewStreamPool[ScanResult](ctx, 8, gogo.WithBufferSize(64))

	// Track outstanding work. When it reaches zero, all directories have been explored.
	var outstanding atomic.Int64
	var closeOnce sync.Once

	// tryClose calls pool.Close() exactly once when all work is done.
	tryClose := func() {
		if outstanding.Load() == 0 {
			closeOnce.Do(func() {
				pool.Close()
			})
		}
	}

	// submitScan submits a directory scan task to the pool.
	submitScan := func(task scanTask) {
		outstanding.Add(1)
		err := pool.Submit(func(ctx context.Context) (ScanResult, error) {
			defer func() {
				outstanding.Add(-1)
				tryClose()
			}()
			return scanDirectory(ctx, task, pool, &outstanding, &closeOnce, cfg.MaxDepth)
		})
		if err != nil {
			// Pool already closed (e.g. context cancelled)
			outstanding.Add(-1)
		}
	}

	// Seed with the root directory.
	submitScan(scanTask{dir: cfg.Root, depth: 0})

	// Drain pool results and forward to the output channel.
	go func() {
		for result := range pool.Results() {
			if result.Error == nil && result.Result.Path != "" {
				outCh <- result.Result
			}
		}
		close(outCh)
	}()

	return outCh
}

// scanDirectory processes a single directory. If it contains .git, it's a repo.
// Otherwise, it submits child directories for scanning.
func scanDirectory(
	ctx context.Context,
	task scanTask,
	pool *gogo.StreamPool[ScanResult],
	outstanding *atomic.Int64,
	closeOnce *sync.Once,
	maxDepth int,
) (ScanResult, error) {
	select {
	case <-ctx.Done():
		return ScanResult{}, ctx.Err()
	default:
	}

	// Check if this directory is a git repo.
	gitDir := filepath.Join(task.dir, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		// Found a repo — detect remote and return. Don't recurse deeper.
		remote := detectRemote(ctx, task.dir)
		return ScanResult{
			Name:   filepath.Base(task.dir),
			Path:   task.dir,
			Remote: remote,
		}, nil
	}

	// Not a repo — scan children if we haven't hit max depth.
	if task.depth >= maxDepth {
		return ScanResult{}, nil
	}

	entries, err := os.ReadDir(task.dir)
	if err != nil {
		// Permission denied, etc. — skip silently.
		return ScanResult{}, nil
	}

	tryClose := func() {
		if outstanding.Load() == 0 {
			closeOnce.Do(func() {
				pool.Close()
			})
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()

		// Skip known junk directories.
		if skipDirs[name] {
			continue
		}

		// Skip hidden directories (except those we explicitly handle).
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Skip symlinks to avoid cycles.
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		childPath := filepath.Join(task.dir, name)

		outstanding.Add(1)
		submitErr := pool.Submit(func(ctx context.Context) (ScanResult, error) {
			defer func() {
				outstanding.Add(-1)
				tryClose()
			}()
			return scanDirectory(ctx, scanTask{dir: childPath, depth: task.depth + 1}, pool, outstanding, closeOnce, maxDepth)
		})
		if submitErr != nil {
			outstanding.Add(-1)
		}
	}

	return ScanResult{}, nil
}

// detectRemote runs `git config --get remote.origin.url` to get the remote.
func detectRemote(ctx context.Context, repoPath string) string {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
