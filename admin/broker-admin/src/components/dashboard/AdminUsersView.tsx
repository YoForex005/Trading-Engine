'use client';

import React, { useState, useEffect } from 'react';
import { Users, Plus, Edit, Power, Search, X } from 'lucide-react';
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
    lastLogin?: string;
    createdAt: string;
}

interface AdminUserFormData {
    username: string;
    email: string;
    password: string;
    role: UserRole;
    status: UserStatus;
}

const ROLE_COLORS: Record<UserRole, string> = {
    SUPER_ADMIN: 'bg-[#E74C3C] text-white',
    BROKER_ADMIN: 'bg-[#3B82F6] text-white',
    RISK_MANAGER: 'bg-[#E67E22] text-white',
    DEALER: 'bg-[#9B59B6] text-white',
    ACCOUNTANT: 'bg-[#2ECC71] text-white',
    SUPPORT: 'bg-[#F1C40F] text-black',
    VIEWER: 'bg-[#95A5A6] text-white',
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

export default function AdminUsersView() {
    const [users, setUsers] = useState<AdminUser[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [filterRole, setFilterRole] = useState<UserRole | 'ALL'>('ALL');
    const [showModal, setShowModal] = useState(false);
    const [editingUser, setEditingUser] = useState<AdminUser | null>(null);

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
            console.error('[AdminUsers] Failed to fetch users:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateUser = () => {
        setEditingUser(null);
        setShowModal(true);
    };

    const handleEditUser = (user: AdminUser) => {
        setEditingUser(user);
        setShowModal(true);
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
            await fetchUsers(); // Refresh list
        } catch (error) {
            console.error('[AdminUsers] Failed to toggle status:', error);
            alert('Failed to update user status');
        }
    };

    // Filter users
    const filteredUsers = users.filter(user => {
        const matchesSearch =
            user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
            user.email.toLowerCase().includes(searchQuery.toLowerCase());
        const matchesRole = filterRole === 'ALL' || user.role === filterRole;
        return matchesSearch && matchesRole;
    });

    return (
        <div className="flex flex-col h-full bg-[#121316] text-[11px]">
            {/* Header */}
            <div className="bg-[#1E2026] border-b border-[#383A42] px-4 py-3 flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <Users size={14} className="text-[#3B82F6]" />
                    <span className="text-white font-bold text-sm">Admin Users Management</span>
                    <span className="text-[#666] text-xs">({filteredUsers.length} users)</span>
                </div>
                <button
                    onClick={handleCreateUser}
                    className="flex items-center gap-1.5 px-3 py-1.5 bg-[#2ECC71] hover:bg-[#27AE60] text-white font-bold text-xs transition-colors"
                >
                    <Plus size={12} />
                    Create Admin
                </button>
            </div>

            {/* Filters */}
            <div className="bg-[#1E2026] border-b border-[#383A42] px-4 py-2 flex items-center gap-3">
                {/* Search */}
                <div className="flex items-center bg-[#252526] border border-[#444] px-2 flex-1 max-w-xs">
                    <Search size={12} className="text-[#888]" />
                    <input
                        type="text"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        placeholder="Search username or email..."
                        className="bg-transparent border-none outline-none h-6 text-white text-xs px-2 flex-1 placeholder-[#666]"
                    />
                </div>

                {/* Role Filter */}
                <select
                    value={filterRole}
                    onChange={(e) => setFilterRole(e.target.value as UserRole | 'ALL')}
                    className="bg-[#252526] border border-[#444] h-6 px-2 text-white text-xs"
                >
                    <option value="ALL">All Roles</option>
                    {ROLES.map(role => (
                        <option key={role} value={role}>{role}</option>
                    ))}
                </select>
            </div>

            {/* Table */}
            <div className="flex-1 overflow-auto">
                {loading ? (
                    <div className="flex items-center justify-center h-full text-[#666]">
                        <span>Loading users...</span>
                    </div>
                ) : (
                    <table className="w-full border-collapse">
                        <thead className="bg-[#2D2D30] text-[#A0A0A0] sticky top-0">
                            <tr>
                                <th className="border-r border-[#444] px-3 py-2 text-left font-normal">Username</th>
                                <th className="border-r border-[#444] px-3 py-2 text-left font-normal">Email</th>
                                <th className="border-r border-[#444] px-3 py-2 text-left font-normal">Role</th>
                                <th className="border-r border-[#444] px-3 py-2 text-left font-normal">Status</th>
                                <th className="border-r border-[#444] px-3 py-2 text-left font-normal">Last Login</th>
                                <th className="border-r border-[#444] px-3 py-2 text-left font-normal">Created</th>
                                <th className="px-3 py-2 text-center font-normal">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredUsers.map((user) => (
                                <tr
                                    key={user.id}
                                    className="border-b border-[#333] hover:bg-[#2A2D35]"
                                >
                                    <td className="px-3 py-2 border-r border-[#333] text-white font-mono">
                                        {user.username}
                                    </td>
                                    <td className="px-3 py-2 border-r border-[#333] text-[#CCC]">
                                        {user.email}
                                    </td>
                                    <td className="px-3 py-2 border-r border-[#333]">
                                        <span className={`px-2 py-1 text-[10px] font-bold ${ROLE_COLORS[user.role]}`}>
                                            {user.role}
                                        </span>
                                    </td>
                                    <td className="px-3 py-2 border-r border-[#333]">
                                        <span className={`px-2 py-1 text-[10px] font-bold ${
                                            user.status === 'ACTIVE'
                                                ? 'bg-[#2ECC71] text-white'
                                                : 'bg-[#E74C3C] text-white'
                                        }`}>
                                            {user.status}
                                        </span>
                                    </td>
                                    <td className="px-3 py-2 border-r border-[#333] text-[#888] font-mono text-[10px]">
                                        {user.lastLogin ? new Date(user.lastLogin).toLocaleString() : 'Never'}
                                    </td>
                                    <td className="px-3 py-2 border-r border-[#333] text-[#888] font-mono text-[10px]">
                                        {new Date(user.createdAt).toLocaleDateString()}
                                    </td>
                                    <td className="px-3 py-2 text-center">
                                        <div className="flex items-center justify-center gap-1">
                                            <button
                                                onClick={() => handleEditUser(user)}
                                                className="p-1 hover:bg-[#3B82F6] text-[#3B82F6] hover:text-white transition-colors"
                                                title="Edit user"
                                            >
                                                <Edit size={12} />
                                            </button>
                                            <button
                                                onClick={() => handleToggleStatus(user)}
                                                className={`p-1 transition-colors ${
                                                    user.status === 'ACTIVE'
                                                        ? 'hover:bg-[#E74C3C] text-[#E74C3C] hover:text-white'
                                                        : 'hover:bg-[#2ECC71] text-[#2ECC71] hover:text-white'
                                                }`}
                                                title={user.status === 'ACTIVE' ? 'Disable user' : 'Enable user'}
                                            >
                                                <Power size={12} />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                            {filteredUsers.length === 0 && (
                                <tr>
                                    <td colSpan={7} className="px-3 py-8 text-center text-[#666]">
                                        No users found
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                )}
            </div>

            {/* Modal */}
            {showModal && (
                <AdminUserModal
                    user={editingUser}
                    onClose={() => setShowModal(false)}
                    onSave={() => {
                        fetchUsers();
                        setShowModal(false);
                    }}
                />
            )}
        </div>
    );
}

// Admin User Modal Component
interface AdminUserModalProps {
    user: AdminUser | null;
    onClose: () => void;
    onSave: () => void;
}

function AdminUserModal({ user, onClose, onSave }: AdminUserModalProps) {
    const [formData, setFormData] = useState<AdminUserFormData>({
        username: user?.username || '',
        email: user?.email || '',
        password: '',
        role: user?.role || 'VIEWER',
        status: user?.status || 'ACTIVE',
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            if (user) {
                // Update existing user
                await api.put(API_CONFIG.ADMIN_USER_BY_ID(user.id), {
                    email: formData.email,
                    role: formData.role,
                    status: formData.status,
                    ...(formData.password && { password: formData.password }),
                });
            } else {
                // Create new user
                await api.post(API_CONFIG.ADMIN_USERS, formData);
            }
            onSave();
        } catch (err: any) {
            setError(err.data?.message || err.message || 'Failed to save user');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="w-full max-w-md bg-[#1E1E1E] border border-[#444] shadow-2xl">
                {/* Header */}
                <div className="flex items-center justify-between bg-[#2D2D30] border-b border-[#444] px-4 py-3">
                    <span className="text-white font-bold text-sm">
                        {user ? 'Edit Admin User' : 'Create Admin User'}
                    </span>
                    <button onClick={onClose} className="text-[#888] hover:text-white">
                        <X size={14} />
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="p-4 space-y-3">
                    {error && (
                        <div className="bg-[#E74C3C]/10 border border-[#E74C3C] text-[#E74C3C] px-3 py-2 text-xs">
                            {error}
                        </div>
                    )}

                    {/* Username */}
                    <div>
                        <label className="block text-[#888] text-xs mb-1">Username</label>
                        <input
                            type="text"
                            value={formData.username}
                            onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                            className="w-full bg-[#252526] border border-[#444] text-white px-2 py-1.5 text-xs"
                            required
                            disabled={!!user} // Cannot change username when editing
                        />
                    </div>

                    {/* Email */}
                    <div>
                        <label className="block text-[#888] text-xs mb-1">Email</label>
                        <input
                            type="email"
                            value={formData.email}
                            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                            className="w-full bg-[#252526] border border-[#444] text-white px-2 py-1.5 text-xs"
                            required
                        />
                    </div>

                    {/* Password */}
                    <div>
                        <label className="block text-[#888] text-xs mb-1">
                            Password {user && '(leave blank to keep current)'}
                        </label>
                        <input
                            type="password"
                            value={formData.password}
                            onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                            className="w-full bg-[#252526] border border-[#444] text-white px-2 py-1.5 text-xs"
                            required={!user}
                        />
                    </div>

                    {/* Role */}
                    <div>
                        <label className="block text-[#888] text-xs mb-1">Role</label>
                        <select
                            value={formData.role}
                            onChange={(e) => setFormData({ ...formData, role: e.target.value as UserRole })}
                            className="w-full bg-[#252526] border border-[#444] text-white px-2 py-1.5 text-xs"
                        >
                            {ROLES.map(role => (
                                <option key={role} value={role}>{role}</option>
                            ))}
                        </select>
                    </div>

                    {/* Status */}
                    <div>
                        <label className="block text-[#888] text-xs mb-1">Status</label>
                        <div className="flex items-center gap-3">
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="status"
                                    value="ACTIVE"
                                    checked={formData.status === 'ACTIVE'}
                                    onChange={(e) => setFormData({ ...formData, status: e.target.value as UserStatus })}
                                />
                                <span className="text-white text-xs">Active</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="status"
                                    value="DISABLED"
                                    checked={formData.status === 'DISABLED'}
                                    onChange={(e) => setFormData({ ...formData, status: e.target.value as UserStatus })}
                                />
                                <span className="text-white text-xs">Disabled</span>
                            </label>
                        </div>
                    </div>

                    {/* Buttons */}
                    <div className="flex gap-2 pt-2">
                        <button
                            type="button"
                            onClick={onClose}
                            className="flex-1 h-8 bg-[#333] hover:bg-[#444] text-white text-xs font-bold transition-colors"
                            disabled={loading}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="flex-1 h-8 bg-[#2ECC71] hover:bg-[#27AE60] text-white text-xs font-bold transition-colors"
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
