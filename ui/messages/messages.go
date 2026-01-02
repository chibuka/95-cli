package messages

// Message types for Bubble Tea communication
// Matching boot.dev's message-based architecture

// StartStepMsg is sent when a step begins
type StartStepMsg struct {
	Step string
}

// ResolveStepMsg is sent when a step completes
type ResolveStepMsg struct {
	Index  int
	Passed *bool
}

// StartTestMsg is sent when a test begins execution
type StartTestMsg struct {
	Text string
}

// ResolveTestMsg is sent when a test completes
type ResolveTestMsg struct {
	StepIndex int
	TestIndex int
	Passed    *bool
	Stdin     string
	Stdout    string
	Stderr    string
	ExitCode  int
}

// DoneStepMsg is sent when everything is complete
type DoneStepMsg struct {
	Success     bool
	TotalTests  int
	PassedTests int
	Feedback    string
}
