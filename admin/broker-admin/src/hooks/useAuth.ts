'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { api, setAuthToken, clearAuthToken, isAuthenticated as checkIsAuthenticated } from '@/services/apiClient';
import { API_CONFIG } from '@/config/api';

/**
 * Authentication state and user info
 */
interface AuthState {
    isAuthenticated: boolean;
    isLoading: boolean;
    user: AdminUser | null;
    token: string | null;
    error: string | null;
}

/**
 * Admin user info returned from backend
 */
interface AdminUser {
    id?: number;
    username: string;
    email?: string;
    role: string;
}

/**
 * Login response from backend
 */
interface LoginResponse {
    success: boolean;
    token: string;
    session: {
        SessionID: string;
        AdminID: number;
        Username: string;
        Role: string;
        IPAddress: string;
        UserAgent: string;
        CreatedAt: string;
        ExpiresAt: string;
        LastActive: string;
    };
}

/**
 * useAuth Hook
 * Manages authentication state, login, logout, and token validation
 *
 * Features:
 * - Login with username/password
 * - Logout and clear tokens
 * - Check authentication status on mount
 * - Store JWT token in localStorage (SSR-safe)
 * - Auto-redirect on successful login
 */
export function useAuth() {
    const router = useRouter();
    const [state, setState] = useState<AuthState>({
        isAuthenticated: false,
        isLoading: true,
        user: null,
        token: null,
        error: null,
    });

    /**
     * Check if user is authenticated on mount
     * Validates token from localStorage
     */
    const checkAuth = useCallback(async () => {
        // SSR-safe: only run on client
        if (typeof window === 'undefined') {
            setState(prev => ({ ...prev, isLoading: false }));
            return;
        }

        try {
            // Check if token exists
            const authenticated = checkIsAuthenticated();

            if (!authenticated) {
                setState({
                    isAuthenticated: false,
                    isLoading: false,
                    user: null,
                    token: null,
                    error: null,
                });
                return;
            }

            // Get token from localStorage
            const token = localStorage.getItem('admin_token') ||
                         localStorage.getItem('rtx_token') ||
                         localStorage.getItem('jwt_token') ||
                         null;

            // TODO: Optionally verify token with backend
            // For now, just trust localStorage token
            // In production, you'd want to call API_CONFIG.AUTH_VERIFY

            if (token) {
                // Extract user info from token if needed (JWT decode)
                // For now, just mark as authenticated
                setState({
                    isAuthenticated: true,
                    isLoading: false,
                    user: null, // Will be populated after token verification
                    token,
                    error: null,
                });
            } else {
                setState({
                    isAuthenticated: false,
                    isLoading: false,
                    user: null,
                    token: null,
                    error: null,
                });
            }
        } catch (err) {
            console.error('[useAuth] Auth check failed:', err);
            setState({
                isAuthenticated: false,
                isLoading: false,
                user: null,
                token: null,
                error: 'Authentication check failed',
            });
        }
    }, []);

    /**
     * Login with username and password
     * Calls backend POST /login endpoint
     */
    const login = useCallback(async (username: string, password: string) => {
        setState(prev => ({ ...prev, isLoading: true, error: null }));

        try {
            // Call backend login endpoint
            // Based on backend/admin/handlers.go HandleLogin
            const response = await api.post<LoginResponse>(
                API_CONFIG.AUTH_LOGIN,
                { username, password },
                { skipAuth: true } // Don't include auth header for login
            );

            if (!response.success || !response.token) {
                throw new Error('Login failed - invalid response from server');
            }

            // Extract user info from session
            const user: AdminUser = {
                id: response.session.AdminID,
                username: response.session.Username,
                role: response.session.Role,
            };

            // Store token in localStorage
            setAuthToken(response.token);

            // Update state
            setState({
                isAuthenticated: true,
                isLoading: false,
                user,
                token: response.token,
                error: null,
            });

            // Redirect to admin dashboard
            router.push('/');
        } catch (err: any) {
            console.error('[useAuth] Login failed:', err);

            let errorMessage = 'Login failed';
            if (err.status === 401) {
                errorMessage = 'Invalid username or password';
            } else if (err.message) {
                errorMessage = err.message;
            }

            setState(prev => ({
                ...prev,
                isLoading: false,
                error: errorMessage,
            }));

            throw err;
        }
    }, [router]);

    /**
     * Logout and clear authentication
     * Removes token from localStorage and redirects to login
     */
    const logout = useCallback(async () => {
        try {
            // Optionally call backend logout endpoint
            // await api.post(API_CONFIG.AUTH_LOGOUT, {});

            // Clear tokens
            clearAuthToken();

            // Update state
            setState({
                isAuthenticated: false,
                isLoading: false,
                user: null,
                token: null,
                error: null,
            });

            // Redirect to login
            router.push('/login');
        } catch (err) {
            console.error('[useAuth] Logout failed:', err);

            // Still clear tokens even if backend call fails
            clearAuthToken();
            setState({
                isAuthenticated: false,
                isLoading: false,
                user: null,
                token: null,
                error: null,
            });
            router.push('/login');
        }
    }, [router]);

    /**
     * Check authentication on mount
     */
    useEffect(() => {
        checkAuth();
    }, [checkAuth]);

    return {
        isAuthenticated: state.isAuthenticated,
        isLoading: state.isLoading,
        user: state.user,
        token: state.token,
        error: state.error,
        login,
        logout,
        checkAuth,
    };
}
