# Flutter App Setup Instructions

## Prerequisites

Before running the Flutter app, you need to have Flutter installed on your system.

## Installation Steps

### 1. Install Flutter SDK

Visit: https://docs.flutter.dev/get-started/install

Or use these quick commands:

**macOS/Linux:**
```bash
# Download Flutter SDK
git clone https://github.com/flutter/flutter.git -b stable
export PATH="$PATH:`pwd`/flutter/bin"

# Verify installation
flutter doctor
```

**Windows:**
Download from: https://docs.flutter.dev/get-started/install/windows

---

## Setup This Project

### Step 1: Install Dependencies

Navigate to the flutter_app directory and run:

```bash
cd /home/user/pli-agent/flutter_app
flutter pub get
```

**What this does**: Downloads all packages listed in `pubspec.yaml`

**Expected output**:
```
Running "flutter pub get" in flutter_app...
Resolving dependencies...
+ flutter_bloc 8.1.3
+ dio 5.4.0
+ freezed 2.4.5
...
Got dependencies!
```

---

### Step 2: Generate Code (for Freezed models)

Since we're using Freezed for immutable models, we need to generate code:

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

**What this does**:
- Generates `.freezed.dart` files for our request/response models
- Generates `.g.dart` files for JSON serialization

**Expected output**:
```
[INFO] Generating build script...
[INFO] Generating build script completed, took 2.3s
[INFO] Creating build script snapshot...
[INFO] Creating build script snapshot completed, took 5.1s
[INFO] Initializing inputs
[INFO] Building new asset graph...
[INFO] Building new asset graph completed, took 1.2s
[INFO] Checking for unexpected pre-existing outputs...
[INFO] Checking for unexpected pre-existing outputs completed, took 0.0s
[INFO] Running build...
[INFO] Running build completed, took 3.5s
[INFO] Caching finalized dependency graph...
[INFO] Caching finalized dependency graph completed, took 0.1s
[INFO] Succeeded after 3.6s with 6 outputs
```

**Generated files**:
- `lib/data/models/requests/agent/create_agent_request.freezed.dart`
- `lib/data/models/requests/agent/create_agent_request.g.dart`
- `lib/data/models/responses/agent/agent_profile_response.freezed.dart`
- `lib/data/models/responses/agent/agent_profile_response.g.dart`

---

### Step 3: Verify Setup

Check that everything is configured correctly:

```bash
flutter doctor
```

**Expected output**:
```
Doctor summary (to see all details, run flutter doctor -v):
[✓] Flutter (Channel stable, 3.x.x, on macOS/Linux/Windows)
[✓] Android toolchain - develop for Android devices
[✓] Xcode - develop for iOS and macOS (if on macOS)
[✓] Chrome - develop for the web
[✓] Android Studio (version 2023.x)
[✓] VS Code (version 1.x.x)
[✓] Connected device (1 available)
```

---

### Step 4: Run the App

Choose your target platform:

**Web (easiest for testing):**
```bash
flutter run -d chrome --dart-define=ENVIRONMENT=development
```

**Android Emulator:**
```bash
flutter run -d android --dart-define=ENVIRONMENT=development
```

**iOS Simulator (macOS only):**
```bash
flutter run -d ios --dart-define=ENVIRONMENT=development
```

**Physical Device:**
```bash
# List available devices
flutter devices

# Run on specific device
flutter run -d <device-id> --dart-define=ENVIRONMENT=development
```

---

## Alternative: Watch Mode for Development

If you'll be making frequent changes to models, use watch mode:

```bash
# Terminal 1: Watch for model changes and auto-generate
flutter pub run build_runner watch

# Terminal 2: Run the app
flutter run -d chrome --dart-define=ENVIRONMENT=development
```

This automatically regenerates code whenever you modify a model file.

---

## Troubleshooting

### Issue: "flutter: command not found"

**Solution**: Add Flutter to your PATH:

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:/path/to/flutter/bin"

# Reload shell
source ~/.bashrc  # or source ~/.zshrc
```

### Issue: "Could not find a file named 'pubspec.yaml'"

**Solution**: Make sure you're in the flutter_app directory:

```bash
cd /home/user/pli-agent/flutter_app
```

### Issue: "Version solving failed"

**Solution**: Update Flutter and try again:

```bash
flutter upgrade
flutter pub get
```

### Issue: Build runner fails

**Solution**: Clean and rebuild:

```bash
flutter pub run build_runner clean
flutter pub run build_runner build --delete-conflicting-outputs
```

### Issue: "No devices found"

**Solution**:

For Web:
```bash
flutter run -d web-server
```

For Android:
```bash
# Start Android emulator first
# Then run
flutter run
```

---

## Project Structure Check

After running `flutter pub get`, your project should have:

```
flutter_app/
├── .dart_tool/           ← Created by pub get
├── .packages             ← Created by pub get
├── pubspec.lock          ← Created by pub get
├── lib/
│   └── (your code)
├── pubspec.yaml
└── README.md
```

After running `build_runner`, you should also have:

```
lib/data/models/
├── requests/
│   └── agent/
│       ├── create_agent_request.dart
│       ├── create_agent_request.freezed.dart  ← Generated
│       └── create_agent_request.g.dart        ← Generated
└── responses/
    └── agent/
        ├── agent_profile_response.dart
        ├── agent_profile_response.freezed.dart  ← Generated
        └── agent_profile_response.g.dart        ← Generated
```

---

## Environment Configuration

The app supports three environments:

### Development (default)
```bash
flutter run --dart-define=ENVIRONMENT=development
# Uses: http://localhost:8080/v1
```

### Staging
```bash
flutter run --dart-define=ENVIRONMENT=staging
# Uses: https://staging-api.postallifeinsurance.gov.in/v1
```

### Production
```bash
flutter run --dart-define=ENVIRONMENT=production
# Uses: https://api.postallifeinsurance.gov.in/v1
```

---

## Quick Reference Commands

```bash
# Install dependencies
flutter pub get

# Generate code
flutter pub run build_runner build --delete-conflicting-outputs

# Watch mode (auto-generate)
flutter pub run build_runner watch

# Clean generated files
flutter pub run build_runner clean

# Run app (development)
flutter run -d chrome --dart-define=ENVIRONMENT=development

# Run app (production)
flutter run -d chrome --dart-define=ENVIRONMENT=production

# Check Flutter setup
flutter doctor

# List available devices
flutter devices

# Analyze code
flutter analyze

# Format code
flutter format lib/

# Run tests
flutter test

# Build for production
flutter build apk --release          # Android
flutter build ios --release          # iOS
flutter build web --release          # Web
```

---

## VS Code Integration

If using VS Code, install these extensions:

1. **Flutter** (Dart-Code.flutter)
2. **Dart** (Dart-Code.dart-code)

Then you can:
- Press `F5` to run the app
- Use command palette (`Cmd/Ctrl + Shift + P`) → "Flutter: Run Flutter Doctor"
- Right-click on a file → "Flutter: New ..." to create widgets/screens

---

## Android Studio Integration

If using Android Studio:

1. Install Flutter plugin
2. Open flutter_app folder
3. Click "Get Dependencies" when prompted
4. Click Run button (green triangle)

---

## Next Steps After Setup

1. ✅ Run `flutter pub get`
2. ✅ Run `flutter pub run build_runner build`
3. ✅ Verify with `flutter doctor`
4. ✅ Run the app with `flutter run`
5. 📖 Read `CONFIGURATION_GUIDE.md` for configuration
6. 📖 Read `MODEL_STRUCTURE_GUIDE.md` for models
7. 🚀 Start developing!

---

## Support

If you encounter issues:

1. Check `flutter doctor -v` for detailed diagnostics
2. Check Flutter logs: Run with `-v` flag for verbose output
3. Check this guide's Troubleshooting section
4. Consult Flutter documentation: https://docs.flutter.dev

---

## Summary

**Minimum Required Steps**:
```bash
# 1. Navigate to project
cd /home/user/pli-agent/flutter_app

# 2. Install dependencies
flutter pub get

# 3. Generate code
flutter pub run build_runner build --delete-conflicting-outputs

# 4. Run app
flutter run -d chrome --dart-define=ENVIRONMENT=development
```

**That's it! You're ready to go! 🚀**

---

**Created**: 2026-01-29
**Flutter Version**: Compatible with Flutter 3.0+
**Dart Version**: Compatible with Dart 3.0+
