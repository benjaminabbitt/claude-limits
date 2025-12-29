package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/benjaminabbitt/claude-limits/internal/models"
)

func TestCacheReadWrite(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a cache with custom directory
	c := &Cache{
		dir:     tmpDir,
		file:    filepath.Join(tmpDir, "test_usage.json"),
		verbose: false,
	}

	// Create test usage data
	rawJSON := json.RawMessage(`{"five_hour_utilization": 75.5}`)
	usage := &models.Usage{}
	_ = json.Unmarshal(rawJSON, usage)

	// Write to cache
	err := c.Write(usage)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists with correct permissions
	info, err := os.Stat(c.file)
	if err != nil {
		t.Fatalf("Cache file not created: %v", err)
	}
	if info.Mode().Perm() != FileMode {
		t.Errorf("File permissions = %o, want %o", info.Mode().Perm(), FileMode)
	}

	// Read from cache with valid TTL
	cached, err := c.Read(60)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if cached == nil {
		t.Fatal("Read returned nil usage")
	}
}

func TestCacheExpiry(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Cache{
		dir:     tmpDir,
		file:    filepath.Join(tmpDir, "test_usage.json"),
		verbose: false,
	}

	// Create test usage data
	rawJSON := json.RawMessage(`{"test": "data"}`)
	usage := &models.Usage{}
	_ = json.Unmarshal(rawJSON, usage)

	// Write to cache
	_ = c.Write(usage)

	// Read with 0 TTL should fail (expired immediately)
	_, err := c.Read(0)
	if err == nil {
		t.Error("Read with 0 TTL should return error")
	}
}

func TestCacheReadNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Cache{
		dir:     tmpDir,
		file:    filepath.Join(tmpDir, "nonexistent.json"),
		verbose: false,
	}

	_, err := c.Read(60)
	if err == nil {
		t.Error("Read of nonexistent file should return error")
	}
}

func TestCacheReadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	err := os.WriteFile(cacheFile, []byte("not valid json"), FileMode)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	c := &Cache{
		dir:     tmpDir,
		file:    cacheFile,
		verbose: false,
	}

	_, err = c.Read(60)
	if err == nil {
		t.Error("Read of invalid JSON should return error")
	}
}

func TestCacheDirectoryPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "newcachedir")

	c := &Cache{
		dir:     cacheDir,
		file:    filepath.Join(cacheDir, "usage.json"),
		verbose: false,
	}

	rawJSON := json.RawMessage(`{"test": "data"}`)
	usage := &models.Usage{}
	_ = json.Unmarshal(rawJSON, usage)

	// Write should create directory
	err := c.Write(usage)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify directory permissions
	info, err := os.Stat(cacheDir)
	if err != nil {
		t.Fatalf("Cache directory not created: %v", err)
	}
	if info.Mode().Perm() != DirMode {
		t.Errorf("Directory permissions = %o, want %o", info.Mode().Perm(), DirMode)
	}
}

func TestCacheDataIntegrity(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Cache{
		dir:     tmpDir,
		file:    filepath.Join(tmpDir, "usage.json"),
		verbose: false,
	}

	// Create usage with known data
	testData := `{"five_hour_utilization": 75.5, "weekly_limit": 100}`
	rawJSON := json.RawMessage(testData)
	usage := &models.Usage{}
	_ = json.Unmarshal(rawJSON, usage)

	// Write and read back
	_ = c.Write(usage)

	// Small sleep to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	cached, err := c.Read(60)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Compare by unmarshaling to maps (handles whitespace differences)
	var original, retrieved map[string]interface{}
	_ = json.Unmarshal(usage.Raw, &original)
	_ = json.Unmarshal(cached.Raw, &retrieved)

	if original["five_hour_utilization"] != retrieved["five_hour_utilization"] {
		t.Errorf("five_hour_utilization mismatch: got %v, want %v",
			retrieved["five_hour_utilization"], original["five_hour_utilization"])
	}
	if original["weekly_limit"] != retrieved["weekly_limit"] {
		t.Errorf("weekly_limit mismatch: got %v, want %v",
			retrieved["weekly_limit"], original["weekly_limit"])
	}
}

func TestNew(t *testing.T) {
	c := New(false)
	if c == nil {
		t.Fatal("New returned nil")
	}
	if c.dir == "" {
		t.Error("Cache directory is empty")
	}
	if c.file == "" {
		t.Error("Cache file is empty")
	}
}
