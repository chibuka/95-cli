package messages

// Msg is a marker interface for all message types
type Msg any

// StartStepMsg is sent when a stage begins
type StartStepMsg struct {
	StageNumber int
	StageName   string
}

// StartTestMsg is sent when a test begins execution
type StartTestMsg struct {
	TestName string
	Stdin    string
}

// ResolveTestMsg is sent when a test completes
type ResolveTestMsg struct {
	StepIndex     int
	TestIndex     int
	Passed        *bool // nil if not yet validated, true/false after backend validation
	Stdin         string
	Stdout        string
	Stderr        string
	ExitCode      int
	FailureReason string
}

// ResolveStepMsg is sent when a stage/step completes
type ResolveStepMsg struct {
	Index  int
	Passed *bool
}

// DoneStepMsg is sent when everything is complete
type DoneStepMsg struct{}
