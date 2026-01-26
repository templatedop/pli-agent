package domain

import (
	"database/sql"
	"time"
)

// ApprovalRequestStatus represents the status of an approval request
type ApprovalRequestStatus string

const (
	ApprovalStatusPending  ApprovalRequestStatus = "PENDING"
	ApprovalStatusApproved ApprovalRequestStatus = "APPROVED"
	ApprovalStatusRejected ApprovalRequestStatus = "REJECTED"
)

// ApprovalRequest represents a pending approval request for profile updates
// E-09: Approval Request Entity
// BR-AGT-PRF-005: Name Update with Audit Logging
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
type ApprovalRequest struct {
	ApprovalRequestID string                `db:"approval_request_id" json:"approval_request_id"`
	AgentID           string                `db:"agent_id" json:"agent_id"`
	Section           string                `db:"section" json:"section"` // personal_info, address, contact, etc.
	RequestedChanges  string                `db:"requested_changes" json:"requested_changes"` // JSON field
	RequestedBy       string                `db:"requested_by" json:"requested_by"`
	RequestedAt       time.Time             `db:"requested_at" json:"requested_at"`
	Status            ApprovalRequestStatus `db:"status" json:"status"`
	ReviewedBy        sql.NullString        `db:"reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt        sql.NullTime          `db:"reviewed_at" json:"reviewed_at,omitempty"`
	ReviewComments    sql.NullString        `db:"review_comments" json:"review_comments,omitempty"`
	CreatedAt         time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt         sql.NullTime          `db:"updated_at" json:"updated_at,omitempty"`
}

// IsPending checks if the approval request is still pending
func (a *ApprovalRequest) IsPending() bool {
	return a.Status == ApprovalStatusPending
}

// IsApproved checks if the approval request is approved
func (a *ApprovalRequest) IsApproved() bool {
	return a.Status == ApprovalStatusApproved
}

// IsRejected checks if the approval request is rejected
func (a *ApprovalRequest) IsRejected() bool {
	return a.Status == ApprovalStatusRejected
}
