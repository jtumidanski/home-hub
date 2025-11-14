package reminder

// Status represents the state of a reminder
type Status string

const (
	// StatusActive indicates a reminder that is pending
	StatusActive Status = "active"
	// StatusSnoozed indicates a reminder that has been snoozed
	StatusSnoozed Status = "snoozed"
	// StatusDismissed indicates a reminder that has been dismissed
	StatusDismissed Status = "dismissed"
)

// IsValid checks if a status value is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusSnoozed, StatusDismissed:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}
