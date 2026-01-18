import React, { useState } from 'react';
import { ArrowDownCircle, ArrowUpCircle, Edit2, Plus, Lock } from 'lucide-react';
import { Account } from '../../types';

interface ModalProps {
    type: string;
    account: Account;
    onClose: () => void;
    onSuccess: () => void;
}

export default function Modal({ type, account, onClose, onSuccess }: ModalProps) {
    const [amount, setAmount] = useState('');
    const [method, setMethod] = useState('MANUAL');
    const [reference, setReference] = useState('');
    const [description, setDescription] = useState('');
    const [leverage, setLeverage] = useState(account.leverage.toString());
    const [loading, setLoading] = useState(false);

    const handleSubmit = async () => {
        setLoading(true);
        try {
            let endpoint = '';
            let body: any = { accountId: account.id, adminId: 'broker-admin' };

            if (type === 'deposit') {
                endpoint = '/admin/deposit';
                body = { ...body, amount: parseFloat(amount), method, reference, description };
            } else if (type === 'withdraw') {
                endpoint = '/admin/withdraw';
                body = { ...body, amount: parseFloat(amount), method, reference, description };
            } else if (type === 'edit') {
                endpoint = '/admin/account/update';
                body = {
                    accountId: account.id,
                    leverage: parseFloat(leverage),
                    marginMode: account.marginMode // Keep existing mode for now
                };
            } else if (type === 'create') {
                endpoint = '/api/account/create';
                // Reuse description as username and reference as password for simplicity
                body = { userId: `user-${Date.now()}`, username: description, password: reference, isDemo: true };
            } else if (type === 'reset-password') {
                endpoint = '/admin/reset-password';
                body = { accountId: account.id, newPassword: reference };
            }

            const res = await fetch(`http://localhost:8080${endpoint}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (res.ok) {
                onSuccess();
            } else {
                const err = await res.text();
                alert('Error: ' + err);
            }
        } catch (err) {
            console.error(err);
            alert('Request failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
            <div className="bg-zinc-900 border border-zinc-700 rounded-xl w-full max-w-md p-6">
                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    {type === 'deposit' && <><ArrowDownCircle className="text-emerald-400" /> Deposit Funds</>}
                    {type === 'withdraw' && <><ArrowUpCircle className="text-red-400" /> Withdraw Funds</>}
                    {type === 'edit' && <><Edit2 className="text-blue-400" /> Edit Account</>}
                    {type === 'create' && <><Plus className="text-emerald-400" /> Create New Account</>}
                    {type === 'reset-password' && <><Lock className="text-yellow-400" /> Reset Password</>}
                </h3>

                <div className="mb-4 p-3 bg-zinc-800 rounded-lg">
                    <div className="text-sm text-zinc-400">Account</div>
                    <div className="font-semibold">{account.accountNumber}</div>
                    <div className="text-sm text-zinc-500">Balance: ${account.balance.toLocaleString()}</div>
                </div>

                {(type === 'deposit' || type === 'withdraw') && (
                    <div className="space-y-4">
                        <div>
                            <label className="text-sm text-zinc-400 mb-1 block">Amount (USD)</label>
                            <input
                                type="number"
                                value={amount}
                                onChange={(e) => setAmount(e.target.value)}
                                placeholder="0.00"
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                            />
                        </div>
                        <div>
                            <label className="text-sm text-zinc-400 mb-1 block">Payment Method</label>
                            <select
                                value={method}
                                onChange={(e) => setMethod(e.target.value)}
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                            >
                                <option value="MANUAL">Manual</option>
                                <option value="BANK">Bank Transfer</option>
                                <option value="CRYPTO">Crypto</option>
                                <option value="CARD">Card</option>
                            </select>
                        </div>
                        <div>
                            <label className="text-sm text-zinc-400 mb-1 block">Reference (optional)</label>
                            <input
                                type="text"
                                value={reference}
                                onChange={(e) => setReference(e.target.value)}
                                placeholder="Transaction ID, crypto hash, etc."
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                            />
                        </div>
                        <div>
                            <label className="text-sm text-zinc-400 mb-1 block">Description</label>
                            <input
                                type="text"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="Notes..."
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                            />
                        </div>
                    </div>
                )}

                {/* New Account Fields */}
                {type === 'create' && (
                    <div className="space-y-4">
                        <div>
                            <label className="text-sm text-zinc-400 mb-1 block">Username (Optional)</label>
                            <input
                                type="text"
                                value={description} // Reuse description as username state for simplicity or add new state
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="Client Name or ID"
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                            />
                        </div>
                        <div>
                            <label className="text-sm text-zinc-400 mb-1 block">Password</label>
                            <input
                                type="password"
                                value={reference} // Reuse reference as password state
                                onChange={(e) => setReference(e.target.value)}
                                placeholder="Secret Password"
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                            />
                        </div>
                    </div>
                )}

                {type === 'edit' && (
                    <div className="space-y-4">
                        <div>
                            <label className="text-sm text-zinc-400 mb-1 block">Leverage</label>
                            <select
                                value={leverage}
                                onChange={(e) => setLeverage(e.target.value)}
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                            >
                                <option value="50">1:50</option>
                                <option value="100">1:100</option>
                                <option value="200">1:200</option>
                                <option value="500">1:500</option>
                            </select>
                        </div>
                    </div>
                )}

                {type === 'reset-password' && (
                    <div className="space-y-4">
                        <div>
                            <label className="text-sm text-zinc-400 mb-1 block">New Password</label>
                            <input
                                type="password"
                                value={reference} // Reuse reference as password state
                                onChange={(e) => setReference(e.target.value)}
                                placeholder="New Secret Password"
                                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg"
                            />
                        </div>
                    </div>
                )}

                <div className="flex gap-2 mt-6">
                    <button
                        onClick={onClose}
                        className="flex-1 px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded-lg transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSubmit}
                        disabled={loading || ((type === 'deposit' || type === 'withdraw') && !amount)}
                        className={`flex-1 px-4 py-2 rounded-lg transition-colors ${type === 'deposit' ? 'bg-emerald-600 hover:bg-emerald-500' :
                            type === 'withdraw' ? 'bg-red-600 hover:bg-red-500' :
                                'bg-blue-600 hover:bg-blue-500'
                            } disabled:opacity-50`}
                    >
                        {loading ? 'Processing...' : 'Confirm'}
                    </button>
                </div>
            </div>
        </div>
    );
}
