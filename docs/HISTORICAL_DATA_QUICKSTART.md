# Historical Data System - Quick Start Guide

## 5-Minute Setup

### 1. Install (Already Included)

All dependencies are already installed in the desktop client. No additional packages needed.

### 2. Basic Usage - React Component

```tsx
import { ChartWithHistory } from './components';

function TradingView() {
  return (
    <ChartWithHistory
      symbol="EURUSD"
      enableHistoricalData={true}
      chartType="candlestick"
      timeframe="1h"
    />
  );
}
```

### 3. Manual Download UI

```tsx
import { HistoryDownloader } from './components';

function DataManagement() {
  return (
    <div className="w-full h-screen">
      <HistoryDownloader />
    </div>
  );
}
```

### 4. Programmatic Download

```tsx
import { historyDataManager } from './services';

async function downloadData() {
  const taskId = await historyDataManager.downloadData(
    'EURUSD',
    { from: '2026-01-01', to: '2026-01-20' },
    (task) => {
      console.log(`Progress: ${task.progress}%`);
      if (task.status === 'completed') {
        console.log('Download complete!');
      }
    }
  );
}
```

### 5. Custom Hook

```tsx
import { useHistoricalData } from './hooks';

function MyChart() {
  const { data, isLoading, error, downloadIfMissing } = useHistoricalData({
    symbol: 'EURUSD',
    dateRange: { from: '2026-01-01', to: '2026-01-20' },
    autoLoad: true
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return <div>Loaded {data.length} ticks</div>;
}
```

## Common Use Cases

### Case 1: Chart with Auto-Download Prompt

```tsx
<ChartWithHistory
  symbol="GBPUSD"
  enableHistoricalData={true}
  historicalDateRange={{
    from: '2026-01-15',
    to: '2026-01-20'
  }}
/>
```

- Shows download prompt if data is missing
- Progress indicator during download
- Seamless integration with live data

### Case 2: Bulk Download in Background

```tsx
import { historyDataManager } from './services';

async function bulkDownload() {
  const symbols = ['EURUSD', 'GBPUSD', 'USDJPY'];
  const dateRange = { from: '2026-01-01', to: '2026-01-20' };

  for (const symbol of symbols) {
    await historyDataManager.downloadData(symbol, dateRange);
  }
}
```

### Case 3: Check Available Data

```tsx
import { historyDataManager } from './services';

async function checkData() {
  const symbols = await historyDataManager.getAvailableSymbols();

  symbols.forEach(info => {
    console.log(`${info.symbol}:`);
    console.log(`  Available: ${info.availableDates.length} days`);
    console.log(`  Downloaded: ${info.downloadedDates.length} days`);
    console.log(`  Progress: ${info.downloadProgress}%`);
  });
}
```

### Case 4: Custom OHLC Aggregation

```tsx
import { historyDataManager, chartDataService } from './services';

async function getOHLC() {
  // Get tick data
  const ticks = await historyDataManager.getTicks(
    'EURUSD',
    { from: '2026-01-20', to: '2026-01-20' }
  );

  // Convert to 1-hour candles
  const ohlc = chartDataService.aggregateToOHLC(ticks, '1h');

  // Calculate Heikin-Ashi
  const ha = chartDataService.calculateHeikinAshi(ohlc);

  return { ohlc, ha };
}
```

### Case 5: Storage Management

```tsx
import { historyDataManager } from './services';

async function manageStorage() {
  // Get storage stats
  const stats = await historyDataManager.getStorageStats();
  console.log(`Total size: ${stats.totalSize / 1024 / 1024} MB`);
  console.log(`Total ticks: ${stats.tickCount}`);

  // Clear old symbol data
  await historyDataManager.clearSymbol('EURUSD');

  // Clear all data
  await historyDataManager.clearAll();
}
```

## API Endpoints Required

Your backend must implement these endpoints:

```
GET  /api/history/available
GET  /api/history/info?symbol=EURUSD
GET  /api/history/ticks?symbol=EURUSD&date=2026-01-20&offset=0&limit=5000
GET  /api/history/count?symbol=EURUSD&date=2026-01-20
GET  /api/history/verify?symbol=EURUSD&date=2026-01-20
```

See `docs/HISTORICAL_DATA_CLIENT.md` for detailed API specifications.

## Admin Panel Integration

```tsx
import { DataManagement } from './components/dashboard/DataManagement';

function AdminDashboard() {
  return (
    <div>
      <h1>Admin Dashboard</h1>
      <DataManagement />
    </div>
  );
}
```

Features:
- Storage metrics
- Backfill operations
- Export/Import
- Task monitoring

## Performance Tips

### 1. Download Recent Data First
```tsx
// Download last 7 days first
await historyDataManager.downloadData(
  'EURUSD',
  {
    from: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    to: new Date().toISOString().split('T')[0]
  }
);

// Then download older data in background
```

### 2. Use Background Downloads
```tsx
const taskId = await historyDataManager.downloadData(
  'EURUSD',
  dateRange,
  null // No callback = background mode
);

// Check status later
const task = historyDataManager.getDownloadTask(taskId);
console.log(task.status, task.progress);
```

### 3. Limit Date Ranges
```tsx
// Good: Small date ranges
const data = await historyDataManager.getTicks(
  'EURUSD',
  { from: '2026-01-20', to: '2026-01-20' }
);

// Avoid: Large date ranges without limiting
// const data = await historyDataManager.getTicks(
//   'EURUSD',
//   { from: '2020-01-01', to: '2026-01-20' }
// );
```

### 4. Use Pagination for Large Queries
```tsx
const ticks = await historyDataManager.getTicks(
  'EURUSD',
  dateRange,
  10000 // Limit to 10,000 ticks
);
```

## Troubleshooting

### Downloads not starting?
- Check network connection
- Verify API endpoints are accessible
- Check browser console for errors

### High storage usage?
```tsx
const stats = await historyDataManager.getStorageStats();
console.log(stats);

// Clear old data
await historyDataManager.clearSymbol('OLD_SYMBOL');
```

### Slow queries?
- Reduce date range
- Add limit parameter
- Check IndexedDB indexes

### Data not showing in chart?
- Verify data is downloaded
- Check date range matches
- Look for console errors

## Next Steps

1. Read full documentation: `docs/HISTORICAL_DATA_CLIENT.md`
2. Implement backend API endpoints
3. Test with real data
4. Optimize for your use case

## Support

For issues or questions:
1. Check documentation
2. Review browser console
3. Test with small data sets first
4. File an issue on GitHub
