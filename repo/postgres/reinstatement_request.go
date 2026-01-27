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
// FR-AGT-PRF-022: Audit History Tracking
// CRITICAL: Single database round trip with batch for INSERT request + INSERT audit
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

	// Use batch to combine INSERT request + INSERT audit log in single round trip
	batch := &pgx.Batch{}

	// Query 1: Insert reinstatement request
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
	err := dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByNameLax[domain.ReinstatementRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to queue reinstatement request insert: %w", err)
	}

	// Query 2: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			request.AgentID,
			"REINSTATEMENT_REQUESTED",
			"request_status",
			domain.RequestStatusPending,
			fmt.Sprintf("Reinstatement requested: %s", request.RequestReasonCode),
			request.CreatedBy,
			time.Now(),
		)

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to queue audit log: %w", err)
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, fmt.Errorf("failed to create reinstatement request: %w", err)
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
// FR-AGT-PRF-022: Audit History Tracking
// CRITICAL: Single database round trip with CTE for validation + UPDATE + INSERT audit
func (r *ReinstatementRequestRepository) Approve(
	ctx context.Context,
	requestID string,
	approvedBy string,
	approvalComments string,
) (*domain.ReinstatementRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	approvedAt := time.Now()

	// Use CTE to validate + UPDATE + INSERT audit log in single round trip
	// Validates request is PENDING and not deleted before updating
	sql := `
		WITH current_request AS (
			SELECT request_status, deleted_at
			FROM agent_reinstatement_requests
			WHERE request_id = $1
		),
		updated AS (
			UPDATE agent_reinstatement_requests
			SET
				request_status = 'APPROVED',
				approved_by = $2,
				approved_at = $3,
				approval_comments = $4,
				workflow_status = 'RUNNING',
				updated_by = $2,
				updated_at = NOW()
			WHERE request_id = $1
				AND request_status = 'PENDING'
				AND deleted_at IS NULL
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT
			agent_id,
			'REINSTATEMENT_APPROVED',
			'request_status',
			'APPROVED',
			'Reinstatement approved: ' || $4,
			$2,
			NOW()
		FROM updated
		RETURNING (SELECT ROW(request_id, agent_id, request_reason_code, request_reason_text, effective_date,
			request_status, approved_by, approved_at, approval_comments, rejected_by, rejected_at, rejection_reason,
			workflow_id, workflow_status, created_by, created_at, updated_by, updated_at, deleted_at, version) FROM updated)
	`

	var result domain.ReinstatementRequest
	err := r.db.QueryRow(cCtx, sql, requestID, approvedBy, approvedAt, approvalComments).Scan(
		&result.RequestID, &result.AgentID, &result.RequestReasonCode, &result.RequestReasonText,
		&result.EffectiveDate, &result.RequestStatus, &result.ApprovedBy, &result.ApprovedAt,
		&result.ApprovalComments, &result.RejectedBy, &result.RejectedAt, &result.RejectionReason,
		&result.WorkflowID, &result.WorkflowStatus, &result.CreatedBy, &result.CreatedAt,
		&result.UpdatedBy, &result.UpdatedAt, &result.DeletedAt, &result.Version,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("request cannot be approved (not found or not in PENDING status)")
		}
		return nil, fmt.Errorf("failed to approve reinstatement request: %w", err)
	}

	return &result, nil
}

// Reject rejects a reinstatement request with audit logging
// AGT-062: Reject Reinstatement Request
// FR-AGT-PRF-013: Reinstatement Process
// BR-AGT-PRF-016: Status updates require mandatory reason
// WF-AGT-PRF-011: Reinstatement Workflow
// FR-AGT-PRF-022: Audit History Tracking
// CRITICAL: Single database round trip with CTE for validation + UPDATE + INSERT audit
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

	rejectedAt := time.Now()

	// Use CTE to validate + UPDATE + INSERT audit log in single round trip
	// Validates request is PENDING and not deleted before updating
	sql := `
		WITH updated AS (
			UPDATE agent_reinstatement_requests
			SET
				request_status = 'REJECTED',
				rejected_by = $2,
				rejected_at = $3,
				rejection_reason = $4,
				workflow_status = 'FAILED',
				updated_by = $2,
				updated_at = NOW()
			WHERE request_id = $1
				AND request_status = 'PENDING'
				AND deleted_at IS NULL
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT
			agent_id,
			'REINSTATEMENT_REJECTED',
			'request_status',
			'REJECTED',
			'Reinstatement rejected: ' || $4,
			$2,
			NOW()
		FROM updated
		RETURNING (SELECT ROW(request_id, agent_id, request_reason_code, request_reason_text, effective_date,
			request_status, approved_by, approved_at, approval_comments, rejected_by, rejected_at, rejection_reason,
			workflow_id, workflow_status, created_by, created_at, updated_by, updated_at, deleted_at, version) FROM updated)
	`

	var result domain.ReinstatementRequest
	err := r.db.QueryRow(cCtx, sql, requestID, rejectedBy, rejectedAt, rejectionReason).Scan(
		&result.RequestID, &result.AgentID, &result.RequestReasonCode, &result.RequestReasonText,
		&result.EffectiveDate, &result.RequestStatus, &result.ApprovedBy, &result.ApprovedAt,
		&result.ApprovalComments, &result.RejectedBy, &result.RejectedAt, &result.RejectionReason,
		&result.WorkflowID, &result.WorkflowStatus, &result.CreatedBy, &result.CreatedAt,
		&result.UpdatedBy, &result.UpdatedAt, &result.DeletedAt, &result.Version,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("request cannot be rejected (not found or not in PENDING status)")
		}
		return nil, fmt.Errorf("failed to reject reinstatement request: %w", err)
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
