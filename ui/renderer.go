package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/ui/messages"
)

var green lipgloss.Style
var red lipgloss.Style
var gray lipgloss.Style
var borderBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

type testModel struct {
	text     string
	passed   *bool
	finished bool
	stdin    string
	stdout   string
	stderr   string
	exitCode int
}

type stepModel struct {
	step     string
	passed   *bool
	finished bool
	tests    []testModel
}

type rootModel struct {
	steps       []stepModel
	spinner     spinner.Model
	isSubmit    bool
	success     bool
	finalized   bool
	clear       bool
	feedback    string
	totalTests  int
	passedTests int
}

func initModel(isSubmit bool) rootModel {
	s := spinner.New()
	s.Spinner = spinner.Dot

	// Initialize styles
	green = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	red = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	gray = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	return rootModel{
		spinner:  s,
		isSubmit: isSubmit,
		steps:    []stepModel{},
	}
}

func (m rootModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.DoneStepMsg:
		m.feedback = msg.Feedback
		m.totalTests = msg.TotalTests
		m.passedTests = msg.PassedTests
		m.success = msg.Success
		m.clear = true
		return m, tea.Quit

	case messages.StartStepMsg:
		m.steps = append(m.steps, stepModel{
			step:  msg.Step,
			tests: []testModel{},
		})
		return m, nil

	case messages.ResolveStepMsg:
		if msg.Index < len(m.steps) {
			m.steps[msg.Index].passed = msg.Passed
			m.steps[msg.Index].finished = true
		}
		return m, nil

	case messages.StartTestMsg:
		if len(m.steps) > 0 {
			m.steps[len(m.steps)-1].tests = append(
				m.steps[len(m.steps)-1].tests,
				testModel{text: msg.Text},
			)
		}
		return m, nil

	case messages.ResolveTestMsg:
		if msg.StepIndex < len(m.steps) && msg.TestIndex < len(m.steps[msg.StepIndex].tests) {
			m.steps[msg.StepIndex].tests[msg.TestIndex].passed = msg.Passed
			m.steps[msg.StepIndex].tests[msg.TestIndex].finished = true
			m.steps[msg.StepIndex].tests[msg.TestIndex].stdin = msg.Stdin
			m.steps[msg.StepIndex].tests[msg.TestIndex].stdout = msg.Stdout
			m.steps[msg.StepIndex].tests[msg.TestIndex].stderr = msg.Stderr
			m.steps[msg.StepIndex].tests[msg.TestIndex].exitCode = msg.ExitCode
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func renderTestHeader(header string, spinner spinner.Model, isFinished bool, isSubmit bool, passed *bool) string {
	cmdStr := renderTest(header, spinner.View(), isFinished, &isSubmit, passed)
	box := borderBox.Render(fmt.Sprintf(" %s ", cmdStr))
	sliced := strings.Split(box, "\n")
	if len(sliced) > 2 {
		sliced[2] = strings.Replace(sliced[2], "─", "┬", 1)
	}
	return strings.Join(sliced, "\n") + "\n"
}

func renderTests(tests []testModel, spinner string, finalized bool) string {
	var str string
	for _, test := range tests {
		testStr := renderTest(test.text, spinner, test.finished, nil, test.passed)
		testStr = fmt.Sprintf("  %s", testStr)

		edges := " ├─"
		for range lipgloss.Height(testStr) - 1 {
			edges += "\n │ "
		}
		str += lipgloss.JoinHorizontal(lipgloss.Top, edges, testStr)
		str += "\n"

		// Only show details for FAILED tests when finalized
		if finalized && test.finished && test.passed != nil && !*test.passed {
			// Show stdin if present
			if test.stdin != "" {
				str += " │   " + gray.Render("Input:") + "\n"
				lines := strings.Split(test.stdin, "\n")
				for i, line := range lines {
					if i == len(lines)-1 && line == "" {
						continue
					}
					str += " │     " + gray.Render(line) + "\n"
				}
			}

			// Show exit code for failed tests
			str += " │   " + gray.Render(fmt.Sprintf("Exit code: %d", test.exitCode)) + "\n"

			// Show stdout if present
			if test.stdout != "" {
				str += " │   " + gray.Render("Output:") + "\n"
				lines := strings.Split(test.stdout, "\n")
				for i, line := range lines {
					if i == len(lines)-1 && line == "" {
						continue
					}
					str += " │     " + gray.Render(line) + "\n"
				}
			}

			// Show stderr if present
			if test.stderr != "" {
				str += " │   " + red.Render("Error:") + "\n"
				lines := strings.Split(test.stderr, "\n")
				for i, line := range lines {
					if i == len(lines)-1 && line == "" {
						continue
					}
					str += " │     " + red.Render(line) + "\n"
				}
			}

			str += " │\n"
		}
	}
	return str
}

func renderTest(text string, spinner string, isFinished bool, isSubmit *bool, passed *bool) string {
	testStr := ""
	if !isFinished {
		testStr += fmt.Sprintf("%s %s", spinner, text)
	} else if isSubmit != nil && !*isSubmit {
		testStr += text
	} else if passed == nil {
		testStr += gray.Render(fmt.Sprintf("?  %s", text))
	} else if *passed {
		testStr += green.Render(fmt.Sprintf("✓  %s", text))
	} else {
		testStr += red.Render(fmt.Sprintf("X  %s", text))
	}
	return testStr
}

func (m rootModel) View() string {
	if m.clear {
		return ""
	}

	s := m.spinner.View()
	var str string

	for _, step := range m.steps {
		str += renderTestHeader(step.step, m.spinner, step.finished, m.isSubmit, step.passed)
		str += renderTests(step.tests, s, m.finalized)
	}

	if m.finalized {
		str += "\n"
		if m.success {
			str += green.Render(fmt.Sprintf("✓ All %d tests passed!", m.totalTests)) + " 🎉\n\n"
			if m.feedback != "" {
				str += gray.Render(m.feedback) + "\n\n"
			}
		} else {
			str += red.Render(fmt.Sprintf("✗ %d of %d tests failed", m.totalTests-m.passedTests, m.totalTests)) + "\n\n"
			if m.feedback != "" {
				str += gray.Render(m.feedback) + "\n\n"
			}
		}
	}

	return str
}

// StartRenderer creates and runs the simple renderer (boot.dev style)
func StartRenderer(isSubmit bool, ch chan tea.Msg) func(success bool, totalTests, passedTests int, feedback string) {
	var wg sync.WaitGroup
	p := tea.NewProgram(initModel(isSubmit), tea.WithoutSignalHandler())

	wg.Add(1)
	go func() {
		defer wg.Done()
		if model, err := p.Run(); err != nil {
			fmt.Printf("UI Error: %v\n", err)
		} else if r, ok := model.(rootModel); ok {
			r.clear = false
			r.finalized = true
			fmt.Print(r.View())
		}
	}()

	go func() {
		for msg := range ch {
			p.Send(msg)
		}
	}()

	return func(success bool, totalTests, passedTests int, feedback string) {
		ch <- messages.DoneStepMsg{
			Success:     success,
			TotalTests:  totalTests,
			PassedTests: passedTests,
			Feedback:    feedback,
		}
		wg.Wait()
	}
}
