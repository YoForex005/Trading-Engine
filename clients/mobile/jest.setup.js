import '@testing-library/jest-native/extend-expect';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock'),
);

// Mock Keychain
jest.mock('react-native-keychain', () => ({
  setGenericPassword: jest.fn().mockResolvedValue(true),
  getGenericPassword: jest.fn().mockResolvedValue({ username: 'test', password: 'test' }),
  resetGenericPassword: jest.fn().mockResolvedValue(true),
}));

// Mock Biometrics
jest.mock('react-native-biometrics', () => ({
  default: jest.fn().mockImplementation(() => ({
    isSensorAvailable: jest.fn().mockResolvedValue({ available: true, biometryType: 'FaceID' }),
    simplePrompt: jest.fn().mockResolvedValue({ success: true }),
    createKeys: jest.fn().mockResolvedValue({ publicKey: 'test-key' }),
    deleteKeys: jest.fn().mockResolvedValue({ keysDeleted: true }),
  })),
}));

// Mock Haptic Feedback
jest.mock('react-native-haptic-feedback', () => ({
  trigger: jest.fn(),
}));

// Mock Firebase
jest.mock('@react-native-firebase/app', () => ({}));
jest.mock('@react-native-firebase/messaging', () => () => ({
  requestPermission: jest.fn().mockResolvedValue(1),
  getToken: jest.fn().mockResolvedValue('test-token'),
  onTokenRefresh: jest.fn(),
  onMessage: jest.fn(),
  setBackgroundMessageHandler: jest.fn(),
  subscribeToTopic: jest.fn(),
  unsubscribeFromTopic: jest.fn(),
}));

// Mock Reanimated
jest.mock('react-native-reanimated', () => {
  const Reanimated = require('react-native-reanimated/mock');
  Reanimated.default.call = () => {};
  return Reanimated;
});

// Silence the warning: Animated: `useNativeDriver` is not supported
jest.mock('react-native/Libraries/Animated/NativeAnimatedHelper');

global.WebSocket = jest.fn().mockImplementation(() => ({
  readyState: 1,
  send: jest.fn(),
  close: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
}));
