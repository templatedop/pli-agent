package domain

import (
	"database/sql"
	"time"
)

// AgentLicenseReminder represents agent license reminder log entity
// E-07: Agent License Reminder Log Entity
// BR-AGT-PRF-014
type AgentLicenseReminder struct {
	// Primary Key
	ReminderID string `json:"reminder_id" db:"reminder_id"`

	// Foreign Key
	LicenseID string `json:"license_id" db:"license_id"`

	// Reminder Details (BR-AGT-PRF-014)
	ReminderType string       `json:"reminder_type" db:"reminder_type"` // 30_DAYS, 15_DAYS, 7_DAYS, EXPIRY_DAY
	ReminderDate time.Time    `json:"reminder_date" db:"reminder_date"`
	SentDate     sql.NullTime `json:"sent_date" db:"sent_date"`

	// Sent Status
	SentStatus string `json:"sent_status" db:"sent_status"` // PENDING, SENT, FAILED

	// Sent Flags
	EmailSent bool `json:"email_sent" db:"email_sent"`
	SMSSent   bool `json:"sms_sent" db:"sms_sent"`

	// Failure Tracking
	FailureReason sql.NullString `json:"failure_reason" db:"failure_reason"`
	RetryCount    int            `json:"retry_count" db:"retry_count"`

	// Metadata
	Metadata sql.NullString `json:"metadata" db:"metadata"` // JSONB

	// Audit Fields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy string    `json:"created_by" db:"created_by"`
}

// ReminderType constants (BR-AGT-PRF-014)
const (
	ReminderType30Days    = "30_DAYS"
	ReminderType15Days    = "15_DAYS"
	ReminderType7Days     = "7_DAYS"
	ReminderTypeExpiryDay = "EXPIRY_DAY"
)

// ReminderStatus constants
const (
	ReminderStatusPending = "PENDING"
	ReminderStatusSent    = "SENT"
	ReminderStatusFailed  = "FAILED"
)
