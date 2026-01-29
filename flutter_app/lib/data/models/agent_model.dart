/// Data Model: Agent Profile
///
/// This model represents the API response/request structure for agent profiles.
/// It maps to the AgentProfile domain entity.
///
/// **Generated Files**:
/// - agent_model.freezed.dart (immutability)
/// - agent_model.g.dart (JSON serialization)
///
/// **Usage**:
/// ```dart
/// // From JSON (API response)
/// final model = AgentModel.fromJson(jsonData);
///
/// // To Entity (for domain layer)
/// final entity = model.toEntity();
///
/// // To JSON (for API request)
/// final json = model.toJson();
/// ```

import 'package:freezed_annotation/freezed_annotation.dart';
import '../../domain/entities/agent_profile.dart';

part 'agent_model.freezed.dart';
part 'agent_model.g.dart';

/// Agent Profile Data Model
@freezed
class AgentModel with _$AgentModel {
  const AgentModel._();

  const factory AgentModel({
    // ========================================================================
    // CORE IDENTIFIERS
    // ========================================================================

    /// Unique agent identifier (e.g., "AGT-2026-000001")
    @JsonKey(name: 'agent_id') required String agentId,

    /// Internal profile ID
    @JsonKey(name: 'profile_id') required String profileId,

    /// Agent type (advisor, advisor_coordinator, etc.)
    @JsonKey(name: 'agent_type') required String agentType,

    /// Full name
    @JsonKey(name: 'full_name') required String fullName,

    /// Current status (active, inactive, suspended, etc.)
    required String status,

    // ========================================================================
    // PERSONAL INFORMATION
    // ========================================================================

    /// PAN number (10 characters)
    @JsonKey(name: 'pan_number') String? panNumber,

    /// Personal information
    @JsonKey(name: 'personal_info') PersonalInfoModel? personalInfo,

    // ========================================================================
    // CONTACT & ADDRESS
    // ========================================================================

    /// List of addresses
    List<AddressModel>? addresses,

    /// List of contacts
    List<ContactModel>? contacts,

    /// List of emails
    List<EmailModel>? emails,

    // ========================================================================
    // OFFICE & HIERARCHY
    // ========================================================================

    /// Office information
    OfficeModel? office,

    /// Advisor coordinator (for advisors only)
    @JsonKey(name: 'advisor_coordinator') CoordinatorModel? advisorCoordinator,

    // ========================================================================
    // LICENSE & BANK
    // ========================================================================

    /// List of licenses
    List<LicenseModel>? licenses,

    /// Bank account details
    @JsonKey(name: 'bank_details') BankDetailsModel? bankDetails,

    // ========================================================================
    // METADATA
    // ========================================================================

    /// Date of joining
    @JsonKey(name: 'date_of_joining') String? dateOfJoining,

    /// Creation timestamp
    @JsonKey(name: 'created_at') String? createdAt,

    /// Last update timestamp
    @JsonKey(name: 'updated_at') String? updatedAt,

    /// Created by user ID
    @JsonKey(name: 'created_by') String? createdBy,

    /// Last updated by user ID
    @JsonKey(name: 'updated_by') String? updatedBy,
  }) = _AgentModel;

  /// Create from JSON
  factory AgentModel.fromJson(Map<String, dynamic> json) =>
      _$AgentModelFromJson(json);

  /// Convert to Domain Entity
  AgentProfile toEntity() {
    return AgentProfile(
      agentId: agentId,
      profileId: profileId,
      agentType: _parseAgentType(agentType),
      fullName: fullName,
      status: _parseAgentStatus(status),
      panNumber: panNumber,
      personalInfo: personalInfo?.toEntity(),
      addresses: addresses?.map((a) => a.toEntity()).toList(),
      contacts: contacts?.map((c) => c.toEntity()).toList(),
      emails: emails?.map((e) => e.toEntity()).toList(),
      office: office?.toEntity(),
      advisorCoordinator: advisorCoordinator?.toEntity(),
      licenses: licenses?.map((l) => l.toEntity()).toList(),
      bankDetails: bankDetails?.toEntity(),
      dateOfJoining: dateOfJoining != null ? DateTime.parse(dateOfJoining!) : null,
      createdAt: createdAt != null ? DateTime.parse(createdAt!) : null,
      updatedAt: updatedAt != null ? DateTime.parse(updatedAt!) : null,
      createdBy: createdBy,
      updatedBy: updatedBy,
    );
  }

  /// Parse agent type from string
  static AgentType _parseAgentType(String type) {
    switch (type.toLowerCase()) {
      case 'advisor':
        return AgentType.advisor;
      case 'advisor_coordinator':
        return AgentType.advisorCoordinator;
      case 'departmental_employee':
        return AgentType.departmentalEmployee;
      case 'field_officer':
        return AgentType.fieldOfficer;
      case 'direct_agent':
        return AgentType.directAgent;
      case 'gds':
        return AgentType.gds;
      default:
        return AgentType.advisor;
    }
  }

  /// Parse agent status from string
  static AgentStatus _parseAgentStatus(String status) {
    switch (status.toLowerCase()) {
      case 'active':
        return AgentStatus.active;
      case 'inactive':
        return AgentStatus.inactive;
      case 'suspended':
        return AgentStatus.suspended;
      case 'terminated':
        return AgentStatus.terminated;
      case 'deactivated':
        return AgentStatus.deactivated;
      default:
        return AgentStatus.inactive;
    }
  }
}

// ============================================================================
// NESTED MODELS
// ============================================================================

/// Personal Information Model
@freezed
class PersonalInfoModel with _$PersonalInfoModel {
  const PersonalInfoModel._();

  const factory PersonalInfoModel({
    @JsonKey(name: 'date_of_birth') String? dateOfBirth,
    String? gender,
    @JsonKey(name: 'marital_status') String? maritalStatus,
    @JsonKey(name: 'father_name') String? fatherName,
    @JsonKey(name: 'mother_name') String? motherName,
    @JsonKey(name: 'spouse_name') String? spouseName,
    String? nationality,
    String? religion,
    String? category,
    @JsonKey(name: 'aadhar_number') String? aadharNumber,
    @JsonKey(name: 'employee_id') String? employeeId,
  }) = _PersonalInfoModel;

  factory PersonalInfoModel.fromJson(Map<String, dynamic> json) =>
      _$PersonalInfoModelFromJson(json);

  PersonalInfo toEntity() {
    return PersonalInfo(
      dateOfBirth: dateOfBirth != null ? DateTime.parse(dateOfBirth!) : null,
      gender: gender,
      maritalStatus: maritalStatus,
      fatherName: fatherName,
      motherName: motherName,
      spouseName: spouseName,
      nationality: nationality,
      religion: religion,
      category: category,
      aadharNumber: aadharNumber,
      employeeId: employeeId,
    );
  }
}

/// Address Model
@freezed
class AddressModel with _$AddressModel {
  const AddressModel._();

  const factory AddressModel({
    @JsonKey(name: 'address_type') required String addressType,
    @JsonKey(name: 'address_line_1') required String addressLine1,
    @JsonKey(name: 'address_line_2') String? addressLine2,
    String? landmark,
    required String city,
    required String district,
    required String state,
    @JsonKey(name: 'pin_code') required String pinCode,
    required String country,
    @JsonKey(name: 'is_primary') @Default(false) bool isPrimary,
  }) = _AddressModel;

  factory AddressModel.fromJson(Map<String, dynamic> json) =>
      _$AddressModelFromJson(json);

  Address toEntity() {
    return Address(
      addressType: _parseAddressType(addressType),
      addressLine1: addressLine1,
      addressLine2: addressLine2,
      landmark: landmark,
      city: city,
      district: district,
      state: state,
      pinCode: pinCode,
      country: country,
      isPrimary: isPrimary,
    );
  }

  static AddressType _parseAddressType(String type) {
    switch (type.toLowerCase()) {
      case 'permanent':
        return AddressType.permanent;
      case 'correspondence':
        return AddressType.correspondence;
      case 'office':
        return AddressType.office;
      default:
        return AddressType.permanent;
    }
  }
}

/// Contact Model
@freezed
class ContactModel with _$ContactModel {
  const ContactModel._();

  const factory ContactModel({
    @JsonKey(name: 'contact_type') required String contactType,
    @JsonKey(name: 'contact_number') required String contactNumber,
    @JsonKey(name: 'is_primary') @Default(false) bool isPrimary,
    @JsonKey(name: 'is_verified') @Default(false) bool isVerified,
  }) = _ContactModel;

  factory ContactModel.fromJson(Map<String, dynamic> json) =>
      _$ContactModelFromJson(json);

  Contact toEntity() {
    return Contact(
      contactType: _parseContactType(contactType),
      contactNumber: contactNumber,
      isPrimary: isPrimary,
      isVerified: isVerified,
    );
  }

  static ContactType _parseContactType(String type) {
    switch (type.toLowerCase()) {
      case 'mobile':
        return ContactType.mobile;
      case 'landline':
        return ContactType.landline;
      case 'whatsapp':
        return ContactType.whatsapp;
      default:
        return ContactType.mobile;
    }
  }
}

/// Email Model
@freezed
class EmailModel with _$EmailModel {
  const EmailModel._();

  const factory EmailModel({
    @JsonKey(name: 'email_type') required String emailType,
    @JsonKey(name: 'email_address') required String emailAddress,
    @JsonKey(name: 'is_primary') @Default(false) bool isPrimary,
    @JsonKey(name: 'is_verified') @Default(false) bool isVerified,
  }) = _EmailModel;

  factory EmailModel.fromJson(Map<String, dynamic> json) =>
      _$EmailModelFromJson(json);

  Email toEntity() {
    return Email(
      emailType: _parseEmailType(emailType),
      emailAddress: emailAddress,
      isPrimary: isPrimary,
      isVerified: isVerified,
    );
  }

  static EmailType _parseEmailType(String type) {
    switch (type.toLowerCase()) {
      case 'personal':
        return EmailType.personal;
      case 'official':
        return EmailType.official;
      default:
        return EmailType.personal;
    }
  }
}

/// Office Model
@freezed
class OfficeModel with _$OfficeModel {
  const OfficeModel._();

  const factory OfficeModel({
    @JsonKey(name: 'office_code') required String officeCode,
    @JsonKey(name: 'office_name') required String officeName,
    @JsonKey(name: 'office_type') String? officeType,
    String? division,
    String? region,
    String? circle,
  }) = _OfficeModel;

  factory OfficeModel.fromJson(Map<String, dynamic> json) =>
      _$OfficeModelFromJson(json);

  Office toEntity() {
    return Office(
      officeCode: officeCode,
      officeName: officeName,
      officeType: officeType,
      division: division,
      region: region,
      circle: circle,
    );
  }
}

/// Coordinator Model
@freezed
class CoordinatorModel with _$CoordinatorModel {
  const CoordinatorModel._();

  const factory CoordinatorModel({
    @JsonKey(name: 'coordinator_id') required String coordinatorId,
    @JsonKey(name: 'coordinator_name') required String coordinatorName,
    @JsonKey(name: 'linked_since') String? linkedSince,
    @JsonKey(name: 'is_active') @Default(true) bool isActive,
  }) = _CoordinatorModel;

  factory CoordinatorModel.fromJson(Map<String, dynamic> json) =>
      _$CoordinatorModelFromJson(json);

  Coordinator toEntity() {
    return Coordinator(
      coordinatorId: coordinatorId,
      coordinatorName: coordinatorName,
      linkedSince: linkedSince != null ? DateTime.parse(linkedSince!) : null,
      isActive: isActive,
    );
  }
}

/// License Model
@freezed
class LicenseModel with _$LicenseModel {
  const LicenseModel._();

  const factory LicenseModel({
    @JsonKey(name: 'license_number') required String licenseNumber,
    @JsonKey(name: 'license_type') required String licenseType,
    @JsonKey(name: 'issue_date') String? issueDate,
    @JsonKey(name: 'expiry_date') String? expiryDate,
    @JsonKey(name: 'issuing_authority') String? issuingAuthority,
    @JsonKey(name: 'is_valid') @Default(true) bool isValid,
  }) = _LicenseModel;

  factory LicenseModel.fromJson(Map<String, dynamic> json) =>
      _$LicenseModelFromJson(json);

  License toEntity() {
    return License(
      licenseNumber: licenseNumber,
      licenseType: licenseType,
      issueDate: issueDate != null ? DateTime.parse(issueDate!) : null,
      expiryDate: expiryDate != null ? DateTime.parse(expiryDate!) : null,
      issuingAuthority: issuingAuthority,
      isValid: isValid,
    );
  }
}

/// Bank Details Model
@freezed
class BankDetailsModel with _$BankDetailsModel {
  const BankDetailsModel._();

  const factory BankDetailsModel({
    @JsonKey(name: 'account_number') required String accountNumber,
    @JsonKey(name: 'account_holder_name') required String accountHolderName,
    @JsonKey(name: 'ifsc_code') required String ifscCode,
    @JsonKey(name: 'bank_name') String? bankName,
    @JsonKey(name: 'branch_name') String? branchName,
    @JsonKey(name: 'account_type') String? accountType,
    @JsonKey(name: 'is_verified') @Default(false) bool isVerified,
  }) = _BankDetailsModel;

  factory BankDetailsModel.fromJson(Map<String, dynamic> json) =>
      _$BankDetailsModelFromJson(json);

  BankDetails toEntity() {
    return BankDetails(
      accountNumber: accountNumber,
      accountHolderName: accountHolderName,
      ifscCode: ifscCode,
      bankName: bankName,
      branchName: branchName,
      accountType: accountType,
      isVerified: isVerified,
    );
  }
}
