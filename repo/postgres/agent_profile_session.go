package repo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
	dbutil "pli-agent-api/db"
)

// AgentProfileSessionRepository handles all database operations for profile creation sessions
// WF-AGT-PRF-001: Profile Creation Workflow
type AgentProfileSessionRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentProfileSessionRepository creates a new session repository
func NewAgentProfileSessionRepository(db *dblib.DB, cfg *config.Config) *AgentProfileSessionRepository {
	return &AgentProfileSessionRepository{
		db:  db,
		cfg: cfg,
	}
}

const sessionTable = "agent_profile_sessions"

// Create creates a new profile creation session
// AGT-001: Initiate Agent Profile Creation
// WF-AGT-PRF-001: Profile Creation Workflow
func (r *AgentProfileSessionRepository) Create(ctx context.Context, session domain.AgentProfileSession) (*domain.AgentProfileSession, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(sessionTable).
		Columns("agent_type", "workflow_state", "current_step", "next_step", "progress_percentage", "status", "initiated_by", "expires_at").
		Values(session.AgentType, session.WorkflowState, session.CurrentStep, session.NextStep, session.ProgressPercentage, session.Status, session.InitiatedBy, session.ExpiresAt).
		Suffix("RETURNING *")

	var result domain.AgentProfileSession
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileSession], &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FindByID retrieves a session by ID
func (r *AgentProfileSessionRepository) FindByID(ctx context.Context, sessionID string) (*domain.AgentProfileSession, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(sessionTable).
		Where(sq.Eq{"session_id": sessionID})

	var session domain.AgentProfileSession
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileSession], &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateWorkflowState updates session workflow state
// AGT-016 to AGT-019: Session Management
func (r *AgentProfileSessionRepository) UpdateWorkflowState(ctx context.Context, sessionID, workflowState, currentStep, nextStep string, progressPercentage int, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(sessionTable).
		Set("workflow_state", workflowState).
		Set("current_step", currentStep).
		Set("next_step", nextStep).
		Set("progress_percentage", progressPercentage).
		Set("last_updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID})

	_, err := dblib.Exec(cCtx, r.db, query)
	return err
}

// SaveFormData saves session form data
// AGT-017: Save Session Checkpoint
func (r *AgentProfileSessionRepository) SaveFormData(ctx context.Context, sessionID string, formDataJSON string, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(sessionTable).
		Set("form_data", formDataJSON).
		Set("last_updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID})

	_, err := dblib.Exec(cCtx, r.db, query)
	return err
}

// UpdateStatus updates session status
// AGT-019: Cancel Session
func (r *AgentProfileSessionRepository) UpdateStatus(ctx context.Context, sessionID, status string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(sessionTable).
		Set("status", status).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID})

	if status == domain.SessionStatusCompleted {
		query = query.Set("completed_at", time.Now())
	}

	_, err := dblib.Exec(cCtx, r.db, query)
	return err
}

// LinkTemporalWorkflow links session to Temporal workflow
func (r *AgentProfileSessionRepository) LinkTemporalWorkflow(ctx context.Context, sessionID, workflowID, runID string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(sessionTable).
		Set("temporal_workflow_id", workflowID).
		Set("temporal_run_id", runID).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID})

	_, err := dblib.Exec(cCtx, r.db, query)
	return err
}

// Complete marks session as completed with created agent ID
// AGT-006: Submit Profile for Creation
func (r *AgentProfileSessionRepository) Complete(ctx context.Context, sessionID, agentID string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(sessionTable).
		Set("status", domain.SessionStatusCompleted).
		Set("workflow_state", domain.WorkflowStateCompleted).
		Set("completed_at", time.Now()).
		Set("created_agent_id", agentID).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID})

	_, err := dblib.Exec(cCtx, r.db, query)
	return err
}

// Cancel cancels a session
// AGT-019: Cancel Session
func (r *AgentProfileSessionRepository) Cancel(ctx context.Context, sessionID, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(sessionTable).
		Set("status", domain.SessionStatusCancelled).
		Set("workflow_state", domain.WorkflowStateCancelled).
		Set("last_updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID})

	_, err := dblib.Exec(cCtx, r.db, query)
	return err
}

// MarkExpiredSessions marks all expired sessions
// Batch operation for cleanup
func (r *AgentProfileSessionRepository) MarkExpiredSessions(ctx context.Context) (int64, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use CTE to update expired sessions and return count
	sql := `
		WITH expired_sessions AS (
			UPDATE agent_profile_sessions
			SET status = $1, workflow_state = $2, updated_at = NOW()
			WHERE status = $3 AND expires_at < NOW()
			RETURNING session_id
		)
		SELECT COUNT(*) FROM expired_sessions
	`

	args := []interface{}{
		domain.SessionStatusExpired,
		domain.WorkflowStateExpired,
		domain.SessionStatusActive,
	}

	batch := &pgx.Batch{}
	var count int64
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowTo[int64], &count)
	if err != nil {
		return 0, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ========================================================================
// ATOMIC BATCH OPERATIONS (Single Round-Trip to Database)
// ========================================================================

// SaveFormDataAndUpdateWorkflowStateReturning atomically saves form data and updates workflow state
// Returns the updated session in a single database round trip
// AGT-002, AGT-004, AGT-005: Profile creation workflow handlers
func (r *AgentProfileSessionRepository) SaveFormDataAndUpdateWorkflowStateReturning(
	ctx context.Context,
	sessionID string,
	formDataJSON string,
	workflowState string,
	currentStep string,
	nextStep string,
	progressPercentage int,
	updatedBy string,
) (*domain.AgentProfileSession, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Single UPDATE with RETURNING - atomic operation
	query := dblib.Psql.Update(sessionTable).
		Set("form_data", formDataJSON).
		Set("workflow_state", workflowState).
		Set("current_step", currentStep).
		Set("next_step", nextStep).
		Set("progress_percentage", progressPercentage).
		Set("last_updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID}).
		Suffix("RETURNING *")

	var result domain.AgentProfileSession
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileSession], &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// LinkTemporalWorkflowAndUpdateStateReturning atomically links Temporal workflow and updates state
// Returns the updated session in a single database round trip
// AGT-006: Submit Profile for Creation
// Ensures atomicity - either both workflow link and state update succeed, or both fail
func (r *AgentProfileSessionRepository) LinkTemporalWorkflowAndUpdateStateReturning(
	ctx context.Context,
	sessionID string,
	workflowID string,
	runID string,
	workflowState string,
	currentStep string,
	nextStep string,
	progressPercentage int,
	updatedBy string,
) (*domain.AgentProfileSession, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Single UPDATE with RETURNING - atomic operation
	// Prevents inconsistent state where workflow is linked but state is not updated
	query := dblib.Psql.Update(sessionTable).
		Set("temporal_workflow_id", workflowID).
		Set("temporal_run_id", runID).
		Set("workflow_state", workflowState).
		Set("current_step", currentStep).
		Set("next_step", nextStep).
		Set("progress_percentage", progressPercentage).
		Set("last_updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID}).
		Suffix("RETURNING *")

	var result domain.AgentProfileSession
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileSession], &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CancelReturning atomically cancels session and returns result
// Single database round trip using UPDATE...RETURNING
// AGT-019: Cancel Session
func (r *AgentProfileSessionRepository) CancelReturning(ctx context.Context, sessionID string, updatedBy string) (*domain.AgentProfileSession, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(sessionTable).
		Set("status", domain.SessionStatusCancelled).
		Set("workflow_state", domain.WorkflowStateCancelled).
		Set("last_updated_by", updatedBy).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID}).
		Suffix("RETURNING *")

	var result domain.AgentProfileSession
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfileSession], &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
