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

// AgentProfileRepository handles all database operations for agent profiles
// E-01: Agent Profile Entity
type AgentProfileRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentProfileRepository creates a new agent profile repository
func NewAgentProfileRepository(db *dblib.DB, cfg *config.Config) *AgentProfileRepository {
	return &AgentProfileRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentProfileTable = "agent_profiles"

// Create inserts a new agent profile
// FR-AGT-PRF-001: New Profile Creation
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
func (r *AgentProfileRepository) Create(ctx context.Context, profile domain.AgentProfile) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for profile creation with audit log in single transaction
	// OPTIMIZATION: Batch operation combines INSERT profile + INSERT audit log
	batch := &pgx.Batch{}

	// Query 1: Insert agent profile
	// VR-AGT-PRF-001 to VR-AGT-PRF-007: Personal Information Validation
	// VR-AGT-PRF-003: PAN Format Validation
	// VR-AGT-PRF-004: Aadhar Format Validation
	query1 := dblib.Psql.Insert(agentProfileTable).
		Columns(
			"agent_type", "employee_id", "office_code", "circle_id", "division_id",
			"advisor_coordinator_id", "title", "first_name", "middle_name", "last_name",
			"gender", "date_of_birth", "category", "marital_status", "aadhar_number",
			"pan_number", "designation_rank", "service_number", "professional_title",
			"status", "status_date", "distribution_channel", "product_class",
			"external_identification_number", "workflow_state", "created_by",
		).
		Values(
			profile.AgentType, profile.EmployeeID, profile.OfficeCode, profile.CircleID,
			profile.DivisionID, profile.AdvisorCoordinatorID, profile.Title, profile.FirstName,
			profile.MiddleName, profile.LastName, profile.Gender, profile.DateOfBirth,
			profile.Category, profile.MaritalStatus, profile.AadharNumber, profile.PANNumber,
			profile.DesignationRank, profile.ServiceNumber, profile.ProfessionalTitle,
			profile.Status, profile.StatusDate, profile.DistributionChannel, profile.ProductClass,
			profile.ExternalIdentificationNumber, profile.WorkflowState, profile.CreatedBy,
		).
		Suffix("RETURNING agent_id, agent_code, created_at, version")

	var result domain.AgentProfile
	err := dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.AgentProfile], &result)
	if err != nil {
		return nil, err
	}

	// Query 2: Insert audit log for profile creation
	// BR-AGT-PRF-005: Name Update with Audit Logging
	query2 := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "action_reason", "performed_by", "performed_at").
		Values(result.AgentID, domain.AuditActionCreate, "Agent profile created", profile.CreatedBy, time.Now())

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
	result = profile
	return &result, nil
}

// FindByID retrieves an agent profile by ID
// FR-AGT-PRF-021: Multi-Criteria Agent Search
func (r *AgentProfileRepository) FindByID(ctx context.Context, agentID string) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil})

	var profile domain.AgentProfile
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfile], &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// FindByAgentCode retrieves an agent profile by agent code
// FR-AGT-PRF-021: Multi-Criteria Agent Search
func (r *AgentProfileRepository) FindByAgentCode(ctx context.Context, agentCode string) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileTable).
		Where(sq.Eq{"agent_code": agentCode, "deleted_at": nil})

	var profile domain.AgentProfile
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfile], &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// FindByPAN retrieves an agent profile by PAN number
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
// VR-AGT-PRF-003: PAN Format Validation
func (r *AgentProfileRepository) FindByPAN(ctx context.Context, panNumber string) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentProfileTable).
		Where(sq.Eq{"pan_number": panNumber, "deleted_at": nil})

	var profile domain.AgentProfile
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfile], &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// List retrieves agent profiles with pagination and filtering
// FR-AGT-PRF-021: Multi-Criteria Agent Search
// BR-AGT-PRF-022: Multi-Criteria Agent Search
type AgentSearchFilters struct {
	Status             string
	AgentType          string
	CircleID           string
	DivisionID         string
	CoordinatorID      string
	OfficeCode         string
	Name               string
	MobileNumber       string
	Email              string
}

func (r *AgentProfileRepository) List(ctx context.Context, filters AgentSearchFilters, skip, limit uint64, orderBy, sortType string) ([]domain.AgentProfile, int64, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Use batch to get count and data in single round trip
	// OPTIMIZATION: Batch combines COUNT + SELECT queries
	batch := &pgx.Batch{}

	// Base query with filters
	baseQuery := sq.Select().From(agentProfileTable).Where(sq.Eq{"deleted_at": nil})

	// Apply filters
	if filters.Status != "" {
		baseQuery = baseQuery.Where(sq.Eq{"status": filters.Status})
	}
	if filters.AgentType != "" {
		baseQuery = baseQuery.Where(sq.Eq{"agent_type": filters.AgentType})
	}
	if filters.CircleID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"circle_id": filters.CircleID})
	}
	if filters.DivisionID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"division_id": filters.DivisionID})
	}
	if filters.CoordinatorID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"advisor_coordinator_id": filters.CoordinatorID})
	}
	if filters.OfficeCode != "" {
		baseQuery = baseQuery.Where(sq.Eq{"office_code": filters.OfficeCode})
	}
	if filters.Name != "" {
		// Search in first_name, middle_name, last_name
		namePattern := "%" + filters.Name + "%"
		baseQuery = baseQuery.Where(
			sq.Or{
				sq.Like{"first_name": namePattern},
				sq.Like{"middle_name": namePattern},
				sq.Like{"last_name": namePattern},
			},
		)
	}

	// Query 1: Count total records
	countQuery := baseQuery.Columns("COUNT(*)")
	var totalCount int64
	err := dblib.QueueReturnRow(batch, countQuery, pgx.RowTo[int64], &totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Query 2: Get paginated data
	dataQuery := baseQuery.
		Columns("*").
		OrderBy(orderBy + " " + sortType).
		Limit(limit).
		Offset(skip)

	var profiles []domain.AgentProfile
	err = dblib.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.AgentProfile], &profiles)
	if err != nil {
		return nil, 0, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, 0, err
	}

	return profiles, totalCount, nil
}

// UpdatePersonalInfo updates agent personal information
// FR-AGT-PRF-006: Profile Update
// BR-AGT-PRF-005: Name Update with Audit Logging
// BR-AGT-PRF-007: Personal Information Update Rules
func (r *AgentProfileRepository) UpdatePersonalInfo(ctx context.Context, agentID string, updates map[string]interface{}, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Update profile
	updateQuery := dblib.Psql.Update(agentProfileTable).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil})

	// Apply updates
	for field, value := range updates {
		updateQuery = updateQuery.Set(field, value)
	}

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Insert audit logs for each field update
	// BR-AGT-PRF-005: Audit Logging
	for field, newValue := range updates {
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "performed_by", "performed_at").
			Values(agentID, domain.AuditActionUpdate, field, newValue, updatedBy, time.Now())

		err = dblib.QueueExecRow(batch, auditQuery)
		if err != nil {
			return err
		}
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// UpdateStatus updates agent status
// FR-AGT-PRF-017: Status Management
// BR-AGT-PRF-016: Status Update with Mandatory Reason
func (r *AgentProfileRepository) UpdateStatus(ctx context.Context, agentID, status, reason, updatedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for status update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Update status
	updateQuery := dblib.Psql.Update(agentProfileTable).
		Set("status", status).
		Set("status_date", time.Now()).
		Set("status_reason", reason).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Insert audit log
	// BR-AGT-PRF-016: Status Update with Mandatory Reason
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionStatusChange, "status", status, reason, updatedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Terminate terminates an agent
// FR-AGT-PRF-018: Agent Termination
// BR-AGT-PRF-017: Agent Termination Workflow
func (r *AgentProfileRepository) Terminate(ctx context.Context, agentID, reason, terminatedBy string, effectiveDate time.Time) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for termination + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Update status to TERMINATED
	// BR-AGT-PRF-017: Agent Termination Workflow
	updateQuery := dblib.Psql.Update(agentProfileTable).
		Set("status", domain.AgentStatusTerminated).
		Set("status_date", effectiveDate).
		Set("status_reason", reason).
		Set("updated_at", time.Now()).
		Set("updated_by", terminatedBy).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Insert audit log for termination
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionTerminate, "status", domain.AgentStatusTerminated, reason, terminatedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// GetActiveAdvisorCoordinators retrieves active advisor coordinators for dropdown
// FR-AGT-PRF-003: Advisor Coordinator Selection
// BR-AGT-PRF-001: Advisor Coordinator Linkage Requirement
// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
func (r *AgentProfileRepository) GetActiveAdvisorCoordinators(ctx context.Context, circleID, divisionID string) ([]domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("agent_id", "agent_code", "first_name", "middle_name", "last_name", "circle_id", "division_id").
		From(agentProfileTable).
		Where(sq.Eq{
			"agent_type":  domain.AgentTypeAdvisorCoordinator,
			"status":      domain.AgentStatusActive,
			"deleted_at":  nil,
		})

	// Apply geographic filters if provided
	// BR-AGT-PRF-002: Advisor Coordinator Geographic Assignment
	if circleID != "" {
		query = query.Where(sq.Eq{"circle_id": circleID})
	}
	if divisionID != "" {
		query = query.Where(sq.Eq{"division_id": divisionID})
	}

	query = query.OrderBy("first_name ASC")

	var coordinators []domain.AgentProfile
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentProfile], &coordinators)
	if err != nil {
		return nil, err
	}

	return coordinators, nil
}

// ValidatePANUniqueness checks if PAN is unique (excluding current agent)
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
func (r *AgentProfileRepository) ValidatePANUniqueness(ctx context.Context, panNumber, excludeAgentID string) (bool, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COUNT(*)").
		From(agentProfileTable).
		Where(sq.Eq{"pan_number": panNumber, "deleted_at": nil}).
		Where(sq.NotEq{"agent_id": excludeAgentID})

	var count int64
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowTo[int64], &count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
