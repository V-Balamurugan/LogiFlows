class AuthUser {
  final String id;
  final String email;
  final String fullName;
  final String? phoneNumber;
  final bool isActive;
  final bool isPlatformAdmin;
  final bool emailVerified;

  AuthUser({
    required this.id,
    required this.email,
    required this.fullName,
    this.phoneNumber,
    required this.isActive,
    required this.isPlatformAdmin,
    required this.emailVerified,
  });

  factory AuthUser.fromJson(Map<String, dynamic> json) {
    return AuthUser(
      id: json['id'] as String,
      email: json['email'] as String,
      fullName: json['full_name'] as String,
      phoneNumber: json['phone_number'] as String?,
      isActive: json['is_active'] as bool? ?? true,
      isPlatformAdmin: json['is_platform_admin'] as bool? ?? false,
      emailVerified: json['email_verified'] as bool? ?? false,
    );
  }
}

class Tenant {
  final String id;
  final String name;
  final String slug;
  final String role;

  Tenant({
    required this.id,
    required this.name,
    required this.slug,
    required this.role,
  });

  factory Tenant.fromJson(Map<String, dynamic> json) {
    return Tenant(
      id: json['id'] as String,
      name: json['name'] as String,
      slug: json['slug'] as String,
      role: json['role'] as String,
    );
  }
}

class AuthResponse {
  final String token;
  final String expiresAt;
  final String? refreshToken;
  final String? refreshTokenExpiresAt;
  final AuthUser user;
  final List<Tenant> tenants;

  AuthResponse({
    required this.token,
    required this.expiresAt,
    this.refreshToken,
    this.refreshTokenExpiresAt,
    required this.user,
    required this.tenants,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      token: json['token'] as String,
      expiresAt: json['expires_at'] as String,
      refreshToken: json['refresh_token'] as String?,
      refreshTokenExpiresAt: json['refresh_token_expires_at'] as String?,
      user: AuthUser.fromJson(json['user'] as Map<String, dynamic>),
      tenants: (json['tenants'] as List<dynamic>? ?? [])
          .map((t) => Tenant.fromJson(t as Map<String, dynamic>))
          .toList(),
    );
  }
}
