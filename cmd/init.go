package cmd

import (
	"fmt"

	"github.com/chibuka/95-cli/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [--cmd <command>] or init <command>",
	Short: "Initialize project with run command",
	Long: `Initialize your project by specifying how to run your code.

The run command is used to execute your program during tests.

Examples:
  95 init --cmd "python main.py"
  95 init --cmd "node index.js"
  95 init --cmd "go run main.go"
  95 init --cmd "./my-binary"

  # Shorthand (positional argument):
  95 init "python main.py"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var runCommand string

		// Support both --cmd flag and positional argument
		if len(args) > 0 {
			runCommand = args[0]
		} else {
			var err error
			runCommand, err = cmd.Flags().GetString("cmd")
			if err != nil {
				return fmt.Errorf("failed to get cmd flag: %w", err)
			}
		}

		if runCommand == "" {
			return fmt.Errorf("run command cannot be empty. Use: 95 init --cmd \"<command>\" or 95 init \"<command>\"")
		}

		// Detect language from run command
		language := config.DetectLanguage(runCommand)

		// Save project config
		err := config.SaveProjectConfig(runCommand, language)
		if err != nil {
			return fmt.Errorf("failed to save project config: %w", err)
		}

		fmt.Printf("✓ Project initialized!\n")
		fmt.Printf("  Run command: %s\n", runCommand)
		fmt.Printf("  Language: %s\n", language)
		fmt.Println("\nTip: Make sure your command includes the entry point file in case you runCommand needs it (e.g., 'python main.py')")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("cmd", "", "Command to run your program (e.g., 'python main.py')")
}
