package task

// Status represents the state of a task
type Status string

const (
	// StatusIncomplete indicates a task that has not been completed
	StatusIncomplete Status = "incomplete"
	// StatusComplete indicates a task that has been completed
	StatusComplete Status = "complete"
)

// IsValid checks if a status value is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusIncomplete, StatusComplete:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}
