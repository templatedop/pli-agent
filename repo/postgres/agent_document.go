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
)

const (
	agentDocumentTable = "agent_documents"
)

// AgentDocumentRepository handles all database operations for agent documents
// E-09: Agent Document Entity
type AgentDocumentRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAgentDocumentRepository creates a new agent document repository
func NewAgentDocumentRepository(db *dblib.DB, cfg *config.Config) *AgentDocumentRepository {
	return &AgentDocumentRepository{
		db:  db,
		cfg: cfg,
	}
}

// Create inserts a new document record with audit logging
// AGT-063: Upload Reinstatement Documents
// FR-AGT-PRF-013: Reinstatement Process
// BR-AGT-PRF-016: Status updates require mandatory reason
// FR-AGT-PRF-022: Audit History Tracking
// CRITICAL: Single database round trip with batch (INSERT document + INSERT audit)
func (r *AgentDocumentRepository) Create(ctx context.Context, document domain.AgentDocument) (*domain.AgentDocument, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Validate file size (max 5MB)
	if !document.IsValidSize() {
		return nil, fmt.Errorf("file size must be between 1 byte and 5MB (current: %.2fMB)", document.GetFileSizeMB())
	}

	// Validate MIME type (PDF, JPG, PNG)
	if !document.IsValidMimeType() {
		return nil, fmt.Errorf("invalid file type: %s (allowed: PDF, JPG, PNG)", document.MimeType)
	}

	// Use batch to combine INSERT document + INSERT audit log in single round trip
	batch := &pgx.Batch{}

	// Query 1: Insert document record
	insertQuery := dblib.Psql.Insert(agentDocumentTable).
		Columns(
			"agent_id", "reference_type", "reference_id", "document_type",
			"file_name", "file_size", "file_path", "mime_type", "uploaded_by",
		).
		Values(
			document.AgentID,
			document.ReferenceType,
			document.ReferenceID,
			document.DocumentType,
			document.FileName,
			document.FileSize,
			document.FilePath,
			document.MimeType,
			document.UploadedBy,
		).
		Suffix("RETURNING *")

	var result domain.AgentDocument
	err := dblib.QueueReturnRow(batch, insertQuery, pgx.RowToStructByNameLax[domain.AgentDocument], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to queue document insert: %w", err)
	}

	// Query 2: Insert audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			document.AgentID,
			"DOCUMENT_UPLOADED",
			"document_type",
			document.DocumentType,
			fmt.Sprintf("Document uploaded for %s: %s", document.ReferenceType, document.FileName),
			document.UploadedBy,
			time.Now(),
		)

	err = dblib.QueueExecRow(batch, auditQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to queue audit log: %w", err)
	}

	// Execute batch
	err = r.db.SendBatch(cCtx, batch).Close()
	if err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	return &result, nil
}

// FindByID retrieves a document by ID
// AGT-063: Upload Reinstatement Documents (查询)
func (r *AgentDocumentRepository) FindByID(ctx context.Context, documentID string) (*domain.AgentDocument, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentDocumentTable).
		Where(sq.Eq{"document_id": documentID, "deleted_at": nil})

	var document domain.AgentDocument
	err := dblib.SelectOne(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentDocument], &document)
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	return &document, nil
}

// FindByReferenceID retrieves all documents for a specific reference
// AGT-063: Upload Reinstatement Documents (查询)
// FR-AGT-PRF-013: Reinstatement Process
func (r *AgentDocumentRepository) FindByReferenceID(
	ctx context.Context,
	referenceType string,
	referenceID string,
) ([]domain.AgentDocument, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentDocumentTable).
		Where(sq.Eq{
			"reference_type": referenceType,
			"reference_id":   referenceID,
			"deleted_at":     nil,
		}).
		OrderBy("uploaded_at DESC")

	var documents []domain.AgentDocument
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentDocument], &documents)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}

	return documents, nil
}

// FindByAgentID retrieves all documents for an agent
// AGT-063: Upload Reinstatement Documents (查询)
func (r *AgentDocumentRepository) FindByAgentID(ctx context.Context, agentID string) ([]domain.AgentDocument, error) {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(agentDocumentTable).
		Where(sq.Eq{"agent_id": agentID, "deleted_at": nil}).
		OrderBy("uploaded_at DESC")

	var documents []domain.AgentDocument
	err := dblib.SelectRows(cCtx, r.db, query, pgx.RowToStructByNameLax[domain.AgentDocument], &documents)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}

	return documents, nil
}

// Delete soft deletes a document
// FR-AGT-PRF-013: Reinstatement Process
// FR-AGT-PRF-022: Audit History Tracking
// CRITICAL: Single database round trip with CTE (UPDATE + INSERT audit)
func (r *AgentDocumentRepository) Delete(ctx context.Context, documentID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use CTE to combine UPDATE + INSERT audit log in single round trip
	// WITH deleted AS (UPDATE... RETURNING ...) INSERT INTO audit_logs SELECT ... FROM deleted
	sql := `
		WITH deleted AS (
			UPDATE agent_documents
			SET deleted_at = NOW()
			WHERE document_id = $1 AND deleted_at IS NULL
			RETURNING agent_id, file_name
		)
		INSERT INTO agent_audit_logs (agent_id, action_type, field_name, new_value, action_reason, performed_by, performed_at)
		SELECT
			agent_id,
			'DOCUMENT_DELETED',
			'document_id',
			$1,
			'Document deleted: ' || file_name,
			$2,
			NOW()
		FROM deleted
	`

	tag, err := r.db.Exec(cCtx, sql, documentID, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("document not found: %s", documentID)
	}

	return nil
}
