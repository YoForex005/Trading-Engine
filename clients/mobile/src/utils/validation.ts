export const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

export const validatePassword = (password: string): {
  valid: boolean;
  errors: string[];
} => {
  const errors: string[] = [];

  if (password.length < 8) {
    errors.push('Password must be at least 8 characters');
  }
  if (!/[A-Z]/.test(password)) {
    errors.push('Password must contain at least one uppercase letter');
  }
  if (!/[a-z]/.test(password)) {
    errors.push('Password must contain at least one lowercase letter');
  }
  if (!/[0-9]/.test(password)) {
    errors.push('Password must contain at least one number');
  }
  if (!/[!@#$%^&*]/.test(password)) {
    errors.push('Password must contain at least one special character');
  }

  return {
    valid: errors.length === 0,
    errors,
  };
};

export const validatePhoneNumber = (phone: string): boolean => {
  const phoneRegex = /^\+?[1-9]\d{1,14}$/;
  return phoneRegex.test(phone.replace(/[\s-]/g, ''));
};

export const validateOrderVolume = (
  volume: number,
  minVolume: number = 0.01,
  maxVolume: number = 1000,
): { valid: boolean; error?: string } => {
  if (volume < minVolume) {
    return { valid: false, error: `Minimum volume is ${minVolume}` };
  }
  if (volume > maxVolume) {
    return { valid: false, error: `Maximum volume is ${maxVolume}` };
  }
  return { valid: true };
};

export const validatePrice = (price: number): { valid: boolean; error?: string } => {
  if (price <= 0) {
    return { valid: false, error: 'Price must be greater than 0' };
  }
  return { valid: true };
};

export const sanitizeInput = (input: string): string => {
  return input.trim().replace(/[<>]/g, '');
};
