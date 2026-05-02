package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ANSI Color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	Bold   = "\033[1m"
)

var ErrSelectionCancelled = errors.New("selection cancelled")

type ConsoleUI struct {
	reader *bufio.Reader
	writer io.Writer
}

type BatchProposal struct {
	OriginalName string
	ProposedName string
	Reasoning    string
}

type BatchSkipped struct {
	OriginalName string
	Reason       string
}

func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// NewConsoleUIWithStreams allows creating a ConsoleUI with custom streams (for testing)
func NewConsoleUIWithStreams(r io.Reader, w io.Writer) *ConsoleUI {
	return &ConsoleUI{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

func (ui *ConsoleUI) Info(msg string) {
	_, _ = fmt.Fprintf(ui.writer, "%s> %s%s\n", Blue, msg, Reset)
}

func (ui *ConsoleUI) PrintModelInfo(model string) {
	_, _ = fmt.Fprintf(ui.writer, "%s> Using model: %s%s%s\n", Gray, Cyan, model, Reset)
}

func (ui *ConsoleUI) PrintDetectedType(mimeType string) {
	_, _ = fmt.Fprintf(ui.writer, "%s> Detected type: %s%s%s\n", Gray, Cyan, mimeType, Reset)
}

func (ui *ConsoleUI) PrintAnalyzing(filename string) {
	_, _ = fmt.Fprintf(ui.writer, "%s> Analyzing '%s%s%s' with Gemini...%s\n", Gray, Bold, filename, Gray, Reset)
}

func (ui *ConsoleUI) PrintProposal(oldName, newName, reasoning string) {
	_, _ = fmt.Fprintf(ui.writer, "\n%s%sProposal %s\n", Purple, Bold, Reset)
	_, _ = fmt.Fprintf(ui.writer, "%sReasoning:%s\n  %s\n\n", Bold, Reset, reasoning)
	_, _ = fmt.Fprintf(ui.writer, "%sRename:%s\n  %s%s%s -> %s%s%s\n\n", Bold, Reset, Red, oldName, Reset, Green, newName, Reset)
}

func (ui *ConsoleUI) PrintBatchProposals(proposals []BatchProposal) {
	if len(proposals) == 0 {
		return
	}

	_, _ = fmt.Fprintf(ui.writer, "\n%s%sProposals %s\n", Purple, Bold, Reset)
	for i, proposal := range proposals {
		_, _ = fmt.Fprintf(ui.writer, "%s%d.%s %s%s%s -> %s%s%s\n", Bold, i+1, Reset, Red, proposal.OriginalName, Reset, Green, proposal.ProposedName, Reset)
		if proposal.Reasoning != "" {
			_, _ = fmt.Fprintf(ui.writer, "   %sReasoning:%s %s\n", Bold, Reset, proposal.Reasoning)
		}
	}
	_, _ = fmt.Fprintln(ui.writer)
}

func (ui *ConsoleUI) PrintBatchSkipped(skipped []BatchSkipped) {
	if len(skipped) == 0 {
		return
	}

	_, _ = fmt.Fprintf(ui.writer, "\n%s%sSkipped / Errors %s\n", Yellow, Bold, Reset)
	for i, item := range skipped {
		_, _ = fmt.Fprintf(ui.writer, "%s%d.%s %s%s%s: %s\n", Bold, i+1, Reset, Yellow, item.OriginalName, Reset, item.Reason)
	}
	_, _ = fmt.Fprintln(ui.writer)
}

func (ui *ConsoleUI) PrintSuccess(newName string) {
	_, _ = fmt.Fprintf(ui.writer, "%s> Success! Renamed to %s%s%s\n", Green, Bold, newName, Reset)
}

func (ui *ConsoleUI) PrintBatchSuccess(count int) {
	_, _ = fmt.Fprintf(ui.writer, "%s> Success! Renamed %s%d%s files.%s\n", Green, Bold, count, Green, Reset)
}

func (ui *ConsoleUI) PrintDryRun() {
	_, _ = fmt.Fprintf(ui.writer, "%s> Dry-run enabled. Skipping rename.%s\n", Yellow, Reset)
}

func (ui *ConsoleUI) PrintCancelled() {
	_, _ = fmt.Fprintf(ui.writer, "%s> Cancelled.%s\n", Red, Reset)
}

func (ui *ConsoleUI) SelectFiles(files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}

	selected := make(map[int]bool)
	for {
		ui.printFileSelection(files, selected)
		_, _ = fmt.Fprintf(ui.writer, "%sSelection [1,3  1-4  a=all  n=none  Enter=done  q=quit]: %s", Bold, Reset)

		response, err := ui.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read user input: %w", err)
		}

		command := strings.ToLower(strings.TrimSpace(response))
		switch command {
		case "":
			if len(selected) == 0 {
				_, _ = fmt.Fprintf(ui.writer, "%sSelect at least one file.%s\n", Yellow, Reset)
				continue
			}
			return selectedFiles(files, selected), nil
		case "q":
			return nil, ErrSelectionCancelled
		case "a":
			for i := range files {
				selected[i] = true
			}
		case "n":
			selected = make(map[int]bool)
		default:
			if err := toggleSelectionCommand(command, len(files), selected); err != nil {
				_, _ = fmt.Fprintf(ui.writer, "%sInvalid selection: %v%s\n", Yellow, err, Reset)
			}
		}
	}
}

func (ui *ConsoleUI) Confirm(question string) (bool, error) {
	_, _ = fmt.Fprintf(ui.writer, "%s%s [y/N]: %s", Bold, question, Reset)
	response, err := ui.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

func (ui *ConsoleUI) Error(msg string) {
	_, _ = fmt.Fprintf(ui.writer, "%sError: %s%s\n", Red, msg, Reset)
}

func (ui *ConsoleUI) printFileSelection(files []string, selected map[int]bool) {
	_, _ = fmt.Fprintf(ui.writer, "\n%s%sSelect files to rename%s\n", Purple, Bold, Reset)
	for i, file := range files {
		mark := " "
		if selected[i] {
			mark = "x"
		}
		_, _ = fmt.Fprintf(ui.writer, "  %2d. [%s] %s\n", i+1, mark, file)
	}
	_, _ = fmt.Fprintln(ui.writer)
}

func selectedFiles(files []string, selected map[int]bool) []string {
	result := make([]string, 0, len(selected))
	for i, file := range files {
		if selected[i] {
			result = append(result, file)
		}
	}
	return result
}

func toggleSelectionCommand(command string, total int, selected map[int]bool) error {
	parts := strings.Split(command, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return fmt.Errorf("empty item")
		}

		if strings.Contains(part, "-") {
			if err := toggleRange(part, total, selected); err != nil {
				return err
			}
			continue
		}

		index, err := parseSelectionIndex(part, total)
		if err != nil {
			return err
		}
		selected[index] = !selected[index]
	}

	return nil
}

func toggleRange(part string, total int, selected map[int]bool) error {
	bounds := strings.Split(part, "-")
	if len(bounds) != 2 {
		return fmt.Errorf("invalid range %q", part)
	}

	start, err := parseSelectionIndex(strings.TrimSpace(bounds[0]), total)
	if err != nil {
		return err
	}
	end, err := parseSelectionIndex(strings.TrimSpace(bounds[1]), total)
	if err != nil {
		return err
	}
	if start > end {
		return fmt.Errorf("range start must be before range end")
	}

	for i := start; i <= end; i++ {
		selected[i] = !selected[i]
	}
	return nil
}

func parseSelectionIndex(value string, total int) (int, error) {
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%q is not a number", value)
	}
	if number < 1 || number > total {
		return 0, fmt.Errorf("%d is outside 1-%d", number, total)
	}
	return number - 1, nil
}
