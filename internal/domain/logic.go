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

type RenameCandidate struct {
	OriginalPath string
	ProposedName string
}

type ResolvedRename struct {
	OriginalPath string
	FinalName    string
	NoOp         bool
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

func ResolveBatchCollisions(candidates []RenameCandidate, checkExists func(string) bool) []ResolvedRename {
	results := make([]ResolvedRename, 0, len(candidates))
	reservedTargets := make(map[string]map[string]struct{})
	selectedSources := make(map[string]map[string]struct{})

	for _, candidate := range candidates {
		dir := filepath.Dir(candidate.OriginalPath)
		name := filepath.Base(candidate.OriginalPath)
		if selectedSources[dir] == nil {
			selectedSources[dir] = make(map[string]struct{})
		}
		selectedSources[dir][name] = struct{}{}
	}

	for _, candidate := range candidates {
		dir := filepath.Dir(candidate.OriginalPath)
		originalName := filepath.Base(candidate.OriginalPath)
		proposedName := filepath.Base(candidate.ProposedName)

		result := ResolvedRename{
			OriginalPath: candidate.OriginalPath,
			FinalName:    proposedName,
		}

		if proposedName == originalName {
			result.NoOp = true
			results = append(results, result)
			continue
		}

		finalName := ResolveCollision(proposedName, func(name string) bool {
			if reservedTargets[dir] != nil {
				if _, ok := reservedTargets[dir][name]; ok {
					return true
				}
			}
			if _, ok := selectedSources[dir][name]; ok {
				return true
			}
			return checkExists(filepath.Join(dir, name))
		})

		if reservedTargets[dir] == nil {
			reservedTargets[dir] = make(map[string]struct{})
		}
		reservedTargets[dir][finalName] = struct{}{}
		result.FinalName = finalName
		results = append(results, result)
	}

	return results
}
