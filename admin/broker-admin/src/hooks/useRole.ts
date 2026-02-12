/**
 * useRole Hook
 * Manages user role and permissions for RBAC-aware UI
 */

import { useState, useEffect, useCallback } from 'react';
import { API_CONFIG } from '@/config/api';
import { api } from '@/services/apiClient';

export type AdminRole =
    | 'SUPER_ADMIN'
    | 'BROKER_ADMIN'
    | 'ADMIN'
    | 'MANAGER'
    | 'DEALER'
    | 'SUPPORT'
    | 'VIEWER';

export type Permission =
    | 'ACCOUNTS_READ'
    | 'ACCOUNTS_WRITE'
    | 'TRADING_READ'
    | 'TRADING_EXECUTE'
    | 'SETTINGS_READ'
    | 'SETTINGS_WRITE'
    | 'USERS_READ'
    | 'USERS_MANAGE'
    | 'REPORTS_READ'
    | 'LP_MANAGE'
    | 'RISK_MANAGE';

interface RoleData {
    role: AdminRole;
    permissions: Permission[];
}

const ROLE_STORAGE_KEY = 'admin_role';
const PERMISSIONS_STORAGE_KEY = 'admin_permissions';

/**
 * Role hierarchy mapping (higher value = more privilege)
 */
const ROLE_HIERARCHY: Record<AdminRole, number> = {
    SUPER_ADMIN: 100,
    BROKER_ADMIN: 80,
    ADMIN: 80, // Alias for BROKER_ADMIN
    MANAGER: 60,
    DEALER: 40,
    SUPPORT: 20,
    VIEWER: 10,
};

/**
 * Default permissions by role (frontend fallback)
 */
const DEFAULT_PERMISSIONS: Record<AdminRole, Permission[]> = {
    SUPER_ADMIN: [
        'ACCOUNTS_READ', 'ACCOUNTS_WRITE', 'TRADING_READ', 'TRADING_EXECUTE',
        'SETTINGS_READ', 'SETTINGS_WRITE', 'USERS_READ', 'USERS_MANAGE',
        'REPORTS_READ', 'LP_MANAGE', 'RISK_MANAGE'
    ],
    BROKER_ADMIN: [
        'ACCOUNTS_READ', 'ACCOUNTS_WRITE', 'TRADING_READ', 'TRADING_EXECUTE',
        'SETTINGS_READ', 'SETTINGS_WRITE', 'USERS_READ', 'REPORTS_READ',
        'LP_MANAGE', 'RISK_MANAGE'
    ],
    ADMIN: [
        'ACCOUNTS_READ', 'ACCOUNTS_WRITE', 'TRADING_READ', 'TRADING_EXECUTE',
        'SETTINGS_READ', 'SETTINGS_WRITE', 'USERS_READ', 'REPORTS_READ',
        'LP_MANAGE', 'RISK_MANAGE'
    ],
    MANAGER: ['ACCOUNTS_READ', 'TRADING_READ', 'REPORTS_READ', 'SETTINGS_READ'],
    DEALER: ['ACCOUNTS_READ', 'TRADING_READ', 'TRADING_EXECUTE'],
    SUPPORT: ['ACCOUNTS_READ', 'TRADING_READ', 'SETTINGS_READ', 'REPORTS_READ'],
    VIEWER: ['ACCOUNTS_READ', 'TRADING_READ', 'SETTINGS_READ', 'REPORTS_READ'],
};

export function useRole() {
    const [role, setRole] = useState<AdminRole | null>(null);
    const [permissions, setPermissions] = useState<Permission[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    /**
     * Load role from localStorage cache
     */
    const loadFromCache = useCallback((): boolean => {
        if (typeof window === 'undefined') return false;

        const cachedRole = localStorage.getItem(ROLE_STORAGE_KEY) as AdminRole | null;
        const cachedPerms = localStorage.getItem(PERMISSIONS_STORAGE_KEY);

        if (cachedRole) {
            setRole(cachedRole);
            if (cachedPerms) {
                try {
                    setPermissions(JSON.parse(cachedPerms));
                } catch {
                    setPermissions(DEFAULT_PERMISSIONS[cachedRole] || []);
                }
            } else {
                setPermissions(DEFAULT_PERMISSIONS[cachedRole] || []);
            }
            return true;
        }
        return false;
    }, []);

    /**
     * Fetch role and permissions from backend
     */
    const fetchRole = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);

            // Try to fetch role from backend
            const response = await api.get<RoleData>(API_CONFIG.ADMIN_ROLE);

            const userRole = response.role;
            const userPermissions = response.permissions || DEFAULT_PERMISSIONS[userRole] || [];

            setRole(userRole);
            setPermissions(userPermissions);

            // Cache in localStorage
            if (typeof window !== 'undefined') {
                localStorage.setItem(ROLE_STORAGE_KEY, userRole);
                localStorage.setItem(PERMISSIONS_STORAGE_KEY, JSON.stringify(userPermissions));
            }
        } catch (err) {
            console.error('[useRole] Failed to fetch role:', err);
            setError(err instanceof Error ? err.message : 'Failed to fetch role');

            // Fallback to cache if API fails
            if (!loadFromCache()) {
                // Default to VIEWER if no cache and API fails
                setRole('VIEWER');
                setPermissions(DEFAULT_PERMISSIONS.VIEWER);
            }
        } finally {
            setLoading(false);
        }
    }, [loadFromCache]);

    /**
     * Check if user has a specific permission
     */
    const hasPermission = useCallback((permission: Permission): boolean => {
        return permissions.includes(permission);
    }, [permissions]);

    /**
     * Check if user role meets or exceeds required role
     */
    const hasRole = useCallback((requiredRole: AdminRole): boolean => {
        if (!role) return false;
        const userLevel = ROLE_HIERARCHY[role] || 0;
        const requiredLevel = ROLE_HIERARCHY[requiredRole] || 0;
        return userLevel >= requiredLevel;
    }, [role]);

    /**
     * Check if user is admin or higher
     */
    const isAdmin = useCallback((): boolean => {
        return hasRole('ADMIN');
    }, [hasRole]);

    /**
     * Check if user is super admin
     */
    const isSuperAdmin = useCallback((): boolean => {
        return role === 'SUPER_ADMIN';
    }, [role]);

    /**
     * Clear role cache (on logout)
     */
    const clearRole = useCallback(() => {
        setRole(null);
        setPermissions([]);
        if (typeof window !== 'undefined') {
            localStorage.removeItem(ROLE_STORAGE_KEY);
            localStorage.removeItem(PERMISSIONS_STORAGE_KEY);
        }
    }, []);

    /**
     * Initialize role on mount
     */
    useEffect(() => {
        // Load from cache first for instant UI
        const hasCache = loadFromCache();

        // Then fetch fresh data from backend
        fetchRole();
    }, [loadFromCache, fetchRole]);

    return {
        role,
        permissions,
        loading,
        error,
        hasPermission,
        hasRole,
        isAdmin,
        isSuperAdmin,
        refetch: fetchRole,
        clearRole,
    };
}
