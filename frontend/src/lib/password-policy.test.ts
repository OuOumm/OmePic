import { describe, expect, it } from 'vitest';
import { isValidAdminPasswordStrength } from './password-policy';

describe('isValidAdminPasswordStrength', () => {
  it('requires more than 8 characters, uppercase, lowercase, and symbol characters', () => {
    expect(isValidAdminPasswordStrength('Abcdefg!')).toBe(false);
    expect(isValidAdminPasswordStrength('abcdefghi!')).toBe(false);
    expect(isValidAdminPasswordStrength('ABCDEFGHI!')).toBe(false);
    expect(isValidAdminPasswordStrength('Abcdefghi')).toBe(false);
    expect(isValidAdminPasswordStrength('Abcdefgh!')).toBe(true);
  });
});
