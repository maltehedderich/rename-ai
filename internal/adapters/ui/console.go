package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ConsoleUI struct{}

func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{}
}

func (ui *ConsoleUI) PrintProposal(oldName, newName, reasoning string) {
	fmt.Printf("\nProposal:\n")
	fmt.Printf("  Reasoning: %s\n", reasoning)
	fmt.Printf("  Rename: %s -> %s\n", oldName, newName)
}

func (ui *ConsoleUI) Confirm(question string) (bool, error) {
	fmt.Printf("%s [y/N]: ", question)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	response = strings.TrimSpace(response)
	return strings.ToLower(response) == "y", nil
}

func (ui *ConsoleUI) Error(msg string) {
	fmt.Printf("Error: %s\n", msg)
}
