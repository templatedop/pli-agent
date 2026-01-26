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

// AgentContactRepository handles all database operations for agent contacts
// E-03: Agent Contact Entity
// BR-AGT-PRF-010: Phone Number Categories
type AgentContactRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentContactRepository creates a new agent contact repository
func NewAgentContactRepository(db *dblib.DB, cfg *config.Config) *AgentContactRepository {
	return &AgentContactRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentContactTable = "agent_contacts"

// Create inserts a new agent contact
// FR-AGT-PRF-010: Contact Management
// VR-AGT-PRF-013: Mobile Format Validation (10 digits)
func (r *AgentContactRepository) Create(ctx context.Context, contact domain.AgentContact) (*domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for contact creation with audit log in single transaction
	// OPTIMIZATION: Batch operation combines INSERT contact + INSERT audit log
	batch := &pgx.Batch{}

	// Query 1: Insert agent contact
	// BR-AGT-PRF-010: Phone Number Categories (MOBILE, OFFICIAL_LANDLINE, RESIDENT_LANDLINE)
	query1 := dblib.Psql.Insert(agentContactTable).
		Columns(
			"agent_id", "contact_type", "contact_number", "is_primary",
			"effective_from", "metadata", "created_by",
		).
		Values(
			contact.AgentID, contact.ContactType, contact.ContactNumber, contact.IsPrimary,
			contact.EffectiveFrom, contact.Metadata, contact.CreatedBy,
		).
		Suffix("RETURNING contact_id, created_at, version")

	var result domain.AgentContact
	err := dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.AgentContact], &result)
	if err != nil {
		return nil, err
	}

	// Query 2: Insert audit log for contact creation
	// BR-AGT-PRF-005: Audit Logging
	query2 := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(contact.AgentID, domain.AuditActionContactUpdate, "contact_type", contact.ContactType, "New contact added", contact.CreatedBy, time.Now())

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
	result = contact
	return &result, nil
}

// FindByID retrieves a contact by ID
func (r *AgentContactRepository) FindByID(ctx context.Context, contactID string) (*domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentContactTable).
		Where(sq.Eq{"contact_id": contactID, "deleted_at": nil})

	var contact domain.AgentContact
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentContact], &contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// FindByAgentID retrieves all contacts for an agent
// FR-AGT-PRF-010: Contact Management
// BR-AGT-PRF-010: Phone Number Categories
func (r *AgentContactRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentContactTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("is_primary DESC, effective_from DESC")

	var contacts []domain.AgentContact
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentContact], &contacts)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// FindByAgentIDAndType retrieves a specific contact type for an agent
// BR-AGT-PRF-010: Phone Number Categories
func (r *AgentContactRepository) FindByAgentIDAndType(ctx context.Context, agentID, contactType string) (*domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentContactTable).
		Where(sq.Eq{"agent_id": agentID, "contact_type": contactType, "deleted_at": nil}).
		OrderBy("is_primary DESC, effective_from DESC").
		Limit(1)

	var contact domain.AgentContact
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentContact], &contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// FindPrimaryContact retrieves the primary contact for an agent
func (r *AgentContactRepository) FindPrimaryContact(ctx context.Context, agentID string) (*domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentContactTable).
		Where(sq.Eq{"agent_id": agentID, "is_primary": true, "deleted_at": nil}).
		OrderBy("effective_from DESC").
		Limit(1)

	var contact domain.AgentContact
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentContact], &contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// Update updates an agent contact
// FR-AGT-PRF-010: Contact Management
// VR-AGT-PRF-013: Mobile Format Validation
func (r *AgentContactRepository) Update(ctx context.Context, contactID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Update contact
	updateQuery := dblib.Psql.Update(agentContactTable).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"contact_id": contactID, "deleted_at": nil})

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
		From(agentContactTable).
		Where(sq.Eq{"contact_id": contactID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit logs for each field update
	// BR-AGT-PRF-005: Audit Logging
	for field, newValue := range updates {
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "performed_by", "performed_at").
			Values(agentID, domain.AuditActionContactUpdate, field, newValue, updatedBy, time.Now())

		err = dblib.QueueExecRow(batch, auditQuery)
		if err != nil {
			return err
		}
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes an agent contact
func (r *AgentContactRepository) Delete(ctx context.Context, contactID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for delete + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Soft delete contact
	updateQuery := dblib.Psql.Update(agentContactTable).
		Set("deleted_at", time.Now()).
		Set("updated_by", deletedBy).
		Where(sq.Eq{"contact_id": contactID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentContactTable).
		Where(sq.Eq{"contact_id": contactID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionDelete, "contact", "Contact deleted", deletedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchCreate inserts multiple agent contacts in a single transaction
// OPTIMIZATION: Batch operation for multiple contact inserts
// FR-AGT-PRF-010: Contact Management
func (r *AgentContactRepository) BatchCreate(ctx context.Context, contacts []domain.AgentContact) ([]domain.AgentContact, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch for multiple contact inserts with audit logs
	// OPTIMIZATION: Batch operation combines multiple INSERTs in single round-trip
	batch := &pgx.Batch{}
	results := make([]domain.AgentContact, len(contacts))

	for i, contact := range contacts {
		// Insert contact
		insertQuery := dblib.Psql.Insert(agentContactTable).
			Columns(
				"agent_id", "contact_type", "contact_number", "is_primary",
				"effective_from", "metadata", "created_by",
			).
			Values(
				contact.AgentID, contact.ContactType, contact.ContactNumber, contact.IsPrimary,
				contact.EffectiveFrom, contact.Metadata, contact.CreatedBy,
			).
			Suffix("RETURNING contact_id, created_at, version")

		err := dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByNameLax[domain.AgentContact], &results[i])
		if err != nil {
			return nil, err
		}

		// Insert audit log
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
			Values(contact.AgentID, domain.AuditActionContactUpdate, "contact_type", contact.ContactType, "New contact added", contact.CreatedBy, time.Now())

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
	for i := range contacts {
		results[i] = contacts[i]
	}

	return results, nil
}

// SetPrimaryContact sets a contact as primary and unsets others
// OPTIMIZATION: Batch operation to update multiple contacts atomically
func (r *AgentContactRepository) SetPrimaryContact(ctx context.Context, contactID, agentID, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch to unset all primary flags and set new primary
	// OPTIMIZATION: Batch combines multiple UPDATEs + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Unset all primary flags for agent
	unsetQuery := dblib.Psql.Update(agentContactTable).
		Set("is_primary", false).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, unsetQuery)
	if err != nil {
		return err
	}

	// Query 2: Set new primary contact
	setPrimaryQuery := dblib.Psql.Update(agentContactTable).
		Set("is_primary", true).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"contact_id": contactID, "deleted_at": nil})

	err = dblib.QueueExecRow(batch, setPrimaryQuery)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionContactUpdate, "is_primary", "true", "Primary contact changed", updatedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}
