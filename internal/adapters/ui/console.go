package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
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

type ConsoleUI struct {
	reader io.Reader
	writer io.Writer
}

func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

// NewConsoleUIWithStreams allows creating a ConsoleUI with custom streams (for testing)
func NewConsoleUIWithStreams(r io.Reader, w io.Writer) *ConsoleUI {
	return &ConsoleUI{
		reader: r,
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

func (ui *ConsoleUI) PrintSuccess(newName string) {
	_, _ = fmt.Fprintf(ui.writer, "%s> Success! Renamed to %s%s%s\n", Green, Bold, newName, Reset)
}

func (ui *ConsoleUI) PrintDryRun() {
	_, _ = fmt.Fprintf(ui.writer, "%s> Dry-run enabled. Skipping rename.%s\n", Yellow, Reset)
}

func (ui *ConsoleUI) PrintCancelled() {
	_, _ = fmt.Fprintf(ui.writer, "%s> Cancelled.%s\n", Red, Reset)
}

func (ui *ConsoleUI) Confirm(question string) (bool, error) {
	_, _ = fmt.Fprintf(ui.writer, "%s%s [y/N]: %s", Bold, question, Reset)
	reader := bufio.NewReader(ui.reader)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	response = strings.TrimSpace(response)
	return strings.ToLower(response) == "y", nil
}

func (ui *ConsoleUI) Error(msg string) {
	_, _ = fmt.Fprintf(ui.writer, "%sError: %s%s\n", Red, msg, Reset)
}
