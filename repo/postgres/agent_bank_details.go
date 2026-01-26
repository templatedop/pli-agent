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
	dbutil "pli-agent-api/db"
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

	// Use CTE to combine INSERT with encryption + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-018: Bank Account Details - Account number stored encrypted using pgcrypto
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_bank_details (
				agent_id, account_type, account_number_encrypted, ifsc_code,
				bank_name, branch_name, effective_from, metadata, created_by
			) VALUES (
				$1, $2, pgp_sym_encrypt($3, current_setting('app.encryption_key')),
				$4, $5, $6, $7, $8, $9
			)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $10, $11, $12, $13, $14, $15
		FROM inserted
		RETURNING (SELECT ROW(bank_id, agent_id, account_type, ifsc_code, bank_name, branch_name,
			effective_from, metadata, created_at, updated_at, created_by, updated_by, deleted_at, version) FROM inserted)
	`

	args := []interface{}{
		bankDetails.AgentID, bankDetails.AccountType, accountNumber, bankDetails.IFSCCode,
		bankDetails.BankName, bankDetails.BranchName, bankDetails.EffectiveFrom,
		bankDetails.Metadata, bankDetails.CreatedBy,
		domain.AuditActionBankUpdate, "bank_details", "REDACTED", "New bank details added", bankDetails.CreatedBy, time.Now(),
	}

	var result domain.AgentBankDetails
	err := dbutil.QueueReturnRowRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentBankDetails], &result)
	if err != nil {
		return nil, err
	}

	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

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

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	// BR-AGT-PRF-005: Audit Logging
	batch := &pgx.Batch{}

	// Build SET clause dynamically
	setClauses := "updated_at = $2, updated_by = $3"
	args := []interface{}{bankID, time.Now(), updatedBy}
	argIndex := 4

	for field, value := range updates {
		setClauses += fmt.Sprintf(", %s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	// If account number is provided, encrypt it
	if accountNumber != "" {
		setClauses += fmt.Sprintf(", account_number_encrypted = pgp_sym_encrypt($%d, current_setting('app.encryption_key'))", argIndex)
		args = append(args, accountNumber)
		argIndex++
	}

	sql := fmt.Sprintf(`
		WITH updated AS (
			UPDATE agent_bank_details
			SET %s
			WHERE bank_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $%d, $%d, $%d, $%d, $%d, $%d
		FROM updated
	`, setClauses, argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4, argIndex+5)

	args = append(args, domain.AuditActionBankUpdate, "bank_details", "REDACTED", "Bank details updated", updatedBy, time.Now())

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
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

	// Use CTE to combine UPDATE + INSERT audit in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must combine at SQL level
	batch := &pgx.Batch{}

	sql := `
		WITH updated AS (
			UPDATE agent_bank_details
			SET deleted_at = $2, updated_by = $3
			WHERE bank_id = $1 AND deleted_at IS NULL
			RETURNING agent_id
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, action_reason, performed_by, performed_at)
		SELECT agent_id, $4, $5, $6, $7, $8
		FROM updated
	`

	args := []interface{}{
		bankID, time.Now(), deletedBy,
		domain.AuditActionDelete, "bank_details", "Bank details deleted", deletedBy, time.Now(),
	}

	err := dbutil.QueueExecRowRaw(batch, sql, args...)
	if err != nil {
		return err
	}

	// Execute batch
	return r.db.SendBatch(cCtx, batch).Close()
}

// BatchCreate inserts multiple agent bank details in a single transaction
// OPTIMIZATION: Batch operation for multiple bank details inserts using UNNEST
// FR-AGT-PRF-019: Bank Details Management
func (r *AgentBankDetailsRepository) BatchCreate(ctx context.Context, bankDetailsList []domain.AgentBankDetails, accountNumbers []string) ([]domain.AgentBankDetails, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	if len(bankDetailsList) != len(accountNumbers) {
		return nil, pgx.ErrNoRows
	}

	// Use UNNEST to bulk insert bank details with audit logs in single query
	// CRITICAL: Golang variables cannot be passed between batch queries - must use UNNEST pattern
	// BR-AGT-PRF-005: Audit Logging
	// BR-AGT-PRF-018: Account number encrypted using pgcrypto
	batch := &pgx.Batch{}

	// Prepare arrays for UNNEST
	agentIDs := make([]string, len(bankDetailsList))
	accountTypes := make([]string, len(bankDetailsList))
	ifscCodes := make([]string, len(bankDetailsList))
	bankNames := make([]string, len(bankDetailsList))
	branchNames := make([]string, len(bankDetailsList))
	effectiveFroms := make([]time.Time, len(bankDetailsList))
	metadatas := make([]interface{}, len(bankDetailsList))
	createdBys := make([]string, len(bankDetailsList))

	for i, bankDetails := range bankDetailsList {
		agentIDs[i] = bankDetails.AgentID
		accountTypes[i] = bankDetails.AccountType
		ifscCodes[i] = bankDetails.IFSCCode
		bankNames[i] = bankDetails.BankName
		branchNames[i] = bankDetails.BranchName
		effectiveFroms[i] = bankDetails.EffectiveFrom
		metadatas[i] = bankDetails.Metadata
		createdBys[i] = bankDetails.CreatedBy
	}

	sql := `
		WITH inserted AS (
			INSERT INTO agent_bank_details (
				agent_id, account_type, account_number_encrypted, ifsc_code,
				bank_name, branch_name, effective_from, metadata, created_by
			)
			SELECT
				agent_id, account_type,
				pgp_sym_encrypt(account_number, current_setting('app.encryption_key')),
				ifsc_code, bank_name, branch_name, effective_from, metadata, created_by
			FROM UNNEST(
				$1::uuid[],
				$2::text[],
				$3::text[],
				$4::text[],
				$5::text[],
				$6::text[],
				$7::timestamp[],
				$8::jsonb[],
				$9::text[]
			) AS t(agent_id, account_type, account_number, ifsc_code, bank_name, branch_name, effective_from, metadata, created_by)
			RETURNING *
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT agent_id, $10, $11, $12, $13, created_by, NOW()
		FROM inserted
		RETURNING (SELECT array_agg(ROW(bank_id, agent_id, account_type, ifsc_code, bank_name, branch_name,
			effective_from, metadata, created_at, created_by, updated_at, updated_by, deleted_at, version)::agent_bank_details) FROM inserted)
	`

	args := []interface{}{
		agentIDs, accountTypes, accountNumbers, ifscCodes, bankNames, branchNames, effectiveFroms, metadatas, createdBys,
		domain.AuditActionBankUpdate, "bank_details", "REDACTED", "New bank details added",
	}

	var results []domain.AgentBankDetails
	err := dbutil.QueueReturnRaw(batch, sql, args, pgx.RowToStructByNameLax[domain.AgentBankDetails], &results)
	if err != nil {
		return nil, err
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, err
	}

	// Copy input data to results if needed
	for i := range bankDetailsList {
		if i < len(results) {
			results[i].AccountType = bankDetailsList[i].AccountType
			results[i].IFSCCode = bankDetailsList[i].IFSCCode
			results[i].BankName = bankDetailsList[i].BankName
		}
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
