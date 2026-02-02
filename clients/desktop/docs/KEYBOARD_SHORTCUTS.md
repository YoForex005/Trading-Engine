# Keyboard Shortcut System

## Overview

MT5-compatible global keyboard shortcut system with conflict detection, input protection, and visual feedback.

## Architecture

### Core Components

1. **KeyboardShortcutManager** (`services/keyboardShortcutManager.ts`)
   - Singleton service managing all shortcuts
   - Priority-based conflict resolution
   - Input field protection
   - Event delegation and cleanup

2. **KeyboardShortcutContext** (`contexts/KeyboardShortcutContext.tsx`)
   - React Context for global access
   - Enable/disable shortcuts
   - Query shortcuts by category
   - Conflict detection

3. **Hooks** (`hooks/useKeyboardShortcut.ts`)
   - `useKeyboardShortcut` - Register single shortcut
   - `useKeyboardShortcuts` - Register multiple shortcuts
   - `useShortcutList` - Get all shortcuts
   - `useShortcutsByCategory` - Filter by category
   - `useShortcutToggle` - Enable/disable dynamically
   - `useShortcutFormat` - Format for display

4. **UI Components**
   - `GlobalShortcuts` - Registers app-wide shortcuts
   - `ShortcutFeedback` - Visual feedback toast
   - `ShortcutHelpDialog` - F1 help dialog
   - `ShortcutHint` - Display shortcuts in menus

## Features

### 1. Input Protection

Shortcuts are automatically disabled when typing in input fields, textareas, or contenteditable elements.

**Exception**: F-keys and Escape always work (even in inputs).

```typescript
// Configure per-shortcut
{
  key: 's',
  ctrl: true,
  allowInInput: false  // Default - disabled in inputs
}

// Force enable in inputs
{
  key: 'Escape',
  allowInInput: true  // Works everywhere
}
```

### 2. Priority-Based Conflict Resolution

When multiple shortcuts use the same key combination, priority determines which takes precedence.

```typescript
{
  key: 's',
  ctrl: true,
  priority: 100  // Higher priority = takes precedence
}
```

**Priority Levels**:
- 100: Critical (File menu, F-keys, Escape)
- 90-95: High priority (View, Charts, Trading)
- 80-85: Medium priority (Navigation, Tools)
- <80: Low priority (Custom shortcuts)

### 3. Modal/Dialog Awareness

Shortcuts respect active modals and dialogs:

```typescript
// Escape closes the topmost modal first
{
  key: 'Escape',
  handler: () => {
    if (isModalOpen) {
      closeModal();
    } else {
      // Other escape behavior
    }
  }
}
```

### 4. Visual Feedback

All shortcuts trigger visual feedback toast showing:
- Shortcut name
- Key combination
- Symbol (for trading actions)

### 5. Help Dialog (F1)

Press F1 to open comprehensive help showing:
- All shortcuts organized by category
- Search functionality
- Conflict warnings
- Enabled/disabled status

## Default Shortcuts (MT5-Compatible)

### File Menu
- `Ctrl+S` - Save workspace
- `Ctrl+Shift+D` - Open data folder
- `Ctrl+P` - Print chart
- `Alt+F4` - Exit application

### View Menu
- `Ctrl+U` - Symbols
- `Alt+B` - Depth of Market
- `Ctrl+M` - Market Watch
- `Ctrl+D` - Data Window
- `Ctrl+N` - Navigator
- `Ctrl+T` - Toolbox
- `Ctrl+R` - Strategy Tester
- `F11` - Full Screen

### Charts Menu
- `Ctrl+I` - Indicator List
- `Ctrl+B` - Object List
- `Alt+1` - Bar Chart
- `Alt+2` - Candlesticks
- `Alt+3` - Line Chart
- `Ctrl+G` - Grid
- `Ctrl+L` - Volumes
- `+` - Zoom In
- `-` - Zoom Out
- `F8` - Chart Properties
- `Backspace` - Delete Last Object
- `Delete` - Delete Selected
- `Ctrl+Z` - Undo

### Tools Menu
- `F9` - New Order
- `Ctrl+O` - Options

### Trading
- `Alt+B` - Quick Buy
- `Alt+S` - Quick Sell

### Navigation
- `Escape` - Close Modal
- `Ctrl+Tab` - Next Tab
- `Ctrl+Shift+Tab` - Previous Tab

### Help
- `F1` - Show Keyboard Shortcuts

## Usage

### 1. Setup (App.tsx)

```typescript
import { KeyboardShortcutProvider } from './contexts/KeyboardShortcutContext';
import { GlobalShortcuts } from './components/GlobalShortcuts';

function App() {
  return (
    <KeyboardShortcutProvider>
      <GlobalShortcuts
        onOpenOrderPanel={handleOpenOrderPanel}
        onQuickBuy={handleQuickBuy}
        onQuickSell={handleQuickSell}
        // ... other handlers
        selectedSymbol={selectedSymbol}
      />
      <YourApp />
    </KeyboardShortcutProvider>
  );
}
```

### 2. Register Custom Shortcuts in Components

```typescript
import { useKeyboardShortcut } from '../hooks/useKeyboardShortcut';

function MyComponent() {
  useKeyboardShortcut({
    key: 's',
    ctrl: true,
    shift: true,
    description: 'Save as',
    category: 'File',
    priority: 90,
    handler: () => {
      console.log('Save as triggered');
    }
  });

  return <div>My Component</div>;
}
```

### 3. Register Multiple Shortcuts

```typescript
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcut';

function ChartComponent() {
  useKeyboardShortcuts([
    {
      key: '+',
      description: 'Zoom In',
      handler: () => zoomIn()
    },
    {
      key: '-',
      description: 'Zoom Out',
      handler: () => zoomOut()
    }
  ]);
}
```

### 4. Display Shortcuts in Menus

```typescript
import { useShortcutHint } from '../contexts/KeyboardShortcutContext';

function MenuItem() {
  const hint = useShortcutHint('file.save');

  return (
    <button>
      Save
      {hint && <span className="text-zinc-500">{hint}</span>}
    </button>
  );
}
```

### 5. Conditionally Enable/Disable

```typescript
import { useShortcutToggle } from '../hooks/useKeyboardShortcut';

function ConditionalComponent({ canSave }) {
  // Enable/disable shortcut based on state
  useShortcutToggle('file.save', canSave);

  return <div>...</div>;
}
```

### 6. Temporarily Disable All Shortcuts

```typescript
import { useDisableShortcuts } from '../contexts/KeyboardShortcutContext';

function InputHeavyDialog() {
  // Disables all shortcuts while this component is mounted
  useDisableShortcuts();

  return (
    <dialog>
      <input type="text" />
      {/* Shortcuts won't interfere with typing */}
    </dialog>
  );
}
```

## Edge Cases Handled

### 1. Multiple Windows
- Shortcuts work in the focused window only
- No cross-window interference

### 2. Modal Dialogs
- Escape closes topmost modal first
- Other shortcuts respect modal state

### 3. Input Focus
- Shortcuts disabled when typing (except F-keys/Escape)
- ContentEditable elements detected
- Select dropdowns protected

### 4. Rapid Key Presses
- Event delegation prevents duplicate triggers
- Debouncing built-in

### 5. Cleanup
- Automatic unregister on component unmount
- No memory leaks

### 6. Platform Differences
- Ctrl vs Command (Mac) handled automatically
- Windows/Linux/Mac compatible

## Conflict Detection

The system automatically detects and logs conflicts:

```typescript
const { hasConflicts, getConflicts } = useKeyboardShortcutContext();

if (hasConflicts()) {
  const conflicts = getConflicts();
  console.warn('Shortcut conflicts:', conflicts);
}
```

Conflicts are resolved by priority:
- Higher priority wins
- Equal priority = first registered wins
- Warning logged to console

## Performance

- **Registration**: O(1) lookup via Map
- **Event handling**: Single global listener with delegation
- **Memory**: Minimal overhead (~1KB per shortcut)
- **Cleanup**: Automatic on unmount

## Testing

```typescript
// Test shortcut registration
import { keyboardShortcutManager } from '../services/keyboardShortcutManager';

test('registers shortcut', () => {
  const unregister = keyboardShortcutManager.register({
    id: 'test',
    key: 's',
    ctrl: true,
    description: 'Test',
    handler: jest.fn()
  });

  expect(keyboardShortcutManager.findShortcut('test')).toBeDefined();
  unregister();
  expect(keyboardShortcutManager.findShortcut('test')).toBeUndefined();
});
```

## Migration from Old System

**Before** (App.tsx inline handlers):
```typescript
useEffect(() => {
  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'F9') {
      // ...
    }
  };
  window.addEventListener('keydown', handleKeyDown);
  return () => window.removeEventListener('keydown', handleKeyDown);
}, [/* many deps */]);
```

**After** (New system):
```typescript
<GlobalShortcuts
  onOpenOrderPanel={handleOpenOrderPanel}
  selectedSymbol={selectedSymbol}
/>
```

**Benefits**:
- Centralized management
- No dependency arrays
- Automatic cleanup
- Visual feedback
- Help dialog
- Conflict detection

## Future Enhancements

1. **Customization UI**: Let users remap shortcuts
2. **Profiles**: Save/load shortcut sets
3. **Chords**: Multi-key sequences (e.g., Ctrl+K Ctrl+S)
4. **Recording**: Record macro sequences
5. **Cloud Sync**: Sync shortcuts across devices
6. **Analytics**: Track most-used shortcuts

## Troubleshooting

### Shortcut Not Working

1. Check if input is focused (shortcuts disabled by default)
2. Check priority (lower priority might be overridden)
3. Open help dialog (F1) to verify registration
4. Check browser console for conflict warnings

### Conflict Warnings

```typescript
// Increase priority to override
{
  key: 's',
  ctrl: true,
  priority: 100  // Higher = wins
}
```

### Memory Leaks

- Ensure components using shortcuts are properly unmounted
- Check for circular references in handlers
- Use cleanup functions returned by hooks

## API Reference

See inline JSDoc comments in:
- `services/keyboardShortcutManager.ts`
- `hooks/useKeyboardShortcut.ts`
- `contexts/KeyboardShortcutContext.tsx`
