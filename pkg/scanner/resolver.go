package scanner

import (
	"os"
	"path/filepath"
	"sync"
)

// ProjectResolver finds project roots by walking directory tree
type ProjectResolver struct {
	cache map[string]string
	mu    sync.RWMutex
}

// NewProjectResolver creates a new resolver instance
func NewProjectResolver() *ProjectResolver {
	return &ProjectResolver{
		cache: make(map[string]string),
	}
}

// ProjectMarkers are files/dirs that indicate a project root
var ProjectMarkers = []string{
	".git",
	"package.json",
	"composer.json",
	"wp-config.php",
	"Gemfile",
	"go.mod",
	"pyproject.toml",
	"Makefile",
	"Cargo.toml",
}

// FindProjectRoot searches up the directory tree for a project root
func (pr *ProjectResolver) FindProjectRoot(startPath string) string {
	if startPath == "" {
		return ""
	}

	// Check cache
	pr.mu.RLock()
	if cached, ok := pr.cache[startPath]; ok {
		pr.mu.RUnlock()
		return cached
	}
	pr.mu.RUnlock()

	// Walk upward
	current := startPath
	for {
		// Check for project markers
		for _, marker := range ProjectMarkers {
			markerPath := filepath.Join(current, marker)
			if _, err := os.Stat(markerPath); err == nil {
				pr.mu.Lock()
				pr.cache[startPath] = current
				pr.mu.Unlock()
				return current
			}
		}

		parent := filepath.Dir(current)
		if parent == current || parent == "/" {
			// Reached root without finding markers
			pr.mu.Lock()
			pr.cache[startPath] = ""
			pr.mu.Unlock()
			return ""
		}
		current = parent
	}
}

// ClearCache clears all cached mappings
func (pr *ProjectResolver) ClearCache() {
	pr.mu.Lock()
	pr.cache = make(map[string]string)
	pr.mu.Unlock()
}

// ClearCacheForPath clears cache entry for a specific path
func (pr *ProjectResolver) ClearCacheForPath(path string) {
	pr.mu.Lock()
	delete(pr.cache, path)
	pr.mu.Unlock()
}
