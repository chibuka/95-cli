package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <stage-uuid>",
	Short: "Run tests and submit results for validation",
	Long: `Run tests against your code and submit results to the server.

This command fetches tests, runs your code locally, and submits the results
for server-side validation. If all tests pass, your progress is saved.

Example:
  95 run d533f704-66aa-4dd7-ae7d-f59f505e9839

Tip: Use '95 test' first to validate locally before submitting.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stageUuid := args[0]
		return runOrTest(stageUuid, true)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
