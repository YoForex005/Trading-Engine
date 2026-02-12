/**
 * NewsImpactAnalyzer Component
 * Shows how economic news events affect symbol prices historically
 */

import {
  Calendar,
  TrendingUp,
  Filter,
  Star,
  Clock,
  BarChart3,
  AlertCircle,
  ChevronRight,
} from 'lucide-react';
import { useNewsImpactStore } from '../store/useNewsImpactStore';

type EventCategory = 'Employment' | 'Inflation' | 'Interest Rates' | 'GDP' | 'PMI' | 'Trade Balance';
type ImpactLevel = 'Low' | 'Medium' | 'High';

interface EconomicEvent {
  id: string;
  type: string;
  category: EventCategory;
  time: string;
  impact: ImpactLevel;
  currency: string;
  affectedSymbols: string[];
  forecast?: string;
  previous?: string;
  actual?: string;
}

interface HistoricalOccurrence {
  date: string;
  forecast: string;
  actual: string;
  previous: string;
  deviation: number; // Actual - Forecast
  priceData: {
    time: number; // Minutes relative to event (-30 to +30)
    open: number;
    high: number;
    low: number;
    close: number;
  }[];
  pipMovement: number;
}

interface EventTypeData {
  type: string;
  category: EventCategory;
  avgPipMovement: number;
  occurrences: HistoricalOccurrence[];
}

// Generate mock upcoming events (next 24h)
function generateUpcomingEvents(): EconomicEvent[] {
  const now = new Date();
  const events: EconomicEvent[] = [
    {
      id: '1',
      type: 'Non-Farm Payrolls (NFP)',
      category: 'Employment',
      time: new Date(now.getTime() + 2 * 60 * 60 * 1000).toISOString(),
      impact: 'High',
      currency: 'USD',
      affectedSymbols: ['EURUSD', 'GBPUSD', 'USDJPY'],
      forecast: '200K',
      previous: '187K',
    },
    {
      id: '2',
      type: 'Consumer Price Index (CPI)',
      category: 'Inflation',
      time: new Date(now.getTime() + 6 * 60 * 60 * 1000).toISOString(),
      impact: 'High',
      currency: 'USD',
      affectedSymbols: ['EURUSD', 'XAUUSD', 'BTCUSD'],
      forecast: '3.2%',
      previous: '3.1%',
    },
    {
      id: '3',
      type: 'FOMC Meeting',
      category: 'Interest Rates',
      time: new Date(now.getTime() + 10 * 60 * 60 * 1000).toISOString(),
      impact: 'High',
      currency: 'USD',
      affectedSymbols: ['EURUSD', 'GBPUSD', 'USDJPY', 'XAUUSD'],
      forecast: '5.25%',
      previous: '5.25%',
    },
    {
      id: '4',
      type: 'GDP Growth Rate',
      category: 'GDP',
      time: new Date(now.getTime() + 14 * 60 * 60 * 1000).toISOString(),
      impact: 'Medium',
      currency: 'EUR',
      affectedSymbols: ['EURUSD', 'EURGBP'],
      forecast: '0.4%',
      previous: '0.1%',
    },
    {
      id: '5',
      type: 'Manufacturing PMI',
      category: 'PMI',
      time: new Date(now.getTime() + 18 * 60 * 60 * 1000).toISOString(),
      impact: 'Medium',
      currency: 'GBP',
      affectedSymbols: ['GBPUSD', 'EURGBP'],
      forecast: '48.5',
      previous: '47.9',
    },
    {
      id: '6',
      type: 'Trade Balance',
      category: 'Trade Balance',
      time: new Date(now.getTime() + 22 * 60 * 60 * 1000).toISOString(),
      impact: 'Low',
      currency: 'USD',
      affectedSymbols: ['USDJPY', 'USDCAD'],
      forecast: '-$65.0B',
      previous: '-$64.3B',
    },
  ];

  return events;
}

// Generate historical data for event types
function generateEventTypeData(): EventTypeData[] {
  const eventTypes = [
    { type: 'Non-Farm Payrolls (NFP)', category: 'Employment' as EventCategory, avgPips: 80 },
    { type: 'Consumer Price Index (CPI)', category: 'Inflation' as EventCategory, avgPips: 70 },
    { type: 'FOMC Meeting', category: 'Interest Rates' as EventCategory, avgPips: 95 },
    { type: 'Interest Rate Decision', category: 'Interest Rates' as EventCategory, avgPips: 85 },
    { type: 'Unemployment Rate', category: 'Employment' as EventCategory, avgPips: 45 },
    { type: 'GDP Growth Rate', category: 'GDP' as EventCategory, avgPips: 55 },
    { type: 'Retail Sales', category: 'GDP' as EventCategory, avgPips: 40 },
    { type: 'Manufacturing PMI', category: 'PMI' as EventCategory, avgPips: 35 },
    { type: 'Services PMI', category: 'PMI' as EventCategory, avgPips: 30 },
    { type: 'Producer Price Index (PPI)', category: 'Inflation' as EventCategory, avgPips: 50 },
    { type: 'Trade Balance', category: 'Trade Balance' as EventCategory, avgPips: 25 },
    { type: 'Industrial Production', category: 'GDP' as EventCategory, avgPips: 28 },
    { type: 'Consumer Confidence', category: 'GDP' as EventCategory, avgPips: 32 },
    { type: 'Core CPI', category: 'Inflation' as EventCategory, avgPips: 65 },
    { type: 'Initial Jobless Claims', category: 'Employment' as EventCategory, avgPips: 38 },
  ];

  return eventTypes.map(et => ({
    type: et.type,
    category: et.category,
    avgPipMovement: et.avgPips,
    occurrences: generateOccurrences(et.avgPips),
  }));
}

function generateOccurrences(avgPips: number): HistoricalOccurrence[] {
  const occurrences: HistoricalOccurrence[] = [];
  const now = new Date();

  for (let i = 0; i < 6; i++) {
    const date = new Date(now);
    date.setMonth(date.getMonth() - i - 1);

    const basePrice = 1.0850;
    const forecast = 3.2 + (Math.random() - 0.5) * 0.5;
    const actual = forecast + (Math.random() - 0.5) * 0.8;
    const previous = forecast + (Math.random() - 0.5) * 0.6;
    const deviation = actual - forecast;

    // Generate 30min before to 30min after (1min candles)
    const priceData = [];
    let price = basePrice;

    for (let min = -30; min <= 30; min++) {
      // Big spike at event time (min = 0)
      let volatility = 0.00005;
      if (min >= -2 && min <= 5) {
        volatility = 0.0002 * (avgPips / 50); // Scale by event impact
      }

      const change = (Math.random() - 0.5) * volatility;
      price += change;

      const candleVolatility = volatility * 0.3;
      const open = price;
      const close = price + (Math.random() - 0.5) * candleVolatility;
      const high = Math.max(open, close) + Math.random() * candleVolatility;
      const low = Math.min(open, close) - Math.random() * candleVolatility;

      priceData.push({
        time: min,
        open: parseFloat(open.toFixed(5)),
        high: parseFloat(high.toFixed(5)),
        low: parseFloat(low.toFixed(5)),
        close: parseFloat(close.toFixed(5)),
      });
    }

    const pipMovement = Math.abs(priceData[0].close - priceData[priceData.length - 1].close) * 10000;

    occurrences.push({
      date: date.toISOString().substring(0, 10),
      forecast: forecast.toFixed(2) + '%',
      actual: actual.toFixed(2) + '%',
      previous: previous.toFixed(2) + '%',
      deviation: parseFloat(deviation.toFixed(2)),
      priceData,
      pipMovement: parseFloat(pipMovement.toFixed(1)),
    });
  }

  return occurrences;
}

export function NewsImpactAnalyzer() {
  const [selectedEventType, setSelectedEventType] = useState<string | null>(null);
  const [categoryFilter, setCategoryFilter] = useState<EventCategory | 'All'>('All');
  const [currentTime, setCurrentTime] = useState(Date.now());

  const { toggleFavorite, isFavorited } = useNewsImpactStore();

  const upcomingEvents = useMemo(() => generateUpcomingEvents(), []);
  const eventTypeData = useMemo(() => generateEventTypeData(), []);

  // Update time every second for countdowns
  useEffect(() => {
    const interval = setInterval(() => setCurrentTime(Date.now()), 1000);
    return () => clearInterval(interval);
  }, []);

  const filteredEventTypes = useMemo(() => {
    if (categoryFilter === 'All') return eventTypeData;
    return eventTypeData.filter(e => e.category === categoryFilter);
  }, [eventTypeData, categoryFilter]);

  const selectedEventData = eventTypeData.find(e => e.type === selectedEventType);

  const getImpactColor = (impact: ImpactLevel) => {
    return impact === 'High' ? 'bg-red-500/20 text-red-400 border-red-500/50' :
           impact === 'Medium' ? 'bg-yellow-500/20 text-yellow-400 border-yellow-500/50' :
           'bg-emerald-500/20 text-emerald-400 border-emerald-500/50';
  };

  const getCountdown = (eventTime: string) => {
    const diff = new Date(eventTime).getTime() - currentTime;
    if (diff < 0) return 'LIVE';

    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((diff % (1000 * 60)) / 1000);

    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
  };

  return (
    <div className="h-full overflow-auto bg-zinc-900 p-4 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Calendar className="w-5 h-5 text-cyan-400" />
          <h2 className="text-lg font-semibold text-white">Economic News Impact Analyzer</h2>
        </div>
      </div>

      {/* Upcoming Events (Next 24h) */}
      <div className="bg-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3 flex items-center gap-2">
          <Clock className="w-4 h-4 text-cyan-400" />
          Upcoming Events (Next 24h)
        </h3>
        <div className="space-y-2">
          {upcomingEvents.map(event => {
            const countdown = getCountdown(event.time);
            return (
              <div
                key={event.id}
                className="flex items-center justify-between p-3 bg-zinc-900 rounded hover:bg-zinc-700/50 cursor-pointer"
                onClick={() => setSelectedEventType(event.type)}
              >
                <div className="flex items-center gap-3 flex-1">
                  <div className={`font-mono text-sm ${countdown === 'LIVE' ? 'text-red-400 font-bold' : 'text-cyan-400'}`}>
                    {countdown}
                  </div>
                  <div className="flex-1">
                    <div className="text-sm font-semibold text-white">{event.type}</div>
                    <div className="text-xs text-zinc-500">
                      {event.currency} • {event.affectedSymbols.join(', ')}
                    </div>
                  </div>
                  <div className={`px-2 py-1 rounded border text-xs font-medium ${getImpactColor(event.impact)}`}>
                    {event.impact}
                  </div>
                  {event.forecast && (
                    <div className="text-xs text-zinc-400">
                      F: {event.forecast} | P: {event.previous}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Category Filter */}
      <div className="bg-zinc-800 rounded-lg p-4">
        <div className="flex items-center gap-2 mb-3">
          <Filter className="w-4 h-4 text-zinc-400" />
          <span className="text-sm font-semibold text-zinc-300">Event Category</span>
        </div>
        <div className="flex gap-2 flex-wrap">
          {(['All', 'Employment', 'Inflation', 'Interest Rates', 'GDP', 'PMI', 'Trade Balance'] as const).map(cat => (
            <button
              key={cat}
              onClick={() => setCategoryFilter(cat)}
              className={`px-3 py-1 rounded text-xs font-medium ${
                categoryFilter === cat
                  ? 'bg-cyan-600 text-white'
                  : 'bg-zinc-700 text-zinc-300 hover:bg-zinc-600'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Event Types with Avg Pip Movement */}
      <div className="bg-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3 flex items-center gap-2">
          <BarChart3 className="w-4 h-4 text-emerald-400" />
          Average Volatility by Event Type
        </h3>
        <div className="grid grid-cols-3 gap-2">
          {filteredEventTypes.map(eventType => (
            <button
              key={eventType.type}
              onClick={() => setSelectedEventType(eventType.type)}
              className={`p-3 rounded text-left ${
                selectedEventType === eventType.type
                  ? 'bg-cyan-600 border-2 border-cyan-400'
                  : 'bg-zinc-900 hover:bg-zinc-700'
              }`}
            >
              <div className="flex items-start justify-between mb-1">
                <div className="text-sm font-semibold text-white">{eventType.type}</div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    toggleFavorite(eventType.type);
                  }}
                  className="p-1"
                >
                  <Star
                    className={`w-3 h-3 ${
                      isFavorited(eventType.type)
                        ? 'fill-yellow-400 text-yellow-400'
                        : 'text-zinc-600'
                    }`}
                  />
                </button>
              </div>
              <div className="text-xs text-zinc-500 mb-2">{eventType.category}</div>
              <div className="flex items-center gap-1">
                <TrendingUp className="w-3 h-3 text-emerald-400" />
                <span className="text-xs font-bold text-emerald-400">
                  ~{eventType.avgPipMovement} pips
                </span>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Historical Impact Analysis */}
      {selectedEventData && (
        <div className="bg-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3 flex items-center gap-2">
            <AlertCircle className="w-4 h-4 text-purple-400" />
            Historical Impact Analysis: {selectedEventData.type}
          </h3>
          <div className="space-y-4">
            {selectedEventData.occurrences.map((occurrence, index) => (
              <div key={index} className="bg-zinc-900 rounded p-3">
                <div className="flex items-center justify-between mb-2">
                  <div className="text-sm font-semibold text-white">{occurrence.date}</div>
                  <div className="flex items-center gap-3 text-xs">
                    <span className="text-zinc-500">Forecast: <span className="text-zinc-300">{occurrence.forecast}</span></span>
                    <span className="text-zinc-500">Actual: <span className={occurrence.deviation >= 0 ? 'text-emerald-400' : 'text-red-400'}>{occurrence.actual}</span></span>
                    <span className="text-zinc-500">Previous: <span className="text-zinc-400">{occurrence.previous}</span></span>
                    <span className={`font-bold ${Math.abs(occurrence.deviation) > 0.3 ? 'text-orange-400' : 'text-zinc-400'}`}>
                      Δ {occurrence.deviation >= 0 ? '+' : ''}{occurrence.deviation.toFixed(2)}%
                    </span>
                  </div>
                </div>

                {/* Price Impact Chart (30min before/after) */}
                <div className="bg-zinc-950 rounded p-2">
                  <div className="text-xs text-zinc-500 mb-1">
                    Price Movement: {occurrence.pipMovement} pips
                  </div>
                  <PriceImpactChart data={occurrence.priceData} />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// Price Impact Chart (30min before/after event)
function PriceImpactChart({
  data
}: {
  data: { time: number; open: number; high: number; low: number; close: number }[];
}) {
  const prices = data.flatMap(c => [c.high, c.low]);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const priceRange = maxPrice - minPrice;

  const svgWidth = 800;
  const svgHeight = 120;
  const candleWidth = Math.max(2, (svgWidth / data.length) - 1);

  return (
    <svg width="100%" height={svgHeight} viewBox={`0 0 ${svgWidth} ${svgHeight}`} className="bg-zinc-950">
      {/* Event time line */}
      <line
        x1={svgWidth / 2}
        y1={10}
        x2={svgWidth / 2}
        y2={svgHeight - 10}
        stroke="#ef4444"
        strokeWidth="2"
        strokeDasharray="4,4"
      />
      <text
        x={svgWidth / 2}
        y={8}
        textAnchor="middle"
        fill="#ef4444"
        fontSize="9"
        fontWeight="bold"
      >
        EVENT
      </text>

      {/* Candles */}
      {data.map((candle, i) => {
        const x = (i / data.length) * svgWidth + candleWidth / 2;
        const openY = svgHeight - 10 - ((candle.open - minPrice) / priceRange) * (svgHeight - 20);
        const closeY = svgHeight - 10 - ((candle.close - minPrice) / priceRange) * (svgHeight - 20);
        const highY = svgHeight - 10 - ((candle.high - minPrice) / priceRange) * (svgHeight - 20);
        const lowY = svgHeight - 10 - ((candle.low - minPrice) / priceRange) * (svgHeight - 20);

        const isBullish = candle.close >= candle.open;
        const bodyTop = Math.min(openY, closeY);
        const bodyHeight = Math.abs(closeY - openY) || 1;

        // Highlight candles near event time
        const isNearEvent = candle.time >= -2 && candle.time <= 5;
        const color = isBullish ? (isNearEvent ? '#22c55e' : '#16a34a') : (isNearEvent ? '#ef4444' : '#dc2626');
        const opacity = isNearEvent ? 1 : 0.7;

        return (
          <g key={i}>
            {/* Wick */}
            <line
              x1={x}
              y1={highY}
              x2={x}
              y2={lowY}
              stroke={color}
              strokeWidth="1"
              opacity={opacity}
            />
            {/* Body */}
            <rect
              x={x - candleWidth / 2}
              y={bodyTop}
              width={candleWidth}
              height={bodyHeight}
              fill={color}
              opacity={opacity}
            />
          </g>
        );
      })}

      {/* Time labels */}
      <text x="5" y={svgHeight - 2} fill="#71717a" fontSize="9">-30m</text>
      <text x={svgWidth - 25} y={svgHeight - 2} fill="#71717a" fontSize="9">+30m</text>
    </svg>
  );
}
