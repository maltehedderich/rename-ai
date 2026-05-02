package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/maltehedderich/rename-ai/internal/adapters/ai"
	"github.com/maltehedderich/rename-ai/internal/adapters/fs"
	"github.com/maltehedderich/rename-ai/internal/adapters/ui"
)

var (
	dryRun bool
	style  string
	model  string
)

var rootCmd = &cobra.Command{
	Use:   "rnai [file ...]",
	Short: "Rename files using GenAI",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		console := ui.NewConsoleUI()
		fileSys := fs.NewOsFileSystem()

		if code := runCLI(ctx, args, dryRun, console, fileSys); code != 0 {
			os.Exit(1)
		}
	},
}

func runCLI(ctx context.Context, args []string, dryRun bool, console *ui.ConsoleUI, fileSys fileSystem) int {
	filePaths, err := chooseInputFiles(args, console, fileSys)
	if err != nil {
		if errors.Is(err, ui.ErrSelectionCancelled) {
			console.PrintCancelled()
			return 0
		}
		console.Error(fmt.Sprintf("Failed to select files: %v", err))
		return 1
	}
	if len(filePaths) == 0 {
		console.Info("No files selected.")
		return 0
	}
	if len(filePaths) == 1 && !fileSys.Exists(filePaths[0]) {
		console.Error(fmt.Sprintf("File not found: %s", filePaths[0]))
		return 1
	}

	key := viper.GetString("GEMINI_API_KEY")
	if key == "" {
		console.Error("GEMINI_API_KEY invalid. Please set GEMINI_API_KEY environment variable.")
		return 1
	}

	modelName := viper.GetString("model")
	if modelName == "" {
		modelName = "gemini-flash-latest"
	}

	console.PrintModelInfo(modelName)

	aiClient, err := ai.NewGeminiProvider(ctx, key, modelName)
	if err != nil {
		console.Error(fmt.Sprintf("Failed to initialize AI client: %v", err))
		return 1
	}

	if len(filePaths) == 1 {
		return runSingleRename(ctx, filePaths[0], dryRun, console, fileSys, aiClient)
	}
	return runBatchRename(ctx, filePaths, dryRun, console, fileSys, aiClient)
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
