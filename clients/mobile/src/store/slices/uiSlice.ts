import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Appearance } from 'react-native';

type ThemeMode = 'light' | 'dark' | 'auto';

interface UIState {
  themeMode: ThemeMode;
  isDarkMode: boolean;
  language: string;
  hapticEnabled: boolean;
}

const systemColorScheme = Appearance.getColorScheme();

const initialState: UIState = {
  themeMode: 'auto',
  isDarkMode: systemColorScheme === 'dark',
  language: 'en',
  hapticEnabled: true,
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    setThemeMode: (state, action: PayloadAction<ThemeMode>) => {
      state.themeMode = action.payload;
      if (action.payload === 'auto') {
        state.isDarkMode = Appearance.getColorScheme() === 'dark';
      } else {
        state.isDarkMode = action.payload === 'dark';
      }
    },
    setIsDarkMode: (state, action: PayloadAction<boolean>) => {
      state.isDarkMode = action.payload;
    },
    setLanguage: (state, action: PayloadAction<string>) => {
      state.language = action.payload;
    },
    setHapticEnabled: (state, action: PayloadAction<boolean>) => {
      state.hapticEnabled = action.payload;
    },
  },
});

export const { setThemeMode, setIsDarkMode, setLanguage, setHapticEnabled } = uiSlice.actions;
export default uiSlice.reducer;
