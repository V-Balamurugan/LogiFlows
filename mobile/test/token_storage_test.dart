import 'package:flutter_test/flutter_test.dart';
import 'package:logiflows_mobile/core/token_storage.dart';

void main() {
  group('Token Storage Unit Tests', () {
    late TokenStorage tokenStorage;

    setUp(() async {
      tokenStorage = InMemorySecureTokenStorage();
      await tokenStorage.clearTokens();
    });

    test('Initial storage state has no tokens and hasValidToken is false', () async {
      final hasToken = await tokenStorage.hasValidToken();
      final accessToken = await tokenStorage.getAccessToken();
      final refreshToken = await tokenStorage.getRefreshToken();

      expect(hasToken, isFalse);
      expect(accessToken, isNull);
      expect(refreshToken, isNull);
    });

    test('saveTokens persists accessToken and marks hasValidToken as true', () async {
      await tokenStorage.saveTokens(accessToken: 'sample-access-jwt-token');

      expect(await tokenStorage.hasValidToken(), isTrue);
      expect(await tokenStorage.getAccessToken(), 'sample-access-jwt-token');
      expect(await tokenStorage.getRefreshToken(), isNull);
    });

    test('saveTokens persists both access and refresh tokens when provided', () async {
      await tokenStorage.saveTokens(
        accessToken: 'access-12345',
        refreshToken: 'refresh-67890',
      );

      expect(await tokenStorage.hasValidToken(), isTrue);
      expect(await tokenStorage.getAccessToken(), 'access-12345');
      expect(await tokenStorage.getRefreshToken(), 'refresh-67890');
    });

    test('clearTokens removes all tokens and sets hasValidToken to false', () async {
      await tokenStorage.saveTokens(
        accessToken: 'access-to-clear',
        refreshToken: 'refresh-to-clear',
      );
      expect(await tokenStorage.hasValidToken(), isTrue);

      await tokenStorage.clearTokens();

      expect(await tokenStorage.hasValidToken(), isFalse);
      expect(await tokenStorage.getAccessToken(), isNull);
      expect(await tokenStorage.getRefreshToken(), isNull);
    });

    test('saveTokens overwrites existing tokens with newer credentials', () async {
      await tokenStorage.saveTokens(
        accessToken: 'old-access-token',
        refreshToken: 'old-refresh-token',
      );

      await tokenStorage.saveTokens(
        accessToken: 'new-rotated-token',
        refreshToken: 'new-refresh-token',
      );

      expect(await tokenStorage.getAccessToken(), 'new-rotated-token');
      expect(await tokenStorage.getRefreshToken(), 'new-refresh-token');
    });
  });
}
