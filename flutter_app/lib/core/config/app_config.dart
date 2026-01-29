/// Application Configuration
/// For API-specific config, see api_config.dart
import 'api_config.dart';

class AppConfig {
  // ============================================================================
  // API CONFIGURATION (Delegated to ApiConfig)
  // See api_config.dart for base URLs, timeouts, and API settings
  // ============================================================================

  /// Get current base URL (delegates to ApiConfig)
  static String get baseUrl => ApiConfig.baseUrl;

  /// Get current environment (delegates to ApiConfig)
  static String get environment => ApiConfig.environment;

  /// Get API connect timeout
  static Duration get connectTimeout => ApiConfig.connectTimeout;

  /// Get API receive timeout
  static Duration get receiveTimeout => ApiConfig.receiveTimeout;

  /// Get API send timeout
  static Duration get sendTimeout => ApiConfig.sendTimeout;

  // Storage Keys
  static const String cachedAgentTypesKey = 'CACHED_AGENT_TYPES';
  static const String cachedStatesKey = 'CACHED_STATES';
  static const String cachedCategoriesKey = 'CACHED_CATEGORIES';
  static const String cachedDesignationsKey = 'CACHED_DESIGNATIONS';
  static const String cachedOfficeTypesKey = 'CACHED_OFFICE_TYPES';
  static const String cachedLicenseTypesKey = 'CACHED_LICENSE_TYPES';

  // Pagination
  static const int defaultPageSize = 20;
  static const int maxPageSize = 100;

  // Validation
  static const int minAge = 18;
  static const int maxAge = 70;
  static const int minNameLength = 2;
  static const int maxNameLength = 50;
  static const int maxAddressLength = 100;
  static const int maxReasonLength = 500;
  static const int minReasonLength = 10;
  static const int pincodeLength = 6;
  static const int aadharLength = 12;
  static const int mobileLength = 10;

  // Session
  static const Duration sessionExpiryDuration = Duration(hours: 2);
  static const Duration otpExpiryDuration = Duration(minutes: 10);
  static const int maxOtpAttempts = 5;

  // File Upload
  static const int maxFileSize = 5 * 1024 * 1024; // 5 MB
  static const List<String> allowedFileExtensions = ['pdf', 'jpg', 'jpeg', 'png'];

  // App Info
  static const String appName = 'PLI Agent Management';
  static const String appVersion = '1.0.0';
  static const String supportEmail = 'support@postallifeinsurance.gov.in';
}
