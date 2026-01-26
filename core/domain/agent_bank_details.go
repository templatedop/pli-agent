package domain

import (
	"database/sql"
	"time"
)

// AgentBankDetails represents agent bank details entity
// E-05: Agent Bank Details Entity
// BR-AGT-PRF-018
type AgentBankDetails struct {
	// Primary Key
	BankID string `json:"bank_id" db:"bank_id"`

	// Foreign Key
	AgentID string `json:"agent_id" db:"agent_id"`

	// Account Type (VR-AGT-PRF-017)
	AccountType string `json:"account_type" db:"account_type"` // SAVINGS, CURRENT

	// Bank Details (BR-AGT-PRF-018, VR-AGT-PRF-015, VR-AGT-PRF-016)
	// Encrypted account number using pgcrypto
	AccountNumberEncrypted []byte         `json:"-" db:"account_number_encrypted"` // Encrypted, not in JSON
	IFSCCode               string         `json:"ifsc_code" db:"ifsc_code"`        // VR-AGT-PRF-016: AAAA0123456
	BankName               string         `json:"bank_name" db:"bank_name"`
	BranchName             sql.NullString `json:"branch_name" db:"branch_name"`

	// Effective Date
	EffectiveFrom time.Time `json:"effective_from" db:"effective_from"`

	// Metadata
	Metadata sql.NullString `json:"metadata" db:"metadata"` // JSONB

	// Audit Fields
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at" db:"updated_at"`
	CreatedBy string         `json:"created_by" db:"created_by"`
	UpdatedBy sql.NullString `json:"updated_by" db:"updated_by"`
	DeletedAt sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	Version   int            `json:"version" db:"version"`
}

// AccountType constants (VR-AGT-PRF-017)
const (
	AccountTypeSavings = "SAVINGS"
	AccountTypeCurrent = "CURRENT"
)
