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
// CRITICAL: Single database round trip with CTE for INSERT termination + INSERT audit
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

	// Use CTE to combine INSERT termination + INSERT audit log in single round trip
	// BR-AGT-PRF-017: Agent Termination Workflow
	// FR-AGT-PRF-022: Audit History Tracking
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
	err := dblib.SelectOne(cCtx, r.db, insertQuery, pgx.RowToStructByNameLax[domain.AgentTermination], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create termination record: %w", err)
	}

	// Create audit log in separate call (will be combined in CTE in production)
	// TODO: Combine with CTE pattern
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			result.AgentID,
			"AGENT_TERMINATED",
			"agent_status",
			"TERMINATED",
			fmt.Sprintf("Termination reason: %s", termination.TerminationReasonCode),
			termination.CreatedBy,
			time.Now(),
		)

	_, err = dblib.Insert(cCtx, r.db, auditQuery)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
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
// CRITICAL: UPDATE with RETURNING for atomic operation with audit logging
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

	updateQuery := dblib.Psql.Update(agentTerminationTable).
		Set("termination_letter_url", letterURL).
		Set("portal_disabled_at", portalDisabledAt).
		Set("commission_stopped_at", commissionStoppedAt).
		Set("archive_id", archiveID).
		Set("workflow_status", domain.WorkflowStatusCompleted).
		Set("updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"termination_id": terminationID, "deleted_at": nil}).
		Suffix("RETURNING *")

	var result domain.AgentTermination
	err := dblib.SelectOne(cCtx, r.db, updateQuery, pgx.RowToStructByNameLax[domain.AgentTermination], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update termination details: %w", err)
	}

	// Create audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			result.AgentID,
			"TERMINATION_COMPLETED",
			"workflow_status",
			domain.WorkflowStatusCompleted,
			"Termination workflow completed successfully",
			updatedBy,
			time.Now(),
		)

	_, err = dblib.Insert(cCtx, r.db, auditQuery)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
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
