# Quick Print Integration Guide

This guide shows how to add print functionality to your existing chart components in under 5 minutes.

## Step 1: Import Dependencies

```tsx
import { useChartPrint } from '../hooks/useChartPrint';
import { PrintSetupDialog } from './dialogs/PrintSetupDialog';
import { PrintPreviewDialog } from './dialogs/PrintPreviewDialog';
import { Printer } from 'lucide-react';
```

## Step 2: Add the Hook

Add the `useChartPrint` hook to your chart component:

```tsx
function YourChartComponent() {
  // Your existing code...
  const chartApiRef = useRef<IChartApi | null>(null);

  // Add this hook
  const {
    showSetupDialog,
    showPreviewDialog,
    printPreferences,
    chartData,
    openPrintSetup,
    handleSetupComplete,
    handleDirectPrint,
    closeDialogs,
  } = useChartPrint({
    accountId: 'your-account-id', // Optional
    symbol: selectedSymbol,
    timeframe: currentTimeframe,
    chartApi: chartApiRef.current,
  });

  // Rest of your component...
}
```

## Step 3: Add Print Button

Add a print button to your chart toolbar:

```tsx
<button
  onClick={openPrintSetup}
  className="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded transition-colors"
  title="Print Chart"
>
  <Printer size={16} />
  <span>Print</span>
</button>
```

## Step 4: Add Dialogs

Add the print dialogs at the end of your component's JSX:

```tsx
return (
  <div>
    {/* Your existing chart JSX */}

    {/* Add these at the end */}
    <PrintSetupDialog
      isOpen={showSetupDialog}
      onClose={closeDialogs}
      onPrint={handleSetupComplete}
      accountId="your-account-id" // Optional
    />

    {showPreviewDialog && chartData && printPreferences && (
      <PrintPreviewDialog
        isOpen={showPreviewDialog}
        onClose={closeDialogs}
        chartData={chartData}
        preferences={printPreferences}
        onPrint={handleDirectPrint}
      />
    )}
  </div>
);
```

## Complete Example

Here's a complete minimal example:

```tsx
import React, { useRef } from 'react';
import { IChartApi } from 'lightweight-charts';
import { Printer } from 'lucide-react';
import { useChartPrint } from '../hooks/useChartPrint';
import { PrintSetupDialog } from './dialogs/PrintSetupDialog';
import { PrintPreviewDialog } from './dialogs/PrintPreviewDialog';

export function MyTradingChart() {
  const chartApiRef = useRef<IChartApi | null>(null);
  const symbol = 'EURUSD';
  const timeframe = '1H';

  const {
    showSetupDialog,
    showPreviewDialog,
    printPreferences,
    chartData,
    openPrintSetup,
    handleSetupComplete,
    handleDirectPrint,
    closeDialogs,
  } = useChartPrint({
    symbol,
    timeframe,
    chartApi: chartApiRef.current,
  });

  return (
    <div className="relative">
      {/* Chart Toolbar */}
      <div className="absolute top-2 right-2 z-10">
        <button
          onClick={openPrintSetup}
          className="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded transition-colors"
        >
          <Printer size={16} />
          Print
        </button>
      </div>

      {/* Your Chart Component */}
      <div ref={chartContainerRef} className="w-full h-full">
        {/* Chart renders here */}
      </div>

      {/* Print Dialogs */}
      <PrintSetupDialog
        isOpen={showSetupDialog}
        onClose={closeDialogs}
        onPrint={handleSetupComplete}
      />

      {showPreviewDialog && chartData && printPreferences && (
        <PrintPreviewDialog
          isOpen={showPreviewDialog}
          onClose={closeDialogs}
          chartData={chartData}
          preferences={printPreferences}
          onPrint={handleDirectPrint}
        />
      )}
    </div>
  );
}
```

## With Indicators and Drawings (Advanced)

If your chart has indicators and drawings:

```tsx
const {
  // ... other hooks
} = useChartPrint({
  symbol,
  timeframe,
  chartApi: chartApiRef.current,
  // Add these functions
  getIndicators: () => {
    return indicatorManager.getActiveIndicators().map(ind => ({
      name: ind.name,
      parameters: ind.parameters,
    }));
  },
  getDrawings: () => {
    return drawingManager.getActiveDrawings().map(drawing => ({
      type: drawing.type,
      label: drawing.label,
    }));
  },
});
```

## Styling Options

### Minimal Button

```tsx
<button onClick={openPrintSetup} title="Print">
  <Printer size={16} />
</button>
```

### Icon-only Button

```tsx
<button
  onClick={openPrintSetup}
  className="p-2 hover:bg-gray-700 rounded"
  title="Print Chart"
>
  <Printer size={18} />
</button>
```

### With Keyboard Shortcut

```tsx
useEffect(() => {
  const handleKeyPress = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'p') {
      e.preventDefault();
      openPrintSetup();
    }
  };

  window.addEventListener('keydown', handleKeyPress);
  return () => window.removeEventListener('keydown', handleKeyPress);
}, [openPrintSetup]);
```

## Troubleshooting

### Chart API is null
Make sure you set the chartApiRef after creating the chart:

```tsx
useEffect(() => {
  const chart = createChart(containerRef.current);
  chartApiRef.current = chart;

  return () => {
    chart.remove();
    chartApiRef.current = null;
  };
}, []);
```

### Dialogs not showing
Check that you've imported both dialog components and they're rendered in your JSX.

### Print button not visible
Ensure the button has proper z-index and positioning to appear above the chart.

## Next Steps

- See [CHART_PRINT_FUNCTIONALITY.md](./CHART_PRINT_FUNCTIONALITY.md) for full documentation
- Check [ChartPrintExample.tsx](../clients/desktop/src/components/ChartPrintExample.tsx) for a complete working example
- Review print preferences customization options
- Test printing on different browsers

## Support

If you encounter issues:
1. Check browser console for errors
2. Verify all imports are correct
3. Ensure chart API is initialized before printing
4. Test with the provided example component first
