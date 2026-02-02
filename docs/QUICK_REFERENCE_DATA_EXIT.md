# Quick Reference: Data Folder & Exit Features

## Keyboard Shortcuts

| Shortcut | Action | Platform |
|----------|--------|----------|
| `Ctrl+Shift+D` | Open Data Folder | All |
| `Esc` | Close Dialog | All |
| Browser Close | Exit Confirmation | Web |

## Data Folder Access

### Open Data Folder
**Keyboard**: `Ctrl+Shift+D`

**What happens**:
- **Web/Desktop**: Opens virtual file explorer dialog
- **Electron**: Opens OS file explorer (Windows Explorer, macOS Finder, etc.)

### File Operations

#### Upload Files
1. Open data folder (Ctrl+Shift+D)
2. Click "Upload" button
3. Select files (multiple allowed)
4. Files uploaded to current folder

#### Download Files
1. Open data folder
2. Click download icon next to file
3. File downloads to browser's download folder

#### Delete Files
1. Open data folder
2. Click delete icon next to file
3. Confirm deletion
4. File removed

#### Navigate Folders
- Click folder in sidebar to open
- Click "Root" button to go back to main folder

### Folder Structure
```
user-data/
├── profiles/       User profiles & settings
├── templates/      Chart & order templates
├── logs/           Application logs
├── cache/          Temporary cached data
├── indicators/     Custom indicators
├── drawings/       Chart drawings
└── workspaces/     Saved workspace configs
```

## Exit Functionality

### Exit with Unsaved Changes

**Trigger**: Close browser tab, click logout, or close app

**Dialog appears if you have**:
- Modified workspace layout
- Open positions
- Pending orders

### Exit Options

#### Save & Exit
- Saves current workspace
- Preserves layout, charts, settings
- Exits application
- **Recommended** if you want to restore your workspace

#### Discard & Exit
- Exits without saving
- Changes lost
- Faster exit
- Use if you don't need current layout

#### Cancel
- Returns to application
- No changes made
- Continue working

### Preferences

#### "Save workspace before exiting"
- ✅ Checked: Workspace saved automatically
- ❌ Unchecked: Workspace not saved

#### "Don't ask again"
- ✅ Checked: No confirmation next time
- ❌ Unchecked: Always ask

**Reset preferences**:
Open browser console (F12) and run:
```javascript
localStorage.removeItem('exitPreferences');
```

### Unsaved Items Indicators

| Color | Type | Meaning |
|-------|------|---------|
| 🔵 Blue | Workspace | Layout modified |
| 🟢 Green | Trades | Open positions |
| 🟡 Yellow | Orders | Pending orders |

### Important Notes

⚠️ **Open Positions**: Your positions remain active even after exit. Monitor them remotely or close before exiting.

✅ **Auto-Save**: If "Don't ask again" + "Save workspace" are checked, workspace saves automatically on every exit.

🔄 **Restore**: Next login restores your saved workspace automatically.

## Common Tasks

### Task: Export Chart Template
1. Save chart template in app
2. Press Ctrl+Shift+D
3. Navigate to "templates" folder
4. Download template file
5. Share with others or backup

### Task: Import Custom Indicator
1. Get indicator file
2. Press Ctrl+Shift+D
3. Navigate to "indicators" folder
4. Click "Upload"
5. Select indicator file
6. Refresh app to load indicator

### Task: Backup Workspace
1. Press Ctrl+Shift+D
2. Navigate to "workspaces" folder
3. Download workspace files
4. Store safely

### Task: Quick Exit (No Save)
1. Close browser tab
2. Click "Discard & Exit"
3. Done

### Task: Always Save on Exit
1. Close app once
2. In exit dialog:
   - ✅ Check "Save workspace before exiting"
   - ✅ Check "Don't ask again"
3. Click "Save & Exit"
4. Future exits auto-save

## API Endpoints (For Developers)

### Get Data Folder Path
```http
GET /api/user/data-folder
```

### List Files
```http
GET /api/user/data-folder/list?path=/subfolder
```

### Download File
```http
GET /api/user/data-folder/download?path=/path/to/file
```

### Upload File
```http
POST /api/user/data-folder/upload
Content-Type: multipart/form-data

file: [file data]
path: /target/folder
```

### Delete File
```http
DELETE /api/user/data-folder/delete
Content-Type: application/json

{
  "path": "/path/to/file"
}
```

### Create Directory
```http
POST /api/user/data-folder/mkdir
Content-Type: application/json

{
  "path": "/path/to/new-dir"
}
```

## Programmatic Usage

### Open Data Folder
```typescript
import { dataFolderManager } from './services/dataFolderManager';

// Check platform
const platform = dataFolderManager.getPlatform();

// Open (Electron)
await dataFolderManager.openInExplorer();
```

### Exit Handling
```typescript
import { exitHandler } from './services/exitHandler';

// Check for unsaved changes
const changes = exitHandler.detectUnsavedChanges();

// Initiate exit
const shouldExit = await exitHandler.initiateExit();
if (shouldExit) {
  window.location.href = '/';
}
```

### Register Cleanup Task
```typescript
exitHandler.registerCleanupTask(async () => {
  // Your cleanup code
  console.log('Cleaning up...');
});
```

## Troubleshooting

### Q: Data folder doesn't open
**A**: Check backend is running on port 7999

### Q: Can't upload files
**A**: Check file size (<100MB) and disk space

### Q: Exit dialog always shows
**A**: Uncheck "Don't ask again" or clear preferences

### Q: Workspace not saving
**A**: Check localStorage quota and "Save workspace" checkbox

### Q: Lost my workspace
**A**: Check "workspaces" folder for backup files

## Tips & Tricks

💡 **Tip 1**: Use templates folder to share chart setups with other traders

💡 **Tip 2**: Enable auto-save to never lose your workspace again

💡 **Tip 3**: Back up your data folder regularly (especially workspaces and templates)

💡 **Tip 4**: Clear cache folder if app becomes slow

💡 **Tip 5**: Check logs folder for debugging errors

## Support

- 📚 Full Documentation: `docs/DATA_FOLDER_EXIT_IMPLEMENTATION.md`
- 📋 Implementation Guide: `docs/IMPLEMENTATION_SUMMARY_DATA_EXIT.md`
- 🐛 Report Issues: GitHub Issues
- 💬 Community: Discord/Slack channel
