# PLI Agent Management - Flutter Application

A comprehensive Flutter application for managing agent profiles in the Postal Life Insurance (PLI) system.

## 📱 Features

### Core Functionalities
- **Agent Profile Management** - Create, update, and view agent profiles
- **License Management** - Add, renew, and track agent licenses
- **Bank Details Management** - Secure bank account management
- **Goal Setting & Tracking** - Set and monitor performance goals
- **Status Management** - Handle agent termination and reinstatement
- **Search & Export** - Advanced search and data export capabilities
- **Audit Trail** - Complete audit history and timeline

### User Journeys (10 Complete Flows)
1. **UJ-001**: Agent Profile Creation with HRMS integration
2. **UJ-002**: Agent Profile Update with approval workflow
3. **UJ-003**: License Management & Renewal
4. **UJ-004**: Agent Termination
5. **UJ-005**: Portal Authentication (prepared for future)
6. **UJ-006**: Self-Service Profile Update
7. **UJ-007**: Bank Details Management
8. **UJ-008**: Agent Goal Setting
9. **UJ-009**: Status Reinstatement
10. **UJ-010**: Profile Search & Export

## 🏗️ Architecture

### Clean Architecture
```
lib/
├── core/           # Core utilities, config, theme
├── data/           # Data layer (API, models, repositories)
├── domain/         # Domain layer (entities, use cases)
├── presentation/   # UI layer (screens, widgets, BLoC)
└── generated/      # Generated API client
```

### State Management
- **BLoC Pattern** for state management
- **GetIt** for dependency injection
- **Dartz** for functional programming (Either type)

### Key Technologies
- Flutter 3.0+
- Dio for HTTP requests
- Freezed for immutable models
- JSON Serialization
- Go Router for navigation
- Hive for local storage
- Form Builder for forms

## 🚀 Getting Started

### Prerequisites
```bash
Flutter SDK >= 3.0.0
Dart SDK >= 3.0.0
```

### Installation

1. Clone the repository
```bash
git clone <repository-url>
cd flutter_app
```

2. Install dependencies
```bash
flutter pub get
```

3. Run code generation (for JSON serialization, BLoC, etc.)
```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

4. Run the app
```bash
# Development
flutter run --dart-define=ENVIRONMENT=development

# Staging
flutter run --dart-define=ENVIRONMENT=staging

# Production
flutter run --dart-define=ENVIRONMENT=production
```

## 🔧 Configuration

### Environment Setup

The app supports three environments:

1. **Development**: `http://localhost:8080/api/v1`
2. **Staging**: `https://staging-api.postallifeinsurance.gov.in/v1`
3. **Production**: `https://api.postallifeinsurance.gov.in/v1`

Set environment using `--dart-define`:
```bash
flutter run --dart-define=ENVIRONMENT=development
```

### API Configuration

Current Status: **Plain APIs** (No authentication required)

Future: Will support JWT/OAuth based authentication

## 📚 API Integration

### Total APIs: 78 Endpoints
Organized across 10 user journeys covering:
- Profile Management (38 APIs)
- Lookup/Dropdown (14 APIs)
- Validation (4 APIs)
- Status Tracking (10 APIs)
- Workflow Management (4 APIs)
- Approvals (3 APIs)
- Others (5 APIs)

### API Specifications
Located in: `../agents/*.yaml`
- Part 1: Profile creation, lookups, validations
- Part 2: Updates, licenses, termination, auth
- Part 3: Self-service, bank details, goals, export

## 🎨 UI/UX

### Design System
- **Colors**: PLI Brand colors (Blue & Gold)
- **Typography**: Roboto font family
- **Components**: Material Design 3
- **Responsive**: Mobile & Tablet support

### Status Colors
- Active: Green
- Suspended: Orange
- Terminated: Red
- Inactive: Grey

### SLA Indicators
- Green: On track
- Yellow: Warning
- Red: Breached

## 📦 Project Structure

```
lib/
├── core/
│   ├── config/          # App configuration
│   │   └── app_config.dart
│   ├── constants/       # Constants & endpoints
│   │   └── api_endpoints.dart
│   ├── errors/          # Error handling
│   │   ├── failures.dart
│   │   └── exceptions.dart
│   ├── network/         # API client
│   │   └── api_client.dart
│   ├── theme/           # App theme
│   │   ├── app_theme.dart
│   │   └── app_colors.dart
│   └── utils/           # Utilities
│       └── validators.dart
├── data/
│   ├── datasources/     # Remote data sources
│   ├── models/          # Data models
│   └── repositories/    # Repository implementations
├── domain/
│   ├── entities/        # Domain entities
│   ├── repositories/    # Repository interfaces
│   └── usecases/        # Business logic
├── presentation/
│   ├── blocs/           # BLoC state management
│   ├── screens/         # UI screens
│   │   ├── home/
│   │   ├── agent_creation/
│   │   ├── agent_profile/
│   │   ├── license_management/
│   │   ├── bank_details/
│   │   ├── goals/
│   │   └── search/
│   └── widgets/         # Reusable widgets
│       ├── common/
│       ├── forms/
│       └── cards/
└── injection_container.dart
```

## ✅ Validation Rules

### PAN
- Format: AAAAA9999A (5 letters, 4 digits, 1 letter)
- Uniqueness check required

### Mobile
- 10 digits
- Must start with 6, 7, 8, or 9

### Aadhar
- 12 digits
- Optional field

### IFSC
- Format: AAAA0NNNNNN
- Validated against bank database

### Age
- Minimum: 18 years
- Maximum: 70 years

### Pincode
- 6 digits numeric

## 🧪 Testing

### Run Tests
```bash
# Unit tests
flutter test

# Widget tests
flutter test test/widget

# Integration tests
flutter test integration_test
```

### Test Coverage
```bash
flutter test --coverage
genhtml coverage/lcov.info -o coverage/html
```

## 📱 Build & Release

### Android
```bash
flutter build apk --release
flutter build appbundle --release
```

### iOS
```bash
flutter build ios --release
```

### Web
```bash
flutter build web --release
```

## 🔐 Security

### Data Encryption
- Bank account numbers are masked
- Sensitive data encrypted at rest
- HTTPS for all API communications

### Validation
- Client-side validation for all forms
- Server-side validation enforced
- PAN uniqueness check
- IFSC code verification

## 📖 Documentation

### Code Documentation
```bash
dart doc .
```

### API Documentation
See `../agents/*.yaml` for complete API specifications

## 🤝 Contributing

1. Follow clean architecture principles
2. Write unit tests for new features
3. Use BLoC for state management
4. Follow Flutter best practices
5. Add validation for all form fields

## 📄 License

Postal Life Insurance (Government of India)

## 📞 Support

For issues and support:
- Email: support@postallifeinsurance.gov.in
- Create an issue in the repository

## 🗺️ Roadmap

### Phase 1 (Current) ✅
- Project setup
- Core infrastructure
- API client integration
- Basic UI screens

### Phase 2 (In Progress)
- All 10 user journeys implementation
- State management setup
- Form validations
- Navigation

### Phase 3 (Planned)
- Offline support
- Push notifications
- Advanced search
- Dashboard analytics

### Phase 4 (Future)
- Authentication integration
- Role-based access
- Advanced reporting
- Multi-language support

## 📊 Technical Debt

- [ ] API client generation from OpenAPI specs
- [ ] Complete test coverage
- [ ] Offline sync implementation
- [ ] Performance optimization
- [ ] Accessibility improvements

## 🎯 Performance

- Target: 60 FPS rendering
- Cold start: < 3 seconds
- API response: < 2 seconds
- Memory usage: < 200 MB

---

**Version**: 1.0.0
**Last Updated**: 2026-01-29
**Maintained by**: PLI Development Team
