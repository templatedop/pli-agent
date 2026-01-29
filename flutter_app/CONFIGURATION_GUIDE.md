# Configuration Guide - Easy Updates & Maintenance

## 📋 Table of Contents

1. [How to Change Base URL](#how-to-change-base-url)
2. [How to Update API Endpoints](#how-to-update-api-endpoints)
3. [How to Modify Request Models](#how-to-modify-request-models)
4. [How to Modify Response Models](#how-to-modify-response-models)
5. [How to Change API Version](#how-to-change-api-version)
6. [How to Add New Endpoint](#how-to-add-new-endpoint)
7. [Quick Reference](#quick-reference)

---

## 1. How to Change Base URL

### ⚙️ File Location
```
lib/core/config/api_config.dart
```

### 📝 Steps

**Step 1**: Open the configuration file
```bash
nano lib/core/config/api_config.dart
```

**Step 2**: Update the appropriate base URL constant

```dart
// For Production
static const String productionBaseUrl = 'https://NEW-PRODUCTION-URL.com';

// For Staging
static const String stagingBaseUrl = 'https://NEW-STAGING-URL.com';

// For Development
static const String developmentBaseUrl = 'http://localhost:8080';
// Or for LAN: 'http://192.168.1.100:8080'
// Or for Android Emulator: 'http://10.0.2.2:8080'
```

**Step 3**: Save the file

**Step 4**: Restart the app - changes apply automatically!

### 🎯 Example

```dart
// BEFORE
static const String productionBaseUrl = 'https://api.postallifeinsurance.gov.in';

// AFTER (New domain)
static const String productionBaseUrl = 'https://api-new.postallifeinsurance.gov.in';
```

### ✅ Verification

Check current configuration:
```dart
ApiConfig.printConfiguration();
// Prints current base URL and environment
```

---

## 2. How to Update API Endpoints

### ⚙️ File Location
```
lib/core/constants/api_endpoints_config.dart
```

### 📝 Steps

**Step 1**: Open the endpoints file
```bash
nano lib/core/constants/api_endpoints_config.dart
```

**Step 2**: Find the endpoint you want to change
Use Ctrl+F (or Cmd+F) to search for the endpoint

**Step 3**: Update the endpoint path

```dart
// BEFORE
static String get agentProfile => _endpoint('/agents/{agent_id}');

// AFTER (API path changed)
static String get agentProfile => _endpoint('/api/agents/{agent_id}');
```

**Step 4**: Save the file

**Step 5**: Restart the app - all API calls automatically use new endpoint!

### 🎯 Example Changes

```dart
// Example 1: Add prefix
// BEFORE: '/agents/{agent_id}'
// AFTER:  '/v2/agents/{agent_id}'
static String get agentProfile => _endpoint('/v2/agents/{agent_id}');

// Example 2: Change resource name
// BEFORE: '/agents/search'
// AFTER:  '/agents/find'
static String get searchAgents => _endpoint('/agents/find');

// Example 3: Add query parameter documentation
/// Search agents
/// GET /agents/search?status={status}&page={page}
static String get searchAgents => _endpoint('/agents/search');
```

### ✅ Verification

Print endpoint to verify:
```dart
print(ApiEndpoints.agentProfile);  // Prints: /agents/{agent_id}
```

---

## 3. How to Modify Request Models

### ⚙️ File Location
```
lib/data/models/requests/{entity}/{action}_{entity}_request.dart
```

Example: `lib/data/models/requests/agent/create_agent_request.dart`

### 📝 Steps to Add a Field

**Step 1**: Open the request model file

**Step 2**: Add the new field in the factory constructor

```dart
@freezed
class CreateAgentRequest with _$CreateAgentRequest {
  const factory CreateAgentRequest({
    required String submitAction,
    required bool confirmation,

    // ✅ ADD NEW FIELD HERE
    @JsonKey(name: 'priority_flag') bool? priorityFlag,  // ← New field
  }) = _CreateAgentRequest;
}
```

**Step 3**: Run code generation

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

**Step 4**: Document the change in the CHANGE LOG section

```dart
// ============================================================================
// CHANGE LOG
// ============================================================================
//
// 2026-01-30 v1.1.0 - Added priority_flag
//   - Added: priority_flag (optional bool)
//   - Reason: Support for priority processing
//   - API Version: v1.1
```

**Step 5**: Use the updated model

```dart
final request = CreateAgentRequest(
  submitAction: 'CREATE',
  confirmation: true,
  priorityFlag: true,  // ← New field in use
);
```

### 📝 Steps to Remove a Field

**Step 1**: Comment out or delete the field

```dart
// String? oldField,  // ← Commented out (deprecated)
```

**Step 2**: Run code generation

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

**Step 3**: Document the removal

### 📝 Steps to Rename API Field

**Step 1**: Update @JsonKey annotation only (keep Dart field name same)

```dart
// API changed "submitted_by" to "creator_id"

// BEFORE
@JsonKey(name: 'submitted_by') required String submittedBy,

// AFTER (Only change @JsonKey, keep Dart name for compatibility)
@JsonKey(name: 'creator_id') required String submittedBy,
```

**Step 2**: Run code generation

**Step 3**: Document the change

---

## 4. How to Modify Response Models

### ⚙️ File Location
```
lib/data/models/responses/{entity}/{entity}_{type}_response.dart
```

Example: `lib/data/models/responses/agent/agent_profile_response.dart`

### 📝 Steps to Add a Field

**Step 1**: Open the response model file

**Step 2**: Add the new field

```dart
@freezed
class AgentProfileResponse with _$AgentProfileResponse {
  const factory AgentProfileResponse({
    required String agentId,
    required String fullName,

    // ✅ ADD NEW FIELD HERE
    @JsonKey(name: 'email_verified') bool? emailVerified,  // ← New field
  }) = _AgentProfileResponse;
}
```

**Step 3**: Run code generation

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

**Step 4**: Document the change

**Step 5**: Access the new field

```dart
final response = AgentProfileResponse.fromJson(jsonData);
print(response.emailVerified);  // Access new field
```

### 🎯 Common Response Changes

```dart
// Add new nested object
@JsonKey(name: 'verification_status') Map<String, dynamic>? verificationStatus,

// Add new array
@JsonKey(name: 'attached_documents') List<Map<String, dynamic>>? attachedDocuments,

// Change field type (String → int)
// BEFORE: String? age,
// AFTER:
int? age,

// Make required field optional (or vice versa)
// BEFORE: required String email,
// AFTER:
String? email,
```

---

## 5. How to Change API Version

### ⚙️ File Location
```
lib/core/config/api_config.dart
```

### 📝 Steps

**Step 1**: Open the configuration file

**Step 2**: Update the API version

```dart
// BEFORE
static const String apiVersion = '/v1';

// AFTER (API upgraded to v2)
static const String apiVersion = '/v2';
```

**Step 3**: Save and restart

**Result**: All endpoints now use `/v2` instead of `/v1`

### 🎯 Example

```
BEFORE:
https://api.postallifeinsurance.gov.in/v1/agents/search

AFTER:
https://api.postallifeinsurance.gov.in/v2/agents/search
```

---

## 6. How to Add New Endpoint

### ⚙️ File Location
```
lib/core/constants/api_endpoints_config.dart
```

### 📝 Steps

**Step 1**: Open the endpoints file

**Step 2**: Find the appropriate section (based on user journey)

**Step 3**: Add new endpoint with documentation

```dart
// ============================================================================
// UJ-002: AGENT PROFILE UPDATE
// ============================================================================

/// Get agent statistics
/// GET /agents/{agent_id}/statistics
/// Response: {statistics}
static String get agentStatistics => _endpoint('/agents/{agent_id}/statistics');
```

**Step 4**: Use the new endpoint

```dart
final url = ApiEndpoints.agentStatistics;
final finalUrl = ApiEndpoints.replacePath(url, {'agent_id': 'AGT-001'});
// Result: '/agents/AGT-001/statistics'
```

---

## 7. Quick Reference

### 📋 Configuration Files Matrix

| What to Change | File Location | Action |
|----------------|---------------|--------|
| **Base URL** | `lib/core/config/api_config.dart` | Update constant |
| **API Version** | `lib/core/config/api_config.dart` | Update `apiVersion` |
| **Endpoint Path** | `lib/core/constants/api_endpoints_config.dart` | Update getter |
| **Request Model** | `lib/data/models/requests/{entity}/` | Add field + run build_runner |
| **Response Model** | `lib/data/models/responses/{entity}/` | Add field + run build_runner |
| **Timeouts** | `lib/core/config/api_config.dart` | Update timeout constants |
| **Headers** | `lib/core/config/api_config.dart` | Update `defaultHeaders` |

### 🔧 Common Commands

```bash
# Generate code after model changes
flutter pub run build_runner build --delete-conflicting-outputs

# Watch mode (auto-generate on save)
flutter pub run build_runner watch

# Clean generated files
flutter pub run build_runner clean

# Run with specific environment
flutter run --dart-define=ENVIRONMENT=production
flutter run --dart-define=ENVIRONMENT=staging
flutter run --dart-define=ENVIRONMENT=development
```

### 📱 Environment Selection

```bash
# Method 1: Command line (recommended)
flutter run --dart-define=ENVIRONMENT=production

# Method 2: Change default in api_config.dart
static const String environment = String.fromEnvironment(
  'ENVIRONMENT',
  defaultValue: 'staging',  // ← Change this
);
```

### 🎯 Typical Workflow

```
1. API changes announced
   ↓
2. Update appropriate file:
   - Base URL → api_config.dart
   - Endpoint → api_endpoints_config.dart
   - Request/Response → models/
   ↓
3. If model changed:
   - Run: flutter pub run build_runner build --delete-conflicting-outputs
   ↓
4. Test the changes
   ↓
5. Document in CHANGE LOG
   ↓
6. Commit & deploy
```

---

## 🎨 Visual File Structure

```
lib/
├── core/
│   ├── config/
│   │   ├── api_config.dart           ← Base URLs, Timeouts, Headers
│   │   └── app_config.dart           ← App-level config
│   └── constants/
│       └── api_endpoints_config.dart ← All API endpoints
│
└── data/
    └── models/
        ├── requests/                  ← Request models
        │   ├── agent/
        │   │   ├── create_agent_request.dart
        │   │   └── update_agent_request.dart
        │   ├── license/
        │   └── bank/
        │
        └── responses/                 ← Response models
            ├── agent/
            │   └── agent_profile_response.dart
            ├── license/
            └── bank/
```

---

## 💡 Best Practices

### ✅ DO

- **Document every change** in CHANGE LOG
- **Keep Dart field names stable** (change @JsonKey instead)
- **Run build_runner** after model changes
- **Test after every update**
- **Use descriptive field names**
- **Group related fields** together

### ❌ DON'T

- **Don't change Dart field names** unnecessarily (breaks existing code)
- **Don't skip code generation** (will cause runtime errors)
- **Don't forget to document** changes
- **Don't hardcode URLs** in code (use config files)
- **Don't mix environments** (use proper environment selection)

---

## 🆘 Troubleshooting

### Problem: Build runner fails

```bash
# Solution: Clean and rebuild
flutter pub run build_runner clean
flutter pub run build_runner build --delete-conflicting-outputs
```

### Problem: Changes not reflecting

```bash
# Solution: Hot restart (not hot reload)
# Press R in terminal or
flutter run
```

### Problem: Wrong environment

```bash
# Solution: Check current environment
ApiConfig.printConfiguration();

# Run with correct environment
flutter run --dart-define=ENVIRONMENT=production
```

### Problem: Endpoint not found

```bash
# Solution: Verify endpoint exists
print(ApiEndpoints.agentProfile);

# Check api_endpoints_config.dart for typos
```

---

## 📞 Support

For help with configuration:

1. Check this guide first
2. Check MODEL_STRUCTURE_GUIDE.md for model changes
3. Check README.md for general documentation
4. Check API YAML files for API specifications

---

## 📚 Related Documentation

- [MODEL_STRUCTURE_GUIDE.md](lib/data/models/MODEL_STRUCTURE_GUIDE.md) - Detailed model documentation
- [README.md](README.md) - Project overview and setup
- [FLUTTER_APP_PLAN.md](../FLUTTER_APP_PLAN.md) - Development roadmap
- [API Specifications](../agents/*.yaml) - OpenAPI specs

---

**Last Updated**: 2026-01-29
**Version**: 1.0.0
**Maintained by**: PLI Development Team

---

## 🎯 Summary

Three files to remember for configuration:

1. **`api_config.dart`** - Base URLs, versions, timeouts
2. **`api_endpoints_config.dart`** - All endpoint paths
3. **`models/`** - Request/Response structures

All changes in these files automatically propagate throughout the entire app!
