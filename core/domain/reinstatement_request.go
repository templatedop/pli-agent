package domain

import (
	"time"
)

// ReinstatementRequest represents an agent reinstatement request
// E-08: Reinstatement Request Entity
// BR-AGT-PRF-016: Status updates require mandatory reason
type ReinstatementRequest struct {
	RequestID         string     `db:"request_id"`
	AgentID           string     `db:"agent_id"`
	RequestReasonCode string     `db:"request_reason_code"`
	RequestReasonText string     `db:"request_reason_text"`
	EffectiveDate     time.Time  `db:"effective_date"`
	RequestStatus     string     `db:"request_status"`
	ApprovedBy        string     `db:"approved_by"`
	ApprovedAt        *time.Time `db:"approved_at"`
	ApprovalComments  string     `db:"approval_comments"`
	RejectedBy        string     `db:"rejected_by"`
	RejectedAt        *time.Time `db:"rejected_at"`
	RejectionReason   string     `db:"rejection_reason"`
	WorkflowID        string     `db:"workflow_id"`
	WorkflowStatus    string     `db:"workflow_status"`
	CreatedBy         string     `db:"created_by"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedBy         string     `db:"updated_by"`
	UpdatedAt         time.Time  `db:"updated_at"`
	DeletedAt         *time.Time `db:"deleted_at"`
	Version           int        `db:"version"`
}

// Request Status Constants
const (
	RequestStatusPending  = "PENDING"
	RequestStatusApproved = "APPROVED"
	RequestStatusRejected = "REJECTED"
)

// Reinstatement Reason Codes
// BR-AGT-PRF-016: Status updates require mandatory reason
const (
	ReinstatementReasonLicenseRenewed      = "LICENSE_RENEWED"
	ReinstatementReasonAppealApproved      = "APPEAL_APPROVED"
	ReinstatementReasonClearanceObtained   = "CLEARANCE_OBTAINED"
	ReinstatementReasonMutualAgreement     = "MUTUAL_AGREEMENT"
	ReinstatementReasonWrongfulTermination = "WRONGFUL_TERMINATION"
)

// IsPending checks if request is pending approval
func (r *ReinstatementRequest) IsPending() bool {
	return r.RequestStatus == RequestStatusPending
}

// IsApproved checks if request is approved
func (r *ReinstatementRequest) IsApproved() bool {
	return r.RequestStatus == RequestStatusApproved
}

// IsRejected checks if request is rejected
func (r *ReinstatementRequest) IsRejected() bool {
	return r.RequestStatus == RequestStatusRejected
}

// IsEffective checks if reinstatement is effective (approved and effective date passed)
func (r *ReinstatementRequest) IsEffective() bool {
	if !r.IsApproved() {
		return false
	}
	return r.EffectiveDate.Before(time.Now()) || r.EffectiveDate.Equal(time.Now())
}

// CanApprove checks if request can be approved
func (r *ReinstatementRequest) CanApprove() bool {
	return r.IsPending() && r.DeletedAt == nil
}

// CanReject checks if request can be rejected
func (r *ReinstatementRequest) CanReject() bool {
	return r.IsPending() && r.DeletedAt == nil
}

// GetReinstatementReasonName returns human-readable reinstatement reason
func (r *ReinstatementRequest) GetReinstatementReasonName() string {
	switch r.RequestReasonCode {
	case ReinstatementReasonLicenseRenewed:
		return "License Renewed"
	case ReinstatementReasonAppealApproved:
		return "Appeal Approved"
	case ReinstatementReasonClearanceObtained:
		return "Clearance Obtained"
	case ReinstatementReasonMutualAgreement:
		return "Mutual Agreement"
	case ReinstatementReasonWrongfulTermination:
		return "Wrongful Termination"
	default:
		return "Unknown"
	}
}

// GetStatusName returns human-readable status
func (r *ReinstatementRequest) GetStatusName() string {
	switch r.RequestStatus {
	case RequestStatusPending:
		return "Pending Approval"
	case RequestStatusApproved:
		return "Approved"
	case RequestStatusRejected:
		return "Rejected"
	default:
		return "Unknown"
	}
}
