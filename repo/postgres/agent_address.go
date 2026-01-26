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

// AgentAddressRepository handles all database operations for agent addresses
// E-02: Agent Address Entity
// BR-AGT-PRF-008: Multiple Address Types Support
// BR-AGT-PRF-009: Communication Address Same as Permanent Option
type AgentAddressRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentAddressRepository creates a new agent address repository
func NewAgentAddressRepository(db *dblib.DB, cfg *config.Config) *AgentAddressRepository {
	return &AgentAddressRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentAddressTable = "agent_addresses"

// Create inserts a new agent address
// FR-AGT-PRF-009: Address Management
// VR-AGT-PRF-008: Pincode Validation (6 digits)
// VR-AGT-PRF-009: Address Line1 Mandatory
// VR-AGT-PRF-010: City Mandatory
// VR-AGT-PRF-011: State Mandatory
// VR-AGT-PRF-012: Country Mandatory
func (r *AgentAddressRepository) Create(ctx context.Context, address domain.AgentAddress) (*domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for address creation with audit log in single transaction
	// OPTIMIZATION: Batch operation combines INSERT address + INSERT audit log
	batch := &pgx.Batch{}

	// Query 1: Insert agent address
	// BR-AGT-PRF-008: Multiple Address Types Support (OFFICIAL, PERMANENT, COMMUNICATION)
	query1 := dblib.Psql.Insert(agentAddressTable).
		Columns(
			"agent_id", "address_type", "address_line1", "address_line2", "village",
			"taluka", "city", "district", "state", "country", "pincode",
			"is_same_as_permanent", "effective_from", "metadata", "created_by",
		).
		Values(
			address.AgentID, address.AddressType, address.AddressLine1, address.AddressLine2,
			address.Village, address.Taluka, address.City, address.District, address.State,
			address.Country, address.Pincode, address.IsSameAsPermanent, address.EffectiveFrom,
			address.Metadata, address.CreatedBy,
		).
		Suffix("RETURNING address_id, created_at, version")

	var result domain.AgentAddress
	err := dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.AgentAddress], &result)
	if err != nil {
		return nil, err
	}

	// Query 2: Insert audit log for address creation
	// BR-AGT-PRF-005: Audit Logging
	query2 := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(address.AgentID, domain.AuditActionAddressUpdate, "address_type", address.AddressType, "New address added", address.CreatedBy, time.Now())

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
	result = address
	return &result, nil
}

// FindByID retrieves an address by ID
func (r *AgentAddressRepository) FindByID(ctx context.Context, addressID string) (*domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAddressTable).
		Where(sq.Eq{"address_id": addressID, "deleted_at": nil})

	var address domain.AgentAddress
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAddress], &address)
	if err != nil {
		return nil, err
	}

	return &address, nil
}

// FindByAgentID retrieves all addresses for an agent
// FR-AGT-PRF-009: Address Management
// BR-AGT-PRF-008: Multiple Address Types Support
func (r *AgentAddressRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAddressTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("effective_from DESC")

	var addresses []domain.AgentAddress
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAddress], &addresses)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

// FindByAgentIDAndType retrieves a specific address type for an agent
// BR-AGT-PRF-008: Multiple Address Types Support
func (r *AgentAddressRepository) FindByAgentIDAndType(ctx context.Context, agentID, addressType string) (*domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentAddressTable).
		Where(sq.Eq{"agent_id": agentID, "address_type": addressType, "deleted_at": nil}).
		OrderBy("effective_from DESC").
		Limit(1)

	var address domain.AgentAddress
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentAddress], &address)
	if err != nil {
		return nil, err
	}

	return &address, nil
}

// Update updates an agent address
// FR-AGT-PRF-009: Address Management
// VR-AGT-PRF-008 to VR-AGT-PRF-012: Address Validations
func (r *AgentAddressRepository) Update(ctx context.Context, addressID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Update address
	updateQuery := dblib.Psql.Update(agentAddressTable).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"address_id": addressID, "deleted_at": nil})

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
		From(agentAddressTable).
		Where(sq.Eq{"address_id": addressID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit logs for each field update
	// BR-AGT-PRF-005: Audit Logging
	for field, newValue := range updates {
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "performed_by", "performed_at").
			Values(agentID, domain.AuditActionAddressUpdate, field, newValue, updatedBy, time.Now())

		err = dblib.QueueExecRow(batch, auditQuery)
		if err != nil {
			return err
		}
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes an agent address
func (r *AgentAddressRepository) Delete(ctx context.Context, addressID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for delete + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Soft delete address
	updateQuery := dblib.Psql.Update(agentAddressTable).
		Set("deleted_at", time.Now()).
		Set("updated_by", deletedBy).
		Where(sq.Eq{"address_id": addressID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentAddressTable).
		Where(sq.Eq{"address_id": addressID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionDelete, "address", "Address deleted", deletedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchCreate inserts multiple agent addresses in a single transaction
// OPTIMIZATION: Batch operation for multiple address inserts
// FR-AGT-PRF-009: Address Management
func (r *AgentAddressRepository) BatchCreate(ctx context.Context, addresses []domain.AgentAddress) ([]domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch for multiple address inserts with audit logs
	// OPTIMIZATION: Batch operation combines multiple INSERTs in single round-trip
	batch := &pgx.Batch{}
	results := make([]domain.AgentAddress, len(addresses))

	for i, address := range addresses {
		// Insert address
		insertQuery := dblib.Psql.Insert(agentAddressTable).
			Columns(
				"agent_id", "address_type", "address_line1", "address_line2", "village",
				"taluka", "city", "district", "state", "country", "pincode",
				"is_same_as_permanent", "effective_from", "metadata", "created_by",
			).
			Values(
				address.AgentID, address.AddressType, address.AddressLine1, address.AddressLine2,
				address.Village, address.Taluka, address.City, address.District, address.State,
				address.Country, address.Pincode, address.IsSameAsPermanent, address.EffectiveFrom,
				address.Metadata, address.CreatedBy,
			).
			Suffix("RETURNING address_id, created_at, version")

		err := dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByNameLax[domain.AgentAddress], &results[i])
		if err != nil {
			return nil, err
		}

		// Insert audit log
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
			Values(address.AgentID, domain.AuditActionAddressUpdate, "address_type", address.AddressType, "New address added", address.CreatedBy, time.Now())

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
	for i := range addresses {
		results[i] = addresses[i]
	}

	return results, nil
}

// CopyCommunicationFromPermanent copies permanent address to communication address
// BR-AGT-PRF-009: Communication Address Same as Permanent Option
func (r *AgentAddressRepository) CopyCommunicationFromPermanent(ctx context.Context, agentID, createdBy string) (*domain.AgentAddress, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch to get permanent address and create communication address
	// OPTIMIZATION: Batch combines SELECT + INSERT
	batch := &pgx.Batch{}

	// Query 1: Get permanent address
	var permanentAddress domain.AgentAddress
	selectQuery := dblib.Psql.Select("*").
		From(agentAddressTable).
		Where(sq.Eq{"agent_id": agentID, "address_type": domain.AddressTypePermanent, "deleted_at": nil}).
		OrderBy("effective_from DESC").
		Limit(1)

	err := dblib.QueueReturnRow(batch, selectQuery, pgx.RowToStructByNameLax[domain.AgentAddress], &permanentAddress)
	if err != nil {
		return nil, err
	}

	// Execute first batch to get permanent address
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Create communication address with same data
	communicationAddress := permanentAddress
	communicationAddress.AddressID = "" // Will be generated
	communicationAddress.AddressType = domain.AddressTypeCommunication
	communicationAddress.IsSameAsPermanent = true
	communicationAddress.CreatedBy = createdBy
	communicationAddress.EffectiveFrom = time.Now()

	// Create new address
	return r.Create(ctx, communicationAddress)
}
