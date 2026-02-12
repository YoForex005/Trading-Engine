/**
 * Economic Calendar Component (MT5-style) - ENHANCED
 * Shows upcoming economic events with filters, alerts, real-time updates, and advanced features
 */

import React, { useState, useMemo, useEffect } from 'react';
import {
  Calendar,
  Filter,
  Bell,
  BellRing,
  X,
  ChevronDown,
  Info,
  Clock,
  TrendingUp,
  TrendingDown,
  Minus,
  LineChart,
  Star,
  ChevronRight
} from 'lucide-react';
import { API_ENDPOINTS } from '../config/api';
import { useAlertStore } from '../store/useAlertStore';

// ============================================================================
// Types
// ============================================================================

export interface EconomicEvent {
  id: string;
  time: string; // ISO date string
  currency: string;
  country: string;
  event: string;
  impact: 'High' | 'Medium' | 'Low';
  period: string;
  actual: string;
  forecast: string;
  previous: string;
  description?: string;
}

interface HistoricalOccurrence {
  date: string;
  actual: string;
  forecast: string;
  previous: string;
  pipMovement: number; // positive or negative
}

type DateRange = 'Today' | 'This Week' | 'Next Week' | 'Custom';
type ViewMode = 'day' | 'week';

interface EconomicCalendarProps {
  onClose?: () => void;
}

// ============================================================================
// Mock Data (Realistic Economic Events)
// ============================================================================

const generateMockEvents = (): EconomicEvent[] => {
  const now = new Date();
  const events: EconomicEvent[] = [];

  // Helper to create date
  const createDate = (daysOffset: number, hour: number, minute: number = 0) => {
    const d = new Date(now);
    d.setDate(d.getDate() + daysOffset);
    d.setHours(hour, minute, 0, 0);
    return d.toISOString();
  };

  // Today's events
  events.push(
    { id: '1', time: createDate(0, 8, 30), currency: 'GBP', country: 'United Kingdom', event: 'Claimant Count Change', impact: 'High', period: 'Jan', actual: '18.2K', forecast: '16.5K', previous: '16.0K', description: 'Measures the change in the number of unemployed people claiming benefits.' },
    { id: '2', time: createDate(0, 10, 0), currency: 'EUR', country: 'Germany', event: 'German PPI m/m', impact: 'Medium', period: 'Dec', actual: '', forecast: '0.2%', previous: '-0.1%', description: 'Producer Price Index measures change in prices of goods sold by manufacturers.' },
    { id: '3', time: createDate(0, 13, 30), currency: 'USD', country: 'United States', event: 'Retail Sales m/m', impact: 'High', period: 'Dec', actual: '', forecast: '0.4%', previous: '0.3%', description: 'Measures the change in total value of sales at retail level.' },
    { id: '4', time: createDate(0, 14, 30), currency: 'USD', country: 'United States', event: 'Core Retail Sales m/m', impact: 'High', period: 'Dec', actual: '', forecast: '0.3%', previous: '0.4%', description: 'Retail sales excluding automobiles.' },
    { id: '5', time: createDate(0, 15, 0), currency: 'USD', country: 'United States', event: 'Empire State Manufacturing Index', impact: 'Medium', period: 'Jan', actual: '', forecast: '-4.5', previous: '-14.5', description: 'Survey of New York manufacturers about current business conditions.' }
  );

  // Tomorrow
  events.push(
    { id: '6', time: createDate(1, 3, 0), currency: 'JPY', country: 'Japan', event: 'Trade Balance', impact: 'Medium', period: 'Dec', actual: '', forecast: '¥0.15T', previous: '¥-0.04T', description: 'Difference in value between imported and exported goods.' },
    { id: '7', time: createDate(1, 7, 0), currency: 'AUD', country: 'Australia', event: 'Employment Change', impact: 'High', period: 'Dec', actual: '', forecast: '15.0K', previous: '-10.1K', description: 'Change in number of employed people.' },
    { id: '8', time: createDate(1, 9, 30), currency: 'EUR', country: 'European Union', event: 'CPI y/y', impact: 'High', period: 'Dec', actual: '', forecast: '2.4%', previous: '2.3%', description: 'Consumer Price Index - measure of inflation.' },
    { id: '9', time: createDate(1, 13, 30), currency: 'USD', country: 'United States', event: 'Building Permits', impact: 'Medium', period: 'Dec', actual: '', forecast: '1.46M', previous: '1.50M', description: 'Measures the change in number of new building permits issued.' },
    { id: '10', time: createDate(1, 14, 15), currency: 'USD', country: 'United States', event: 'Fed Chair Powell Speaks', impact: 'High', period: '', actual: '', forecast: '', previous: '', description: 'Federal Reserve Chair Jerome Powell speech - can cause market volatility.' }
  );

  // Day after tomorrow
  events.push(
    { id: '11', time: createDate(2, 1, 30), currency: 'CNY', country: 'China', event: 'GDP q/y', impact: 'High', period: 'Q4', actual: '', forecast: '5.3%', previous: '4.9%', description: 'Gross Domestic Product - broadest measure of economic activity.' },
    { id: '12', time: createDate(2, 8, 30), currency: 'CAD', country: 'Canada', event: 'CPI m/m', impact: 'High', period: 'Dec', actual: '', forecast: '-0.3%', previous: '0.1%', description: 'Consumer Price Index month-over-month change.' },
    { id: '13', time: createDate(2, 10, 0), currency: 'EUR', country: 'Germany', event: 'ZEW Economic Sentiment', impact: 'Medium', period: 'Jan', actual: '', forecast: '21.0', previous: '23.0', description: 'Survey of institutional investors and analysts on economic outlook.' },
    { id: '14', time: createDate(2, 14, 30), currency: 'USD', country: 'United States', event: 'Unemployment Claims', impact: 'Medium', period: '', actual: '', forecast: '208K', previous: '202K', description: 'Number of individuals filing for unemployment insurance for the first time.' },
    { id: '15', time: createDate(2, 15, 0), currency: 'USD', country: 'United States', event: 'Philly Fed Manufacturing Index', impact: 'Medium', period: 'Jan', actual: '', forecast: '-5.0', previous: '-10.5', description: 'Survey of manufacturers in Philadelphia region.' }
  );

  // Future events
  events.push(
    { id: '16', time: createDate(3, 13, 30), currency: 'USD', country: 'United States', event: 'GDP q/q', impact: 'High', period: 'Q4', actual: '', forecast: '2.6%', previous: '3.1%', description: 'Gross Domestic Product quarterly change.' },
    { id: '17', time: createDate(4, 7, 0), currency: 'NZD', country: 'New Zealand', event: 'CPI q/q', impact: 'High', period: 'Q4', actual: '', forecast: '0.7%', previous: '0.6%', description: 'Consumer Price Index quarterly change.' },
    { id: '18', time: createDate(5, 12, 45), currency: 'EUR', country: 'European Union', event: 'ECB Press Conference', impact: 'High', period: '', actual: '', forecast: '', previous: '', description: 'European Central Bank President press conference following rate decision.' },
    { id: '19', time: createDate(6, 13, 30), currency: 'USD', country: 'United States', event: 'Non-Farm Payrolls', impact: 'High', period: 'Jan', actual: '', forecast: '180K', previous: '256K', description: 'Most important employment report - number of jobs added/lost.' },
    { id: '20', time: createDate(6, 13, 30), currency: 'USD', country: 'United States', event: 'Unemployment Rate', impact: 'High', period: 'Jan', actual: '', forecast: '4.2%', previous: '4.1%', description: 'Percentage of total workforce that is unemployed and actively seeking employment.' }
  );

  return events;
};

// Mock historical data
const getMockHistoricalData = (eventName: string): HistoricalOccurrence[] => {
  // Return last 3 occurrences with simulated data
  return [
    {
      date: '2026-01-03',
      actual: eventName.includes('NFP') ? '256K' : eventName.includes('CPI') ? '2.3%' : '0.3%',
      forecast: eventName.includes('NFP') ? '200K' : eventName.includes('CPI') ? '2.2%' : '0.2%',
      previous: eventName.includes('NFP') ? '227K' : eventName.includes('CPI') ? '2.1%' : '0.1%',
      pipMovement: Math.floor(Math.random() * 60) - 30, // -30 to +30 pips
    },
    {
      date: '2025-12-06',
      actual: eventName.includes('NFP') ? '227K' : eventName.includes('CPI') ? '2.1%' : '0.2%',
      forecast: eventName.includes('NFP') ? '183K' : eventName.includes('CPI') ? '2.0%' : '0.1%',
      previous: eventName.includes('NFP') ? '150K' : eventName.includes('CPI') ? '1.9%' : '0.0%',
      pipMovement: Math.floor(Math.random() * 60) - 30,
    },
    {
      date: '2025-11-01',
      actual: eventName.includes('NFP') ? '150K' : eventName.includes('CPI') ? '1.9%' : '0.1%',
      forecast: eventName.includes('NFP') ? '140K' : eventName.includes('CPI') ? '1.8%' : '0.0%',
      previous: eventName.includes('NFP') ? '254K' : eventName.includes('CPI') ? '1.7%' : '-0.1%',
      pipMovement: Math.floor(Math.random() * 60) - 30,
    },
  ];
};

// ============================================================================
// Main Component
// ============================================================================

export const EconomicCalendar: React.FC<EconomicCalendarProps> = ({ onClose }) => {
  const [events] = useState<EconomicEvent[]>(generateMockEvents());
  const [dateRange, setDateRange] = useState<DateRange>('This Week');
  const [selectedCurrencies, setSelectedCurrencies] = useState<string[]>([]);
  const [selectedImpacts, setSelectedImpacts] = useState<string[]>(['High', 'Medium', 'Low']);
  const [detailEvent, setDetailEvent] = useState<EconomicEvent | null>(null);
  const [customDateFrom, setCustomDateFrom] = useState('');
  const [customDateTo, setCustomDateTo] = useState('');
  const [currentTime, setCurrentTime] = useState(Date.now()); // For live countdown
  const [viewMode, setViewMode] = useState<ViewMode>('day');
  const [favorites, setFavorites] = useState<string[]>(() => {
    const stored = localStorage.getItem('rtx5-event-favorites');
    return stored ? JSON.parse(stored) : [];
  });
  const [showFavoritesOnly, setShowFavoritesOnly] = useState(false);
  const [expandedEventId, setExpandedEventId] = useState<string | null>(null);

  const { alerts, addAlert } = useAlertStore();
  const currencies = ['USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', 'NZD', 'CHF', 'CNY'];

  // Live countdown timer - updates every second
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(Date.now());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  // Persist favorites to localStorage
  useEffect(() => {
    localStorage.setItem('rtx5-event-favorites', JSON.stringify(favorites));
  }, [favorites]);

  // Filter events by date range
  const filteredByDate = useMemo(() => {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

    return events.filter(event => {
      const eventDate = new Date(event.time);

      if (dateRange === 'Today') {
        const eventDay = new Date(eventDate.getFullYear(), eventDate.getMonth(), eventDate.getDate());
        return eventDay.getTime() === today.getTime();
      } else if (dateRange === 'This Week') {
        const weekEnd = new Date(today);
        weekEnd.setDate(weekEnd.getDate() + 7);
        return eventDate >= today && eventDate <= weekEnd;
      } else if (dateRange === 'Next Week') {
        const weekStart = new Date(today);
        weekStart.setDate(weekStart.getDate() + 7);
        const weekEnd = new Date(weekStart);
        weekEnd.setDate(weekEnd.getDate() + 7);
        return eventDate >= weekStart && eventDate <= weekEnd;
      } else if (dateRange === 'Custom' && customDateFrom && customDateTo) {
        const from = new Date(customDateFrom);
        const to = new Date(customDateTo);
        to.setHours(23, 59, 59);
        return eventDate >= from && eventDate <= to;
      }
      return true;
    });
  }, [events, dateRange, customDateFrom, customDateTo]);

  // Filter by currency, impact, and favorites
  const filteredEvents = useMemo(() => {
    return filteredByDate.filter(event => {
      const currencyMatch = selectedCurrencies.length === 0 || selectedCurrencies.includes(event.currency);
      const impactMatch = selectedImpacts.includes(event.impact);
      const favoriteMatch = !showFavoritesOnly || favorites.includes(event.event);
      return currencyMatch && impactMatch && favoriteMatch;
    });
  }, [filteredByDate, selectedCurrencies, selectedImpacts, showFavoritesOnly, favorites]);

  // Group by date for day view
  const groupedEvents = useMemo(() => {
    const groups: Record<string, EconomicEvent[]> = {};
    filteredEvents.forEach(event => {
      const date = new Date(event.time);
      const dateKey = date.toLocaleDateString('en-US', {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
        year: 'numeric'
      });
      if (!groups[dateKey]) groups[dateKey] = [];
      groups[dateKey].push(event);
    });

    return Object.entries(groups).sort((a, b) => {
      return new Date(a[1][0].time).getTime() - new Date(b[1][0].time).getTime();
    });
  }, [filteredEvents]);

  // Group by weekday for week view
  const weekViewData = useMemo(() => {
    const weekdays = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday'];
    const grouped: Record<string, EconomicEvent[]> = {
      Monday: [],
      Tuesday: [],
      Wednesday: [],
      Thursday: [],
      Friday: []
    };

    filteredEvents.forEach(event => {
      const date = new Date(event.time);
      const dayName = date.toLocaleDateString('en-US', { weekday: 'long' });
      if (grouped[dayName]) {
        grouped[dayName].push(event);
      }
    });

    return weekdays.map(day => ({
      day,
      events: grouped[day].sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime())
    }));
  }, [filteredEvents]);

  // Find next event with countdown
  const nextEvent = useMemo(() => {
    const now = new Date();
    const upcoming = filteredEvents
      .filter(e => new Date(e.time) > now)
      .sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime());
    return upcoming[0] || null;
  }, [filteredEvents]);

  // Get countdown string (HH:MM:SS)
  const getCountdown = (eventTime: string): { text: string; isLive: boolean; isUrgent: boolean } => {
    const now = currentTime;
    const event = new Date(eventTime);
    const diff = event.getTime() - now;

    // Event is happening now (within ±15 minutes window)
    if (Math.abs(diff) <= 15 * 60 * 1000) {
      return { text: 'LIVE', isLive: true, isUrgent: false };
    }

    // Event has passed
    if (diff < 0) {
      return { text: '-', isLive: false, isUrgent: false };
    }

    // Event is within 24 hours
    if (diff <= 24 * 60 * 60 * 1000) {
      const hours = Math.floor(diff / (1000 * 60 * 60));
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
      const seconds = Math.floor((diff % (1000 * 60)) / 1000);
      const isUrgent = diff <= 5 * 60 * 1000; // Less than 5 minutes
      return {
        text: `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`,
        isLive: false,
        isUrgent
      };
    }

    // More than 24 hours away
    return { text: '-', isLive: false, isUrgent: false };
  };

  // Check if event is within 30 minutes
  const isUpcoming = (eventTime: string) => {
    const now = new Date();
    const event = new Date(eventTime);
    const diff = event.getTime() - now.getTime();
    return diff > 0 && diff <= 30 * 60 * 1000; // 30 minutes
  };

  // Get impact color
  const getImpactColor = (impact: string) => {
    switch (impact) {
      case 'High': return 'bg-red-500';
      case 'Medium': return 'bg-orange-500';
      case 'Low': return 'bg-yellow-500';
      default: return 'bg-gray-500';
    }
  };

  // Get difference indicator
  const getDifferenceIndicator = (actual: string, forecast: string) => {
    if (!actual || !forecast) return null;

    const actualNum = parseFloat(actual.replace(/[^0-9.-]/g, ''));
    const forecastNum = parseFloat(forecast.replace(/[^0-9.-]/g, ''));

    if (isNaN(actualNum) || isNaN(forecastNum)) return null;

    if (actualNum > forecastNum) {
      return <TrendingUp size={14} className="text-emerald-400" />;
    } else if (actualNum < forecastNum) {
      return <TrendingDown size={14} className="text-red-400" />;
    }
    return <Minus size={14} className="text-zinc-500" />;
  };

  // Toggle currency filter
  const toggleCurrency = (currency: string) => {
    setSelectedCurrencies(prev =>
      prev.includes(currency)
        ? prev.filter(c => c !== currency)
        : [...prev, currency]
    );
  };

  // Toggle impact filter
  const toggleImpact = (impact: string) => {
    setSelectedImpacts(prev =>
      prev.includes(impact)
        ? prev.filter(i => i !== impact)
        : [...prev, impact]
    );
  };

  // Toggle favorite
  const toggleFavorite = (eventName: string) => {
    setFavorites(prev =>
      prev.includes(eventName)
        ? prev.filter(e => e !== eventName)
        : [...prev, eventName]
    );
  };

  // Check if event has alert
  const hasAlert = (event: EconomicEvent): boolean => {
    return alerts.some(alert =>
      alert.symbol === event.currency &&
      alert.status === 'active' &&
      alert.enabled
    );
  };

  // Handle set alert (5 minutes before event)
  const handleSetAlert = (event: EconomicEvent) => {
    const eventTime = new Date(event.time);
    const alertTime = new Date(eventTime.getTime() - 5 * 60 * 1000); // 5 minutes before

    // Create alert via useAlertStore
    addAlert({
      symbol: event.currency,
      condition: 'price_above', // Dummy condition
      price: 0, // Event-based alert
      expiration: eventTime.getTime(),
      notifyMethod: 'both',
      message: `Economic Event: ${event.event} in 5 minutes`
    });

    // Also dispatch custom event for chart integration if needed
    window.dispatchEvent(new CustomEvent('set-event-alert', {
      detail: {
        symbol: event.currency,
        event: event.event,
        time: event.time
      }
    }));
  };

  // Handle show on chart
  const handleShowOnChart = (event: EconomicEvent) => {
    window.dispatchEvent(new CustomEvent('addChartMarker', {
      detail: {
        time: event.time,
        name: event.event,
        currency: event.currency,
        impact: event.impact
      }
    }));
  };

  return (
    <div className="flex flex-col h-full bg-[#1e1e1e] text-zinc-300">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-zinc-700 bg-[#252526]">
        <div className="flex items-center gap-3">
          <Calendar size={18} className="text-blue-500" />
          <h2 className="text-sm font-semibold text-zinc-200">Economic Calendar</h2>
          {nextEvent && (
            <div className="flex items-center gap-2 ml-4 text-xs text-zinc-400">
              <Clock size={12} />
              <span>Next: {nextEvent.event} in {Math.floor((new Date(nextEvent.time).getTime() - new Date().getTime()) / 60000)}m</span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-2">
          {/* View Mode Toggle */}
          <div className="flex items-center gap-1 bg-zinc-800 rounded p-0.5">
            <button
              onClick={() => setViewMode('day')}
              className={`px-2 py-1 text-xs rounded transition-colors ${
                viewMode === 'day' ? 'bg-blue-600 text-white' : 'text-zinc-400 hover:text-white'
              }`}
            >
              Day
            </button>
            <button
              onClick={() => setViewMode('week')}
              className={`px-2 py-1 text-xs rounded transition-colors ${
                viewMode === 'week' ? 'bg-blue-600 text-white' : 'text-zinc-400 hover:text-white'
              }`}
            >
              Week
            </button>
          </div>
          {/* Favorites Toggle */}
          <button
            onClick={() => setShowFavoritesOnly(!showFavoritesOnly)}
            className={`flex items-center gap-1 px-2 py-1 text-xs rounded transition-colors ${
              showFavoritesOnly ? 'bg-yellow-600 text-white' : 'bg-zinc-800 text-zinc-400 hover:text-white'
            }`}
            title="Show favorites only"
          >
            <Star size={12} fill={showFavoritesOnly ? 'currentColor' : 'none'} />
            {favorites.length > 0 && <span>({favorites.length})</span>}
          </button>
          {onClose && (
            <button
              onClick={onClose}
              className="text-zinc-400 hover:text-white transition-colors p-1 hover:bg-zinc-700/50 rounded"
            >
              <X size={16} />
            </button>
          )}
        </div>
      </div>

      {/* Date Range Selector */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-zinc-700 bg-[#1e1e1e]">
        <span className="text-xs text-zinc-500 mr-2">Range:</span>
        {(['Today', 'This Week', 'Next Week', 'Custom'] as DateRange[]).map((range) => (
          <button
            key={range}
            onClick={() => setDateRange(range)}
            className={`px-3 py-1 text-xs rounded transition-colors ${
              dateRange === range
                ? 'bg-blue-600 text-white'
                : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700 hover:text-white'
            }`}
          >
            {range}
          </button>
        ))}
        {dateRange === 'Custom' && (
          <div className="flex items-center gap-2 ml-2">
            <input
              type="date"
              value={customDateFrom}
              onChange={(e) => setCustomDateFrom(e.target.value)}
              className="bg-[#252526] border border-zinc-600 rounded px-2 py-1 text-xs text-white focus:outline-none focus:border-blue-500"
            />
            <span className="text-zinc-600">to</span>
            <input
              type="date"
              value={customDateTo}
              onChange={(e) => setCustomDateTo(e.target.value)}
              className="bg-[#252526] border border-zinc-600 rounded px-2 py-1 text-xs text-white focus:outline-none focus:border-blue-500"
            />
          </div>
        )}
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-2 px-4 py-2 border-b border-zinc-700 bg-[#252526]">
        <Filter size={14} className="text-zinc-500" />

        {/* Currency Filters */}
        <div className="flex flex-wrap gap-1">
          {currencies.map((currency) => (
            <button
              key={currency}
              onClick={() => toggleCurrency(currency)}
              className={`px-2 py-1 text-xs rounded transition-colors ${
                selectedCurrencies.includes(currency)
                  ? 'bg-blue-600 text-white'
                  : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
              }`}
            >
              {currency}
            </button>
          ))}
        </div>

        <div className="h-4 w-px bg-zinc-700 mx-2" />

        {/* Impact Filters */}
        <div className="flex gap-1">
          {['High', 'Medium', 'Low'].map((impact) => (
            <button
              key={impact}
              onClick={() => toggleImpact(impact)}
              className={`flex items-center gap-1 px-2 py-1 text-xs rounded transition-colors ${
                selectedImpacts.includes(impact)
                  ? 'bg-zinc-700 text-white'
                  : 'bg-zinc-800 text-zinc-500 hover:bg-zinc-700'
              }`}
            >
              <div className={`w-2 h-2 rounded-full ${getImpactColor(impact)}`} />
              {impact}
            </button>
          ))}
        </div>

        {selectedCurrencies.length > 0 && (
          <button
            onClick={() => setSelectedCurrencies([])}
            className="ml-auto text-xs text-zinc-500 hover:text-white"
          >
            Clear filters
          </button>
        )}
      </div>

      {/* Content Area */}
      <div className="flex-1 overflow-auto">
        {viewMode === 'day' ? (
          <DayView
            groupedEvents={groupedEvents}
            isUpcoming={isUpcoming}
            getCountdown={getCountdown}
            getImpactColor={getImpactColor}
            getDifferenceIndicator={getDifferenceIndicator}
            favorites={favorites}
            toggleFavorite={toggleFavorite}
            hasAlert={hasAlert}
            handleSetAlert={handleSetAlert}
            handleShowOnChart={handleShowOnChart}
            setDetailEvent={setDetailEvent}
            expandedEventId={expandedEventId}
            setExpandedEventId={setExpandedEventId}
          />
        ) : (
          <WeekView
            weekViewData={weekViewData}
            getImpactColor={getImpactColor}
            favorites={favorites}
            toggleFavorite={toggleFavorite}
            setDetailEvent={setDetailEvent}
          />
        )}
      </div>

      {/* Event Detail Popup */}
      {detailEvent && (
        <EventDetailPopup
          event={detailEvent}
          onClose={() => setDetailEvent(null)}
          onSetAlert={handleSetAlert}
          onShowOnChart={handleShowOnChart}
          hasAlert={hasAlert(detailEvent)}
          isFavorite={favorites.includes(detailEvent.event)}
          toggleFavorite={() => toggleFavorite(detailEvent.event)}
        />
      )}
    </div>
  );
};

// ============================================================================
// Day View Component
// ============================================================================

interface DayViewProps {
  groupedEvents: [string, EconomicEvent[]][];
  isUpcoming: (time: string) => boolean;
  getCountdown: (time: string) => { text: string; isLive: boolean; isUrgent: boolean };
  getImpactColor: (impact: string) => string;
  getDifferenceIndicator: (actual: string, forecast: string) => React.ReactNode;
  favorites: string[];
  toggleFavorite: (eventName: string) => void;
  hasAlert: (event: EconomicEvent) => boolean;
  handleSetAlert: (event: EconomicEvent) => void;
  handleShowOnChart: (event: EconomicEvent) => void;
  setDetailEvent: (event: EconomicEvent) => void;
  expandedEventId: string | null;
  setExpandedEventId: (id: string | null) => void;
}

const DayView: React.FC<DayViewProps> = ({
  groupedEvents,
  isUpcoming,
  getCountdown,
  getImpactColor,
  getDifferenceIndicator,
  favorites,
  toggleFavorite,
  hasAlert,
  handleSetAlert,
  handleShowOnChart,
  setDetailEvent,
  expandedEventId,
  setExpandedEventId
}) => {
  return (
    <table className="w-full text-xs border-collapse">
      <thead className="sticky top-0 bg-[#2d3436] text-zinc-400 z-10">
        <tr>
          <th className="px-3 py-2 text-left font-medium w-16">Time</th>
          <th className="px-3 py-2 text-left font-medium w-20">Countdown</th>
          <th className="px-3 py-2 text-left font-medium w-12">Cur.</th>
          <th className="px-3 py-2 text-center font-medium w-20">Impact</th>
          <th className="px-3 py-2 text-left font-medium">Event</th>
          <th className="px-3 py-2 text-center font-medium w-16">Period</th>
          <th className="px-3 py-2 text-right font-medium w-20">Actual</th>
          <th className="px-3 py-2 text-right font-medium w-20">Forecast</th>
          <th className="px-3 py-2 text-right font-medium w-20">Previous</th>
          <th className="px-3 py-2 text-center font-medium w-24">Actions</th>
        </tr>
      </thead>
      <tbody>
        {groupedEvents.length === 0 ? (
          <tr>
            <td colSpan={10} className="text-center py-12 text-zinc-500">
              No events found for selected filters
            </td>
          </tr>
        ) : (
          groupedEvents.map(([dateLabel, events]) => (
            <React.Fragment key={dateLabel}>
              {/* Date Header */}
              <tr className="bg-[#262626] border-y border-zinc-800">
                <td colSpan={10} className="px-3 py-1.5 font-medium text-blue-400 text-xs">
                  {dateLabel}
                </td>
              </tr>
              {/* Events */}
              {events.map((event) => {
                const countdown = getCountdown(event.time);
                const isExpanded = expandedEventId === event.id;
                const historicalData = isExpanded ? getMockHistoricalData(event.event) : [];

                return (
                  <React.Fragment key={event.id}>
                    <tr
                      className={`hover:bg-zinc-800/50 transition-colors border-b border-zinc-800/50 ${
                        isUpcoming(event.time) ? 'bg-yellow-500/10 border-yellow-500/30' : ''
                      } ${countdown.isLive ? 'bg-red-500/10 border-red-500/30' : ''}`}
                    >
                      <td className="px-3 py-2 font-mono text-zinc-400">
                        {new Date(event.time).toLocaleTimeString('en-US', {
                          hour: '2-digit',
                          minute: '2-digit',
                          hour12: false
                        })}
                      </td>
                      <td className="px-3 py-2 font-mono">
                        {countdown.isLive ? (
                          <span className="bg-red-600 text-white px-2 py-0.5 rounded font-bold text-[10px] animate-pulse">
                            LIVE
                          </span>
                        ) : countdown.isUrgent ? (
                          <span className="text-red-400 font-bold animate-pulse">
                            {countdown.text}
                          </span>
                        ) : countdown.text !== '-' ? (
                          <span className="text-yellow-400">{countdown.text}</span>
                        ) : (
                          <span className="text-zinc-600">—</span>
                        )}
                      </td>
                      <td className="px-3 py-2 font-bold text-zinc-200">{event.currency}</td>
                      <td className="px-3 py-2 text-center">
                        <ImpactMeter impact={event.impact} getImpactColor={getImpactColor} />
                      </td>
                      <td className="px-3 py-2 font-medium">
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => setExpandedEventId(isExpanded ? null : event.id)}
                            className="text-zinc-500 hover:text-blue-400 transition-colors"
                          >
                            <ChevronRight
                              size={14}
                              className={`transform transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                            />
                          </button>
                          <span
                            className="cursor-pointer hover:text-blue-400"
                            onClick={() => setDetailEvent(event)}
                          >
                            {event.event}
                          </span>
                        </div>
                      </td>
                      <td className="px-3 py-2 text-center text-zinc-500">{event.period}</td>
                      <td className={`px-3 py-2 text-right font-mono font-semibold ${
                        event.actual ? 'text-white' : 'text-zinc-600'
                      }`}>
                        <div className="flex items-center justify-end gap-1">
                          {event.actual || '-'}
                          {getDifferenceIndicator(event.actual, event.forecast)}
                        </div>
                      </td>
                      <td className="px-3 py-2 text-right font-mono text-zinc-400">
                        {event.forecast || '-'}
                      </td>
                      <td className="px-3 py-2 text-right font-mono text-zinc-500">
                        {event.previous || '-'}
                      </td>
                      <td className="px-3 py-2">
                        <div className="flex items-center justify-center gap-1">
                          {/* Favorite */}
                          <button
                            onClick={() => toggleFavorite(event.event)}
                            className="p-1 text-zinc-500 hover:text-yellow-500 hover:bg-zinc-700/50 rounded transition-colors"
                            title="Add to favorites"
                          >
                            <Star size={12} fill={favorites.includes(event.event) ? 'currentColor' : 'none'} />
                          </button>
                          {/* Alert */}
                          <button
                            onClick={() => handleSetAlert(event)}
                            className={`p-1 transition-colors rounded ${
                              hasAlert(event)
                                ? 'text-yellow-500 bg-yellow-500/10'
                                : 'text-zinc-500 hover:text-yellow-500 hover:bg-zinc-700/50'
                            }`}
                            title="Set alert (5 min before)"
                          >
                            {hasAlert(event) ? <BellRing size={12} /> : <Bell size={12} />}
                          </button>
                          {/* Show on Chart */}
                          <button
                            onClick={() => handleShowOnChart(event)}
                            className="p-1 text-zinc-500 hover:text-blue-500 hover:bg-zinc-700/50 rounded transition-colors"
                            title="Show on chart"
                          >
                            <LineChart size={12} />
                          </button>
                        </div>
                      </td>
                    </tr>
                    {/* Historical Data Expansion */}
                    {isExpanded && (
                      <tr className="bg-zinc-900/50 border-b border-zinc-800">
                        <td colSpan={10} className="px-8 py-3">
                          <div className="text-xs">
                            <div className="text-zinc-400 font-semibold mb-2">Historical Impact (Last 3 Occurrences)</div>
                            <div className="grid grid-cols-3 gap-3">
                              {historicalData.map((occ, idx) => (
                                <div key={idx} className="bg-[#1e1e1e] rounded p-2 border border-zinc-800">
                                  <div className="text-zinc-500 text-[10px] mb-1">{occ.date}</div>
                                  <div className="space-y-0.5">
                                    <div className="flex justify-between">
                                      <span className="text-zinc-500">Actual:</span>
                                      <span className="text-white font-mono">{occ.actual}</span>
                                    </div>
                                    <div className="flex justify-between">
                                      <span className="text-zinc-500">Forecast:</span>
                                      <span className="text-zinc-400 font-mono">{occ.forecast}</span>
                                    </div>
                                    <div className="flex justify-between">
                                      <span className="text-zinc-500">Previous:</span>
                                      <span className="text-zinc-500 font-mono">{occ.previous}</span>
                                    </div>
                                    <div className="flex justify-between border-t border-zinc-700 pt-0.5 mt-0.5">
                                      <span className="text-zinc-500">Pip Move:</span>
                                      <span className={`font-mono font-bold ${
                                        occ.pipMovement > 0 ? 'text-emerald-400' : occ.pipMovement < 0 ? 'text-red-400' : 'text-zinc-500'
                                      }`}>
                                        {occ.pipMovement > 0 ? '+' : ''}{occ.pipMovement} pips
                                      </span>
                                    </div>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })}
            </React.Fragment>
          ))
        )}
      </tbody>
    </table>
  );
};

// ============================================================================
// Week View Component
// ============================================================================

interface WeekViewProps {
  weekViewData: { day: string; events: EconomicEvent[] }[];
  getImpactColor: (impact: string) => string;
  favorites: string[];
  toggleFavorite: (eventName: string) => void;
  setDetailEvent: (event: EconomicEvent) => void;
}

const WeekView: React.FC<WeekViewProps> = ({
  weekViewData,
  getImpactColor,
  favorites,
  toggleFavorite,
  setDetailEvent
}) => {
  return (
    <div className="grid grid-cols-5 gap-2 p-3 h-full">
      {weekViewData.map(({ day, events }) => (
        <div key={day} className="bg-[#252526] rounded border border-zinc-700 flex flex-col">
          <div className="px-3 py-2 border-b border-zinc-700 text-center font-semibold text-zinc-300 text-xs">
            {day}
          </div>
          <div className="flex-1 overflow-auto p-2 space-y-1">
            {events.length === 0 ? (
              <div className="text-center text-zinc-600 text-[10px] mt-4">No events</div>
            ) : (
              events.map((event) => (
                <div
                  key={event.id}
                  className={`rounded p-1.5 cursor-pointer hover:bg-zinc-700/50 transition-colors border ${
                    event.impact === 'High' ? 'border-red-500/30 bg-red-500/5' :
                    event.impact === 'Medium' ? 'border-orange-500/30 bg-orange-500/5' :
                    'border-yellow-500/30 bg-yellow-500/5'
                  }`}
                  onClick={() => setDetailEvent(event)}
                >
                  <div className="flex items-center justify-between mb-0.5">
                    <span className="font-mono text-[10px] text-zinc-400">
                      {new Date(event.time).toLocaleTimeString('en-US', {
                        hour: '2-digit',
                        minute: '2-digit',
                        hour12: false
                      })}
                    </span>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        toggleFavorite(event.event);
                      }}
                      className="text-zinc-600 hover:text-yellow-500"
                    >
                      <Star size={10} fill={favorites.includes(event.event) ? 'currentColor' : 'none'} />
                    </button>
                  </div>
                  <div className="flex items-start gap-1">
                    <div className={`w-1.5 h-1.5 rounded-full mt-1 ${getImpactColor(event.impact)}`} />
                    <div className="flex-1 min-w-0">
                      <div className="text-[10px] font-medium text-zinc-200 truncate" title={event.event}>
                        {event.event}
                      </div>
                      <div className="text-[9px] text-zinc-500">{event.currency}</div>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      ))}
    </div>
  );
};

// ============================================================================
// Impact Meter Component
// ============================================================================

interface ImpactMeterProps {
  impact: 'High' | 'Medium' | 'Low';
  getImpactColor: (impact: string) => string;
}

const ImpactMeter: React.FC<ImpactMeterProps> = ({ impact, getImpactColor }) => {
  const bars = impact === 'High' ? 3 : impact === 'Medium' ? 2 : 1;
  const historicalVolatility = impact === 'High' ? '80-150 pips' : impact === 'Medium' ? '30-80 pips' : '10-30 pips';

  return (
    <div className="flex items-center justify-center gap-0.5 group relative">
      {[1, 2, 3].map((i) => (
        <div
          key={i}
          className={`w-1 h-3 rounded-sm ${
            i <= bars ? getImpactColor(impact) : 'bg-zinc-700'
          }`}
        />
      ))}
      {/* Tooltip */}
      <div className="absolute bottom-full mb-2 hidden group-hover:block z-50 pointer-events-none">
        <div className="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-[10px] text-zinc-300 whitespace-nowrap shadow-lg">
          Historical Volatility: {historicalVolatility}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Event Detail Popup
// ============================================================================

interface EventDetailPopupProps {
  event: EconomicEvent;
  onClose: () => void;
  onSetAlert: (event: EconomicEvent) => void;
  onShowOnChart: (event: EconomicEvent) => void;
  hasAlert: boolean;
  isFavorite: boolean;
  toggleFavorite: () => void;
}

const EventDetailPopup: React.FC<EventDetailPopupProps> = ({
  event,
  onClose,
  onSetAlert,
  onShowOnChart,
  hasAlert,
  isFavorite,
  toggleFavorite
}) => {
  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-[#1e1e1e] border border-zinc-700 rounded-lg shadow-2xl w-[500px] animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between px-4 py-3 border-b border-zinc-700">
          <div className="flex items-center gap-2">
            <Info size={16} className="text-blue-500" />
            <h3 className="text-sm font-semibold text-zinc-200">Event Details</h3>
          </div>
          <button
            onClick={onClose}
            className="text-zinc-400 hover:text-white transition-colors"
          >
            <X size={16} />
          </button>
        </div>

        <div className="p-4 space-y-4">
          <div>
            <h4 className="text-base font-semibold text-white mb-2">{event.event}</h4>
            <div className="flex items-center gap-3 text-xs text-zinc-400">
              <span className="font-bold text-zinc-200">{event.currency}</span>
              <span>•</span>
              <span>{event.country}</span>
              <span>•</span>
              <span>{new Date(event.time).toLocaleString()}</span>
            </div>
          </div>

          {event.description && (
            <div className="text-xs text-zinc-400 bg-[#252526] rounded p-3">
              {event.description}
            </div>
          )}

          <div className="grid grid-cols-3 gap-4 text-xs">
            <div className="bg-[#252526] rounded p-3">
              <div className="text-zinc-500 mb-1">Actual</div>
              <div className={`text-lg font-bold font-mono ${event.actual ? 'text-white' : 'text-zinc-600'}`}>
                {event.actual || '-'}
              </div>
            </div>
            <div className="bg-[#252526] rounded p-3">
              <div className="text-zinc-500 mb-1">Forecast</div>
              <div className="text-lg font-bold font-mono text-zinc-400">
                {event.forecast || '-'}
              </div>
            </div>
            <div className="bg-[#252526] rounded p-3">
              <div className="text-zinc-500 mb-1">Previous</div>
              <div className="text-lg font-bold font-mono text-zinc-500">
                {event.previous || '-'}
              </div>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-between gap-2 px-4 py-3 border-t border-zinc-700 bg-[#252526]">
          <button
            onClick={toggleFavorite}
            className={`flex items-center gap-2 px-3 py-2 rounded transition-colors text-sm ${
              isFavorite
                ? 'bg-yellow-600 hover:bg-yellow-700 text-white'
                : 'bg-zinc-700 hover:bg-zinc-600 text-zinc-300'
            }`}
          >
            <Star size={14} fill={isFavorite ? 'currentColor' : 'none'} />
            {isFavorite ? 'Favorited' : 'Favorite'}
          </button>
          <div className="flex gap-2">
            <button
              onClick={() => {
                onShowOnChart(event);
                onClose();
              }}
              className="flex items-center gap-2 px-3 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded transition-colors"
            >
              <LineChart size={14} />
              Show on Chart
            </button>
            <button
              onClick={() => {
                onSetAlert(event);
                onClose();
              }}
              className={`flex items-center gap-2 px-3 py-2 text-sm rounded transition-colors ${
                hasAlert
                  ? 'bg-yellow-700 text-white'
                  : 'bg-yellow-600 hover:bg-yellow-700 text-white'
              }`}
            >
              {hasAlert ? <BellRing size={14} /> : <Bell size={14} />}
              {hasAlert ? 'Alert Set' : 'Set Alert'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
