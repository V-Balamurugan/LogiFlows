import 'package:flutter/foundation.dart';

class ApiConfig {
  static String? _customHost;

  /// Explicitly set custom host (useful for physical device testing, custom LAN IP, or unit tests).
  static void setCustomHost(String? host) {
    _customHost = host;
  }

  /// Automatically resolves the host for the current runtime platform:
  /// - Web browser (kIsWeb): uses browser's active host (e.g. 'localhost' or LAN IP).
  /// - Android emulator: '10.0.2.2' (loopback alias to host machine).
  /// - Windows / macOS / Linux desktop or iOS simulator: 'localhost'.
  static String get defaultHost {
    if (_customHost != null && _customHost!.isNotEmpty) {
      return _customHost!;
    }
    if (kIsWeb) {
      try {
        final host = Uri.base.host;
        if (host.isNotEmpty && host != '0.0.0.0') {
          return host;
        }
      } catch (_) {}
      return 'localhost';
    }
    if (defaultTargetPlatform == TargetPlatform.android) {
      return '10.0.2.2';
    }
    return 'localhost';
  }

  static String get baseUrl => 'http://$defaultHost:8080/api/v1';
  static const String fallbackLocalhostUrl = 'http://localhost:8080/api/v1';
  static String get aiServiceUrl => 'http://$defaultHost:8000/api/v1';

  // Observability
  static String get healthEndpoint => '$baseUrl/health';
  static String get readinessEndpoint => '$baseUrl/readiness';

  // Authentication & Identity
  static String get loginEndpoint => '$baseUrl/auth/login';
  static String get registerEndpoint => '$baseUrl/auth/register';
  static String get meEndpoint => '$baseUrl/auth/me';
  static String get refreshEndpoint => '$baseUrl/auth/refresh';
  static String get logoutEndpoint => '$baseUrl/auth/logout';

  // AI Inference
  static String get predictDelayEndpoint => '$aiServiceUrl/predict/delay-risk';
}
