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
  final String availabilityStatus;
  final String verificationStatus;
  final String? employmentType;
  final String? joiningDate;
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
    this.availabilityStatus = 'AVAILABLE',
    this.verificationStatus = 'PENDING',
    this.employmentType,
    this.joiningDate,
    required this.isActive,
  });

  String get fullName => '$firstName $lastName'.trim();
  String get role => operationalRole;
  bool get isDriver => operationalRole == 'DRIVER';
  bool get isAvailable => availabilityStatus == 'AVAILABLE' && status == 'ACTIVE';

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
      employeeCode: json['employee_code'] as String? ?? '',
      firstName: parsedFirst,
      lastName: parsedLast,
      email: json['email'] as String?,
      phone: json['phone'] as String?,
      designation: json['designation'] as String? ?? 'Staff',
      operationalRole: json['operational_role'] as String? ?? json['role'] as String? ?? 'OPERATOR',
      licenseNumber: json['license_number'] as String?,
      status: json['status'] as String? ?? 'ACTIVE',
      availabilityStatus: json['availability_status'] as String? ?? 'AVAILABLE',
      verificationStatus: json['verification_status'] as String? ?? 'PENDING',
      employmentType: json['employment_type'] as String?,
      joiningDate: json['joining_date'] as String?,
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
      'availability_status': availabilityStatus,
      'verification_status': verificationStatus,
      if (employmentType != null) 'employment_type': employmentType,
      if (joiningDate != null) 'joining_date': joiningDate,
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
  final String availabilityStatus;
  final bool isActive;
  final bool isElectric;
  final String? currentDriverName;
  final String? currentDriverId;
  final String? currentDriverCode;

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
    this.availabilityStatus = 'AVAILABLE',
    required this.isActive,
    bool? isElectric,
    this.currentDriverName,
    this.currentDriverId,
    this.currentDriverCode,
  }) : isElectric = isElectric ?? (vehicleType == 'ELECTRIC_VAN');

  bool get isAvailable => availabilityStatus == 'AVAILABLE' && status == 'AVAILABLE';
  bool get isAssigned => status == 'ASSIGNED' || currentDriverId != null;

  factory VehicleModel.fromJson(Map<String, dynamic> json) {
    final vType = json['vehicle_type'] as String? ?? 'VAN';
    final electricVal = json['is_electric'] as bool? ?? (vType == 'ELECTRIC_VAN');

    return VehicleModel(
      id: json['id'] as String,
      tenantId: json['tenant_id'] as String,
      assignedBranchId: json['assigned_branch_id'] as String? ?? json['branch_id'] as String?,
      branchName: json['branch_name'] as String?,
      registrationNumber: json['registration_number'] as String,
      vehicleType: vType,
      makeModel: json['make_model'] as String?,
      year: json['year'] as int?,
      maxWeightKg: (json['max_weight_kg'] as num?)?.toDouble() ?? 500.0,
      maxVolumeCbm: (json['max_volume_cbm'] as num?)?.toDouble() ?? 3.0,
      status: json['status'] as String? ?? 'AVAILABLE',
      availabilityStatus: json['availability_status'] as String? ?? 'AVAILABLE',
      isActive: json['is_active'] as bool? ?? true,
      isElectric: electricVal,
      currentDriverName: json['current_driver_name'] as String?,
      currentDriverId: json['current_driver_id'] as String?,
      currentDriverCode: json['current_driver_code'] as String?,
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
      'availability_status': availabilityStatus,
      'is_active': isActive,
      if (currentDriverName != null) 'current_driver_name': currentDriverName,
      if (currentDriverId != null) 'current_driver_id': currentDriverId,
      if (currentDriverCode != null) 'current_driver_code': currentDriverCode,
    };
  }
}

class AssignmentModel {
  final String id;
  final String tenantId;
  final String vehicleId;
  final String employeeId;
  final String? registrationNumber;
  final String? vehicleType;
  final String? driverName;
  final String? driverCode;
  final String? assignedAt;
  final String? unassignedAt;
  final String status;
  final String? notes;

  AssignmentModel({
    required this.id,
    required this.tenantId,
    required this.vehicleId,
    required this.employeeId,
    this.registrationNumber,
    this.vehicleType,
    this.driverName,
    this.driverCode,
    this.assignedAt,
    this.unassignedAt,
    required this.status,
    this.notes,
  });

  bool get isActive => status == 'ACTIVE';

  factory AssignmentModel.fromJson(Map<String, dynamic> json) {
    return AssignmentModel(
      id: json['id'] as String,
      tenantId: json['tenant_id'] as String,
      vehicleId: json['vehicle_id'] as String,
      employeeId: json['employee_id'] as String,
      registrationNumber: json['registration_number'] as String?,
      vehicleType: json['vehicle_type'] as String?,
      driverName: json['driver_name'] as String?,
      driverCode: json['driver_code'] as String?,
      assignedAt: json['assigned_at'] as String?,
      unassignedAt: json['unassigned_at'] as String?,
      status: json['status'] as String? ?? 'ACTIVE',
      notes: json['notes'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'tenant_id': tenantId,
      'vehicle_id': vehicleId,
      'employee_id': employeeId,
      if (registrationNumber != null) 'registration_number': registrationNumber,
      if (vehicleType != null) 'vehicle_type': vehicleType,
      if (driverName != null) 'driver_name': driverName,
      if (driverCode != null) 'driver_code': driverCode,
      if (assignedAt != null) 'assigned_at': assignedAt,
      if (unassignedAt != null) 'unassigned_at': unassignedAt,
      'status': status,
      if (notes != null) 'notes': notes,
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

class AssignedVehicleModel {
  final String id;
  final String registrationNumber;
  final String vehicleType;
  final String? makeModel;
  final String status;

  AssignedVehicleModel({
    required this.id,
    required this.registrationNumber,
    required this.vehicleType,
    this.makeModel,
    required this.status,
  });

  factory AssignedVehicleModel.fromJson(Map<String, dynamic> json) {
    return AssignedVehicleModel(
      id: json['id'] as String,
      registrationNumber: json['registration_number'] as String,
      vehicleType: json['vehicle_type'] as String,
      makeModel: json['make_model'] as String?,
      status: json['status'] as String? ?? 'ACTIVE',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'registration_number': registrationNumber,
      'vehicle_type': vehicleType,
      if (makeModel != null) 'make_model': makeModel,
      'status': status,
    };
  }
}

class EmployeeMeModel {
  final EmployeeModel employee;
  final String systemRole;
  final AssignedVehicleModel? assignedVehicle;

  EmployeeMeModel({
    required this.employee,
    required this.systemRole,
    this.assignedVehicle,
  });

  factory EmployeeMeModel.fromJson(Map<String, dynamic> json) {
    return EmployeeMeModel(
      employee: EmployeeModel.fromJson(json['employee'] as Map<String, dynamic>),
      systemRole: json['system_role'] as String? ?? 'EMPLOYEE',
      assignedVehicle: json['assigned_vehicle'] != null
          ? AssignedVehicleModel.fromJson(json['assigned_vehicle'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'employee': employee.toJson(),
      'system_role': systemRole,
      if (assignedVehicle != null) 'assigned_vehicle': assignedVehicle!.toJson(),
    };
  }
}

class EmployeeAccountStatusModel {
  final String employeeId;
  final String employeeCode;
  final String fullName;
  final String operationalRole;
  final String status;
  final bool hasAccount;
  final String? userId;
  final String? userEmail;
  final bool? userIsActive;
  final String? systemRole;
  final String? branchId;
  final String? branchName;

  EmployeeAccountStatusModel({
    required this.employeeId,
    required this.employeeCode,
    required this.fullName,
    required this.operationalRole,
    required this.status,
    required this.hasAccount,
    this.userId,
    this.userEmail,
    this.userIsActive,
    this.systemRole,
    this.branchId,
    this.branchName,
  });

  factory EmployeeAccountStatusModel.fromJson(Map<String, dynamic> json) {
    return EmployeeAccountStatusModel(
      employeeId: json['employee_id'] as String,
      employeeCode: json['employee_code'] as String,
      fullName: json['full_name'] as String,
      operationalRole: json['operational_role'] as String,
      status: json['status'] as String,
      hasAccount: json['has_account'] as bool? ?? false,
      userId: json['user_id'] as String?,
      userEmail: json['user_email'] as String?,
      userIsActive: json['user_is_active'] as bool?,
      systemRole: json['system_role'] as String?,
      branchId: json['branch_id'] as String?,
      branchName: json['branch_name'] as String?,
    );
  }
}

class BranchEmployeeSummaryModel {
  final String id;
  final String employeeCode;
  final String firstName;
  final String lastName;
  final String? email;
  final String? phone;
  final String designation;
  final String operationalRole;
  final String status;
  final String availabilityStatus;
  final bool isActive;

  BranchEmployeeSummaryModel({
    required this.id,
    required this.employeeCode,
    required this.firstName,
    required this.lastName,
    this.email,
    this.phone,
    required this.designation,
    required this.operationalRole,
    required this.status,
    required this.availabilityStatus,
    required this.isActive,
  });

  factory BranchEmployeeSummaryModel.fromJson(Map<String, dynamic> json) {
    return BranchEmployeeSummaryModel(
      id: json['id'] as String,
      employeeCode: json['employee_code'] as String? ?? '',
      firstName: json['first_name'] as String? ?? '',
      lastName: json['last_name'] as String? ?? '',
      email: json['email'] as String?,
      phone: json['phone'] as String?,
      designation: json['designation'] as String? ?? '',
      operationalRole: json['operational_role'] as String? ?? '',
      status: json['status'] as String? ?? 'ACTIVE',
      availabilityStatus: json['availability_status'] as String? ?? 'AVAILABLE',
      isActive: json['is_active'] as bool? ?? true,
    );
  }
}

class BranchVehicleSummaryModel {
  final String id;
  final String registrationNumber;
  final String vehicleType;
  final String? makeModel;
  final int? year;
  final double maxWeightKg;
  final double maxVolumeCbm;
  final String status;
  final String availabilityStatus;
  final bool isActive;
  final String? currentDriverName;

  BranchVehicleSummaryModel({
    required this.id,
    required this.registrationNumber,
    required this.vehicleType,
    this.makeModel,
    this.year,
    required this.maxWeightKg,
    required this.maxVolumeCbm,
    required this.status,
    required this.availabilityStatus,
    required this.isActive,
    this.currentDriverName,
  });

  factory BranchVehicleSummaryModel.fromJson(Map<String, dynamic> json) {
    return BranchVehicleSummaryModel(
      id: json['id'] as String,
      registrationNumber: json['registration_number'] as String,
      vehicleType: json['vehicle_type'] as String,
      makeModel: json['make_model'] as String?,
      year: json['year'] as int?,
      maxWeightKg: (json['max_weight_kg'] as num?)?.toDouble() ?? 0.0,
      maxVolumeCbm: (json['max_volume_cbm'] as num?)?.toDouble() ?? 0.0,
      status: json['status'] as String? ?? 'ACTIVE',
      availabilityStatus: json['availability_status'] as String? ?? 'AVAILABLE',
      isActive: json['is_active'] as bool? ?? true,
      currentDriverName: json['current_driver_name'] as String?,
    );
  }
}
