package port

import "io"

// Standard status messages for all operations
var (
	ListSuccess   StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "list retrieved successfully", Success: true}
	FetchSuccess  StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "data retrieved successfully", Success: true}
	CreateSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 201, Message: "resource created successfully", Success: true}
	UpdateSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "resource updated successfully", Success: true}
	DeleteSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "resource deleted successfully", Success: true}
	CustomEnv     StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "This is environment specific", Success: true}
)

// Agent-specific status messages
var (
	ProfileCreateSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 201, Message: "agent profile created successfully", Success: true}
	ProfileUpdateSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "agent profile updated successfully", Success: true}
	ProfileFetchSuccess  StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "profile retrieved successfully", Success: true}
	StatusChangeSuccess  StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "agent status updated successfully", Success: true}
	ValidationSuccess    StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "validation successful", Success: true}
	SessionCreateSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 201, Message: "session created successfully", Success: true}
	OTPSendSuccess       StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "OTP sent successfully", Success: true}
	OTPVerifySuccess     StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "OTP verified successfully", Success: true}
	SearchSuccess        StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "search completed successfully", Success: true}
	FormFetchSuccess     StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "form data retrieved successfully", Success: true}
	PendingApproval      StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 202, Message: "update pending approval", Success: true}
	ApprovalSuccess      StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "update approved successfully", Success: true}
	RejectionSuccess     StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "update rejected successfully", Success: true}
	AuditHistorySuccess  StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "audit history retrieved successfully", Success: true}

	// License Management (Phase 7: AGT-029 to AGT-038)
	LicenseListSuccess            StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "licenses retrieved successfully", Success: true}
	LicenseCreateSuccess          StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 201, Message: "license created successfully", Success: true}
	LicenseDetailSuccess          StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "license details retrieved successfully", Success: true}
	LicenseUpdateSuccess          StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "license updated successfully", Success: true}
	LicenseRenewSuccess           StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "license renewed successfully", Success: true}
	LicenseDeleteSuccess          StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "license deleted successfully", Success: true}
	LicenseTypeListSuccess        StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "license types retrieved successfully", Success: true}
	LicenseExpiringListSuccess    StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "expiring licenses retrieved successfully", Success: true}
	LicenseReminderSuccess        StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "license reminders retrieved successfully", Success: true}
	LicenseBatchDeactivateSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "batch deactivation completed successfully", Success: true}

	// Status Management (Phase 8: AGT-039 to AGT-041)
	AgentTerminationSuccess      StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 201, Message: "agent termination initiated successfully", Success: true}
	TerminationLetterRetrieved   StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "termination letter retrieved successfully", Success: true}
	ReinstatementRequestCreated  StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 201, Message: "reinstatement request created successfully", Success: true}
	ReinstatementApprovalSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "reinstatement approved successfully", Success: true}
	ReinstatementRejectionSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "reinstatement rejected successfully", Success: true}
)

// StatusCodeAndMessage is embedded in all response structs
// Provides consistent status code, success flag, and message
type StatusCodeAndMessage struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

// Status returns HTTP status code (interface compliance)
func (s StatusCodeAndMessage) Status() int {
	return s.StatusCode
}

func (s StatusCodeAndMessage) ResponseType() string {
	return "standard"
}

func (s StatusCodeAndMessage) GetContentType() string {
	return "application/json"
}

func (s StatusCodeAndMessage) GetContentDisposition() string {
	return ""
}

func (s StatusCodeAndMessage) Object() []byte {
	return nil
}

// FileResponse for file downloads/uploads
type FileResponse struct {
	ContentDisposition string
	ContentType        string
	Data               []byte        // Memory-based payload
	Reader             io.ReadCloser // Optional streaming source
}

func (s FileResponse) GetContentType() string {
	return s.ContentType
}

func (s FileResponse) GetContentDisposition() string {
	return s.ContentDisposition
}

func (s FileResponse) ResponseType() string {
	return "file"
}

func (s FileResponse) Status() int {
	return 200
}

func (s FileResponse) Object() []byte {
	return s.Data
}

// Stream copies Reader to w if available; else writes Data
func (s FileResponse) Stream(w io.Writer) error {
	if s.Reader == nil {
		if len(s.Data) > 0 {
			_, err := w.Write(s.Data)
			return err
		}
		return nil
	}
	defer s.Reader.Close()
	_, err := io.Copy(w, s.Reader)
	return err
}

// MetaDataResponse provides pagination metadata
// Embed this in list response structs
// FR-AGT-PRF-021: Multi-Criteria Agent Search Support
type MetaDataResponse struct {
	Skip                 uint64 `json:"skip,default=0"`
	Limit                uint64 `json:"limit,default=10"`
	OrderBy              string `json:"order_by,omitempty"`
	SortType             string `json:"sort_type,omitempty"`
	TotalRecordsCount    int    `json:"total_records_count,omitempty"`
	ReturnedRecordsCount uint64 `json:"returned_records_count"`
}

// NewMetaDataResponse creates metadata response from pagination params
func NewMetaDataResponse(skip, limit, total uint64) MetaDataResponse {
	return MetaDataResponse{
		Skip:                 skip,
		Limit:                limit,
		TotalRecordsCount:    int(total),
		ReturnedRecordsCount: limit,
	}
}
