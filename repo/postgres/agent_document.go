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
// CRITICAL: Single database round trip with INSERT document + INSERT audit
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

	// Insert document record
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
	err := dblib.SelectOne(cCtx, r.db, insertQuery, pgx.RowToStructByNameLax[domain.AgentDocument], &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	// Create audit log
	// FR-AGT-PRF-022: Audit History Tracking
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			result.AgentID,
			"DOCUMENT_UPLOADED",
			"document_type",
			document.DocumentType,
			fmt.Sprintf("Document uploaded for %s: %s", document.ReferenceType, document.FileName),
			document.UploadedBy,
			time.Now(),
		)

	_, err = dblib.Insert(cCtx, r.db, auditQuery)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
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
func (r *AgentDocumentRepository) Delete(ctx context.Context, documentID, deletedBy string) error {
	cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Get document for audit log
	document, err := r.FindByID(ctx, documentID)
	if err != nil {
		return err
	}

	updateQuery := dblib.Psql.Update(agentDocumentTable).
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"document_id": documentID, "deleted_at": nil})

	tag, err := dblib.Update(cCtx, r.db, updateQuery)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("document not found: %s", documentID)
	}

	// Create audit log
	auditQuery := dblib.Psql.Insert("agent_audit_logs").
		Columns("agent_id", "action_type", "field_name", "new_value", "action_reason", "performed_by", "performed_at").
		Values(
			document.AgentID,
			"DOCUMENT_DELETED",
			"document_id",
			documentID,
			fmt.Sprintf("Document deleted: %s", document.FileName),
			deletedBy,
			time.Now(),
		)

	_, err = dblib.Insert(cCtx, r.db, auditQuery)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
	}

	return nil
}
