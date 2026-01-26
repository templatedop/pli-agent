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

// AgentBankDetailsRepository handles all database operations for agent bank details
// E-05: Agent Bank Details Entity
// BR-AGT-PRF-018: POSB Account/Bank Account Details for Commission Disbursement
type AgentBankDetailsRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentBankDetailsRepository creates a new agent bank details repository
func NewAgentBankDetailsRepository(db *dblib.DB, cfg *config.Config) *AgentBankDetailsRepository {
	return &AgentBankDetailsRepository{
		db:  db,
		cfg: cfg,
	}
}

const agentBankDetailsTable = "agent_bank_details"

// Create inserts new agent bank details
// FR-AGT-PRF-019: Bank Details Management
// BR-AGT-PRF-018: Bank Account Details for Commission Disbursement
// VR-AGT-PRF-015: Account Number Format (encrypted storage)
// VR-AGT-PRF-016: IFSC Code Format (AAAA0123456)
// VR-AGT-PRF-017: Account Type Validation (SAVINGS, CURRENT)
func (r *AgentBankDetailsRepository) Create(ctx context.Context, bankDetails domain.AgentBankDetails, accountNumber string) (*domain.AgentBankDetails, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for bank details creation with encryption and audit log in single transaction
	// OPTIMIZATION: Batch operation combines INSERT with pgcrypto encryption + INSERT audit log
	batch := &pgx.Batch{}

	// Query 1: Insert agent bank details with encrypted account number
	// BR-AGT-PRF-018: Bank Account Details - Account number stored encrypted using pgcrypto
	// Note: pgcrypto_encrypt() function should be used in actual implementation
	// For now, using placeholder - actual encryption key management should be configured
	query1 := dblib.Psql.Insert(agentBankDetailsTable).
		Columns(
			"agent_id", "account_type", "account_number_encrypted", "ifsc_code",
			"bank_name", "branch_name", "effective_from", "metadata", "created_by",
		).
		Values(
			bankDetails.AgentID, bankDetails.AccountType,
			sq.Expr("pgp_sym_encrypt(?, current_setting('app.encryption_key'))", accountNumber),
			bankDetails.IFSCCode, bankDetails.BankName, bankDetails.BranchName,
			bankDetails.EffectiveFrom, bankDetails.Metadata, bankDetails.CreatedBy,
		).
		Suffix("RETURNING bank_id, created_at, version")

	var result domain.AgentBankDetails
	err := dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.AgentBankDetails], &result)
	if err != nil {
		return nil, err
	}

	// Query 2: Insert audit log for bank details creation
	// BR-AGT-PRF-005: Audit Logging
	query2 := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(bankDetails.AgentID, domain.AuditActionBankUpdate, "bank_details", "REDACTED", "New bank details added", bankDetails.CreatedBy, time.Now())

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
	result = bankDetails
	return &result, nil
}

// FindByID retrieves bank details by ID
func (r *AgentBankDetailsRepository) FindByID(ctx context.Context, bankID string) (*domain.AgentBankDetails, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentBankDetailsTable).
		Where(sq.Eq{"bank_id": bankID, "deleted_at": nil})

	var bankDetails domain.AgentBankDetails
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentBankDetails], &bankDetails)
	if err != nil {
		return nil, err
	}

	return &bankDetails, nil
}

// FindByAgentID retrieves all bank details for an agent
// FR-AGT-PRF-019: Bank Details Management
// BR-AGT-PRF-018: POSB Account/Bank Account Details
func (r *AgentBankDetailsRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentBankDetails, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentBankDetailsTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("effective_from DESC")

	var bankDetailsList []domain.AgentBankDetails
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentBankDetails], &bankDetailsList)
	if err != nil {
		return nil, err
	}

	return bankDetailsList, nil
}

// FindActiveByAgentID retrieves the most recent active bank details for an agent
// BR-AGT-PRF-018: Bank Account Details for Commission Disbursement
func (r *AgentBankDetailsRepository) FindActiveByAgentID(ctx context.Context, agentID string) (*domain.AgentBankDetails, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentBankDetailsTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		Where(sq.LtOrEq{"effective_from": time.Now()}).
		OrderBy("effective_from DESC").
		Limit(1)

	var bankDetails domain.AgentBankDetails
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentBankDetails], &bankDetails)
	if err != nil {
		return nil, err
	}

	return &bankDetails, nil
}

// GetDecryptedAccountNumber retrieves the decrypted account number
// BR-AGT-PRF-018: Bank Account Details - Decrypt for commission processing
// Note: This should only be called by authorized services
func (r *AgentBankDetailsRepository) GetDecryptedAccountNumber(ctx context.Context, bankID string) (string, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use pgcrypto to decrypt the account number
	query := dblib.Psql.Select("pgp_sym_decrypt(account_number_encrypted, current_setting('app.encryption_key')) as account_number").
		From(agentBankDetailsTable).
		Where(sq.Eq{"bank_id": bankID, "deleted_at": nil})

	var accountNumber string
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowTo[string], &accountNumber)
	if err != nil {
		return "", err
	}

	return accountNumber, nil
}

// Update updates agent bank details
// FR-AGT-PRF-019: Bank Details Management
// VR-AGT-PRF-015 to VR-AGT-PRF-017: Bank Validations
func (r *AgentBankDetailsRepository) Update(ctx context.Context, bankID string, updates map[string]interface{}, updatedBy, accountNumber string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for update + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Update bank details
	updateQuery := dblib.Psql.Update(agentBankDetailsTable).
		Set("updated_at", time.Now()).
		Set("updated_by", updatedBy).
		Where(sq.Eq{"bank_id": bankID, "deleted_at": nil})

	// Apply updates
	for field, value := range updates {
		updateQuery = updateQuery.Set(field, value)
	}

	// If account number is provided, encrypt it
	if accountNumber != "" {
		updateQuery = updateQuery.Set("account_number_encrypted",
			sq.Expr("pgp_sym_encrypt(?, current_setting('app.encryption_key'))", accountNumber))
	}

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentBankDetailsTable).
		Where(sq.Eq{"bank_id": bankID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log (without sensitive data)
	// BR-AGT-PRF-005: Audit Logging
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionBankUpdate, "bank_details", "REDACTED", "Bank details updated", updatedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// Delete soft deletes agent bank details
func (r *AgentBankDetailsRepository) Delete(ctx context.Context, bankID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use batch for delete + audit log
	// OPTIMIZATION: Batch combines UPDATE + INSERT audit
	batch := &pgx.Batch{}

	// Query 1: Soft delete bank details
	updateQuery := dblib.Psql.Update(agentBankDetailsTable).
		Set("deleted_at", time.Now()).
		Set("updated_by", deletedBy).
		Where(sq.Eq{"bank_id": bankID, "deleted_at": nil})

	err := dblib.QueueExecRow(batch, updateQuery)
	if err != nil {
		return err
	}

	// Query 2: Get agent_id for audit log
	var agentID string
	selectQuery := dblib.Psql.Select("agent_id").
		From(agentBankDetailsTable).
		Where(sq.Eq{"bank_id": bankID})

	err = dblib.QueueReturnRow(batch, selectQuery, pgx.RowTo[string], &agentID)
	if err != nil {
		return err
	}

	// Query 3: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "action_reason", "performed_by", "performed_at").
		Values(agentID, domain.AuditActionDelete, "bank_details", "Bank details deleted", deletedBy, time.Now())

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchCreate inserts multiple agent bank details in a single transaction
// OPTIMIZATION: Batch operation for multiple bank details inserts
// FR-AGT-PRF-019: Bank Details Management
func (r *AgentBankDetailsRepository) BatchCreate(ctx context.Context, bankDetailsList []domain.AgentBankDetails, accountNumbers []string) ([]domain.AgentBankDetails, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	if len(bankDetailsList) != len(accountNumbers) {
		return nil, pgx.ErrNoRows
	}

	// Use batch for multiple bank details inserts with audit logs
	// OPTIMIZATION: Batch operation combines multiple INSERTs in single round-trip
	batch := &pgx.Batch{}
	results := make([]domain.AgentBankDetails, len(bankDetailsList))

	for i, bankDetails := range bankDetailsList {
		// Insert bank details with encrypted account number
		insertQuery := dblib.Psql.Insert(agentBankDetailsTable).
			Columns(
				"agent_id", "account_type", "account_number_encrypted", "ifsc_code",
				"bank_name", "branch_name", "effective_from", "metadata", "created_by",
			).
			Values(
				bankDetails.AgentID, bankDetails.AccountType,
				sq.Expr("pgp_sym_encrypt(?, current_setting('app.encryption_key'))", accountNumbers[i]),
				bankDetails.IFSCCode, bankDetails.BankName, bankDetails.BranchName,
				bankDetails.EffectiveFrom, bankDetails.Metadata, bankDetails.CreatedBy,
			).
			Suffix("RETURNING bank_id, created_at, version")

		err := dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByNameLax[domain.AgentBankDetails], &results[i])
		if err != nil {
			return nil, err
		}

		// Insert audit log
		auditQuery := dblib.Psql.Insert("agent_audit_logs").
			Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
			Values(bankDetails.AgentID, domain.AuditActionBankUpdate, "bank_details", "REDACTED", "New bank details added", bankDetails.CreatedBy, time.Now())

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
	for i := range bankDetailsList {
		results[i] = bankDetailsList[i]
	}

	return results, nil
}

// ValidateIFSCCode validates IFSC code format and optionally fetches bank details
// VR-AGT-PRF-016: IFSC Code Format (AAAA0123456)
// Note: This is a placeholder for IFSC validation logic
// In production, this should integrate with bank master data or external API
func (r *AgentBankDetailsRepository) ValidateIFSCCode(ctx context.Context, ifscCode string) (bankName string, branchName string, isValid bool, err error) {
	// Placeholder implementation
	// In production, this should:
	// 1. Validate IFSC format: 4 alpha + 7 numeric (first digit 0)
	// 2. Query bank master table or external API
	// 3. Return bank name and branch name if valid
	return "", "", false, nil
}
