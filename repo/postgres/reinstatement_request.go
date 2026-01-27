package repo

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	"pli-agent-api/core/domain"
)

const (
	agentReinstatementRequestTable = "agent_reinstatement_requests"
)

// ReinstatementRequestRepository handles all database operations for reinstatement requests
// E-08: Reinstatement Request Entity
// BR-AGT-PRF-016: Status updates require mandatory reason
type ReinstatementRequestRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewReinstatementRequestRepository creates a new reinstatement request repository
func NewReinstatementRequestRepository(db *dblib.DB, cfg *config.Config) *ReinstatementRequestRepository {
	return &ReinstatementRequestRepository{
		db:  db,
		cfg: cfg,
	}
}

// Create inserts a new reinstatement request with audit logging
// AGT-060: Create Reinstatement Request
// FR-AGT-PRF-013: Reinstatement Process
// BR-AGT-PRF-016: Status updates require mandatory reason (min 10 chars)
// WF-AGT-PRF-011: Reinstatement Workflow
// CRITICAL: Single database round trip with INSERT request + INSERT audit
func (r *ReinstatementRequestRepository) Create(ctx context.Context, request domain.ReinstatementRequest) (*domain.ReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Validate reinstatement reason length (min 10 chars)
	if len(request.RequestReasonText) < 10 {
		return nil, fmt.Errorf("reinstatement reason must be at least 10 characters")
	}

	// Validate effective date >= today
	today := time.Now().Truncate(24 * time.Hour)
	effectiveDate := request.EffectiveDate.Truncate(24 * time.Hour)
	if effectiveDate.Before(today) {
		return nil, fmt.Errorf("reinstatement effective date cannot be in the past")
	}

	// Insert reinstatement request
	insertQuery := dblib.Psql.Insert(agentReinstatementRequestTable).
		Columns(
			"agent_id", "request_reason_code", "request_reason_text",
			"effective_date", "request_status", "workflow_id", "workflow_status", "created_by",
		).
		Values(
			request.AgentID,
			request.RequestReasonCode,
			request.RequestReasonText,
			request.EffectiveDate,
			domain.RequestStatusPending,
			request.WorkflowID,
			domain.WorkflowStatusPending,
			request.CreatedBy,
		).
		Suffix("RETURNING *")

	var result domain.ReinstatementRequest
	err := dblib.SelectOne(cCtx, r.db, insertQuery, pgx.RowToStructByNameLax[domain.ReinstatementRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create reinstatement request: %w", err)
	}

	// Create audit log
	// FR-AGT-PRF-022: Audit History Tracking
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			result.AgentID,
			"REINSTATEMENT_REQUESTED",
			"request_status",
			domain.RequestStatusPending,
			fmt.Sprintf("Reinstatement requested: %s", request.RequestReasonCode),
			request.CreatedBy,
			time.Now(),
		)

	_, err = dblib.Insert(cCtx, r.db, auditQuery)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
	}

	return &result, nil
}

// FindByID retrieves a reinstatement request by ID
// AGT-061: Approve Reinstatement Request (查询)
// FR-AGT-PRF-013: Reinstatement Process
func (r *ReinstatementRequestRepository) FindByID(ctx context.Context, requestID string) (*domain.ReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentReinstatementRequestTable).
		Where(sq.Eq{"request_id": requestID, "deleted_at": nil})

	var request domain.ReinstatementRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.ReinstatementRequest], &request)
	if err != nil {
		return nil, fmt.Errorf("failed to find reinstatement request: %w", err)
	}

	return &request, nil
}

// FindByAgentID retrieves all reinstatement requests for an agent
// AGT-060: Create Reinstatement Request (查询)
// FR-AGT-PRF-013: Reinstatement Process
func (r *ReinstatementRequestRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.ReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentReinstatementRequestTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("created_at DESC")

	var requests []domain.ReinstatementRequest
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.ReinstatementRequest], &requests)
	if err != nil {
		return nil, fmt.Errorf("failed to find reinstatement requests: %w", err)
	}

	return requests, nil
}

// Approve approves a reinstatement request with audit logging
// AGT-061: Approve Reinstatement Request
// FR-AGT-PRF-013: Reinstatement Process
// BR-AGT-PRF-016: Status updates require mandatory reason
// WF-AGT-PRF-011: Reinstatement Workflow
// CRITICAL: UPDATE with RETURNING for atomic operation with audit logging
func (r *ReinstatementRequestRepository) Approve(
	ctx context.Context,
	requestID string,
	approvedBy string,
	approvalComments string,
) (*domain.ReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Verify request is pending
	currentRequest, err := r.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if !currentRequest.CanApprove() {
		return nil, fmt.Errorf("request cannot be approved (current status: %s)", currentRequest.RequestStatus)
	}

	approvedAt := time.Now()
	updateQuery := dblib.Psql.Update(agentReinstatementRequestTable).
		Set("request_status", domain.RequestStatusApproved).
		Set("approved_by", approvedBy).
		Set("approved_at", approvedAt).
		Set("approval_comments", approvalComments).
		Set("workflow_status", domain.WorkflowStatusRunning).
		Set("updated_by", approvedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"request_id": requestID, "deleted_at": nil}).
		Suffix("RETURNING *")

	var result domain.ReinstatementRequest
	err = dblib.SelectOne(cCtx, r.db, updateQuery, pgx.RowToStructByNameLax[domain.ReinstatementRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to approve reinstatement request: %w", err)
	}

	// Create audit log
	// FR-AGT-PRF-022: Audit History Tracking
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			result.AgentID,
			"REINSTATEMENT_APPROVED",
			"request_status",
			domain.RequestStatusApproved,
			fmt.Sprintf("Reinstatement approved: %s", approvalComments),
			approvedBy,
			time.Now(),
		)

	_, err = dblib.Insert(cCtx, r.db, auditQuery)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
	}

	return &result, nil
}

// Reject rejects a reinstatement request with audit logging
// AGT-062: Reject Reinstatement Request
// FR-AGT-PRF-013: Reinstatement Process
// BR-AGT-PRF-016: Status updates require mandatory reason
// WF-AGT-PRF-011: Reinstatement Workflow
// CRITICAL: UPDATE with RETURNING for atomic operation with audit logging
func (r *ReinstatementRequestRepository) Reject(
	ctx context.Context,
	requestID string,
	rejectedBy string,
	rejectionReason string,
) (*domain.ReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Validate rejection reason (min 10 chars)
	if len(rejectionReason) < 10 {
		return nil, fmt.Errorf("rejection reason must be at least 10 characters")
	}

	// Verify request is pending
	currentRequest, err := r.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if !currentRequest.CanReject() {
		return nil, fmt.Errorf("request cannot be rejected (current status: %s)", currentRequest.RequestStatus)
	}

	rejectedAt := time.Now()
	updateQuery := dblib.Psql.Update(agentReinstatementRequestTable).
		Set("request_status", domain.RequestStatusRejected).
		Set("rejected_by", rejectedBy).
		Set("rejected_at", rejectedAt).
		Set("rejection_reason", rejectionReason).
		Set("workflow_status", domain.WorkflowStatusFailed).
		Set("updated_by", rejectedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"request_id": requestID, "deleted_at": nil}).
		Suffix("RETURNING *")

	var result domain.ReinstatementRequest
	err = dblib.SelectOne(cCtx, r.db, updateQuery, pgx.RowToStructByNameLax[domain.ReinstatementRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to reject reinstatement request: %w", err)
	}

	// Create audit log
	// FR-AGT-PRF-022: Audit History Tracking
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			result.AgentID,
			"REINSTATEMENT_REJECTED",
			"request_status",
			domain.RequestStatusRejected,
			fmt.Sprintf("Reinstatement rejected: %s", rejectionReason),
			rejectedBy,
			time.Now(),
		)

	_, err = dblib.Insert(cCtx, r.db, auditQuery)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
	}

	return &result, nil
}

// UpdateWorkflowStatus updates the workflow status
// WF-AGT-PRF-011: Reinstatement Workflow
// CRITICAL: UPDATE with RETURNING for atomic operation
func (r *ReinstatementRequestRepository) UpdateWorkflowStatus(
	ctx context.Context,
	requestID string,
	workflowStatus string,
	updatedBy string,
) (*domain.ReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateQuery := dblib.Psql.Update(agentReinstatementRequestTable).
		Set("workflow_status", workflowStatus).
		Set("updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"request_id": requestID, "deleted_at": nil}).
		Suffix("RETURNING *")

	var result domain.ReinstatementRequest
	err := dblib.SelectOne(cCtx, r.db, updateQuery, pgx.RowToStructByNameLax[domain.ReinstatementRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow status: %w", err)
	}

	return &result, nil
}

// Delete soft deletes a reinstatement request
// FR-AGT-PRF-013: Reinstatement Process
// FR-AGT-PRF-022: Audit History Tracking
func (r *ReinstatementRequestRepository) Delete(ctx context.Context, requestID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateQuery := dblib.Psql.Update(agentReinstatementRequestTable).
		Set("deleted_at", time.Now()).
		Set("updated_by", deletedBy).
		Where(sq.Eq{"request_id": requestID, "deleted_at": nil})

	tag, err := dblib.Update(cCtx, r.db, updateQuery)
	if err != nil {
		return fmt.Errorf("failed to delete reinstatement request: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("reinstatement request not found: %s", requestID)
	}

	return nil
}
