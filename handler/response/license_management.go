package response

import (
	"pli-agent-api/core/domain"
	"pli-agent-api/core/port"
)

// AgentLicensesResponse returns all licenses for an agent
// AGT-029: Get Agent Licenses
type AgentLicensesResponse struct {
	port.StatusCodeAndMessage
	AgentID  string                  `json:"agent_id"`
	Licenses []LicenseWithExpiryInfo `json:"licenses"`
}

// LicenseWithExpiryInfo includes license with expiry status
type LicenseWithExpiryInfo struct {
	LicenseID             string `json:"license_id"`
	LicenseNumber         string `json:"license_number"`
	LicenseLine           string `json:"license_line"`
	LicenseType           string `json:"license_type"`
	ResidentStatus        string `json:"resident_status"`
	LicenseDate           string `json:"license_date"`
	RenewalDate           string `json:"renewal_date"`
	AuthorityDate         string `json:"authority_date"`
	RenewalCount          int    `json:"renewal_count"`
	LicenseStatus         string `json:"license_status"`
	IsPrimary             bool   `json:"is_primary"`
	DaysUntilExpiry       int    `json:"days_until_expiry"`
	ExpiryStatus          string `json:"expiry_status"` // GREEN, YELLOW, RED
	LicentiateExamPassed  bool   `json:"licentiate_exam_passed"`
}

// AddLicenseResponse returns newly created license with reminders
// AGT-030: Add License
type AddLicenseResponse struct {
	port.StatusCodeAndMessage
	License             LicenseWithExpiryInfo   `json:"license"`
	RenewalCalculation  RenewalCalculationInfo  `json:"renewal_calculation"`
	RemindersScheduled  []ReminderInfo          `json:"reminders_scheduled"`
}

// RenewalCalculationInfo shows how renewal date was calculated
type RenewalCalculationInfo struct {
	LicenseType    string `json:"license_type"`
	RenewalPeriod  string `json:"renewal_period"`
	Reason         string `json:"reason"`
}

// LicenseDetailsResponse returns complete license details with history
// AGT-031: Get License Details
type LicenseDetailsResponse struct {
	port.StatusCodeAndMessage
	License          LicenseWithExpiryInfo `json:"license"`
	RenewalHistory   []RenewalHistoryEntry `json:"renewal_history"`
	ReminderSchedule []ReminderInfo        `json:"reminder_schedule"`
}

// RenewalHistoryEntry represents a license renewal event
type RenewalHistoryEntry struct {
	RenewedAt     string `json:"renewed_at"`
	RenewalType   string `json:"renewal_type"`
	PreviousDate  string `json:"previous_date"`
	NewDate       string `json:"new_date"`
	RenewedBy     string `json:"renewed_by"`
}

// UpdateLicenseResponse returns updated license
// AGT-032: Update License
type UpdateLicenseResponse struct {
	port.StatusCodeAndMessage
	License       LicenseWithExpiryInfo `json:"license"`
	UpdatedFields []string              `json:"updated_fields"`
}

// RenewLicenseResponse returns renewed license with calculation details
// AGT-033: Renew License
type RenewLicenseResponse struct {
	port.StatusCodeAndMessage
	License             LicenseWithExpiryInfo  `json:"license"`
	RenewalCalculation  RenewalCalculationInfo `json:"renewal_calculation"`
	PreviousRenewalDate string                 `json:"previous_renewal_date"`
	NewRenewalDate      string                 `json:"new_renewal_date"`
	ConvertedToPermanent bool                   `json:"converted_to_permanent"`
}

// LicenseTypesResponse returns available license types
// AGT-035: Get License Types
type LicenseTypesResponse struct {
	port.StatusCodeAndMessage
	LicenseTypes []LicenseTypeInfo `json:"license_types"`
}

// LicenseTypeInfo describes a license type
type LicenseTypeInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ExpiringLicensesResponse returns licenses expiring soon
// AGT-036: Get Expiring Licenses
type ExpiringLicensesResponse struct {
	port.StatusCodeAndMessage
	port.MetaDataResponse
	Licenses []ExpiringLicenseInfo `json:"licenses"`
}

// ExpiringLicenseInfo includes agent details for expiring license
type ExpiringLicenseInfo struct {
	LicenseID       string `json:"license_id"`
	AgentID         string `json:"agent_id"`
	AgentName       string `json:"agent_name"`
	LicenseNumber   string `json:"license_number"`
	LicenseType     string `json:"license_type"`
	RenewalDate     string `json:"renewal_date"`
	DaysUntilExpiry int    `json:"days_until_expiry"`
	AgentMobile     string `json:"agent_mobile"`
	AgentEmail      string `json:"agent_email"`
	OfficeCode      string `json:"office_code"`
}

// LicenseRemindersResponse returns reminder schedule
// AGT-037: Get License Reminders
type LicenseRemindersResponse struct {
	port.StatusCodeAndMessage
	LicenseID string         `json:"license_id"`
	Reminders []ReminderInfo `json:"reminders"`
}

// ReminderInfo represents a license reminder
type ReminderInfo struct {
	ReminderID    string `json:"reminder_id,omitempty"`
	ReminderType  string `json:"reminder_type"`  // 30_DAYS, 15_DAYS, 7_DAYS, EXPIRY_DAY
	ScheduledDate string `json:"scheduled_date"`
	Status        string `json:"status,omitempty"` // PENDING, SENT, FAILED
	SentDate      string `json:"sent_date,omitempty"`
	EmailSent     bool   `json:"email_sent,omitempty"`
	SMSSent       bool   `json:"sms_sent,omitempty"`
}

// ExpiryDeactivationResponse returns batch deactivation results
// AGT-038: Trigger License Expiry Deactivation
type ExpiryDeactivationResponse struct {
	port.StatusCodeAndMessage
	BatchID             string   `json:"batch_id"`
	BatchDate           string   `json:"batch_date"`
	AgentsDeactivated   int      `json:"agents_deactivated"`
	AgentIDs            []string `json:"agent_ids"`
	NotificationsSent   int      `json:"notifications_sent"`
	DryRun              bool     `json:"dry_run"`
}

// Helper function to convert domain license to response format
func ToLicenseWithExpiryInfo(license *domain.AgentLicense) LicenseWithExpiryInfo {
	return LicenseWithExpiryInfo{
		LicenseID:            license.LicenseID,
		LicenseNumber:        license.LicenseNumber,
		LicenseLine:          license.LicenseLine,
		LicenseType:          license.LicenseType,
		ResidentStatus:       license.ResidentStatus,
		LicenseDate:          license.LicenseDate.Format("2006-01-02"),
		RenewalDate:          license.RenewalDate.Format("2006-01-02"),
		AuthorityDate:        license.AuthorityDate.Format("2006-01-02"),
		RenewalCount:         license.RenewalCount,
		LicenseStatus:        license.LicenseStatus,
		IsPrimary:            license.IsPrimary,
		DaysUntilExpiry:      license.DaysUntilExpiry(),
		ExpiryStatus:         license.GetExpiryStatus(),
		LicentiateExamPassed: license.LicentiatExamPassed,
	}
}
