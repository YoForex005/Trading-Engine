# File Menu Quick Reference

## Quick Implementation Guide

### 1. Import the FileMenu Component

```typescript
import { FileMenu } from './components/layout/FileMenu';
```

### 2. Add to Your MenuBar

```typescript
{activeMenu === 'File' && (
    <FileMenu
        onSave={() => handleSave()}
        onSaveAsPicture={() => exportChart()}
        onOpenDataFolder={() => openDataFolder()}
        onPrint={() => printChart()}
        onExit={() => handleExit()}
        hasUnsavedChanges={hasUnsavedChanges}
    />
)}
```

## Component Summary

| Component | Location | Purpose |
|-----------|----------|---------|
| FileMenu | `layout/FileMenu.tsx` | Main dropdown menu |
| SaveWorkspaceDialog | `dialogs/SaveWorkspaceDialog.tsx` | Save/overwrite workspace |
| PrintSetupDialog | `dialogs/PrintSetupDialog.tsx` | Configure print settings |
| PrintPreviewDialog | `dialogs/PrintPreviewDialog.tsx` | Preview before printing |
| ExitConfirmDialog | `dialogs/ExitConfirmDialog.tsx` | Confirm exit with unsaved changes |

## Menu Items Summary

| Item | Shortcut | Status | Dialog Opened |
|------|----------|--------|---------------|
| Save | Ctrl+S | Active | SaveWorkspaceDialog (if unsaved) |
| Save As Picture | - | Active | None (direct export) |
| Open Data Folder | Ctrl+Shift+D | Active | None (opens folder) |
| Print | Ctrl+P | Active | None (prints directly) |
| Print Preview | - | Active | PrintPreviewDialog |
| Print Setup | - | Active | PrintSetupDialog |
| Exit | - | Active | ExitConfirmDialog (if unsaved) |

## Keyboard Shortcuts

Add to your App component:

```typescript
useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
        // Save
        if (e.ctrlKey && e.key === 's') {
            e.preventDefault();
            handleSave();
        }

        // Open Data Folder
        if (e.ctrlKey && e.shiftKey && e.key === 'D') {
            e.preventDefault();
            openDataFolder();
        }

        // Print
        if (e.ctrlKey && e.key === 'p') {
            e.preventDefault();
            handlePrint();
        }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
}, []);
```

## Props Reference

### FileMenu Props
```typescript
interface FileMenuProps {
    onSave?: () => void;               // Save workspace
    onSaveAsPicture?: () => void;      // Export chart as image
    onOpenDataFolder?: () => void;     // Open data directory
    onPrint?: () => void;              // Print chart
    onPrintPreview?: () => void;       // Preview print
    onPrintSetup?: () => void;         // Configure print
    onExit?: () => void;               // Exit application
    hasUnsavedChanges?: boolean;       // Show save dialogs
}
```

### SaveWorkspaceDialog Props
```typescript
interface SaveWorkspaceDialogProps {
    onConfirm: (workspaceName?: string) => void;
    onCancel: () => void;
    existingWorkspace?: string;
}
```

### PrintSetupDialog Props
```typescript
interface PrintSetupDialogProps {
    onConfirm: (settings: PrintSettings) => void;
    onCancel: () => void;
    defaultSettings?: Partial<PrintSettings>;
}
```

### PrintPreviewDialog Props
```typescript
interface PrintPreviewDialogProps {
    onPrint: () => void;
    onClose: () => void;
}
```

### ExitConfirmDialog Props
```typescript
interface ExitConfirmDialogProps {
    onSave: () => void;
    onDontSave: () => void;
    onCancel: () => void;
}
```

## Styling Classes

All components use these Tailwind classes for consistency:

### Colors
- Background: `bg-[#1e1e1e]`
- Surface: `bg-[#252528]`
- Border: `border-zinc-700`
- Text: `text-zinc-300`
- Hover: `hover:bg-[#2a2e39]`
- Active: `bg-blue-600`

### Typography
- Menu items: `text-[12px]`
- Dialog title: `text-sm font-semibold`
- Input fields: `text-xs`

## Testing

Run tests:
```bash
npm test -- FileMenu.test.tsx
npm test -- dialogs.test.tsx
```

## Common Issues

### Dialog Not Showing
- Check z-index (should be `z-[200]`)
- Ensure dialog state is managed correctly

### Keyboard Shortcuts Not Working
- Add global event listeners in App component
- Prevent default browser behavior with `e.preventDefault()`

### Styling Conflicts
- Check Tailwind CSS is properly configured
- Verify all required classes are included in build

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## Performance Tips

1. Lazy load dialogs only when needed
2. Use React.memo for frequently re-rendered components
3. Debounce keyboard shortcuts if needed

## Accessibility

All components include:
- ARIA labels
- Keyboard navigation (Tab, Enter, Escape)
- Focus indicators
- Screen reader support

## Next Steps

1. Implement actual save/load workspace functionality
2. Add chart export in multiple formats
3. Integrate with backend for persistent storage
4. Add cloud sync for workspaces
