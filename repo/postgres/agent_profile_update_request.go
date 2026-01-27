package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
)

const agentProfileUpdateRequestTable = "agent_profile_update_requests"

// AgentProfileUpdateRequestRepository handles profile update request operations
type AgentProfileUpdateRequestRepository struct {
	db  dblib.DB
	cfg *config.Config
}

// NewAgentProfileUpdateRequestRepository creates a new update request repository
func NewAgentProfileUpdateRequestRepository(db dblib.DB, cfg *config.Config) *AgentProfileUpdateRequestRepository {
	return &AgentProfileUpdateRequestRepository{
		db:  db,
		cfg: cfg,
	}
}

// Create creates a new profile update request
// AGT-025: Update Profile Section (when critical fields require approval)
func (r *AgentProfileUpdateRequestRepository) Create(
	ctx context.Context,
	agentID string,
	section string,
	fieldUpdates map[string]interface{},
	reason string,
	requestedBy string,
) (*domain.AgentProfileUpdateRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Convert field updates to JSON
	fieldUpdatesJSON, err := json.Marshal(fieldUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal field updates: %w", err)
	}

	requestID := uuid.New().String()
	now := time.Now()

	query := dblib.Psql.Insert(agentProfileUpdateRequestTable).
		Columns(
			"request_id", "agent_id", "section", "field_updates", "reason",
			"requested_by", "requested_at", "status", "created_at", "updated_at",
		).
		Values(
			requestID, agentID, section, fieldUpdatesJSON, reason,
			requestedBy, now, domain.UpdateRequestStatusPending, now, now,
		).
		Suffix("RETURNING *")

	var result domain.AgentProfileUpdateRequest
	err = dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileUpdateRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create update request: %w", err)
	}

	return &result, nil
}

// FindByID retrieves an update request by ID
// AGT-026, AGT-027: Approve/Reject Profile Update
func (r *AgentProfileUpdateRequestRepository) FindByID(ctx context.Context, requestID string) (*domain.AgentProfileUpdateRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileUpdateRequestTable).
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"deleted_at": nil},
		})

	var result domain.AgentProfileUpdateRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileUpdateRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to find update request: %w", err)
	}

	return &result, nil
}

// Approve approves an update request
// AGT-026: Approve Profile Update
func (r *AgentProfileUpdateRequestRepository) Approve(
	ctx context.Context,
	requestID string,
	approvedBy string,
	comments string,
) (*domain.AgentProfileUpdateRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	query := dblib.Psql.Update(agentProfileUpdateRequestTable).
		Set("status", domain.UpdateRequestStatusApproved).
		Set("approved_by", approvedBy).
		Set("approved_at", now).
		Set("comments", comments).
		Set("updated_at", now).
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"status": domain.UpdateRequestStatusPending},
			sq.Eq{"deleted_at": nil},
		}).
		Suffix("RETURNING *")

	var result domain.AgentProfileUpdateRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileUpdateRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to approve update request: %w", err)
	}

	return &result, nil
}

// Reject rejects an update request
// AGT-027: Reject Profile Update
func (r *AgentProfileUpdateRequestRepository) Reject(
	ctx context.Context,
	requestID string,
	rejectedBy string,
	comments string,
) (*domain.AgentProfileUpdateRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	query := dblib.Psql.Update(agentProfileUpdateRequestTable).
		Set("status", domain.UpdateRequestStatusRejected).
		Set("rejected_by", rejectedBy).
		Set("rejected_at", now).
		Set("comments", comments).
		Set("updated_at", now).
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"status": domain.UpdateRequestStatusPending},
			sq.Eq{"deleted_at": nil},
		}).
		Suffix("RETURNING *")

	var result domain.AgentProfileUpdateRequest
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileUpdateRequest], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to reject update request: %w", err)
	}

	return &result, nil
}

// ListByAgentID retrieves all update requests for an agent
// For viewing pending/historical update requests
func (r *AgentProfileUpdateRequestRepository) ListByAgentID(
	ctx context.Context,
	agentID string,
	status *string,
) ([]domain.AgentProfileUpdateRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileUpdateRequestTable).
		Where(sq.And{
			sq.Eq{"agent_id": agentID},
			sq.Eq{"deleted_at": nil},
		}).
		OrderBy("created_at DESC")

	// Optional status filter
	if status != nil && *status != "" {
		query = query.Where(sq.Eq{"status": *status})
	}

	var results []domain.AgentProfileUpdateRequest
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileUpdateRequest], &results)
	if err != nil {
		return nil, fmt.Errorf("failed to list update requests: %w", err)
	}

	return results, nil
}

// ApproveAndApplyUpdates approves request and returns both request and field updates in SINGLE database call
// AGT-026: Approve Profile Update - Optimized version
// Returns: approved request and field updates to apply
func (r *AgentProfileUpdateRequestRepository) ApproveAndApplyUpdates(
	ctx context.Context,
	requestID string,
	approvedBy string,
	comments string,
) (*domain.AgentProfileUpdateRequest, map[string]interface{}, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	// Use CTE to fetch request + mark as approved in single query
	sql := `
		WITH request_data AS (
			SELECT * FROM agent_profile_update_requests
			WHERE request_id = $1 AND status = $2 AND deleted_at IS NULL
		),
		approved_request AS (
			UPDATE agent_profile_update_requests
			SET
				status = $3,
				approved_by = $4,
				approved_at = $5,
				comments = $6,
				updated_at = $5
			WHERE request_id = $1 AND status = $2 AND deleted_at IS NULL
			RETURNING *
		)
		SELECT
			r.request_id, r.agent_id, r.section, r.field_updates, r.reason,
			r.requested_by, r.requested_at, r.status, r.approved_by, r.approved_at,
			r.rejected_by, r.rejected_at, r.comments, r.created_at, r.updated_at, r.deleted_at
		FROM approved_request r
	`

	var result domain.AgentProfileUpdateRequest
	err := r.db.QueryRow(cCtx, sql,
		requestID,
		domain.UpdateRequestStatusPending,
		domain.UpdateRequestStatusApproved,
		approvedBy,
		now,
		comments,
	).Scan(
		&result.RequestID, &result.AgentID, &result.Section, &result.FieldUpdates,
		&result.Reason, &result.RequestedBy, &result.RequestedAt, &result.Status,
		&result.ApprovedBy, &result.ApprovedAt, &result.RejectedBy, &result.RejectedAt,
		&result.Comments, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to approve update request: %w", err)
	}

	// Parse field updates from JSON
	var fieldUpdates map[string]interface{}
	if result.FieldUpdates.Valid {
		err = json.Unmarshal([]byte(result.FieldUpdates.String), &fieldUpdates)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse field updates: %w", err)
		}
	}

	return &result, fieldUpdates, nil
}

// RejectAndReturn rejects request and returns it in SINGLE database call
// AGT-027: Reject Profile Update - Optimized version
func (r *AgentProfileUpdateRequestRepository) RejectAndReturn(
	ctx context.Context,
	requestID string,
	rejectedBy string,
	comments string,
) (*domain.AgentProfileUpdateRequest, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	// Use CTE to fetch + update in single query
	sql := `
		WITH request_data AS (
			SELECT * FROM agent_profile_update_requests
			WHERE request_id = $1 AND status = $2 AND deleted_at IS NULL
		)
		UPDATE agent_profile_update_requests
		SET
			status = $3,
			rejected_by = $4,
			rejected_at = $5,
			comments = $6,
			updated_at = $5
		WHERE request_id = $1 AND status = $2 AND deleted_at IS NULL
		RETURNING *
	`

	var result domain.AgentProfileUpdateRequest
	err := r.db.QueryRow(cCtx, sql,
		requestID,
		domain.UpdateRequestStatusPending,
		domain.UpdateRequestStatusRejected,
		rejectedBy,
		now,
		comments,
	).Scan(
		&result.RequestID, &result.AgentID, &result.Section, &result.FieldUpdates,
		&result.Reason, &result.RequestedBy, &result.RequestedAt, &result.Status,
		&result.ApprovedBy, &result.ApprovedAt, &result.RejectedBy, &result.RejectedAt,
		&result.Comments, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reject update request: %w", err)
	}

	return &result, nil
}
