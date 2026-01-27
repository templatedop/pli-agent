package request

import (
	"time"
)

// TerminateAgentRequest represents request to terminate an agent
// AGT-039: Terminate Agent
// BR-AGT-PRF-016: Status Update with Reason
// BR-AGT-PRF-017: Agent Termination Workflow
type TerminateAgentRequest struct {
	// Termination Details
	TerminationReason     string    `json:"termination_reason" validate:"required,min=20" example:"Agent violated company policies regarding fraudulent claims"`
	TerminationReasonCode string    `json:"termination_reason_code" validate:"required,oneof=RESIGNATION MISCONDUCT NON_PERFORMANCE FRAUD LICENSE_EXPIRED OTHER" example:"MISCONDUCT"`
	EffectiveDate         time.Time `json:"effective_date" validate:"required" example:"2024-01-15T00:00:00Z"`
	TerminatedBy          string    `json:"terminated_by" validate:"required,min=3" example:"admin@company.com"`

	// Optional Workflow ID (if using Temporal)
	WorkflowID string `json:"workflow_id,omitempty" example:"termination-workflow-AGT-001-2024"`

	// Optional Comments
	Comments string `json:"comments,omitempty" example:"Additional notes about termination"`
}

// ReinstateAgentRequest represents request to reinstate a terminated agent
// AGT-041: Reinstate Agent
// WF-AGT-PRF-011: Reinstatement Workflow
type ReinstateAgentRequest struct {
	// Reinstatement Details
	ReinstatementReason string `json:"reinstatement_reason" validate:"required,min=10" example:"Agent has completed corrective training and shown improvement"`
	RequestedBy         string `json:"requested_by" validate:"required,min=3" example:"admin@company.com"`

	// Optional Workflow ID (if using Temporal)
	WorkflowID string `json:"workflow_id,omitempty" example:"reinstatement-workflow-AGT-001-2024"`

	// Optional Comments
	Comments string `json:"comments,omitempty" example:"Additional notes about reinstatement request"`
}

// ApproveReinstatementRequest represents approval of a reinstatement request
type ApproveReinstatementRequest struct {
	ApprovedBy string `json:"approved_by" validate:"required,min=3" example:"manager@company.com"`
	Conditions string `json:"conditions,omitempty" example:"Agent must complete quarterly compliance training"`
	ProbationDays int `json:"probation_days" validate:"omitempty,min=0,max=365" example:"90"`
}

// RejectReinstatementRequest represents rejection of a reinstatement request
type RejectReinstatementRequest struct {
	RejectedBy      string `json:"rejected_by" validate:"required,min=3" example:"manager@company.com"`
	RejectionReason string `json:"rejection_reason" validate:"required,min=10" example:"Insufficient evidence of behavior improvement"`
}
