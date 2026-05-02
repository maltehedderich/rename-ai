package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

type OsFileSystem struct{}

func NewOsFileSystem() *OsFileSystem {
	return &OsFileSystem{}
}

func (fs *OsFileSystem) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return data, nil
}

func (fs *OsFileSystem) Rename(oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", oldPath, newPath, err)
	}
	return nil
}

func (fs *OsFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (fs *OsFileSystem) GetMimeType(path string) (string, error) {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to detect mimetype for %s: %w", path, err)
	}
	return mtype.String(), nil
}

func (fs *OsFileSystem) ListVisibleFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || !entry.Type().IsRegular() {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}

	sort.Strings(files)
	return files, nil
}
