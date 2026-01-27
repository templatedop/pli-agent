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
	agentTerminationTable = "agent_termination_records"
)

// AgentTerminationRepository handles all database operations for agent terminations
// E-07: Agent Termination Entity
// BR-AGT-PRF-017: Agent Termination Workflow
type AgentTerminationRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentTerminationRepository creates a new agent termination repository
func NewAgentTerminationRepository(db *dblib.DB, cfg *config.Config) *AgentTerminationRepository {
	return &AgentTerminationRepository{
		db:  db,
		cfg: cfg,
	}
}

// Create inserts a new termination record with audit logging
// AGT-039: Terminate Agent
// FR-AGT-PRF-018: Agent Termination Workflow
// BR-AGT-PRF-017: Agent Termination Workflow
// VR-AGT-PRF-020: Termination Reason Mandatory (min 20 chars)
// VR-AGT-PRF-021: Termination Date Future or Today
// WF-AGT-PRF-004: Termination Workflow
// FR-AGT-PRF-022: Audit History Tracking
// CRITICAL: Single database round trip with batch for INSERT termination + INSERT audit
func (r *AgentTerminationRepository) Create(ctx context.Context, termination domain.AgentTermination) (*domain.AgentTermination, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Validate termination reason length (min 20 chars)
	if len(termination.TerminationReasonText) < 20 {
		return nil, fmt.Errorf("termination reason must be at least 20 characters")
	}

	// Validate effective date >= today
	today := time.Now().Truncate(24 * time.Hour)
	effectiveDate := termination.EffectiveDate.Truncate(24 * time.Hour)
	if effectiveDate.Before(today) {
		return nil, fmt.Errorf("termination effective date cannot be in the past")
	}

	// Use batch to combine INSERT termination + INSERT audit log in single round trip
	batch := &pgx.Batch{}

	// Query 1: Insert termination record
	insertQuery := dblib.Psql.Insert(agentTerminationTable).
		Columns(
			"agent_id", "termination_reason_code", "termination_reason_text",
			"effective_date", "workflow_id", "workflow_status", "created_by",
		).
		Values(
			termination.AgentID,
			termination.TerminationReasonCode,
			termination.TerminationReasonText,
			termination.EffectiveDate,
			termination.WorkflowID,
			domain.WorkflowStatusPending,
			termination.CreatedBy,
		).
		Suffix("RETURNING *")

	var result domain.AgentTermination
	err := dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByNameLax[domain.AgentTermination], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to queue termination insert: %w", err)
	}

	// Query 2: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			termination.AgentID,
			"AGENT_TERMINATED",
			"agent_status",
			"TERMINATED",
			fmt.Sprintf("Termination reason: %s", termination.TerminationReasonCode),
			termination.CreatedBy,
			time.Now(),
		)

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to queue audit log: %w", err)
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, fmt.Errorf("failed to create termination record: %w", err)
	}

	return &result, nil
}

// FindByID retrieves a termination record by ID
// AGT-039: Terminate Agent (查询)
// FR-AGT-PRF-018: Agent Termination Workflow
func (r *AgentTerminationRepository) FindByID(ctx context.Context, terminationID string) (*domain.AgentTermination, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentTerminationTable).
		Where(sq.Eq{"termination_id": terminationID, "deleted_at": nil})

	var termination domain.AgentTermination
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentTermination], &termination)
	if err != nil {
		return nil, fmt.Errorf("failed to find termination record: %w", err)
	}

	return &termination, nil
}

// FindByAgentID retrieves all termination records for an agent
// AGT-039: Terminate Agent (查询)
// FR-AGT-PRF-018: Agent Termination Workflow
func (r *AgentTerminationRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentTermination, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentTerminationTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("created_at DESC")

	var terminations []domain.AgentTermination
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentTermination], &terminations)
	if err != nil {
		return nil, fmt.Errorf("failed to find termination records: %w", err)
	}

	return terminations, nil
}

// UpdateWorkflowStatus updates the workflow status
// WF-AGT-PRF-004: Termination Workflow
// CRITICAL: UPDATE with RETURNING for atomic operation
func (r *AgentTerminationRepository) UpdateWorkflowStatus(
	ctx context.Context,
	terminationID string,
	workflowStatus string,
	updatedBy string,
) (*domain.AgentTermination, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateQuery := dblib.Psql.Update(agentTerminationTable).
		Set("workflow_status", workflowStatus).
		Set("updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"termination_id": terminationID, "deleted_at": nil}).
		Suffix("RETURNING *")

	var result domain.AgentTermination
	err := dblib.SelectOne(cCtx, r.db, updateQuery, pgx.RowToStructByNameLax[domain.AgentTermination], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow status: %w", err)
	}

	return &result, nil
}

// UpdateTerminationDetails updates termination details after workflow completion
// WF-AGT-PRF-004: Termination Workflow
// BR-AGT-PRF-017: Agent Termination Workflow
// FR-AGT-PRF-022: Audit History Tracking
// CRITICAL: Single database round trip with CTE for UPDATE + INSERT audit
func (r *AgentTerminationRepository) UpdateTerminationDetails(
	ctx context.Context,
	terminationID string,
	letterURL string,
	portalDisabledAt, commissionStoppedAt time.Time,
	archiveID string,
	updatedBy string,
) (*domain.AgentTermination, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit log in single round trip
	sql := `
		WITH updated AS (
			UPDATE agent_termination_records
			SET
				termination_letter_url = $2,
				portal_disabled_at = $3,
				commission_stopped_at = $4,
				archive_id = $5,
				workflow_status = 'COMPLETED',
				updated_by = $6,
				updated_at = NOW()
			WHERE termination_id = $1 AND deleted_at IS NULL
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT
			agent_id,
			'TERMINATION_COMPLETED',
			'workflow_status',
			'COMPLETED',
			'Termination workflow completed successfully',
			$6,
			NOW()
		FROM updated
		RETURNING (SELECT ROW(termination_id, agent_id, termination_reason_code, termination_reason_text, effective_date,
			termination_letter_url, portal_disabled_at, commission_stopped_at, archive_id, workflow_id, workflow_status,
			created_by, created_at, updated_by, updated_at, deleted_at, version) FROM updated)
	`

	var result domain.AgentTermination
	err := r.db.QueryRow(cCtx, sql, terminationID, letterURL, portalDisabledAt, commissionStoppedAt, archiveID, updatedBy).Scan(
		&result.TerminationID, &result.AgentID, &result.TerminationReasonCode, &result.TerminationReasonText,
		&result.EffectiveDate, &result.TerminationLetterURL, &result.PortalDisabledAt, &result.CommissionStoppedAt,
		&result.ArchiveID, &result.WorkflowID, &result.WorkflowStatus, &result.CreatedBy, &result.CreatedAt,
		&result.UpdatedBy, &result.UpdatedAt, &result.DeletedAt, &result.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update termination details: %w", err)
	}

	return &result, nil
}

// Delete soft deletes a termination record
// FR-AGT-PRF-018: Agent Termination Workflow
// FR-AGT-PRF-022: Audit History Tracking
func (r *AgentTerminationRepository) Delete(ctx context.Context, terminationID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateQuery := dblib.Psql.Update(agentTerminationTable).
		Set("deleted_at", time.Now()).
		Set("updated_by", deletedBy).
		Where(sq.Eq{"termination_id": terminationID, "deleted_at": nil})

	tag, err := dblib.Update(cCtx, r.db, updateQuery)
	if err != nil {
		return fmt.Errorf("failed to delete termination record: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("termination record not found: %s", terminationID)
	}

	return nil
}
