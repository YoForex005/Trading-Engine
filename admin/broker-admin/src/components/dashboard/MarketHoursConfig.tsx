/**
 * Market Hours / Sessions Configuration
 * Admin page to configure trading session schedules per symbol
 */

'use client';

import React, { useState, useMemo } from 'react';
import {
  Search,
  Clock,
  Calendar,
  Plus,
  X,
  Save,
  RotateCcw,
  Globe,
  ChevronDown,
  Trash2,
  CheckCircle,
} from 'lucide-react';
import { API_CONFIG } from '@/config/api';

// Types
type AssetClass = 'Forex' | 'Crypto' | 'Indices' | 'Commodities' | 'Metals';
type SessionType = 'PRE_MARKET' | 'REGULAR' | 'AFTER_HOURS' | 'CLOSED';
type DayOfWeek = 'Monday' | 'Tuesday' | 'Wednesday' | 'Thursday' | 'Friday' | 'Saturday' | 'Sunday';

interface Symbol {
  symbol: string;
  name: string;
  assetClass: AssetClass;
  timezone: string;
}

interface SessionBlock {
  id: string;
  startTime: string; // HH:MM format
  endTime: string; // HH:MM format
  type: SessionType;
}

interface DaySchedule {
  day: DayOfWeek;
  sessions: SessionBlock[];
}

interface Holiday {
  id: string;
  date: string; // YYYY-MM-DD
  description: string;
}

interface SymbolSchedule {
  symbol: string;
  timezone: string;
  schedule: DaySchedule[];
  holidays: Holiday[];
}

// Session type colors
const SESSION_COLORS: Record<SessionType, { bg: string; text: string; border: string }> = {
  PRE_MARKET: { bg: 'bg-yellow-500/20', text: 'text-yellow-400', border: 'border-yellow-500' },
  REGULAR: { bg: 'bg-emerald-500/20', text: 'text-emerald-400', border: 'border-emerald-500' },
  AFTER_HOURS: { bg: 'bg-blue-500/20', text: 'text-blue-400', border: 'border-blue-500' },
  CLOSED: { bg: 'bg-zinc-700/20', text: 'text-zinc-500', border: 'border-zinc-600' },
};

// Mock symbols grouped by asset class
const MOCK_SYMBOLS: Symbol[] = [
  // Forex
  { symbol: 'EURUSD', name: 'Euro / US Dollar', assetClass: 'Forex', timezone: 'UTC' },
  { symbol: 'GBPUSD', name: 'British Pound / US Dollar', assetClass: 'Forex', timezone: 'UTC' },
  { symbol: 'USDJPY', name: 'US Dollar / Japanese Yen', assetClass: 'Forex', timezone: 'UTC' },
  { symbol: 'AUDUSD', name: 'Australian Dollar / US Dollar', assetClass: 'Forex', timezone: 'UTC' },
  { symbol: 'USDCAD', name: 'US Dollar / Canadian Dollar', assetClass: 'Forex', timezone: 'UTC' },
  // Crypto
  { symbol: 'BTCUSD', name: 'Bitcoin / US Dollar', assetClass: 'Crypto', timezone: 'UTC' },
  { symbol: 'ETHUSD', name: 'Ethereum / US Dollar', assetClass: 'Crypto', timezone: 'UTC' },
  { symbol: 'XRPUSD', name: 'Ripple / US Dollar', assetClass: 'Crypto', timezone: 'UTC' },
  // Indices
  { symbol: 'SPX500', name: 'S&P 500 Index', assetClass: 'Indices', timezone: 'America/New_York' },
  { symbol: 'US30', name: 'Dow Jones Industrial Average', assetClass: 'Indices', timezone: 'America/New_York' },
  { symbol: 'NAS100', name: 'NASDAQ 100 Index', assetClass: 'Indices', timezone: 'America/New_York' },
  { symbol: 'UK100', name: 'FTSE 100 Index', assetClass: 'Indices', timezone: 'Europe/London' },
  { symbol: 'DE30', name: 'DAX 30 Index', assetClass: 'Indices', timezone: 'Europe/Berlin' },
  // Commodities
  { symbol: 'WTIUSD', name: 'WTI Crude Oil', assetClass: 'Commodities', timezone: 'America/New_York' },
  { symbol: 'BRENTUSD', name: 'Brent Crude Oil', assetClass: 'Commodities', timezone: 'Europe/London' },
  { symbol: 'NATGAS', name: 'Natural Gas', assetClass: 'Commodities', timezone: 'America/New_York' },
  // Metals
  { symbol: 'XAUUSD', name: 'Gold / US Dollar', assetClass: 'Metals', timezone: 'UTC' },
  { symbol: 'XAGUSD', name: 'Silver / US Dollar', assetClass: 'Metals', timezone: 'UTC' },
];

// Timezone options
const TIMEZONES = [
  { value: 'UTC', label: 'UTC' },
  { value: 'America/New_York', label: 'Eastern Time (ET)' },
  { value: 'America/Chicago', label: 'Central Time (CT)' },
  { value: 'America/Los_Angeles', label: 'Pacific Time (PT)' },
  { value: 'Europe/London', label: 'London (GMT/BST)' },
  { value: 'Europe/Berlin', label: 'Berlin (CET/CEST)' },
  { value: 'Asia/Tokyo', label: 'Tokyo (JST)' },
  { value: 'Asia/Hong_Kong', label: 'Hong Kong (HKT)' },
  { value: 'Australia/Sydney', label: 'Sydney (AEST/AEDT)' },
];

const DAYS_OF_WEEK: DayOfWeek[] = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];

// Generate default schedule
const generateDefaultSchedule = (): DaySchedule[] => {
  return DAYS_OF_WEEK.map(day => ({
    day,
    sessions: [{ id: `${day}-1`, startTime: '00:00', endTime: '23:59', type: 'CLOSED' as SessionType }],
  }));
};

// Preset schedules
const PRESETS = {
  'Forex 24/5': (): DaySchedule[] => {
    const schedule = generateDefaultSchedule();
    schedule.forEach((day, index) => {
      if (index < 5) { // Monday-Friday
        day.sessions = [{ id: `${day.day}-1`, startTime: '00:00', endTime: '23:59', type: 'REGULAR' as SessionType }];
      }
    });
    return schedule;
  },
  'Crypto 24/7': (): DaySchedule[] => {
    return DAYS_OF_WEEK.map(day => ({
      day,
      sessions: [{ id: `${day}-1`, startTime: '00:00', endTime: '23:59', type: 'REGULAR' as SessionType }],
    }));
  },
  'US Market Hours': (): DaySchedule[] => {
    const schedule = generateDefaultSchedule();
    schedule.forEach((day, index) => {
      if (index < 5) { // Monday-Friday
        day.sessions = [
          { id: `${day.day}-1`, startTime: '04:00', endTime: '09:30', type: 'PRE_MARKET' as SessionType },
          { id: `${day.day}-2`, startTime: '09:30', endTime: '16:00', type: 'REGULAR' as SessionType },
          { id: `${day.day}-3`, startTime: '16:00', endTime: '20:00', type: 'AFTER_HOURS' as SessionType },
        ];
      }
    });
    return schedule;
  },
  'EU Market Hours': (): DaySchedule[] => {
    const schedule = generateDefaultSchedule();
    schedule.forEach((day, index) => {
      if (index < 5) { // Monday-Friday
        day.sessions = [
          { id: `${day.day}-1`, startTime: '08:00', endTime: '17:30', type: 'REGULAR' as SessionType },
        ];
      }
    });
    return schedule;
  },
};

export default function MarketHoursConfig() {
  const [selectedSymbol, setSelectedSymbol] = useState<Symbol | null>(MOCK_SYMBOLS[0]);
  const [searchQuery, setSearchQuery] = useState('');
  const [editingDay, setEditingDay] = useState<DayOfWeek | null>(null);
  const [showAddSession, setShowAddSession] = useState(false);
  const [showAddHoliday, setShowAddHoliday] = useState(false);

  // Symbol schedules state (in real app, this would be fetched from API)
  const [schedules, setSchedules] = useState<Record<string, SymbolSchedule>>({});

  // Get or create schedule for selected symbol
  const currentSchedule = useMemo(() => {
    if (!selectedSymbol) return null;
    return schedules[selectedSymbol.symbol] || {
      symbol: selectedSymbol.symbol,
      timezone: selectedSymbol.timezone,
      schedule: generateDefaultSchedule(),
      holidays: [],
    };
  }, [selectedSymbol, schedules]);

  // New session form state
  const [newSession, setNewSession] = useState<Partial<SessionBlock>>({
    startTime: '09:00',
    endTime: '17:00',
    type: 'REGULAR',
  });

  // New holiday form state
  const [newHoliday, setNewHoliday] = useState({
    date: '',
    description: '',
  });

  // Filtered symbols
  const filteredSymbols = useMemo(() => {
    const query = searchQuery.toLowerCase();
    return MOCK_SYMBOLS.filter(s =>
      s.symbol.toLowerCase().includes(query) || s.name.toLowerCase().includes(query)
    );
  }, [searchQuery]);

  // Group symbols by asset class
  const symbolsByClass = useMemo(() => {
    const grouped: Record<AssetClass, Symbol[]> = {
      Forex: [],
      Crypto: [],
      Indices: [],
      Commodities: [],
      Metals: [],
    };
    filteredSymbols.forEach(symbol => {
      grouped[symbol.assetClass].push(symbol);
    });
    return grouped;
  }, [filteredSymbols]);

  // Summary stats
  const stats = useMemo(() => {
    const symbolsConfigured = Object.keys(schedules).length;
    const currentlyOpen = 0; // TODO: Calculate based on current time
    const nextOpen = 'Monday 09:30 ET'; // TODO: Calculate
    const nextClose = 'Friday 16:00 ET'; // TODO: Calculate

    return { symbolsConfigured, currentlyOpen, nextOpen, nextClose };
  }, [schedules]);

  // Apply preset
  const applyPreset = (presetName: keyof typeof PRESETS) => {
    if (!selectedSymbol || !currentSchedule) return;
    const presetSchedule = PRESETS[presetName]();
    const updatedSchedule: SymbolSchedule = {
      ...currentSchedule,
      schedule: presetSchedule,
    };
    setSchedules({
      ...schedules,
      [selectedSymbol.symbol]: updatedSchedule,
    });
  };

  // Add session to day
  const addSession = () => {
    if (!selectedSymbol || !currentSchedule || !editingDay) return;

    const daySchedule = currentSchedule.schedule.find(d => d.day === editingDay);
    if (!daySchedule) return;

    const newSessionBlock: SessionBlock = {
      id: `${editingDay}-${Date.now()}`,
      startTime: newSession.startTime || '09:00',
      endTime: newSession.endTime || '17:00',
      type: newSession.type || 'REGULAR',
    };

    daySchedule.sessions.push(newSessionBlock);
    daySchedule.sessions.sort((a, b) => a.startTime.localeCompare(b.startTime));

    setSchedules({
      ...schedules,
      [selectedSymbol.symbol]: { ...currentSchedule },
    });

    setShowAddSession(false);
    setNewSession({ startTime: '09:00', endTime: '17:00', type: 'REGULAR' });
  };

  // Remove session
  const removeSession = (day: DayOfWeek, sessionId: string) => {
    if (!selectedSymbol || !currentSchedule) return;

    const daySchedule = currentSchedule.schedule.find(d => d.day === day);
    if (!daySchedule) return;

    daySchedule.sessions = daySchedule.sessions.filter(s => s.id !== sessionId);

    setSchedules({
      ...schedules,
      [selectedSymbol.symbol]: { ...currentSchedule },
    });
  };

  // Add holiday
  const addHoliday = () => {
    if (!selectedSymbol || !currentSchedule || !newHoliday.date) return;

    const holiday: Holiday = {
      id: `holiday-${Date.now()}`,
      date: newHoliday.date,
      description: newHoliday.description,
    };

    currentSchedule.holidays.push(holiday);
    currentSchedule.holidays.sort((a, b) => a.date.localeCompare(b.date));

    setSchedules({
      ...schedules,
      [selectedSymbol.symbol]: { ...currentSchedule },
    });

    setShowAddHoliday(false);
    setNewHoliday({ date: '', description: '' });
  };

  // Remove holiday
  const removeHoliday = (holidayId: string) => {
    if (!selectedSymbol || !currentSchedule) return;

    currentSchedule.holidays = currentSchedule.holidays.filter(h => h.id !== holidayId);

    setSchedules({
      ...schedules,
      [selectedSymbol.symbol]: { ...currentSchedule },
    });
  };

  // Update timezone
  const updateTimezone = (timezone: string) => {
    if (!selectedSymbol || !currentSchedule) return;

    setSchedules({
      ...schedules,
      [selectedSymbol.symbol]: {
        ...currentSchedule,
        timezone,
      },
    });
  };

  // Save changes (would make API call)
  const handleSave = () => {
    console.log('Saving market hours config:', schedules);
    // TODO: API call to save schedules
  };

  // Reset to default
  const handleReset = () => {
    if (!selectedSymbol) return;
    const { [selectedSymbol.symbol]: _, ...rest } = schedules;
    setSchedules(rest);
  };

  return (
    <div className="h-full flex flex-col bg-[#18181b] text-zinc-300">
      {/* Summary Cards */}
      <div className="grid grid-cols-4 gap-4 p-4">
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-blue-500/20 flex items-center justify-center">
              <CheckCircle size={20} className="text-blue-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-blue-400">{stats.symbolsConfigured}</div>
              <div className="text-xs text-zinc-500">Symbols Configured</div>
            </div>
          </div>
        </div>

        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-emerald-500/20 flex items-center justify-center">
              <Clock size={20} className="text-emerald-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-emerald-400">{stats.currentlyOpen}</div>
              <div className="text-xs text-zinc-500">Currently Open Markets</div>
            </div>
          </div>
        </div>

        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-yellow-500/20 flex items-center justify-center">
              <Calendar size={20} className="text-yellow-400" />
            </div>
            <div>
              <div className="text-sm font-bold text-yellow-400">{stats.nextOpen}</div>
              <div className="text-xs text-zinc-500">Next Market Open</div>
            </div>
          </div>
        </div>

        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded bg-rose-500/20 flex items-center justify-center">
              <Calendar size={20} className="text-rose-400" />
            </div>
            <div>
              <div className="text-sm font-bold text-rose-400">{stats.nextClose}</div>
              <div className="text-xs text-zinc-500">Next Market Close</div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex gap-4 p-4 pt-0 overflow-hidden">
        {/* Left Sidebar - Symbol List */}
        <div className="w-64 flex flex-col bg-[#27272a] border border-[#3f3f46] rounded overflow-hidden">
          {/* Search */}
          <div className="p-3 border-b border-[#3f3f46]">
            <div className="relative">
              <Search size={14} className="absolute left-2 top-1/2 transform -translate-y-1/2 text-zinc-500" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search symbols..."
                className="w-full pl-8 pr-3 py-1.5 text-xs bg-[#18181b] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          {/* Symbol List */}
          <div className="flex-1 overflow-auto custom-scrollbar">
            {Object.entries(symbolsByClass).map(([assetClass, symbols]) => (
              symbols.length > 0 && (
                <div key={assetClass} className="mb-2">
                  <div className="px-3 py-2 text-xs font-bold text-zinc-500 bg-[#1e1e1e] sticky top-0">
                    {assetClass}
                  </div>
                  {symbols.map(symbol => (
                    <div
                      key={symbol.symbol}
                      onClick={() => setSelectedSymbol(symbol)}
                      className={`px-3 py-2 text-xs cursor-pointer transition-colors ${
                        selectedSymbol?.symbol === symbol.symbol
                          ? 'bg-blue-500/20 text-blue-400 border-l-2 border-blue-500'
                          : 'hover:bg-zinc-800/30 text-zinc-300'
                      }`}
                    >
                      <div className="font-semibold">{symbol.symbol}</div>
                      <div className="text-[10px] text-zinc-500">{symbol.name}</div>
                      {schedules[symbol.symbol] && (
                        <div className="text-[10px] text-emerald-400 mt-0.5">✓ Configured</div>
                      )}
                    </div>
                  ))}
                </div>
              )
            ))}
          </div>
        </div>

        {/* Right Content */}
        <div className="flex-1 flex flex-col gap-4 overflow-hidden">
          {selectedSymbol && currentSchedule ? (
            <>
              {/* Header */}
              <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-lg font-bold text-zinc-200">{selectedSymbol.symbol}</h2>
                    <div className="text-xs text-zinc-500">{selectedSymbol.name}</div>
                  </div>
                  <div className="flex items-center gap-3">
                    {/* Timezone Selector */}
                    <div className="flex items-center gap-2">
                      <Globe size={14} className="text-zinc-500" />
                      <select
                        value={currentSchedule.timezone}
                        onChange={(e) => updateTimezone(e.target.value)}
                        className="px-2 py-1 text-xs bg-[#18181b] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                      >
                        {TIMEZONES.map(tz => (
                          <option key={tz.value} value={tz.value}>{tz.label}</option>
                        ))}
                      </select>
                    </div>

                    {/* Action Buttons */}
                    <button
                      onClick={handleSave}
                      className="px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold flex items-center gap-1"
                    >
                      <Save size={14} />
                      Save
                    </button>
                    <button
                      onClick={handleReset}
                      className="px-3 py-1.5 rounded bg-zinc-700 hover:bg-zinc-600 text-white text-xs font-semibold flex items-center gap-1"
                    >
                      <RotateCcw size={14} />
                      Reset
                    </button>
                  </div>
                </div>

                {/* Quick Presets */}
                <div className="mt-3 flex items-center gap-2">
                  <span className="text-xs text-zinc-500">Quick Presets:</span>
                  {Object.keys(PRESETS).map(preset => (
                    <button
                      key={preset}
                      onClick={() => applyPreset(preset as keyof typeof PRESETS)}
                      className="px-2 py-1 rounded bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 text-xs"
                    >
                      {preset}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex-1 flex gap-4 overflow-hidden">
                {/* Schedule Editor */}
                <div className="flex-1 bg-[#27272a] border border-[#3f3f46] rounded p-4 overflow-auto custom-scrollbar">
                  <h3 className="text-sm font-bold text-zinc-400 mb-3">Weekly Schedule</h3>

                  <div className="space-y-2">
                    {currentSchedule.schedule.map(daySchedule => (
                      <div key={daySchedule.day} className="border border-[#3f3f46] rounded p-3">
                        <div className="flex items-center justify-between mb-2">
                          <div className="text-xs font-bold text-zinc-300">{daySchedule.day}</div>
                          <button
                            onClick={() => {
                              setEditingDay(daySchedule.day);
                              setShowAddSession(true);
                            }}
                            className="text-xs text-blue-400 hover:text-blue-300"
                          >
                            + Add Session
                          </button>
                        </div>

                        <div className="space-y-1">
                          {daySchedule.sessions.map(session => (
                            <div
                              key={session.id}
                              className={`flex items-center justify-between p-2 rounded border ${SESSION_COLORS[session.type].bg} ${SESSION_COLORS[session.type].border}`}
                            >
                              <div className="flex items-center gap-2">
                                <div className={`text-xs font-semibold ${SESSION_COLORS[session.type].text}`}>
                                  {session.type.replace('_', ' ')}
                                </div>
                                <div className="text-xs text-zinc-400">
                                  {session.startTime} - {session.endTime}
                                </div>
                              </div>
                              <button
                                onClick={() => removeSession(daySchedule.day, session.id)}
                                className="p-1 hover:bg-zinc-700 rounded text-zinc-500 hover:text-rose-400"
                              >
                                <X size={12} />
                              </button>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Holiday Calendar */}
                <div className="w-80 bg-[#27272a] border border-[#3f3f46] rounded p-4 overflow-auto custom-scrollbar">
                  <div className="flex items-center justify-between mb-3">
                    <h3 className="text-sm font-bold text-zinc-400">Holiday Calendar</h3>
                    <button
                      onClick={() => setShowAddHoliday(true)}
                      className="p-1 rounded hover:bg-zinc-700 text-blue-400"
                    >
                      <Plus size={14} />
                    </button>
                  </div>

                  <div className="space-y-2">
                    {currentSchedule.holidays.length === 0 ? (
                      <div className="text-xs text-zinc-500 text-center py-4">No holidays configured</div>
                    ) : (
                      currentSchedule.holidays.map(holiday => (
                        <div
                          key={holiday.id}
                          className="flex items-start justify-between p-2 bg-[#18181b] border border-[#3f3f46] rounded"
                        >
                          <div>
                            <div className="text-xs font-semibold text-zinc-300">{holiday.date}</div>
                            <div className="text-xs text-zinc-500 mt-0.5">{holiday.description}</div>
                          </div>
                          <button
                            onClick={() => removeHoliday(holiday.id)}
                            className="p-1 hover:bg-zinc-700 rounded text-zinc-500 hover:text-rose-400"
                          >
                            <Trash2 size={12} />
                          </button>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-zinc-500">
              Select a symbol to configure market hours
            </div>
          )}
        </div>
      </div>

      {/* Add Session Modal */}
      {showAddSession && editingDay && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-[#27272a] border border-[#3f3f46] rounded-lg p-6 w-96">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-zinc-200">Add Session - {editingDay}</h3>
              <button
                onClick={() => setShowAddSession(false)}
                className="p-1 hover:bg-zinc-700 rounded text-zinc-400"
              >
                <X size={20} />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-xs text-zinc-500 mb-1">Session Type</label>
                <select
                  value={newSession.type}
                  onChange={(e) => setNewSession({ ...newSession, type: e.target.value as SessionType })}
                  className="w-full px-3 py-2 text-sm bg-[#18181b] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                >
                  <option value="PRE_MARKET">Pre-Market</option>
                  <option value="REGULAR">Regular</option>
                  <option value="AFTER_HOURS">After Hours</option>
                  <option value="CLOSED">Closed</option>
                </select>
              </div>

              <div>
                <label className="block text-xs text-zinc-500 mb-1">Start Time</label>
                <input
                  type="time"
                  value={newSession.startTime}
                  onChange={(e) => setNewSession({ ...newSession, startTime: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-[#18181b] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs text-zinc-500 mb-1">End Time</label>
                <input
                  type="time"
                  value={newSession.endTime}
                  onChange={(e) => setNewSession({ ...newSession, endTime: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-[#18181b] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  onClick={addSession}
                  className="flex-1 px-4 py-2 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-semibold"
                >
                  Add Session
                </button>
                <button
                  onClick={() => setShowAddSession(false)}
                  className="flex-1 px-4 py-2 rounded bg-zinc-700 hover:bg-zinc-600 text-white text-sm font-semibold"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Add Holiday Modal */}
      {showAddHoliday && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-[#27272a] border border-[#3f3f46] rounded-lg p-6 w-96">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-zinc-200">Add Holiday</h3>
              <button
                onClick={() => setShowAddHoliday(false)}
                className="p-1 hover:bg-zinc-700 rounded text-zinc-400"
              >
                <X size={20} />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-xs text-zinc-500 mb-1">Date</label>
                <input
                  type="date"
                  value={newHoliday.date}
                  onChange={(e) => setNewHoliday({ ...newHoliday, date: e.target.value })}
                  className="w-full px-3 py-2 text-sm bg-[#18181b] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs text-zinc-500 mb-1">Description</label>
                <input
                  type="text"
                  value={newHoliday.description}
                  onChange={(e) => setNewHoliday({ ...newHoliday, description: e.target.value })}
                  placeholder="e.g., Christmas Day"
                  className="w-full px-3 py-2 text-sm bg-[#18181b] border border-[#3f3f46] rounded text-zinc-300 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  onClick={addHoliday}
                  className="flex-1 px-4 py-2 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-semibold"
                >
                  Add Holiday
                </button>
                <button
                  onClick={() => setShowAddHoliday(false)}
                  className="flex-1 px-4 py-2 rounded bg-zinc-700 hover:bg-zinc-600 text-white text-sm font-semibold"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
