# Command Bus Architecture Guide

## Overview

The Command Bus is a centralized, type-safe event dispatching system for the RTX Trading Terminal. It provides a single source of truth for all toolbar commands, with full support for event replay, history tracking, and clean subscription management.

**Key Features:**
- Full TypeScript type safety with discriminated unions
- Centralized pub/sub dispatching
- Command history (last 100 commands)
- Event replay capability for debugging
- Automatic cleanup (no memory leaks)
- React context integration

## Architecture

### File Structure

```
clients/desktop/src/
├── types/
│   └── commands.ts              # Type definitions (Command, CommandBus interface)
├── services/
│   └── commandBus.ts            # CommandBus implementation (singleton)
├── hooks/
│   └── useCommandBus.ts         # React hooks for command bus
├── contexts/
│   └── CommandBusContext.tsx    # React context provider
├── examples/
│   └── CommandBusExample.tsx    # Usage examples
├── test/
│   └── commandBus.test.ts       # Unit tests
└── App.tsx                       # Wrapped with CommandBusProvider
```

### Core Components

#### 1. Type Definitions (`types/commands.ts`)

```typescript
type Command =
  | { type: 'SET_CHART_TYPE'; payload: { chartType: 'candlestick' | 'bar' | 'line' | 'area' } }
  | { type: 'OPEN_ORDER_PANEL'; payload: { symbol: string; price: { bid: number; ask: number } } }
  // ... other command types

type CommandType = Command['type'];  // Discriminated union of all command types
type PayloadFor<T> = Extract<Command, { type: T }>['payload'];  // Extract payload for a type
```

#### 2. Command Bus Service (`services/commandBus.ts`)

Singleton instance implementing the pub/sub pattern:

```typescript
interface CommandBus {
  dispatch(command: Command): void;
  subscribe<T extends CommandType>(type: T, handler: CommandHandler<T>): Unsubscribe;
  getHistory(): Command[];
  clearHistory(): void;
  replay(commands: Command[]): Promise<void>;
}

export const commandBus = new CommandBusImpl();
```

#### 3. React Hooks (`hooks/useCommandBus.ts`)

Multiple hooks for different use cases:

```typescript
// Full feature hook
useCommandBus() => {
  dispatch(command: Command): void;
  useCommandSubscription<T>(type: T, handler: CommandHandler<T>): void;
  getHistory(): Command[];
  replay(commands: Command[]): Promise<void>;
}

// Simple dispatch only
useCommandDispatch() => (command: Command) => void;

// Simple listener
useCommandListener(type: CommandType, handler: CommandHandler): void;

// History management
useCommandHistory() => {
  getHistory(): Command[];
  clear(): void;
  replay(commands?: Command[]): Promise<void>;
}
```

#### 4. React Context (`contexts/CommandBusContext.tsx`)

Makes the command bus available to all components:

```typescript
<CommandBusProvider>
  <YourApp />
</CommandBusProvider>

// Inside any component:
const bus = useCommandBusContext();
```

## Usage Examples

### Example 1: Dispatch a Command

```typescript
import { useCommandDispatch } from '../hooks/useCommandBus';

function ChartTypeSelector() {
  const dispatch = useCommandDispatch();

  const handleSelectLine = () => {
    dispatch({
      type: 'SET_CHART_TYPE',
      payload: { chartType: 'line' },
    });
  };

  return <button onClick={handleSelectLine}>Line Chart</button>;
}
```

### Example 2: Listen for Commands

```typescript
import { useCommandListener } from '../hooks/useCommandBus';

function ChartTypeDisplay() {
  const [chartType, setChartType] = useState('candlestick');

  useCommandListener('SET_CHART_TYPE', (payload) => {
    setChartType(payload.chartType);
  });

  return <p>Current chart: {chartType}</p>;
}
```

### Example 3: Full Hook with Multiple Features

```typescript
import { useCommandBus } from '../hooks/useCommandBus';

function AdvancedExample() {
  const { dispatch, useCommandSubscription, getHistory, replay } = useCommandBus();

  // Subscribe to multiple command types
  useCommandSubscription('ZOOM_IN', () => {
    console.log('Zoomed in');
  });

  useCommandSubscription('ZOOM_OUT', () => {
    console.log('Zoomed out');
  });

  const handleZoom = (direction: 'in' | 'out') => {
    dispatch({
      type: direction === 'in' ? 'ZOOM_IN' : 'ZOOM_OUT',
      payload: {},
    });
  };

  const handleShowHistory = () => {
    const history = getHistory();
    console.log('Last 5 commands:', history.slice(-5));
  };

  return (
    <>
      <button onClick={() => handleZoom('in')}>Zoom In</button>
      <button onClick={() => handleZoom('out')}>Zoom Out</button>
      <button onClick={handleShowHistory}>Show History</button>
    </>
  );
}
```

### Example 4: Context Integration

```typescript
import { useCommandBusContext } from '../contexts/CommandBusContext';

function ContextExample() {
  const bus = useCommandBusContext();

  const handleAction = () => {
    bus.dispatch({
      type: 'OPEN_ORDER_PANEL',
      payload: {
        symbol: 'EURUSD',
        price: { bid: 1.0950, ask: 1.0952 },
      },
    });
  };

  return <button onClick={handleAction}>Open Order Panel</button>;
}
```

### Example 5: Command History & Replay

```typescript
import { useCommandHistory } from '../hooks/useCommandBus';

function HistoryDebugger() {
  const { getHistory, clear, replay } = useCommandHistory();

  const handleShowHistory = () => {
    const history = getHistory();
    console.table(history);
  };

  const handleReplay = async () => {
    const history = getHistory();
    if (history.length > 0) {
      console.log('Replaying all commands...');
      await replay(history);
    }
  };

  return (
    <>
      <button onClick={handleShowHistory}>Show History</button>
      <button onClick={handleReplay}>Replay All</button>
      <button onClick={clear}>Clear History</button>
    </>
  );
}
```

## Supported Commands

### Order Management
- `OPEN_ORDER_PANEL`: Open order panel with symbol and current prices

### Chart Display
- `SET_CHART_TYPE`: Change chart type (candlestick, bar, line, area)
- `SET_TIMEFRAME`: Change chart timeframe

### Chart Tools
- `TOGGLE_CROSSHAIR`: Show/hide crosshair
- `ZOOM_IN`: Zoom chart in
- `ZOOM_OUT`: Zoom chart out
- `SELECT_TOOL`: Select drawing tool (cursor, trendline, hline, vline, text)

### Indicators
- `OPEN_INDICATOR_NAVIGATOR`: Open indicator selection panel
- `ADD_INDICATOR`: Add indicator to chart

### Drawing
- `SAVE_DRAWING`: Save a drawing/annotation
- `DELETE_DRAWING`: Delete a drawing/annotation

### Window Management
- `TILE_WINDOWS`: Tile windows (horizontal, vertical, grid)

## Type Safety

All commands are fully type-safe using TypeScript's discriminated unions:

```typescript
// ✅ This is type-checked
dispatch({
  type: 'SET_CHART_TYPE',
  payload: { chartType: 'line' },  // Only valid chart types allowed
});

// ❌ TypeScript error - invalid chart type
dispatch({
  type: 'SET_CHART_TYPE',
  payload: { chartType: 'invalid' },  // Error: not assignable
});

// ❌ TypeScript error - wrong payload for command type
dispatch({
  type: 'SET_CHART_TYPE',
  payload: { symbol: 'EURUSD' },  // Error: chartType is required
});
```

## Advanced Features

### 1. Command History

```typescript
const history = commandBus.getHistory();
// Returns last 100 commands in order
// Useful for:
// - Debugging user actions
// - Creating audit trails
// - Testing replay functionality
```

### 2. Event Replay

```typescript
const commands = [
  { type: 'SET_CHART_TYPE', payload: { chartType: 'line' } },
  { type: 'ZOOM_IN', payload: {} },
  { type: 'ZOOM_IN', payload: {} },
];

await commandBus.replay(commands);
// Replays commands with small delays between each
// Useful for:
// - Debugging complex user workflows
// - Testing UI behavior
// - Reproducing issues
```

### 3. Subscription Management

```typescript
// Subscribe returns an unsubscribe function
const unsubscribe = commandBus.subscribe('ZOOM_IN', () => {
  console.log('Zoomed in');
});

// Automatically cleanup when unsubscribed
unsubscribe();  // No more events received
```

### 4. Error Handling

Errors in command handlers are caught and logged:

```typescript
commandBus.subscribe('SET_CHART_TYPE', () => {
  throw new Error('Something went wrong');
  // Error is caught and logged, doesn't break other handlers
});
```

## Performance Considerations

1. **Memory**: History is limited to 100 commands to prevent memory leaks
2. **Cleanup**: Unsubscribe function automatically removes handlers
3. **Batch Updates**: Use requestAnimationFrame if dispatching many commands
4. **WeakMaps**: Internal cleanup uses efficient data structures

## Testing

The command bus is fully testable:

```typescript
it('should dispatch commands', () => {
  const handler = vi.fn();
  const unsubscribe = commandBus.subscribe('ZOOM_IN', handler);

  commandBus.dispatch({ type: 'ZOOM_IN', payload: {} });

  expect(handler).toHaveBeenCalledWith({});
  unsubscribe();
});
```

## Best Practices

1. **Use useCommandDispatch for simple cases:**
   ```typescript
   const dispatch = useCommandDispatch();
   ```

2. **Use useCommandListener for single listeners:**
   ```typescript
   useCommandListener('SET_CHART_TYPE', (payload) => {
     // Handle command
   });
   ```

3. **Use useCommandBus for complex components:**
   ```typescript
   const { dispatch, useCommandSubscription } = useCommandBus();
   ```

4. **Cleanup subscriptions properly:**
   ```typescript
   const unsubscribe = commandBus.subscribe('ZOOM_IN', handler);
   // Later...
   unsubscribe();  // Always clean up!
   ```

5. **Type your payload handlers:**
   ```typescript
   useCommandListener('SET_CHART_TYPE', (payload: { chartType: string }) => {
     // payload is properly typed
   });
   ```

## Debugging

### Enable Development Logging

Development mode logs all commands:

```typescript
// In commandBus.ts (development only)
console.log(`[CommandBus] Dispatching: ${command.type}`, 'payload:', command.payload);
```

### View Command History

```typescript
// In browser console
import { commandBus } from './services/commandBus';
console.table(commandBus.getHistory());
```

### Replay Commands

```typescript
// In browser console
await commandBus.replay(commandBus.getHistory());
```

## Migration Guide

### From Direct State Management

**Before:**
```typescript
const [chartType, setChartType] = useState('candlestick');

const handleChangeChart = (type: string) => {
  setChartType(type);
  // Need to manually sync with other components
};
```

**After:**
```typescript
const dispatch = useCommandDispatch();

const handleChangeChart = (type: string) => {
  dispatch({
    type: 'SET_CHART_TYPE',
    payload: { chartType: type },
  });
  // All subscribers automatically notified
};
```

## Troubleshooting

### "useCommandBusContext must be used within a CommandBusProvider"

**Solution:** Ensure your component is wrapped with `<CommandBusProvider>` in App.tsx

### Commands not being dispatched

**Check:**
1. Component is inside CommandBusProvider
2. Command type is spelled correctly
3. Payload structure matches the Command type

### Memory leaks

**Solution:** Always call the unsubscribe function returned by subscribe():

```typescript
const unsubscribe = commandBus.subscribe(...);
// Later in cleanup
unsubscribe();
```

## Future Enhancements

Possible extensions to the command bus:

1. **Command Middleware**: Transform/intercept commands
2. **Undo/Redo**: History-based undo/redo functionality
3. **Persistence**: Save/load command sequences
4. **Remote Sync**: Sync commands across browser tabs
5. **Time Travel**: Debug mode for replaying at different speeds

## References

- TypeScript Discriminated Unions: https://www.typescriptlang.org/docs/handbook/2/narrowing.html#discriminated-unions
- Pub/Sub Pattern: https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern
- React Context: https://react.dev/reference/react/useContext

## Support

For issues or questions about the Command Bus:

1. Check the examples: `clients/desktop/src/examples/CommandBusExample.tsx`
2. View tests: `clients/desktop/src/test/commandBus.test.ts`
3. Create an issue with details about the problem
