import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:logiflows_mobile/core/api_config.dart';

void main() {
  group('ApiConfig Unit Tests (Phase 0 Foundation & Cross-Platform Host Resolution)', () {
    tearDown(() {
      ApiConfig.setCustomHost(null);
      debugDefaultTargetPlatformOverride = null;
    });

    test('TC-P0-MOB-001: ApiConfig provides correct fallback and AI service port configurations', () {
      expect(ApiConfig.fallbackLocalhostUrl, 'http://localhost:8080/api/v1');
      expect(ApiConfig.aiServiceUrl.contains(':8000/api/v1'), isTrue);
      expect(ApiConfig.predictDelayEndpoint.contains(':8000/api/v1/predict/delay-risk'), isTrue);
    });

    test('TC-P0-MOB-008: Platform-aware host resolution dynamically selects 10.0.2.2 for Android and localhost for Desktop', () {
      // Test Android emulator resolution
      debugDefaultTargetPlatformOverride = TargetPlatform.android;
      expect(ApiConfig.defaultHost, '10.0.2.2');
      expect(ApiConfig.baseUrl, 'http://10.0.2.2:8080/api/v1');
      expect(ApiConfig.loginEndpoint, 'http://10.0.2.2:8080/api/v1/auth/login');

      // Test Windows desktop resolution
      debugDefaultTargetPlatformOverride = TargetPlatform.windows;
      expect(ApiConfig.defaultHost, 'localhost');
      expect(ApiConfig.baseUrl, 'http://localhost:8080/api/v1');
      expect(ApiConfig.loginEndpoint, 'http://localhost:8080/api/v1/auth/login');

      // Test iOS simulator resolution
      debugDefaultTargetPlatformOverride = TargetPlatform.iOS;
      expect(ApiConfig.defaultHost, 'localhost');
      expect(ApiConfig.baseUrl, 'http://localhost:8080/api/v1');
    });

    test('TC-P0-MOB-008: setCustomHost overrides dynamic platform resolution', () {
      ApiConfig.setCustomHost('192.168.1.120');

      expect(ApiConfig.defaultHost, '192.168.1.120');
      expect(ApiConfig.baseUrl, 'http://192.168.1.120:8080/api/v1');
      expect(ApiConfig.healthEndpoint, 'http://192.168.1.120:8080/api/v1/health');
      expect(ApiConfig.readinessEndpoint, 'http://192.168.1.120:8080/api/v1/readiness');
      expect(ApiConfig.aiServiceUrl, 'http://192.168.1.120:8000/api/v1');

      // Reset
      ApiConfig.setCustomHost(null);
      expect(ApiConfig.defaultHost != '192.168.1.120', isTrue);
    });

    test('TC-P1-MOB-001: ApiConfig generates consistent endpoints for identity and authentication', () {
      ApiConfig.setCustomHost('localhost');

      expect(ApiConfig.loginEndpoint, 'http://localhost:8080/api/v1/auth/login');
      expect(ApiConfig.registerEndpoint, 'http://localhost:8080/api/v1/auth/register');
      expect(ApiConfig.meEndpoint, 'http://localhost:8080/api/v1/auth/me');
      expect(ApiConfig.refreshEndpoint, 'http://localhost:8080/api/v1/auth/refresh');
      expect(ApiConfig.logoutEndpoint, 'http://localhost:8080/api/v1/auth/logout');
    });
  });
}
