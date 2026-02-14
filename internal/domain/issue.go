package domain

import "time"

// IssueStatus represents the lifecycle state of an issue.
type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusCompleted  IssueStatus = "completed"
	IssueStatusClosed     IssueStatus = "closed"
)

// Issue represents a task within a project.
type Issue struct {
	ID           int64       `json:"id" db:"id"`
	ProjectID    int64       `json:"project_id" db:"project_id"`
	Title        string      `json:"title" db:"title"`
	Body         *string     `json:"body,omitempty" db:"body"`
	Status       IssueStatus `json:"status" db:"status"`
	AISessionID  *string     `json:"ai_session_id,omitempty" db:"ai_session_id"`
	AIResult     *string     `json:"ai_result,omitempty" db:"ai_result"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// WithStatus returns a new Issue with the given status.
func (i Issue) WithStatus(status IssueStatus) Issue {
	return Issue{
		ID:          i.ID,
		ProjectID:   i.ProjectID,
		Title:       i.Title,
		Body:        i.Body,
		Status:      status,
		AISessionID: i.AISessionID,
		AIResult:    i.AIResult,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   time.Now(),
	}
}
