/**
 * Validation Utilities - Shared validation functions
 * Eliminates duplicated validation logic across components
 */

export type ValidationError = string | null;

export const validators = {
  /**
   * Validates that a value is not empty
   */
  required(value: string | number | null | undefined, fieldName: string): ValidationError {
    if (value === null || value === undefined || value === '') {
      return `${fieldName} is required`;
    }
    return null;
  },

  /**
   * Validates that a number is positive (> 0)
   */
  positive(value: number, fieldName: string): ValidationError {
    if (value <= 0) {
      return `${fieldName} must be positive`;
    }
    return null;
  },

  /**
   * Validates that a number is non-negative (>= 0)
   */
  nonNegative(value: number, fieldName: string): ValidationError {
    if (value < 0) {
      return `${fieldName} must be non-negative`;
    }
    return null;
  },

  /**
   * Validates that a number is within a range (inclusive)
   */
  range(value: number, min: number, max: number, fieldName: string): ValidationError {
    if (value < min || value > max) {
      return `${fieldName} must be between ${min} and ${max}`;
    }
    return null;
  },

  /**
   * Validates minimum value
   */
  minimum(value: number, min: number, fieldName: string): ValidationError {
    if (value < min) {
      return `${fieldName} must be at least ${min}`;
    }
    return null;
  },

  /**
   * Validates maximum value
   */
  maximum(value: number, max: number, fieldName: string): ValidationError {
    if (value > max) {
      return `${fieldName} must be at most ${max}`;
    }
    return null;
  },

  /**
   * Validates string minimum length
   */
  minLength(value: string, minLen: number, fieldName: string): ValidationError {
    if (value.length < minLen) {
      return `${fieldName} must be at least ${minLen} characters`;
    }
    return null;
  },

  /**
   * Validates string maximum length
   */
  maxLength(value: string, maxLen: number, fieldName: string): ValidationError {
    if (value.length > maxLen) {
      return `${fieldName} must be at most ${maxLen} characters`;
    }
    return null;
  },

  /**
   * Validates that value is one of allowed options
   */
  oneOf<T>(value: T, allowed: T[], fieldName: string): ValidationError {
    if (!allowed.includes(value)) {
      return `${fieldName} must be one of: ${allowed.join(', ')}`;
    }
    return null;
  },

  /**
   * Validates email format (basic)
   */
  email(value: string): ValidationError {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(value)) {
      return 'Invalid email format';
    }
    return null;
  },

  /**
   * Combines multiple validation results
   * Returns first error or null if all valid
   */
  combine(...results: ValidationError[]): ValidationError {
    for (const result of results) {
      if (result !== null) {
        return result;
      }
    }
    return null;
  },
};

/**
 * Helper to validate an object against a schema
 */
export type ValidationSchema<T> = {
  [K in keyof T]?: (value: T[K]) => ValidationError;
};

export function validateObject<T extends Record<string, unknown>>(
  obj: T,
  schema: ValidationSchema<T>
): Record<keyof T, ValidationError> {
  const errors = {} as Record<keyof T, ValidationError>;

  for (const key in schema) {
    const validator = schema[key];
    if (validator) {
      errors[key] = validator(obj[key]);
    }
  }

  return errors;
}

/**
 * Checks if validation errors object has any errors
 */
export function hasErrors<T>(errors: Record<keyof T, ValidationError>): boolean {
  return Object.values(errors).some((error) => error !== null);
}
