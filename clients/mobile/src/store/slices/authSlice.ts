import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { User } from '@/types';

interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  biometricEnabled: boolean;
}

const initialState: AuthState = {
  user: null,
  token: null,
  refreshToken: null,
  isAuthenticated: false,
  biometricEnabled: false,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{ user: User; token: string; refreshToken: string }>,
    ) => {
      state.user = action.payload.user;
      state.token = action.payload.token;
      state.refreshToken = action.payload.refreshToken;
      state.isAuthenticated = true;
      state.biometricEnabled = action.payload.user.biometricEnabled;
    },
    setUser: (state, action: PayloadAction<User>) => {
      state.user = action.payload;
      state.biometricEnabled = action.payload.biometricEnabled;
    },
    logout: (state) => {
      state.user = null;
      state.token = null;
      state.refreshToken = null;
      state.isAuthenticated = false;
      state.biometricEnabled = false;
    },
    setBiometricEnabled: (state, action: PayloadAction<boolean>) => {
      state.biometricEnabled = action.payload;
      if (state.user) {
        state.user.biometricEnabled = action.payload;
      }
    },
  },
});

export const { setCredentials, setUser, logout, setBiometricEnabled } = authSlice.actions;
export default authSlice.reducer;
