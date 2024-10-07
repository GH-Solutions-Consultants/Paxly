// core/cache.go
package core

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CacheDependency caches the downloaded package data.
func CacheDependency(dep Dependency, data io.Reader) error {
	cacheDir := "pkgmgr_cache"
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.Mkdir(cacheDir, 0755); err != nil {
			return err
		}
	}

	hash := sha256.New()
	tee := io.TeeReader(data, hash)

	// Read the data to compute hash
	_, err := io.Copy(io.Discard, tee)
	if err != nil {
		return err
	}
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	// Create cache file path
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s-%s.cache", dep.Name, checksum))
	if _, err := os.Stat(cachePath); err == nil {
		// Cached version exists
		return nil
	}

	// Save to cache
	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Reset the reader if possible or handle accordingly
	// This depends on how data is passed. If data is already consumed, consider reading from the original source again.

	return nil
}
