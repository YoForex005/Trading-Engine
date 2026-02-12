'use client';

import { useState, FormEvent } from 'react';
import { useAuth } from '@/hooks/useAuth';

export default function LoginPage() {
    const { login, isLoading, error: authError } = useAuth();
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setError('');

        try {
            await login(username, password);
            // If successful, useAuth will handle redirect
        } catch (err: any) {
            // Error is set in useAuth hook
            console.error('[Login] Error:', err);
        }
    };

    return (
        <div className="min-h-screen bg-[#0A0A0B] flex items-center justify-center p-4">
            <div className="w-full max-w-md">
                {/* RTX5 Branding */}
                <div className="text-center mb-8">
                    <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-[#F5C542] to-[#D4AC0D] rounded-lg mb-4">
                        <span className="text-2xl font-bold text-black">RTX</span>
                    </div>
                    <h1 className="text-2xl font-bold text-white mb-2">RTX Manager</h1>
                    <p className="text-[#888] text-sm">Broker Administration Dashboard</p>
                </div>

                {/* Login Form */}
                <div className="bg-[#1E2026] border border-[#383A42] shadow-2xl">
                    {/* Header */}
                    <div className="bg-[#2D2D30] border-b border-[#383A42] px-6 py-4">
                        <h2 className="text-white font-bold text-sm">Admin Login</h2>
                    </div>

                    {/* Form */}
                    <form onSubmit={handleSubmit} className="p-6 space-y-4">
                        {/* Error Message */}
                        {(error || authError) && (
                            <div className="bg-[#E74C3C]/10 border border-[#E74C3C] text-[#E74C3C] px-4 py-3 text-xs">
                                <div className="flex items-center gap-2">
                                    <span className="font-bold">⚠</span>
                                    <span>{error || authError}</span>
                                </div>
                            </div>
                        )}

                        {/* Username Field */}
                        <div>
                            <label className="block text-[#888] text-xs mb-2">
                                Username
                            </label>
                            <input
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                className="w-full bg-[#252526] border border-[#444] text-white px-3 py-2 text-sm focus:outline-none focus:border-[#F5C542]"
                                placeholder="Enter your username"
                                required
                                autoFocus
                                disabled={isLoading}
                            />
                        </div>

                        {/* Password Field */}
                        <div>
                            <label className="block text-[#888] text-xs mb-2">
                                Password
                            </label>
                            <input
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="w-full bg-[#252526] border border-[#444] text-white px-3 py-2 text-sm focus:outline-none focus:border-[#F5C542]"
                                placeholder="Enter your password"
                                required
                                disabled={isLoading}
                            />
                        </div>

                        {/* Login Button */}
                        <button
                            type="submit"
                            disabled={isLoading}
                            className={`
                                w-full h-10 font-bold text-sm transition-colors
                                ${isLoading
                                    ? 'bg-[#666] text-[#333] cursor-not-allowed'
                                    : 'bg-gradient-to-b from-[#F5C542] to-[#D4AC0D] text-black hover:from-[#FFD54F] hover:to-[#F5C542]'}
                            `}
                        >
                            {isLoading ? (
                                <span className="flex items-center justify-center gap-2">
                                    <span className="inline-block w-4 h-4 border-2 border-[#333] border-t-transparent rounded-full animate-spin" />
                                    Logging in...
                                </span>
                            ) : (
                                'Login'
                            )}
                        </button>
                    </form>

                    {/* Footer */}
                    <div className="bg-[#252526] border-t border-[#383A42] px-6 py-3 text-[10px] text-[#666]">
                        RTX Trading Engine v5.0 • Broker Terminal
                    </div>
                </div>

                {/* Additional Info */}
                <div className="mt-4 text-center text-[10px] text-[#555]">
                    Default credentials: <span className="text-[#888] font-mono">admin</span> / <span className="text-[#888] font-mono">Admin@123</span>
                </div>
            </div>
        </div>
    );
}
