'use client';

import React, { useState, useEffect } from 'react';
import { Shield, QrCode, Key, Loader2, CheckCircle2, XCircle, AlertTriangle, Copy, Download } from 'lucide-react';
import { API_CONFIG } from '@/config/api';
import { api } from '@/services/apiClient';

interface TOTPSetupResponse {
    secret: string;
    qrCodeUrl: string;
    message: string;
}

interface BackupCodesResponse {
    backupCodes: string[];
    message: string;
}

type SetupStep = 'initial' | 'qr-display' | 'verification' | 'completed';

export default function TwoFactorSetup() {
    const [is2FAEnabled, setIs2FAEnabled] = useState<boolean>(false);
    const [loading, setLoading] = useState(false);
    const [step, setStep] = useState<SetupStep>('initial');
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);

    // Setup data
    const [qrCodeUrl, setQrCodeUrl] = useState<string>('');
    const [secret, setSecret] = useState<string>('');
    const [verificationCode, setVerificationCode] = useState<string>('');
    const [backupCodes, setBackupCodes] = useState<string[]>([]);

    // Check 2FA status on mount
    useEffect(() => {
        checkTwoFactorStatus();
    }, []);

    const checkTwoFactorStatus = async () => {
        try {
            // This would be an endpoint like GET /admin/2fa/status
            // For now, we'll derive it from the login response or user profile
            // Placeholder - implement when backend provides status endpoint
            setIs2FAEnabled(false);
        } catch (err) {
            console.error('[2FA] Failed to check status:', err);
        }
    };

    /**
     * STEP 1: Initialize 2FA setup
     */
    const handleSetupStart = async () => {
        setLoading(true);
        setError(null);
        setSuccess(null);

        try {
            const response = await api.post<TOTPSetupResponse>(API_CONFIG.TOTP_SETUP);

            setQrCodeUrl(response.qrCodeUrl);
            setSecret(response.secret);
            setStep('qr-display');
            setSuccess('QR code generated successfully');
        } catch (err: any) {
            setError(err.data?.error || 'Failed to generate 2FA QR code');
        } finally {
            setLoading(false);
        }
    };

    /**
     * STEP 2: Verify TOTP code
     */
    const handleVerifyCode = async () => {
        if (verificationCode.length !== 6) {
            setError('Please enter a 6-digit code');
            return;
        }

        setLoading(true);
        setError(null);

        try {
            await api.post(API_CONFIG.TOTP_VERIFY, {
                code: verificationCode,
            });

            setStep('verification');
            setSuccess('Code verified successfully');
        } catch (err: any) {
            setError(err.data?.error || 'Invalid 2FA code. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    /**
     * STEP 3: Enable 2FA and get backup codes
     */
    const handleEnable2FA = async () => {
        setLoading(true);
        setError(null);

        try {
            const response = await api.post<BackupCodesResponse>(API_CONFIG.TOTP_ENABLE, {
                code: verificationCode,
            });

            setBackupCodes(response.backupCodes);
            setIs2FAEnabled(true);
            setStep('completed');
            setSuccess('2FA enabled successfully! Save your backup codes.');
        } catch (err: any) {
            setError(err.data?.error || 'Failed to enable 2FA');
        } finally {
            setLoading(false);
        }
    };

    /**
     * Disable 2FA
     */
    const handleDisable2FA = async () => {
        const code = prompt('Enter your current 2FA code to disable:');
        if (!code) return;

        setLoading(true);
        setError(null);

        try {
            await api.post(API_CONFIG.TOTP_DISABLE, { code });

            setIs2FAEnabled(false);
            setStep('initial');
            setQrCodeUrl('');
            setSecret('');
            setVerificationCode('');
            setBackupCodes([]);
            setSuccess('2FA disabled successfully');
        } catch (err: any) {
            setError(err.data?.error || 'Failed to disable 2FA');
        } finally {
            setLoading(false);
        }
    };

    /**
     * Regenerate backup codes
     */
    const handleRegenerateBackupCodes = async () => {
        if (!confirm('This will invalidate your current backup codes. Continue?')) {
            return;
        }

        setLoading(true);
        setError(null);

        try {
            const response = await api.post<BackupCodesResponse>(API_CONFIG.TOTP_BACKUP_CODES);
            setBackupCodes(response.backupCodes);
            setSuccess('New backup codes generated successfully');
        } catch (err: any) {
            setError(err.data?.error || 'Failed to regenerate backup codes');
        } finally {
            setLoading(false);
        }
    };

    /**
     * Copy backup codes to clipboard
     */
    const handleCopyBackupCodes = () => {
        navigator.clipboard.writeText(backupCodes.join('\n'));
        setSuccess('Backup codes copied to clipboard');
        setTimeout(() => setSuccess(null), 3000);
    };

    /**
     * Download backup codes as text file
     */
    const handleDownloadBackupCodes = () => {
        const text = `RTX5 Trading Engine - 2FA Backup Codes\n\n${backupCodes.join('\n')}\n\nKeep these codes in a secure location.\nEach code can only be used once.`;
        const blob = new Blob([text], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'rtx5-2fa-backup-codes.txt';
        a.click();
        URL.revokeObjectURL(url);
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center gap-3">
                <Shield className={`w-6 h-6 ${is2FAEnabled ? 'text-green-400' : 'text-zinc-400'}`} />
                <div>
                    <h3 className="text-lg font-semibold">Two-Factor Authentication (2FA)</h3>
                    <p className="text-sm text-zinc-400">
                        {is2FAEnabled
                            ? 'Your account is protected with 2FA'
                            : 'Add an extra layer of security to your account'}
                    </p>
                </div>
            </div>

            {/* Status Messages */}
            {error && (
                <div className="flex items-start gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
                    <XCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
                    <p className="text-sm text-red-300">{error}</p>
                </div>
            )}

            {success && (
                <div className="flex items-start gap-2 p-3 bg-green-500/10 border border-green-500/30 rounded-lg">
                    <CheckCircle2 className="w-5 h-5 text-green-400 flex-shrink-0 mt-0.5" />
                    <p className="text-sm text-green-300">{success}</p>
                </div>
            )}

            {/* 2FA Disabled - Setup Flow */}
            {!is2FAEnabled && step === 'initial' && (
                <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
                    <p className="text-sm text-zinc-300 mb-4">
                        Enable two-factor authentication using Google Authenticator, Authy, or any TOTP-compatible app.
                    </p>
                    <button
                        onClick={handleSetupStart}
                        disabled={loading}
                        className="px-4 py-2 bg-emerald-500/20 border border-emerald-500/30 text-emerald-400 rounded-lg hover:bg-emerald-500/30 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                    >
                        {loading ? (
                            <>
                                <Loader2 className="w-4 h-4 animate-spin" />
                                Setting up...
                            </>
                        ) : (
                            <>
                                <Shield className="w-4 h-4" />
                                Enable 2FA
                            </>
                        )}
                    </button>
                </div>
            )}

            {/* STEP 1: QR Code Display */}
            {!is2FAEnabled && step === 'qr-display' && (
                <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6 space-y-4">
                    <div className="flex items-start gap-3">
                        <div className="w-8 h-8 bg-emerald-500/20 rounded-full flex items-center justify-center flex-shrink-0">
                            <span className="text-emerald-400 font-bold">1</span>
                        </div>
                        <div>
                            <h4 className="font-semibold mb-2">Scan QR Code</h4>
                            <p className="text-sm text-zinc-400 mb-4">
                                Open your authenticator app and scan this QR code:
                            </p>
                        </div>
                    </div>

                    {/* QR Code */}
                    <div className="flex justify-center p-6 bg-white rounded-lg">
                        {qrCodeUrl ? (
                            <img src={qrCodeUrl} alt="2FA QR Code" className="w-64 h-64" />
                        ) : (
                            <div className="w-64 h-64 flex items-center justify-center">
                                <Loader2 className="w-8 h-8 animate-spin text-zinc-400" />
                            </div>
                        )}
                    </div>

                    {/* Manual Entry */}
                    <div className="p-4 bg-zinc-800 rounded-lg">
                        <p className="text-xs text-zinc-400 mb-2">Or enter this code manually:</p>
                        <div className="flex items-center gap-2">
                            <code className="flex-1 text-sm font-mono text-emerald-400 break-all">{secret}</code>
                            <button
                                onClick={() => {
                                    navigator.clipboard.writeText(secret);
                                    setSuccess('Secret copied to clipboard');
                                    setTimeout(() => setSuccess(null), 2000);
                                }}
                                className="p-2 hover:bg-zinc-700 rounded transition-colors"
                                title="Copy secret"
                            >
                                <Copy className="w-4 h-4" />
                            </button>
                        </div>
                    </div>

                    {/* Verification Input */}
                    <div>
                        <label className="text-sm text-zinc-400 mb-2 block">Enter the 6-digit code from your app:</label>
                        <div className="flex gap-2">
                            <input
                                type="text"
                                value={verificationCode}
                                onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                                placeholder="000000"
                                maxLength={6}
                                className="flex-1 px-4 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-white text-center text-lg tracking-widest font-mono focus:outline-none focus:ring-2 focus:ring-emerald-500"
                            />
                            <button
                                onClick={handleVerifyCode}
                                disabled={loading || verificationCode.length !== 6}
                                className="px-6 py-2 bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : 'Verify'}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* STEP 2: Verification Success */}
            {!is2FAEnabled && step === 'verification' && (
                <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6 space-y-4">
                    <div className="flex items-start gap-3">
                        <CheckCircle2 className="w-6 h-6 text-green-400 flex-shrink-0" />
                        <div>
                            <h4 className="font-semibold mb-2">Verification Successful</h4>
                            <p className="text-sm text-zinc-400 mb-4">
                                Your code was verified successfully. Click below to complete setup and receive backup codes.
                            </p>
                        </div>
                    </div>

                    <button
                        onClick={handleEnable2FA}
                        disabled={loading}
                        className="w-full px-4 py-3 bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                    >
                        {loading ? (
                            <>
                                <Loader2 className="w-5 h-5 animate-spin" />
                                Enabling 2FA...
                            </>
                        ) : (
                            <>
                                <Key className="w-5 h-5" />
                                Complete Setup & Get Backup Codes
                            </>
                        )}
                    </button>
                </div>
            )}

            {/* STEP 3: Backup Codes Display */}
            {step === 'completed' && backupCodes.length > 0 && (
                <div className="bg-amber-500/10 border border-amber-500/30 rounded-xl p-6 space-y-4">
                    <div className="flex items-start gap-3">
                        <AlertTriangle className="w-6 h-6 text-amber-400 flex-shrink-0" />
                        <div>
                            <h4 className="font-semibold text-amber-300 mb-2">Save Your Backup Codes</h4>
                            <p className="text-sm text-amber-200/80">
                                Store these codes in a secure location. Each code can only be used once if you lose access to your authenticator app.
                            </p>
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-2 p-4 bg-zinc-900 rounded-lg font-mono text-sm">
                        {backupCodes.map((code, idx) => (
                            <div key={idx} className="text-emerald-400">
                                {code}
                            </div>
                        ))}
                    </div>

                    <div className="flex gap-2">
                        <button
                            onClick={handleCopyBackupCodes}
                            className="flex-1 px-4 py-2 bg-zinc-800 border border-zinc-700 text-white rounded-lg hover:bg-zinc-700 transition-colors flex items-center justify-center gap-2"
                        >
                            <Copy className="w-4 h-4" />
                            Copy Codes
                        </button>
                        <button
                            onClick={handleDownloadBackupCodes}
                            className="flex-1 px-4 py-2 bg-zinc-800 border border-zinc-700 text-white rounded-lg hover:bg-zinc-700 transition-colors flex items-center justify-center gap-2"
                        >
                            <Download className="w-4 h-4" />
                            Download
                        </button>
                    </div>
                </div>
            )}

            {/* 2FA Enabled - Management */}
            {is2FAEnabled && (
                <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6 space-y-4">
                    <div className="flex items-center gap-2 text-green-400 mb-4">
                        <CheckCircle2 className="w-5 h-5" />
                        <span className="font-semibold">Two-Factor Authentication is Enabled</span>
                    </div>

                    <div className="space-y-3">
                        <button
                            onClick={handleRegenerateBackupCodes}
                            disabled={loading}
                            className="w-full px-4 py-2 bg-zinc-800 border border-zinc-700 text-white rounded-lg hover:bg-zinc-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                        >
                            {loading ? (
                                <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                                <>
                                    <Key className="w-4 h-4" />
                                    Regenerate Backup Codes
                                </>
                            )}
                        </button>

                        <button
                            onClick={handleDisable2FA}
                            disabled={loading}
                            className="w-full px-4 py-2 bg-red-500/20 border border-red-500/30 text-red-400 rounded-lg hover:bg-red-500/30 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Disable 2FA
                        </button>
                    </div>

                    {/* Display backup codes if regenerated */}
                    {backupCodes.length > 0 && (
                        <div className="mt-4 p-4 bg-amber-500/10 border border-amber-500/30 rounded-lg">
                            <p className="text-xs text-amber-300 mb-3 flex items-center gap-2">
                                <AlertTriangle className="w-4 h-4" />
                                New backup codes generated. Previous codes are invalid.
                            </p>
                            <div className="grid grid-cols-2 gap-2 p-3 bg-zinc-900 rounded font-mono text-xs">
                                {backupCodes.map((code, idx) => (
                                    <div key={idx} className="text-emerald-400">{code}</div>
                                ))}
                            </div>
                            <div className="flex gap-2 mt-3">
                                <button
                                    onClick={handleCopyBackupCodes}
                                    className="flex-1 px-3 py-1.5 bg-zinc-800 border border-zinc-700 text-white text-sm rounded hover:bg-zinc-700 transition-colors flex items-center justify-center gap-2"
                                >
                                    <Copy className="w-3 h-3" />
                                    Copy
                                </button>
                                <button
                                    onClick={handleDownloadBackupCodes}
                                    className="flex-1 px-3 py-1.5 bg-zinc-800 border border-zinc-700 text-white text-sm rounded hover:bg-zinc-700 transition-colors flex items-center justify-center gap-2"
                                >
                                    <Download className="w-3 h-3" />
                                    Download
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
