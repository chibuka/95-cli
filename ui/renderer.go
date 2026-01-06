package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/ui/messages"
)

var (
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	orange = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	// red color, unused for now
	_    = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	gray = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type testModel struct {
	name          string
	running       bool
	passed        *bool // nil if not yet validated, true/false after backend validation
	stdin         string
	stdout        string
	stderr        string
	exitCode      int
	failureReason string
	shown         bool // Track if this test has been displayed
}

type stepModel struct {
	stageNumber int
	stageName   string
	passed      *bool // nil if not yet validated, true/false after backend validation
	tests       []testModel
}

// StartRenderer creates and runs the renderer
func StartRenderer(isSubmit bool, ch chan messages.Msg) func(success bool, totalTests, passedTests int, feedback string) {
	currentStep := -1
	steps := []stepModel{}

	fmt.Println() // Initial newline

	go func() {
		for msg := range ch {
			switch msg := msg.(type) {
			case messages.StartStepMsg:
				// New stage starting
				currentStep = len(steps)
				steps = append(steps, stepModel{
					stageNumber: msg.StageNumber,
					stageName:   msg.StageName,
					tests:       []testModel{},
				})

				// Print stage header
				header := fmt.Sprintf("Stage %02d: %s", msg.StageNumber, msg.StageName)
				fmt.Println(header)

			case messages.StartTestMsg:
				// New test starting - only store it, don't print anything
				if currentStep >= 0 {
					steps[currentStep].tests = append(steps[currentStep].tests, testModel{
						name:    msg.TestName,
						running: true,
						stdin:   msg.Stdin,
						shown:   false,
					})
					// No printing here - we'll only print when the test completes
				}

			case messages.ResolveTestMsg:
				// Test completed - update the test in our model and display it
				if msg.StepIndex < len(steps) && msg.TestIndex < len(steps[msg.StepIndex].tests) {
					test := &steps[msg.StepIndex].tests[msg.TestIndex]
					test.running = false
					test.passed = msg.Passed
					test.stdout = msg.Stdout
					test.stderr = msg.Stderr
					test.exitCode = msg.ExitCode
					test.failureReason = msg.FailureReason

					// In test mode (isSubmit=false), show all tests immediately
					// In run mode (isSubmit=true), only show validated tests (passed != nil)
					shouldDisplay := !isSubmit || msg.Passed != nil
					if !test.shown && shouldDisplay {
						test.shown = true
						displayTest(test, msg.TestIndex, len(steps[msg.StepIndex].tests), isSubmit)
					}
				}

			case messages.ResolveStepMsg:
				// Stage completed
				if msg.Index < len(steps) {
					steps[msg.Index].passed = msg.Passed
				}
			}
		}
	}()

	return func(success bool, totalTests, passedTests int, feedback string) {
		close(ch)

		// Wait a bit for remaining messages to be processed
		time.Sleep(100 * time.Millisecond)

		if isSubmit {
			// Print final summary
			fmt.Println()
			if success {
				fmt.Println(orange.Render(fmt.Sprintf("✓ All %d tests passed!", totalTests)))
				fmt.Println()
				fmt.Println(gray.Render("→ Check your browser for live progress updates and stage completion!"))
			} else {
				fmt.Println(orange.Render(fmt.Sprintf("✗ %d/%d tests passed", passedTests, totalTests)))
			}
		}

		if feedback != "" {
			fmt.Println(gray.Render(feedback))
		}
		fmt.Println()
	}
}

// checks if stderr contains only build/compilation noise (not actual errors)
func isBuildNoise(stderr string) bool {
	// Common patterns in build output that aren't actual errors
	buildPatterns := []string{
		"Finished",        // Rust cargo
		"Compiling",       // Rust cargo
		"Running",         // Rust cargo, Go
		"Build succeeded", // Various tools
		"Build complete",  // Various tools
	}

	lines := strings.Split(stderr, "\n")
	meaningfulLines := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check if line matches build noise patterns
		isBuild := false
		for _, pattern := range buildPatterns {
			if strings.Contains(trimmed, pattern) {
				isBuild = true
				break
			}
		}

		if !isBuild {
			meaningfulLines++
		}
	}

	// If all non-empty lines are build noise, consider it not an error
	return meaningfulLines == 0
}

// displayTest prints a completed test with all its details
func displayTest(test *testModel, testIndex int, totalTestsInStage int, isSubmit bool) {
	// Determine if this is the last test in the stage
	isLast := testIndex == totalTestsInStage-1
	connector := "├─"
	if isLast {
		connector = "└─"
	}

	indent := "     "

	// TEST MODE: Show all stdin/stdout without validation icons
	if !isSubmit {
		// Print test name without status icon (no validation)
		fmt.Printf("  %s %s\n", connector, test.name)

		// Always show command/output pairs in test mode
		if test.stdin != "" {
			commands := strings.Split(strings.TrimRight(test.stdin, "\n"), "\n")

			// Parse output by splitting on "$ " to get each command's response
			outputParts := strings.Split(test.stdout, "$ ")
			// Remove empty first element (before first prompt)
			if len(outputParts) > 0 && outputParts[0] == "" {
				outputParts = outputParts[1:]
			}

			fmt.Println()
			for i, cmd := range commands {
				// Print command
				fmt.Println(indent + orange.Render("$ ") + cmd)

				// Match output for this command
				if i < len(outputParts) {
					output := strings.TrimSpace(outputParts[i])

					if output == "" {
						fmt.Println(indent + gray.Render("  (no output)"))
					} else {
						// Split multi-line outputs
						for _, line := range strings.Split(output, "\n") {
							if line != "" {
								fmt.Println(indent + gray.Render("  "+line))
							}
						}
					}
				}
				fmt.Println()
			}
		}

		// Show stderr if present (but skip common build noise)
		if test.stderr != "" && !isBuildNoise(test.stderr) {
			fmt.Println(indent + orange.Render("Error:"))
			formatErrorOutput(test.stderr, indent)
			fmt.Println()
		}
		return
	}

	// RUN MODE: Show validation status, only show details for failures
	// Determine status icon (only validated tests are displayed)
	var statusIcon string
	if test.passed != nil && *test.passed {
		statusIcon = green.Render("✓")
	} else {
		statusIcon = orange.Render("✗")
	}

	// Print the test result line
	fmt.Printf("  %s %s %s\n", connector, statusIcon, test.name)

	// Only show details for FAILED tests (keeps output clean for passing tests)
	isPassed := test.passed != nil && *test.passed
	if !isPassed {
		// Parse stdin commands and match with output
		if test.stdin != "" {
			commands := strings.Split(strings.TrimRight(test.stdin, "\n"), "\n")

			// Parse output by splitting on "$ " to get each command's response
			outputParts := strings.Split(test.stdout, "$ ")
			// Remove empty first element (before first prompt)
			if len(outputParts) > 0 && outputParts[0] == "" {
				outputParts = outputParts[1:]
			}

			fmt.Println()
			for i, cmd := range commands {
				// Print command
				fmt.Println(indent + orange.Render("$ ") + gray.Render(cmd))

				// Match output for this command
				if i < len(outputParts) {
					output := strings.TrimSpace(outputParts[i])

					if output == "" {
						fmt.Println(indent + gray.Render("  (no output)"))
					} else {
						// Split multi-line outputs
						for _, line := range strings.Split(output, "\n") {
							if line != "" {
								fmt.Println(indent + gray.Render("  "+line))
							}
						}
					}
				}
				fmt.Println()
			}
		}

		// Show stderr if present (but skip common build noise)
		if test.stderr != "" && !isBuildNoise(test.stderr) {
			fmt.Println(indent + orange.Render("Error:"))
			formatErrorOutput(test.stderr, indent)
			fmt.Println()
		}

		if test.failureReason != "" {
			fmt.Println(indent + orange.Render("Error: "+test.failureReason))
			fmt.Println()
		}
	}
}

// formatErrorOutput formats error messages for better readability
func formatErrorOutput(stderr string, indent string) {
	lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check if this is an HTTP request line (e.g., "GET /")
		if strings.HasPrefix(trimmed, "GET ") || strings.HasPrefix(trimmed, "POST ") ||
			strings.HasPrefix(trimmed, "PUT ") || strings.HasPrefix(trimmed, "DELETE ") ||
			strings.HasPrefix(trimmed, "PATCH ") {
			// Format as a command-like line
			fmt.Println(indent + orange.Render("  $ ") + gray.Render(trimmed))
		} else if strings.Contains(trimmed, "→") {
			// Format guidance/hint lines with proper indentation
			// Split on "→" to handle cases where it's in the middle
			parts := strings.SplitN(trimmed, "→", 2)
			if len(parts) == 2 {
				before := strings.TrimSpace(parts[0])
				guidance := strings.TrimSpace(parts[1])
				if before != "" {
					fmt.Println(indent + orange.Render("  "+before))
				}
				// Split multi-line guidance
				guidanceLines := strings.Split(guidance, "\n")
				for i, gline := range guidanceLines {
					gline = strings.TrimSpace(gline)
					if gline != "" {
						if i == 0 {
							fmt.Println(indent + gray.Render("  → "+gline))
						} else {
							fmt.Println(indent + gray.Render("    "+gline))
						}
					}
				}
			} else {
				// Fallback: just print the line
				fmt.Println(indent + orange.Render("  "+line))
			}
		} else {
			// Regular error line
			fmt.Println(indent + orange.Render("  "+line))
		}
	}
}
