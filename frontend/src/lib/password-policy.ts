const CHANGE_PASSWORD_MIN_LENGTH_EXCLUSIVE = 8;
const uppercasePattern = /\p{Lu}/u;
const lowercasePattern = /\p{Ll}/u;
const symbolPattern = /[\p{P}\p{S}]/u;

export function isValidAdminPasswordStrength(password: string): boolean {
  return (
    Array.from(password).length > CHANGE_PASSWORD_MIN_LENGTH_EXCLUSIVE &&
    uppercasePattern.test(password) &&
    lowercasePattern.test(password) &&
    symbolPattern.test(password)
  );
}
