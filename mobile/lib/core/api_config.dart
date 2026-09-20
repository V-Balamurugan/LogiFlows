class ApiConfig {
  // Use 10.0.2.2 for Android emulator, localhost for iOS simulator or Web
  static const String baseUrl = 'http://localhost:8080/api/v1';
  static const String aiServiceUrl = 'http://localhost:8000/api/v1';

  static const String healthEndpoint = '$baseUrl/health';
  static const String readinessEndpoint = '$baseUrl/readiness';
  static const String predictDelayEndpoint = '$aiServiceUrl/predict/delay-risk';
}
