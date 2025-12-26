'use client';

import { useState, useEffect } from 'react';
import {
  Users, Settings, BarChart3, Shield, Activity, ArrowUpDown, Wallet,
  AlertTriangle, CheckCircle, XCircle, RefreshCw, Plus, DollarSign,
  ArrowDownCircle, ArrowUpCircle, Edit2, Eye, Search, Filter
} from 'lucide-react';

interface Account {
  id: number;
  accountNumber: string;
  userId: string;
  balance: number;
  equity?: number;
  margin?: number;
  freeMargin?: number;
  marginLevel?: number;
  leverage: number;
  marginMode: string;
  currency: string;
  status: string;
  isDemo: boolean;
}

interface LedgerEntry {
  id: number;
  accountId: number;
  type: string;
  amount: number;
  balanceAfter: number;
  description: string;
  paymentMethod: string;
  paymentRef: string;
  adminId: string;
  createdAt: string;
}

interface RoutingRule {
  id: string;
  groupPattern: string;
  symbolPattern: string;
  minVolume: number;
  maxVolume: number;
  action: string;
  targetLp: string;
  priority: number;
}

export default function AdminDashboard() {
  const [activeTab, setActiveTab] = useState('accounts');
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [ledger, setLedger] = useState<LedgerEntry[]>([]);
  const [routes, setRoutes] = useState<RoutingRule[]>([]);
  const [lpStatus, setLpStatus] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [showModal, setShowModal] = useState<string | null>(null);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      const [accountsRes, ledgerRes, routesRes, lpRes] = await Promise.all([
        fetch('http://localhost:8080/admin/accounts'),
        fetch('http://localhost:8080/admin/ledger'),
        fetch('http://localhost:8080/admin/routes'),
        fetch('http://localhost:8080/admin/lp-status'),
      ]);

      if (accountsRes.ok) setAccounts(await accountsRes.json());
      if (ledgerRes.ok) setLedger(await ledgerRes.json());
      if (routesRes.ok) setRoutes(await routesRes.json());
      if (lpRes.ok) setLpStatus(await lpRes.json());
    } catch (error) {
      console.error('Failed to fetch data:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex h-screen bg-zinc-950 text-zinc-200">
      {/* Sidebar */}
      <aside className="w-64 border-r border-zinc-800 flex flex-col bg-zinc-900/50">
        <div className="p-4 border-b border-zinc-800">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-lg flex items-center justify-center font-bold text-white">
              RTX
            </div>
            <div>
              <h1 className="font-semibold">Broker Admin</h1>
              <p className="text-xs text-zinc-500">B-Book Management</p>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          <NavItem icon={<Users size={18} />} label="Accounts" active={activeTab === 'accounts'} onClick={() => setActiveTab('accounts')} />
          <NavItem icon={<Wallet size={18} />} label="Ledger" active={activeTab === 'ledger'} onClick={() => setActiveTab('ledger')} />
          <NavItem icon={<ArrowUpDown size={18} />} label="Routing Rules" active={activeTab === 'routing'} onClick={() => setActiveTab('routing')} />
          <NavItem icon={<Activity size={18} />} label="LP Status" active={activeTab === 'lp'} onClick={() => setActiveTab('lp')} />
          <NavItem icon={<Shield size={18} />} label="Risk Monitor" active={activeTab === 'risk'} onClick={() => setActiveTab('risk')} />
          <NavItem icon={<Settings size={18} />} label="Settings" active={activeTab === 'settings'} onClick={() => setActiveTab('settings')} />
        </nav>

        <div className="p-4 border-t border-zinc-800">
          <div className="flex items-center gap-2 text-xs text-zinc-500">
            <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            <span>B-Book Engine Online</span>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        <header className="h-14 border-b border-zinc-800 flex items-center justify-between px-6 bg-zinc-900/30">
          <h2 className="text-lg font-semibold capitalize">{activeTab}</h2>
          <button
            onClick={fetchData}
            className="p-2 hover:bg-zinc-800 rounded-lg transition-colors"
          >
            <RefreshCw size={18} className={loading ? 'animate-spin' : ''} />
          </button>
        </header>

        <div className="p-6">
          {activeTab === 'accounts' && (
            <AccountsView
              accounts={accounts}
              onSelect={setSelectedAccount}
              onAction={(action) => setShowModal(action)}
              selected={selectedAccount}
              onRefresh={fetchData}
            />
          )}
          {activeTab === 'ledger' && <LedgerView ledger={ledger} />}
          {activeTab === 'routing' && <RoutingView routes={routes} />}
          {activeTab === 'lp' && <LPStatusView status={lpStatus} />}
          {activeTab === 'risk' && <RiskView accounts={accounts} />}
          {activeTab === 'settings' && <SettingsView />}
        </div>
      </main>

      {/* Modals */}
      {showModal && selectedAccount && (
        <Modal
          type={showModal}
          account={selectedAccount}
          onClose={() => setShowModal(null)}
          onSuccess={() => { setShowModal(null); fetchData(); }}
        />
      )}
    </div>
  );
}

function NavItem({ icon, label, active, onClick }: { icon: React.ReactNode; label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${active ? 'bg-emerald-600/20 text-emerald-400' : 'hover:bg-zinc-800 text-zinc-400'
        }`}
    >
      {icon}
      {label}
    </button>
  );
}

// ===== ACCOUNTS VIEW =====
function AccountsView({ accounts, onSelect, onAction, selected, onRefresh }: {
  accounts: Account[],
  onSelect: (a: Account) => void,
  onAction: (a: string) => void,
  selected: Account | null,
  onRefresh: () => void
}) {
  const [search, setSearch] = useState('');

  const filtered = accounts.filter(a =>
    a.accountNumber?.toLowerCase().includes(search.toLowerCase()) ||
    a.userId?.toLowerCase().includes(search.toLowerCase())
  );

  const handleCreateAccount = async () => {
    try {
      await fetch('http://localhost:8080/api/account/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ userId: `user-${Date.now()}`, isDemo: true })
      });
      onRefresh();
    } catch (err) {
      console.error('Failed to create account:', err);
    }
  };

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
            <input
              type="text"
              placeholder="Search accounts..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9 pr-4 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm w-64"
            />
          </div>
        </div>
        <button
          onClick={handleCreateAccount}
          className="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 rounded-lg text-sm transition-colors"
        >
          <Plus size={16} />
          Create Account
        </button>
      </div>

      {/* Accounts Table */}
      <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden">
        <table className="w-full">
          <thead className="bg-zinc-800/50">
            <tr className="text-left text-xs text-zinc-500 uppercase">
              <th className="p-3">Account</th>
              <th className="p-3">Balance</th>
              <th className="p-3">Leverage</th>
              <th className="p-3">Mode</th>
              <th className="p-3">Status</th>
              <th className="p-3">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map(account => (
              <tr
                key={account.id}
                className={`border-t border-zinc-800 hover:bg-zinc-800/30 cursor-pointer ${selected?.id === account.id ? 'bg-zinc-800/50' : ''}`}
                onClick={() => onSelect(account)}
              >
                <td className="p-3">
                  <div>
                    <div className="font-medium">{account.accountNumber}</div>
                    <div className="text-xs text-zinc-500">{account.userId}</div>
                  </div>
                </td>
                <td className="p-3">
                  <span className="font-mono text-emerald-400">${account.balance.toLocaleString()}</span>
                </td>
                <td className="p-3">1:{account.leverage}</td>
                <td className="p-3">
                  <span className={`px-2 py-0.5 rounded text-xs ${account.marginMode === 'HEDGING' ? 'bg-blue-500/20 text-blue-400' : 'bg-purple-500/20 text-purple-400'}`}>
                    {account.marginMode}
                  </span>
                </td>
                <td className="p-3">
                  <span className={`px-2 py-0.5 rounded text-xs ${account.status === 'ACTIVE' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-400'}`}>
                    {account.status}
                  </span>
                </td>
                <td className="p-3">
                  <div className="flex items-center gap-1">
                    <button
                      onClick={(e) => { e.stopPropagation(); onSelect(account); onAction('deposit'); }}
                      className="p-1.5 hover:bg-emerald-500/20 rounded text-emerald-400" title="Deposit"
                    >
                      <ArrowDownCircle size={16} />
                    </button>
                    <button
                      onClick={(e) => { e.stopPropagation(); onSelect(account); onAction('withdraw'); }}
                      className="p-1.5 hover:bg-red-500/20 rounded text-red-400" title="Withdraw"
                    >
                      <ArrowUpCircle size={16} />
                    </button>
                    <button
                      onClick={(e) => { e.stopPropagation(); onSelect(account); onAction('edit'); }}
                      className="p-1.5 hover:bg-zinc-700 rounded text-zinc-400" title="Edit"
                    >
                      <Edit2 size={16} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ===== LEDGER VIEW =====
function LedgerView({ ledger }: { ledger: LedgerEntry[] }) {
  return (
    <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden">
      <div className="p-4 border-b border-zinc-800">
        <h3 className="font-semibold">Transaction Ledger</h3>
      </div>
      <table className="w-full">
        <thead className="bg-zinc-800/50">
          <tr className="text-left text-xs text-zinc-500 uppercase">
            <th className="p-3">ID</th>
            <th className="p-3">Account</th>
            <th className="p-3">Type</th>
            <th className="p-3">Amount</th>
            <th className="p-3">Balance After</th>
            <th className="p-3">Method</th>
            <th className="p-3">Reference</th>
            <th className="p-3">Date</th>
          </tr>
        </thead>
        <tbody>
          {ledger.map(entry => (
            <tr key={entry.id} className="border-t border-zinc-800 hover:bg-zinc-800/30">
              <td className="p-3 font-mono text-xs">{entry.id}</td>
              <td className="p-3">#{entry.accountId}</td>
              <td className="p-3">
                <span className={`px-2 py-0.5 rounded text-xs ${entry.type === 'DEPOSIT' ? 'bg-emerald-500/20 text-emerald-400' :
                    entry.type === 'WITHDRAW' ? 'bg-red-500/20 text-red-400' :
                      entry.type === 'REALIZED_PNL' ? 'bg-blue-500/20 text-blue-400' :
                        'bg-zinc-700 text-zinc-400'
                  }`}>
                  {entry.type}
                </span>
              </td>
              <td className={`p-3 font-mono ${entry.amount >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                {entry.amount >= 0 ? '+' : ''}{entry.amount.toFixed(2)}
              </td>
              <td className="p-3 font-mono">${entry.balanceAfter.toFixed(2)}</td>
              <td className="p-3 text-xs text-zinc-400">{entry.paymentMethod || '-'}</td>
              <td className="p-3 text-xs text-zinc-500 max-w-32 truncate" title={entry.paymentRef}>{entry.paymentRef || '-'}</td>
              <td className="p-3 text-xs text-zinc-500">{new Date(entry.createdAt).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ===== MODAL =====
function Modal({ type, account, onClose, onSuccess }: { type: string, account: Account, onClose: () => void, onSuccess: () => void }) {
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
        // TODO: Implement leverage update
        endpoint = '/admin/adjust';
        body = { ...body, amount: 0, description: `Leverage changed to 1:${leverage}` };
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

// ===== ROUTING VIEW =====
function RoutingView({ routes }: { routes: RoutingRule[] }) {
  return (
    <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden">
      <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
        <h3 className="font-semibold">Order Routing Rules (A-Book vs B-Book)</h3>
        <button className="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 rounded-lg text-sm transition-colors">
          + Add Rule
        </button>
      </div>
      <table className="w-full">
        <thead className="bg-zinc-800/50">
          <tr className="text-left text-xs text-zinc-500 uppercase">
            <th className="p-3">Group</th>
            <th className="p-3">Symbol</th>
            <th className="p-3">Volume Range</th>
            <th className="p-3">Action</th>
            <th className="p-3">Target LP</th>
            <th className="p-3">Priority</th>
          </tr>
        </thead>
        <tbody>
          {routes.length === 0 ? (
            <tr>
              <td colSpan={6} className="p-8 text-center text-zinc-500">
                No routing rules. All orders will be B-Booked.
              </td>
            </tr>
          ) : (
            routes.map(rule => (
              <tr key={rule.id} className="border-t border-zinc-800 hover:bg-zinc-800/30">
                <td className="p-3">{rule.groupPattern}</td>
                <td className="p-3">{rule.symbolPattern}</td>
                <td className="p-3">{rule.minVolume} - {rule.maxVolume}</td>
                <td className="p-3">
                  <span className={`px-2 py-0.5 rounded text-xs ${rule.action === 'A_BOOK' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-orange-500/20 text-orange-400'}`}>
                    {rule.action}
                  </span>
                </td>
                <td className="p-3 text-zinc-400">{rule.targetLp || '-'}</td>
                <td className="p-3">{rule.priority}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}

// ===== LP STATUS VIEW =====
function LPStatusView({ status }: { status: Record<string, string> }) {
  return (
    <div className="space-y-4">
      {Object.entries(status).length === 0 ? (
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-8 text-center">
          <Activity className="w-12 h-12 mx-auto text-zinc-600 mb-4" />
          <h3 className="font-semibold mb-2">No LP Connections</h3>
          <p className="text-sm text-zinc-500">Running in pure B-Book mode (internal execution only)</p>
        </div>
      ) : (
        Object.entries(status).map(([id, s]) => (
          <div key={id} className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                {s === 'LOGGED_IN' ? <CheckCircle className="text-emerald-500" /> : s === 'CONNECTING' ? <RefreshCw className="text-yellow-500 animate-spin" /> : <XCircle className="text-red-500" />}
                <h3 className="font-semibold">{id}</h3>
              </div>
              <span className={`px-3 py-1 rounded-full text-xs ${s === 'LOGGED_IN' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-zinc-700 text-zinc-400'}`}>
                {s}
              </span>
            </div>
          </div>
        ))
      )}
    </div>
  );
}

// ===== RISK VIEW =====
function RiskView({ accounts }: { accounts: Account[] }) {
  const totalBalance = accounts.reduce((sum, a) => sum + a.balance, 0);

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-3 gap-4">
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
          <h4 className="text-sm text-zinc-500 mb-2">Total Accounts</h4>
          <p className="text-2xl font-bold">{accounts.length}</p>
        </div>
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
          <h4 className="text-sm text-zinc-500 mb-2">Total Balance</h4>
          <p className="text-2xl font-bold text-emerald-400">${totalBalance.toLocaleString()}</p>
        </div>
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
          <h4 className="text-sm text-zinc-500 mb-2">Execution Mode</h4>
          <p className="text-2xl font-bold text-orange-400">B-Book</p>
        </div>
      </div>
    </div>
  );
}

// ===== SETTINGS VIEW =====
function SettingsView() {
  return (
    <div className="space-y-6">
      <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
        <h3 className="font-semibold mb-4">Execution Mode</h3>
        <div className="flex gap-4">
          <label className="flex items-center gap-2 p-4 bg-orange-500/10 border border-orange-500/30 rounded-lg cursor-pointer">
            <input type="radio" name="mode" value="bbook" defaultChecked className="accent-orange-500" />
            <div>
              <div className="font-medium text-orange-400">B-Book (Internal)</div>
              <div className="text-xs text-zinc-500">All orders executed internally</div>
            </div>
          </label>
          <label className="flex items-center gap-2 p-4 bg-zinc-800 border border-zinc-700 rounded-lg cursor-pointer">
            <input type="radio" name="mode" value="abook" className="accent-emerald-500" />
            <div>
              <div className="font-medium">A-Book (LP)</div>
              <div className="text-xs text-zinc-500">Orders routed to liquidity providers</div>
            </div>
          </label>
        </div>
      </div>

      <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
        <h3 className="font-semibold mb-4">Default Account Settings</h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-sm text-zinc-400 mb-1 block">Default Leverage</label>
            <select className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg">
              <option>1:100</option>
              <option>1:200</option>
              <option>1:500</option>
            </select>
          </div>
          <div>
            <label className="text-sm text-zinc-400 mb-1 block">Margin Mode</label>
            <select className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg">
              <option>HEDGING</option>
              <option>NETTING</option>
            </select>
          </div>
        </div>
      </div>
    </div>
  );
}
