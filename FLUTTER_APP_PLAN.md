# Flutter App Development Plan - Agent Profile Management

## Project Overview

**App Name**: PLI Agent Management
**Platform**: Flutter (Mobile + Web)
**Backend**: Agent Profile Management API (78 endpoints)
**Architecture**: Clean Architecture with BLoC pattern
**Auth**: Plain APIs (No authentication required currently)

---

## Phase 1: Project Setup & API Client Generation (Day 1)

### 1.1 Flutter Project Setup
- Create Flutter project with proper folder structure
- Configure dependencies (dio, flutter_bloc, get_it, freezed, etc.)
- Set up environment configuration
- Configure base URL for API

### 1.2 API Client Generation
- Merge 3 OpenAPI YAML files
- Generate Dart API client using openapi-generator-dart
- Configure API client with Dio HTTP client
- Add error handling and interceptors

### 1.3 Project Structure
```
lib/
├── core/
│   ├── config/
│   │   └── api_config.dart
│   ├── constants/
│   │   ├── app_constants.dart
│   │   └── api_endpoints.dart
│   ├── errors/
│   │   ├── failures.dart
│   │   └── exceptions.dart
│   ├── network/
│   │   ├── api_client.dart
│   │   └── network_info.dart
│   ├── theme/
│   │   ├── app_theme.dart
│   │   └── app_colors.dart
│   └── utils/
│       ├── validators.dart
│       └── formatters.dart
├── data/
│   ├── datasources/
│   │   ├── agent_remote_datasource.dart
│   │   ├── license_remote_datasource.dart
│   │   └── lookup_remote_datasource.dart
│   ├── models/
│   │   ├── agent_model.dart
│   │   ├── license_model.dart
│   │   └── response_models.dart
│   └── repositories/
│       ├── agent_repository_impl.dart
│       └── license_repository_impl.dart
├── domain/
│   ├── entities/
│   │   ├── agent.dart
│   │   ├── license.dart
│   │   └── workflow_state.dart
│   ├── repositories/
│   │   ├── agent_repository.dart
│   │   └── license_repository.dart
│   └── usecases/
│       ├── create_agent_profile.dart
│       ├── update_agent_profile.dart
│       ├── manage_licenses.dart
│       └── search_agents.dart
├── presentation/
│   ├── blocs/
│   │   ├── agent/
│   │   │   ├── agent_bloc.dart
│   │   │   ├── agent_event.dart
│   │   │   └── agent_state.dart
│   │   ├── license/
│   │   └── lookup/
│   ├── screens/
│   │   ├── home/
│   │   ├── agent_creation/
│   │   ├── agent_profile/
│   │   ├── license_management/
│   │   ├── bank_details/
│   │   ├── goals/
│   │   └── search/
│   └── widgets/
│       ├── common/
│       ├── forms/
│       └── cards/
├── generated/
│   └── api/              # Generated API client
├── injection_container.dart
└── main.dart
```

---

## Phase 2: Core Infrastructure (Day 2)

### 2.1 Domain Layer
- Define entity classes (Agent, License, Address, Contact, etc.)
- Create repository interfaces
- Implement use cases for each user journey

### 2.2 Data Layer
- Implement data models with JSON serialization
- Create remote data sources wrapping generated API client
- Implement repositories with error handling
- Add data mappers (Model ↔ Entity)

### 2.3 Dependency Injection
- Set up GetIt service locator
- Register all dependencies
- Configure API client instances

---

## Phase 3: State Management (Day 3)

### 3.1 BLoC Setup
- Create BLoCs for each feature:
  - AgentProfileBloc (Create, Update, View)
  - LicenseManagementBloc
  - BankDetailsBloc
  - GoalsBloc
  - SearchBloc
  - LookupBloc (Dropdowns)

### 3.2 State Classes
- Define states (Initial, Loading, Loaded, Error)
- Define events for user actions
- Implement business logic in BLoCs

---

## Phase 4: UI Implementation - User Journeys (Day 4-7)

### Journey 1: Agent Profile Creation (UJ-001) - Critical
**Screens**:
1. Agent Type Selection
2. HRMS Integration (for Departmental Employees)
3. Coordinator Selection (for Advisors)
4. Profile Details Form (Multi-step)
5. Review & Submit

**APIs Used**: AGT-001 to AGT-021

### Journey 2: Agent Profile Update (UJ-002)
**Screens**:
1. Agent Search
2. Profile Dashboard
3. Section Update Forms
4. Approval Status

**APIs Used**: AGT-022 to AGT-028

### Journey 3: License Management (UJ-003) - Critical
**Screens**:
1. License List View
2. Add/Edit License Form
3. License Renewal
4. Reminder Schedule View

**APIs Used**: AGT-029 to AGT-038

### Journey 4: Agent Termination (UJ-004)
**Screens**:
1. Termination Form
2. Termination Letter Preview
3. Confirmation Dialog

**APIs Used**: AGT-039, AGT-040

### Journey 5: Portal Authentication (UJ-005)
**Screens**:
1. Login Screen (Agent ID + Password)
2. OTP Verification Screen
3. Account Locked Screen

**APIs Used**: AGT-042 to AGT-046
*Note*: Currently APIs are plain, so this will be prepared for future auth implementation

### Journey 6: Self-Service Profile Update (UJ-006)
**Screens**:
1. Agent Dashboard
2. Self-Update Forms (Address, Contact, Bank)
3. OTP Verification for Changes

**APIs Used**: AGT-047 to AGT-049, AGT-068

### Journey 7: Bank Details Management (UJ-007) - Critical
**Screens**:
1. Bank Details List (Masked)
2. Add Bank Account Form (with IFSC validation)
3. OTP Verification

**APIs Used**: AGT-050 to AGT-054

### Journey 8: Goal Setting (UJ-008)
**Screens**:
1. Goals Dashboard
2. Set Goals Form
3. Goal Progress View
4. Goal Templates

**APIs Used**: AGT-055 to AGT-059, AGT-069

### Journey 9: Status Reinstatement (UJ-009)
**Screens**:
1. Reinstatement Request Form
2. Document Upload
3. Approval Workflow Status

**APIs Used**: AGT-060 to AGT-063, AGT-070, AGT-071

### Journey 10: Search & Export (UJ-010)
**Screens**:
1. Advanced Search Form
2. Search Results List
3. Export Configuration
4. Export Status & Download

**APIs Used**: AGT-022, AGT-064 to AGT-067

---

## Phase 5: Common Features (Day 8)

### 5.1 Reusable Widgets
- Custom Form Fields with validation
- Dropdown builders (Agent Types, States, Categories, etc.)
- Date Picker with validation
- File Picker for documents
- Status indicators
- Progress indicators
- Workflow state visualizer

### 5.2 Navigation
- Set up go_router for navigation
- Define all routes
- Implement deep linking support

### 5.3 Form Validations
- PAN validation (Format + Uniqueness check)
- Mobile number validation (10 digits, starts with 6-9)
- Email validation
- Aadhar validation (12 digits)
- IFSC validation
- Age calculation and validation (18-70 years)
- Pincode validation (6 digits)

---

## Phase 6: Enhanced Features (Day 9)

### 6.1 Offline Support
- Local database with Hive/Drift
- Cache lookup data (dropdowns)
- Queue API calls when offline
- Sync when back online

### 6.2 File Handling
- Document upload for reinstatement
- PDF download for termination letters
- Export file download (Excel/PDF)

### 6.3 Notifications
- Display in-app notifications
- Show success/error messages
- Alert for expiring licenses

---

## Phase 7: Testing & Polish (Day 10)

### 7.1 Testing
- Unit tests for BLoCs
- Widget tests for screens
- Integration tests for user flows
- API integration tests

### 7.2 Error Handling
- Network error handling
- Validation error display
- User-friendly error messages

### 7.3 UI/UX Polish
- Loading states
- Empty states
- Error states
- Animations and transitions

---

## Technical Stack

### Core Dependencies
```yaml
dependencies:
  flutter:
    sdk: flutter

  # State Management
  flutter_bloc: ^8.1.3
  equatable: ^2.0.5

  # Dependency Injection
  get_it: ^7.6.4
  injectable: ^2.3.2

  # Networking
  dio: ^5.4.0
  retrofit: ^4.0.3
  pretty_dio_logger: ^1.3.1

  # JSON Serialization
  json_annotation: ^4.8.1
  freezed_annotation: ^2.4.1

  # Code Generation
  build_runner: ^2.4.6
  json_serializable: ^6.7.1
  freezed: ^2.4.5
  retrofit_generator: ^8.0.4

  # UI
  flutter_form_builder: ^9.1.1
  form_builder_validators: ^9.1.0
  flutter_svg: ^2.0.9
  cached_network_image: ^3.3.0

  # Navigation
  go_router: ^13.0.0

  # Local Storage
  shared_preferences: ^2.2.2
  hive: ^2.2.3
  hive_flutter: ^1.1.0

  # File Handling
  file_picker: ^6.1.1
  path_provider: ^2.1.1

  # Utilities
  intl: ^0.18.1
  dartz: ^0.10.1
  connectivity_plus: ^5.0.2

dev_dependencies:
  flutter_test:
    sdk: flutter
  mocktail: ^1.0.1
  bloc_test: ^9.1.5
```

---

## API Client Generation Command

```bash
# Install openapi-generator
dart pub global activate openapi_generator_cli

# Generate API client
openapi-generator generate \
  -i agents/merged_api_spec.yaml \
  -g dart-dio \
  -o lib/generated/api \
  --additional-properties=pubName=pli_agent_api,useEnumExtension=true
```

---

## Environment Configuration

### Base URLs
- **Development**: `http://localhost:8080/api/v1`
- **Staging**: `https://staging-api.postallifeinsurance.gov.in/v1`
- **Production**: `https://api.postallifeinsurance.gov.in/v1`

### Current Status
- No authentication required (APIs are plain)
- Will be prepared for future JWT/OAuth integration

---

## Key Features Implementation Priority

### P0 (Critical) - Days 1-5
1. ✅ Project setup & API client generation
2. ✅ Agent Profile Creation (UJ-001)
3. ✅ License Management (UJ-003)
4. ✅ Bank Details Management (UJ-007)
5. ✅ Agent Search

### P1 (High) - Days 6-8
1. Agent Profile Update (UJ-002)
2. Self-Service Updates (UJ-006)
3. Goal Setting (UJ-008)
4. Portal Authentication UI (UJ-005)

### P2 (Medium) - Days 9-10
1. Agent Termination (UJ-004)
2. Status Reinstatement (UJ-009)
3. Export functionality (UJ-010)
4. Offline support

---

## Delivery Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| Phase 1 | Day 1 | Project setup + API client |
| Phase 2 | Day 2 | Core infrastructure (Domain + Data layers) |
| Phase 3 | Day 3 | State management setup |
| Phase 4 | Day 4-7 | All 10 user journeys UI |
| Phase 5 | Day 8 | Common features + Navigation |
| Phase 6 | Day 9 | Enhanced features |
| Phase 7 | Day 10 | Testing + Polish |

**Total**: 10 working days for complete Flutter app

---

## Success Criteria

1. ✅ All 78 APIs integrated
2. ✅ All 10 user journeys implemented
3. ✅ Form validations matching API specs
4. ✅ Clean architecture implemented
5. ✅ Responsive UI for mobile + tablet
6. ✅ Proper error handling
7. ✅ Loading and empty states
8. ✅ Ready for future auth integration

---

## Next Steps

1. Merge 3 OpenAPI YAML files
2. Generate API client
3. Create Flutter project structure
4. Implement Phase 1 tasks
5. Continue with subsequent phases

---

**Last Updated**: 2026-01-29
