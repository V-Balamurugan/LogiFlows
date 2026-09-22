import 'dart:async';

/// Abstract contract for secure credential and token persistence in mobile.
abstract class TokenStorage {
  Future<void> saveTokens({required String accessToken, String? refreshToken});
  Future<String?> getAccessToken();
  Future<String?> getRefreshToken();
  Future<void> clearTokens();
  Future<bool> hasValidToken();
}

/// In-memory and platform-fallback token storage implementation.
/// In production, this integrates with flutter_secure_storage / Android KeyStore / iOS Keychain.
class InMemorySecureTokenStorage implements TokenStorage {
  static final InMemorySecureTokenStorage _instance = InMemorySecureTokenStorage._internal();
  factory InMemorySecureTokenStorage() => _instance;
  InMemorySecureTokenStorage._internal();

  String? _accessToken;
  String? _refreshToken;

  @override
  Future<void> saveTokens({required String accessToken, String? refreshToken}) async {
    _accessToken = accessToken;
    if (refreshToken != null && refreshToken.isNotEmpty) {
      _refreshToken = refreshToken;
    }
  }

  @override
  Future<String?> getAccessToken() async {
    return _accessToken;
  }

  @override
  Future<String?> getRefreshToken() async {
    return _refreshToken;
  }

  @override
  Future<void> clearTokens() async {
    _accessToken = null;
    _refreshToken = null;
  }

  @override
  Future<bool> hasValidToken() async {
    return _accessToken != null && _accessToken!.isNotEmpty;
  }
}
