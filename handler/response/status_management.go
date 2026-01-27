package response

import (
	"database/sql"
	"time"

	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
)

// TerminationRecordDTO represents a termination record
type TerminationRecordDTO struct {
	TerminationID         string     `json:"termination_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	AgentID               string     `json:"agent_id" example:"AGT-001"`
	TerminationDate       time.Time  `json:"termination_date" example:"2024-01-15T10:30:00Z"`
	EffectiveDate         time.Time  `json:"effective_date" example:"2024-01-15T00:00:00Z"`
	TerminationReason     string     `json:"termination_reason" example:"Agent violated company policies"`
	TerminationReasonCode string     `json:"termination_reason_code" example:"MISCONDUCT"`
	TerminatedBy          string     `json:"terminated_by" example:"admin@company.com"`
	WorkflowID            *string    `json:"workflow_id,omitempty" example:"termination-workflow-AGT-001-2024"`
	WorkflowStatus        string     `json:"workflow_status" example:"IN_PROGRESS"`
	StatusUpdated         bool       `json:"status_updated" example:"true"`
	PortalDisabled        bool       `json:"portal_disabled" example:"false"`
	CommissionStopped     bool       `json:"commission_stopped" example:"true"`
	LetterGenerated       bool       `json:"letter_generated" example:"false"`
	DataArchived          bool       `json:"data_archived" example:"false"`
	NotificationsSent     bool       `json:"notifications_sent" example:"false"`
	TerminationLetterURL  *string    `json:"termination_letter_url,omitempty" example:"https://storage.example.com/letters/term-AGT-001.pdf"`
	CreatedAt             time.Time  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty" example:"2024-01-15T11:00:00Z"`
}

// ReinstatementRequestDTO represents a reinstatement request
type ReinstatementRequestDTO struct {
	ReinstatementID         string     `json:"reinstatement_id" example:"660e8400-e29b-41d4-a716-446655440002"`
	AgentID                 string     `json:"agent_id" example:"AGT-001"`
	RequestDate             time.Time  `json:"request_date" example:"2024-02-01T09:00:00Z"`
	ReinstatementReason     string     `json:"reinstatement_reason" example:"Agent completed training"`
	RequestedBy             string     `json:"requested_by" example:"admin@company.com"`
	Status                  string     `json:"status" example:"PENDING"`
	ApprovedBy              *string    `json:"approved_by,omitempty" example:"manager@company.com"`
	ApprovedAt              *time.Time `json:"approved_at,omitempty" example:"2024-02-05T14:30:00Z"`
	RejectedBy              *string    `json:"rejected_by,omitempty"`
	RejectedAt              *time.Time `json:"rejected_at,omitempty"`
	RejectionReason         *string    `json:"rejection_reason,omitempty"`
	WorkflowID              *string    `json:"workflow_id,omitempty" example:"reinstatement-workflow-AGT-001-2024"`
	ReinstatementConditions *string    `json:"reinstatement_conditions,omitempty" example:"Complete quarterly training"`
	ProbationPeriodDays     *int       `json:"probation_period_days,omitempty" example:"90"`
	CreatedAt               time.Time  `json:"created_at" example:"2024-02-01T09:00:00Z"`
	UpdatedAt               *time.Time `json:"updated_at,omitempty" example:"2024-02-05T14:30:00Z"`
}

// DataArchiveDTO represents archived agent data
type DataArchiveDTO struct {
	ArchiveID        string     `json:"archive_id" example:"770e8400-e29b-41d4-a716-446655440003"`
	AgentID          string     `json:"agent_id" example:"AGT-001"`
	ArchiveDate      time.Time  `json:"archive_date" example:"2024-01-15T12:00:00Z"`
	ArchiveType      string     `json:"archive_type" example:"TERMINATION"`
	RetentionUntil   time.Time  `json:"retention_until" example:"2031-01-15T12:00:00Z"`
	DataChecksum     *string    `json:"data_checksum,omitempty" example:"sha256:abc123..."`
	StorageLocation  *string    `json:"storage_location,omitempty" example:"s3://archives/agents/AGT-001"`
	StorageSizeBytes *int64     `json:"storage_size_bytes,omitempty" example:"1048576"`
	ArchivedBy       string     `json:"archived_by" example:"system"`
	CreatedAt        time.Time  `json:"created_at" example:"2024-01-15T12:00:00Z"`
}

// TerminateAgentResponse represents response after terminating an agent
// AGT-039: Terminate Agent
type TerminateAgentResponse struct {
	port.StatusCodeAndMessage
	Message           string               `json:"message" example:"Agent termination initiated successfully"`
	TerminationRecord TerminationRecordDTO `json:"termination_record"`
	NextSteps         []string             `json:"next_steps" example:"Portal access will be disabled,Commission processing will stop,Termination letter will be generated"`
}

// GetTerminationLetterResponse represents response for termination letter
// AGT-040: Get Termination Letter
type GetTerminationLetterResponse struct {
	port.StatusCodeAndMessage
	Message              string               `json:"message" example:"Termination letter retrieved successfully"`
	TerminationRecord    TerminationRecordDTO `json:"termination_record"`
	LetterURL            string               `json:"letter_url" example:"https://storage.example.com/letters/term-AGT-001.pdf"`
	LetterGeneratedAt    time.Time            `json:"letter_generated_at" example:"2024-01-15T13:00:00Z"`
}

// ReinstateAgentResponse represents response after creating reinstatement request
// AGT-041: Reinstate Agent
type ReinstateAgentResponse struct {
	port.StatusCodeAndMessage
	Message              string                  `json:"message" example:"Reinstatement request created successfully"`
	ReinstatementRequest ReinstatementRequestDTO `json:"reinstatement_request"`
	NextSteps            []string                `json:"next_steps" example:"Request pending approval,Manager will review,Agent will be notified of decision"`
}

// ApproveReinstatementResponse represents response after approving reinstatement
type ApproveReinstatementResponse struct {
	port.StatusCodeAndMessage
	Message              string                  `json:"message" example:"Reinstatement approved successfully"`
	ReinstatementRequest ReinstatementRequestDTO `json:"reinstatement_request"`
	AgentStatus          string                  `json:"agent_status" example:"ACTIVE"`
}

// Helper functions to convert domain to DTO

func ToTerminationRecordDTO(record domain.AgentTerminationRecord) TerminationRecordDTO {
	dto := TerminationRecordDTO{
		TerminationID:         record.TerminationID,
		AgentID:               record.AgentID,
		TerminationDate:       record.TerminationDate,
		EffectiveDate:         record.EffectiveDate,
		TerminationReason:     record.TerminationReason,
		TerminationReasonCode: record.TerminationReasonCode,
		TerminatedBy:          record.TerminatedBy,
		WorkflowStatus:        record.WorkflowStatus,
		StatusUpdated:         record.StatusUpdated,
		PortalDisabled:        record.PortalDisabled,
		CommissionStopped:     record.CommissionStopped,
		LetterGenerated:       record.LetterGenerated,
		DataArchived:          record.DataArchived,
		NotificationsSent:     record.NotificationsSent,
		CreatedAt:             record.CreatedAt,
	}

	if record.WorkflowID.Valid {
		dto.WorkflowID = &record.WorkflowID.String
	}

	if record.TerminationLetterURL.Valid {
		dto.TerminationLetterURL = &record.TerminationLetterURL.String
	}

	if record.UpdatedAt.Valid {
		dto.UpdatedAt = &record.UpdatedAt.Time
	}

	return dto
}

func ToReinstatementRequestDTO(request domain.AgentReinstatementRequest) ReinstatementRequestDTO {
	dto := ReinstatementRequestDTO{
		ReinstatementID:     request.ReinstatementID,
		AgentID:             request.AgentID,
		RequestDate:         request.RequestDate,
		ReinstatementReason: request.ReinstatementReason,
		RequestedBy:         request.RequestedBy,
		Status:              request.Status,
		CreatedAt:           request.CreatedAt,
	}

	if request.ApprovedBy.Valid {
		dto.ApprovedBy = &request.ApprovedBy.String
	}

	if request.ApprovedAt.Valid {
		dto.ApprovedAt = &request.ApprovedAt.Time
	}

	if request.RejectedBy.Valid {
		dto.RejectedBy = &request.RejectedBy.String
	}

	if request.RejectedAt.Valid {
		dto.RejectedAt = &request.RejectedAt.Time
	}

	if request.RejectionReason.Valid {
		dto.RejectionReason = &request.RejectionReason.String
	}

	if request.WorkflowID.Valid {
		dto.WorkflowID = &request.WorkflowID.String
	}

	if request.ReinstatementConditions.Valid {
		dto.ReinstatementConditions = &request.ReinstatementConditions.String
	}

	if request.ProbationPeriodDays.Valid {
		days := int(request.ProbationPeriodDays.Int32)
		dto.ProbationPeriodDays = &days
	}

	if request.UpdatedAt.Valid {
		dto.UpdatedAt = &request.UpdatedAt.Time
	}

	return dto
}

func ToDataArchiveDTO(archive domain.AgentDataArchive) DataArchiveDTO {
	dto := DataArchiveDTO{
		ArchiveID:      archive.ArchiveID,
		AgentID:        archive.AgentID,
		ArchiveDate:    archive.ArchiveDate,
		ArchiveType:    archive.ArchiveType,
		RetentionUntil: archive.RetentionUntil,
		ArchivedBy:     archive.ArchivedBy,
		CreatedAt:      archive.CreatedAt,
	}

	if archive.DataChecksum.Valid {
		dto.DataChecksum = &archive.DataChecksum.String
	}

	if archive.StorageLocation.Valid {
		dto.StorageLocation = &archive.StorageLocation.String
	}

	if archive.StorageSizeBytes.Valid {
		dto.StorageSizeBytes = &archive.StorageSizeBytes.Int64
	}

	return dto
}

// Convert null string to pointer
func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// Convert null time to pointer
func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
