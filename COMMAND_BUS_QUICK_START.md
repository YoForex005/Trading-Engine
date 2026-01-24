# Command Bus Quick Start

## Installation (Already Done!)

✅ App.tsx is already wrapped with `CommandBusProvider`

## Basic Usage

### 1. Dispatch a Command
```typescript
import { useCommandDispatch } from '../hooks/useCommandBus';

function MyComponent() {
  const dispatch = useCommandDispatch();

  return (
    <button onClick={() => dispatch({
      type: 'SET_CHART_TYPE',
      payload: { chartType: 'line' }
    })}>
      Line Chart
    </button>
  );
}
```

### 2. Listen for Commands
```typescript
import { useCommandListener } from '../hooks/useCommandBus';

function MyComponent() {
  useCommandListener('SET_CHART_TYPE', (payload) => {
    console.log('Chart type changed to:', payload.chartType);
  });

  return <div>Chart type listener active</div>;
}
```

### 3. View & Replay History
```typescript
import { useCommandHistory } from '../hooks/useCommandBus';

function DebugPanel() {
  const { getHistory, replay, clear } = useCommandHistory();

  return (
    <>
      <button onClick={() => console.table(getHistory())}>
        Show History
      </button>
      <button onClick={() => replay()}>Replay All</button>
      <button onClick={clear}>Clear</button>
    </>
  );
}
```

## All Available Hooks

### useCommandBus()
Full-featured hook with everything
```typescript
const { dispatch, useCommandSubscription, getHistory, replay } = useCommandBus();
```

### useCommandDispatch()
Just send commands
```typescript
const dispatch = useCommandDispatch();
```

### useCommandListener()
Just receive commands
```typescript
useCommandListener('ZOOM_IN', () => console.log('Zoomed'));
```

### useCommandHistory()
Manage history
```typescript
const { getHistory, clear, replay } = useCommandHistory();
```

### useCommandBusContext()
Access bus directly from context
```typescript
const bus = useCommandBusContext();
bus.dispatch({ type: 'ZOOM_IN', payload: {} });
```

## All Supported Commands

```typescript
// Order Management
'OPEN_ORDER_PANEL'           // payload: { symbol, price: { bid, ask } }

// Chart Display
'SET_CHART_TYPE'             // payload: { chartType }
'SET_TIMEFRAME'              // payload: { timeframe }

// Chart Tools
'TOGGLE_CROSSHAIR'           // payload: {}
'ZOOM_IN'                    // payload: {}
'ZOOM_OUT'                   // payload: {}
'SELECT_TOOL'                // payload: { tool }

// Indicators
'OPEN_INDICATOR_NAVIGATOR'   // payload: {}
'ADD_INDICATOR'              // payload: { name, params }

// Drawing
'SAVE_DRAWING'               // payload: { type, points }
'DELETE_DRAWING'             // payload: { id }

// Window Management
'TILE_WINDOWS'               // payload: { mode }
```

## Common Patterns

### Toolbar Button
```typescript
const dispatch = useCommandDispatch();
<button onClick={() => dispatch({
  type: 'SET_CHART_TYPE',
  payload: { chartType: 'candlestick' }
})}>
  Candlestick
</button>
```

### Real-time Display
```typescript
const [tool, setTool] = useState('cursor');
useCommandListener('SELECT_TOOL', (p) => setTool(p.tool));
return <span>Current Tool: {tool}</span>;
```

### Multi-Handler Setup
```typescript
const { useCommandSubscription } = useCommandBus();

useCommandSubscription('ZOOM_IN', () => updateChart());
useCommandSubscription('ZOOM_IN', () => updateIndicators());
useCommandSubscription('ZOOM_IN', () => updateDrawings());
```

### Context Integration
```typescript
const bus = useCommandBusContext();

const handleMenuClick = (action) => {
  switch(action) {
    case 'zoom':
      bus.dispatch({ type: 'ZOOM_IN', payload: {} });
      break;
    case 'tool':
      bus.dispatch({ type: 'SELECT_TOOL', payload: { tool: 'trendline' } });
      break;
  }
};
```

## Type Safety

Everything is type-checked:

```typescript
// ✅ Correct
dispatch({ type: 'SET_CHART_TYPE', payload: { chartType: 'line' } });

// ❌ Error: 'invalid' is not assignable to type 'candlestick | bar | line | area'
dispatch({ type: 'SET_CHART_TYPE', payload: { chartType: 'invalid' } });

// ❌ Error: ZOOM_IN expects empty payload
dispatch({ type: 'ZOOM_IN', payload: { chartType: 'line' } });
```

## Files Reference

| File | Purpose |
|------|---------|
| `types/commands.ts` | Type definitions |
| `services/commandBus.ts` | Core implementation |
| `hooks/useCommandBus.ts` | React hooks |
| `contexts/CommandBusContext.tsx` | Context provider |
| `examples/CommandBusExample.tsx` | Usage examples |

## Debugging

### View History in Console
```javascript
// Browser console
import { commandBus } from './services/commandBus';
console.table(commandBus.getHistory());
```

### Replay Commands
```javascript
// Browser console
await commandBus.replay(commandBus.getHistory());
```

### Get Subscriber Count
```javascript
// Browser console
commandBus.getSubscriberCount('ZOOM_IN');
```

## Keyboard Integration Example

```typescript
function useKeyboardCommandBus() {
  const dispatch = useCommandDispatch();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === '+') {
        dispatch({ type: 'ZOOM_IN', payload: {} });
      } else if (e.key === '-') {
        dispatch({ type: 'ZOOM_OUT', payload: {} });
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [dispatch]);
}
```

## Context Menu Integration Example

```typescript
function ContextMenu() {
  const dispatch = useCommandDispatch();

  const menu = [
    { label: 'Line Chart', action: () => dispatch({
      type: 'SET_CHART_TYPE',
      payload: { chartType: 'line' }
    })},
    { label: 'Zoom In', action: () => dispatch({
      type: 'ZOOM_IN',
      payload: {}
    })},
  ];

  return (
    <ul>
      {menu.map(item => (
        <li key={item.label} onClick={item.action}>
          {item.label}
        </li>
      ))}
    </ul>
  );
}
```

## Best Practices

1. **Use the right hook for the job:**
   - Dispatch only → `useCommandDispatch()`
   - Listen only → `useCommandListener()`
   - Complex → `useCommandBus()`

2. **Always cleanup:**
   ```typescript
   const unsubscribe = commandBus.subscribe(...);
   // Later
   unsubscribe();
   ```

3. **Handle payloads correctly:**
   ```typescript
   // Ensure payload matches command type
   useCommandListener('SET_CHART_TYPE', (payload: { chartType: string }) => {
     // payload is properly typed
   });
   ```

4. **Test with replay:**
   ```typescript
   // In tests
   await commandBus.replay([
     { type: 'SET_CHART_TYPE', payload: { chartType: 'line' } },
     { type: 'ZOOM_IN', payload: {} }
   ]);
   ```

## Need Help?

See full documentation: `COMMAND_BUS_GUIDE.md`

See examples: `clients/desktop/src/examples/CommandBusExample.tsx`

View tests: `clients/desktop/src/test/commandBus.test.ts`
