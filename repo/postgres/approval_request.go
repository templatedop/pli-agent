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

// ApprovalRequestRepository handles all database operations for approval requests
// E-09: Approval Request Entity
type ApprovalRequestRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewApprovalRequestRepository creates a new approval request repository
func NewApprovalRequestRepository(db *dblib.DB, cfg *config.Config) *ApprovalRequestRepository {
	return &ApprovalRequestRepository{
		db:  db,
		cfg: cfg,
	}
}

const approvalRequestTable = "approval_requests"

// Create inserts a new approval request
func (r *ApprovalRequestRepository) Create(ctx context.Context, request domain.ApprovalRequest) (*domain.ApprovalRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(approvalRequestTable).
		Columns(
			"agent_id", "section", "requested_changes", "requested_by",
			"requested_at", "status",
		).
		Values(
			request.AgentID, request.Section, request.RequestedChanges,
			request.RequestedBy, request.RequestedAt, request.Status,
		).
		Suffix("RETURNING *")

	var result domain.ApprovalRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.ApprovalRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	return &result, nil
}

// FindByID retrieves an approval request by ID
func (r *ApprovalRequestRepository) FindByID(ctx context.Context, approvalRequestID string) (*domain.ApprovalRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(approvalRequestTable).
		Where(sq.Eq{"approval_request_id": approvalRequestID})

	var result domain.ApprovalRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.ApprovalRequest], &result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("approval request not found: %s", approvalRequestID)
		}
		return nil, fmt.Errorf("failed to get approval request: %w", err)
	}

	return &result, nil
}

// ApproveReturning approves an approval request with UPDATE...RETURNING
// AGT-026: Approve Profile Update
// CRITICAL: Single atomic operation
func (r *ApprovalRequestRepository) ApproveReturning(
	ctx context.Context,
	approvalRequestID, reviewedBy string,
	comments *string,
) (*domain.ApprovalRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(approvalRequestTable).
		Set("status", domain.ApprovalStatusApproved).
		Set("reviewed_by", reviewedBy).
		Set("reviewed_at", time.Now()).
		Set("updated_at", time.Now()).
		Where(sq.And{
			sq.Eq{"approval_request_id": approvalRequestID},
			sq.Eq{"status": domain.ApprovalStatusPending},
		}).
		Suffix("RETURNING *")

	if comments != nil {
		query = query.Set("review_comments", *comments)
	}

	var result domain.ApprovalRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.ApprovalRequest], &result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("approval request not found or already processed: %s", approvalRequestID)
		}
		return nil, fmt.Errorf("failed to approve request: %w", err)
	}

	return &result, nil
}

// RejectReturning rejects an approval request with UPDATE...RETURNING
// AGT-027: Reject Profile Update
// CRITICAL: Single atomic operation
func (r *ApprovalRequestRepository) RejectReturning(
	ctx context.Context,
	approvalRequestID, rejectedBy, comments string,
) (*domain.ApprovalRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(approvalRequestTable).
		Set("status", domain.ApprovalStatusRejected).
		Set("reviewed_by", rejectedBy).
		Set("reviewed_at", time.Now()).
		Set("review_comments", comments).
		Set("updated_at", time.Now()).
		Where(sq.And{
			sq.Eq{"approval_request_id": approvalRequestID},
			sq.Eq{"status": domain.ApprovalStatusPending},
		}).
		Suffix("RETURNING *")

	var result domain.ApprovalRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.ApprovalRequest], &result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("approval request not found or already processed: %s", approvalRequestID)
		}
		return nil, fmt.Errorf("failed to reject request: %w", err)
	}

	return &result, nil
}
