package domain

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type RenameRequest struct {
	OriginalPath string
	Content      []byte
	MimeType     string
	Extension    string
}

type RenameResult struct {
	OriginalName string
	ProposedName string
	Reasoning    string
}

// SanitizeFilename removes invalid characters and enforces kebab-case
func SanitizeFilename(name string) string {
	// Remove invalid characters for filenames (this is a simplified regex)
	reg, _ := regexp.Compile("[^a-zA-Z0-9._-]+")
	safe := reg.ReplaceAllString(name, "-")

	// Enforce lower case for kebab-case
	return strings.ToLower(safe)
}

// ResolveCollision checks if a file exists and appends a counter if it does.
// This function assumes the FileSystem interface is available to check existence,
// or it takes a 'checkExists' function.
func ResolveCollision(baseName string, checkExists func(string) bool) string {
	if !checkExists(baseName) {
		return baseName
	}

	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	counter := 1
	for {
		newName := fmt.Sprintf("%s-%d%s", nameWithoutExt, counter, ext)
		if !checkExists(newName) {
			return newName
		}
		counter++
	}
}
