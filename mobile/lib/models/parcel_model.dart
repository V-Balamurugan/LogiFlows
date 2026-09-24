class ParcelModel {
  final String id;
  final String trackingNumber;
  final String tenantId;
  final String originBranchId;
  final String destinationBranchId;
  final String? currentBranchId;
  final String? currentCustodianId;
  final String senderName;
  final String senderPhone;
  final String? senderEmail;
  final String senderAddress;
  final String senderCity;
  final String recipientName;
  final String recipientPhone;
  final String? recipientEmail;
  final String recipientAddress;
  final String recipientCity;
  final double weightKg;
  final double? lengthCm;
  final double? widthCm;
  final double? heightCm;
  final String status;
  final String serviceType;
  final double? declaredValue;
  final String? notes;
  final DateTime createdAt;
  final DateTime updatedAt;

  ParcelModel({
    required this.id,
    required this.trackingNumber,
    required this.tenantId,
    required this.originBranchId,
    required this.destinationBranchId,
    this.currentBranchId,
    this.currentCustodianId,
    required this.senderName,
    required this.senderPhone,
    this.senderEmail,
    required this.senderAddress,
    required this.senderCity,
    required this.recipientName,
    required this.recipientPhone,
    this.recipientEmail,
    required this.recipientAddress,
    required this.recipientCity,
    required this.weightKg,
    this.lengthCm,
    this.widthCm,
    this.heightCm,
    required this.status,
    required this.serviceType,
    this.declaredValue,
    this.notes,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ParcelModel.fromJson(Map<String, dynamic> json) {
    return ParcelModel(
      id: json['id'] as String,
      trackingNumber: json['tracking_number'] as String,
      tenantId: json['tenant_id'] as String,
      originBranchId: json['origin_branch_id'] as String,
      destinationBranchId: json['destination_branch_id'] as String,
      currentBranchId: json['current_branch_id'] as String?,
      currentCustodianId: json['current_custodian_id'] as String?,
      senderName: json['sender_name'] as String? ?? '',
      senderPhone: json['sender_phone'] as String? ?? '',
      senderEmail: json['sender_email'] as String?,
      senderAddress: json['sender_address'] as String? ?? '',
      senderCity: json['sender_city'] as String? ?? '',
      recipientName: json['recipient_name'] as String? ?? '',
      recipientPhone: json['recipient_phone'] as String? ?? '',
      recipientEmail: json['recipient_email'] as String?,
      recipientAddress: json['recipient_address'] as String? ?? '',
      recipientCity: json['recipient_city'] as String? ?? '',
      weightKg: (json['weight_kg'] as num?)?.toDouble() ?? 0.0,
      lengthCm: (json['length_cm'] as num?)?.toDouble(),
      widthCm: (json['width_cm'] as num?)?.toDouble(),
      heightCm: (json['height_cm'] as num?)?.toDouble(),
      status: json['status'] as String? ?? 'CREATED',
      serviceType: json['service_type'] as String? ?? 'STANDARD',
      declaredValue: (json['declared_value'] as num?)?.toDouble(),
      notes: json['notes'] as String?,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'tracking_number': trackingNumber,
      'tenant_id': tenantId,
      'origin_branch_id': originBranchId,
      'destination_branch_id': destinationBranchId,
      'current_branch_id': currentBranchId,
      'current_custodian_id': currentCustodianId,
      'sender_name': senderName,
      'sender_phone': senderPhone,
      'sender_email': senderEmail,
      'sender_address': senderAddress,
      'sender_city': senderCity,
      'recipient_name': recipientName,
      'recipient_phone': recipientPhone,
      'recipient_email': recipientEmail,
      'recipient_address': recipientAddress,
      'recipient_city': recipientCity,
      'weight_kg': weightKg,
      'length_cm': lengthCm,
      'width_cm': widthCm,
      'height_cm': heightCm,
      'status': status,
      'service_type': serviceType,
      'declared_value': declaredValue,
      'notes': notes,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }
}

class DeliveryTaskModel {
  final String id;
  final String tenantId;
  final String parcelId;
  final String? parcelTrackingNumber;
  final String driverId;
  final String? driverName;
  final String? vehicleId;
  final String branchId;
  final String status;
  final String priority;
  final String? recipientName;
  final String? recipientPhone;
  final String? recipientAddress;
  final String? notes;
  final DateTime? scheduledDate;
  final DateTime? dispatchedAt;
  final DateTime? completedAt;
  final DateTime createdAt;

  DeliveryTaskModel({
    required this.id,
    required this.tenantId,
    required this.parcelId,
    this.parcelTrackingNumber,
    required this.driverId,
    this.driverName,
    this.vehicleId,
    required this.branchId,
    required this.status,
    required this.priority,
    this.recipientName,
    this.recipientPhone,
    this.recipientAddress,
    this.notes,
    this.scheduledDate,
    this.dispatchedAt,
    this.completedAt,
    required this.createdAt,
  });

  factory DeliveryTaskModel.fromJson(Map<String, dynamic> json) {
    return DeliveryTaskModel(
      id: json['id'] as String,
      tenantId: json['tenant_id'] as String,
      parcelId: json['parcel_id'] as String,
      parcelTrackingNumber: json['parcel_tracking_number'] as String?,
      driverId: json['driver_id'] as String,
      driverName: json['driver_name'] as String?,
      vehicleId: json['vehicle_id'] as String?,
      branchId: json['branch_id'] as String,
      status: json['status'] as String? ?? 'ASSIGNED',
      priority: json['priority'] as String? ?? 'NORMAL',
      recipientName: json['recipient_name'] as String?,
      recipientPhone: json['recipient_phone'] as String?,
      recipientAddress: json['recipient_address'] as String?,
      notes: json['notes'] as String?,
      scheduledDate: json['scheduled_date'] != null
          ? DateTime.parse(json['scheduled_date'] as String)
          : null,
      dispatchedAt: json['dispatched_at'] != null
          ? DateTime.parse(json['dispatched_at'] as String)
          : null,
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'tenant_id': tenantId,
      'parcel_id': parcelId,
      'parcel_tracking_number': parcelTrackingNumber,
      'driver_id': driverId,
      'driver_name': driverName,
      'vehicle_id': vehicleId,
      'branch_id': branchId,
      'status': status,
      'priority': priority,
      'recipient_name': recipientName,
      'recipient_phone': recipientPhone,
      'recipient_address': recipientAddress,
      'notes': notes,
      'scheduled_date': scheduledDate?.toIso8601String(),
      'dispatched_at': dispatchedAt?.toIso8601String(),
      'completed_at': completedAt?.toIso8601String(),
      'created_at': createdAt.toIso8601String(),
    };
  }
}

class DeliveryAttemptModel {
  final String id;
  final String taskId;
  final int attemptNumber;
  final String outcome;
  final String? reason;
  final String? notes;
  final DateTime attemptedAt;

  DeliveryAttemptModel({
    required this.id,
    required this.taskId,
    required this.attemptNumber,
    required this.outcome,
    this.reason,
    this.notes,
    required this.attemptedAt,
  });

  factory DeliveryAttemptModel.fromJson(Map<String, dynamic> json) {
    return DeliveryAttemptModel(
      id: json['id'] as String,
      taskId: json['task_id'] as String,
      attemptNumber: json['attempt_number'] as int? ?? 1,
      outcome: json['outcome'] as String? ?? 'FAILED',
      reason: json['reason'] as String?,
      notes: json['notes'] as String?,
      attemptedAt: json['attempted_at'] != null
          ? DateTime.parse(json['attempted_at'] as String)
          : DateTime.now(),
    );
  }
}

class BranchTransferModel {
  final String id;
  final String manifestNumber;
  final String tenantId;
  final String sourceBranchId;
  final String destinationBranchId;
  final String? vehicleId;
  final String? driverId;
  final String status;
  final int totalParcels;
  final DateTime? dispatchedAt;
  final DateTime? receivedAt;
  final DateTime createdAt;

  BranchTransferModel({
    required this.id,
    required this.manifestNumber,
    required this.tenantId,
    required this.sourceBranchId,
    required this.destinationBranchId,
    this.vehicleId,
    this.driverId,
    required this.status,
    required this.totalParcels,
    this.dispatchedAt,
    this.receivedAt,
    required this.createdAt,
  });

  factory BranchTransferModel.fromJson(Map<String, dynamic> json) {
    return BranchTransferModel(
      id: json['id'] as String,
      manifestNumber: json['manifest_number'] as String,
      tenantId: json['tenant_id'] as String,
      sourceBranchId: json['source_branch_id'] as String,
      destinationBranchId: json['destination_branch_id'] as String,
      vehicleId: json['vehicle_id'] as String?,
      driverId: json['driver_id'] as String?,
      status: json['status'] as String? ?? 'PENDING',
      totalParcels: json['total_parcels'] as int? ?? 0,
      dispatchedAt: json['dispatched_at'] != null
          ? DateTime.parse(json['dispatched_at'] as String)
          : null,
      receivedAt: json['received_at'] != null
          ? DateTime.parse(json['received_at'] as String)
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }
}

class PublicTrackingModel {
  final String trackingNumber;
  final String status;
  final String serviceType;
  final String? originCity;
  final String? destinationCity;
  final String? currentLocation;
  final DateTime? estimatedDelivery;
  final List<TrackingEventModel> events;

  PublicTrackingModel({
    required this.trackingNumber,
    required this.status,
    required this.serviceType,
    this.originCity,
    this.destinationCity,
    this.currentLocation,
    this.estimatedDelivery,
    this.events = const [],
  });

  factory PublicTrackingModel.fromJson(Map<String, dynamic> json) {
    final eventsRaw = json['events'] as List<dynamic>? ?? [];
    return PublicTrackingModel(
      trackingNumber: json['tracking_number'] as String,
      status: json['status'] as String? ?? 'CREATED',
      serviceType: json['service_type'] as String? ?? 'STANDARD',
      originCity: json['origin_city'] as String?,
      destinationCity: json['destination_city'] as String?,
      currentLocation: json['current_location'] as String?,
      estimatedDelivery: json['estimated_delivery'] != null
          ? DateTime.parse(json['estimated_delivery'] as String)
          : null,
      events: eventsRaw
          .map((e) => TrackingEventModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}

class TrackingEventModel {
  final String status;
  final String? location;
  final String? description;
  final DateTime timestamp;

  TrackingEventModel({
    required this.status,
    this.location,
    this.description,
    required this.timestamp,
  });

  factory TrackingEventModel.fromJson(Map<String, dynamic> json) {
    return TrackingEventModel(
      status: json['status'] as String,
      location: json['location'] as String?,
      description: json['description'] as String?,
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'] as String)
          : DateTime.now(),
    );
  }
}
