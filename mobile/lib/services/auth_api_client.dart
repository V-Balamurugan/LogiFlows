import 'dart:convert';
import 'package:http/http.dart' as http;
import '../core/api_config.dart';
import '../core/token_storage.dart';
import '../models/auth_models.dart';

class AuthApiClient {
  final TokenStorage tokenStorage;
  final http.Client httpClient;

  AuthApiClient({
    TokenStorage? tokenStorage,
    http.Client? httpClient,
  })  : tokenStorage = tokenStorage ?? InMemorySecureTokenStorage(),
        httpClient = httpClient ?? http.Client();

  Map<String, String> _buildHeaders({String? token}) {
    final headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    };
    if (token != null && token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  /// Authenticates with user credentials, persists token pair, and returns AuthResponse.
  Future<AuthResponse> login({required String email, required String password}) async {
    final uri = Uri.parse(ApiConfig.loginEndpoint);
    final http.Response response;
    try {
      response = await httpClient.post(
        uri,
        headers: _buildHeaders(),
        body: jsonEncode({
          'email': email.trim().toLowerCase(),
          'password': password,
        }),
      );
    } catch (e) {
      if (e is http.ClientException ||
          e.toString().contains('ClientException') ||
          e.toString().contains('Failed to fetch') ||
          e.toString().contains('Connection refused') ||
          e.toString().contains('SocketException')) {
        throw Exception('Unable to connect to LogiFlows API at ${uri.host}:${uri.port}. Please verify the server is active.');
      }
      rethrow;
    }

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final authResp = AuthResponse.fromJson(data);

      await tokenStorage.saveTokens(
        accessToken: authResp.token,
        refreshToken: authResp.refreshToken,
      );

      return authResp;
    } else {
      final errorMsg = _extractErrorMessage(response);
      throw Exception(errorMsg);
    }
  }

  /// Onboards a new driver/operator along with their organization.
  Future<AuthResponse> register({
    required String email,
    required String password,
    required String fullName,
    required String companyName,
    String? phoneNumber,
  }) async {
    final uri = Uri.parse(ApiConfig.registerEndpoint);
    final http.Response response;
    try {
      response = await httpClient.post(
        uri,
        headers: _buildHeaders(),
        body: jsonEncode({
          'email': email.trim().toLowerCase(),
          'password': password,
          'full_name': fullName.trim(),
          'company_name': companyName.trim(),
          if (phoneNumber != null && phoneNumber.isNotEmpty) 'phone_number': phoneNumber.trim(),
        }),
      );
    } catch (e) {
      if (e is http.ClientException ||
          e.toString().contains('ClientException') ||
          e.toString().contains('Failed to fetch') ||
          e.toString().contains('Connection refused') ||
          e.toString().contains('SocketException')) {
        throw Exception('Unable to connect to LogiFlows API at ${uri.host}:${uri.port}. Please verify the server is active.');
      }
      rethrow;
    }

    if (response.statusCode == 201) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final authResp = AuthResponse.fromJson(data);

      await tokenStorage.saveTokens(
        accessToken: authResp.token,
        refreshToken: authResp.refreshToken,
      );

      return authResp;
    } else {
      final errorMsg = _extractErrorMessage(response);
      throw Exception(errorMsg);
    }
  }

  /// Retrieves the current profile using the saved access token.
  Future<AuthUser> getCurrentUser() async {
    final token = await tokenStorage.getAccessToken();
    if (token == null) {
      throw Exception('Unauthenticated: No access token found');
    }

    final uri = Uri.parse(ApiConfig.meEndpoint);
    final http.Response response;
    try {
      response = await httpClient.get(
        uri,
        headers: _buildHeaders(token: token),
      );
    } catch (e) {
      if (e is http.ClientException ||
          e.toString().contains('ClientException') ||
          e.toString().contains('Failed to fetch') ||
          e.toString().contains('Connection refused') ||
          e.toString().contains('SocketException')) {
        throw Exception('Unable to connect to LogiFlows API at ${uri.host}:${uri.port}. Please verify the server is active.');
      }
      rethrow;
    }

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      return AuthUser.fromJson(data['user'] as Map<String, dynamic>);
    } else if (response.statusCode == 401) {
      // Try refresh
      final refreshed = await refreshToken();
      return refreshed.user;
    } else {
      throw Exception('Failed to retrieve user profile');
    }
  }

  /// Performs single-use refresh token exchange.
  Future<AuthResponse> refreshToken() async {
    final rt = await tokenStorage.getRefreshToken();
    if (rt == null) {
      await tokenStorage.clearTokens();
      throw Exception('Session expired. Please log in again.');
    }

    final uri = Uri.parse(ApiConfig.refreshEndpoint);
    final http.Response response;
    try {
      response = await httpClient.post(
        uri,
        headers: _buildHeaders(),
        body: jsonEncode({'refresh_token': rt}),
      );
    } catch (e) {
      await tokenStorage.clearTokens();
      throw Exception('Unable to connect to LogiFlows API at ${uri.host}:${uri.port}. Please verify the server is active.');
    }

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final authResp = AuthResponse.fromJson(data);

      await tokenStorage.saveTokens(
        accessToken: authResp.token,
        refreshToken: authResp.refreshToken,
      );

      return authResp;
    } else {
      await tokenStorage.clearTokens();
      throw Exception('Session revoked. Please log in again.');
    }
  }

  /// Invalidate server-side session and clear client credentials.
  Future<void> logout() async {
    final token = await tokenStorage.getAccessToken();
    final rt = await tokenStorage.getRefreshToken();

    try {
      final uri = Uri.parse(ApiConfig.logoutEndpoint);
      await httpClient.post(
        uri,
        headers: _buildHeaders(token: token),
        body: jsonEncode({'refresh_token': rt ?? ''}),
      );
    } catch (_) {
      // Ignore network errors during logout
    } finally {
      await tokenStorage.clearTokens();
    }
  }

  String _extractErrorMessage(http.Response response) {
    try {
      final body = jsonDecode(response.body) as Map<String, dynamic>;
      if (body.containsKey('error') && body['error'] is Map) {
        final err = body['error'] as Map<String, dynamic>;
        return err['message'] as String? ?? 'Authentication request failed (${response.statusCode})';
      }
    } catch (_) {}
    return 'Authentication request failed (${response.statusCode})';
  }
}
