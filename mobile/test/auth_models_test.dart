import 'package:flutter_test/flutter_test.dart';
import 'package:logiflows_mobile/models/auth_models.dart';

void main() {
  group('Auth Models Unit Tests', () {
    test('AuthUser deserializes correctly from complete JSON payload', () {
      final json = {
        'id': 'u-123e4567-e89b-12d3-a456-426614174000',
        'email': 'driver@logiflows.com',
        'full_name': 'Ramesh Kumar',
        'phone_number': '+919876543210',
        'is_active': true,
        'is_platform_admin': false,
        'email_verified': true,
      };

      final user = AuthUser.fromJson(json);

      expect(user.id, 'u-123e4567-e89b-12d3-a456-426614174000');
      expect(user.email, 'driver@logiflows.com');
      expect(user.fullName, 'Ramesh Kumar');
      expect(user.phoneNumber, '+919876543210');
      expect(user.isActive, isTrue);
      expect(user.isPlatformAdmin, isFalse);
      expect(user.emailVerified, isTrue);
    });

    test('AuthUser applies safe defaults when optional boolean flags are omitted', () {
      final json = {
        'id': 'u-simple',
        'email': 'minimal@logiflows.com',
        'full_name': 'Minimal Driver',
        'phone_number': null,
      };

      final user = AuthUser.fromJson(json);

      expect(user.id, 'u-simple');
      expect(user.phoneNumber, isNull);
      expect(user.isActive, isTrue);
      expect(user.isPlatformAdmin, isFalse);
      expect(user.emailVerified, isFalse);
    });

    test('Tenant deserializes correctly from JSON payload', () {
      final json = {
        'id': 't-999',
        'name': 'Speedy Logistics India',
        'slug': 'speedy-logistics',
        'role': 'DRIVER',
      };

      final tenant = Tenant.fromJson(json);

      expect(tenant.id, 't-999');
      expect(tenant.name, 'Speedy Logistics India');
      expect(tenant.slug, 'speedy-logistics');
      expect(tenant.role, 'DRIVER');
    });

    test('AuthResponse deserializes nested user, tenants, and tokens', () {
      final json = {
        'token': 'jwt.header.payload.signature',
        'expires_at': '2026-09-22T21:00:00Z',
        'refresh_token': 'refresh-secret-uuid',
        'refresh_token_expires_at': '2026-09-29T21:00:00Z',
        'user': {
          'id': 'u-100',
          'email': 'driver1@logiflows.com',
          'full_name': 'Driver One',
          'phone_number': '+911234567890',
          'is_active': true,
          'is_platform_admin': false,
          'email_verified': true,
        },
        'tenants': [
          {
            'id': 't-1',
            'name': 'Primary Carrier',
            'slug': 'primary-carrier',
            'role': 'DRIVER',
          },
          {
            'id': 't-2',
            'name': 'Secondary Fleet',
            'slug': 'secondary-fleet',
            'role': 'OPERATOR',
          }
        ],
      };

      final response = AuthResponse.fromJson(json);

      expect(response.token, 'jwt.header.payload.signature');
      expect(response.expiresAt, '2026-09-22T21:00:00Z');
      expect(response.refreshToken, 'refresh-secret-uuid');
      expect(response.refreshTokenExpiresAt, '2026-09-29T21:00:00Z');
      expect(response.user.fullName, 'Driver One');
      expect(response.tenants.length, 2);
      expect(response.tenants[0].slug, 'primary-carrier');
      expect(response.tenants[1].role, 'OPERATOR');
    });

    test('AuthResponse safely handles empty or null tenants array', () {
      final json = {
        'token': 'jwt.minimal.token',
        'expires_at': '2026-09-22T21:00:00Z',
        'user': {
          'id': 'u-200',
          'email': 'solo@logiflows.com',
          'full_name': 'Solo User',
          'is_active': true,
          'is_platform_admin': false,
          'email_verified': false,
        },
        'tenants': null,
      };

      final response = AuthResponse.fromJson(json);

      expect(response.token, 'jwt.minimal.token');
      expect(response.refreshToken, isNull);
      expect(response.tenants, isEmpty);
    });
  });
}
