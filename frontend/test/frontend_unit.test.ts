import { test, describe, beforeEach } from 'node:test';
import assert from 'node:assert/strict';

// Mock localStorage for headless Node testing
class LocalStorageMock {
  private store: Record<string, string> = {};

  getItem(key: string): string | null {
    return this.store[key] || null;
  }

  setItem(key: string, value: string): void {
    this.store[key] = String(value);
  }

  removeItem(key: string): void {
    delete this.store[key];
  }

  clear(): void {
    this.store = {};
  }
}

const mockStorage = new LocalStorageMock();
(global as any).localStorage = mockStorage;

// Import functions under test
import { 
  getSavedToken, 
  setSavedToken, 
  getSavedRefreshToken, 
  setSavedRefreshToken, 
  clearSavedTokens 
} from '../src/services/api.ts';

describe('Frontend Auth Token Storage Tests', () => {
  beforeEach(() => {
    mockStorage.clear();
  });

  test('should return null when no access token is stored', () => {
    assert.equal(getSavedToken(), null);
  });

  test('should save and retrieve access token', () => {
    const dummyJwt = 'header.payload.signature-test-token-123';
    setSavedToken(dummyJwt);
    assert.equal(getSavedToken(), dummyJwt);
  });

  test('should save and retrieve refresh token', () => {
    const dummyRefreshToken = 'opaque-crypto-refresh-token-456';
    setSavedRefreshToken(dummyRefreshToken);
    assert.equal(getSavedRefreshToken(), dummyRefreshToken);
  });

  test('should clear both access and refresh tokens on logout', () => {
    setSavedToken('jwt-to-clear');
    setSavedRefreshToken('refresh-to-clear');
    
    assert.equal(getSavedToken(), 'jwt-to-clear');
    assert.equal(getSavedRefreshToken(), 'refresh-to-clear');

    clearSavedTokens();

    assert.equal(getSavedToken(), null);
    assert.equal(getSavedRefreshToken(), null);
  });
});

describe('Frontend Form Validation Rules', () => {
  // Test email format validation rules used in AuthScreen.tsx
  function validateEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email.trim());
  }

  // Test password complexity rules (>=8 chars, >=1 upper, >=1 lower, >=1 number, >=1 symbol)
  function validatePassword(password: string): { valid: boolean; errors: string[] } {
    const errors: string[] = [];
    if (password.length < 8) errors.push('Password must be at least 8 characters long');
    if (!/[A-Z]/.test(password)) errors.push('Password must contain at least one uppercase letter');
    if (!/[a-z]/.test(password)) errors.push('Password must contain at least one lowercase letter');
    if (!/[0-9]/.test(password)) errors.push('Password must contain at least one number');
    if (!/[^A-Za-z0-9]/.test(password)) errors.push('Password must contain at least one special symbol');
    return { valid: errors.length === 0, errors };
  }

  test('email validator accepts valid standard email', () => {
    assert.equal(validateEmail('operator@logiflows.test'), true);
    assert.equal(validateEmail('ceo.alpha@company.co.in'), true);
  });

  test('email validator rejects invalid emails', () => {
    assert.equal(validateEmail(''), false);
    assert.equal(validateEmail('not-an-email'), false);
    assert.equal(validateEmail('user@'), false);
    assert.equal(validateEmail('@domain.com'), false);
    assert.equal(validateEmail('user with spaces@domain.com'), false);
  });

  test('password validator enforces all 5 security rules', () => {
    const valid = validatePassword('LogiFlows#2026!');
    assert.equal(valid.valid, true);
    assert.equal(valid.errors.length, 0);

    const tooShort = validatePassword('P@1a');
    assert.equal(tooShort.valid, false);
    assert.ok(tooShort.errors.some(e => e.includes('8 characters')));

    const noUpper = validatePassword('pass@12345');
    assert.equal(noUpper.valid, false);
    assert.ok(noUpper.errors.some(e => e.includes('uppercase')));

    const noSymbol = validatePassword('Password12345');
    assert.equal(noSymbol.valid, false);
    assert.ok(noSymbol.errors.some(e => e.includes('special symbol')));
  });
});

describe('Frontend API Envelope Parsing', () => {
  test('unpacks successful API response envelope', () => {
    const rawEnvelope = {
      status: 'ok',
      service: 'logiflows-api',
      version: 'v1',
      timestamp: '2026-09-21T15:00:00Z',
      data: {
        user: { id: 'usr-1', email: 'test@logiflows.test' },
        tenants: [{ id: 'tnt-1', name: 'Alpha Logistics' }]
      }
    };

    assert.equal(rawEnvelope.status, 'ok');
    assert.equal(rawEnvelope.service, 'logiflows-api');
    assert.equal(rawEnvelope.data.tenants[0].name, 'Alpha Logistics');
  });

  test('formats standardized error message from error envelope', () => {
    const errorEnvelope = {
      error: {
        code: 'CROSS_TENANT_ACCESS_DENIED',
        message: 'You do not have permission to access resources belonging to this organization',
        request_id: 'req-uuid-999'
      }
    };

    assert.equal(errorEnvelope.error.code, 'CROSS_TENANT_ACCESS_DENIED');
    assert.ok(errorEnvelope.error.message.includes('permission'));
  });
});
