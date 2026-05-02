package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/maltehedderich/rename-ai/internal/adapters/ui"
	"github.com/maltehedderich/rename-ai/internal/domain"
)

const maxConcurrentAIRequests = 4

type fileSystem interface {
	ReadFile(path string) ([]byte, error)
	Rename(oldPath, newPath string) error
	Exists(path string) bool
	GetMimeType(path string) (string, error)
	ListVisibleFiles(dir string) ([]string, error)
}

type nameGenerator interface {
	GenerateName(ctx context.Context, content []byte, mimeType string, currentExt string) (string, string, error)
}

type renameProgress struct {
	MimeType      string
	AnalyzingName string
}

type renameWorkResult struct {
	OriginalPath string
	OriginalName string
	ProposedName string
	Reasoning    string
	MimeType     string
	Err          error
}

type renamePlan struct {
	renames []renameExecution
	skipped []renameSkipped
}

type renameExecution struct {
	OriginalPath string
	TargetPath   string
	OriginalName string
	FinalName    string
	Reasoning    string
}

type renameSkipped struct {
	OriginalName string
	Reason       string
	IsError      bool
}

func chooseInputFiles(args []string, console *ui.ConsoleUI, fileSys fileSystem) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	files, err := fileSys.ListVisibleFiles(".")
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	return console.SelectFiles(files)
}

func runSingleRename(ctx context.Context, filePath string, dryRun bool, console *ui.ConsoleUI, fileSys fileSystem, generator nameGenerator) int {
	result := buildRenameProposal(ctx, filePath, fileSys, generator, func(progress renameProgress) {
		if progress.MimeType != "" {
			console.PrintDetectedType(progress.MimeType)
		}
		if progress.AnalyzingName != "" {
			console.PrintAnalyzing(progress.AnalyzingName)
		}
	})
	if result.Err != nil {
		console.Error(result.Err.Error())
		return 1
	}

	dir := filepath.Dir(filePath)
	finalName := domain.ResolveCollision(result.ProposedName, func(name string) bool {
		return fileSys.Exists(filepath.Join(dir, name))
	})

	console.PrintProposal(result.OriginalName, finalName, result.Reasoning)

	if dryRun {
		console.PrintDryRun()
		return 0
	}

	confirm, err := console.Confirm("Rename?")
	if err != nil {
		console.Error(fmt.Sprintf("Input error: %v", err))
		return 1
	}
	if !confirm {
		console.PrintCancelled()
		return 0
	}

	newPath := filepath.Join(dir, finalName)
	if err := fileSys.Rename(filePath, newPath); err != nil {
		console.Error(fmt.Sprintf("Rename failed: %v", err))
		return 1
	}

	console.PrintSuccess(finalName)
	return 0
}

func runBatchRename(ctx context.Context, filePaths []string, dryRun bool, console *ui.ConsoleUI, fileSys fileSystem, generator nameGenerator) int {
	console.Info(fmt.Sprintf("Analyzing %d files with Gemini...", len(filePaths)))

	results := collectRenameWork(ctx, filePaths, fileSys, generator, maxConcurrentAIRequests)
	plan := buildRenamePlan(results, fileSys)

	console.PrintBatchProposals(uiBatchProposals(plan.renames))
	console.PrintBatchSkipped(uiBatchSkipped(plan.skipped))

	if len(plan.renames) == 0 {
		console.Info("No files to rename.")
		if planHasErrors(plan) {
			return 1
		}
		return 0
	}

	if dryRun {
		console.PrintDryRun()
		if planHasErrors(plan) {
			return 1
		}
		return 0
	}

	confirm, err := console.Confirm(fmt.Sprintf("Rename %d files?", len(plan.renames)))
	if err != nil {
		console.Error(fmt.Sprintf("Input error: %v", err))
		return 1
	}
	if !confirm {
		console.PrintCancelled()
		return 0
	}

	failures := applyRenamePlan(plan, console, fileSys)
	if failures > 0 {
		return 1
	}
	if planHasErrors(plan) {
		return 1
	}
	return 0
}

func buildRenameProposal(ctx context.Context, filePath string, fileSys fileSystem, generator nameGenerator, progress func(renameProgress)) renameWorkResult {
	result := renameWorkResult{
		OriginalPath: filePath,
		OriginalName: filepath.Base(filePath),
	}

	if !fileSys.Exists(filePath) {
		result.Err = fmt.Errorf("file not found: %s", filePath)
		return result
	}

	mimeType, err := fileSys.GetMimeType(filePath)
	if err != nil {
		result.Err = fmt.Errorf("failed to detect mime type: %w", err)
		return result
	}
	result.MimeType = mimeType
	if progress != nil {
		progress(renameProgress{MimeType: mimeType})
	}

	if err := domain.IsAllowedMimeType(mimeType); err != nil {
		result.Err = fmt.Errorf("validation failed: %w", err)
		return result
	}

	if progress != nil {
		progress(renameProgress{AnalyzingName: result.OriginalName})
	}

	content, err := fileSys.ReadFile(filePath)
	if err != nil {
		result.Err = fmt.Errorf("failed to read file: %w", err)
		return result
	}

	newName, reasoning, err := generator.GenerateName(ctx, content, mimeType, filepath.Ext(filePath))
	if err != nil {
		result.Err = fmt.Errorf("AI Generation failed: %w", err)
		return result
	}

	result.ProposedName = domain.SanitizeFilename(newName)
	result.Reasoning = reasoning
	return result
}

func collectRenameWork(ctx context.Context, filePaths []string, fileSys fileSystem, generator nameGenerator, limit int) []renameWorkResult {
	if len(filePaths) == 0 {
		return nil
	}
	if limit < 1 {
		limit = 1
	}
	if limit > len(filePaths) {
		limit = len(filePaths)
	}

	type job struct {
		index int
		path  string
	}
	type indexedResult struct {
		index  int
		result renameWorkResult
	}

	jobs := make(chan job)
	results := make([]renameWorkResult, len(filePaths))
	resultCh := make(chan indexedResult, len(filePaths))

	var workers sync.WaitGroup
	for i := 0; i < limit; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range jobs {
				resultCh <- indexedResult{
					index:  job.index,
					result: buildRenameProposal(ctx, job.path, fileSys, generator, nil),
				}
			}
		}()
	}

	go func() {
		for i, path := range filePaths {
			jobs <- job{index: i, path: path}
		}
		close(jobs)
		workers.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		results[result.index] = result.result
	}

	return results
}

func buildRenamePlan(results []renameWorkResult, fileSys fileSystem) renamePlan {
	plan := renamePlan{
		renames: make([]renameExecution, 0, len(results)),
		skipped: make([]renameSkipped, 0),
	}
	successful := make([]renameWorkResult, 0, len(results))
	candidates := make([]domain.RenameCandidate, 0, len(results))

	for _, result := range results {
		if result.Err != nil {
			plan.skipped = append(plan.skipped, renameSkipped{
				OriginalName: result.OriginalName,
				Reason:       result.Err.Error(),
				IsError:      true,
			})
			continue
		}

		successful = append(successful, result)
		candidates = append(candidates, domain.RenameCandidate{
			OriginalPath: result.OriginalPath,
			ProposedName: result.ProposedName,
		})
	}

	resolved := domain.ResolveBatchCollisions(candidates, fileSys.Exists)
	for i, result := range resolved {
		work := successful[i]
		if result.NoOp {
			plan.skipped = append(plan.skipped, renameSkipped{
				OriginalName: work.OriginalName,
				Reason:       "already has the proposed name",
			})
			continue
		}

		plan.renames = append(plan.renames, renameExecution{
			OriginalPath: work.OriginalPath,
			TargetPath:   filepath.Join(filepath.Dir(work.OriginalPath), result.FinalName),
			OriginalName: work.OriginalName,
			FinalName:    result.FinalName,
			Reasoning:    work.Reasoning,
		})
	}

	return plan
}

func applyRenamePlan(plan renamePlan, console *ui.ConsoleUI, fileSys fileSystem) int {
	failures := 0
	successes := 0
	for _, rename := range plan.renames {
		if err := fileSys.Rename(rename.OriginalPath, rename.TargetPath); err != nil {
			console.Error(fmt.Sprintf("Rename failed for %s: %v", rename.OriginalName, err))
			failures++
			continue
		}
		successes++
	}

	if successes > 0 {
		console.PrintBatchSuccess(successes)
	}
	return failures
}

func uiBatchProposals(renames []renameExecution) []ui.BatchProposal {
	proposals := make([]ui.BatchProposal, 0, len(renames))
	for _, rename := range renames {
		proposals = append(proposals, ui.BatchProposal{
			OriginalName: rename.OriginalName,
			ProposedName: rename.FinalName,
			Reasoning:    rename.Reasoning,
		})
	}
	return proposals
}

func uiBatchSkipped(skipped []renameSkipped) []ui.BatchSkipped {
	items := make([]ui.BatchSkipped, 0, len(skipped))
	for _, item := range skipped {
		items = append(items, ui.BatchSkipped{
			OriginalName: item.OriginalName,
			Reason:       item.Reason,
		})
	}
	return items
}

func planHasErrors(plan renamePlan) bool {
	for _, skipped := range plan.skipped {
		if skipped.IsError {
			return true
		}
	}
	return false
}
