package cmd

import (
	"fmt"

	"github.com/chibuka/95-cli/client"
	"github.com/chibuka/95-cli/internal/config"
	"github.com/chibuka/95-cli/internal/runner"
	"github.com/chibuka/95-cli/ui"
	"github.com/chibuka/95-cli/ui/messages"
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

		// Load project config to get run command
		projectCfg, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}

		if projectCfg.RunCommand == "" {
			return fmt.Errorf("no run command found. Run '95cli init --cmd \"your command\"' first")
		}

		// Load global config to get auth
		globalCfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if globalCfg.AccessToken == "" {
			return fmt.Errorf("not logged in. Run '95cli login' first")
		}

		// Fetch cascaded tests (stages 1..X) from backend
		cascadedConfig, err := client.FetchCascadedTests(stageUuid, globalCfg)
		if err != nil {
			return fmt.Errorf("failed to fetch tests: %w", err)
		}

		// Start renderer (not submitting)
		ch := make(chan messages.Msg, 10)
		done := ui.StartRenderer(false, ch)

		// Run tests for all prerequisite stages
		totalTests := 0

		for stepIdx, stageInfo := range cascadedConfig.StagesToRun {
			// Parse test config for this stage
			testConfig, err := client.ParseStageTests(stageInfo)
			if err != nil {
				done(false, 0, 0, fmt.Sprintf("Failed to parse tests for stage %d: %v", stageInfo.StageNumber, err))
				return fmt.Errorf("failed to parse tests: %w", err)
			}

			totalTests += len(testConfig.Tests)

			// Send start stage message
			ch <- messages.StartStepMsg{
				StageNumber: stageInfo.StageNumber,
				StageName:   stageInfo.StageName,
			}

			// Run tests for this stage
			for testIdx, test := range testConfig.Tests {
				// Start test
				ch <- messages.StartTestMsg{
					TestName: test.TestName,
					Stdin:    test.Stdin,
				}

				// Execute setup operations
				if err := runner.ExecuteSetup(test.Setup); err != nil {
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    nil, // No validation in test mode
						Stdin:     test.Stdin,
						Stdout:    "",
						Stderr:    fmt.Sprintf("Setup failed: %v", err),
						ExitCode:  -1,
					}
					continue
				}

				// Ensure cleanup runs even if test fails
				defer func(cleanup *client.TestCleanup) {
					if cleanup != nil {
						if err := runner.ExecuteCleanup(cleanup); err != nil {
							fmt.Printf("Warning: cleanup failed: %v\n", err)
						}
					}
				}(test.Cleanup)

				result, err := runner.RunTest(projectCfg.RunCommand, test.Stdin, test.TimeoutSeconds)
				if err != nil {
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    nil, // No validation in test mode - just show output
						Stdin:     test.Stdin,
						Stdout:    "",
						Stderr:    err.Error(),
						ExitCode:  -1,
					}
				} else {
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    nil, // No validation in test mode - just show output
						Stdin:     test.Stdin,
						Stdout:    result.Stdout,
						Stderr:    result.Stderr,
						ExitCode:  result.ExitCode,
					}
				}
			}
		}

		// Complete without submission (no validation, just showing outputs)
		done(true, totalTests, 0, fmt.Sprintf("Run '95 run %s' to submit your results", stageUuid))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
