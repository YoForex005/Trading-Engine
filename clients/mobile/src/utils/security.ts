import { Platform } from 'react-native';
import * as Keychain from 'react-native-keychain';
import ReactNativeBiometrics from 'react-native-biometrics';

const rnBiometrics = new ReactNativeBiometrics();

// Secure token storage
export const secureStorage = {
  setItem: async (key: string, value: string): Promise<void> => {
    try {
      await Keychain.setGenericPassword(key, value, {
        service: key,
        accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED,
      });
    } catch (error) {
      console.error('Error storing secure item:', error);
      throw error;
    }
  },

  getItem: async (key: string): Promise<string | null> => {
    try {
      const credentials = await Keychain.getGenericPassword({ service: key });
      if (credentials) {
        return credentials.password;
      }
      return null;
    } catch (error) {
      console.error('Error retrieving secure item:', error);
      return null;
    }
  },

  removeItem: async (key: string): Promise<void> => {
    try {
      await Keychain.resetGenericPassword({ service: key });
    } catch (error) {
      console.error('Error removing secure item:', error);
    }
  },
};

// Biometric authentication
export const biometricAuth = {
  isAvailable: async (): Promise<boolean> => {
    try {
      const { available } = await rnBiometrics.isSensorAvailable();
      return available;
    } catch (error) {
      return false;
    }
  },

  getBiometryType: async (): Promise<string | null> => {
    try {
      const { biometryType } = await rnBiometrics.isSensorAvailable();
      return biometryType;
    } catch (error) {
      return null;
    }
  },

  authenticate: async (reason: string = 'Authenticate to continue'): Promise<boolean> => {
    try {
      const { success } = await rnBiometrics.simplePrompt({
        promptMessage: reason,
        cancelButtonText: 'Cancel',
      });
      return success;
    } catch (error) {
      console.error('Biometric authentication error:', error);
      return false;
    }
  },

  createKeys: async (): Promise<boolean> => {
    try {
      const { publicKey } = await rnBiometrics.createKeys();
      return !!publicKey;
    } catch (error) {
      console.error('Error creating biometric keys:', error);
      return false;
    }
  },

  deleteKeys: async (): Promise<boolean> => {
    try {
      const { keysDeleted } = await rnBiometrics.deleteKeys();
      return keysDeleted;
    } catch (error) {
      console.error('Error deleting biometric keys:', error);
      return false;
    }
  },
};

// Root/Jailbreak detection (simplified - use a library like jail-monkey for production)
export const deviceSecurity = {
  isDeviceSecure: (): boolean => {
    // In production, use a proper library like jail-monkey
    // This is a simplified check
    return true;
  },

  isDebugMode: (): boolean => {
    return __DEV__;
  },
};

// Session timeout
export class SessionManager {
  private static timeoutId: NodeJS.Timeout | null = null;
  private static timeoutDuration = 15 * 60 * 1000; // 15 minutes
  private static onTimeout: (() => void) | null = null;

  static start(callback: () => void): void {
    this.onTimeout = callback;
    this.reset();
  }

  static reset(): void {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
    }
    this.timeoutId = setTimeout(() => {
      if (this.onTimeout) {
        this.onTimeout();
      }
    }, this.timeoutDuration);
  }

  static stop(): void {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
    this.onTimeout = null;
  }
}

// SSL Pinning configuration (implement in native modules)
export const sslPinning = {
  // Public key hashes for your API servers
  publicKeyHashes: [
    // Add your server's public key hashes here
    // Example: 'sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA='
  ],
};
