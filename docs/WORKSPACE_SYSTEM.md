# Workspace Save/Load System

## Overview

The workspace save/load system provides comprehensive state persistence for the trading platform, allowing users to save and restore complete workspace configurations including charts, indicators, drawings, layouts, and settings.

## Architecture

### Frontend Components

#### 1. Type Definitions (`src/types/workspace.ts`)

Complete TypeScript type definitions for:
- **Workspace**: Main workspace configuration
- **ChartState**: Individual chart configurations with indicators and drawings
- **LayoutConfig**: Panel visibility and window positions
- **MarketWatchConfig**: Market watch symbols and settings
- **OrderPanelSettings**: Order entry panel configuration
- **AutoSaveState**: Auto-save status tracking

#### 2. Workspace Manager (`src/services/workspaceManager.ts`)

Core service responsible for:
- **State Capture**: Extracts current workspace state from DOM/store
- **Serialization**: Converts workspace state to JSON
- **Validation**: Ensures workspace data integrity
- **Auto-save**: Automatic periodic saves with change detection
- **State Application**: Restores workspace configuration to UI

**Key Methods:**
```typescript
captureWorkspaceState(name, description, userId): Promise<Workspace>
saveWorkspace(name?, description?, overwrite?): Promise<void>
loadWorkspace(workspaceId): Promise<void>
```

#### 3. Workspace API (`src/services/workspaceApi.ts`)

HTTP client for backend communication:
- Save workspace with conflict detection
- Load workspace by ID
- List all workspaces
- Delete workspace
- Duplicate workspace
- Set default workspace
- Import/export workspace as JSON

**Endpoints:**
```
POST   /api/workspaces              - Save workspace
GET    /api/workspaces/:id          - Load workspace
GET    /api/workspaces              - List workspaces
DELETE /api/workspaces/:id          - Delete workspace
POST   /api/workspaces/:id/duplicate - Duplicate workspace
POST   /api/workspaces/:id/set-default - Set as default
```

#### 4. React Hook (`src/hooks/useWorkspace.ts`)

Convenient React hook providing:
- Workspace state management
- Loading/saving operations
- Error handling
- Auto-save control
- Change detection

**Usage:**
```typescript
const {
  currentWorkspace,
  workspaceList,
  autoSaveState,
  saveWorkspace,
  loadWorkspace,
  hasUnsavedChanges
} = useWorkspace();
```

#### 5. UI Component (`src/components/WorkspaceManager.tsx`)

User interface for workspace management:
- Current workspace display with save button
- Workspace list with search
- Create new workspace
- Load/delete/duplicate operations
- Auto-save indicator
- Error notifications

### Backend Components

#### 1. API Handlers (`backend/api/workspaces.go`)

Go HTTP handlers for workspace operations:

**Key Handlers:**
- `HandleSaveWorkspace`: Atomic save with transaction support
- `HandleLoadWorkspace`: Retrieve workspace by ID
- `HandleListWorkspaces`: Query workspaces with filters
- `HandleDeleteWorkspace`: Remove workspace

**Features:**
- Conflict detection (optimistic locking)
- Atomic transactions for data integrity
- JSON validation
- User isolation

#### 2. Database Schema (`backend/migrations/001_create_workspaces_table.sql`)

SQLite schema with:
- Primary workspace table
- JSON columns for state storage
- Indexes for performance
- Triggers for automatic timestamps
- Sample default workspace

**Table Structure:**
```sql
CREATE TABLE workspaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    user_id TEXT NOT NULL,
    account_id TEXT,
    charts TEXT NOT NULL,           -- JSON
    layout TEXT NOT NULL,            -- JSON
    market_watch TEXT NOT NULL,      -- JSON
    order_panel TEXT NOT NULL,       -- JSON
    version TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    thumbnail TEXT,
    is_default INTEGER,
    tags TEXT                        -- JSON
);
```

## Features

### 1. Auto-Save

Automatic periodic saves with:
- **Change Detection**: Monitors workspace state for modifications
- **Debouncing**: 30-second delay after last change
- **Visual Indicator**: Shows save status and last saved time
- **Error Recovery**: Retries on network failures
- **User Control**: Enable/disable and configure interval

### 2. Conflict Resolution

Handles concurrent modifications:
- **Optimistic Locking**: Timestamp-based conflict detection
- **Conflict Dialog**: Presents user with merge options
- **Overwrite Control**: Optional force-save parameter

### 3. State Capture

Comprehensive state extraction:
- **Charts**: Symbol, timeframe, chart type
- **Indicators**: Type, parameters, visibility
- **Drawings**: Trend lines, Fibonacci, annotations
- **Layout**: Panel positions and visibility
- **Settings**: User preferences and defaults

### 4. Validation

Multi-level validation:
- **Frontend**: Real-time validation before save
- **Backend**: Server-side validation with SQL constraints
- **JSON Schema**: Validates JSON structure
- **Business Rules**: Name length, chart count limits

### 5. Error Handling

Robust error management:
- **Network Errors**: Retry with exponential backoff
- **Validation Errors**: Detailed error messages
- **Conflicts**: Merge conflict resolution UI
- **Unauthorized**: Automatic re-authentication
- **Rollback**: Transaction rollback on failure

## Integration Guide

### 1. Backend Integration

Add workspace routes to your HTTP server:

```go
// In your main server setup
func RegisterRoutes(mux *http.ServeMux, server *api.Server) {
    // Existing routes...

    // Workspace routes
    mux.HandleFunc("/api/workspaces", server.HandleListWorkspaces)
    mux.HandleFunc("/api/workspaces/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/workspaces") {
            server.HandleSaveWorkspace(w, r)
        } else if r.Method == http.MethodGet {
            server.HandleLoadWorkspace(w, r)
        } else if r.Method == http.MethodDelete {
            server.HandleDeleteWorkspace(w, r)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })
}
```

Connect database:

```go
// In your Server struct, add DB field
type Server struct {
    // ... existing fields
    db *sql.DB
}

// Implement getDB method
func (s *Server) getDB() (*sql.DB, error) {
    if s.db == nil {
        return nil, errors.New("database not initialized")
    }
    return s.db, nil
}
```

Run migration:

```bash
sqlite3 your_database.db < backend/migrations/001_create_workspaces_table.sql
```

### 2. Frontend Integration

Add workspace manager to your app:

```tsx
import { WorkspaceManager } from './components/WorkspaceManager';
import { useWorkspace } from './hooks/useWorkspace';

function App() {
  const [showWorkspaceManager, setShowWorkspaceManager] = useState(false);
  const { autoSaveState, hasUnsavedChanges } = useWorkspace();

  return (
    <div>
      {/* Your existing UI */}

      {/* Workspace menu item */}
      <button onClick={() => setShowWorkspaceManager(true)}>
        Workspaces
        {hasUnsavedChanges && <span className="badge">*</span>}
      </button>

      {/* Workspace manager dialog */}
      {showWorkspaceManager && (
        <WorkspaceManager onClose={() => setShowWorkspaceManager(false)} />
      )}

      {/* Auto-save indicator */}
      {autoSaveState.saving && (
        <div className="saving-indicator">Saving...</div>
      )}
    </div>
  );
}
```

### 3. Chart Integration

Register chart state with workspace manager:

```tsx
// In your chart component
useEffect(() => {
  // Store chart data globally for workspace capture
  window.__chartIndicators = window.__chartIndicators || {};
  window.__chartIndicators[chartId] = indicators;

  window.__chartDrawings = window.__chartDrawings || {};
  window.__chartDrawings[chartId] = drawings;

  window.__chartSettings = window.__chartSettings || {};
  window.__chartSettings[chartId] = settings;

  // Trigger change event for auto-save
  window.dispatchEvent(new CustomEvent('workspace:change'));
}, [chartId, indicators, drawings, settings]);

// Listen for workspace restore
useEffect(() => {
  const handleApplyChart = (event: CustomEvent) => {
    const chartState = event.detail;
    // Apply chart state to your chart
    setSymbol(chartState.symbol);
    setTimeframe(chartState.timeframe);
    setIndicators(chartState.indicators);
    setDrawings(chartState.drawings);
    // ... etc
  };

  window.addEventListener('workspace:apply-chart', handleApplyChart);
  return () => window.removeEventListener('workspace:apply-chart', handleApplyChart);
}, []);
```

### 4. Layout Integration

Store layout configuration:

```tsx
// In your layout manager
useEffect(() => {
  window.__layoutConfig = {
    theme: currentTheme,
    panels: {
      marketWatch: { visible: marketWatchVisible, width: marketWatchWidth },
      // ... other panels
    },
    windows: floatingWindows,
    splitRatios: { leftSidebar, rightSidebar, bottomPanel }
  };

  window.dispatchEvent(new CustomEvent('workspace:change'));
}, [currentTheme, panels, windows]);

// Listen for layout restore
useEffect(() => {
  const handleLayoutUpdate = (event: CustomEvent) => {
    const layout = event.detail;
    setTheme(layout.theme);
    setPanels(layout.panels);
    setWindows(layout.windows);
    // ... apply layout
  };

  window.addEventListener('workspace:layout-updated', handleLayoutUpdate);
  return () => window.removeEventListener('workspace:layout-updated', handleLayoutUpdate);
}, []);
```

## API Reference

### Workspace Manager API

```typescript
// Capture current state
const workspace = await workspaceManager.captureWorkspaceState(
  'My Workspace',
  'Description',
  'user123'
);

// Save workspace
await workspaceManager.saveWorkspace('Updated Name', 'New description');

// Load workspace
await workspaceManager.loadWorkspace('ws_12345');

// Get current workspace
const current = workspaceManager.getCurrentWorkspace();

// Auto-save control
workspaceManager.setAutoSaveEnabled(true);
workspaceManager.setAutoSaveInterval(30000); // 30 seconds

// Listen for changes
const unsubscribe = workspaceManager.addChangeListener((hasChanges) => {
  console.log('Workspace changed:', hasChanges);
});
```

### Workspace API

```typescript
// Save workspace
const response = await workspaceApi.saveWorkspace({
  workspace: workspaceData,
  overwrite: true
});

// Load workspace
const response = await workspaceApi.loadWorkspace('ws_12345');

// List workspaces
const response = await workspaceApi.listWorkspaces('user123');

// Delete workspace
const response = await workspaceApi.deleteWorkspace('ws_12345');

// Duplicate workspace
const response = await workspaceApi.duplicateWorkspace('ws_12345', 'Copy of My Workspace');

// Set default
const response = await workspaceApi.setDefaultWorkspace('ws_12345');

// Export/Import
const workspaceJson = await workspaceApi.exportWorkspace('ws_12345');
await workspaceApi.importWorkspace(workspaceJson);
```

## Error Handling

### Common Errors

1. **NETWORK_ERROR**: Connection failed
   - Auto-retry with backoff
   - Show offline indicator
   - Queue operations for retry

2. **VALIDATION_ERROR**: Invalid data
   - Show validation errors
   - Prevent save until fixed
   - Guide user to corrections

3. **CONFLICT**: Concurrent modification
   - Present conflict dialog
   - Show both versions
   - Offer merge options

4. **UNAUTHORIZED**: Authentication failed
   - Clear auth token
   - Redirect to login
   - Preserve unsaved changes

5. **NOT_FOUND**: Workspace doesn't exist
   - Show error message
   - Offer to create new
   - List available workspaces

### Error Recovery

```typescript
try {
  await saveWorkspace();
} catch (error) {
  if (error.code === 'NETWORK_ERROR') {
    // Retry logic
    await retryWithBackoff(() => saveWorkspace());
  } else if (error.code === 'CONFLICT') {
    // Show conflict resolution UI
    showConflictDialog(error.details.conflictVersion);
  } else {
    // Show error notification
    showError(error.message);
  }
}
```

## Performance Considerations

### 1. Change Detection

- Lightweight state snapshots
- Debounced change events
- Shallow comparison where possible

### 2. Serialization

- Stream large JSON objects
- Compress thumbnail images
- Lazy load workspace list

### 3. Database

- Indexed queries for fast retrieval
- Pagination for large lists
- Background cleanup of old workspaces

### 4. Auto-Save

- Configurable interval (default 30s)
- Only saves when changes detected
- Skips save if already saving
- Cancels pending save on manual save

## Security

### 1. Authentication

- JWT token required for all operations
- User isolation enforced at DB level
- Token refresh on expiry

### 2. Authorization

- Users can only access their own workspaces
- Admin role for global access (optional)
- Account-level isolation (optional)

### 3. Validation

- Input sanitization
- SQL injection prevention (parameterized queries)
- JSON validation
- Size limits on workspace data

### 4. Data Protection

- HTTPS required in production
- Encrypted storage (optional)
- Audit logging (optional)
- Backup and recovery procedures

## Testing

### Unit Tests

```typescript
// Test workspace capture
test('captures workspace state', async () => {
  const workspace = await workspaceManager.captureWorkspaceState('Test', '', 'user1');
  expect(workspace.name).toBe('Test');
  expect(workspace.charts).toBeDefined();
});

// Test validation
test('validates workspace name', () => {
  const result = validateWorkspace({ name: '', charts: [] });
  expect(result.valid).toBe(false);
  expect(result.errors).toContain('Workspace name is required');
});

// Test auto-save
test('auto-saves after changes', async () => {
  jest.useFakeTimers();
  triggerChange();
  jest.advanceTimersByTime(30000);
  expect(saveWorkspace).toHaveBeenCalled();
});
```

### Integration Tests

```bash
# Test save/load cycle
curl -X POST http://localhost:7999/api/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"workspace": {...}, "overwrite": true}'

# Test list
curl http://localhost:7999/api/workspaces?userId=user123 \
  -H "Authorization: Bearer $TOKEN"

# Test delete
curl -X DELETE http://localhost:7999/api/workspaces/ws_12345 \
  -H "Authorization: Bearer $TOKEN"
```

## Future Enhancements

1. **Cloud Sync**: Sync workspaces across devices
2. **Version History**: Track workspace changes over time
3. **Sharing**: Share workspaces with other users
4. **Templates**: Pre-built workspace templates
5. **Import/Export**: Import from other platforms (MT4, cTrader)
6. **Collaborative Editing**: Real-time collaborative workspaces
7. **Workspace Analytics**: Usage tracking and recommendations
8. **Backup/Restore**: Automated backup and point-in-time restore

## Troubleshooting

### Issue: Auto-save not working

**Solution:**
1. Check auto-save is enabled: `workspaceManager.setAutoSaveEnabled(true)`
2. Verify change events are firing: `window.dispatchEvent(new CustomEvent('workspace:change'))`
3. Check console for errors

### Issue: Workspace not loading

**Solution:**
1. Verify workspace ID is correct
2. Check user has permission
3. Inspect network tab for 404/403 errors
4. Verify database contains workspace

### Issue: Charts not restoring correctly

**Solution:**
1. Ensure chart components listen for `workspace:apply-chart` event
2. Verify chart state is being stored in `window.__chartIndicators`, etc.
3. Check chart IDs match between save and restore

### Issue: Conflict errors on save

**Solution:**
1. Use `overwrite: true` to force save
2. Implement conflict resolution UI
3. Reload workspace before editing
4. Increase auto-save frequency

## Support

For issues or questions:
- Check documentation at `/docs/WORKSPACE_SYSTEM.md`
- Review code comments in source files
- Contact development team
- File issue on project repository
