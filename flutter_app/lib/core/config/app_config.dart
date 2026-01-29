class AppConfig {
  // Environment
  static const String environment = String.fromEnvironment(
    'ENVIRONMENT',
    defaultValue: 'development',
  );

  // API Base URLs
  static const String productionBaseUrl = 'https://api.postallifeinsurance.gov.in/v1';
  static const String stagingBaseUrl = 'https://staging-api.postallifeinsurance.gov.in/v1';
  static const String developmentBaseUrl = 'http://localhost:8080/api/v1';

  // Get current base URL based on environment
  static String get baseUrl {
    switch (environment) {
      case 'production':
        return productionBaseUrl;
      case 'staging':
        return stagingBaseUrl;
      case 'development':
      default:
        return developmentBaseUrl;
    }
  }

  // API Timeouts
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);
  static const Duration sendTimeout = Duration(seconds: 30);

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
