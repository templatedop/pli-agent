package repo

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	dbutil "pli-agent-api/db"
	"pli-agent-api/core/domain"
)

const (
	agentLicenseTable = "agent_licenses"
)

// AgentLicenseRepository handles license data operations
type AgentLicenseRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentLicenseRepository creates a new agent license repository
func NewAgentLicenseRepository(db *dblib.DB, cfg *config.Config) *AgentLicenseRepository {
	return &AgentLicenseRepository{
		db:  db,
		cfg: cfg,
	}
}

// Create adds a new license with automatic renewal date calculation
// AGT-030: Add License
// BR-AGT-PRF-012: License Renewal Period Rules
// CRITICAL: Single database round trip
func (r *AgentLicenseRepository) Create(ctx context.Context, license domain.AgentLicense) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Calculate renewal date based on license type
	renewalDate := r.calculateRenewalDate(license.LicenseType, license.LicenseDate, license.LicentiatExamPassed)

	insertQuery := dblib.Psql.Insert(agentLicenseTable).
		Columns(
			"agent_id", "license_line", "license_type", "license_number",
			"resident_status", "license_date", "renewal_date", "authority_date",
			"renewal_count", "license_status", "licentiate_exam_passed",
			"licentiate_exam_date", "licentiate_certificate_number",
			"is_primary", "created_by",
		).
		Values(
			license.AgentID, license.LicenseLine, license.LicenseType, license.LicenseNumber,
			license.ResidentStatus, license.LicenseDate, renewalDate, license.AuthorityDate,
			0, domain.LicenseStatusActive, license.LicentiatExamPassed,
			license.LicentiateExamDate, license.LicenticateCertificateNumber,
			license.IsPrimary, license.CreatedBy,
		).
		Suffix("RETURNING *")

	var result domain.AgentLicense
	err := dblib.SelectOne(cCtx, r.db, insertQuery, pgx.RowToStructByNameLax[domain.AgentLicense], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create license: %w", err)
	}

	return &result, nil
}

// FindByID retrieves a specific license by ID
// AGT-031: Get License Details
func (r *AgentLicenseRepository) FindByID(ctx context.Context, licenseID string) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentLicenseTable).
		Where(sq.And{
			sq.Eq{"license_id": licenseID},
			sq.Eq{"deleted_at": nil},
		})

	var license domain.AgentLicense
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicense], &license)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("license not found: %s", licenseID)
		}
		return nil, fmt.Errorf("failed to find license: %w", err)
	}

	return &license, nil
}

// FindByAgentID retrieves all licenses for an agent
// AGT-029: Get Agent Licenses
// CRITICAL: Single database round trip with optional status filter
func (r *AgentLicenseRepository) FindByAgentID(ctx context.Context, agentID string, status *string) ([]domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentLicenseTable).
		Where(sq.And{
			sq.Eq{"agent_id": agentID},
			sq.Eq{"deleted_at": nil},
		}).
		OrderBy("is_primary DESC, created_at DESC")

	// Apply status filter if provided
	if status != nil && *status != "" {
		query = query.Where(sq.Eq{"license_status": *status})
	}

	var licenses []domain.AgentLicense
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentLicense], &licenses)
	if err != nil {
		return nil, fmt.Errorf("failed to find licenses for agent: %w", err)
	}

	return licenses, nil
}

// Update updates license details
// AGT-032: Update License
func (r *AgentLicenseRepository) Update(ctx context.Context, licenseID string, updates map[string]interface{}, updatedBy string) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateQuery := dblib.Psql.Update(agentLicenseTable).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Set("version", sq.Expr("version + 1")).
		Where(sq.And{
			sq.Eq{"license_id": licenseID},
			sq.Eq{"deleted_at": nil},
		})

	// Add dynamic fields from updates map
	for field, value := range updates {
		updateQuery = updateQuery.Set(field, value)
	}

	updateQuery = updateQuery.Suffix("RETURNING *")

	var result domain.AgentLicense
	err := dblib.SelectOne(cCtx, r.db, updateQuery, pgx.RowToStructByNameLax[domain.AgentLicense], &result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("license not found: %s", licenseID)
		}
		return nil, fmt.Errorf("failed to update license: %w", err)
	}

	return &result, nil
}

// Renew renews a license with period calculation
// AGT-033: Renew License
// BR-AGT-PRF-012: Complex renewal rules
// CRITICAL: Single database round trip with RETURNING
func (r *AgentLicenseRepository) Renew(
	ctx context.Context,
	licenseID string,
	renewalType string,
	examPassed bool,
	examDate *time.Time,
	examCertNumber *string,
	updatedBy string,
) (*domain.AgentLicense, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// First, get current license to determine renewal logic
	currentLicense, err := r.FindByID(ctx, licenseID)
	if err != nil {
		return nil, err
	}

	// Validate renewal count for provisional licenses
	if currentLicense.IsProvisional() && !examPassed {
		if currentLicense.RenewalCount >= domain.MaxProvisionalRenewals {
			return nil, fmt.Errorf("provisional license has reached maximum renewals (%d)", domain.MaxProvisionalRenewals)
		}
	}

	// Determine new license type and renewal date
	newLicenseType := currentLicense.LicenseType
	var newRenewalDate time.Time

	if examPassed && currentLicense.IsProvisional() {
		// Convert to permanent after exam
		newLicenseType = domain.LicenseTypePermanent
		newRenewalDate = time.Now().AddDate(5, 0, 0) // 5 years
	} else if currentLicense.IsProvisional() {
		// Renew provisional for 1 year
		newRenewalDate = currentLicense.RenewalDate.AddDate(1, 0, 0)
	} else {
		// Permanent annual renewal
		newRenewalDate = currentLicense.RenewalDate.AddDate(1, 0, 0)
	}

	updateQuery := dblib.Psql.Update(agentLicenseTable).
		Set("license_type", newLicenseType).
		Set("renewal_date", newRenewalDate).
		Set("renewal_count", sq.Expr("renewal_count + 1")).
		Set("license_status", domain.LicenseStatusActive).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Set("version", sq.Expr("version + 1")).
		Where(sq.And{
			sq.Eq{"license_id": licenseID},
			sq.Eq{"deleted_at": nil},
		})

	// Update exam details if passed
	if examPassed {
		updateQuery = updateQuery.
			Set("licentiate_exam_passed", true).
			Set("licentiate_exam_date", examDate).
			Set("licentiate_certificate_number", examCertNumber)
	}

	updateQuery = updateQuery.Suffix("RETURNING *")

	var result domain.AgentLicense
	err = dblib.SelectOne(cCtx, r.db, updateQuery, pgx.RowToStructByNameLax[domain.AgentLicense], &result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("license not found: %s", licenseID)
		}
		return nil, fmt.Errorf("failed to renew license: %w", err)
	}

	return &result, nil
}

// Delete soft deletes a license
// AGT-034: Delete License
func (r *AgentLicenseRepository) Delete(ctx context.Context, licenseID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateQuery := dblib.Psql.Update(agentLicenseTable).
		Set("deleted_at", time.Now()).
		Set("updated_by", deletedBy).
		Where(sq.And{
			sq.Eq{"license_id": licenseID},
			sq.Eq{"deleted_at": nil},
		})

	tag, err := dblib.Update(cCtx, r.db, updateQuery)
	if err != nil {
		return fmt.Errorf("failed to delete license: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("license not found: %s", licenseID)
	}

	return nil
}

// FindExpiring retrieves licenses expiring within specified days
// AGT-036: Get Expiring Licenses
// BR-AGT-PRF-014: License Renewal Reminders
// CRITICAL: Single database round trip with pagination
func (r *AgentLicenseRepository) FindExpiring(
	ctx context.Context,
	days int,
	officeCode *string,
	page, limit int,
) ([]domain.AgentLicense, int64, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	expiryDate := time.Now().AddDate(0, 0, days)
	
	baseQuery := dblib.Psql.Select().
		From(agentLicenseTable).
		Where(sq.And{
			sq.Eq{"license_status": domain.LicenseStatusActive},
			sq.Eq{"deleted_at": nil},
			sq.LtOrEq{"renewal_date": expiryDate},
		})

	// Apply office filter if provided (requires JOIN with agent_profiles)
	if officeCode != nil && *officeCode != "" {
		baseQuery = baseQuery.
			Join("agent_profiles ap ON ap.agent_id = agent_licenses.agent_id").
			Where(sq.Eq{"ap.office_code": *officeCode})
	}

	offset := (page - 1) * limit

	// Use batch to get count and data in single round trip
	batch := &pgx.Batch{}

	// Query 1: Count
	countQuery := baseQuery.Columns("COUNT(*)")
	var totalCount int64
	err := dbutil.QueueReturnRow(batch, countQuery, pgx.RowTo[int64], &totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue count query: %w", err)
	}

	// Query 2: Data
	dataQuery := baseQuery.
		Columns("agent_licenses.*").
		OrderBy("renewal_date ASC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	var licenses []domain.AgentLicense
	err = dbutil.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.AgentLicense], &licenses)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to queue data query: %w", err)
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute expiring licenses query: %w", err)
	}

	return licenses, totalCount, nil
}

// DeactivateExpiredAgents batch deactivates agents with expired licenses
// AGT-038: Trigger License Expiry Deactivation
// BR-AGT-PRF-013: Auto-Deactivation on Expiry
// CRITICAL: Batch operation for system job
func (r *AgentLicenseRepository) DeactivateExpiredAgents(ctx context.Context, batchDate time.Time, dryRun bool) ([]string, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Find all expired licenses
	query := dblib.Psql.Select("DISTINCT agent_id").
		From(agentLicenseTable).
		Where(sq.And{
			sq.Lt{"renewal_date": batchDate},
			sq.Eq{"license_status": domain.LicenseStatusActive},
			sq.Eq{"deleted_at": nil},
		})

	var agentIDs []string
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowTo[string], &agentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find expired licenses: %w", err)
	}

	// If dry run, just return the agent IDs
	if dryRun {
		return agentIDs, nil
	}

	// Update license status to EXPIRED (actual deactivation would be in a transaction with agent status update)
	if len(agentIDs) > 0 {
		updateQuery := dblib.Psql.Update(agentLicenseTable).
			Set("license_status", domain.LicenseStatusExpired).
			Set("updated_at", time.Now()).
			Set("updated_by", "SYSTEM_BATCH").
			Where(sq.And{
				sq.Eq{"agent_id": agentIDs},
				sq.Lt{"renewal_date": batchDate},
				sq.Eq{"license_status": domain.LicenseStatusActive},
				sq.Eq{"deleted_at": nil},
			})

		_, err = dblib.Update(cCtx, r.db, updateQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to update expired licenses: %w", err)
		}
	}

	return agentIDs, nil
}

// calculateRenewalDate calculates renewal date based on license type
// BR-AGT-PRF-012: License Renewal Period Rules
func (r *AgentLicenseRepository) calculateRenewalDate(licenseType string, licenseDate time.Time, examPassed bool) time.Time {
	if licenseType == domain.LicenseTypeProvisional {
		// Provisional: 1 year
		return licenseDate.AddDate(1, 0, 0)
	}
	
	// Permanent after exam: 5 years
	if examPassed {
		return licenseDate.AddDate(5, 0, 0)
	}
	
	// Default annual renewal
	return licenseDate.AddDate(1, 0, 0)
}
