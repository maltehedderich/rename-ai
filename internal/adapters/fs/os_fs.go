package fs

import (
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/maltehedderich/rename-ai/internal/ports"
)

type OsFileSystem struct{}

func NewOsFileSystem() ports.FileSystem {
	return &OsFileSystem{}
}

func (fs *OsFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (fs *OsFileSystem) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (fs *OsFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (fs *OsFileSystem) GetMimeType(path string) (string, error) {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return "", err
	}
	return mtype.String(), nil
}
