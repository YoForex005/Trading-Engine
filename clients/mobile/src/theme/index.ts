import { Dimensions } from 'react-native';

const { width, height } = Dimensions.get('window');

export const colors = {
  // Light theme
  light: {
    primary: '#007AFF',
    secondary: '#5856D6',
    success: '#34C759',
    warning: '#FF9500',
    error: '#FF3B30',
    background: '#FFFFFF',
    surface: '#F2F2F7',
    surfaceVariant: '#E5E5EA',
    text: '#000000',
    textSecondary: '#8E8E93',
    border: '#C6C6C8',
    buy: '#34C759',
    sell: '#FF3B30',
    chart: {
      up: '#34C759',
      down: '#FF3B30',
      grid: '#E5E5EA',
      text: '#8E8E93',
    },
  },
  // Dark theme
  dark: {
    primary: '#0A84FF',
    secondary: '#5E5CE6',
    success: '#32D74B',
    warning: '#FF9F0A',
    error: '#FF453A',
    background: '#000000',
    surface: '#1C1C1E',
    surfaceVariant: '#2C2C2E',
    text: '#FFFFFF',
    textSecondary: '#8E8E93',
    border: '#38383A',
    buy: '#32D74B',
    sell: '#FF453A',
    chart: {
      up: '#32D74B',
      down: '#FF453A',
      grid: '#2C2C2E',
      text: '#8E8E93',
    },
  },
};

export const spacing = {
  xs: 4,
  sm: 8,
  md: 16,
  lg: 24,
  xl: 32,
  xxl: 48,
};

export const typography = {
  h1: {
    fontSize: 32,
    fontWeight: '700' as const,
    lineHeight: 40,
  },
  h2: {
    fontSize: 28,
    fontWeight: '700' as const,
    lineHeight: 36,
  },
  h3: {
    fontSize: 24,
    fontWeight: '600' as const,
    lineHeight: 32,
  },
  h4: {
    fontSize: 20,
    fontWeight: '600' as const,
    lineHeight: 28,
  },
  body1: {
    fontSize: 17,
    fontWeight: '400' as const,
    lineHeight: 24,
  },
  body2: {
    fontSize: 15,
    fontWeight: '400' as const,
    lineHeight: 22,
  },
  caption: {
    fontSize: 13,
    fontWeight: '400' as const,
    lineHeight: 18,
  },
  button: {
    fontSize: 17,
    fontWeight: '600' as const,
    lineHeight: 24,
  },
};

export const borderRadius = {
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  full: 9999,
};

export const shadows = {
  sm: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  md: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 4,
  },
  lg: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.15,
    shadowRadius: 8,
    elevation: 8,
  },
};

export const dimensions = {
  window: {
    width,
    height,
  },
  isSmallDevice: width < 375,
  isTablet: width >= 768,
};

export type Theme = {
  colors: typeof colors.light;
  spacing: typeof spacing;
  typography: typeof typography;
  borderRadius: typeof borderRadius;
  shadows: typeof shadows;
  dimensions: typeof dimensions;
};

export const lightTheme: Theme = {
  colors: colors.light,
  spacing,
  typography,
  borderRadius,
  shadows,
  dimensions,
};

export const darkTheme: Theme = {
  colors: colors.dark,
  spacing,
  typography,
  borderRadius,
  shadows,
  dimensions,
};
