package domain

import "time"

// NotificationType represents the kind of notification.
type NotificationType string

const (
	NotificationIssueCreated   NotificationType = "issue_created"
	NotificationIssueCompleted NotificationType = "issue_completed"
	NotificationIssueFailed    NotificationType = "issue_failed"
	NotificationAIStarted      NotificationType = "ai_started"
)

// Notification represents an in-app notification for a user.
type Notification struct {
	ID        int64            `json:"id" db:"id"`
	UserID    int64            `json:"user_id" db:"user_id"`
	IssueID   *int64           `json:"issue_id,omitempty" db:"issue_id"`
	Type      NotificationType `json:"type" db:"type"`
	Title     string           `json:"title" db:"title"`
	Message   string           `json:"message" db:"message"`
	Read      bool             `json:"read" db:"read"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}
