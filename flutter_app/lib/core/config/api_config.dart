/// API Configuration
///
/// IMPORTANT: This file contains all API configuration.
/// Update base URLs and API versions here for easy maintenance.

class ApiConfig {
  // ============================================================================
  // BASE URLS - UPDATE THESE FOR DIFFERENT ENVIRONMENTS
  // ============================================================================

  /// Production API Base URL
  /// Change this when production server URL changes
  static const String productionBaseUrl = 'https://api.postallifeinsurance.gov.in';

  /// Staging API Base URL
  /// Change this when staging server URL changes
  static const String stagingBaseUrl = 'https://staging-api.postallifeinsurance.gov.in';

  /// Development API Base URL
  /// Change this to your local development server
  static const String developmentBaseUrl = 'http://localhost:8080';

  /// Alternative Development URLs (Uncomment as needed)
  // static const String developmentBaseUrl = 'http://192.168.1.100:8080'; // LAN
  // static const String developmentBaseUrl = 'http://10.0.2.2:8080';      // Android Emulator

  // ============================================================================
  // API VERSIONING
  // ============================================================================

  /// API Version Path
  /// Change this when API version changes (e.g., v2, v3)
  static const String apiVersion = '/v1';

  /// Full API Path = Base URL + API Version
  /// Example: https://api.postallifeinsurance.gov.in/v1

  // ============================================================================
  // ENVIRONMENT SELECTION
  // ============================================================================

  /// Current Environment
  /// Options: 'development', 'staging', 'production'
  /// Can be overridden with --dart-define=ENVIRONMENT=production
  static const String environment = String.fromEnvironment(
    'ENVIRONMENT',
    defaultValue: 'development',
  );

  /// Get Full Base URL based on current environment
  static String get baseUrl {
    final baseUrlWithoutVersion = _getBaseUrlForEnvironment();
    return '$baseUrlWithoutVersion$apiVersion';
  }

  /// Get Base URL without version (for special cases)
  static String get baseUrlWithoutVersion => _getBaseUrlForEnvironment();

  /// Internal method to get base URL for environment
  static String _getBaseUrlForEnvironment() {
    switch (environment.toLowerCase()) {
      case 'production':
      case 'prod':
        return productionBaseUrl;
      case 'staging':
      case 'stage':
        return stagingBaseUrl;
      case 'development':
      case 'dev':
      default:
        return developmentBaseUrl;
    }
  }

  // ============================================================================
  // REQUEST TIMEOUT CONFIGURATION
  // ============================================================================

  /// Connection timeout in seconds
  /// Increase if server is slow
  static const int connectTimeoutSeconds = 30;

  /// Receive timeout in seconds
  /// Increase for large file downloads
  static const int receiveTimeoutSeconds = 30;

  /// Send timeout in seconds
  /// Increase for large file uploads
  static const int sendTimeoutSeconds = 30;

  /// Get timeout as Duration
  static Duration get connectTimeout => Duration(seconds: connectTimeoutSeconds);
  static Duration get receiveTimeout => Duration(seconds: receiveTimeoutSeconds);
  static Duration get sendTimeout => Duration(seconds: sendTimeoutSeconds);

  // ============================================================================
  // RETRY CONFIGURATION
  // ============================================================================

  /// Number of retry attempts for failed requests
  static const int maxRetryAttempts = 3;

  /// Delay between retries in seconds
  static const int retryDelaySeconds = 2;

  // ============================================================================
  // HEADERS CONFIGURATION
  // ============================================================================

  /// Default request headers
  /// Add or modify headers here
  static Map<String, String> get defaultHeaders => {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    // Add custom headers here if needed
    // 'X-Api-Key': 'your-api-key',
    // 'X-App-Version': '1.0.0',
  };

  // ============================================================================
  // LOGGING CONFIGURATION
  // ============================================================================

  /// Enable/Disable API logging
  static const bool enableLogging = true;

  /// Enable logging only in development
  static bool get shouldEnableLogging {
    return enableLogging && (environment == 'development' || environment == 'staging');
  }

  // ============================================================================
  // HELPER METHODS
  // ============================================================================

  /// Get full URL for an endpoint
  /// Example: getFullUrl('/agents/search')
  /// Returns: https://api.postallifeinsurance.gov.in/v1/agents/search
  static String getFullUrl(String endpoint) {
    // Remove leading slash if present
    final cleanEndpoint = endpoint.startsWith('/') ? endpoint.substring(1) : endpoint;
    return '$baseUrl/$cleanEndpoint';
  }

  /// Get environment info for debugging
  static Map<String, dynamic> getEnvironmentInfo() {
    return {
      'environment': environment,
      'baseUrl': baseUrl,
      'apiVersion': apiVersion,
      'connectTimeout': connectTimeoutSeconds,
      'receiveTimeout': receiveTimeoutSeconds,
      'sendTimeout': sendTimeoutSeconds,
      'loggingEnabled': shouldEnableLogging,
    };
  }

  /// Print current configuration (for debugging)
  static void printConfiguration() {
    print('=== API Configuration ===');
    print('Environment: $environment');
    print('Base URL: $baseUrl');
    print('API Version: $apiVersion');
    print('Connect Timeout: ${connectTimeoutSeconds}s');
    print('Receive Timeout: ${receiveTimeoutSeconds}s');
    print('Send Timeout: ${sendTimeoutSeconds}s');
    print('Logging Enabled: $shouldEnableLogging');
    print('========================');
  }
}

// ============================================================================
// QUICK CONFIGURATION GUIDE
// ============================================================================
//
// HOW TO CHANGE BASE URL:
// 1. Update the corresponding base URL constant above (productionBaseUrl, stagingBaseUrl, developmentBaseUrl)
// 2. Run the app - changes will be applied automatically
//
// HOW TO CHANGE API VERSION:
// 1. Update apiVersion constant (e.g., '/v2', '/v3')
// 2. Run the app - all endpoints will use new version
//
// HOW TO CHANGE ENVIRONMENT:
// 1. Run with: flutter run --dart-define=ENVIRONMENT=staging
// 2. Or change default value in environment constant
//
// HOW TO CHANGE TIMEOUTS:
// 1. Update the timeout constants in seconds
// 2. Changes apply to all API calls
//
// HOW TO ADD CUSTOM HEADERS:
// 1. Add to defaultHeaders map
// 2. Headers will be sent with every request
//
// HOW TO DISABLE LOGGING:
// 1. Set enableLogging = false
// 2. Or it auto-disables in production
// ============================================================================
