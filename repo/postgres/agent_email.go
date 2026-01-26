package repo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
)

// AgentEmailRepository handles all database operations for agent emails
// E-04: Agent Email Entity
// BR-AGT-PRF-011: Email Address Management
type AgentEmailRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentEmailRepository creates a new agent email repository
func NewAgentEmailRepository(db *dblib.DB, cfg *config.Config) *AgentEmailRepository {
	return &AgentEmailRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentEmailTable = "agent_emails"

// Create inserts a new agent email
// FR-AGT-PRF-011: Email Management
// VR-AGT-PRF-014: Email Format Validation
func (r *AgentEmailRepository) Create(ctx context.Context, email domain.AgentEmail) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for email creation with audit log in single transaction
	// OPTIMIZATION: Batch operation combines INSERT email + INSERT audit log
	batch := &pgx.Batch{}

	// Query 1: Insert agent email
	// BR-AGT-PRF-011: Email Address Categories (OFFICIAL, PERMANENT, COMMUNICATION)
	query1 := dblib.Psql.Insert(agentEmailTable).
		Columns(
			"agent_id", "email_type", "email_address", "is_primary",
			"effective_from", "metadata", "created_by",
		).
		Values(
			email.AgentID, email.EmailType, email.EmailAddress, email.IsPrimary,
			email.EffectiveFrom, email.Metadata, email.CreatedBy,
		).
		Suffix("RETURNING email_id, created_at, version")

	var result domain.AgentEmail
	err := dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.AgentEmail], &result)
	if err != nil {
		return nil, err
	}

	// Query 2: Insert audit log for email creation
	// BR-AGT-PRF-005: Audit Logging
	query2 := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(email.AgentID, domain.AuditActionEmailUpdate, "email_type", email.EmailType, "New email added", email.CreatedBy, time.Now())

	err = dblib.QueueExecRow(batch, query2)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Copy input data to result
	result = email
	return &result, nil
}

// FindByID retrieves an email by ID
func (r *AgentEmailRepository) FindByID(ctx context.Context, emailID string) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"email_id": emailID, "deleted_at": nil})

	var email domain.AgentEmail
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// FindByAgentID retrieves all emails for an agent
// FR-AGT-PRF-011: Email Management
// BR-AGT-PRF-011: Email Address Categories
func (r *AgentEmailRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("is_primary DESC, effective_from DESC")

	var emails []domain.AgentEmail
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &emails)
	if err != nil {
		return nil, err
	}

	return emails, nil
}

// FindByAgentIDAndType retrieves a specific email type for an agent
// BR-AGT-PRF-011: Email Address Categories
func (r *AgentEmailRepository) FindByAgentIDAndType(ctx context.Context, agentID, emailType string) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"agent_id": agentID, "email_type": emailType, "deleted_at": nil}).
		OrderBy("is_primary DESC, effective_from DESC").
		Limit(1)

	var email domain.AgentEmail
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// FindPrimaryEmail retrieves the primary email for an agent
func (r *AgentEmailRepository) FindPrimaryEmail(ctx context.Context, agentID string) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"agent_id": agentID, "is_primary": true, "deleted_at": nil}).
		OrderBy("effective_from DESC").
		Limit(1)

	var email domain.AgentEmail
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// FindByEmailAddress retrieves an email by email address
// VR-AGT-PRF-014: Email Format Validation - Check uniqueness
func (r *AgentEmailRepository) FindByEmailAddress(ctx context.Context, emailAddress string) (*domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentEmailTable).
		Where(sq.Eq{"email_address": emailAddress, "deleted_at": nil}).
		Limit(1)

	var email domain.AgentEmail
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentEmail], &email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// Update updates an agent email
// FR-AGT-PRF-011: Email Management
// VR-AGT-PRF-014: Email Format Validation
func (r *AgentEmailRepository) Update(ctx context.Context, emailID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Update email
	updateQuery := dblib.Psql.Update(agentEmailTable).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"email_id": emailID, "deleted_at": nil})

	// Apply updates
	for field, value := range updates {
		updateQuery = updateQuery.Set(field, value)
	}

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentEmailTable).
		Where(sq.Eq{"email_id": emailID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit logs for each field update
	// BR-AGT-PRF-005: Audit Logging
	for field, newValue := range updates {
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "performed_by", "performed_at").
			Values(agentID, domain.AuditActionEmailUpdate, field, newValue, updatedBy, time.Now())

		err = dblib.QueueExecRow(batch, auditQuery)
		if err != nil {
			return err
		}
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes an agent email
func (r *AgentEmailRepository) Delete(ctx context.Context, emailID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for delete + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Soft delete email
	updateQuery := dblib.Psql.Update(agentEmailTable).
		Set("deleted_at", time.Now()).
		Set("updated_by", deletedBy).
		Where(sq.Eq{"email_id": emailID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentEmailTable).
		Where(sq.Eq{"email_id": emailID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionDelete, "email", "Email deleted", deletedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchCreate inserts multiple agent emails in a single transaction
// OPTIMIZATION: Batch operation for multiple email inserts
// FR-AGT-PRF-011: Email Management
func (r *AgentEmailRepository) BatchCreate(ctx context.Context, emails []domain.AgentEmail) ([]domain.AgentEmail, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch for multiple email inserts with audit logs
	// OPTIMIZATION: Batch operation combines multiple INSERTs in single round-trip
	batch := &pgx.Batch{}
	results := make([]domain.AgentEmail, len(emails))

	for i, email := range emails {
		// Insert email
		insertQuery := dblib.Psql.Insert(agentEmailTable).
			Columns(
				"agent_id", "email_type", "email_address", "is_primary",
				"effective_from", "metadata", "created_by",
			).
			Values(
				email.AgentID, email.EmailType, email.EmailAddress, email.IsPrimary,
				email.EffectiveFrom, email.Metadata, email.CreatedBy,
			).
			Suffix("RETURNING email_id, created_at, version")

		err := dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByNameLax[domain.AgentEmail], &results[i])
		if err != nil {
			return nil, err
		}

		// Insert audit log
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
			Values(email.AgentID, domain.AuditActionEmailUpdate, "email_type", email.EmailType, "New email added", email.CreatedBy, time.Now())

		err = dblib.QueueExecRow(batch, auditQuery)
		if err != nil {
			return nil, err
		}
	}

	// Execute batch
	err := r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Copy input data to results
	for i := range emails {
		results[i] = emails[i]
	}

	return results, nil
}

// SetPrimaryEmail sets an email as primary and unsets others
// OPTIMIZATION: Batch operation to update multiple emails atomically
func (r *AgentEmailRepository) SetPrimaryEmail(ctx context.Context, emailID, agentID, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch to unset all primary flags and set new primary
	// OPTIMIZATION: Batch combines multiple UPDATEs + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Unset all primary flags for agent
	unsetQuery := dblib.Psql.Update(agentEmailTable).
		Set("is_primary", false).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, unsetQuery)
	if err != nil {
		return err
	}

	// Query 2: Set new primary email
	setPrimaryQuery := dblib.Psql.Update(agentEmailTable).
		Set("is_primary", true).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"email_id": emailID, "deleted_at": nil})

	err = dblib.QueueExecRow(batch, setPrimaryQuery)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionEmailUpdate, "is_primary", "true", "Primary email changed", updatedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// ValidateEmailUniqueness checks if email address is unique (excluding current email)
// VR-AGT-PRF-014: Email Format Validation
func (r *AgentEmailRepository) ValidateEmailUniqueness(ctx context.Context, emailAddress, excludeEmailID string) (bool, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COUNT(*)").
		From(agentEmailTable).
		Where(sq.Eq{"email_address": emailAddress, "deleted_at": nil}).
		Where(sq.NotEq{"email_id": excludeEmailID})

	var count int64
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowTo[int64], &count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
