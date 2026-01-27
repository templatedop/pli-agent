package domain

import (
	"time"
)

// AgentDocument represents a document uploaded for agent processes
// E-09: Agent Document Entity
type AgentDocument struct {
	DocumentID    string     `db:"document_id"`
	AgentID       string     `db:"agent_id"`
	ReferenceType string     `db:"reference_type"`
	ReferenceID   string     `db:"reference_id"`
	DocumentType  string     `db:"document_type"`
	FileName      string     `db:"file_name"`
	FileSize      int64      `db:"file_size"`
	FilePath      string     `db:"file_path"`
	MimeType      string     `db:"mime_type"`
	UploadedBy    string     `db:"uploaded_by"`
	UploadedAt    time.Time  `db:"uploaded_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
	Version       int        `db:"version"`
}

// Reference Type Constants
const (
	ReferenceTypeReinstatement = "REINSTATEMENT"
	ReferenceTypeTermination   = "TERMINATION"
)

// Document Type Constants
const (
	DocumentTypeLicenseRenewal       = "LICENSE_RENEWAL"
	DocumentTypeClearanceCertificate = "CLEARANCE_CERTIFICATE"
	DocumentTypeAppealApproval       = "APPEAL_APPROVAL"
	DocumentTypeOther                = "OTHER"
)

// Allowed MIME Types
const (
	MimeTypePDF  = "application/pdf"
	MimeTypeJPG  = "image/jpeg"
	MimeTypePNG  = "image/png"
)

// Max file size: 5MB
const MaxDocumentSize = 5 * 1024 * 1024 // 5MB in bytes

// IsDeleted checks if document is soft deleted
func (d *AgentDocument) IsDeleted() bool {
	return d.DeletedAt != nil
}

// GetFileSizeMB returns file size in megabytes
func (d *AgentDocument) GetFileSizeMB() float64 {
	return float64(d.FileSize) / (1024 * 1024)
}

// IsValidMimeType checks if mime type is allowed
func (d *AgentDocument) IsValidMimeType() bool {
	switch d.MimeType {
	case MimeTypePDF, MimeTypeJPG, MimeTypePNG:
		return true
	default:
		return false
	}
}

// IsValidSize checks if file size is within limit
func (d *AgentDocument) IsValidSize() bool {
	return d.FileSize > 0 && d.FileSize <= MaxDocumentSize
}

// GetDocumentTypeName returns human-readable document type
func (d *AgentDocument) GetDocumentTypeName() string {
	switch d.DocumentType {
	case DocumentTypeLicenseRenewal:
		return "License Renewal"
	case DocumentTypeClearanceCertificate:
		return "Clearance Certificate"
	case DocumentTypeAppealApproval:
		return "Appeal Approval"
	case DocumentTypeOther:
		return "Other"
	default:
		return "Unknown"
	}
}

// GetReferenceTypeName returns human-readable reference type
func (d *AgentDocument) GetReferenceTypeName() string {
	switch d.ReferenceType {
	case ReferenceTypeReinstatement:
		return "Reinstatement"
	case ReferenceTypeTermination:
		return "Termination"
	default:
		return "Unknown"
	}
}
