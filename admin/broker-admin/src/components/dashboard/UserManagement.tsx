'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
    Users,
    Plus,
    Edit,
    Power,
    Search,
    X,
    Shield,
    Key,
    RefreshCw,
    Eye,
    EyeOff,
    Copy,
    Check,
} from 'lucide-react';
import { api } from '@/services/apiClient';
import { API_CONFIG } from '@/config/api';

type UserRole = 'SUPER_ADMIN' | 'BROKER_ADMIN' | 'RISK_MANAGER' | 'DEALER' | 'ACCOUNTANT' | 'SUPPORT' | 'VIEWER';
type UserStatus = 'ACTIVE' | 'DISABLED';

interface AdminUser {
    id: string;
    username: string;
    email: string;
    role: UserRole;
    status: UserStatus;
    twoFactorEnabled: boolean;
    lastLogin?: string;
    createdAt: string;
}

interface AdminUserFormData {
    username: string;
    email: string;
    password: string;
    role: UserRole;
}

type SortField = 'username' | 'email' | 'role' | 'status' | 'lastLogin' | 'createdAt';
type SortDirection = 'asc' | 'desc';

const ROLE_COLORS: Record<UserRole, string> = {
    SUPER_ADMIN: 'bg-red-600 text-white',
    BROKER_ADMIN: 'bg-blue-600 text-white',
    DEALER: 'bg-green-600 text-white',
    RISK_MANAGER: 'bg-yellow-600 text-black',
    SUPPORT: 'bg-gray-600 text-white',
    VIEWER: 'bg-gray-500 text-white',
    ACCOUNTANT: 'bg-purple-600 text-white',
};

const ROLES: UserRole[] = [
    'SUPER_ADMIN',
    'BROKER_ADMIN',
    'RISK_MANAGER',
    'DEALER',
    'ACCOUNTANT',
    'SUPPORT',
    'VIEWER'
];

export default function UserManagement() {
    const [users, setUsers] = useState<AdminUser[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showEditModal, setShowEditModal] = useState(false);
    const [showPasswordModal, setShowPasswordModal] = useState(false);
    const [editingUser, setEditingUser] = useState<AdminUser | null>(null);
    const [tempPassword, setTempPassword] = useState('');
    const [sortField, setSortField] = useState<SortField>('createdAt');
    const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
    const [currentPage, setCurrentPage] = useState(1);

    const pageSize = 25;

    // Fetch users on mount
    useEffect(() => {
        fetchUsers();
    }, []);

    const fetchUsers = async () => {
        try {
            setLoading(true);
            const data = await api.get<AdminUser[]>(API_CONFIG.ADMIN_USERS);
            setUsers(data || []);
        } catch (error) {
            console.error('[UserManagement] Failed to fetch users:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateUser = () => {
        setShowCreateModal(true);
    };

    const handleEditUser = (user: AdminUser) => {
        setEditingUser(user);
        setShowEditModal(true);
    };

    const handleToggleStatus = async (user: AdminUser) => {
        const action = user.status === 'ACTIVE' ? 'disable' : 'enable';
        const confirmMsg = `Are you sure you want to ${action} user "${user.username}"?`;

        if (!confirm(confirmMsg)) return;

        try {
            const endpoint = user.status === 'ACTIVE'
                ? API_CONFIG.ADMIN_USER_DISABLE(user.id)
                : API_CONFIG.ADMIN_USER_ENABLE(user.id);

            await api.post(endpoint);
            await fetchUsers();
        } catch (error) {
            console.error('[UserManagement] Failed to toggle status:', error);
            alert('Failed to update user status');
        }
    };

    const handleResetPassword = async (user: AdminUser) => {
        const confirmMsg = `Reset password for "${user.username}"? A temporary password will be generated.`;

        if (!confirm(confirmMsg)) return;

        try {
            const response = await api.post<{ tempPassword: string }>(
                API_CONFIG.ADMIN_USER_RESET_PASSWORD(user.id)
            );
            setTempPassword(response.tempPassword);
            setEditingUser(user);
            setShowPasswordModal(true);
        } catch (error) {
            console.error('[UserManagement] Failed to reset password:', error);
            alert('Failed to reset password');
        }
    };

    // Filter users by search query
    const filteredUsers = useMemo(() => {
        return users.filter(user => {
            const searchLower = searchQuery.toLowerCase();
            return (
                user.username.toLowerCase().includes(searchLower) ||
                user.email.toLowerCase().includes(searchLower)
            );
        });
    }, [users, searchQuery]);

    // Sort users
    const sortedUsers = useMemo(() => {
        const sorted = [...filteredUsers].sort((a, b) => {
            let aVal: any = a[sortField];
            let bVal: any = b[sortField];

            // Handle date sorting
            if (sortField === 'lastLogin' || sortField === 'createdAt') {
                aVal = aVal ? new Date(aVal).getTime() : 0;
                bVal = bVal ? new Date(bVal).getTime() : 0;
            }

            // Handle string sorting (case-insensitive)
            if (typeof aVal === 'string') {
                aVal = aVal.toLowerCase();
                bVal = bVal.toLowerCase();
            }

            if (sortDirection === 'asc') {
                return aVal > bVal ? 1 : aVal < bVal ? -1 : 0;
            } else {
                return aVal < bVal ? 1 : aVal > bVal ? -1 : 0;
            }
        });
        return sorted;
    }, [filteredUsers, sortField, sortDirection]);

    // Paginate users
    const paginatedUsers = useMemo(() => {
        const startIdx = (currentPage - 1) * pageSize;
        return sortedUsers.slice(startIdx, startIdx + pageSize);
    }, [sortedUsers, currentPage]);

    const totalPages = Math.ceil(sortedUsers.length / pageSize);

    // Handle sort
    const handleSort = (field: SortField) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('asc');
        }
    };

    // Format timestamp
    const formatTimestamp = (timestamp: string | undefined): string => {
        if (!timestamp) return 'Never';
        const date = new Date(timestamp);
        return date.toLocaleString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    return (
        <div className="h-full bg-[#0A0A0B] flex flex-col overflow-hidden">
            {/* Header */}
            <div className="flex-shrink-0 p-4 border-b border-zinc-800">
                <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-3">
                        <Users className="w-6 h-6 text-blue-400" />
                        <h1 className="text-2xl font-bold text-white">Admin Users</h1>
                        <span className="text-sm text-zinc-400">
                            ({sortedUsers.length} {sortedUsers.length === 1 ? 'user' : 'users'})
                        </span>
                    </div>
                    <button
                        onClick={handleCreateUser}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
                    >
                        <Plus className="w-4 h-4" />
                        <span className="font-medium">Create User</span>
                    </button>
                </div>

                {/* Search */}
                <div className="relative max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                    <input
                        type="text"
                        value={searchQuery}
                        onChange={(e) => {
                            setSearchQuery(e.target.value);
                            setCurrentPage(1);
                        }}
                        placeholder="Search by name or email..."
                        className="w-full pl-10 pr-4 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-white text-sm placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                    />
                </div>
            </div>

            {/* Table */}
            <div className="flex-1 overflow-auto">
                {loading ? (
                    <div className="flex items-center justify-center h-full">
                        <div className="text-center">
                            <RefreshCw className="w-8 h-8 animate-spin text-blue-500 mx-auto mb-3" />
                            <p className="text-zinc-400">Loading users...</p>
                        </div>
                    </div>
                ) : (
                    <table className="w-full text-sm">
                        <thead className="bg-zinc-800 sticky top-0 z-10">
                            <tr>
                                <th
                                    onClick={() => handleSort('username')}
                                    className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                >
                                    <div className="flex items-center gap-1">
                                        Name
                                        {sortField === 'username' && (
                                            <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                        )}
                                    </div>
                                </th>
                                <th
                                    onClick={() => handleSort('email')}
                                    className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                >
                                    <div className="flex items-center gap-1">
                                        Email
                                        {sortField === 'email' && (
                                            <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                        )}
                                    </div>
                                </th>
                                <th
                                    onClick={() => handleSort('role')}
                                    className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                >
                                    <div className="flex items-center gap-1">
                                        Role
                                        {sortField === 'role' && (
                                            <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                        )}
                                    </div>
                                </th>
                                <th
                                    onClick={() => handleSort('status')}
                                    className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                >
                                    <div className="flex items-center gap-1">
                                        Status
                                        {sortField === 'status' && (
                                            <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                        )}
                                    </div>
                                </th>
                                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400">
                                    2FA
                                </th>
                                <th
                                    onClick={() => handleSort('lastLogin')}
                                    className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                >
                                    <div className="flex items-center gap-1">
                                        Last Login
                                        {sortField === 'lastLogin' && (
                                            <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                        )}
                                    </div>
                                </th>
                                <th
                                    onClick={() => handleSort('createdAt')}
                                    className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 cursor-pointer hover:text-white"
                                >
                                    <div className="flex items-center gap-1">
                                        Created
                                        {sortField === 'createdAt' && (
                                            <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                                        )}
                                    </div>
                                </th>
                                <th className="px-4 py-3 text-center text-xs font-semibold text-zinc-400">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {paginatedUsers.length === 0 ? (
                                <tr>
                                    <td colSpan={8} className="px-4 py-12 text-center text-zinc-500">
                                        No users found
                                    </td>
                                </tr>
                            ) : (
                                paginatedUsers.map((user) => (
                                    <tr
                                        key={user.id}
                                        className="border-b border-zinc-800 hover:bg-zinc-900/50"
                                    >
                                        <td className="px-4 py-3 text-white font-medium">
                                            {user.username}
                                        </td>
                                        <td className="px-4 py-3 text-zinc-300">
                                            {user.email}
                                        </td>
                                        <td className="px-4 py-3">
                                            <span className={`px-2 py-1 rounded text-xs font-semibold ${ROLE_COLORS[user.role]}`}>
                                                {user.role}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3">
                                            <span className={`px-2 py-1 rounded text-xs font-semibold ${
                                                user.status === 'ACTIVE'
                                                    ? 'bg-green-600 text-white'
                                                    : 'bg-red-600 text-white'
                                            }`}>
                                                {user.status}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3">
                                            <Shield
                                                className={`w-4 h-4 ${
                                                    user.twoFactorEnabled
                                                        ? 'text-green-400'
                                                        : 'text-gray-600'
                                                }`}
                                            />
                                        </td>
                                        <td className="px-4 py-3 text-zinc-400 text-xs font-mono">
                                            {formatTimestamp(user.lastLogin)}
                                        </td>
                                        <td className="px-4 py-3 text-zinc-400 text-xs font-mono">
                                            {formatTimestamp(user.createdAt)}
                                        </td>
                                        <td className="px-4 py-3">
                                            <div className="flex items-center justify-center gap-1">
                                                <button
                                                    onClick={() => handleEditUser(user)}
                                                    className="p-1.5 text-blue-400 hover:bg-blue-500/20 rounded transition-colors"
                                                    title="Edit user"
                                                >
                                                    <Edit className="w-4 h-4" />
                                                </button>
                                                <button
                                                    onClick={() => handleToggleStatus(user)}
                                                    className={`p-1.5 rounded transition-colors ${
                                                        user.status === 'ACTIVE'
                                                            ? 'text-red-400 hover:bg-red-500/20'
                                                            : 'text-green-400 hover:bg-green-500/20'
                                                    }`}
                                                    title={user.status === 'ACTIVE' ? 'Disable user' : 'Enable user'}
                                                >
                                                    <Power className="w-4 h-4" />
                                                </button>
                                                <button
                                                    onClick={() => handleResetPassword(user)}
                                                    className="p-1.5 text-yellow-400 hover:bg-yellow-500/20 rounded transition-colors"
                                                    title="Reset password"
                                                >
                                                    <Key className="w-4 h-4" />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                )}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
                <div className="flex-shrink-0 flex items-center justify-between px-4 py-3 border-t border-zinc-800 bg-zinc-900">
                    <div className="text-sm text-zinc-400">
                        Showing {((currentPage - 1) * pageSize) + 1} - {Math.min(currentPage * pageSize, sortedUsers.length)} of {sortedUsers.length}
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                            disabled={currentPage === 1}
                            className="px-3 py-1 bg-zinc-800 text-zinc-300 rounded text-sm hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Previous
                        </button>
                        <div className="text-sm text-zinc-400">
                            Page {currentPage} of {totalPages}
                        </div>
                        <button
                            onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                            disabled={currentPage === totalPages}
                            className="px-3 py-1 bg-zinc-800 text-zinc-300 rounded text-sm hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Next
                        </button>
                    </div>
                </div>
            )}

            {/* Modals */}
            {showCreateModal && (
                <CreateUserModal
                    onClose={() => setShowCreateModal(false)}
                    onSave={() => {
                        fetchUsers();
                        setShowCreateModal(false);
                    }}
                />
            )}

            {showEditModal && editingUser && (
                <EditUserModal
                    user={editingUser}
                    onClose={() => {
                        setShowEditModal(false);
                        setEditingUser(null);
                    }}
                    onSave={() => {
                        fetchUsers();
                        setShowEditModal(false);
                        setEditingUser(null);
                    }}
                />
            )}

            {showPasswordModal && editingUser && (
                <PasswordResetModal
                    user={editingUser}
                    tempPassword={tempPassword}
                    onClose={() => {
                        setShowPasswordModal(false);
                        setEditingUser(null);
                        setTempPassword('');
                    }}
                />
            )}
        </div>
    );
}

// Create User Modal
interface CreateUserModalProps {
    onClose: () => void;
    onSave: () => void;
}

function CreateUserModal({ onClose, onSave }: CreateUserModalProps) {
    const [formData, setFormData] = useState<AdminUserFormData>({
        username: '',
        email: '',
        password: '',
        role: 'VIEWER',
    });
    const [autoGenerate, setAutoGenerate] = useState(false);
    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    // Generate random password
    const generatePassword = () => {
        const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*';
        let password = '';
        for (let i = 0; i < 16; i++) {
            password += charset[Math.floor(Math.random() * charset.length)];
        }
        return password;
    };

    // Handle auto-generate toggle
    useEffect(() => {
        if (autoGenerate) {
            setFormData(prev => ({ ...prev, password: generatePassword() }));
        }
    }, [autoGenerate]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            await api.post(API_CONFIG.ADMIN_USERS, formData);
            onSave();
        } catch (err: any) {
            setError(err.data?.message || err.message || 'Failed to create user');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-lg shadow-2xl">
                {/* Header */}
                <div className="flex items-center justify-between border-b border-zinc-800 px-6 py-4">
                    <h3 className="text-lg font-semibold text-white">Create Admin User</h3>
                    <button onClick={onClose} className="text-zinc-400 hover:text-white">
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    {error && (
                        <div className="bg-red-500/10 border border-red-500 text-red-400 px-4 py-3 rounded text-sm">
                            {error}
                        </div>
                    )}

                    {/* Name */}
                    <div>
                        <label className="block text-sm text-zinc-400 mb-2">Name</label>
                        <input
                            type="text"
                            value={formData.username}
                            onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                            className="w-full bg-zinc-800 border border-zinc-700 text-white px-3 py-2 rounded focus:outline-none focus:border-blue-500"
                            required
                        />
                    </div>

                    {/* Email */}
                    <div>
                        <label className="block text-sm text-zinc-400 mb-2">Email</label>
                        <input
                            type="email"
                            value={formData.email}
                            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                            className="w-full bg-zinc-800 border border-zinc-700 text-white px-3 py-2 rounded focus:outline-none focus:border-blue-500"
                            required
                        />
                    </div>

                    {/* Password */}
                    <div>
                        <label className="block text-sm text-zinc-400 mb-2">Password</label>
                        <div className="space-y-2">
                            <div className="flex items-center gap-2">
                                <label className="flex items-center gap-2 cursor-pointer">
                                    <input
                                        type="checkbox"
                                        checked={autoGenerate}
                                        onChange={(e) => setAutoGenerate(e.target.checked)}
                                        className="w-4 h-4"
                                    />
                                    <span className="text-sm text-zinc-400">Auto-generate</span>
                                </label>
                            </div>
                            <div className="relative">
                                <input
                                    type={showPassword ? 'text' : 'password'}
                                    value={formData.password}
                                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                                    className="w-full bg-zinc-800 border border-zinc-700 text-white px-3 py-2 pr-10 rounded focus:outline-none focus:border-blue-500 font-mono"
                                    required
                                    disabled={autoGenerate}
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-white"
                                >
                                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Role */}
                    <div>
                        <label className="block text-sm text-zinc-400 mb-2">Role</label>
                        <select
                            value={formData.role}
                            onChange={(e) => setFormData({ ...formData, role: e.target.value as UserRole })}
                            className="w-full bg-zinc-800 border border-zinc-700 text-white px-3 py-2 rounded focus:outline-none focus:border-blue-500"
                        >
                            {ROLES.map(role => (
                                <option key={role} value={role}>{role}</option>
                            ))}
                        </select>
                    </div>

                    {/* Buttons */}
                    <div className="flex gap-3 pt-4">
                        <button
                            type="button"
                            onClick={onClose}
                            className="flex-1 px-4 py-2 bg-zinc-800 text-white rounded hover:bg-zinc-700 transition-colors"
                            disabled={loading}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors disabled:opacity-50"
                            disabled={loading}
                        >
                            {loading ? 'Creating...' : 'Create'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}

// Edit User Modal
interface EditUserModalProps {
    user: AdminUser;
    onClose: () => void;
    onSave: () => void;
}

function EditUserModal({ user, onClose, onSave }: EditUserModalProps) {
    const [formData, setFormData] = useState({
        username: user.username,
        role: user.role,
        forcePasswordReset: false,
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            await api.put(API_CONFIG.ADMIN_USER_BY_ID(user.id), {
                username: formData.username,
                role: formData.role,
                forcePasswordReset: formData.forcePasswordReset,
            });
            onSave();
        } catch (err: any) {
            setError(err.data?.message || err.message || 'Failed to update user');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-lg shadow-2xl">
                {/* Header */}
                <div className="flex items-center justify-between border-b border-zinc-800 px-6 py-4">
                    <h3 className="text-lg font-semibold text-white">Edit Admin User</h3>
                    <button onClick={onClose} className="text-zinc-400 hover:text-white">
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    {error && (
                        <div className="bg-red-500/10 border border-red-500 text-red-400 px-4 py-3 rounded text-sm">
                            {error}
                        </div>
                    )}

                    {/* Name */}
                    <div>
                        <label className="block text-sm text-zinc-400 mb-2">Name</label>
                        <input
                            type="text"
                            value={formData.username}
                            onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                            className="w-full bg-zinc-800 border border-zinc-700 text-white px-3 py-2 rounded focus:outline-none focus:border-blue-500"
                            required
                        />
                    </div>

                    {/* Email (read-only) */}
                    <div>
                        <label className="block text-sm text-zinc-400 mb-2">Email</label>
                        <input
                            type="email"
                            value={user.email}
                            className="w-full bg-zinc-800/50 border border-zinc-700 text-zinc-500 px-3 py-2 rounded cursor-not-allowed"
                            disabled
                        />
                    </div>

                    {/* Role */}
                    <div>
                        <label className="block text-sm text-zinc-400 mb-2">Role</label>
                        <select
                            value={formData.role}
                            onChange={(e) => setFormData({ ...formData, role: e.target.value as UserRole })}
                            className="w-full bg-zinc-800 border border-zinc-700 text-white px-3 py-2 rounded focus:outline-none focus:border-blue-500"
                        >
                            {ROLES.map(role => (
                                <option key={role} value={role}>{role}</option>
                            ))}
                        </select>
                    </div>

                    {/* Force Password Reset */}
                    <div>
                        <label className="flex items-center gap-2 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={formData.forcePasswordReset}
                                onChange={(e) => setFormData({ ...formData, forcePasswordReset: e.target.checked })}
                                className="w-4 h-4"
                            />
                            <span className="text-sm text-zinc-400">Force password reset on next login</span>
                        </label>
                    </div>

                    {/* Buttons */}
                    <div className="flex gap-3 pt-4">
                        <button
                            type="button"
                            onClick={onClose}
                            className="flex-1 px-4 py-2 bg-zinc-800 text-white rounded hover:bg-zinc-700 transition-colors"
                            disabled={loading}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors disabled:opacity-50"
                            disabled={loading}
                        >
                            {loading ? 'Saving...' : 'Save'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}

// Password Reset Modal
interface PasswordResetModalProps {
    user: AdminUser;
    tempPassword: string;
    onClose: () => void;
}

function PasswordResetModal({ user, tempPassword, onClose }: PasswordResetModalProps) {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(tempPassword);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-lg shadow-2xl">
                {/* Header */}
                <div className="flex items-center justify-between border-b border-zinc-800 px-6 py-4">
                    <h3 className="text-lg font-semibold text-white">Password Reset</h3>
                    <button onClick={onClose} className="text-zinc-400 hover:text-white">
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Content */}
                <div className="p-6 space-y-4">
                    <div className="bg-yellow-500/10 border border-yellow-500 text-yellow-400 px-4 py-3 rounded text-sm">
                        <p className="font-semibold mb-1">Temporary password generated for {user.username}</p>
                        <p className="text-xs">User must change this password on next login.</p>
                    </div>

                    <div>
                        <label className="block text-sm text-zinc-400 mb-2">Temporary Password</label>
                        <div className="relative">
                            <input
                                type="text"
                                value={tempPassword}
                                readOnly
                                className="w-full bg-zinc-800 border border-zinc-700 text-white px-3 py-2 pr-10 rounded font-mono text-sm"
                            />
                            <button
                                type="button"
                                onClick={handleCopy}
                                className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-white"
                                title="Copy password"
                            >
                                {copied ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4" />}
                            </button>
                        </div>
                        {copied && (
                            <p className="text-xs text-green-400 mt-1">Copied to clipboard!</p>
                        )}
                    </div>

                    <div className="bg-blue-500/10 border border-blue-500 text-blue-400 px-4 py-3 rounded text-xs">
                        <p className="font-semibold mb-1">Important:</p>
                        <ul className="list-disc list-inside space-y-1">
                            <li>Save this password securely - it won't be shown again</li>
                            <li>Share it with the user through a secure channel</li>
                            <li>User will be forced to change it on next login</li>
                        </ul>
                    </div>

                    {/* Close Button */}
                    <button
                        onClick={onClose}
                        className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>
    );
}
