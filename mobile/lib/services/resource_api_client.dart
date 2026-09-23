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

  /// Fetches employees with optional filters
  Future<List<EmployeeModel>> getEmployees(
    String tenantId, {
    String? operationalRole,
    String? availabilityStatus,
    String? branchId,
    String? status,
    int limit = 50,
  }) async {
    final queryParams = <String, String>{
      'limit': limit.toString(),
      if (operationalRole != null) 'operational_role': operationalRole,
      if (availabilityStatus != null) 'availability_status': availabilityStatus,
      if (branchId != null) 'branch_id': branchId,
      if (status != null) 'status': status,
    };

    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/employees')
        .replace(queryParameters: queryParams);

    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final employeesList = (data['employees'] as List<dynamic>? ?? []);
      return employeesList
          .map((e) => EmployeeModel.fromJson(e as Map<String, dynamic>))
          .toList();
    } else {
      throw Exception('Failed to load employees: ${response.statusCode}');
    }
  }

  /// Fetches verified, active drivers currently available for assignment
  Future<List<EmployeeModel>> getAvailableDrivers(
    String tenantId, {
    String? branchId,
  }) async {
    final queryParams = <String, String>{
      if (branchId != null) 'branch_id': branchId,
    };

    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/employees/available-drivers')
        .replace(queryParameters: queryParams.isNotEmpty ? queryParams : null);

    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final driversList = (data['drivers'] as List<dynamic>? ?? []);
      return driversList
          .map((d) => EmployeeModel.fromJson(d as Map<String, dynamic>))
          .toList();
    } else {
      throw Exception('Failed to load available drivers: ${response.statusCode}');
    }
  }

  /// Updates employee operational and availability status
  Future<bool> updateEmployeeStatus(
    String tenantId,
    String employeeId, {
    String? availabilityStatus,
    String? status,
    String? verificationStatus,
  }) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/employees/$employeeId/status');
    final headers = await _authHeaders();
    final body = jsonEncode({
      if (availabilityStatus != null) 'availability_status': availabilityStatus,
      if (status != null) 'status': status,
      if (verificationStatus != null) 'verification_status': verificationStatus,
    });

    final response = await _httpClient.patch(uri, headers: headers, body: body);
    return response.statusCode == 200;
  }

  /// Updates vehicle operational status and availability
  Future<bool> updateVehicleStatus(
    String tenantId,
    String vehicleId, {
    String? status,
    String? availabilityStatus,
  }) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/vehicles/$vehicleId/status');
    final headers = await _authHeaders();
    final body = jsonEncode({
      if (status != null) 'status': status,
      if (availabilityStatus != null) 'availability_status': availabilityStatus,
    });

    final response = await _httpClient.patch(uri, headers: headers, body: body);
    return response.statusCode == 200;
  }

  /// Fetches assignment history
  Future<List<AssignmentModel>> getAssignments(
    String tenantId, {
    String? vehicleId,
    String? employeeId,
    String? status,
    int limit = 50,
  }) async {
    final queryParams = <String, String>{
      'limit': limit.toString(),
      if (vehicleId != null) 'vehicle_id': vehicleId,
      if (employeeId != null) 'employee_id': employeeId,
      if (status != null) 'status': status,
    };

    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/assignments')
        .replace(queryParameters: queryParams);

    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final assignmentsList = (data['assignments'] as List<dynamic>? ?? []);
      return assignmentsList
          .map((a) => AssignmentModel.fromJson(a as Map<String, dynamic>))
          .toList();
    } else {
      throw Exception('Failed to load assignments: ${response.statusCode}');
    }
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

  /// Fetches the authenticated employee's self profile, role, and assigned vehicle
  Future<EmployeeMeModel> getMyProfile(String tenantId) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/employees/me');
    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      return EmployeeMeModel.fromJson(data);
    } else {
      throw Exception('Failed to load self-profile: ${response.statusCode}');
    }
  }

  /// Fetches all employees assigned to a specific branch hub
  Future<List<BranchEmployeeSummaryModel>> getBranchEmployees(
    String tenantId,
    String branchId,
  ) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/branches/$branchId/employees');
    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final list = (data['employees'] as List<dynamic>? ?? []);
      return list
          .map((e) => BranchEmployeeSummaryModel.fromJson(e as Map<String, dynamic>))
          .toList();
    } else {
      throw Exception('Failed to load branch employees: ${response.statusCode}');
    }
  }

  /// Fetches all vehicles stationed at a specific branch hub
  Future<List<BranchVehicleSummaryModel>> getBranchVehicles(
    String tenantId,
    String branchId,
  ) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/branches/$branchId/vehicles');
    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      final list = (data['vehicles'] as List<dynamic>? ?? []);
      return list
          .map((v) => BranchVehicleSummaryModel.fromJson(v as Map<String, dynamic>))
          .toList();
    } else {
      throw Exception('Failed to load branch vehicles: ${response.statusCode}');
    }
  }

  /// Fetches account link and credential status for an employee
  Future<EmployeeAccountStatusModel> getEmployeeAccountStatus(
    String tenantId,
    String employeeId,
  ) async {
    final uri = Uri.parse('${ApiConfig.baseUrl}/tenants/$tenantId/employees/$employeeId/account-status');
    final headers = await _authHeaders();
    final response = await _httpClient.get(uri, headers: headers);

    if (response.statusCode == 200) {
      final jsonBody = jsonDecode(response.body) as Map<String, dynamic>;
      final data = jsonBody['data'] as Map<String, dynamic>;
      return EmployeeAccountStatusModel.fromJson(data);
    } else {
      throw Exception('Failed to load account status: ${response.statusCode}');
    }
  }
}
