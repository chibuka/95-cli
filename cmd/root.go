/*
Copyright © 2025 MAROUANE BOUFAROUJ <boufaroujmarouan@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/chibuka/95-cli/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "95",
	Short: "Practice coding challenges and level up your skills",
	Long: `95 CLI - Build your coding skills, one challenge at a time

The 95 CLI lets you practice coding challenges from your terminal.
Your code runs locally, gets validated server-side, and tracks your progress.

Quick Start:
  1. Authenticate:      95 login
  2. Setup project:     95 init --cmd "python main.py"
  3. Test locally:      95 test <stage-uuid>
  4. Submit solution:   95 run <stage-uuid>

For more information, visit: https://95.dev`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	config.Init()
}
