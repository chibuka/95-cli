package cmd

import (
	"fmt"

	"github.com/chibuka/95-cli/client"
	"github.com/chibuka/95-cli/internal/config"
	"github.com/chibuka/95-cli/internal/runner"
	"github.com/chibuka/95-cli/ui"
	"github.com/chibuka/95-cli/ui/messages"
)

func runOrTest(stageUuid string, isSubmit bool) error {
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
		return fmt.Errorf("%w", err)
	}

	if globalCfg.AccessToken == "" {
		return fmt.Errorf("not logged in. Run '95cli login' first")
	}

	// Fetch cascaded tests (stages 1..X) from backend
	cascadedConfig, err := client.FetchCascadedTests(stageUuid, globalCfg)
	if err != nil {
		return fmt.Errorf("failed to fetch tests: %w", err)
	}

	// Start renderer
	ch := make(chan messages.Msg, 10)
	done := ui.StartRenderer(isSubmit, ch)

	// Run tests for all prerequisite stages
	totalTests := 0
	totalPassed := 0

	var lastSubmissionResult *client.SubmissionResult
	if isSubmit {
		fmt.Printf("Running stages 0 through %d (%d total stages)\n\n",
			cascadedConfig.TargetStageNumber, len(cascadedConfig.StagesToRun))
	}

	for stepIdx, stageInfo := range cascadedConfig.StagesToRun {
		// Parse test config for this stage
		testConfig, err := client.ParseStageTests(stageInfo)
		if err != nil {
			done(false, 0, 0, fmt.Sprintf("Failed to parse tests for stage %d: %v", stageInfo.StageNumber, err))
			return nil
		}

		totalTests += len(testConfig.Tests)

		// Send start stage message
		ch <- messages.StartStepMsg{
			StageNumber: stageInfo.StageNumber,
			StageName:   stageInfo.StageName,
		}

		var results []client.TestResult
		passedCount := 0

		for testIdx, test := range testConfig.Tests {
			// Start test
			ch <- messages.StartTestMsg{
				TestName: test.TestName,
				Stdin:    getTestInput(test),
			}

			result, err := runSingleTest(test, testConfig, projectCfg.RunCommand)

			if err != nil {
				ch <- messages.ResolveTestMsg{
					StepIndex: stepIdx,
					TestIndex: testIdx,
					Passed:    nil,
					Stdin:     getTestInput(test),
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
				stdout := formatTestOutput(testConfig.TestType, result)
				ch <- messages.ResolveTestMsg{
					StepIndex: stepIdx,
					TestIndex: testIdx,
					Passed:    nil, // Will be determined by backend validation if isSubmit
					Stdin:     getTestInput(test),
					Stdout:    stdout,
					Stderr:    result.Stderr,
					ExitCode:  result.ExitCode,
				}
				result.TestName = test.TestName
			}
			results = append(results, *result)
		}

		if isSubmit {
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
				failureReasons := make(map[string]string)
				for _, failure := range submissionResult.TestFailures {
					failedTestNames[failure.TestName] = true
					failureReasons[failure.TestName] = failure.Reason
				}

				// Update each test's pass/fail status
				for testIdx, testResult := range results {
					passed := !failedTestNames[testResult.TestName]
					if passed {
						passedCount++
					}
					stdout := formatTestOutput(testConfig.TestType, &testResult)
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    &passed,
						Stdin:     testConfig.Tests[testIdx].Stdin,
						Stdout:    stdout,
						Stderr:    testResult.Stderr,
						ExitCode:  testResult.ExitCode,
					}
				}
			} else {
				// All tests passed
				passedCount = len(results)
				for testIdx, testResult := range results {
					passed := true
					stdout := formatTestOutput(testConfig.TestType, &testResult)
					ch <- messages.ResolveTestMsg{
						StepIndex: stepIdx,
						TestIndex: testIdx,
						Passed:    &passed,
						Stdin:     testConfig.Tests[testIdx].Stdin,
						Stdout:    stdout,
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
	}

	if isSubmit {
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
	} else {
		done(true, totalTests, 0, fmt.Sprintf("Run '95 run %s' to submit your results", stageUuid))
	}

	return nil
}

func runSingleTest(test client.Test, testConfig *client.TestConfig, runCommand string) (*client.TestResult, error) {
	// Execute setup operations
	if err := runner.ExecuteSetup(test.Setup); err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}

	// Ensure cleanup runs even if test fails
	defer func() {
		if test.Cleanup != nil {
			if err := runner.ExecuteCleanup(test.Cleanup); err != nil {
				fmt.Printf("Warning: cleanup failed: %v\n", err)
			}
		}
	}()

	var result *client.TestResult
	var err error

	// Run test based on type
	if testConfig.TestType == "http_server" {
		// Run HTTP test
		if testConfig.ProgramConfig == nil || testConfig.ServerConfig == nil {
			return nil, fmt.Errorf("HTTP test configuration missing programConfig or serverConfig")
		}

		result, err = runner.RunHTTPTest(
			testConfig.ProgramConfig,
			testConfig.ServerConfig,
			runCommand,
			test,
		)
	} else {
		// Run CLI test
		result, err = runner.RunCLITest(
			runCommand,
			test.Stdin,
			test.TimeoutSeconds,
		)
	}

	return result, err
}
