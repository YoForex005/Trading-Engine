# Save Workspace Testing Checklist

## Pre-Testing Setup

- [ ] Backend server is running (`go run ./cmd/server`)
- [ ] Frontend dev server is running (`npm run dev`)
- [ ] Database directory exists (`./data/`)
- [ ] User is logged in to the application

## Functional Tests

### 1. Basic Save Functionality
- [ ] Click File menu > Save OR press Ctrl+S
- [ ] SaveWorkspaceDialog appears
- [ ] Enter workspace name "Test Workspace 1"
- [ ] Click Save button
- [ ] Dialog closes
- [ ] Green toast notification appears: "Workspace 'Test Workspace 1' saved successfully!"
- [ ] Check database file exists: `./data/workspaces.db`

### 2. Keyboard Shortcut (Ctrl+S)
- [ ] Press Ctrl+S
- [ ] Dialog opens immediately
- [ ] Enter workspace name "Keyboard Test"
- [ ] Press Enter key
- [ ] Workspace saves successfully
- [ ] Toast notification appears

### 3. Overwrite Protection
- [ ] Press Ctrl+S
- [ ] Enter existing workspace name "Test Workspace 1"
- [ ] Click Save
- [ ] Overwrite warning appears with amber color
- [ ] Warning message: "Workspace 'Test Workspace 1' already exists"
- [ ] Two buttons visible: "Back" and "Overwrite"
- [ ] Click Back
- [ ] Returns to name input screen
- [ ] Click Save again
- [ ] Click Overwrite
- [ ] Workspace saves successfully

### 4. Empty Name Validation
- [ ] Press Ctrl+S
- [ ] Leave workspace name empty (clear default)
- [ ] Save button is disabled (gray)
- [ ] Cannot submit empty name

### 5. Dialog Cancellation
- [ ] Press Ctrl+S
- [ ] Enter workspace name
- [ ] Click Cancel button
- [ ] Dialog closes without saving
- [ ] No toast notification appears
- [ ] Press Ctrl+S
- [ ] Press Escape key
- [ ] Dialog closes without saving

### 6. Menu Integration
- [ ] Click File menu
- [ ] Click Save menu item
- [ ] Dialog opens
- [ ] Menu closes automatically
- [ ] Save workspace successfully

### 7. Error Handling
- [ ] Stop backend server
- [ ] Press Ctrl+S
- [ ] Enter workspace name
- [ ] Click Save
- [ ] Red error message appears in dialog
- [ ] Error text is clear and actionable
- [ ] Restart backend
- [ ] Click Save again
- [ ] Workspace saves successfully

### 8. Loading States
- [ ] Add network throttling (if possible)
- [ ] Press Ctrl+S
- [ ] Enter workspace name
- [ ] Click Save
- [ ] Loading spinner appears on Save button
- [ ] Button text changes to "Saving..."
- [ ] All inputs are disabled during save
- [ ] After completion, dialog closes

### 9. Workspace Data Persistence
- [ ] Open multiple charts (File > Charts > New Chart Window)
- [ ] Adjust dock height (drag bottom panel)
- [ ] Select different symbol
- [ ] Change volume setting
- [ ] Press Ctrl+S and save as "Full State Test"
- [ ] Verify workspace saved
- [ ] Check localStorage contains workspace data
- [ ] Verify database contains all workspace data

### 10. Multiple User Workspaces
- [ ] Save workspace "User1 Workspace"
- [ ] Logout
- [ ] Login as different user
- [ ] Press Ctrl+S
- [ ] Save workspace "User2 Workspace"
- [ ] Verify both workspaces exist in database
- [ ] Verify workspaces are user-specific

## API Tests (Backend)

### Using curl or Postman

#### 1. Save Workspace API
```bash
curl -X POST http://localhost:7999/api/workspaces \
  -H "Content-Type: application/json" \
  -d '{
    "workspace": {
      "name": "API Test Workspace",
      "userId": "1",
      "accountId": "1",
      "charts": "[{\"id\":\"chart1\",\"symbol\":\"EURUSD\",\"timeframe\":\"1m\"}]",
      "layout": "{\"dockHeight\":250}",
      "marketWatch": "{\"selectedSymbol\":\"EURUSD\"}",
      "orderPanel": "{\"volume\":0.01}",
      "version": "1.0.0",
      "isDefault": false
    },
    "overwrite": false
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "workspaceId": "ws_1234567890",
  "message": "Workspace saved successfully"
}
```

- [ ] Status code: 200 OK
- [ ] Response contains workspaceId
- [ ] Success is true

#### 2. List Workspaces API
```bash
curl http://localhost:7999/api/workspaces?userId=1
```

**Expected Response:**
```json
{
  "success": true,
  "workspaces": [
    {
      "id": "ws_1234567890",
      "name": "API Test Workspace",
      "userId": "1",
      "accountId": "1",
      "chartCount": 1,
      "updatedAt": "2024-01-01T12:00:00Z",
      "isDefault": false,
      "tags": []
    }
  ]
}
```

- [ ] Status code: 200 OK
- [ ] Returns array of workspaces
- [ ] Contains previously saved workspace

#### 3. Load Workspace API
```bash
curl http://localhost:7999/api/workspaces/ws_1234567890
```

**Expected Response:**
```json
{
  "success": true,
  "workspace": {
    "id": "ws_1234567890",
    "name": "API Test Workspace",
    "userId": "1",
    "charts": "[...]",
    "layout": "{...}",
    "marketWatch": "{...}",
    "orderPanel": "{...}",
    "version": "1.0.0"
  }
}
```

- [ ] Status code: 200 OK
- [ ] Returns complete workspace object
- [ ] All fields populated correctly

#### 4. Delete Workspace API
```bash
curl -X DELETE http://localhost:7999/api/workspaces/ws_1234567890
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Workspace deleted successfully"
}
```

- [ ] Status code: 200 OK
- [ ] Success is true
- [ ] Workspace no longer in list

#### 5. Overwrite Protection API
```bash
# Save twice with same workspace ID
curl -X POST http://localhost:7999/api/workspaces \
  -H "Content-Type: application/json" \
  -d '{"workspace": {..., "id": "ws_123"}, "overwrite": false}'
```

**Expected Response (second call):**
```json
{
  "success": false,
  "conflict": true,
  "message": "Workspace has been modified by another session",
  "conflictVersion": {...}
}
```

- [ ] Status code: 409 Conflict
- [ ] Conflict flag is true
- [ ] Returns conflicting version

## Database Tests

### Using SQLite CLI
```bash
sqlite3 ./data/workspaces.db
```

#### 1. Check Table Schema
```sql
.schema workspaces
```

- [ ] Table exists
- [ ] All columns present
- [ ] Indexes created

#### 2. Verify Saved Data
```sql
SELECT id, name, user_id, version FROM workspaces;
```

- [ ] Workspaces visible
- [ ] Data matches saved workspaces

#### 3. Check JSON Fields
```sql
SELECT charts, layout FROM workspaces WHERE name = 'Test Workspace 1';
```

- [ ] JSON data is valid
- [ ] Contains expected structure

## Performance Tests

### 1. Quick Save (Latency)
- [ ] Press Ctrl+S
- [ ] Enter name and save
- [ ] Measure time from click to toast
- [ ] Target: < 500ms

### 2. Large Workspace
- [ ] Open 10+ chart windows
- [ ] Add many drawings/indicators
- [ ] Save workspace
- [ ] Verify all data saved
- [ ] Load time acceptable

### 3. Concurrent Saves
- [ ] Open multiple browser tabs
- [ ] Save different workspaces simultaneously
- [ ] No data corruption
- [ ] All saves succeed

## Edge Cases

### 1. Special Characters in Name
- [ ] Save workspace with name: "Test ! @ # $ % Workspace"
- [ ] Verify saves and loads correctly

### 2. Long Workspace Name
- [ ] Enter 100+ character name
- [ ] Verify saves successfully
- [ ] Name displays properly

### 3. Network Interruption
- [ ] Start save
- [ ] Disconnect network mid-save
- [ ] Verify error handling
- [ ] Reconnect and retry

### 4. Server Restart During Operation
- [ ] Press Ctrl+S
- [ ] Restart backend server
- [ ] Complete save
- [ ] Verify error or success

### 5. Database File Permissions
- [ ] Make database file read-only
- [ ] Try to save workspace
- [ ] Verify error message
- [ ] Restore permissions

## Browser Compatibility

Test in multiple browsers:

### Chrome
- [ ] Save functionality works
- [ ] Keyboard shortcuts work
- [ ] UI renders correctly

### Firefox
- [ ] Save functionality works
- [ ] Keyboard shortcuts work
- [ ] UI renders correctly

### Edge
- [ ] Save functionality works
- [ ] Keyboard shortcuts work
- [ ] UI renders correctly

### Safari (if available)
- [ ] Save functionality works
- [ ] Keyboard shortcuts work
- [ ] UI renders correctly

## Regression Tests

### 1. Existing Features Still Work
- [ ] Login/logout functionality
- [ ] Chart operations
- [ ] Order placement
- [ ] Position management
- [ ] Market watch
- [ ] Other menu items

### 2. No Console Errors
- [ ] Open browser console
- [ ] Perform save operations
- [ ] No errors or warnings logged

### 3. No Memory Leaks
- [ ] Open/save workspaces 20+ times
- [ ] Check memory usage in DevTools
- [ ] Memory usage stable

## Sign-Off

### Test Environment
- OS: _________________
- Browser: _________________
- Backend Version: _________________
- Frontend Version: _________________
- Date: _________________

### Results Summary
- Total Tests: _____
- Passed: _____
- Failed: _____
- Blocked: _____

### Issues Found
1. _________________
2. _________________
3. _________________

### Tester Sign-Off
Name: _________________
Signature: _________________
Date: _________________

### Approval
Approved By: _________________
Date: _________________

## Notes
- Any additional observations
- Performance metrics
- Suggestions for improvement
