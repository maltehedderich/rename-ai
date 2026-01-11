package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/maltehedderich/rename-ai/internal/adapters/ai"
	"github.com/maltehedderich/rename-ai/internal/adapters/fs"
	"github.com/maltehedderich/rename-ai/internal/adapters/ui"
	"github.com/maltehedderich/rename-ai/internal/domain"
)

var (
	dryRun bool
	style  string
	model  string
)

var rootCmd = &cobra.Command{
	Use:   "rnai [file]",
	Short: "Rename files using GenAI",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		ctx := context.Background()

		// 1. Initialize Adapters
		console := ui.NewConsoleUI()
		fileSys := fs.NewOsFileSystem()

		// Validation
		if !fileSys.Exists(filePath) {
			console.Error(fmt.Sprintf("File not found: %s", filePath))
			os.Exit(1)
		}

		// Config / Auth
		key := viper.GetString("GEMINI_API_KEY")
		if key == "" {
			console.Error("GEMINI_API_KEY invalid. Please set GEMINI_API_KEY environment variable.")
			os.Exit(1)
		}

		// Get model from viper (flag or env)
		modelName := viper.GetString("model")
		if modelName == "" {
			modelName = "gemini-flash-latest" // Default fallback if not set by flag or env (though flag default handles this)
		}

		console.PrintModelInfo(modelName)

		aiClient, err := ai.NewGeminiProvider(ctx, key, modelName)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to initialize AI client: %v", err))
			os.Exit(1)
		}

		// 2. Execution Flow

		mimeType, err := fileSys.GetMimeType(filePath)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to detect mime type: %v", err))
			os.Exit(1)
		}
		console.PrintDetectedType(mimeType)

		if err := domain.IsAllowedMimeType(mimeType); err != nil {
			console.Error(fmt.Sprintf("Validation failed: %v", err))
			os.Exit(1)
		}

		console.PrintAnalyzing(filepath.Base(filePath))
		content, err := fileSys.ReadFile(filePath)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to read file: %v", err))
			os.Exit(1)
		}

		// Generate Name
		currentExt := filepath.Ext(filePath)
		newName, reasoning, err := aiClient.GenerateName(ctx, content, mimeType, currentExt)
		if err != nil {
			console.Error(fmt.Sprintf("AI Generation failed: %v", err))
			os.Exit(1)
		}

		// Sanitize & Domain Logic
		safeName := domain.SanitizeFilename(newName)

		// Collision Check
		// Need absolute path for checking existence in the same dir
		dir := filepath.Dir(filePath)
		finalName := domain.ResolveCollision(safeName, func(name string) bool {
			return fileSys.Exists(filepath.Join(dir, name))
		})

		// 3. User Interaction
		console.PrintProposal(filepath.Base(filePath), finalName, reasoning)

		if dryRun {
			console.PrintDryRun()
			return
		}

		confirm, err := console.Confirm("Rename?")
		if err != nil {
			console.Error(fmt.Sprintf("Input error: %v", err))
			os.Exit(1)
		}

		if confirm {
			newPath := filepath.Join(dir, finalName)
			if err := fileSys.Rename(filePath, newPath); err != nil {
				console.Error(fmt.Sprintf("Rename failed: %v", err))
				os.Exit(1)
			}
			console.PrintSuccess(finalName)
		} else {
			console.PrintCancelled()
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simulate rename without executing")
	rootCmd.PersistentFlags().StringVar(&style, "style", "kebab", "Naming style (kebab, snake)")
	rootCmd.PersistentFlags().StringVar(&model, "model", "gemini-flash-latest", "Gemini model to use")
	_ = viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
}

func initConfig() {
	viper.AutomaticEnv() // Read from env variables
	_ = viper.BindEnv("model", "GEMINI_MODEL")
}
