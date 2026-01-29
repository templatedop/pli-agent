import '../config/app_config.dart';

class Validators {
  // PAN Validation
  static String? validatePan(String? value) {
    if (value == null || value.isEmpty) {
      return 'PAN is required';
    }

    // PAN format: AAAAA9999A (5 letters, 4 digits, 1 letter)
    final panRegex = RegExp(r'^[A-Z]{5}[0-9]{4}[A-Z]{1}$');

    if (!panRegex.hasMatch(value)) {
      return 'Invalid PAN format (e.g., ABCDE1234F)';
    }

    return null;
  }

  // Aadhar Validation
  static String? validateAadhar(String? value) {
    if (value == null || value.isEmpty) {
      return null; // Aadhar is optional
    }

    // Remove spaces
    final aadhar = value.replaceAll(' ', '');

    if (aadhar.length != AppConfig.aadharLength) {
      return 'Aadhar must be 12 digits';
    }

    if (!RegExp(r'^[0-9]{12}$').hasMatch(aadhar)) {
      return 'Aadhar must contain only digits';
    }

    return null;
  }

  // Mobile Number Validation
  static String? validateMobile(String? value) {
    if (value == null || value.isEmpty) {
      return 'Mobile number is required';
    }

    if (value.length != AppConfig.mobileLength) {
      return 'Mobile number must be 10 digits';
    }

    // Must start with 6, 7, 8, or 9
    if (!RegExp(r'^[6-9][0-9]{9}$').hasMatch(value)) {
      return 'Invalid mobile number';
    }

    return null;
  }

  // Email Validation
  static String? validateEmail(String? value) {
    if (value == null || value.isEmpty) {
      return 'Email is required';
    }

    final emailRegex = RegExp(
      r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$',
    );

    if (!emailRegex.hasMatch(value)) {
      return 'Invalid email address';
    }

    return null;
  }

  // Name Validation
  static String? validateName(String? value, {String fieldName = 'Name'}) {
    if (value == null || value.isEmpty) {
      return '$fieldName is required';
    }

    if (value.length < AppConfig.minNameLength) {
      return '$fieldName must be at least ${AppConfig.minNameLength} characters';
    }

    if (value.length > AppConfig.maxNameLength) {
      return '$fieldName must not exceed ${AppConfig.maxNameLength} characters';
    }

    // Only letters, spaces, and hyphens
    if (!RegExp(r'^[a-zA-Z\s\-\.]+$').hasMatch(value)) {
      return '$fieldName can only contain letters, spaces, hyphens, and dots';
    }

    return null;
  }

  // Pincode Validation
  static String? validatePincode(String? value) {
    if (value == null || value.isEmpty) {
      return 'Pincode is required';
    }

    if (value.length != AppConfig.pincodeLength) {
      return 'Pincode must be 6 digits';
    }

    if (!RegExp(r'^[0-9]{6}$').hasMatch(value)) {
      return 'Invalid pincode';
    }

    return null;
  }

  // IFSC Code Validation
  static String? validateIfsc(String? value) {
    if (value == null || value.isEmpty) {
      return null; // IFSC is optional for POSB accounts
    }

    // IFSC format: AAAA0NNNNNN (4 letters, 0, 6 alphanumeric)
    final ifscRegex = RegExp(r'^[A-Z]{4}0[A-Z0-9]{6}$');

    if (!ifscRegex.hasMatch(value)) {
      return 'Invalid IFSC code (e.g., SBIN0001234)';
    }

    return null;
  }

  // Bank Account Number Validation
  static String? validateAccountNumber(String? value) {
    if (value == null || value.isEmpty) {
      return 'Account number is required';
    }

    if (value.length < 9 || value.length > 18) {
      return 'Account number must be 9-18 digits';
    }

    if (!RegExp(r'^[0-9]+$').hasMatch(value)) {
      return 'Account number must contain only digits';
    }

    return null;
  }

  // Date of Birth Validation (Age check)
  static String? validateDateOfBirth(DateTime? value) {
    if (value == null) {
      return 'Date of birth is required';
    }

    final now = DateTime.now();
    final age = now.year - value.year;

    // Check if birthday hasn't occurred this year
    final adjustedAge = (now.month < value.month ||
                        (now.month == value.month && now.day < value.day))
        ? age - 1
        : age;

    if (adjustedAge < AppConfig.minAge) {
      return 'Minimum age is ${AppConfig.minAge} years';
    }

    if (adjustedAge > AppConfig.maxAge) {
      return 'Maximum age is ${AppConfig.maxAge} years';
    }

    return null;
  }

  // Required Field Validation
  static String? validateRequired(String? value, {String fieldName = 'This field'}) {
    if (value == null || value.trim().isEmpty) {
      return '$fieldName is required';
    }
    return null;
  }

  // Address Validation
  static String? validateAddress(String? value) {
    if (value == null || value.isEmpty) {
      return 'Address is required';
    }

    if (value.length > AppConfig.maxAddressLength) {
      return 'Address must not exceed ${AppConfig.maxAddressLength} characters';
    }

    return null;
  }

  // Reason Validation (for termination, reinstatement, etc.)
  static String? validateReason(String? value, {String fieldName = 'Reason'}) {
    if (value == null || value.isEmpty) {
      return '$fieldName is required';
    }

    if (value.length < AppConfig.minReasonLength) {
      return '$fieldName must be at least ${AppConfig.minReasonLength} characters';
    }

    if (value.length > AppConfig.maxReasonLength) {
      return '$fieldName must not exceed ${AppConfig.maxReasonLength} characters';
    }

    return null;
  }

  // License Number Validation
  static String? validateLicenseNumber(String? value) {
    if (value == null || value.isEmpty) {
      return 'License number is required';
    }

    // Basic validation - alphanumeric with hyphens
    if (!RegExp(r'^[A-Z0-9\-]+$').hasMatch(value)) {
      return 'License number can only contain letters, numbers, and hyphens';
    }

    return null;
  }

  // OTP Validation
  static String? validateOtp(String? value) {
    if (value == null || value.isEmpty) {
      return 'OTP is required';
    }

    if (value.length != 6) {
      return 'OTP must be 6 digits';
    }

    if (!RegExp(r'^[0-9]{6}$').hasMatch(value)) {
      return 'OTP must contain only digits';
    }

    return null;
  }

  // Employee ID Validation
  static String? validateEmployeeId(String? value) {
    if (value == null || value.isEmpty) {
      return 'Employee ID is required';
    }

    if (value.length < 5 || value.length > 20) {
      return 'Employee ID must be 5-20 characters';
    }

    return null;
  }

  // Agent ID Validation
  static String? validateAgentId(String? value) {
    if (value == null || value.isEmpty) {
      return 'Agent ID is required';
    }

    // Agent ID format: AGT-YYYY-NNNNNN
    final agentIdRegex = RegExp(r'^AGT-\d{4}-\d{6}$');

    if (!agentIdRegex.hasMatch(value)) {
      return 'Invalid Agent ID format (e.g., AGT-2026-000001)';
    }

    return null;
  }

  // Goal Target Validation
  static String? validateGoalTarget(String? value) {
    if (value == null || value.isEmpty) {
      return 'Target value is required';
    }

    final target = int.tryParse(value);

    if (target == null) {
      return 'Target must be a number';
    }

    if (target <= 0) {
      return 'Target must be greater than 0';
    }

    return null;
  }

  // Amount Validation
  static String? validateAmount(String? value) {
    if (value == null || value.isEmpty) {
      return 'Amount is required';
    }

    final amount = double.tryParse(value);

    if (amount == null) {
      return 'Invalid amount';
    }

    if (amount <= 0) {
      return 'Amount must be greater than 0';
    }

    return null;
  }
}
