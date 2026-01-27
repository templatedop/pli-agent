package domain

import (
	"database/sql"
	"time"
)

// AgentNotification represents a notification sent to an agent
// Phase 9: Search & Dashboard APIs
// AGT-077: Agent Notification History
// FR-AGT-PRF-021: Self-Service Update notifications
type AgentNotification struct {
	NotificationID   string         `json:"notification_id" db:"notification_id"`
	AgentID          string         `json:"agent_id" db:"agent_id"`
	NotificationType string         `json:"notification_type" db:"notification_type"` // EMAIL, SMS, INTERNAL
	Template         string         `json:"template" db:"template"`
	Recipient        string         `json:"recipient" db:"recipient"`
	Subject          sql.NullString `json:"subject" db:"subject"`
	Message          sql.NullString `json:"message" db:"message"`
	SentAt           time.Time      `json:"sent_at" db:"sent_at"`
	DeliveredAt      sql.NullTime   `json:"delivered_at" db:"delivered_at"`
	ReadAt           sql.NullTime   `json:"read_at" db:"read_at"`
	Status           string         `json:"status" db:"status"` // SENT, DELIVERED, READ, FAILED
	FailureReason    sql.NullString `json:"failure_reason" db:"failure_reason"`
	Metadata         sql.NullString `json:"metadata" db:"metadata"` // JSONB
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
}

// Notification status constants
const (
	NotificationStatusSent      = "SENT"
	NotificationStatusDelivered = "DELIVERED"
	NotificationStatusRead      = "READ"
	NotificationStatusFailed    = "FAILED"
)

// Notification type constants
const (
	NotificationTypeEmail    = "EMAIL"
	NotificationTypeSMS      = "SMS"
	NotificationTypeInternal = "INTERNAL"
)

// AgentSearchResult represents a search result for AGT-022
// Phase 9: Multi-criteria agent search
type AgentSearchResult struct {
	AgentID            string         `json:"agent_id" db:"agent_id"`
	AgentCode          sql.NullString `json:"agent_code" db:"agent_code"`
	Name               string         `json:"name" db:"name"` // Concatenated full name
	AgentType          string         `json:"agent_type" db:"agent_type"`
	PANNumber          string         `json:"pan_number" db:"pan_number"`
	MobileNumber       sql.NullString `json:"mobile_number" db:"mobile_number"`
	EmailAddress       sql.NullString `json:"email_address" db:"email_address"`
	Status             string         `json:"status" db:"status"`
	AdvisorCoordinator sql.NullString `json:"advisor_coordinator" db:"advisor_coordinator"`
	OfficeName         sql.NullString `json:"office_name" db:"office_name"`
	OfficeCode         string         `json:"office_code" db:"office_code"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
}

// AgentProfileComplete represents a complete agent profile with all related entities
// Phase 9: AGT-023 Get Complete Profile
// Uses JSON aggregation to fetch everything in single query
type AgentProfileComplete struct {
	AgentProfile
	Addresses    []AgentAddress     `json:"addresses"`
	Contacts     []AgentContact     `json:"contacts"`
	Emails       []AgentEmail       `json:"emails"`
	BankDetails  []AgentBankDetails `json:"bank_details"`
	Licenses     []AgentLicense     `json:"licenses"`
	WorkflowInfo *WorkflowInfo      `json:"workflow_info,omitempty"`
}

// WorkflowInfo represents current workflow state for AGT-023
type WorkflowInfo struct {
	CurrentStep        string  `json:"current_step"`
	ProgressPercentage float64 `json:"progress_percentage"`
}

// HierarchyNode represents a node in agent hierarchy chain
// Phase 9: AGT-073 Get Agent Hierarchy
// Uses recursive CTE to build hierarchy
type HierarchyNode struct {
	AgentID   string         `json:"agent_id" db:"agent_id"`
	AgentCode sql.NullString `json:"agent_code" db:"agent_code"`
	Name      string         `json:"name" db:"name"`
	AgentType string         `json:"agent_type" db:"agent_type"`
	Level     int            `json:"level" db:"level"` // 1 = current agent, 2 = coordinator, 3 = manager
}

// TimelineEvent represents an event in agent activity timeline
// Phase 9: AGT-076 Agent Activity Timeline
// Combines audit logs, license changes, status changes
type TimelineEvent struct {
	Timestamp    time.Time      `json:"timestamp" db:"timestamp"`
	EventType    string         `json:"event_type" db:"event_type"` // PROFILE_CHANGE, LICENSE_UPDATE, STATUS_CHANGE
	Description  string         `json:"description" db:"description"`
	PerformedBy  sql.NullString `json:"performed_by" db:"performed_by"`
	FieldName    sql.NullString `json:"field_name" db:"field_name"`
	OldValue     sql.NullString `json:"old_value" db:"old_value"`
	NewValue     sql.NullString `json:"new_value" db:"new_value"`
	ActionReason sql.NullString `json:"action_reason" db:"action_reason"`
}

// DashboardMetrics represents dashboard metrics for AGT-068
// Phase 9: Agent Dashboard
type DashboardMetrics struct {
	ProfileSummary      ProfileSummary      `json:"profile_summary"`
	PerformanceMetrics  PerformanceMetrics  `json:"performance_metrics"`
	PendingTasks        []PendingTask       `json:"pending_tasks"`
	RecentNotifications []AgentNotification `json:"recent_notifications"`
}

// ProfileSummary for dashboard
type ProfileSummary struct {
	AgentID   string `json:"agent_id"`
	Name      string `json:"name"`
	AgentType string `json:"agent_type"`
	Status    string `json:"status"`
	PANNumber string `json:"pan_number"`
}

// PerformanceMetrics for dashboard
type PerformanceMetrics struct {
	PoliciesSold     int     `json:"policies_sold"`
	PremiumCollected float64 `json:"premium_collected"`
	TargetsAchieved  struct {
		MonthlyTarget int     `json:"monthly_target"`
		Achieved      int     `json:"achieved"`
		Percentage    float64 `json:"percentage"`
	} `json:"targets_achieved"`
}

// PendingTask represents a pending task for agent
type PendingTask struct {
	Task     string    `json:"task"`
	Priority string    `json:"priority"` // HIGH, MEDIUM, LOW
	DueDate  time.Time `json:"due_date"`
	Overdue  bool      `json:"overdue"`
}

// PaginationMetadata for search results
type PaginationMetadata struct {
	CurrentPage    int `json:"current_page"`
	TotalPages     int `json:"total_pages"`
	TotalResults   int `json:"total_results"`
	ResultsPerPage int `json:"results_per_page"`
}

// SearchFilters for AGT-022 multi-criteria search
type SearchFilters struct {
	AgentID      *string
	Name         *string
	PANNumber    *string
	MobileNumber *string
	Status       *string
	OfficeCode   *string
	AgentType    *string
}
