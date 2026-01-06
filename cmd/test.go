package cmd

import (
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test <stage-uuid>",
	Short: "Run tests locally without submitting",
	Long: `Run tests locally to validate your solution before submitting.

This command fetches the test configuration from the server, runs your code
against the tests, and shows you the results. No submission is made to the server.

Example:
  95 test d533f704-66aa-4dd7-ae7d-f59f505e9839

After tests pass, use '95 run' to submit your solution and track progress.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stageUuid := args[0]
		return runOrTest(stageUuid, false)
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
