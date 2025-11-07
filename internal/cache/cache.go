package cache

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"os/user"
	"path"
)

// Manager defines the interface for cache operations
// Following DIP: depend on abstractions, not concretions
// Following ISP: focused interface with only necessary methods
type Manager interface {
	// IsCached checks if the file is already cached and unchanged
	IsCached(filePath string, fileContent []byte) bool
	// UpdateCache updates the cache with the new file hash
	UpdateCache(filePath string, content []byte) error
}

// NoOpCacheManager is a cache manager that does nothing (Null Object Pattern)
// Following OCP: can extend behavior without modifying existing code
type NoOpCacheManager struct{}

// NewNoOpCacheManager creates a new no-op cache manager
func NewNoOpCacheManager() Manager {
	return &NoOpCacheManager{}
}

// IsCached always returns false for no-op cache
func (n *NoOpCacheManager) IsCached(filePath string, fileContent []byte) bool {
	return false
}

// UpdateCache does nothing for no-op cache
func (n *NoOpCacheManager) UpdateCache(filePath string, content []byte) error {
	return nil
}

// FileSystemCacheManager implements file-system based caching
// Following SRP: single responsibility for file-system cache operations
type FileSystemCacheManager struct {
	cacheDir string
}

// NewFileSystemCacheManager creates a new file-system cache manager
func NewFileSystemCacheManager() (Manager, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	cacheDir := path.Join(u.HomeDir, ".cache", "goimports-reviser")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, err
	}

	return &FileSystemCacheManager{
		cacheDir: cacheDir,
	}, nil
}

// IsCached checks if the file is cached and unchanged
func (f *FileSystemCacheManager) IsCached(filePath string, fileContent []byte) bool {
	cacheFile := f.getCacheFilePath(filePath)

	cachedHash, err := os.ReadFile(cacheFile)
	if err != nil {
		return false
	}

	currentHash := f.computeHash(fileContent)
	return string(cachedHash) == currentHash
}

// UpdateCache updates the cache with the new content hash
func (f *FileSystemCacheManager) UpdateCache(filePath string, content []byte) error {
	cacheFile := f.getCacheFilePath(filePath)
	fileHash := f.computeHash(content)

	// Create cache file if it doesn't exist
	if fileInfo, err := os.Stat(cacheFile); err != nil || fileInfo.IsDir() {
		if _, err := os.Create(cacheFile); err != nil {
			return err
		}
	}

	// Write the hash to cache file
	file, err := os.OpenFile(cacheFile, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := file.WriteString(fileHash); err != nil {
		return err
	}

	return nil
}

// getCacheFilePath returns the cache file path for a given file path
func (f *FileSystemCacheManager) getCacheFilePath(filePath string) string {
	hash := md5.Sum([]byte(filePath))
	return path.Join(f.cacheDir, hex.EncodeToString(hash[:]))
}

// computeHash computes the MD5 hash of content
func (f *FileSystemCacheManager) computeHash(content []byte) string {
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}
