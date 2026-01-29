# Alert System Integration Example

## Quick Start Integration

### Step 1: Add AlertsContainer to App.tsx

The AlertsContainer is already integrated in `/src/App.tsx`:

```tsx
import { AlertsContainer } from './components/AlertsContainer';

function App() {
  const wsRef = useRef<WebSocket | null>(null);

  return (
    <div>
      {/* Real-time alerts - already added */}
      <AlertsContainer wsConnection={wsRef.current} />

      {/* Rest of your app */}
    </div>
  );
}
```

### Step 2: Add AlertRulesManager to Settings/Admin Panel

Option A: Add to AdminPanel.tsx

```tsx
import { AlertRulesManager } from './AlertRulesManager';
import { Bell } from 'lucide-react';

// Inside AdminPanel component, add a new tab:
const [activeTab, setActiveTab] = useState<'accounts' | 'symbols' | 'config' | 'alerts'>('accounts');

// Add tab button
<button
  onClick={() => setActiveTab('alerts')}
  className={/* styling */}
>
  <Bell className="w-4 h-4" />
  Alert Rules
</button>

// Add tab content
{activeTab === 'alerts' && (
  <div className="flex-1 overflow-hidden">
    <AlertRulesManager />
  </div>
)}
```

Option B: Create dedicated Alerts Panel in BottomDock

```tsx
// In BottomDock.tsx
import { AlertRulesManager } from './AlertRulesManager';

// Add new tab
<button onClick={() => setActiveTab('alerts')}>
  <Bell className="w-4 h-4" />
  <span>Alerts</span>
</button>

// Add panel
{activeTab === 'alerts' && (
  <div className="flex-1 overflow-hidden">
    <AlertRulesManager />
  </div>
)}
```

### Step 3: Backend WebSocket Alert Emission

The backend should emit alerts when conditions are met:

```go
// Example Go backend alert emission
type Alert struct {
    Type      string `json:"type"`
    ID        string `json:"id"`
    Severity  string `json:"severity"`
    Message   string `json:"message"`
    Timestamp int64  `json:"timestamp"`
}

func (s *Server) checkAlertRules(tick Tick) {
    for _, rule := range s.alertRules {
        if !rule.Enabled {
            continue
        }

        triggered := false

        switch rule.Condition.Type {
        case "price_above":
            if rule.Condition.Symbol == tick.Symbol && tick.Bid > rule.Condition.Threshold {
                triggered = true
            }
        case "price_below":
            if rule.Condition.Symbol == tick.Symbol && tick.Bid < rule.Condition.Threshold {
                triggered = true
            }
        // ... other conditions
        }

        if triggered {
            alert := Alert{
                Type:      "alert",
                ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
                Severity:  rule.Severity,
                Message:   rule.Message,
                Timestamp: time.Now().UnixMilli(),
            }

            // Broadcast to all WebSocket clients
            s.broadcastJSON(alert)
        }
    }
}
```

### Step 4: Backend REST Endpoints

Required endpoints for AlertRulesManager:

```go
// GET /api/alerts/rules
func (s *Server) handleGetAlertRules(c *gin.Context) {
    rules := s.getAlertRules()
    c.JSON(200, rules)
}

// POST /api/alerts/rules
func (s *Server) handleCreateAlertRule(c *gin.Context) {
    var rule AlertRule
    if err := c.ShouldBindJSON(&rule); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    rule.ID = generateID()
    s.saveAlertRule(rule)
    c.JSON(201, rule)
}

// PUT /api/alerts/rules/:id
func (s *Server) handleUpdateAlertRule(c *gin.Context) {
    id := c.Param("id")
    var updates AlertRule
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    s.updateAlertRule(id, updates)
    c.JSON(200, updates)
}

// DELETE /api/alerts/rules/:id
func (s *Server) handleDeleteAlertRule(c *gin.Context) {
    id := c.Param("id")
    s.deleteAlertRule(id)
    c.JSON(200, gin.H{"success": true})
}

// POST /api/alerts/rules/:id/test
func (s *Server) handleTestAlertRule(c *gin.Context) {
    id := c.Param("id")
    rule := s.getAlertRule(id)

    // Create test alert
    alert := Alert{
        Type:      "alert",
        ID:        fmt.Sprintf("test-%d", time.Now().UnixNano()),
        Severity:  rule.Severity,
        Message:   "[TEST] " + rule.Message,
        Timestamp: time.Now().UnixMilli(),
    }

    // Broadcast test alert
    s.broadcastJSON(alert)
    c.JSON(200, gin.H{"success": true})
}
```

## Example Alert Rules

### Price Alert
```json
{
  "name": "EURUSD High Price",
  "enabled": true,
  "condition": {
    "type": "price_above",
    "symbol": "EURUSD",
    "threshold": 1.1000
  },
  "severity": "HIGH",
  "message": "EURUSD price exceeded 1.1000"
}
```

### P&L Alert
```json
{
  "name": "Daily Loss Limit",
  "enabled": true,
  "condition": {
    "type": "pnl_below",
    "threshold": -500.00
  },
  "severity": "CRITICAL",
  "message": "Daily loss limit reached: -$500"
}
```

### Margin Alert
```json
{
  "name": "Margin Call Warning",
  "enabled": true,
  "condition": {
    "type": "margin_level_below",
    "threshold": 120.0
  },
  "severity": "CRITICAL",
  "message": "Margin level below 120% - Risk of margin call"
}
```

## Testing the Integration

### 1. Test WebSocket Alert Manually

Open browser console and send a test alert:

```javascript
// In browser console (when WebSocket is connected)
const ws = /* get websocket from App.tsx */;
ws.send(JSON.stringify({
  type: 'alert',
  id: 'test-123',
  severity: 'HIGH',
  message: 'This is a test alert from browser console',
  timestamp: Date.now()
}));
```

### 2. Test via Backend

Create a test endpoint:

```go
// POST /api/alerts/test
func (s *Server) handleTestAlert(c *gin.Context) {
    alert := Alert{
        Type:      "alert",
        ID:        fmt.Sprintf("test-%d", time.Now().UnixNano()),
        Severity:  "HIGH",
        Message:   "Test alert from backend",
        Timestamp: time.Now().UnixMilli(),
    }
    s.broadcastJSON(alert)
    c.JSON(200, gin.H{"success": true})
}
```

Then trigger with curl:

```bash
curl -X POST http://localhost:8080/api/alerts/test
```

### 3. Test Alert Rules Manager

1. Open the app
2. Navigate to the alerts panel (Admin Panel → Alert Rules tab)
3. Click "New Rule"
4. Fill in the form:
   - Name: "Test Rule"
   - Condition: "Price Above"
   - Symbol: "EURUSD"
   - Threshold: 1.0000 (low threshold for testing)
   - Severity: "HIGH"
   - Message: "Test alert rule triggered"
5. Click "Create"
6. Click "Test" button on the newly created rule
7. Verify alert appears in top-right with sound

## Styling Customization

### Change Alert Position

```tsx
// In AlertsContainer.tsx
// Default: top-16 right-4 (top-right)
// Options:
- top-16 left-4    // top-left
- bottom-16 right-4 // bottom-right
- bottom-16 left-4  // bottom-left
```

### Change Severity Colors

```tsx
// In AlertCard.tsx
const SEVERITY_CONFIG = {
  HIGH: {
    icon: AlertTriangle,
    bgColor: 'bg-purple-500/10',     // Change colors here
    borderColor: 'border-purple-500/30',
    textColor: 'text-purple-400',
    dotColor: 'bg-purple-500',
  },
};
```

### Change Auto-Hide Duration

```typescript
// In AlertsContainer.tsx
const AUTO_HIDE_DURATION = 5000; // 5 seconds instead of 10

if (alert.severity !== 'CRITICAL') {
  const timeout = setTimeout(() => {
    dismissAlert(alert.id);
  }, AUTO_HIDE_DURATION);
}
```

## Mobile Responsive Design

The alerts are already mobile-responsive:

- Desktop: Fixed width (320-400px)
- Mobile: Full width with horizontal padding
- Touch-friendly buttons (larger tap targets)
- Stacks vertically on narrow screens

## Performance Tips

1. **Limit Active Alerts**: Show max 5 alerts at once
2. **Debounce Rule Checks**: Don't check on every tick, use interval
3. **Store History**: Move dismissed alerts to database instead of memory
4. **Rate Limiting**: Prevent alert spam (max 1 per rule per minute)

```go
// Example rate limiting
type AlertRateLimiter struct {
    lastTriggered map[string]time.Time
    mutex         sync.RWMutex
}

func (l *AlertRateLimiter) CanTrigger(ruleID string, cooldown time.Duration) bool {
    l.mutex.Lock()
    defer l.mutex.Unlock()

    lastTime, exists := l.lastTriggered[ruleID]
    if !exists || time.Since(lastTime) > cooldown {
        l.lastTriggered[ruleID] = time.Now()
        return true
    }
    return false
}
```

## Complete Integration Checklist

- [x] AlertCard.tsx created
- [x] AlertsContainer.tsx created
- [x] AlertRulesManager.tsx created
- [x] API endpoints added to api.ts
- [x] AlertsContainer integrated in App.tsx
- [ ] Add AlertRulesManager to AdminPanel or BottomDock
- [ ] Implement backend WebSocket alert emission
- [ ] Implement backend REST endpoints for rules
- [ ] Test WebSocket alert delivery
- [ ] Test rule CRUD operations
- [ ] Test notification permissions
- [ ] Test sound playback
- [ ] Test mobile vibrations (if applicable)

## Next Steps

1. Add AlertRulesManager to your UI (AdminPanel or BottomDock)
2. Implement backend alert rule storage (database or in-memory)
3. Implement backend alert evaluation logic
4. Add rate limiting to prevent alert spam
5. Consider adding alert history/analytics
6. Set up error tracking for WebSocket disconnections
