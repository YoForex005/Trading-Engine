'use client';

import { useState } from 'react';
import { useAuth } from '@/hooks/useAuth';

/**
 * LoginPage Component
 * Professional dark-themed admin login matching RTX5 branding
 * Authenticates against backend POST /login endpoint
 */
export default function LoginPage() {
    const { login, isLoading, error: authError } = useAuth();
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        // Basic validation
        if (!username.trim() || !password.trim()) {
            setError('Username and password are required');
            return;
        }

        try {
            await login(username, password);
            // If successful, AuthGuard will handle redirect
        } catch (err) {
            // Error is already set in useAuth hook
            console.error('[LoginPage] Login failed:', err);
        }
    };

    return (
        <div className="min-h-screen bg-[#0A0A0B] flex items-center justify-center px-4">
            <div className="w-full max-w-md">
                {/* Logo / Branding */}
                <div className="text-center mb-8">
                    <div className="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-br from-[#F5C542] to-[#D4AC0D] rounded-2xl mb-4 shadow-lg">
                        <span className="text-3xl font-bold text-black">RTX</span>
                    </div>
                    <h1 className="text-2xl font-bold text-white mb-2">
                        Broker Admin Panel
                    </h1>
                    <p className="text-[#888] text-sm">
                        Sign in to manage your trading platform
                    </p>
                </div>

                {/* Login Form */}
                <div className="bg-[#1E2026] border border-[#383A42] rounded-lg p-8 shadow-2xl">
                    <form onSubmit={handleSubmit} className="space-y-6">
                        {/* Username Field */}
                        <div>
                            <label
                                htmlFor="username"
                                className="block text-sm font-medium text-[#CCC] mb-2"
                            >
                                Username
                            </label>
                            <input
                                id="username"
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                disabled={isLoading}
                                placeholder="Enter your username"
                                className="w-full px-4 py-3 bg-[#121316] border border-[#383A42] rounded-lg text-white placeholder-[#666] focus:outline-none focus:border-[#F5C542] focus:ring-1 focus:ring-[#F5C542] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                autoComplete="username"
                                autoFocus
                            />
                        </div>

                        {/* Password Field */}
                        <div>
                            <label
                                htmlFor="password"
                                className="block text-sm font-medium text-[#CCC] mb-2"
                            >
                                Password
                            </label>
                            <input
                                id="password"
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                disabled={isLoading}
                                placeholder="Enter your password"
                                className="w-full px-4 py-3 bg-[#121316] border border-[#383A42] rounded-lg text-white placeholder-[#666] focus:outline-none focus:border-[#F5C542] focus:ring-1 focus:ring-[#F5C542] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                autoComplete="current-password"
                            />
                        </div>

                        {/* Error Message */}
                        {(error || authError) && (
                            <div className="bg-red-900/20 border border-red-600/50 rounded-lg p-3 flex items-start gap-2">
                                <svg
                                    className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        strokeLinecap="round"
                                        strokeLinejoin="round"
                                        strokeWidth={2}
                                        d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                    />
                                </svg>
                                <span className="text-sm text-red-400">
                                    {error || authError}
                                </span>
                            </div>
                        )}

                        {/* Login Button */}
                        <button
                            type="submit"
                            disabled={isLoading}
                            className="w-full py-3 bg-gradient-to-r from-[#F5C542] to-[#D4AC0D] hover:from-[#D4AC0D] hover:to-[#F5C542] text-black font-bold rounded-lg transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl flex items-center justify-center gap-2"
                        >
                            {isLoading ? (
                                <>
                                    <span className="inline-block w-4 h-4 border-2 border-black border-t-transparent rounded-full animate-spin" />
                                    <span>Signing in...</span>
                                </>
                            ) : (
                                <span>Sign In</span>
                            )}
                        </button>
                    </form>

                    {/* Default Credentials Hint (for demo) */}
                    <div className="mt-6 pt-6 border-t border-[#383A42]">
                        <p className="text-xs text-[#666] text-center">
                            Default credentials: <span className="text-[#888] font-mono">admin</span> / <span className="text-[#888] font-mono">Admin@123</span>
                        </p>
                    </div>
                </div>

                {/* Footer */}
                <div className="mt-6 text-center text-xs text-[#666]">
                    RTX5 Trading Platform &copy; {new Date().getFullYear()}
                </div>
            </div>
        </div>
    );
}
