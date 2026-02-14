package domain

import "time"

// JobStatus represents the state of an AI job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// AIJob represents a background job for Claude Code execution.
type AIJob struct {
	ID          int64     `json:"id" db:"id"`
	IssueID     int64     `json:"issue_id" db:"issue_id"`
	Status      JobStatus `json:"status" db:"status"`
	Attempts    int       `json:"attempts" db:"attempts"`
	MaxAttempts int       `json:"max_attempts" db:"max_attempts"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	ErrorMsg    *string   `json:"error_msg,omitempty" db:"error_msg"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
