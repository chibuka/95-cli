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

		// Load project config to get run command
		projectCfg, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}

		if projectCfg.RunCommand == "" {
			return fmt.Errorf("no run command found. Run '95cli init --cmd \"your command\"' first")
		}

		if projectCfg.Language == "" {
			return fmt.Errorf("no language detected. Please re-run '95cli init --cmd \"your command\"' to update project config")
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

		// Show what we're running
		fmt.Printf("Running stages 0 through %d (%d total stages)\n\n",
			cascadedConfig.TargetStageNumber, len(cascadedConfig.StagesToRun))

		// Start renderer
		ch := make(chan messages.Msg, 10)
		done := ui.StartRenderer(true, ch)

		// Run tests for all prerequisite stages
		totalTests := 0
		totalPassed := 0
		var lastSubmissionResult *client.SubmissionResult

		for stepIdx, stageInfo := range cascadedConfig.StagesToRun {
			// Parse test config for this stage
			testConfig, err := client.ParseStageTests(stageInfo)
			if err != nil {
				done(false, totalTests, totalPassed, fmt.Sprintf("Failed to parse tests for stage %d: %v", stageInfo.StageNumber, err))
				return nil
			}

			totalTests += len(testConfig.Tests)

			// Send start step message
			ch <- messages.StartStepMsg{
				StageNumber: stageInfo.StageNumber,
			StageName:   stageInfo.StageName,
			}

			// Run tests for this stage
			var results []client.TestResult
			passedCount := 0

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
				Passed:    nil,
				Stdin:     test.Stdin,
				Stdout:    "",
				Stderr:    fmt.Sprintf("Setup failed: %v", err),
				ExitCode:  -1,
			}
			results = append(results, client.TestResult{
				TestName: test.TestName,
				ExitCode: -1,
				Stderr:   fmt.Sprintf("Setup failed: %v", err),
			})
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
					// Test execution error (not validation failure)
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    nil, // Unknown until backend validates
						Stdin:     test.Stdin,
						Stdout:    "",
						Stderr:    err.Error(),
						ExitCode:  -1,
					}
					result = &client.TestResult{
						TestName: test.TestName,
						ExitCode: -1,
						Stderr:   err.Error(),
					}
				} else {
					// Test ran successfully, send output (don't determine pass/fail locally)
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    nil, // Will be determined by backend validation
						Stdin:     test.Stdin,
						Stdout:    result.Stdout,
						Stderr:    result.Stderr,
						ExitCode:  result.ExitCode,
					}
					result.TestName = test.TestName
				}
				results = append(results, *result)
			}

			// Submit results for this stage to backend for validation
			submissionResult, err := client.SubmitResults(
				stageInfo.StageUuid,
				projectCfg.Language,
				globalCfg,
				results,
				&cascadedConfig.TargetStageNumber,
			)
			if err != nil {
				done(false, totalTests, totalPassed, fmt.Sprintf("Submission failed for stage %d: %v", stageInfo.StageNumber, err))
				return nil
			}

			lastSubmissionResult = submissionResult

			// Update pass/fail status based on backend validation
			if submissionResult.TestFailures != nil {
				// Map test failures to update UI
				failedTestNames := make(map[string]bool)
				for _, failure := range submissionResult.TestFailures {
					failedTestNames[failure.TestName] = true
				}

				// Update each test's pass/fail status
				for testIdx, testResult := range results {
					passed := !failedTestNames[testResult.TestName]
					if passed {
						passedCount++
					}
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    &passed,
						Stdin:     testConfig.Tests[testIdx].Stdin,
						Stdout:    testResult.Stdout,
						Stderr:    testResult.Stderr,
						ExitCode:  testResult.ExitCode,
					}
				}
			} else {
				// All tests passed
				passedCount = len(results)
				for testIdx, testResult := range results {
					passed := true
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    &passed,
						Stdin:     testConfig.Tests[testIdx].Stdin,
						Stdout:    testResult.Stdout,
						Stderr:    testResult.Stderr,
						ExitCode:  testResult.ExitCode,
					}
				}
			}

			totalPassed += passedCount

			// Mark step as complete with backend validation result
			ch <- messages.ResolveStepMsg{
				Index:  stepIdx,
				Passed: &submissionResult.Passed,
			}

			// If this stage failed, stop running further stages
			if !submissionResult.Passed {
				fmt.Printf("\n⚠ Stage %d failed. Stopping at stage %d (requested stage %d)\n\n",
					stageInfo.StageNumber, stageInfo.StageNumber, cascadedConfig.TargetStageNumber)
				break
			}
		}

		// Complete with results from the last submission
		if lastSubmissionResult != nil {
			done(
				lastSubmissionResult.Passed,
				totalTests,
				totalPassed,
				lastSubmissionResult.Feedback,
			)
		} else {
			done(false, totalTests, totalPassed, "No tests were run")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
