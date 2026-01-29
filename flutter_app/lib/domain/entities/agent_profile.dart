/// Domain Entity: Agent Profile
///
/// This represents a clean business object in the domain layer.
/// It contains only business logic, no framework dependencies.
///
/// **Clean Architecture Layer**: Domain (innermost layer)
/// **Dependencies**: None (pure Dart)
///
/// **Usage**:
/// ```dart
/// final agent = AgentProfile(
///   agentId: 'AGT-2026-000001',
///   fullName: 'Rajesh Kumar Sharma',
///   status: AgentStatus.active,
/// );
/// ```

import 'package:equatable/equatable.dart';

/// Agent Profile domain entity
class AgentProfile extends Equatable {
  // ============================================================================
  // CORE IDENTIFICATION
  // ============================================================================

  /// Unique agent identifier (Format: AGT-YYYY-NNNNNN)
  final String agentId;

  /// Internal profile ID (UUID)
  final String profileId;

  /// Type of agent (ADVISOR, COORDINATOR, etc.)
  final AgentType agentType;

  /// Full name of the agent
  final String fullName;

  /// Current status
  final AgentStatus status;

  // ============================================================================
  // PROFILE DETAILS
  // ============================================================================

  /// PAN number (AAAAA9999A format)
  final String? panNumber;

  /// Date when status was last changed
  final DateTime? statusDate;

  /// Profile effective date
  final DateTime? effectiveDate;

  /// Profile creation timestamp
  final DateTime? createdAt;

  /// User who created the profile
  final String? createdBy;

  // ============================================================================
  // NESTED OBJECTS
  // ============================================================================

  /// Personal information
  final PersonalInfo? personalInfo;

  /// List of addresses
  final List<Address>? addresses;

  /// List of contact numbers
  final List<Contact>? contacts;

  /// List of email addresses
  final List<Email>? emails;

  /// Office association
  final Office? office;

  /// Advisor coordinator (for Advisors)
  final Coordinator? advisorCoordinator;

  /// Licenses
  final List<License>? licenses;

  /// Bank details (masked for security)
  final List<BankDetails>? bankDetails;

  // ============================================================================
  // CONSTRUCTOR
  // ============================================================================

  const AgentProfile({
    required this.agentId,
    required this.profileId,
    required this.agentType,
    required this.fullName,
    required this.status,
    this.panNumber,
    this.statusDate,
    this.effectiveDate,
    this.createdAt,
    this.createdBy,
    this.personalInfo,
    this.addresses,
    this.contacts,
    this.emails,
    this.office,
    this.advisorCoordinator,
    this.licenses,
    this.bankDetails,
  });

  // ============================================================================
  // BUSINESS LOGIC METHODS
  // ============================================================================

  /// Check if agent is active
  bool get isActive => status == AgentStatus.active;

  /// Check if agent is suspended
  bool get isSuspended => status == AgentStatus.suspended;

  /// Check if agent is terminated
  bool get isTerminated => status == AgentStatus.terminated;

  /// Check if agent can perform transactions
  bool get canPerformTransactions => isActive;

  /// Get primary contact number
  String? get primaryContact {
    if (contacts == null || contacts!.isEmpty) return null;
    final primary = contacts!.firstWhere(
      (c) => c.isPrimary,
      orElse: () => contacts!.first,
    );
    return primary.contactNumber;
  }

  /// Get primary email
  String? get primaryEmail {
    if (emails == null || emails!.isEmpty) return null;
    final primary = emails!.firstWhere(
      (e) => e.isPrimary,
      orElse: () => emails!.first,
    );
    return primary.emailAddress;
  }

  /// Get permanent address
  Address? get permanentAddress {
    if (addresses == null || addresses!.isEmpty) return null;
    return addresses!.firstWhere(
      (a) => a.addressType == AddressType.permanent,
      orElse: () => addresses!.first,
    );
  }

  /// Check if profile is complete
  bool get isProfileComplete {
    return panNumber != null &&
        personalInfo != null &&
        addresses != null &&
        addresses!.isNotEmpty &&
        contacts != null &&
        contacts!.isNotEmpty;
  }

  /// Check if agent requires license
  bool get requiresLicense {
    return agentType == AgentType.advisor ||
        agentType == AgentType.advisorCoordinator;
  }

  /// Check if agent has valid license
  bool get hasValidLicense {
    if (licenses == null || licenses!.isEmpty) return false;
    return licenses!.any((license) => license.isValid);
  }

  // ============================================================================
  // EQUALITY & HASH CODE (via Equatable)
  // ============================================================================

  @override
  List<Object?> get props => [
        agentId,
        profileId,
        agentType,
        fullName,
        status,
        panNumber,
        statusDate,
        effectiveDate,
        createdAt,
        createdBy,
        personalInfo,
        addresses,
        contacts,
        emails,
        office,
        advisorCoordinator,
        licenses,
        bankDetails,
      ];

  // ============================================================================
  // COPY WITH (for immutability)
  // ============================================================================

  AgentProfile copyWith({
    String? agentId,
    String? profileId,
    AgentType? agentType,
    String? fullName,
    AgentStatus? status,
    String? panNumber,
    DateTime? statusDate,
    DateTime? effectiveDate,
    DateTime? createdAt,
    String? createdBy,
    PersonalInfo? personalInfo,
    List<Address>? addresses,
    List<Contact>? contacts,
    List<Email>? emails,
    Office? office,
    Coordinator? advisorCoordinator,
    List<License>? licenses,
    List<BankDetails>? bankDetails,
  }) {
    return AgentProfile(
      agentId: agentId ?? this.agentId,
      profileId: profileId ?? this.profileId,
      agentType: agentType ?? this.agentType,
      fullName: fullName ?? this.fullName,
      status: status ?? this.status,
      panNumber: panNumber ?? this.panNumber,
      statusDate: statusDate ?? this.statusDate,
      effectiveDate: effectiveDate ?? this.effectiveDate,
      createdAt: createdAt ?? this.createdAt,
      createdBy: createdBy ?? this.createdBy,
      personalInfo: personalInfo ?? this.personalInfo,
      addresses: addresses ?? this.addresses,
      contacts: contacts ?? this.contacts,
      emails: emails ?? this.emails,
      office: office ?? this.office,
      advisorCoordinator: advisorCoordinator ?? this.advisorCoordinator,
      licenses: licenses ?? this.licenses,
      bankDetails: bankDetails ?? this.bankDetails,
    );
  }
}

// ============================================================================
// ENUMS
// ============================================================================

/// Agent type enumeration
enum AgentType {
  advisor,
  advisorCoordinator,
  departmentalEmployee,
  fieldOfficer,
  directAgent,
  gds;

  /// Get display name
  String get displayName {
    switch (this) {
      case AgentType.advisor:
        return 'Advisor';
      case AgentType.advisorCoordinator:
        return 'Advisor Coordinator';
      case AgentType.departmentalEmployee:
        return 'Departmental Employee';
      case AgentType.fieldOfficer:
        return 'Field Officer';
      case AgentType.directAgent:
        return 'Direct Agent';
      case AgentType.gds:
        return 'GDS';
    }
  }
}

/// Agent status enumeration
enum AgentStatus {
  active,
  inactive,
  suspended,
  terminated,
  deactivated;

  /// Get display name
  String get displayName {
    switch (this) {
      case AgentStatus.active:
        return 'Active';
      case AgentStatus.inactive:
        return 'Inactive';
      case AgentStatus.suspended:
        return 'Suspended';
      case AgentStatus.terminated:
        return 'Terminated';
      case AgentStatus.deactivated:
        return 'Deactivated';
    }
  }

  /// Check if status allows transactions
  bool get allowsTransactions => this == AgentStatus.active;
}

// ============================================================================
// NESTED VALUE OBJECTS (Placeholder - will be created separately)
// ============================================================================

/// Personal information value object
class PersonalInfo extends Equatable {
  final String title;
  final String firstName;
  final String? middleName;
  final String lastName;
  final DateTime dateOfBirth;
  final String gender;
  final String? maritalStatus;
  final String? category;
  final String? aadharNumber;

  const PersonalInfo({
    required this.title,
    required this.firstName,
    this.middleName,
    required this.lastName,
    required this.dateOfBirth,
    required this.gender,
    this.maritalStatus,
    this.category,
    this.aadharNumber,
  });

  int get age {
    final now = DateTime.now();
    int age = now.year - dateOfBirth.year;
    if (now.month < dateOfBirth.month ||
        (now.month == dateOfBirth.month && now.day < dateOfBirth.day)) {
      age--;
    }
    return age;
  }

  @override
  List<Object?> get props => [
        title,
        firstName,
        middleName,
        lastName,
        dateOfBirth,
        gender,
        maritalStatus,
        category,
        aadharNumber,
      ];
}

/// Address value object
class Address extends Equatable {
  final AddressType addressType;
  final String addressLine1;
  final String? addressLine2;
  final String city;
  final String state;
  final String pincode;
  final String country;

  const Address({
    required this.addressType,
    required this.addressLine1,
    this.addressLine2,
    required this.city,
    required this.state,
    required this.pincode,
    this.country = 'India',
  });

  @override
  List<Object?> get props => [
        addressType,
        addressLine1,
        addressLine2,
        city,
        state,
        pincode,
        country,
      ];
}

enum AddressType { permanent, communication, official }

/// Contact value object
class Contact extends Equatable {
  final ContactType contactType;
  final String contactNumber;
  final bool isPrimary;

  const Contact({
    required this.contactType,
    required this.contactNumber,
    this.isPrimary = false,
  });

  @override
  List<Object?> get props => [contactType, contactNumber, isPrimary];
}

enum ContactType { mobile, alternateMobile, landline }

/// Email value object
class Email extends Equatable {
  final EmailType emailType;
  final String emailAddress;
  final bool isPrimary;

  const Email({
    required this.emailType,
    required this.emailAddress,
    this.isPrimary = false,
  });

  @override
  List<Object?> get props => [emailType, emailAddress, isPrimary];
}

enum EmailType { official, personal }

/// Office value object
class Office extends Equatable {
  final String officeCode;
  final String officeType;
  final String? officeName;

  const Office({
    required this.officeCode,
    required this.officeType,
    this.officeName,
  });

  @override
  List<Object?> get props => [officeCode, officeType, officeName];
}

/// Coordinator value object
class Coordinator extends Equatable {
  final String coordinatorId;
  final String coordinatorName;
  final String status;

  const Coordinator({
    required this.coordinatorId,
    required this.coordinatorName,
    required this.status,
  });

  @override
  List<Object?> get props => [coordinatorId, coordinatorName, status];
}

/// License value object (placeholder)
class License extends Equatable {
  final String licenseId;
  final String licenseNumber;
  final String licenseType;
  final DateTime licenseDate;
  final DateTime renewalDate;
  final String status;

  const License({
    required this.licenseId,
    required this.licenseNumber,
    required this.licenseType,
    required this.licenseDate,
    required this.renewalDate,
    required this.status,
  });

  bool get isValid {
    return status == 'Active' && renewalDate.isAfter(DateTime.now());
  }

  @override
  List<Object?> get props => [
        licenseId,
        licenseNumber,
        licenseType,
        licenseDate,
        renewalDate,
        status,
      ];
}

/// Bank details value object (placeholder)
class BankDetails extends Equatable {
  final String bankDetailsId;
  final String accountType;
  final String accountNumberMasked;
  final String? bankName;
  final bool verified;

  const BankDetails({
    required this.bankDetailsId,
    required this.accountType,
    required this.accountNumberMasked,
    this.bankName,
    this.verified = false,
  });

  @override
  List<Object?> get props => [
        bankDetailsId,
        accountType,
        accountNumberMasked,
        bankName,
        verified,
      ];
}
