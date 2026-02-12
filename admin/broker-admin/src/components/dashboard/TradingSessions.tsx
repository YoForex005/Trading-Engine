'use client';

import { useState, useEffect, useMemo } from 'react';
import { Clock, Globe, Calendar, AlertCircle, Plus, Edit2, Trash2, X, Check } from 'lucide-react';
import { API_CONFIG } from '../../config/api';

// ============================================
// TypeScript Interfaces (matching backend)
// ============================================

interface BreakPeriod {
  startTime: string; // HH:MM format
  endTime: string;
  reason: string;
}

interface TradingSession {
  id: number;
  name: string;
  description: string;
  openTime: string;
  closeTime: string;
  timezone: string;
  breakPeriods: BreakPeriod[];
  weekendClosed: boolean;
  closedDays: string[];
  holidaySchedule: string;
  symbolsCount: number;
  createdAt: string;
  updatedAt: string;
}

interface SessionTemplate {
  id: number;
  name: string;
  type: string; // "24/5 Forex", "24/7 Crypto", "NYSE", "LSE", "Tokyo", "Custom"
  description: string;
  openTime: string;
  closeTime: string;
  timezone: string;
  breakPeriods: BreakPeriod[];
  weekendClosed: boolean;
  closedDays: string[];
  isDefault: boolean;
  createdAt: string;
}

interface SymbolSession {
  symbol: string;
  sessionId: number;
  sessionName: string;
  openTime: string;
  closeTime: string;
  timezone: string;
  breakPeriods: BreakPeriod[];
  weekendClosed: boolean;
  closedDays: string[];
  customOverride: boolean;
}

interface Holiday {
  id: number;
  date: string; // YYYY-MM-DD
  name: string;
  affectedSessions: string[];
  closureType: string; // "full" or "partial"
  partialHours?: string;
  region: string;
  createdAt: string;
}

interface MarketStatus {
  sessionId: number;
  sessionName: string;
  status: string; // "open", "closed", "pre-market", "after-hours", "break"
  currentTime: string;
  nextStateChange: string;
  timeToNextChange: string;
  isHoliday: boolean;
  holidayName?: string;
  currentBreak?: string;
}

// ============================================
// Component
// ============================================

export default function TradingSessions() {
  const [sessions, setSessions] = useState<TradingSession[]>([]);
  const [templates, setTemplates] = useState<SessionTemplate[]>([]);
  const [symbolSessions, setSymbolSessions] = useState<SymbolSession[]>([]);
  const [holidays, setHolidays] = useState<Holiday[]>([]);
  const [marketStatuses, setMarketStatuses] = useState<MarketStatus[]>([]);

  const [selectedSessionId, setSelectedSessionId] = useState<number | null>(null);
  const [showEditSessionModal, setShowEditSessionModal] = useState(false);
  const [showAddHolidayModal, setShowAddHolidayModal] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortColumn, setSortColumn] = useState<string>('symbol');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

  // Stats
  const totalSessions = sessions.length;
  const totalSymbols = symbolSessions.length;
  const holidaysThisMonth = useMemo(() => {
    const currentMonth = new Date().getMonth();
    return holidays.filter(h => new Date(h.date).getMonth() === currentMonth).length;
  }, [holidays]);
  const currentlyOpenSessions = useMemo(() => {
    return marketStatuses.filter(s => s.status === 'open').length;
  }, [marketStatuses]);

  // Fetch all data on mount
  useEffect(() => {
    const fetchData = async () => {
      const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
      if (!token) return;

      try {
        const headers = {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        };

        const [sessionsRes, templatesRes, holidaysRes, statusRes] = await Promise.all([
          fetch(API_CONFIG.TRADING_SESSIONS_LIST, { headers }),
          fetch(API_CONFIG.TRADING_SESSIONS_TEMPLATES, { headers }),
          fetch(API_CONFIG.TRADING_SESSIONS_HOLIDAYS, { headers }),
          fetch(API_CONFIG.TRADING_SESSIONS_STATUS, { headers })
        ]);

        const [sessionsData, templatesData, holidaysData, statusData] = await Promise.all([
          sessionsRes.json(),
          templatesRes.json(),
          holidaysRes.json(),
          statusRes.json()
        ]);

        setSessions(sessionsData.sessions || []);
        setTemplates(templatesData.templates || []);
        setHolidays(holidaysData.holidays || []);
        setMarketStatuses(statusData.statuses || []);

        // Fetch symbol sessions for first session (Forex as example - in real app would fetch all)
        // For demo, we'll generate mock symbol sessions
        generateMockSymbolSessions(sessionsData.sessions || []);

      } catch (error) {
        console.error('[TradingSessions] Error fetching data:', error);
      }
    };

    fetchData();

    // Refresh market status every 60 seconds
    const intervalId = setInterval(() => {
      refreshMarketStatus();
    }, 60000);

    return () => clearInterval(intervalId);
  }, []);

  const generateMockSymbolSessions = (sessions: TradingSession[]) => {
    // Generate 150 symbol sessions mapped to the 5 sessions
    const symbols: SymbolSession[] = [];

    // Forex pairs (50 symbols) - Session 1
    const forexPairs = [
      "EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "NZDUSD", "USDCAD",
      "EURGBP", "EURJPY", "EURAUD", "EURCHF", "GBPJPY", "GBPAUD", "GBPCHF",
      "AUDJPY", "AUDNZD", "AUDCAD", "AUDCHF", "NZDJPY", "NZDCAD", "NZDCHF",
      "CADJPY", "CADCHF", "CHFJPY", "EURCZK", "EURHUF", "EURPLN", "EURNOK",
      "EURSEK", "EURDKK", "EURTRY", "GBPNZD", "GBPPLN", "USDMXN", "USDNOK",
      "USDSEK", "USDDKK", "USDPLN", "USDTRY", "USDZAR", "USDHUF", "USDCZK",
      "EURZAR", "GBPZAR", "AUDSGD", "NZDSGD", "SGDJPY", "USDHKD", "USDSGD",
      "USDCNH"
    ];

    const session1 = sessions[0];
    if (session1) {
      forexPairs.forEach(symbol => {
        symbols.push({
          symbol,
          sessionId: session1.id,
          sessionName: session1.name,
          openTime: session1.openTime,
          closeTime: session1.closeTime,
          timezone: session1.timezone,
          breakPeriods: session1.breakPeriods,
          weekendClosed: session1.weekendClosed,
          closedDays: session1.closedDays,
          customOverride: false
        });
      });
    }

    // Crypto (30 symbols) - Session 2
    const cryptoPairs = [
      "BTCUSD", "ETHUSD", "XRPUSD", "SOLUSD", "BNBUSD", "ADAUSD", "DOGUSD",
      "DOTUSD", "MATICUSD", "LINKUSD", "AVAXUSD", "UNIUSD", "ATOMUSD", "LTCUSD",
      "ETCUSD", "XMRUSD", "BCHUSD", "ALGOUSD", "FILUSD", "VETUSDT", "ICPUSD",
      "FTMUSD", "SANDUSD", "MANAUSD", "GRTUSD", "ENJUSD", "CHZUSD", "THEUSD",
      "AXSUSD", "SHIBUSDT"
    ];

    const session2 = sessions[1];
    if (session2) {
      cryptoPairs.forEach(symbol => {
        symbols.push({
          symbol,
          sessionId: session2.id,
          sessionName: session2.name,
          openTime: session2.openTime,
          closeTime: session2.closeTime,
          timezone: session2.timezone,
          breakPeriods: session2.breakPeriods,
          weekendClosed: session2.weekendClosed,
          closedDays: session2.closedDays,
          customOverride: false
        });
      });
    }

    // US Stocks (40 symbols) - Session 3
    const usStocks = [
      "AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "NVDA", "META", "BRK.B",
      "JNJ", "V", "WMT", "JPM", "MA", "PG", "UNH", "DIS", "HD", "BAC",
      "ADBE", "CRM", "NFLX", "PYPL", "INTC", "CMCSA", "VZ", "T", "PFE",
      "KO", "PEP", "ABT", "MRK", "CSCO", "NKE", "TMO", "ABBV", "ACN",
      "AVGO", "TXN", "DHR", "QCOM"
    ];

    const session3 = sessions[2];
    if (session3) {
      usStocks.forEach(symbol => {
        symbols.push({
          symbol,
          sessionId: session3.id,
          sessionName: session3.name,
          openTime: session3.openTime,
          closeTime: session3.closeTime,
          timezone: session3.timezone,
          breakPeriods: session3.breakPeriods,
          weekendClosed: session3.weekendClosed,
          closedDays: session3.closedDays,
          customOverride: false
        });
      });
    }

    // UK Stocks (20 symbols) - Session 4
    const ukStocks = [
      "BP", "HSBA", "RIO", "GSK", "AZN", "SHEL", "DGE", "ULVR", "BATS",
      "REL", "NG", "LSEG", "AAL", "BARC", "LLOY", "PRU", "VOD", "BT.A",
      "IMB", "CCH"
    ];

    const session4 = sessions[3];
    if (session4) {
      ukStocks.forEach(symbol => {
        symbols.push({
          symbol,
          sessionId: session4.id,
          sessionName: session4.name,
          openTime: session4.openTime,
          closeTime: session4.closeTime,
          timezone: session4.timezone,
          breakPeriods: session4.breakPeriods,
          weekendClosed: session4.weekendClosed,
          closedDays: session4.closedDays,
          customOverride: false
        });
      });
    }

    // Japan Stocks (10 symbols) - Session 5
    const japanStocks = [
      "7203.T", "9984.T", "6758.T", "8306.T", "6861.T", "9433.T",
      "7267.T", "8035.T", "6098.T", "4063.T"
    ];

    const session5 = sessions[4];
    if (session5) {
      japanStocks.forEach(symbol => {
        symbols.push({
          symbol,
          sessionId: session5.id,
          sessionName: session5.name,
          openTime: session5.openTime,
          closeTime: session5.closeTime,
          timezone: session5.timezone,
          breakPeriods: session5.breakPeriods,
          weekendClosed: session5.weekendClosed,
          closedDays: session5.closedDays,
          customOverride: false
        });
      });
    }

    setSymbolSessions(symbols);
  };

  const refreshMarketStatus = async () => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('adminToken') : null;
    if (!token) return;

    try {
      const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      };

      const statusRes = await fetch(API_CONFIG.TRADING_SESSIONS_STATUS, { headers });
      const statusData = await statusRes.json();
      setMarketStatuses(statusData.statuses || []);
    } catch (error) {
      console.error('[TradingSessions] Error refreshing market status:', error);
    }
  };

  // Get session icon emoji
  const getSessionIcon = (name: string): string => {
    if (name.includes('Forex')) return '💱';
    if (name.includes('Crypto')) return '₿';
    if (name.includes('US')) return '🇺🇸';
    if (name.includes('UK')) return '🇬🇧';
    if (name.includes('Japan')) return '🇯🇵';
    return '🌐';
  };

  // Get template badge color
  const getTemplateBadgeColor = (type: string): string => {
    const colors: Record<string, string> = {
      '24/5 Forex': '#3B82F6',
      '24/7 Crypto': '#10B981',
      'NYSE': '#F59E0B',
      'LSE': '#8B5CF6',
      'Tokyo': '#EF4444',
      'Custom': '#6B7280'
    };
    return colors[type] || '#6B7280';
  };

  // Get status badge color
  const getStatusBadgeColor = (status: string): string => {
    const colors: Record<string, string> = {
      'open': '#10B981',
      'closed': '#EF4444',
      'pre-market': '#F59E0B',
      'after-hours': '#8B5CF6',
      'break': '#6B7280'
    };
    return colors[status] || '#6B7280';
  };

  // Get closure type badge color
  const getClosureTypeBadgeColor = (type: string): string => {
    return type === 'full' ? '#EF4444' : '#F59E0B';
  };

  // Filter and sort symbol sessions
  const filteredSymbols = useMemo(() => {
    let filtered = symbolSessions;

    // Search filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(s =>
        s.symbol.toLowerCase().includes(query) ||
        s.sessionName.toLowerCase().includes(query)
      );
    }

    // Sort
    filtered.sort((a, b) => {
      let aVal: any = a[sortColumn as keyof SymbolSession];
      let bVal: any = b[sortColumn as keyof SymbolSession];

      if (typeof aVal === 'string') aVal = aVal.toLowerCase();
      if (typeof bVal === 'string') bVal = bVal.toLowerCase();

      if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
      if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
      return 0;
    });

    return filtered;
  }, [symbolSessions, searchQuery, sortColumn, sortDirection]);

  const handleSort = (column: string) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  const applyTemplate = async (template: SessionTemplate) => {
    console.log(`[TradingSessions] Applying template: ${template.name}`);
    // In real app, would create a new session from template via POST
    alert(`Template "${template.name}" would be applied (not implemented in demo)`);
  };

  const handleAddHoliday = async (e: React.FormEvent) => {
    e.preventDefault();
    // In real app, would POST to API_CONFIG.TRADING_SESSIONS_HOLIDAYS
    console.log('[TradingSessions] Add holiday (not implemented in demo)');
    setShowAddHolidayModal(false);
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#E0E0E0] overflow-hidden">
      {/* Stats Cards */}
      <div className="flex-shrink-0 p-3 border-b border-[#383A42]">
        <div className="grid grid-cols-4 gap-3">
          <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
            <div className="text-xs text-[#888] mb-1">Total Sessions</div>
            <div className="text-2xl font-bold text-[#F5C542]">{totalSessions}</div>
          </div>
          <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
            <div className="text-xs text-[#888] mb-1">Total Symbols</div>
            <div className="text-2xl font-bold text-[#3B82F6]">{totalSymbols}</div>
          </div>
          <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
            <div className="text-xs text-[#888] mb-1">Holidays This Month</div>
            <div className="text-2xl font-bold text-[#F97316]">{holidaysThisMonth}</div>
          </div>
          <div className="bg-[#1E2026] rounded p-3 border border-[#383A42]">
            <div className="text-xs text-[#888] mb-1">Currently Open</div>
            <div className="text-2xl font-bold text-[#10B981]">{currentlyOpenSessions}</div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex gap-3 p-3 overflow-hidden">
        {/* Left Column: Sessions + Templates */}
        <div className="w-1/3 flex flex-col gap-3 overflow-hidden">
          {/* Session List Cards */}
          <div className="flex-1 flex flex-col bg-[#1E2026] rounded border border-[#383A42] overflow-hidden">
            <div className="flex-shrink-0 px-3 py-2 border-b border-[#383A42] flex items-center justify-between">
              <div className="text-sm font-bold text-[#F5C542]">Trading Sessions ({sessions.length})</div>
            </div>
            <div className="flex-1 overflow-y-auto p-2 space-y-2">
              {sessions.map(session => {
                const status = marketStatuses.find(s => s.sessionId === session.id);
                return (
                  <div
                    key={session.id}
                    className={`bg-[#121316] rounded p-3 border cursor-pointer transition-colors ${
                      selectedSessionId === session.id
                        ? 'border-[#F5C542]'
                        : 'border-[#383A42] hover:border-[#555]'
                    }`}
                    onClick={() => setSelectedSessionId(session.id)}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <span className="text-xl">{getSessionIcon(session.name)}</span>
                        <div>
                          <div className="text-sm font-bold text-white">{session.name}</div>
                          <div className="text-xs text-[#888]">{session.description}</div>
                        </div>
                      </div>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedSessionId(session.id);
                          setShowEditSessionModal(true);
                        }}
                        className="p-1 hover:bg-[#383A42] rounded"
                      >
                        <Edit2 size={14} className="text-[#888]" />
                      </button>
                    </div>
                    <div className="grid grid-cols-2 gap-2 text-xs">
                      <div>
                        <span className="text-[#888]">Open: </span>
                        <span className="text-white font-mono">{session.openTime}</span>
                      </div>
                      <div>
                        <span className="text-[#888]">Close: </span>
                        <span className="text-white font-mono">{session.closeTime}</span>
                      </div>
                      <div>
                        <span className="text-[#888]">Timezone: </span>
                        <span className="text-white">{session.timezone}</span>
                      </div>
                      <div>
                        <span className="text-[#888]">Symbols: </span>
                        <span className="text-[#F5C542] font-bold">{session.symbolsCount}</span>
                      </div>
                    </div>
                    {status && (
                      <div className="mt-2 pt-2 border-t border-[#383A42] flex items-center justify-between">
                        <div
                          className="text-xs px-2 py-0.5 rounded font-bold"
                          style={{ backgroundColor: getStatusBadgeColor(status.status) + '20', color: getStatusBadgeColor(status.status) }}
                        >
                          {status.status.toUpperCase()}
                        </div>
                        <div className="text-xs text-[#888]">{status.timeToNextChange}</div>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          {/* Template Gallery */}
          <div className="flex-1 flex flex-col bg-[#1E2026] rounded border border-[#383A42] overflow-hidden">
            <div className="flex-shrink-0 px-3 py-2 border-b border-[#383A42] flex items-center justify-between">
              <div className="text-sm font-bold text-[#F5C542]">Session Templates ({templates.length})</div>
            </div>
            <div className="flex-1 overflow-y-auto p-2 space-y-2">
              {templates.map(template => (
                <div
                  key={template.id}
                  className="bg-[#121316] rounded p-3 border border-[#383A42] hover:border-[#555] transition-colors"
                >
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <div className="text-sm font-bold text-white mb-1">{template.name}</div>
                      <div className="text-xs text-[#888] mb-2">{template.description}</div>
                      <div
                        className="inline-block text-xs px-2 py-0.5 rounded font-bold"
                        style={{ backgroundColor: getTemplateBadgeColor(template.type) + '20', color: getTemplateBadgeColor(template.type) }}
                      >
                        {template.type}
                      </div>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-2 text-xs mt-2">
                    <div>
                      <span className="text-[#888]">Hours: </span>
                      <span className="text-white font-mono">{template.openTime}-{template.closeTime}</span>
                    </div>
                    <div>
                      <span className="text-[#888]">Weekend: </span>
                      <span className={template.weekendClosed ? "text-[#EF4444]" : "text-[#10B981]"}>
                        {template.weekendClosed ? 'Closed' : 'Open'}
                      </span>
                    </div>
                  </div>
                  <button
                    onClick={() => applyTemplate(template)}
                    className="w-full mt-2 px-3 py-1.5 bg-[#F5C542] hover:bg-[#E5B532] text-[#121316] rounded text-xs font-bold transition-colors"
                  >
                    Quick Apply
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Middle Column: Symbol-Session Mapping Table */}
        <div className="flex-1 flex flex-col bg-[#1E2026] rounded border border-[#383A42] overflow-hidden">
          <div className="flex-shrink-0 px-3 py-2 border-b border-[#383A42] flex items-center justify-between">
            <div className="text-sm font-bold text-[#F5C542]">Symbol-Session Mapping ({filteredSymbols.length})</div>
            <input
              type="text"
              placeholder="Search symbols..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="px-2 py-1 bg-[#121316] border border-[#383A42] rounded text-xs text-white focus:outline-none focus:border-[#F5C542]"
            />
          </div>
          <div className="flex-1 overflow-auto">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-[#383A42]">
                <tr>
                  <th
                    className="px-3 py-2 text-left cursor-pointer hover:bg-[#383A42] transition-colors"
                    onClick={() => handleSort('symbol')}
                  >
                    Symbol {sortColumn === 'symbol' && (sortDirection === 'asc' ? '↑' : '↓')}
                  </th>
                  <th
                    className="px-3 py-2 text-left cursor-pointer hover:bg-[#383A42] transition-colors"
                    onClick={() => handleSort('sessionName')}
                  >
                    Session {sortColumn === 'sessionName' && (sortDirection === 'asc' ? '↑' : '↓')}
                  </th>
                  <th className="px-3 py-2 text-left">Trading Hours</th>
                  <th className="px-3 py-2 text-left">Timezone</th>
                  <th className="px-3 py-2 text-left">Break Periods</th>
                  <th className="px-3 py-2 text-left">Weekend</th>
                  <th className="px-3 py-2 text-left">Custom</th>
                </tr>
              </thead>
              <tbody>
                {filteredSymbols.map((symbolSession, idx) => (
                  <tr
                    key={`${symbolSession.symbol}-${idx}`}
                    className="border-b border-[#383A42] hover:bg-[#25272E] transition-colors"
                  >
                    <td className="px-3 py-2 font-mono font-bold text-white">{symbolSession.symbol}</td>
                    <td className="px-3 py-2 text-[#888]">{symbolSession.sessionName}</td>
                    <td className="px-3 py-2 font-mono text-white">
                      {symbolSession.openTime} - {symbolSession.closeTime}
                    </td>
                    <td className="px-3 py-2 text-[#888]">{symbolSession.timezone}</td>
                    <td className="px-3 py-2 text-[#888]">
                      {symbolSession.breakPeriods.length > 0 ? (
                        <div className="text-xs">
                          {symbolSession.breakPeriods.map((bp, i) => (
                            <div key={i}>{bp.startTime}-{bp.endTime}</div>
                          ))}
                        </div>
                      ) : (
                        <span className="text-[#555]">None</span>
                      )}
                    </td>
                    <td className="px-3 py-2">
                      <span className={symbolSession.weekendClosed ? "text-[#EF4444]" : "text-[#10B981]"}>
                        {symbolSession.weekendClosed ? 'Closed' : 'Open'}
                      </span>
                    </td>
                    <td className="px-3 py-2">
                      {symbolSession.customOverride ? (
                        <span className="text-[#F59E0B]">✓</span>
                      ) : (
                        <span className="text-[#555]">—</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Right Column: Holiday Calendar + Market Status */}
        <div className="w-1/3 flex flex-col gap-3 overflow-hidden">
          {/* Market Status Panel */}
          <div className="flex-1 flex flex-col bg-[#1E2026] rounded border border-[#383A42] overflow-hidden">
            <div className="flex-shrink-0 px-3 py-2 border-b border-[#383A42] flex items-center justify-between">
              <div className="text-sm font-bold text-[#F5C542]">Market Status (Real-Time)</div>
              <button
                onClick={refreshMarketStatus}
                className="text-xs text-[#888] hover:text-[#F5C542] transition-colors"
              >
                Refresh
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-2 space-y-2">
              {marketStatuses.map(status => (
                <div
                  key={status.sessionId}
                  className="bg-[#121316] rounded p-3 border border-[#383A42]"
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="text-sm font-bold text-white">{status.sessionName}</div>
                    <div
                      className="text-xs px-2 py-0.5 rounded font-bold"
                      style={{ backgroundColor: getStatusBadgeColor(status.status) + '20', color: getStatusBadgeColor(status.status) }}
                    >
                      {status.status.toUpperCase()}
                    </div>
                  </div>
                  {status.isHoliday ? (
                    <div className="flex items-center gap-2 text-xs text-[#EF4444]">
                      <AlertCircle size={14} />
                      <span>Holiday: {status.holidayName}</span>
                    </div>
                  ) : (
                    <div className="space-y-1 text-xs">
                      <div className="flex justify-between">
                        <span className="text-[#888]">Next Change:</span>
                        <span className="text-white font-mono">{status.nextStateChange}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-[#888]">Time Until:</span>
                        <span className="text-[#F5C542] font-bold">{status.timeToNextChange}</span>
                      </div>
                      {status.currentBreak && (
                        <div className="flex items-center gap-2 text-[#F59E0B]">
                          <Clock size={14} />
                          <span>Break: {status.currentBreak}</span>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Holiday Calendar */}
          <div className="flex-1 flex flex-col bg-[#1E2026] rounded border border-[#383A42] overflow-hidden">
            <div className="flex-shrink-0 px-3 py-2 border-b border-[#383A42] flex items-center justify-between">
              <div className="text-sm font-bold text-[#F5C542]">Holiday Calendar 2026 ({holidays.length})</div>
              <button
                onClick={() => setShowAddHolidayModal(true)}
                className="flex items-center gap-1 px-2 py-1 bg-[#F5C542] hover:bg-[#E5B532] text-[#121316] rounded text-xs font-bold transition-colors"
              >
                <Plus size={14} />
                Add
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-2 space-y-2">
              {holidays.map(holiday => (
                <div
                  key={holiday.id}
                  className="bg-[#121316] rounded p-3 border border-[#383A42]"
                >
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <div className="text-sm font-bold text-white mb-1">{holiday.name}</div>
                      <div className="text-xs text-[#888] font-mono">{holiday.date}</div>
                    </div>
                    <div
                      className="text-xs px-2 py-0.5 rounded font-bold"
                      style={{
                        backgroundColor: getClosureTypeBadgeColor(holiday.closureType) + '20',
                        color: getClosureTypeBadgeColor(holiday.closureType)
                      }}
                    >
                      {holiday.closureType === 'full' ? 'FULL' : 'PARTIAL'}
                    </div>
                  </div>
                  <div className="text-xs space-y-1">
                    <div>
                      <span className="text-[#888]">Region: </span>
                      <span className="text-white">{holiday.region}</span>
                    </div>
                    {holiday.partialHours && (
                      <div>
                        <span className="text-[#888]">Hours: </span>
                        <span className="text-[#F59E0B] font-mono">{holiday.partialHours}</span>
                      </div>
                    )}
                    <div className="text-[#888] mt-1">
                      Affects: {holiday.affectedSessions.join(', ')}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Edit Session Modal */}
      {showEditSessionModal && selectedSessionId && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] rounded border border-[#383A42] w-[600px] max-h-[80vh] overflow-hidden">
            <div className="px-4 py-3 border-b border-[#383A42] flex items-center justify-between">
              <div className="text-sm font-bold text-[#F5C542]">Edit Session</div>
              <button
                onClick={() => setShowEditSessionModal(false)}
                className="p-1 hover:bg-[#383A42] rounded transition-colors"
              >
                <X size={16} className="text-[#888]" />
              </button>
            </div>
            <div className="p-4 overflow-y-auto max-h-[calc(80vh-120px)]">
              <form className="space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs text-[#888] mb-1">Open Time (HH:MM)</label>
                    <input
                      type="text"
                      placeholder="22:00"
                      className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-[#F5C542]"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-[#888] mb-1">Close Time (HH:MM)</label>
                    <input
                      type="text"
                      placeholder="22:00"
                      className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-[#F5C542]"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Timezone</label>
                  <select className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-[#F5C542]">
                    <option value="UTC">UTC</option>
                    <option value="America/New_York">America/New_York (EST/EDT)</option>
                    <option value="Europe/London">Europe/London (GMT/BST)</option>
                    <option value="Asia/Tokyo">Asia/Tokyo (JST)</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Break Periods</label>
                  <div className="text-xs text-[#666] italic">Add break periods (e.g., server maintenance, lunch break)</div>
                </div>
                <div className="flex items-center gap-2">
                  <input type="checkbox" id="weekendClosed" className="rounded" />
                  <label htmlFor="weekendClosed" className="text-sm text-white">Weekend Closed</label>
                </div>
              </form>
            </div>
            <div className="px-4 py-3 border-t border-[#383A42] flex items-center justify-end gap-2">
              <button
                onClick={() => setShowEditSessionModal(false)}
                className="px-4 py-2 bg-[#383A42] hover:bg-[#4A4C54] text-white rounded text-sm font-bold transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => {
                  console.log('[TradingSessions] Save session (not implemented in demo)');
                  setShowEditSessionModal(false);
                }}
                className="px-4 py-2 bg-[#F5C542] hover:bg-[#E5B532] text-[#121316] rounded text-sm font-bold transition-colors"
              >
                Save Changes
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Holiday Modal */}
      {showAddHolidayModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-[#1E2026] rounded border border-[#383A42] w-[600px] max-h-[80vh] overflow-hidden">
            <div className="px-4 py-3 border-b border-[#383A42] flex items-center justify-between">
              <div className="text-sm font-bold text-[#F5C542]">Add Holiday</div>
              <button
                onClick={() => setShowAddHolidayModal(false)}
                className="p-1 hover:bg-[#383A42] rounded transition-colors"
              >
                <X size={16} className="text-[#888]" />
              </button>
            </div>
            <form onSubmit={handleAddHoliday}>
              <div className="p-4 overflow-y-auto max-h-[calc(80vh-120px)] space-y-3">
                <div>
                  <label className="block text-xs text-[#888] mb-1">Date</label>
                  <input
                    type="date"
                    required
                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-[#F5C542]"
                  />
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Holiday Name</label>
                  <input
                    type="text"
                    placeholder="e.g., Christmas Day"
                    required
                    className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-[#F5C542]"
                  />
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Affected Sessions (select multiple)</label>
                  <div className="space-y-2 mt-2">
                    {sessions.map(session => (
                      <label key={session.id} className="flex items-center gap-2">
                        <input type="checkbox" className="rounded" />
                        <span className="text-sm text-white">{session.name}</span>
                      </label>
                    ))}
                  </div>
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Closure Type</label>
                  <select className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-[#F5C542]">
                    <option value="full">Full Closure</option>
                    <option value="partial">Partial Closure</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs text-[#888] mb-1">Region</label>
                  <select className="w-full px-3 py-2 bg-[#121316] border border-[#383A42] rounded text-sm text-white focus:outline-none focus:border-[#F5C542]">
                    <option value="US">United States</option>
                    <option value="UK">United Kingdom</option>
                    <option value="Japan">Japan</option>
                    <option value="Global">Global</option>
                  </select>
                </div>
              </div>
              <div className="px-4 py-3 border-t border-[#383A42] flex items-center justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setShowAddHolidayModal(false)}
                  className="px-4 py-2 bg-[#383A42] hover:bg-[#4A4C54] text-white rounded text-sm font-bold transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-[#F5C542] hover:bg-[#E5B532] text-[#121316] rounded text-sm font-bold transition-colors"
                >
                  Add Holiday
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
