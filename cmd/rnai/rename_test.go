package main

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCollectRenameWorkAndBuildRenamePlan(t *testing.T) {
	t.Parallel()

	fileSys := newFakeFileSystem()
	for _, path := range []string{"third.txt", "first.txt", "bad.txt", "noop.txt"} {
		fileSys.addFile(path, "text/plain", []byte(path))
	}
	generator := &fakeNameGenerator{
		names: map[string]fakeGeneratedName{
			"third.txt": {filename: "same.txt", reasoning: "third reason"},
			"first.txt": {filename: "same.txt", reasoning: "first reason"},
			"noop.txt":  {filename: "noop.txt", reasoning: "noop reason"},
		},
		errors: map[string]error{
			"bad.txt": errors.New("boom"),
		},
		delay: 5 * time.Millisecond,
	}

	results := collectRenameWork(
		context.Background(),
		[]string{"third.txt", "first.txt", "bad.txt", "noop.txt"},
		fileSys,
		generator,
		2,
	)
	plan := buildRenamePlan(results, fileSys)

	gotRenames := make([]string, 0, len(plan.renames))
	for _, rename := range plan.renames {
		gotRenames = append(gotRenames, rename.OriginalName+"->"+rename.FinalName)
	}
	wantRenames := []string{
		"third.txt->same.txt",
		"first.txt->same-1.txt",
	}
	if !reflect.DeepEqual(gotRenames, wantRenames) {
		t.Fatalf("renames = %#v; want %#v", gotRenames, wantRenames)
	}

	gotSkipped := make([]string, 0, len(plan.skipped))
	for _, skipped := range plan.skipped {
		gotSkipped = append(gotSkipped, skipped.OriginalName+":"+skipped.Reason)
	}
	if len(gotSkipped) != 2 {
		t.Fatalf("skipped = %#v; want 2 items", gotSkipped)
	}
	if !strings.Contains(gotSkipped[0], "bad.txt:AI Generation failed: boom") {
		t.Fatalf("first skipped item = %q; want AI error for bad.txt", gotSkipped[0])
	}
	if gotSkipped[1] != "noop.txt:already has the proposed name" {
		t.Fatalf("second skipped item = %q; want noop skip", gotSkipped[1])
	}
	if generator.maxInFlight > 2 {
		t.Fatalf("max concurrent GenerateName calls = %d; want <= 2", generator.maxInFlight)
	}
}

type fakeFileSystem struct {
	files      map[string]fakeFile
	existing   map[string]bool
	renamed    map[string]string
	renameErrs map[string]error
}

type fakeFile struct {
	mimeType string
	content  []byte
}

func newFakeFileSystem() *fakeFileSystem {
	return &fakeFileSystem{
		files:      make(map[string]fakeFile),
		existing:   make(map[string]bool),
		renamed:    make(map[string]string),
		renameErrs: make(map[string]error),
	}
}

func (fs *fakeFileSystem) addFile(path string, mimeType string, content []byte) {
	fs.files[path] = fakeFile{mimeType: mimeType, content: content}
	fs.existing[path] = true
}

func (fs *fakeFileSystem) ReadFile(path string) ([]byte, error) {
	file, ok := fs.files[path]
	if !ok {
		return nil, errors.New("read failed")
	}
	return file.content, nil
}

func (fs *fakeFileSystem) Rename(oldPath, newPath string) error {
	if err := fs.renameErrs[oldPath]; err != nil {
		return err
	}
	fs.renamed[oldPath] = newPath
	fs.existing[oldPath] = false
	fs.existing[newPath] = true
	return nil
}

func (fs *fakeFileSystem) Exists(path string) bool {
	return fs.existing[path]
}

func (fs *fakeFileSystem) GetMimeType(path string) (string, error) {
	file, ok := fs.files[path]
	if !ok {
		return "", errors.New("mime failed")
	}
	return file.mimeType, nil
}

func (fs *fakeFileSystem) ListVisibleFiles(dir string) ([]string, error) {
	return nil, nil
}

type fakeNameGenerator struct {
	names       map[string]fakeGeneratedName
	errors      map[string]error
	delay       time.Duration
	mu          sync.Mutex
	inFlight    int
	maxInFlight int
}

type fakeGeneratedName struct {
	filename  string
	reasoning string
}

func (g *fakeNameGenerator) GenerateName(ctx context.Context, content []byte, mimeType string, currentExt string) (string, string, error) {
	g.mu.Lock()
	g.inFlight++
	if g.inFlight > g.maxInFlight {
		g.maxInFlight = g.inFlight
	}
	g.mu.Unlock()

	if g.delay > 0 {
		time.Sleep(g.delay)
	}

	g.mu.Lock()
	g.inFlight--
	g.mu.Unlock()

	if err := g.errors[string(content)]; err != nil {
		return "", "", err
	}
	name := g.names[string(content)]
	return name.filename, name.reasoning, nil
}
