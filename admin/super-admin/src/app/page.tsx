'use client';

import { useState, useEffect } from 'react';
import {
  Building2, CloudCog, Database, Users, Shield, Activity,
  Server, Globe, Zap, Plus, Trash2, Edit, Power
} from 'lucide-react';

interface Broker {
  id: string;
  name: string;
  status: 'active' | 'suspended' | 'pending';
  users: number;
  volume: number;
  createdAt: string;
}

interface LPConfig {
  id: string;
  name: string;
  type: 'FIX' | 'API' | 'Bridge';
  host: string;
  status: string;
  enabled: boolean;
}

export default function SuperAdminDashboard() {
  const [activeTab, setActiveTab] = useState('overview');
  const [brokers, setBrokers] = useState<Broker[]>([
    { id: 'broker_001', name: 'Alpha Trading Ltd', status: 'active', users: 1250, volume: 45600000, createdAt: '2024-01-15' },
    { id: 'broker_002', name: 'Beta Markets', status: 'active', users: 890, volume: 23400000, createdAt: '2024-03-20' },
    { id: 'broker_003', name: 'Gamma Forex', status: 'pending', users: 0, volume: 0, createdAt: '2024-12-20' },
  ]);
  const [lps, setLPs] = useState<LPConfig[]>([
    { id: 'lmax_prod', name: 'LMAX Exchange', type: 'FIX', host: 'fix.lmax.com:443', status: 'LOGGED_IN', enabled: true },
    { id: 'mass_markets', name: 'Mass Markets', type: 'FIX', host: 'fix.massmarkets.com:443', status: 'DISCONNECTED', enabled: false },
    { id: 'currenex', name: 'Currenex', type: 'FIX', host: 'fix.currenex.com:443', status: 'DISCONNECTED', enabled: false },
  ]);

  return (
    <div className="flex h-screen bg-zinc-950 text-zinc-200">
      {/* Sidebar */}
      <aside className="w-72 border-r border-zinc-800 flex flex-col bg-gradient-to-b from-violet-950/20 to-zinc-950">
        <div className="p-4 border-b border-zinc-800">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-violet-600 to-purple-600 rounded-lg flex items-center justify-center font-bold text-white shadow-lg shadow-violet-500/20">
              RTX
            </div>
            <div>
              <h1 className="font-semibold">Super Admin</h1>
              <p className="text-xs text-zinc-500">Platform Control Center</p>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          <div className="text-xs text-zinc-600 uppercase tracking-wider mb-2 px-3">Platform</div>
          <NavItem icon={<Activity size={18} />} label="Overview" active={activeTab === 'overview'} onClick={() => setActiveTab('overview')} />
          <NavItem icon={<Building2 size={18} />} label="White Labels" active={activeTab === 'whitelabels'} onClick={() => setActiveTab('whitelabels')} />

          <div className="text-xs text-zinc-600 uppercase tracking-wider mb-2 px-3 pt-4">Infrastructure</div>
          <NavItem icon={<CloudCog size={18} />} label="LP Configuration" active={activeTab === 'lps'} onClick={() => setActiveTab('lps')} />
          <NavItem icon={<Database size={18} />} label="Databases" active={activeTab === 'databases'} onClick={() => setActiveTab('databases')} />
          <NavItem icon={<Server size={18} />} label="Servers" active={activeTab === 'servers'} onClick={() => setActiveTab('servers')} />

          <div className="text-xs text-zinc-600 uppercase tracking-wider mb-2 px-3 pt-4">Security</div>
          <NavItem icon={<Shield size={18} />} label="Kill Switches" active={activeTab === 'killswitch'} onClick={() => setActiveTab('killswitch')} />
          <NavItem icon={<Users size={18} />} label="Admins" active={activeTab === 'admins'} onClick={() => setActiveTab('admins')} />
        </nav>

        <div className="p-4 border-t border-zinc-800 bg-violet-950/20">
          <div className="flex items-center gap-2 text-xs">
            <Zap className="w-4 h-4 text-violet-400" />
            <span className="text-zinc-400">RTX Engine v1.0.0</span>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        <header className="h-14 border-b border-zinc-800 flex items-center justify-between px-6 bg-zinc-900/30">
          <h2 className="text-lg font-semibold capitalize">{activeTab === 'whitelabels' ? 'White Label Brokers' : activeTab === 'lps' ? 'Liquidity Providers' : activeTab}</h2>
          <div className="flex items-center gap-2 text-xs">
            <Globe className="w-4 h-4 text-emerald-500" />
            <span>3 Regions Active</span>
          </div>
        </header>

        <div className="p-6">
          {activeTab === 'overview' && <OverviewView brokers={brokers} lps={lps} />}
          {activeTab === 'whitelabels' && <WhiteLabelsView brokers={brokers} />}
          {activeTab === 'lps' && <LPConfigView lps={lps} setLPs={setLPs} />}
          {activeTab === 'killswitch' && <KillSwitchView />}
          {activeTab === 'databases' && <PlaceholderView title="Database Management" icon={<Database />} />}
          {activeTab === 'servers' && <PlaceholderView title="Server Cluster Management" icon={<Server />} />}
          {activeTab === 'admins' && <PlaceholderView title="Admin Users" icon={<Users />} />}
        </div>
      </main>
    </div>
  );
}

function NavItem({ icon, label, active, onClick }: { icon: React.ReactNode; label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${active ? 'bg-violet-600/20 text-violet-400' : 'hover:bg-zinc-800 text-zinc-400'
        }`}
    >
      {icon}
      {label}
    </button>
  );
}

function OverviewView({ brokers, lps }: { brokers: any[]; lps: any[] }) {
  const totalUsers = brokers.reduce((sum, b) => sum + b.users, 0);
  const totalVolume = brokers.reduce((sum, b) => sum + b.volume, 0);
  const activeLPs = lps.filter(lp => lp.status === 'LOGGED_IN').length;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-4 gap-4">
        <StatCard label="Active Brokers" value={brokers.filter(b => b.status === 'active').length.toString()} color="violet" />
        <StatCard label="Total Traders" value={totalUsers.toLocaleString()} color="blue" />
        <StatCard label="Monthly Volume" value={`$${(totalVolume / 1000000).toFixed(1)}M`} color="emerald" />
        <StatCard label="Connected LPs" value={`${activeLPs}/${lps.length}`} color="orange" />
      </div>

      <div className="grid grid-cols-2 gap-6">
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
          <h3 className="font-semibold mb-4">Top Brokers by Volume</h3>
          {brokers.filter(b => b.status === 'active').sort((a, b) => b.volume - a.volume).slice(0, 5).map(broker => (
            <div key={broker.id} className="flex items-center justify-between py-2 border-b border-zinc-800 last:border-0">
              <span>{broker.name}</span>
              <span className="text-zinc-400">${(broker.volume / 1000000).toFixed(2)}M</span>
            </div>
          ))}
        </div>
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
          <h3 className="font-semibold mb-4">System Health</h3>
          <div className="space-y-3">
            <HealthItem label="Trading Core" status="healthy" />
            <HealthItem label="FIX Gateway" status="healthy" />
            <HealthItem label="Market Data" status="healthy" />
            <HealthItem label="Database Cluster" status="healthy" />
          </div>
        </div>
      </div>
    </div>
  );
}

function HealthItem({ label, status }: { label: string; status: 'healthy' | 'warning' | 'error' }) {
  const colors = {
    healthy: 'bg-emerald-500',
    warning: 'bg-yellow-500',
    error: 'bg-red-500'
  };
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm">{label}</span>
      <div className="flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full ${colors[status]}`} />
        <span className="text-xs text-zinc-500 capitalize">{status}</span>
      </div>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: string; color: string }) {
  const colorClasses: Record<string, string> = {
    violet: 'bg-violet-500/10 text-violet-400 border-violet-500/20',
    blue: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
    emerald: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    orange: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  };

  return (
    <div className={`p-4 rounded-xl border ${colorClasses[color]}`}>
      <div className="text-sm opacity-75 mb-1">{label}</div>
      <div className="text-2xl font-bold">{value}</div>
    </div>
  );
}

function WhiteLabelsView({ brokers }: { brokers: any[] }) {
  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <button className="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-500 rounded-lg transition-colors">
          <Plus size={16} />
          Add White Label
        </button>
      </div>

      <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl overflow-hidden">
        <table className="w-full">
          <thead className="bg-zinc-800/50">
            <tr className="text-left text-xs text-zinc-500 uppercase">
              <th className="p-4">Broker</th>
              <th className="p-4">Status</th>
              <th className="p-4">Users</th>
              <th className="p-4">Monthly Volume</th>
              <th className="p-4">Created</th>
              <th className="p-4">Actions</th>
            </tr>
          </thead>
          <tbody>
            {brokers.map(broker => (
              <tr key={broker.id} className="border-t border-zinc-800 hover:bg-zinc-800/30">
                <td className="p-4 font-medium">{broker.name}</td>
                <td className="p-4">
                  <span className={`px-2 py-0.5 rounded text-xs ${broker.status === 'active' ? 'bg-emerald-500/20 text-emerald-400' :
                      broker.status === 'pending' ? 'bg-yellow-500/20 text-yellow-400' :
                        'bg-red-500/20 text-red-400'
                    }`}>
                    {broker.status}
                  </span>
                </td>
                <td className="p-4">{broker.users.toLocaleString()}</td>
                <td className="p-4">${(broker.volume / 1000000).toFixed(2)}M</td>
                <td className="p-4 text-zinc-500">{broker.createdAt}</td>
                <td className="p-4">
                  <div className="flex gap-2">
                    <button className="p-1.5 hover:bg-zinc-700 rounded"><Edit size={14} /></button>
                    <button className="p-1.5 hover:bg-red-500/20 rounded text-red-400"><Trash2 size={14} /></button>
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

function LPConfigView({ lps, setLPs }: { lps: any[]; setLPs: (lps: any[]) => void }) {
  const toggleLP = (id: string) => {
    setLPs(lps.map(lp => lp.id === id ? { ...lp, enabled: !lp.enabled, status: !lp.enabled ? 'CONNECTING' : 'DISCONNECTED' } : lp));
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <button className="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-500 rounded-lg transition-colors">
          <Plus size={16} />
          Add LP Connection
        </button>
      </div>

      <div className="grid gap-4">
        {lps.map(lp => (
          <div key={lp.id} className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className={`w-3 h-3 rounded-full ${lp.status === 'LOGGED_IN' ? 'bg-emerald-500' : lp.status === 'CONNECTING' ? 'bg-yellow-500 animate-pulse' : 'bg-zinc-600'}`} />
                <div>
                  <h3 className="font-semibold">{lp.name}</h3>
                  <p className="text-xs text-zinc-500">{lp.type} • {lp.host}</p>
                </div>
              </div>
              <div className="flex items-center gap-4">
                <span className={`px-2 py-1 rounded text-xs ${lp.status === 'LOGGED_IN' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-zinc-700 text-zinc-400'}`}>
                  {lp.status}
                </span>
                <button
                  onClick={() => toggleLP(lp.id)}
                  className={`p-2 rounded-lg transition-colors ${lp.enabled ? 'bg-emerald-500/20 text-emerald-400' : 'bg-zinc-800 text-zinc-500'}`}
                >
                  <Power size={18} />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function KillSwitchView() {
  return (
    <div className="space-y-6">
      <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-6">
        <h3 className="font-semibold text-red-400 mb-2 flex items-center gap-2">
          <Shield className="w-5 h-5" />
          Emergency Controls
        </h3>
        <p className="text-sm text-zinc-400 mb-4">Use these controls to immediately halt trading operations. Use with caution.</p>

        <div className="grid grid-cols-3 gap-4">
          <button className="p-4 bg-red-600/20 hover:bg-red-600/30 border border-red-500/30 rounded-xl text-center transition-colors">
            <Power className="w-8 h-8 mx-auto mb-2 text-red-400" />
            <p className="font-medium">Halt All Trading</p>
          </button>
          <button className="p-4 bg-orange-600/20 hover:bg-orange-600/30 border border-orange-500/30 rounded-xl text-center transition-colors">
            <CloudCog className="w-8 h-8 mx-auto mb-2 text-orange-400" />
            <p className="font-medium">Disconnect All LPs</p>
          </button>
          <button className="p-4 bg-yellow-600/20 hover:bg-yellow-600/30 border border-yellow-500/30 rounded-xl text-center transition-colors">
            <Users className="w-8 h-8 mx-auto mb-2 text-yellow-400" />
            <p className="font-medium">Freeze New Orders</p>
          </button>
        </div>
      </div>

      <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4">
        <h3 className="font-semibold mb-4">Symbol-Specific Controls</h3>
        <div className="space-y-2">
          {['EURUSD', 'GBPUSD', 'USDJPY', 'BTCUSD', 'XAUUSD'].map(symbol => (
            <div key={symbol} className="flex items-center justify-between p-3 bg-zinc-800/30 rounded-lg">
              <span className="font-mono">{symbol}</span>
              <div className="flex items-center gap-4">
                <span className="text-xs text-emerald-400">Trading</span>
                <button className="px-3 py-1 text-xs bg-red-500/20 text-red-400 rounded hover:bg-red-500/30 transition-colors">
                  Suspend
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function PlaceholderView({ title, icon }: { title: string; icon: React.ReactNode }) {
  return (
    <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-12 text-center">
      <div className="w-16 h-16 mx-auto mb-4 bg-zinc-800 rounded-xl flex items-center justify-center text-zinc-600">
        {icon}
      </div>
      <h3 className="font-semibold mb-2">{title}</h3>
      <p className="text-sm text-zinc-500">Coming soon in Phase 5</p>
    </div>
  );
}
