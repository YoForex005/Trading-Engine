# Historical Data Quick Start

## For Developers

### Quick Integration Checklist

- [x] IndexedDB layer implemented (`src/db/ticksDB.ts`)
- [x] API client created (`src/api/historyClient.ts`)
- [x] Service manager implemented (`src/services/historyDataManager.ts`)
- [x] React hook available (`src/hooks/useHistoricalData.ts`)
- [x] Download UI component created (`src/components/HistoryDownloader.tsx`)
- [x] Chart integration component created (`src/components/ChartWithHistory.tsx`)
- [x] Exports added to all index files
- [x] Types defined (`src/types/history.ts`)
- [x] App.tsx updated with ChartWithHistory
- [x] BottomDock.tsx updated with Historical Data tab

### File Structure

```
clients/desktop/src/
├── db/
│   ├── ticksDB.ts              # IndexedDB wrapper
│   └── index.ts                # Exports ticksDB
├── api/
│   ├── historyClient.ts        # Backend API client
│   └── index.ts                # Exports historyClient
├── services/
│   ├── historyDataManager.ts   # Download orchestrator
│   └── index.ts                # Exports historyDataManager
├── hooks/
│   ├── useHistoricalData.ts    # React hook
│   └── index.ts                # Exports useHistoricalData
├── components/
│   ├── HistoryDownloader.tsx   # Download manager UI
│   ├── ChartWithHistory.tsx    # Enhanced chart
│   └── BottomDock.tsx          # Updated with new tab
├── types/
│   ├── history.ts              # Type definitions
│   └── index.ts                # Exports all types
└── App.tsx                     # Updated with ChartWithHistory
```

## Usage Examples

### Using the Hook

```typescript
import { useHistoricalData } from '../hooks/useHistoricalData';

function MyComponent() {
  const {
    data,           // TickData[]
    isLoading,      // boolean
    error,          // Error | null
    progress,       // number (0-100)
    symbolInfo,     // SymbolDataInfo | null
    refetch,        // () => Promise<void>
    downloadIfMissing  // () => Promise<void>
  } = useHistoricalData({
    symbol: 'EURUSD',
    dateRange: { from: '2026-01-13', to: '2026-01-20' },
    autoLoad: true,
    onProgress: (progress) => console.log(`Download: ${progress}%`)
  });

  return (
    <div>
      {isLoading && <p>Loading...</p>}
      {error && <p>Error: {error.message}</p>}
      {data.length > 0 && <p>Loaded {data.length} ticks</p>}
    </div>
  );
}
```

### Using the Service Directly

```typescript
import { historyDataManager } from '../services/historyDataManager';

// Download data
const taskId = await historyDataManager.downloadData(
  'EURUSD',
  { from: '2026-01-13', to: '2026-01-20' },
  (task) => {
    console.log(`Progress: ${task.progress}%`);
    if (task.status === 'completed') {
      console.log('Download complete!');
    }
  }
);

// Get cached data
const ticks = await historyDataManager.getTicks(
  'EURUSD',
  { from: '2026-01-13', to: '2026-01-20' }
);

// Get storage stats
const stats = await historyDataManager.getStorageStats();
console.log(`Total ticks: ${stats.tickCount}`);
console.log(`Total size: ${stats.totalSize} bytes`);

// Clear cache
await historyDataManager.clearSymbol('EURUSD');
await historyDataManager.clearAll();
```

### Using IndexedDB Directly

```typescript
import { ticksDB } from '../db/ticksDB';

// Initialize
await ticksDB.initialize();

// Store ticks
await ticksDB.storeTicks('EURUSD', '2026-01-20', [
  { symbol: 'EURUSD', bid: 1.0950, ask: 1.0952, timestamp: Date.now() }
]);

// Retrieve ticks
const ticks = await ticksDB.getTicksByDate('EURUSD', '2026-01-20');

// Check if data exists
const hasData = await ticksDB.hasData('EURUSD', '2026-01-20');

// Get downloaded dates
const dates = await ticksDB.getDownloadedDates('EURUSD');

// Delete data
await ticksDB.deleteTicks('EURUSD', '2026-01-20');
```

## Testing in Browser

### 1. Open DevTools Console
```javascript
// Access IndexedDB
const request = indexedDB.open('TradingEngineTicksDB');
request.onsuccess = () => console.log('DB exists');

// Check stored data
const transaction = request.result.transaction(['ticks'], 'readonly');
const store = transaction.objectStore('ticks');
const countRequest = store.count();
countRequest.onsuccess = () => console.log('Tick count:', countRequest.result);
```

### 2. Test API Endpoints
```bash
# Get available symbols
curl http://localhost:7999/api/history/available

# Get symbol info
curl http://localhost:7999/api/history/info?symbol=EURUSD

# Get ticks
curl "http://localhost:7999/api/history/ticks?symbol=EURUSD&date=2026-01-20&offset=0&limit=100"

# Get tick count
curl "http://localhost:7999/api/history/count?symbol=EURUSD&date=2026-01-20"
```

### 3. UI Testing Flow
1. Open app → Click "Historical Data" tab
2. Select EURUSD from symbol list
3. Set date range (e.g., last 7 days)
4. Click "Download Historical Data"
5. Watch progress bar fill
6. Click pause/resume to test controls
7. Check storage stats in header
8. Switch to chart view to see data

## Backend API Requirements

### Endpoint Format
```
Base URL: http://localhost:7999
Endpoints:
  GET /api/history/available
  GET /api/history/info?symbol={symbol}
  GET /api/history/ticks?symbol={symbol}&date={date}&offset={offset}&limit={limit}
  GET /api/history/count?symbol={symbol}&date={date}
  GET /api/history/verify?symbol={symbol}&date={date}&checksum={checksum}
```

### Response Formats

**Available Symbols:**
```json
[
  {
    "symbol": "EURUSD",
    "availableDates": ["2026-01-13", "2026-01-14", "2026-01-15"],
    "totalTicks": 150000,
    "firstDate": "2026-01-13",
    "lastDate": "2026-01-20"
  }
]
```

**Ticks:**
```json
{
  "ticks": [
    {
      "symbol": "EURUSD",
      "bid": 1.09501,
      "ask": 1.09523,
      "timestamp": 1737396000000
    }
  ],
  "total": 50000
}
```

## Configuration

### Update API URL
Edit `clients/desktop/.env`:
```env
VITE_API_URL=http://localhost:7999
```

### Adjust Download Settings
Edit `clients/desktop/src/api/historyClient.ts`:
```typescript
const CHUNK_SIZE = 5000;      // Ticks per chunk
const MAX_RETRIES = 3;        // Retry attempts
const RETRY_DELAY = 1000;     // Retry delay (ms)
```

Edit `clients/desktop/src/services/historyDataManager.ts`:
```typescript
private maxConcurrentDownloads = 3;  // Parallel downloads
```

## Common Tasks

### Enable/Disable Historical Data
In `App.tsx`:
```typescript
const [enableHistoricalData, setEnableHistoricalData] = useState(true);
```

### Change Default Date Range
In `ChartWithHistory.tsx`:
```typescript
const defaultDateRange = useMemo<DateRange>(() => ({
  from: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
  to: new Date().toISOString().split('T')[0]
}), []);
```

### Add Custom Error Handling
```typescript
import { HistoryApiError } from '../api/historyClient';

try {
  await historyClient.getTicksByDate('EURUSD', '2026-01-20');
} catch (error) {
  if (error instanceof HistoryApiError) {
    console.error(`API Error ${error.statusCode}: ${error.message}`);
  }
}
```

## Troubleshooting Quick Fixes

| Issue | Solution |
|-------|----------|
| "Failed to fetch" | Check backend is running on port 7999 |
| IndexedDB quota exceeded | Clear browser data or reduce date range |
| Download stuck | Cancel task and restart |
| No data in chart | Verify symbol name matches exactly |
| CORS error | Add CORS headers to backend |
| Slow downloads | Reduce CHUNK_SIZE or maxConcurrentDownloads |

## Dependencies

All required dependencies are already in `package.json`:
- `react` ^19.2.0
- `lucide-react` ^0.562.0 (for icons)

No additional packages needed - uses native browser IndexedDB API.

## Next Steps

1. **Start Backend**: Ensure port 7999 is running with history endpoints
2. **Test Downloads**: Download sample data via UI
3. **Verify Storage**: Check IndexedDB in browser DevTools
4. **Test Chart**: Switch to chart view and verify data loads
5. **Performance Test**: Download large date ranges
6. **Error Test**: Test with backend offline

## Support

For issues or questions:
1. Check browser console for errors
2. Verify backend endpoints are responding
3. Review `HISTORICAL_DATA_INTEGRATION_GUIDE.md` for detailed documentation
4. Test with sample data first before production use
