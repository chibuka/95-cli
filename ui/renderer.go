package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/ui/messages"
)

var (
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	gray   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

type testModel struct {
	name     string
	running  bool
	passed   *bool  // nil if not yet validated, true/false after backend validation
	stdin    string
	stdout   string
	stderr   string
	exitCode int
}

type stepModel struct {
	stageNumber int
	stageName   string
	passed      *bool  // nil if not yet validated, true/false after backend validation
	tests       []testModel
}

// StartRenderer creates and runs the renderer
func StartRenderer(isSubmit bool, ch chan messages.Msg) func(success bool, totalTests, passedTests int, feedback string) {
	currentStep := -1
	currentTest := -1
	steps := []stepModel{}

	fmt.Println() // Initial newline

	go func() {
		for msg := range ch {
			switch msg := msg.(type) {
			case messages.StartStepMsg:
				// New stage starting
				currentStep = len(steps)
				currentTest = -1
				steps = append(steps, stepModel{
					stageNumber: msg.StageNumber,
					stageName:   msg.StageName,
					tests:       []testModel{},
				})

				// Print stage header
				header := fmt.Sprintf("Stage %02d: %s", msg.StageNumber, msg.StageName)
				fmt.Println(cyan.Render("● ") + header)

			case messages.StartTestMsg:
				// New test starting
				if currentStep >= 0 {
					currentTest = len(steps[currentStep].tests)
					steps[currentStep].tests = append(steps[currentStep].tests, testModel{
						name:    msg.TestName,
						running: true,
						stdin:   msg.Stdin,
					})

					// Print test starting (we'll update this line when it completes)
					isLast := currentTest == len(steps[currentStep].tests)-1
					connector := "├─"
					if isLast {
						connector = "└─"
					}
					fmt.Printf("  %s ⧗ %s\n", connector, msg.TestName)
				}

			case messages.ResolveTestMsg:
				// Test completed - update the test in our model
				if msg.StepIndex < len(steps) && msg.TestIndex < len(steps[msg.StepIndex].tests) {
					test := &steps[msg.StepIndex].tests[msg.TestIndex]
					test.running = false
					test.passed = msg.Passed
					test.stdout = msg.Stdout
					test.stderr = msg.Stderr
					test.exitCode = msg.ExitCode

					// Move cursor up one line and clear it, then print updated status
					fmt.Print("\033[1A\033[2K")

					isLast := msg.TestIndex == len(steps[msg.StepIndex].tests)-1
					connector := "├─"
					if isLast {
						connector = "└─"
					}

					// Determine status icon
					statusIcon := yellow.Render("⧗")
					if test.passed != nil {
						if *test.passed {
							statusIcon = green.Render("✓")
						} else {
							statusIcon = red.Render("✗")
						}
					}

					fmt.Printf("  %s %s %s\n", connector, statusIcon, gray.Render(test.name))

					// Print test details indented
					indent := "  │  "
					if isLast {
						indent = "     "
					}

					// Show stdin if present
					if test.stdin != "" {
						fmt.Println(indent + gray.Render("Stdin:"))
						for _, line := range strings.Split(strings.TrimRight(test.stdin, "\n"), "\n") {
							fmt.Println(indent + gray.Render("  "+line))
						}
					}

					// Show stdout if present
					if test.stdout != "" {
						fmt.Println(indent + gray.Render("Output:"))
						for _, line := range strings.Split(strings.TrimRight(test.stdout, "\n"), "\n") {
							fmt.Println(indent + gray.Render("  "+line))
						}
					}

					// Show stderr only if test failed
					if test.exitCode != 0 && test.stderr != "" {
						fmt.Println(indent + red.Render("Error:"))
						for _, line := range strings.Split(strings.TrimRight(test.stderr, "\n"), "\n") {
							fmt.Println(indent + red.Render("  "+line))
						}
					}

					fmt.Println() // Extra line after test details
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
		// (hacky but simple solution)
		// TODO: use proper synchronization
		// time.Sleep(100 * time.Millisecond)

		// Print final summary
		fmt.Println()
		if success {
			fmt.Println(green.Render(fmt.Sprintf("✓ All %d tests passed!", totalTests)))
		} else {
			fmt.Println(red.Render(fmt.Sprintf("✗ %d/%d tests passed", passedTests, totalTests)))
		}

		if feedback != "" {
			fmt.Println(gray.Render(feedback))
		}
		fmt.Println()
	}
}
