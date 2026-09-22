import 'dart:convert';
import 'package:http/http.dart' as http;
import '../core/api_config.dart';
import '../core/token_storage.dart';
import '../models/resource_models.dart';

class ResourceApiClient {
  final http.Client _httpClient;
  final TokenStorage tokenStorage;

  ResourceApiClient({
    http.Client? httpClient,
    TokenStorage? tokenStorage,
  })  : _httpClient = httpClient ?? http.Client(),
        tokenStorage = tokenStorage ?? InMemorySecureTokenStorage();

  Future<Map<String, String>> _authHeaders() async {
    final token = await tokenStorage.getAccessToken();
    return {
      'Content-Type': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  /// Fetches branches with optional spatial proximity filtering
  Future<List<BranchModel>> getBranches(
    String tenantId, {
    double? nearLat,
    double? nearLng,
    double? radiusKm,
    int limit = 50,
  }) async {
    final queryParams = <String, String>{
      'limit': limit.toString(),
      if (nearLat != null) 'near_lat': nearLat.toString(),
      if (nearLng != null) 'near_lng': nearLng.toString(),
      if (radiusKm != null) 'radius_km': radiusKm.toString(),
    };

    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/branches')
        .replace(queryParameters: queryParams);

    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final branchesList = (data['branches'] as List<dynamic>? ?? []);
      return branchesList
          .map((b) => BranchModel.fromJson(b as Map<String, dynamic>))
          .toList();
    } else {
      throw Exception('Failed to load branches: ${response.statusCode}');
    }
  }

  /// Fetches fleet vehicles with optional filtering
  Future<List<VehicleModel>> getVehicles(
    String tenantId, {
    String? vehicleType,
    String? status,
    String? branchId,
    int limit = 50,
  }) async {
    final queryParams = <String, String>{
      'limit': limit.toString(),
      if (vehicleType != null) 'vehicle_type': vehicleType,
      if (status != null) 'status': status,
      if (branchId != null) 'branch_id': branchId,
    };

    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/vehicles')
        .replace(queryParameters: queryParams);

    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final vehiclesList = (data['vehicles'] as List<dynamic>? ?? []);
      return vehiclesList
          .map((v) => VehicleModel.fromJson(v as Map<String, dynamic>))
          .toList();
    } else {
      throw Exception('Failed to load fleet vehicles: ${response.statusCode}');
    }
  }

  /// Assigns a driver to a vehicle
  Future<bool> assignVehicle(
    String tenantId,
    String vehicleId,
    String driverId, {
    String? notes,
  }) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/vehicles/$vehicleId/assign');
    final headers = await _authHeaders();
    final body = jsonEncode({
      'driver_id': driverId,
      if (notes != null) 'notes': notes,
    });

    final response = await _httpClient.post(uri, headers: headers, body: body);

    if (response.statusCode == 201) {
      return true;
    } else if (response.statusCode == 409) {
      throw Exception('Double-assignment conflict: driver or vehicle already assigned.');
    } else {
      throw Exception('Failed to assign vehicle: ${response.statusCode}');
    }
  }

  /// Unassigns a vehicle from its current driver
  Future<bool> unassignVehicle(String tenantId, String vehicleId) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/vehicles/$vehicleId/unassign');
    final headers = await _authHeaders();

    final response = await _httpClient.post(uri, headers: headers);
    return response.statusCode == 200;
  }

  /// Fetches company/tenant organization details
  Future<CompanyModel> getCompanyDetails(String tenantId) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId');
    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      return CompanyModel.fromJson(data);
    } else {
      throw Exception('Failed to load company details: ${response.statusCode}');
    }
  }
}
