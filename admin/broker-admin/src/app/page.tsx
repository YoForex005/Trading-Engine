'use client';

import { useState, useEffect } from 'react';
import {
  Users, Settings, Shield, Activity, ArrowUpDown, Wallet, RefreshCw, BookOpen, Layers
} from 'lucide-react';

import { Account, LedgerEntry, RoutingRule } from '../types';
import NavItem from '../components/ui/NavItem';
import Modal from '../components/ui/Modal';
import AccountsView from '../components/dashboard/AccountsView';
import LedgerView from '../components/dashboard/LedgerView';
import RoutingView from '../components/dashboard/RoutingView';
import LPStatusView from '../components/dashboard/LPStatusView';
import LPManagementView from '../components/dashboard/LPManagementView';
import RiskView from '../components/dashboard/RiskView';
import SettingsView from '../components/dashboard/SettingsView';
import SymbolsView from '../components/dashboard/SymbolsView';

export default function AdminDashboard() {
  const [activeTab, setActiveTab] = useState('accounts');
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [ledger, setLedger] = useState<LedgerEntry[]>([]);
  const [routes, setRoutes] = useState<RoutingRule[]>([]);
  const [lpStatus, setLpStatus] = useState<Record<string, any>>({});

  const [loading, setLoading] = useState(true);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [showModal, setShowModal] = useState<string | null>(null);

  const [executionMode, setExecutionMode] = useState('BBOOK');
  const [currentLP, setCurrentLP] = useState('OANDA');

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      // Fetch Accounts (Critical)
      try {
        const res = await fetch('http://localhost:8080/admin/accounts');
        if (res.ok) setAccounts(await res.json());
      } catch (e) {
        console.error('Failed to fetch accounts', e);
      }

      // Fetch Ledger (Non-Critical)
      try {
        const res = await fetch('http://localhost:8080/admin/ledger');
        if (res.ok) setLedger(await res.json());
      } catch (e) { console.error('Ledger fetch failed', e); }

      // Fetch Routes (Non-Critical)
      try {
        const res = await fetch('http://localhost:8080/admin/routes');
        if (res.ok) setRoutes(await res.json());
      } catch (e) { console.error('Routes fetch failed', e); }

      // Fetch LP Status (Non-Critical)
      try {
        const res = await fetch('http://localhost:8080/admin/lp-status');
        if (res.ok) setLpStatus(await res.json());
      } catch (e) { console.error('LP Status fetch failed', e); }

      // Fetch Config/Mode
      try {
        const res = await fetch('http://localhost:8080/api/config');
        if (res.ok) {
          const data = await res.json();
          setExecutionMode(data.executionMode);
          setCurrentLP(data.priceFeedLP);
        }
      } catch (e) { console.error('Config fetch failed', e); }

    } catch (error) {
      console.error('Global fetch error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleModeChange = async (mode: string) => {
    try {
      const res = await fetch('http://localhost:8080/api/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ executionMode: mode })
      });
      if (res.ok) {
        setExecutionMode(mode);
        alert('Execution mode updated to ' + mode);
        fetchData();
      }
    } catch (e) {
      alert('Failed to update mode');
    }
  };

  const handleLPChange = async (lp: string) => {
    try {
      const res = await fetch('http://localhost:8080/api/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ priceFeedLP: lp })
      });
      if (res.ok) {
        setCurrentLP(lp);
        alert(`LP changed to ${lp}. Please RESTART the backend to connect.`);
        fetchData();
      }
    } catch (e) {
      alert('Failed to update LP');
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
          <NavItem icon={<Layers size={18} />} label="LP Management" active={activeTab === 'lpmanage'} onClick={() => setActiveTab('lpmanage')} />
          <NavItem icon={<BookOpen size={18} />} label="Feeds" active={activeTab === 'symbols'} onClick={() => setActiveTab('symbols')} />
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
          {activeTab === 'lpmanage' && <LPManagementView />}
          {activeTab === 'symbols' && <SymbolsView />}
          {activeTab === 'risk' && <RiskView accounts={accounts} />}
          {activeTab === 'settings' && <SettingsView mode={executionMode} onModeChange={handleModeChange} lp={currentLP} onLPChange={handleLPChange} />}
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
