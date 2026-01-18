# Alerting System Implementation

## Overview

The alerting system provides real-time WebSocket-based alerts with sound notifications, browser notifications, mobile vibrations, and dynamic rule management.

## Components

### 1. AlertCard (`/src/components/AlertCard.tsx`)

Individual alert card with severity indicators and action buttons.

**Features:**
- Severity levels: LOW, MEDIUM, HIGH, CRITICAL
- Color-coded indicators
- Timestamp display
- Three actions: Acknowledge, Snooze, Dismiss
- Slide-in animation

**Props:**
```typescript
type AlertCardProps = {
  alert: Alert;
  onAcknowledge: (id: string) => void;
  onSnooze: (id: string, minutes: number) => void;
  onDismiss: (id: string) => void;
};

type Alert = {
  id: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  message: string;
  timestamp: number;
  acknowledged?: boolean;
  snoozedUntil?: number;
};
```

### 2. AlertsContainer (`/src/components/AlertsContainer.tsx`)

Main container managing the alert stack with WebSocket integration.

**Features:**
- WebSocket subscription to "alerts" channel
- Sound playback on new alert (Audio API)
- Browser notifications (Notification API)
- Mobile vibration (Vibration API)
- Auto-hide non-critical alerts after 10 seconds
- Badge counter showing unacknowledged count
- Alert snoozing (5min, 15min, 1h)

**Props:**
```typescript
type AlertsContainerProps = {
  wsConnection?: WebSocket | null;
};
```

**Usage:**
```tsx
import { AlertsContainer } from './components/AlertsContainer';

function App() {
  const wsRef = useRef<WebSocket | null>(null);

  return (
    <>
      <AlertsContainer wsConnection={wsRef.current} />
      {/* Rest of app */}
    </>
  );
}
```

### 3. AlertRulesManager (`/src/components/AlertRulesManager.tsx`)

UI for creating, editing, and managing alert rules.

**Features:**
- List all configured rules
- Create new rules with dialog
- Edit existing rules
- Delete rules
- Enable/disable toggle
- Test rule button (triggers test alert)
- Dynamic severity color coding

**Rule Types:**
- Price Above/Below (for symbols)
- P&L Above/Below (for account)
- Margin Level Below (for account)

**Usage:**
```tsx
import { AlertRulesManager } from './components/AlertRulesManager';

function SettingsPanel() {
  return (
    <div className="p-4">
      <AlertRulesManager />
    </div>
  );
}
```

## WebSocket Integration

### Message Format

Alerts are sent from the backend via WebSocket with this format:

```json
{
  "type": "alert",
  "id": "alert-123",
  "severity": "HIGH",
  "message": "EURUSD price above 1.1000",
  "timestamp": 1642534800000
}
```

### Backend Requirements

The backend should:

1. Evaluate alert rules on each tick/update
2. Send WebSocket messages on "alerts" channel
3. Provide REST endpoints:
   - `GET /api/alerts/rules` - Get all rules
   - `POST /api/alerts/rules` - Create rule
   - `PUT /api/alerts/rules/:id` - Update rule
   - `DELETE /api/alerts/rules/:id` - Delete rule
   - `POST /api/alerts/rules/:id/test` - Test rule

## API Integration

Alert rules API is available via `/src/services/api.ts`:

```typescript
import { api } from './services/api';

// Get all rules
const rules = await api.alerts.getRules();

// Create rule
await api.alerts.createRule({
  name: 'EURUSD High Price',
  enabled: true,
  condition: {
    type: 'price_above',
    symbol: 'EURUSD',
    threshold: 1.1000,
  },
  severity: 'HIGH',
  message: 'EURUSD price exceeded 1.1000',
});

// Update rule
await api.alerts.updateRule('rule-id', {
  enabled: false,
});

// Delete rule
await api.alerts.deleteRule('rule-id');

// Test rule
await api.alerts.testRule('rule-id');
```

## Notification Permissions

The system automatically requests browser notification permission on mount. Users will see a browser prompt asking to allow notifications.

**Permission Handling:**
- Granted: Shows browser notifications
- Denied: Only shows in-app alerts
- Default: Requests permission on first mount

## Sound Notifications

Default sound is a simple beep (embedded as data URL). To use a custom sound:

```typescript
// In AlertsContainer.tsx
const ALERT_SOUND_URL = '/sounds/alert.mp3'; // Change to your audio file
```

Supported formats: MP3, WAV, OGG

## Mobile Support

On mobile devices, the system uses the Vibration API:

- **LOW/MEDIUM**: Single vibration (100ms)
- **HIGH/CRITICAL**: Pattern vibration (200ms, pause, 200ms)

## Severity Color Coding

| Severity | Color | Use Case |
|----------|-------|----------|
| LOW | Blue | Informational alerts |
| MEDIUM | Yellow | Important notices |
| HIGH | Orange | Urgent attention required |
| CRITICAL | Red | Critical issues, stays until acknowledged |

## Auto-Hide Behavior

- **LOW/MEDIUM/HIGH**: Auto-hide after 10 seconds
- **CRITICAL**: Persists until user acknowledges

## Testing

### Manual Test
1. Start the app
2. Create an alert rule via AlertRulesManager
3. Click "Test" button on the rule
4. Verify alert appears with sound/notification

### Programmatic Test
```typescript
// Send test WebSocket message
wsConnection.send(JSON.stringify({
  type: 'alert',
  id: 'test-' + Date.now(),
  severity: 'HIGH',
  message: 'This is a test alert',
  timestamp: Date.now(),
}));
```

## Customization

### Change Alert Position
```tsx
// In AlertsContainer.tsx
<div className="fixed top-16 right-4 z-40"> {/* Change position here */}
```

### Change Auto-Hide Duration
```typescript
// In AlertsContainer.tsx
if (alert.severity !== 'CRITICAL') {
  const timeout = setTimeout(() => {
    dismissAlert(alert.id);
  }, 10000); // Change duration here (milliseconds)
}
```

### Add Custom Severity Levels
```typescript
// In AlertCard.tsx
const SEVERITY_CONFIG = {
  // Add your custom severity
  EXTREME: {
    icon: AlertOctagon,
    bgColor: 'bg-purple-500/10',
    borderColor: 'border-purple-500/30',
    textColor: 'text-purple-400',
    dotColor: 'bg-purple-500',
  },
};
```

## File Structure

```
src/
├── components/
│   ├── AlertCard.tsx           # Individual alert card
│   ├── AlertsContainer.tsx     # Alert stack + WebSocket
│   └── AlertRulesManager.tsx   # Rule management UI
└── services/
    └── api.ts                  # Added alertsApi
```

## Dependencies

All dependencies are already included:

- `lucide-react` - Icons (Bell, Clock, CheckCircle, etc.)
- `react` - Component framework
- Native browser APIs:
  - WebSocket API
  - Audio API
  - Notification API
  - Vibration API

## Browser Compatibility

- **WebSocket**: All modern browsers
- **Audio**: All modern browsers
- **Notifications**: Chrome 50+, Firefox 44+, Safari 16+
- **Vibration**: Chrome (Android), Firefox (Android) - Desktop browsers ignore

## Performance Considerations

- Alerts are rendered with virtualization (only visible alerts in DOM)
- Auto-hide cleanup prevents memory leaks
- WebSocket messages are throttled to prevent UI lag
- Sound playback is limited to one at a time

## Security Notes

- Alert messages should be sanitized server-side
- Ensure WebSocket connection is authenticated
- Validate alert rule thresholds to prevent abuse
- Rate-limit test rule endpoint to prevent spam

## Future Enhancements

Potential improvements:

1. **Alert History** - Store dismissed alerts in database
2. **Alert Groups** - Group related alerts together
3. **Custom Sounds** - Per-severity or per-rule sound selection
4. **Email/SMS** - Send critical alerts via email/SMS
5. **Alert Templates** - Pre-configured rule templates
6. **Multi-account Support** - Different rules per account
7. **Alert Analytics** - Dashboard showing alert frequency/patterns
8. **Scheduled Alerts** - Time-based alert muting
9. **Alert Priorities** - More granular priority system
10. **Rich Alerts** - Include charts/images in alert messages

## Troubleshooting

### Alerts Not Appearing
- Check WebSocket connection is active
- Verify backend is sending correct message format
- Check browser console for errors

### Sound Not Playing
- Ensure browser audio is not muted
- Check ALERT_SOUND_URL is valid
- Some browsers block autoplay - requires user interaction first

### Notifications Not Showing
- Check notification permission granted
- Verify browser supports Notification API
- Check browser notification settings

### WebSocket Disconnects
- Verify auth token is valid
- Check backend WebSocket server is running
- Look for network issues (firewall, proxy)

## Support

For issues or questions:
1. Check browser console for errors
2. Verify WebSocket messages in Network tab
3. Test with curl/Postman to verify backend endpoints
4. Enable verbose logging in AlertsContainer
