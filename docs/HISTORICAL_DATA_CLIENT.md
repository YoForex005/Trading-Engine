# Historical Data Client-Side System

## Overview

A comprehensive client-side historical data retrieval and caching system for the trading platform. This system enables efficient storage, retrieval, and management of historical tick data with offline support, progressive downloads, and seamless chart integration.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Application                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────┐    ┌──────────────────┐               │
│  │  UI Components   │    │  React Hooks     │               │
│  │                  │    │                  │               │
│  │  HistoryDown-    │───▶│  useHistorical-  │               │
│  │  loader          │    │  Data            │               │
│  │                  │    │                  │               │
│  │  ChartWith-      │    └────────┬─────────┘               │
│  │  History         │             │                         │
│  └──────────────────┘             │                         │
│                                   │                         │
│  ┌─────────────────────────────────▼─────────────────────┐  │
│  │         HistoryDataManager (Service Layer)           │  │
│  │  - Download orchestration                            │  │
│  │  - Queue management                                  │  │
│  │  - Progress tracking                                 │  │
│  │  - Cache coordination                                │  │
│  └───────────┬──────────────────────────┬────────────────┘  │
│              │                          │                   │
│  ┌───────────▼──────────┐   ┌──────────▼─────────────┐     │
│  │   HistoryClient      │   │   TicksDB (IndexedDB)  │     │
│  │   (API Client)       │   │   - Local storage      │     │
│  │  - HTTP requests     │   │   - Efficient queries  │     │
│  │  - Retry logic       │   │   - LRU eviction       │     │
│  │  - Chunked download  │   │   - Compression        │     │
│  └──────────────────────┘   └────────────────────────┘     │
│                                                               │
└───────────────────────────────────────────────────────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │  Backend API   │
                    │  /api/history  │
                    └────────────────┘
```

## Components

### 1. Data Types (`types/history.ts`)

Core TypeScript types for the historical data system:

- `TickData` - Individual tick data point
- `DateRange` - Date range specification
- `SymbolDataInfo` - Symbol metadata and availability
- `DownloadTask` - Download task tracking
- `StorageStats` - Storage metrics
- `CacheOptions` - Cache configuration

### 2. Storage Layer (`db/ticksDB.ts`)

IndexedDB-based storage for persistent caching:

**Features:**
- Composite key indexing (symbol + date + timestamp)
- Efficient range queries
- Automatic compression for large datasets
- Metadata tracking
- Storage statistics

**Key Methods:**
```typescript
await ticksDB.initialize();
await ticksDB.storeTicks(symbol, date, ticks);
const ticks = await ticksDB.getTicks(symbol, dateRange);
const hasData = await ticksDB.hasData(symbol, date);
const stats = await ticksDB.getStorageStats();
await ticksDB.clearSymbol(symbol);
```

### 3. API Client (`api/historyClient.ts`)

HTTP client for fetching historical data from the backend:

**Features:**
- Automatic retry with exponential backoff
- Chunked downloads (5000 ticks per chunk)
- Cancellable downloads
- Data integrity verification
- Progress tracking

**Key Methods:**
```typescript
const info = await historyClient.getSymbolInfo(symbol);
const chunk = await historyClient.getTicksByDate(symbol, date);
await historyClient.downloadTicks(taskId, symbol, date, onChunk, onProgress);
historyClient.cancelDownload(taskId);
```

### 4. Data Manager (`services/historyDataManager.ts`)

Orchestration layer for download and cache management:

**Features:**
- Concurrent download management (max 3 concurrent)
- Download queue with prioritization
- Automatic cache fallback
- Progress callbacks
- Task state management

**Key Methods:**
```typescript
const symbols = await historyDataManager.getAvailableSymbols();
const taskId = await historyDataManager.downloadData(symbol, dateRange, onProgress);
const ticks = await historyDataManager.getTicks(symbol, dateRange);
historyDataManager.pauseDownload(taskId);
historyDataManager.resumeDownload(taskId);
```

### 5. UI Components

#### HistoryDownloader (`components/HistoryDownloader.tsx`)

Full-featured UI for managing historical data downloads:

**Features:**
- Symbol selection with search
- Date range picker
- Download progress tracking
- Pause/Resume/Cancel controls
- Storage statistics display
- Background download support

**Usage:**
```tsx
import { HistoryDownloader } from './components/HistoryDownloader';

function App() {
  return <HistoryDownloader />;
}
```

#### ChartWithHistory (`components/ChartWithHistory.tsx`)

Enhanced chart component with historical data integration:

**Features:**
- Automatic historical data loading
- Download prompts for missing data
- Progress indicators
- Seamless live/historical data merge
- Error handling

**Usage:**
```tsx
import { ChartWithHistory } from './components/ChartWithHistory';

function TradingView() {
  return (
    <ChartWithHistory
      symbol="EURUSD"
      enableHistoricalData={true}
      historicalDateRange={{
        from: '2026-01-01',
        to: '2026-01-20'
      }}
    />
  );
}
```

### 6. React Hook (`hooks/useHistoricalData.ts`)

React hook for easy data access:

**Features:**
- Automatic data loading
- Download on demand
- Progress tracking
- Error handling
- Symbol info integration

**Usage:**
```tsx
import { useHistoricalData } from './hooks/useHistoricalData';

function MyComponent() {
  const {
    data,
    isLoading,
    error,
    progress,
    symbolInfo,
    downloadIfMissing
  } = useHistoricalData({
    symbol: 'EURUSD',
    dateRange: { from: '2026-01-01', to: '2026-01-20' },
    autoLoad: true,
    onProgress: (p) => console.log(`Progress: ${p}%`)
  });

  return (
    <div>
      {isLoading && <p>Loading... {progress}%</p>}
      {error && <p>Error: {error.message}</p>}
      <p>Loaded {data.length} ticks</p>
    </div>
  );
}
```

### 7. Chart Data Service (`services/chartDataService.ts`)

Utilities for aggregating tick data into OHLC candles:

**Features:**
- Tick to OHLC aggregation
- Heikin-Ashi calculation
- Timeframe resampling
- Historical/live data merging

**Usage:**
```typescript
import { chartDataService } from './services/chartDataService';

const ohlc = chartDataService.aggregateToOHLC(ticks, '1h');
const ha = chartDataService.calculateHeikinAshi(ohlc);
const resampled = chartDataService.resampleOHLC(ohlc, '4h');
```

## Admin Panel Integration

### DataManagement Component (`admin/broker-admin/src/components/dashboard/DataManagement.tsx`)

Admin dashboard for server-side data management:

**Features:**
- Storage metrics overview
- Backfill operations
- Data export (CSV)
- Data import
- Task monitoring

## Backend API Requirements

The client-side system expects the following API endpoints:

### GET /api/history/available
Returns available symbols and date ranges:
```json
[
  {
    "symbol": "EURUSD",
    "availableDates": ["2026-01-01", "2026-01-02", ...],
    "totalTicks": 1000000,
    "firstDate": "2026-01-01",
    "lastDate": "2026-01-20"
  }
]
```

### GET /api/history/info?symbol=EURUSD
Returns symbol-specific information:
```json
{
  "symbol": "EURUSD",
  "availableDates": [...],
  "totalTicks": 1000000,
  "firstDate": "2026-01-01",
  "lastDate": "2026-01-20"
}
```

### GET /api/history/ticks?symbol=EURUSD&date=2026-01-20&offset=0&limit=5000
Returns tick data chunks:
```json
{
  "ticks": [
    {
      "symbol": "EURUSD",
      "bid": 1.12345,
      "ask": 1.12355,
      "timestamp": 1737331200000,
      "volume": 100
    }
  ],
  "total": 50000
}
```

### GET /api/history/count?symbol=EURUSD&date=2026-01-20
Returns tick count for a date:
```json
{
  "count": 50000
}
```

### GET /api/history/verify?symbol=EURUSD&date=2026-01-20
Verifies data integrity:
```json
{
  "valid": true,
  "count": 50000,
  "checksum": "abc123"
}
```

## Storage Management

### IndexedDB Schema

**Ticks Store:**
- Key: `[symbol, date, timestamp]`
- Indexes:
  - `symbol`
  - `date`
  - `symbolDate` (composite)
  - `timestamp`

**Metadata Store:**
- Key: `id` (format: `{symbol}-{date}`)
- Data: `{ symbol, date, count, size, timestamp }`

### Storage Limits

- **Chunk Size:** 5000 ticks per download
- **Compression Threshold:** 10,000 ticks
- **Concurrent Downloads:** 3 maximum
- **Typical Storage:** ~32 bytes per tick

### LRU Eviction

Storage stats are tracked, and old data can be cleared:

```typescript
// Clear old symbol data
await historyDataManager.clearSymbol('EURUSD');

// Clear all data
await historyDataManager.clearAll();
```

## Performance Optimization

### Progressive Loading
1. Load recent data first (last 7 days)
2. Download older data in background
3. Priority queue for user-requested ranges

### Efficient Queries
- IndexedDB range queries for date filtering
- Composite indexes for fast lookups
- Lazy loading for large datasets

### Caching Strategy
1. Check local cache first
2. Fallback to server if missing
3. Store downloaded data automatically
4. Background sync for updates

## Error Handling

### Retry Logic
- Automatic retry with exponential backoff
- Max 3 retries per request
- Client errors (4xx) don't retry
- Server errors (5xx) trigger retry

### User Feedback
- Progress indicators
- Error messages
- Download status
- Storage statistics

## Future Enhancements

### Planned Features
1. **Compression:** LZ4 compression for storage
2. **Web Workers:** Background processing
3. **Service Workers:** Offline support
4. **Delta Updates:** Incremental data sync
5. **Smart Preloading:** Predictive data loading
6. **Cross-Tab Sync:** SharedArrayBuffer for multi-tab

### Performance Targets
- **Initial Load:** < 500ms
- **Download Speed:** > 10,000 ticks/sec
- **Query Time:** < 50ms for 1M ticks
- **Storage Efficiency:** > 70% compression ratio

## Testing

### Unit Tests
```bash
npm run test
```

### Integration Tests
```bash
npm run test:integration
```

### Performance Tests
```bash
npm run test:performance
```

## Troubleshooting

### Common Issues

**Issue: Downloads fail repeatedly**
- Check network connection
- Verify API endpoint availability
- Check browser console for errors

**Issue: High storage usage**
- Check storage stats
- Clear old symbols
- Adjust cache limits

**Issue: Slow queries**
- Verify IndexedDB indexes
- Check data volume
- Consider date range limits

## License

MIT License - See LICENSE file for details
