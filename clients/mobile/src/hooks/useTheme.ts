import { useSelector } from 'react-redux';
import type { RootState } from '@/store';
import { lightTheme, darkTheme, Theme } from '@/theme';

export const useTheme = (): Theme => {
  const isDarkMode = useSelector((state: RootState) => state.ui.isDarkMode);
  return isDarkMode ? darkTheme : lightTheme;
};
