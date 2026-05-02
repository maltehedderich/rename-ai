package fs_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/maltehedderich/rename-ai/internal/adapters/fs"
)

func TestOsFileSystem_ListVisibleFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "b.txt"))
	mustWriteFile(t, filepath.Join(dir, "a.txt"))
	mustWriteFile(t, filepath.Join(dir, ".hidden.txt"))
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	fileSys := fs.NewOsFileSystem()
	got, err := fileSys.ListVisibleFiles(dir)
	if err != nil {
		t.Fatalf("ListVisibleFiles() error = %v", err)
	}

	want := []string{
		filepath.Join(dir, "a.txt"),
		filepath.Join(dir, "b.txt"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListVisibleFiles() = %#v; want %#v", got, want)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
