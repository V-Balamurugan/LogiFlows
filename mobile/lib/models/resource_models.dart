class BranchModel {
  final String id;
  final String tenantId;
  final String branchCode;
  final String name;
  final String address;
  final String city;
  final String? state;
  final String? postalCode;
  final String country;
  final double latitude;
  final double longitude;
  final double coverageRadiusKm;
  final String status;
  final bool isActive;
  final double? distanceKm;

  BranchModel({
    required this.id,
    required this.tenantId,
    required this.branchCode,
    required this.name,
    required this.address,
    required this.city,
    this.state,
    this.postalCode,
    required this.country,
    required this.latitude,
    required this.longitude,
    required this.coverageRadiusKm,
    required this.status,
    required this.isActive,
    this.distanceKm,
  });

  factory BranchModel.fromJson(Map<String, dynamic> json) {
    return BranchModel(
      id: json['id'] as String,
      tenantId: json['tenant_id'] as String,
      branchCode: json['branch_code'] as String,
      name: json['name'] as String,
      address: json['address'] as String? ?? json['address_line1'] as String? ?? '',
      city: json['city'] as String,
      state: json['state'] as String?,
      postalCode: json['postal_code'] as String?,
      country: json['country'] as String? ?? 'India',
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
      coverageRadiusKm: (json['coverage_radius_km'] as num?)?.toDouble() ?? 15.0,
      status: json['status'] as String? ?? 'ACTIVE',
      isActive: json['is_active'] as bool? ?? true,
      distanceKm: (json['distance_km'] as num?)?.toDouble(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'tenant_id': tenantId,
      'branch_code': branchCode,
      'name': name,
      'address': address,
      'city': city,
      if (state != null) 'state': state,
      if (postalCode != null) 'postal_code': postalCode,
      'country': country,
      'latitude': latitude,
      'longitude': longitude,
      'coverage_radius_km': coverageRadiusKm,
      'status': status,
      'is_active': isActive,
      if (distanceKm != null) 'distance_km': distanceKm,
    };
  }
}

class EmployeeModel {
  final String id;
  final String tenantId;
  final String? branchId;
  final String? branchName;
  final String employeeCode;
  final String firstName;
  final String lastName;
  final String? email;
  final String? phone;
  final String designation;
  final String operationalRole;
  final String? licenseNumber;
  final String status;
  final bool isActive;

  EmployeeModel({
    required this.id,
    required this.tenantId,
    this.branchId,
    this.branchName,
    required this.employeeCode,
    required this.firstName,
    required this.lastName,
    this.email,
    this.phone,
    required this.designation,
    required this.operationalRole,
    this.licenseNumber,
    required this.status,
    required this.isActive,
  });

  String get fullName => '$firstName $lastName'.trim();
  String get role => operationalRole;

  factory EmployeeModel.fromJson(Map<String, dynamic> json) {
    final rawFullName = json['full_name'] as String? ?? '';
    final nameParts = rawFullName.trim().split(RegExp(r'\s+'));
    final parsedFirst = json['first_name'] as String? ?? (nameParts.isNotEmpty ? nameParts.first : '');
    final parsedLast = json['last_name'] as String? ?? (nameParts.length > 1 ? nameParts.sublist(1).join(' ') : '');

    return EmployeeModel(
      id: json['id'] as String,
      tenantId: json['tenant_id'] as String,
      branchId: json['branch_id'] as String?,
      branchName: json['branch_name'] as String?,
      employeeCode: json['employee_code'] as String,
      firstName: parsedFirst,
      lastName: parsedLast,
      email: json['email'] as String?,
      phone: json['phone'] as String?,
      designation: json['designation'] as String? ?? 'Staff',
      operationalRole: json['operational_role'] as String? ?? json['role'] as String? ?? 'OPERATOR',
      licenseNumber: json['license_number'] as String?,
      status: json['status'] as String? ?? 'ACTIVE',
      isActive: json['is_active'] as bool? ?? true,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'tenant_id': tenantId,
      if (branchId != null) 'branch_id': branchId,
      if (branchName != null) 'branch_name': branchName,
      'employee_code': employeeCode,
      'first_name': firstName,
      'last_name': lastName,
      if (email != null) 'email': email,
      if (phone != null) 'phone': phone,
      'designation': designation,
      'operational_role': operationalRole,
      if (licenseNumber != null) 'license_number': licenseNumber,
      'status': status,
      'is_active': isActive,
    };
  }
}

class VehicleModel {
  final String id;
  final String tenantId;
  final String? assignedBranchId;
  final String? branchName;
  final String registrationNumber;
  final String vehicleType;
  final String? makeModel;
  final int? year;
  final double maxWeightKg;
  final double maxVolumeCbm;
  final String status;
  final bool isActive;
  final bool isElectric;
  final String? currentDriverName;
  final String? currentDriverId;

  VehicleModel({
    required this.id,
    required this.tenantId,
    this.assignedBranchId,
    this.branchName,
    required this.registrationNumber,
    required this.vehicleType,
    this.makeModel,
    this.year,
    required this.maxWeightKg,
    required this.maxVolumeCbm,
    required this.status,
    required this.isActive,
    bool? isElectric,
    this.currentDriverName,
    this.currentDriverId,
  }) : isElectric = isElectric ?? (vehicleType == 'ELECTRIC_VAN');

  bool get isAvailable => status == 'AVAILABLE';
  bool get isAssigned => status == 'ASSIGNED';

  factory VehicleModel.fromJson(Map<String, dynamic> json) {
    final vType = json['vehicle_type'] as String? ?? 'VAN';
    final electricVal = json['is_electric'] as bool? ?? (vType == 'ELECTRIC_VAN');

    return VehicleModel(
      id: json['id'] as String,
      tenantId: json['tenant_id'] as String,
      assignedBranchId: json['assigned_branch_id'] as String?,
      branchName: json['branch_name'] as String?,
      registrationNumber: json['registration_number'] as String,
      vehicleType: vType,
      makeModel: json['make_model'] as String?,
      year: json['year'] as int?,
      maxWeightKg: (json['max_weight_kg'] as num?)?.toDouble() ?? 500.0,
      maxVolumeCbm: (json['max_volume_cbm'] as num?)?.toDouble() ?? 3.0,
      status: json['status'] as String? ?? 'AVAILABLE',
      isActive: json['is_active'] as bool? ?? true,
      isElectric: electricVal,
      currentDriverName: json['current_driver_name'] as String?,
      currentDriverId: json['current_driver_id'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'tenant_id': tenantId,
      if (assignedBranchId != null) 'assigned_branch_id': assignedBranchId,
      if (branchName != null) 'branch_name': branchName,
      'registration_number': registrationNumber,
      'vehicle_type': vehicleType,
      if (makeModel != null) 'make_model': makeModel,
      if (year != null) 'year': year,
      'max_weight_kg': maxWeightKg,
      'max_volume_cbm': maxVolumeCbm,
      'status': status,
      'is_active': isActive,
      if (currentDriverName != null) 'current_driver_name': currentDriverName,
      if (currentDriverId != null) 'current_driver_id': currentDriverId,
    };
  }
}

class CompanyModel {
  final String id;
  final String name;
  final String slug;
  final String status;
  final String? contactEmail;
  final String? createdAt;

  CompanyModel({
    required this.id,
    required this.name,
    required this.slug,
    required this.status,
    this.contactEmail,
    this.createdAt,
  });

  factory CompanyModel.fromJson(Map<String, dynamic> json) {
    return CompanyModel(
      id: json['id'] as String,
      name: json['name'] as String,
      slug: json['slug'] as String,
      status: json['status'] as String? ?? 'ACTIVE',
      contactEmail: json['contact_email'] as String?,
      createdAt: json['created_at'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'slug': slug,
      'status': status,
      if (contactEmail != null) 'contact_email': contactEmail,
      if (createdAt != null) 'created_at': createdAt,
    };
  }
}
