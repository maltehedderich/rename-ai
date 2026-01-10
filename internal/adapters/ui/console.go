package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
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

func (ui *ConsoleUI) PrintProposal(oldName, newName, reasoning string) {
	fmt.Fprintf(ui.writer, "\nProposal:\n")
	fmt.Fprintf(ui.writer, "  Reasoning: %s\n", reasoning)
	fmt.Fprintf(ui.writer, "  Rename: %s -> %s\n", oldName, newName)
}

func (ui *ConsoleUI) Confirm(question string) (bool, error) {
	fmt.Fprintf(ui.writer, "%s [y/N]: ", question)
	reader := bufio.NewReader(ui.reader)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	response = strings.TrimSpace(response)
	return strings.ToLower(response) == "y", nil
}

func (ui *ConsoleUI) Error(msg string) {
	fmt.Fprintf(ui.writer, "Error: %s\n", msg)
}
