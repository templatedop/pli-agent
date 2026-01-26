package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"pli-agent-api/core/domain"
	dbutil "pli-agent-api/db"
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

	// Use CTE to combine INSERT profile + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// VR-AGT-PRF-001 to VR-AGT-PRF-007: Personal Information Validation
	// VR-AGT-PRF-003: PAN Format Validation
	// VR-AGT-PRF-004: Aadhar Format Validation
	// BR-AGT-PRF-005: Name Update with Audit Logging
	batch := &pgx.Batch{}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_profiles (
				agent_type, employee_id, office_code, circle_id, division_id,
				advisor_coordinator_id, title, first_name, middle_name, last_name,
				gender, date_of_birth, category, marital_status, aadhar_number,
				pan_number, designation_rank, service_number, professional_title,
				status, status_date, distribution_channel, product_class,
				external_identification_number, workflow_state, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, action_reason, performed_by, performed_at)
		SELECT agent_id, $27, $28, $29, $30
		FROM inserted
		RETURNING (SELECT ROW(agent_id, agent_code, agent_type, employee_id, office_code, circle_id, division_id,
			advisor_coordinator_id, title, first_name, middle_name, last_name, gender, date_of_birth,
			category, marital_status, aadhar_number, pan_number, designation_rank, service_number,
			professional_title, status, status_date, status_reason, distribution_channel, product_class,
			external_identification_number, workflow_state, created_at, created_by, updated_at, updated_by,
			deleted_at, version) FROM inserted)
	`

	args := []interface{}{
		profile.AgentType, profile.EmployeeID, profile.OfficeCode, profile.CircleID,
		profile.DivisionID, profile.AdvisorCoordinatorID, profile.Title, profile.FirstName,
		profile.MiddleName, profile.LastName, profile.Gender, profile.DateOfBirth,
		profile.Category, profile.MaritalStatus, profile.AadharNumber, profile.PANNumber,
		profile.DesignationRank, profile.ServiceNumber, profile.ProfessionalTitle,
		profile.Status, profile.StatusDate, profile.DistributionChannel, profile.ProductClass,
		profile.ExternalIdentificationNumber, profile.WorkflowState, profile.CreatedBy,
		domain.AuditActionCreate, "Agent profile created", profile.CreatedBy, time.Now(),
	}

	var result domain.AgentProfile
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentProfile], &result)
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
	Status        string
	AgentType     string
	CircleID      string
	DivisionID    string
	CoordinatorID string
	OfficeCode    string
	Name          string
	MobileNumber  string
	Email         string
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

	// Use CTE to combine UPDATE + INSERT audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Build SET clause dynamically
	setClauses := "updated_at = $2, updated_by = $3"
	args := []interface{}{agentID, time.Now(), updatedBy}
	argIndex := 4

	for field, value := range updates {
		setClauses += fmt.Sprintf(", %s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	// Build audit log values for UNNEST
	fieldNames := []string{}
	newValues := []interface{}{}
	for field, value := range updates {
		fieldNames = append(fieldNames, field)
		newValues = append(newValues, value)
	}

	sql := fmt.Sprintf(`
		WITH updated AS (
			UPDATE agent_profiles
			SET %s
			WHERE agent_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, performed_by, performed_at)
		SELECT agent_id, $%d, unnest($%d::text[]), unnest($%d::text[]), $%d, $%d
		FROM updated
	`, setClauses, argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4)

	args = append(args, domain.AuditActionUpdate, fieldNames, newValues, updatedBy, time.Now())

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
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

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-016: Status Update with Mandatory Reason
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_profiles
			SET status = $2, status_date = $3, status_reason = $4, updated_at = $5, updated_by = $6
			WHERE agent_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $7, $8, $9, $10, $11, $12
		FROM updated
	`

	args := []interface{}{
		agentID, status, time.Now(), reason, time.Now(), updatedBy,
		domain.AuditActionStatusChange, "status", status, reason, updatedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
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

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-017: Agent Termination Workflow
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_profiles
			SET status = $2, status_date = $3, status_reason = $4, updated_at = $5, updated_by = $6
			WHERE agent_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $7, $8, $9, $10, $11, $12
		FROM updated
	`

	args := []interface{}{
		agentID, domain.AgentStatusTerminated, effectiveDate, reason, time.Now(), terminatedBy,
		domain.AuditActionTerminate, "status", domain.AgentStatusTerminated, reason, terminatedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
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
			"agent_type": domain.AgentTypeAdvisorCoordinator,
			"status":     domain.AgentStatusActive,
			"deleted_at": nil,
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

// ========================================================================
// ATOMIC BATCH OPERATIONS (Single Round-Trip to Database)
// ========================================================================

// CreateWithRelatedEntitiesInput holds all data for atomic profile creation
type CreateWithRelatedEntitiesInput struct {
	Profile   domain.AgentProfile
	Addresses []domain.AgentAddress
	Contacts  []domain.AgentContact
	Emails    []domain.AgentEmail
}

// CreateWithRelatedEntities atomically creates agent profile with all related entities
// Single database round trip using CTE pattern with UNNEST for bulk inserts
// ACT-024: CreateAgentProfileActivity
// FR-AGT-PRF-001: New Profile Creation
// Ensures atomicity: Either all entities are created, or none are (transaction)
func (r *AgentProfileRepository) CreateWithRelatedEntities(ctx context.Context, input CreateWithRelatedEntitiesInput) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	profile := input.Profile

	// Build arrays for UNNEST bulk insert
	var (
		// Address arrays
		addrAgentIDs    []string
		addrTypes       []string
		addrLine1s      []string
		addrLine2s      []string
		addrLine3s      []string
		addrCities      []string
		addrDistricts   []string
		addrStates      []string
		addrCountries   []string
		addrPincodes    []string
		addrIsPrimaries []bool
		addrValidFroms  []time.Time
		addrCreatedBys  []string

		// Contact arrays
		contactAgentIDs    []string
		contactTypes       []string
		contactNumbers     []string
		contactIsPrimaries []bool
		contactIsVerifieds []bool
		contactCreatedBys  []string

		// Email arrays
		emailAgentIDs    []string
		emailAddresses   []string
		emailIsPrimaries []bool
		emailIsVerifieds []bool
		emailCreatedBys  []string
	)

	// We'll use a placeholder for agent_id - it will be filled from the CTE
	agentIDPlaceholder := "<<AGENT_ID>>"

	// Prepare address arrays
	for _, addr := range input.Addresses {
		addrAgentIDs = append(addrAgentIDs, agentIDPlaceholder)
		addrTypes = append(addrTypes, addr.AddressType)
		addrLine1s = append(addrLine1s, addr.Line1)
		addrLine2s = append(addrLine2s, addr.Line2.String)
		addrLine3s = append(addrLine3s, addr.Line3.String)
		addrCities = append(addrCities, addr.City)
		addrDistricts = append(addrDistricts, addr.District.String)
		addrStates = append(addrStates, addr.State)
		addrCountries = append(addrCountries, addr.Country)
		addrPincodes = append(addrPincodes, addr.Pincode)
		addrIsPrimaries = append(addrIsPrimaries, addr.IsPrimary)
		addrValidFroms = append(addrValidFroms, addr.ValidFrom)
		addrCreatedBys = append(addrCreatedBys, profile.CreatedBy)
	}

	// Prepare contact arrays
	for _, contact := range input.Contacts {
		contactAgentIDs = append(contactAgentIDs, agentIDPlaceholder)
		contactTypes = append(contactTypes, contact.ContactType)
		contactNumbers = append(contactNumbers, contact.ContactNumber)
		contactIsPrimaries = append(contactIsPrimaries, contact.IsPrimary)
		contactIsVerifieds = append(contactIsVerifieds, contact.IsVerified)
		contactCreatedBys = append(contactCreatedBys, profile.CreatedBy)
	}

	// Prepare email arrays
	for _, email := range input.Emails {
		emailAgentIDs = append(emailAgentIDs, agentIDPlaceholder)
		emailAddresses = append(emailAddresses, email.EmailAddress)
		emailIsPrimaries = append(emailIsPrimaries, email.IsPrimary)
		emailIsVerifieds = append(emailIsVerifieds, email.IsVerified)
		emailCreatedBys = append(emailCreatedBys, profile.CreatedBy)
	}

	// Build the complex CTE query
	// Uses CTEs to:
	// 1. INSERT profile and get agent_id
	// 2. INSERT addresses using UNNEST with agent_id from step 1
	// 3. INSERT contacts using UNNEST with agent_id from step 1
	// 4. INSERT emails using UNNEST with agent_id from step 1
	// 5. INSERT audit log with agent_id from step 1
	// All in single atomic transaction
	sql := `
		WITH inserted_profile AS (
			INSERT INTO agent_profiles (
				agent_type, employee_id, office_code, circle_id, division_id,
				advisor_coordinator_id, title, first_name, middle_name, last_name,
				gender, date_of_birth, category, marital_status, aadhar_number,
				pan_number, designation_rank, service_number, professional_title,
				status, status_date, distribution_channel, product_class,
				external_identification_number, workflow_state, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
			RETURNING *
		),
		inserted_addresses AS (
			INSERT INTO agent_addresses (agent_id, address_type, line1, line2, line3, city, district, state, country, pincode, is_primary, valid_from, created_by)
			SELECT ip.agent_id, addr_type, addr_line1, addr_line2, addr_line3, addr_city, addr_district, addr_state, addr_country, addr_pincode, addr_is_primary, addr_valid_from, addr_created_by
			FROM inserted_profile ip
			CROSS JOIN UNNEST(
				$27::text[], $28::text[], $29::text[], $30::text[], $31::text[], $32::text[], $33::text[], $34::text[], $35::text[],
				$36::boolean[], $37::timestamp[], $38::text[]
			) AS t(addr_type, addr_line1, addr_line2, addr_line3, addr_city, addr_district, addr_state, addr_country, addr_pincode, addr_is_primary, addr_valid_from, addr_created_by)
			WHERE ARRAY_LENGTH($27::text[], 1) > 0
			RETURNING *
		),
		inserted_contacts AS (
			INSERT INTO agent_contacts (agent_id, contact_type, contact_number, is_primary, is_verified, created_by)
			SELECT ip.agent_id, contact_type, contact_number, contact_is_primary, contact_is_verified, contact_created_by
			FROM inserted_profile ip
			CROSS JOIN UNNEST(
				$39::text[], $40::text[], $41::boolean[], $42::boolean[], $43::text[]
			) AS t(contact_type, contact_number, contact_is_primary, contact_is_verified, contact_created_by)
			WHERE ARRAY_LENGTH($39::text[], 1) > 0
			RETURNING *
		),
		inserted_emails AS (
			INSERT INTO agent_emails (agent_id, email_address, is_primary, is_verified, created_by)
			SELECT ip.agent_id, email_address, email_is_primary, email_is_verified, email_created_by
			FROM inserted_profile ip
			CROSS JOIN UNNEST(
				$44::text[], $45::boolean[], $46::boolean[], $47::text[]
			) AS t(email_address, email_is_primary, email_is_verified, email_created_by)
			WHERE ARRAY_LENGTH($44::text[], 1) > 0
			RETURNING *
		),
		inserted_audit AS (
			INSERT INTO agent_audit_logs (agent_id, action_type, action_reason, performed_by, performed_at)
			SELECT agent_id, $48, $49, $50, $51
			FROM inserted_profile
			RETURNING *
		)
		SELECT * FROM inserted_profile
	`

	args := []interface{}{
		// Profile fields ($1 to $26)
		profile.AgentType, profile.EmployeeID, profile.OfficeCode, profile.CircleID,
		profile.DivisionID, profile.AdvisorCoordinatorID, profile.Title, profile.FirstName,
		profile.MiddleName, profile.LastName, profile.Gender, profile.DateOfBirth,
		profile.Category, profile.MaritalStatus, profile.AadharNumber, profile.PANNumber,
		profile.DesignationRank, profile.ServiceNumber, profile.ProfessionalTitle,
		profile.Status, profile.StatusDate, profile.DistributionChannel, profile.ProductClass,
		profile.ExternalIdentificationNumber, profile.WorkflowState, profile.CreatedBy,
		// Address arrays ($27 to $38)
		addrTypes, addrLine1s, addrLine2s, addrLine3s, addrCities, addrDistricts,
		addrStates, addrCountries, addrPincodes, addrIsPrimaries, addrValidFroms, addrCreatedBys,
		// Contact arrays ($39 to $43)
		contactTypes, contactNumbers, contactIsPrimaries, contactIsVerifieds, contactCreatedBys,
		// Email arrays ($44 to $47)
		emailAddresses, emailIsPrimaries, emailIsVerifieds, emailCreatedBys,
		// Audit fields ($48 to $51)
		domain.AuditActionCreate, "Agent profile created", profile.CreatedBy, time.Now(),
	}

	batch := &pgx.Batch{}
	var result domain.AgentProfile
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentProfile], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to queue profile creation: %w", err)
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, fmt.Errorf("failed to execute profile creation: %w", err)
	}

	return &result, nil
}


// SearchAgents performs multi-criteria agent search
// AGT-022: Multi-criteria Agent Search
// FR-AGT-PRF-004: Agent Search Functionality
// BR-AGT-PRF-022: Multi-Criteria Search Support
// CRITICAL: Single database round trip using batch for count + data
func (r *AgentProfileRepository) SearchAgents(
	ctx context.Context,
	agentID, name, panNumber, mobileNumber, status, officeCode string,
	page, limit int,
) ([]domain.AgentProfile, int64, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Build base query with Squirrel
	baseQuery := sq.Select().From(agentProfileTable).Where(sq.Eq{"deleted_at": nil})

	// Apply filters dynamically
	if agentID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"agent_id": agentID})
	}
	if name != "" {
		namePattern := "%" + name + "%"
		baseQuery = baseQuery.Where(
			sq.Or{
				sq.ILike{"first_name": namePattern},
				sq.ILike{"last_name": namePattern},
			},
		)
	}
	if panNumber != "" {
		baseQuery = baseQuery.Where(sq.Eq{"pan_number": panNumber})
	}
	if status != "" {
		baseQuery = baseQuery.Where(sq.Eq{"status": status})
	}
	if officeCode != "" {
		baseQuery = baseQuery.Where(sq.Eq{"office_code": officeCode})
	}

	// Note: Mobile number search would require JOIN with contacts table
	// For now, skipping mobile filter in Squirrel approach
	// TODO: Implement mobile search with JOIN if needed

	// Calculate offset
	offset := (page - 1) * limit

	// Use batch to get count and data in single round trip
	batch := &pgx.Batch{}

	// Query 1: Count total records
	countQuery := baseQuery.Columns("COUNT(*)")
	var totalCount int64
	err := dbutil.QueueReturnRow(batch, countQuery, pgx.RowTo[int64], &totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue count query: %w", err)
	}

	// Query 2: Get paginated data
	dataQuery := baseQuery.
		Columns("*").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	var agents []domain.AgentProfile
	err = dbutil.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.AgentProfile], &agents)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue data query: %w", err)
	}

	// Execute batch - single database round trip
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search: %w", err)
	}

	return agents, totalCount, nil
}

// GetAgentProfileWithDetails retrieves complete agent profile with related data
// AGT-023: Get Agent Profile Details
// FR-AGT-PRF-005: Profile Dashboard View
// BR-AGT-PRF-023: Dashboard Profile View
// CRITICAL: Single database round trip using JSON aggregation
func (r *AgentProfileRepository) GetAgentProfileWithDetails(ctx context.Context, agentID string) (map[string]interface{}, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE and JSON aggregation to fetch profile + addresses + contacts + emails in single query
	// This is complex JSON building, so using raw SQL with dblib.SelectOne
	sql := `
		WITH profile_data AS (
			SELECT * FROM agent_profiles WHERE agent_id = $1 AND deleted_at IS NULL
		),
		addresses_data AS (
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'address_id', address_id,
						'address_type', address_type,
						'line1', line1,
						'line2', line2,
						'line3', line3,
						'city', city,
						'district', district,
						'state', state,
						'country', country,
						'pincode', pincode,
						'is_primary', is_primary,
						'valid_from', valid_from,
						'valid_to', valid_to
					) ORDER BY is_primary DESC, created_at
				),
				'[]'::json
			) as addresses
			FROM agent_addresses
			WHERE agent_id = $1 AND deleted_at IS NULL
		),
		contacts_data AS (
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'contact_id', contact_id,
						'contact_type', contact_type,
						'contact_number', contact_number,
						'is_primary', is_primary,
						'is_verified', is_verified,
						'verified_at', verified_at
					) ORDER BY is_primary DESC, created_at
				),
				'[]'::json
			) as contacts
			FROM agent_contacts
			WHERE agent_id = $1 AND deleted_at IS NULL
		),
		emails_data AS (
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'email_id', email_id,
						'email_address', email_address,
						'is_primary', is_primary,
						'is_verified', is_verified,
						'verified_at', verified_at
					) ORDER BY is_primary DESC, created_at
				),
				'[]'::json
			) as emails
			FROM agent_emails
			WHERE agent_id = $1 AND deleted_at IS NULL
		)
		SELECT
			row_to_json(pd.*)::text as profile,
			ad.addresses::text as addresses,
			cd.contacts::text as contacts,
			ed.emails::text as emails
		FROM profile_data pd
		CROSS JOIN addresses_data ad
		CROSS JOIN contacts_data cd
		CROSS JOIN emails_data ed
	`

	// Use struct to receive the JSON strings
	type ProfileResult struct {
		Profile   string `db:"profile"`
		Addresses string `db:"addresses"`
		Contacts  string `db:"contacts"`
		Emails    string `db:"emails"`
	}

	// Execute with dblib.SelectRows (raw SQL, non-Squirrel)
	rows, err := r.db.Query(cCtx, sql, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent profile: %w", err)
	}

	// Use dblib.SelectRows pattern
	results, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ProfileResult])
	if err != nil {
		return nil, fmt.Errorf("failed to collect profile data: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	result := results[0]

	// Build response map
	responseMap := map[string]interface{}{
		"profile":   result.Profile,
		"addresses": result.Addresses,
		"contacts":  result.Contacts,
		"emails":    result.Emails,
	}

	return responseMap, nil
}

// GetAgentUpdateFormData retrieves data for updating a specific section
// AGT-024: Get Profile Update Form
// FR-AGT-PRF-006: Personal Information Update
// Returns section-specific data for form population
func (r *AgentProfileRepository) GetAgentUpdateFormData(ctx context.Context, agentID, section string) (map[string]interface{}, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	var sql string
	switch section {
	case "personal_info":
		sql = `
			SELECT row_to_json(t)::text as form_data
			FROM (
				SELECT
					agent_id, title, first_name, middle_name, last_name,
					gender, date_of_birth, category, marital_status,
					aadhar_number, pan_number, version
				FROM agent_profiles
				WHERE agent_id = $1 AND deleted_at IS NULL
			) t`
	case "address":
		sql = `
			SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)::text as form_data
			FROM (
				SELECT
					address_id, address_type, line1, line2, line3,
					city, district, state, country, pincode,
					is_primary, version
				FROM agent_addresses
				WHERE agent_id = $1 AND deleted_at IS NULL
			) t`
	case "contact":
		sql = `
			SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)::text as form_data
			FROM (
				SELECT
					contact_id, contact_type, contact_number,
					is_primary, version
				FROM agent_contacts
				WHERE agent_id = $1 AND deleted_at IS NULL
			) t`
	default:
		return nil, fmt.Errorf("unsupported section: %s", section)
	}

	// Execute with dblib.SelectOne pattern (raw SQL)
	rows, err := r.db.Query(cCtx, sql, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get update form data: %w", err)
	}

	// Use pgx.CollectOneRow for single row query
	formDataJSON, err := pgx.CollectOneRow(rows, pgx.RowTo[string])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("agent not found or no data for section: %s", section)
		}
		return nil, fmt.Errorf("failed to scan form data: %w", err)
	}

	// Return as map
	result := map[string]interface{}{
		"form_data": formDataJSON,
	}

	return result, nil
}

// UpdateAgentPersonalInfoReturning updates personal information with UPDATE...RETURNING
// AGT-025: Update Profile Section (personal_info)
// FR-AGT-PRF-006: Personal Information Update
// BR-AGT-PRF-005: Name Update with Audit Logging
// BR-AGT-PRF-006: PAN Update with Format and Uniqueness Validation
// CRITICAL: Single atomic operation with UPDATE...RETURNING
func (r *AgentProfileRepository) UpdateAgentPersonalInfoReturning(
	ctx context.Context,
	agentID string,
	updates map[string]interface{},
	updatedBy string,
) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Build dynamic UPDATE query with Squirrel
	updateQuery := dblib.Psql.Update(agentProfileTable).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Set("version", sq.Expr("version + 1")).
		Where(sq.And{
			sq.Eq{"agent_id": agentID},
			sq.Eq{"deleted_at": nil},
		})

	// Add dynamic fields from updates map
	for field, value := range updates {
		updateQuery = updateQuery.Set(field, value)
	}

	// Add RETURNING clause
	updateQuery = updateQuery.Suffix("RETURNING *")

	var result domain.AgentProfile
	err := dblib.SelectOne(cCtx, r.db, updateQuery, pgx.RowToStructByNameLax[domain.AgentProfile], &result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("agent not found: %s", agentID)
		}
		return nil, fmt.Errorf("failed to update personal info: %w", err)
	}

	return &result, nil
}

// CreateApprovalRequestWithChanges handles both approval and direct update cases
// CRITICAL: Single database round trip using batch
// If requiresApproval: Creates approval request
// If !requiresApproval: Directly updates profile based on section
func (r *AgentProfileRepository) CreateApprovalRequestWithChanges(
	ctx context.Context,
	agentID, section string,
	changes map[string]interface{},
	requestedBy string,
	requiresApproval bool,
) (approvalRequestID *string, updatedProfile *domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	if requiresApproval {
		// Case 1: Create approval request
		changesJSON, err := json.Marshal(changes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal changes: %w", err)
		}

		insertQuery := dblib.Psql.Insert("approval_requests").
			Columns("agent_id", "section", "requested_changes", "requested_by", "requested_at", "status").
			Values(agentID, section, string(changesJSON), requestedBy, time.Now(), "PENDING").
			Suffix("RETURNING approval_request_id")

		var approvalID string
		err = dbutil.QueueReturnRow(batch, insertQuery, pgx.RowTo[string], &approvalID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to queue approval request: %w", err)
		}

		// Execute batch
		err = r.db.SendBatch(cCtx, batch).Close()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create approval request: %w", err)
		}

		return &approvalID, nil, nil

	} else {
		// Case 2: Direct update based on section
		var profile domain.AgentProfile

		switch section {
		case "personal_info":
			// Extract update map from changes
			updateMap := make(map[string]interface{})
			for field, changeObj := range changes {
				if change, ok := changeObj.(map[string]interface{}); ok {
					if newValue, exists := change["new_value"]; exists {
						updateMap[field] = newValue
					}
				} else {
					// If not in change object format, use directly
					updateMap[field] = changeObj
				}
			}

			// Build dynamic UPDATE query with Squirrel
			updateQuery := dblib.Psql.Update(agentProfileTable).
				Set("updated_at", time.Now()).
				Set("updated_by", requestedBy).
				Set("version", sq.Expr("version + 1")).
				Where(sq.And{
					sq.Eq{"agent_id": agentID},
					sq.Eq{"deleted_at": nil},
				})

			// Add dynamic fields from updates map
			for field, value := range updateMap {
				updateQuery = updateQuery.Set(field, value)
			}

			// Add RETURNING clause
			updateQuery = updateQuery.Suffix("RETURNING *")

			// Queue the update query
			err := dbutil.QueueReturnRow(batch, updateQuery, pgx.RowToStructByNameLax[domain.AgentProfile], &profile)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to queue profile update: %w", err)
			}

		case "address":
			// TODO: Implement address update
			return nil, nil, fmt.Errorf("address update not implemented yet")

		case "contact":
			// TODO: Implement contact update
			return nil, nil, fmt.Errorf("contact update not implemented yet")

		default:
			return nil, nil, fmt.Errorf("unsupported section: %s", section)
		}

		// Execute batch
		err := r.db.SendBatch(cCtx, batch).Close()
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, nil, fmt.Errorf("agent not found: %s", agentID)
			}
			return nil, nil, fmt.Errorf("failed to update profile: %w", err)
		}

		return nil, &profile, nil
	}
}

// ApproveAndApplyProfileChangesReturning approves request and applies changes in single transaction
// AGT-026: Approve Profile Update
// CRITICAL: Single database round trip using CTE to combine approval + profile update
func (r *AgentProfileRepository) ApproveAndApplyProfileChangesReturning(
	ctx context.Context,
	approvalRequestID, reviewedBy string,
	comments *string,
) (*domain.AgentProfile, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Build CTE that:
	// 1. Updates approval_requests to APPROVED
	// 2. Gets the requested changes
	// 3. Updates agent_profiles with those changes
	// All in ONE database round trip
	commentsVal := ""
	if comments != nil {
		commentsVal = *comments
	}

	sql := `
		WITH updated_approval AS (
			UPDATE approval_requests
			SET
				status = 'APPROVED',
				reviewed_by = $2,
				reviewed_at = NOW(),
				review_comments = $3,
				updated_at = NOW()
			WHERE approval_request_id = $1 AND status = 'PENDING'
			RETURNING agent_id, section, requested_changes
		),
		changes_parsed AS (
			SELECT
				agent_id,
				section,
				requested_changes::jsonb as changes
			FROM updated_approval
		),
		updated_profile AS (
			UPDATE agent_profiles ap
			SET
				first_name = COALESCE((cp.changes->>'first_name')::text, ap.first_name),
				middle_name = COALESCE((cp.changes->>'middle_name')::text, ap.middle_name),
				last_name = COALESCE((cp.changes->>'last_name')::text, ap.last_name),
				pan_number = COALESCE((cp.changes->>'pan_number')::text, ap.pan_number),
				updated_at = NOW(),
				updated_by = $2,
				version = ap.version + 1
			FROM changes_parsed cp
			WHERE ap.agent_id = cp.agent_id AND cp.section = 'personal_info'
			RETURNING ap.*
		)
		SELECT * FROM updated_profile
	`

	var result domain.AgentProfile
	err := r.db.QueryRow(cCtx, sql, approvalRequestID, reviewedBy, commentsVal).Scan(
		&result.AgentID, &result.AgentCode, &result.AgentType, &result.EmployeeID,
		&result.OfficeCode, &result.CircleID, &result.DivisionID, &result.AdvisorCoordinatorID,
		&result.Title, &result.FirstName, &result.MiddleName, &result.LastName,
		&result.Gender, &result.DateOfBirth, &result.Category, &result.MaritalStatus,
		&result.AadharNumber, &result.PANNumber, &result.DesignationRank,
		&result.ServiceNumber, &result.ProfessionalTitle, &result.Status,
		&result.StatusDate, &result.StatusReason, &result.DistributionChannel,
		&result.ProductClass, &result.ExternalIdentificationNumber,
		&result.WorkflowState, &result.CreatedAt, &result.CreatedBy,
		&result.UpdatedAt, &result.UpdatedBy, &result.DeletedAt, &result.Version,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("approval request not found or already processed")
		}
		return nil, fmt.Errorf("failed to approve and apply changes: %w", err)
	}

	return &result, nil
}
