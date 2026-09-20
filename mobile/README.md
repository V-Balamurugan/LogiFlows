# LogiFlows Mobile Application

> Flutter & Dart cross-platform mobile client for delivery drivers, field couriers, and branch custody transfers.

---

## Architecture
- **Framework**: Flutter (Material 3)
- **Language**: Dart 3.x
- **Target OS**: Android, iOS, Web
- **State Management Pattern**: Riverpod / Bloc (Prepared)

---

## Directory Structure
```
mobile/
├── lib/
│   ├── main.dart             # Application root and Driver Operations screen
│   └── core/
│       └── api_config.dart   # Centralized endpoints for Backend and AI Service
├── test/
│   └── widget_test.dart      # Widget interaction unit test
├── pubspec.yaml              # Package dependencies and asset manifests
└── README.md
```

---

## Running on a Machine with Flutter Installed

```bash
# 1. Get packages
flutter pub get

# 2. Run unit & widget tests
flutter test

# 3. Launch on Android Emulator, iOS Simulator, or Chrome
flutter run
```
