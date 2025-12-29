// Package cache provides TTL-based caching for usage data.
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	apierrors "github.com/benjaminabbitt/claude-limits/internal/errors"
	"github.com/benjaminabbitt/claude-limits/internal/models"
)

// File permission constants
const (
	DirMode  = 0700 // rwx------ for cache directory (private)
	FileMode = 0600 // rw------- for cache file (contains API data)
)

// Data represents cached usage data with a timestamp
type Data struct {
	Timestamp time.Time       `json:"timestamp"`
	Usage     json.RawMessage `json:"usage"`
}

// Cache manages the usage cache
type Cache struct {
	dir     string
	file    string
	verbose bool
}

// New creates a new Cache instance
func New(verbose bool) *Cache {
	dir := getCacheDir()
	return &Cache{
		dir:     dir,
		file:    filepath.Join(dir, "usage.json"),
		verbose: verbose,
	}
}

// getCacheDir returns the platform-appropriate cache directory
func getCacheDir() string {
	// Use os.UserCacheDir for cross-platform cache location:
	// - Linux: $XDG_CACHE_HOME or ~/.cache
	// - macOS: ~/Library/Caches
	// - Windows: %LocalAppData%
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(cacheDir, "claudelimits")
}

// Read attempts to read cached data if it's still valid
func (c *Cache) Read(ttlSeconds int) (*models.Usage, error) {
	data, err := os.ReadFile(c.file)
	if err != nil {
		return nil, apierrors.NewCacheError("read", c.file, err)
	}

	var cache Data
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, apierrors.NewCacheError("parse", c.file, err)
	}

	// Check if cache is still valid
	if time.Since(cache.Timestamp) > time.Duration(ttlSeconds)*time.Second {
		return nil, apierrors.ErrCacheExpired
	}

	var usage models.Usage
	if err := json.Unmarshal(cache.Usage, &usage); err != nil {
		return nil, apierrors.NewCacheError("parse", c.file, err)
	}

	return &usage, nil
}

// Write saves usage data to the cache
func (c *Cache) Write(usage *models.Usage) error {
	cache := Data{
		Timestamp: time.Now(),
		Usage:     usage.Raw,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return apierrors.NewCacheError("marshal", c.file, err)
	}

	// Create cache directory if needed
	if err := os.MkdirAll(c.dir, DirMode); err != nil {
		return apierrors.NewCacheError("mkdir", c.dir, err)
	}

	if err := os.WriteFile(c.file, data, FileMode); err != nil {
		return apierrors.NewCacheError("write", c.file, err)
	}

	return nil
}

// Dir returns the cache directory path
func (c *Cache) Dir() string {
	return c.dir
}

// File returns the cache file path
func (c *Cache) File() string {
	return c.file
}
