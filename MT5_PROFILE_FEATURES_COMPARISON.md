# MT5 Profile Features vs Trading Engine Implementation

**Research Date**: 2026-02-02
**Platform**: Trading Engine (Go Backend + React Frontend)
**Comparison Target**: MetaTrader 5 (MT5) Profile System

---

## Table of Contents

1. [MT5 Profile System Overview](#mt5-profile-system-overview)
2. [Current Implementation Analysis](#current-implementation-analysis)
3. [Feature Comparison Matrix](#feature-comparison-matrix)
4. [Gap Analysis](#gap-analysis)
5. [Recommendations](#recommendations)
6. [Implementation Roadmap](#implementation-roadmap)

---

## MT5 Profile System Overview

### What is an MT5 Profile?

MT5 **Profiles** are comprehensive workspace configurations that save and restore the complete trading environment state. A profile stores:

#### 1. **Multiple Named Profiles**
- Create unlimited named profiles (e.g., "Scalping", "Swing Trading", "Analysis Only")
- Quick-switch between different trading setups
- Each profile is independent with separate configurations

#### 2. **Chart Templates & Layouts**
- Save chart configurations (symbols, timeframes, indicators)
- Store chart window positions and sizes
- Save zoom levels, chart colors, and drawing tools
- Store indicator parameters for later use

#### 3. **Workspace Configuration**
- Panel layouts (left, right, bottom docks)
- Window positions on multi-monitor setups
- Dockable window arrangements
- Size and visibility of each panel

#### 4. **Indicator Settings**
- Selected indicators and their parameters
- Indicator color schemes
- Applied period and calculation methods
- Indicator alert thresholds

#### 5. **Expert Advisors (EAs) per Profile**
- Different EAs active in different profiles
- EA-specific settings per profile
- Trading mode per profile
- Strategy parameters

#### 6. **Custom Toolbars**
- Custom toolbar layouts
- Saved toolbar button configurations
- Toolbar button groupings

#### 7. **Profile Import/Export**
- Export profiles as files (.tpl format)
- Share configurations with other traders
- Import pre-configured profiles
- Backup profiles

#### 8. **Quick Profile Switching**
- Keyboard shortcuts for quick switch
- Profile dropdown selector
- One-click profile restoration
- Automatic layout restoration

### MT5 Profile Architecture

**File Location**: `Profiles` folder in MT5 data directory
```
C:\Users\[Username]\AppData\Roaming\MetaQuotes\Terminal\[ServerName]\Profiles
```

**Typical Profile Structure**:
```
My Profile.tpl
├── Charts (multiple)
│   ├── chart1.ini
│   ├── chart2.ini
│   └── chart3.ini
├── Workspace layout
├── Toolbar configuration
├── Indicator settings
└── Expert Advisor configuration
```

### Key MT5 Profile Features

| Feature | Description |
|---------|-------------|
| **Profile Switching** | Instant switch between complete setups |
| **Template Saving** | Save individual chart configurations |
| **Multi-Chart Setup** | Multiple charts open with different symbols/timeframes |
| **Indicator Persistence** | All indicator settings saved |
| **EA Management** | Different EAs per profile |
| **Layout Recovery** | Restore exact window positions/sizes |
| **Import/Export** | Share profiles (.tpl files) |
| **Auto-Load** | Load profile on platform startup |
| **Custom Toolbars** | Personalized toolbar buttons |

---

## Current Implementation Analysis

### What We Have

#### 1. **Basic Settings Storage** (Partial)
**Location**: `clients/desktop/src/store/useAppStore.ts`

```typescript
persist(
  (set) => ({...}),
  {
    name: 'trading-app-storage',
    partialize: (state) => ({
      selectedSymbol: state.selectedSymbol,      // ✅ Persisted
      chartType: state.chartType,                 // ✅ Persisted
      timeframe: state.timeframe,                 // ✅ Persisted
      orderVolume: state.orderVolume,             // ✅ Persisted
    }),
  }
)
```

**Status**: Minimal localStorage support for basic UI state
- Only 4 settings persisted
- No profile concept
- No named configurations
- No chart template support

#### 2. **Options Dialog** (Placeholder)
**Location**: `clients/desktop/src/components/settings/OptionsDialog.tsx`

```typescript
const tabs: { id: TabId; label: string }[] = [
  { id: 'server', label: 'Server' },
  { id: 'charts', label: 'Charts' },      // Chart settings tab exists
  { id: 'trade', label: 'Trade' },
  { id: 'automation', label: 'Automation & Bots' },
  // ... more tabs
];
```

**Status**: UI exists but mostly non-functional
- ChartsTab has options for chart behavior
- Trade Tab has symbol/volume/deviation defaults
- Automation Tab mentions "profile has been changed" (!) but not implemented

#### 3. **Client Profile** (Wrong Type - for Risk Classification)
**Location**: `backend/cbook/client_profile.go`

```go
type ClientProfile struct {
  AccountID   int64
  UserID      string
  TotalTrades int64
  WinRate     float64
  // ... risk metrics for routing decisions
}
```

**Status**: This is for B-Book routing, NOT user profiles
- Tracks client trading behavior (win rates, toxicity scores)
- Used for routing decisions (A-Book vs B-Book)
- NOT related to workspace configuration

#### 4. **Chart Component State** (Basic)
**Location**: `clients/desktop/src/components/TradingChart.tsx`

```typescript
const [chartType, setChartType] = useState('candlestick');
const [timeframe, setTimeframe] = useState('1m');
// ... individual component state
```

**Status**: Inline state, not persistable
- No template saving
- No multi-chart support
- No indicator parameter persistence

### What We're Missing

1. **Profile System**: No named profiles concept
2. **Chart Templates**: Cannot save/load chart configurations
3. **Workspace Layouts**: No docking/layout system
4. **Multi-Chart Management**: Limited to single chart view
5. **Indicator Persistence**: Indicator settings not saved
6. **Import/Export**: No profile file export/import
7. **Toolbar Configuration**: No custom toolbar support
8. **Quick Switching**: No profile selector dropdown
9. **Profile Database**: No backend storage for profiles
10. **Multi-Monitor Support**: No layout for multi-screen setups

---

## Feature Comparison Matrix

### Summary Table

| Category | Feature | MT5 | Trading Engine | Status |
|----------|---------|-----|---|---|
| **Profiles** | Multiple Named Profiles | ✅ Yes | ❌ No | Missing |
| | Profile Switching | ✅ Yes | ❌ No | Missing |
| | Auto-Load Profile | ✅ Yes | ❌ No | Missing |
| **Chart Templates** | Save Chart Config | ✅ Yes | ❌ No | Missing |
| | Load Chart Template | ✅ Yes | ❌ No | Missing |
| | Template Parameters | ✅ Yes | ❌ No | Missing |
| **Workspace** | Save Window Layout | ✅ Yes | ❌ No | Missing |
| | Restore Window Positions | ✅ Yes | ❌ No | Missing |
| | Multi-Monitor Support | ✅ Yes | ❌ No | Missing |
| | Dockable Panels | ✅ Yes | ⚠️ Fixed | Limited |
| **Indicators** | Persist Indicator List | ✅ Yes | ❌ No | Missing |
| | Save Indicator Settings | ✅ Yes | ❌ No | Missing |
| | Indicator Templates | ✅ Yes | ❌ No | Missing |
| **Expert Advisors** | EA per Profile | ✅ Yes | N/A | N/A (No EA support) |
| | EA Parameter Profiles | ✅ Yes | N/A | N/A |
| **Custom UI** | Custom Toolbars | ✅ Yes | ⚠️ Basic | Limited |
| | Toolbar Customization | ✅ Yes | ❌ No | Missing |
| **Import/Export** | Profile Export (.tpl) | ✅ Yes | ❌ No | Missing |
| | Profile Import | ✅ Yes | ❌ No | Missing |
| | Share Profiles | ✅ Yes | ❌ No | Missing |
| **Settings** | General Settings | ✅ Yes | ⚠️ Partial | 4 settings only |
| | Trade Defaults | ✅ Yes | ⚠️ Partial | Limited |
| | Notification Settings | ✅ Yes | ❌ No | Missing |

### Detailed Feature Breakdown

#### 1. Multiple Named Profiles

**MT5 Implementation**:
```
Profiles/
├── Default.tpl
├── Scalping.tpl
├── Swing Trading.tpl
├── Analysis Only.tpl
└── Cryptocurrency.tpl
```

**Trading Engine**: ❌ Not implemented
- No profile selection UI
- No named configuration storage
- No profile list management

**Estimated Effort**: Medium (2-3 weeks)

---

#### 2. Profile Switching

**MT5 Implementation**:
- Dropdown selector in toolbar
- Keyboard shortcut (Ctrl+Alt+P + number)
- Automatic restoration of entire workspace
- Instant layout recovery

**Trading Engine**: ❌ Not implemented
- No UI for profile selection
- No instant switching mechanism
- No workspace restoration

**Estimated Effort**: Low-Medium (1-2 weeks, after profiles exist)

---

#### 3. Chart Templates

**MT5 Implementation**:
```
Right-click on chart → Template → Save As...
├── Saves chart symbol
├── Saves timeframe
├── Saves all indicators with settings
├── Saves colors and styles
├── Saves drawings
└── Auto-loads when selecting template
```

**Trading Engine**: ❌ Not implemented
- OptionsDialog has ChartsTab but non-functional
- No template save/load mechanism
- No indicator persistence

**Estimated Effort**: Medium (3-4 weeks)

**Example Data Structure**:
```json
{
  "chartTemplate": {
    "id": "uuid",
    "name": "EURUSD Scalping",
    "symbol": "EURUSD",
    "timeframe": "5m",
    "chartType": "candlestick",
    "indicators": [
      {
        "type": "RSI",
        "period": 14,
        "level1": 30,
        "level2": 70,
        "color": "#00FF00"
      },
      {
        "type": "MA",
        "period": 20,
        "method": "EMA",
        "color": "#FF0000"
      }
    ],
    "drawings": [...],
    "colors": {...},
    "createdAt": "2026-02-02T10:00:00Z"
  }
}
```

---

#### 4. Workspace Layout Persistence

**MT5 Implementation**:
- Saves all open windows' positions and sizes
- Remembers which panels are visible
- Restores multi-monitor layout
- Preserves zoom levels

**Trading Engine**: ⚠️ Partial
- Fixed docking layout (not flexible)
- No position/size persistence
- No multi-monitor support

**Estimated Effort**: High (4-6 weeks)

---

#### 5. Indicator Settings Persistence

**MT5 Implementation**:
```
Chart Window → Indicator Properties
├── Period
├── Method (SMA, EMA, WMA, etc.)
├── Applied price
├── Colors and styles
├── Alert thresholds
└── All auto-saved with chart template
```

**Trading Engine**: ❌ Not implemented
- No indicator system yet
- No parameter storage
- No indicator list persistence

**Estimated Effort**: High (6+ weeks for full indicator suite)

---

#### 6. Expert Advisor Profile Support

**MT5 Implementation**:
- Different EAs can run in different profiles
- EA parameters saved per profile
- Automatic EA loading with profile
- Multiple EA instances possible

**Trading Engine**: N/A
- No EA support at all
- Would require full EA runtime implementation

**Estimated Effort**: Very High (16-20 weeks for EA system)

---

#### 7. Custom Toolbars

**MT5 Implementation**:
- Right-click toolbar → Customize
- Add/remove buttons
- Reorder buttons
- Create custom button groups
- Save toolbar layout with profile

**Trading Engine**: ⚠️ Basic menu exists
- Fixed toolbar (TopToolbar.tsx)
- No customization UI
- Buttons not reorderable

**Estimated Effort**: Medium (3-4 weeks)

---

#### 8. Profile Import/Export

**MT5 Implementation**:
```
File → Profiles → Export Profile...
Creates: "MyProfile.tpl" file
├── Can be shared with other traders
├── Can be imported by others
├── Version-safe (compatible across builds)
└── Compressed for file size
```

**Trading Engine**: ❌ Not implemented
- No export mechanism
- No import UI
- No file handling

**Estimated Effort**: Low-Medium (2-3 weeks)

---

#### 9. General Settings Persistence

**MT5 Implementation**:
- 100+ settings stored
- Server preferences
- Chart preferences
- Trade preferences
- Notification preferences
- Language/locale settings

**Trading Engine**: ⚠️ Minimal (4 settings only)
```typescript
// Current state persisted:
selectedSymbol
chartType
timeframe
orderVolume

// Missing:
server selection
chart behavior
trade defaults
notifications
language/locale
automation rules
```

**Estimated Effort**: Low-Medium (2 weeks)

---

## Gap Analysis

### Critical Gaps (Must Have)

| # | Feature | Impact | Current State | Effort | Priority |
|---|---------|--------|---|---|---|
| 1 | Profile System | UI/UX Foundation | ❌ None | 2-3w | CRITICAL |
| 2 | Chart Templates | Core Trading Feature | ❌ None | 3-4w | CRITICAL |
| 3 | Settings Persistence | User Experience | ⚠️ 4 only | 2w | HIGH |
| 4 | Import/Export | Sharing & Backup | ❌ None | 2-3w | HIGH |

### Important Gaps (Should Have)

| # | Feature | Impact | Current State | Effort | Priority |
|---|---------|--------|---|---|---|
| 5 | Workspace Layout | Professional Use | ❌ None | 4-6w | MEDIUM |
| 6 | Indicator Settings | Advanced Trading | ❌ None | 6+w | MEDIUM |
| 7 | Custom Toolbars | Customization | ⚠️ Basic | 3-4w | MEDIUM |
| 8 | Profile Switching UI | Usability | ❌ None | 1-2w | MEDIUM |

### Nice-to-Have Gaps (Could Have)

| # | Feature | Impact | Current State | Effort | Priority |
|---|---------|--------|---|---|---|
| 9 | Multi-Monitor Support | Enterprise | ❌ None | 3-4w | LOW |
| 10 | EA Profile Support | (awaits EA system) | N/A | 16+w | LOW |
| 11 | Notifications Settings | Advanced | ❌ None | 2-3w | LOW |
| 12 | Keyboard Shortcuts | Power Users | ⚠️ Partial | 1w | LOW |

---

## Detailed Recommendations

### Phase 1: Foundation (Weeks 1-3)

#### 1.1 Profile Database Schema

**Backend**: Add profile storage in Go

```go
// backend/models/profile.go

type Profile struct {
    ID          string                 `json:"id"`
    AccountID   string                 `json:"accountId"`
    Name        string                 `json:"name"`        // "Scalping", "Swing", etc.
    Description string                 `json:"description"`
    IsDefault   bool                   `json:"isDefault"`

    // UI Configuration
    UIState     map[string]interface{} `json:"uiState"`

    // Chart Configuration
    Charts      []ChartConfig          `json:"charts"`

    // Settings
    Settings    ProfileSettings        `json:"settings"`

    // Metadata
    CreatedAt   time.Time              `json:"createdAt"`
    UpdatedAt   time.Time              `json:"updatedAt"`
    ExportedAt  *time.Time             `json:"exportedAt"`
}

type ChartConfig struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Symbol      string                 `json:"symbol"`
    Timeframe   string                 `json:"timeframe"`
    ChartType   string                 `json:"chartType"`
    Indicators  []IndicatorConfig      `json:"indicators"`
    Properties  map[string]interface{} `json:"properties"`
}

type ProfileSettings struct {
    DefaultSymbol  string `json:"defaultSymbol"`
    DefaultVolume  float64 `json:"defaultVolume"`
    OneClickTrade  bool   `json:"oneClickTrade"`
    ShowTradeHistory bool `json:"showTradeHistory"`
    // ... more settings
}
```

#### 1.2 Backend API Endpoints

```
POST   /api/profiles                    # Create new profile
GET    /api/profiles                    # List all profiles
GET    /api/profiles/{id}               # Get specific profile
PUT    /api/profiles/{id}               # Update profile
DELETE /api/profiles/{id}               # Delete profile
POST   /api/profiles/{id}/export        # Export as file
POST   /api/profiles/import             # Import from file
POST   /api/profiles/{id}/set-default   # Set as default
```

#### 1.3 Frontend Store

**Update**: `clients/desktop/src/store/useAppStore.ts`

```typescript
interface Profile {
  id: string;
  accountId: string;
  name: string;
  description: string;
  isDefault: boolean;
  uiState: Record<string, any>;
  charts: ChartConfig[];
  settings: ProfileSettings;
  createdAt: string;
  updatedAt: string;
}

interface ProfileState {
  profiles: Profile[];
  activeProfileId: string | null;
  isLoadingProfiles: boolean;

  // Actions
  loadProfiles: () => Promise<void>;
  createProfile: (name: string, description: string) => Promise<Profile>;
  updateProfile: (id: string, updates: Partial<Profile>) => Promise<void>;
  deleteProfile: (id: string) => Promise<void>;
  switchProfile: (id: string) => Promise<void>;
  exportProfile: (id: string) => Promise<Blob>;
  importProfile: (file: File) => Promise<Profile>;
}
```

#### 1.4 Profile Selector UI Component

**New File**: `clients/desktop/src/components/ProfileSelector.tsx`

```typescript
export const ProfileSelector: React.FC = () => {
  const { profiles, activeProfileId, switchProfile } = useProfileStore();

  return (
    <div className="flex items-center gap-2">
      <select
        value={activeProfileId || ''}
        onChange={(e) => switchProfile(e.target.value)}
      >
        {profiles.map(profile => (
          <option key={profile.id} value={profile.id}>
            {profile.name}
          </option>
        ))}
      </select>
      <button onClick={() => createNewProfile()}>+ New</button>
    </div>
  );
};
```

**Effort**: 2-3 weeks

---

### Phase 2: Chart Templates (Weeks 4-7)

#### 2.1 Chart Template Data Model

```go
type ChartTemplate struct {
    ID          string                 `json:"id"`
    ProfileID   string                 `json:"profileId"`
    Name        string                 `json:"name"`
    Symbol      string                 `json:"symbol"`
    Timeframe   string                 `json:"timeframe"`

    // Chart Configuration
    ChartType   string                 `json:"chartType"`
    Colors      ChartColors            `json:"colors"`

    // Indicators
    Indicators  []IndicatorTemplate    `json:"indicators"`

    // Drawings
    Drawings    []DrawingObject        `json:"drawings"`

    // Advanced
    GridEnabled bool                   `json:"gridEnabled"`
    GridStyle   string                 `json:"gridStyle"`
    Zoom        float64                `json:"zoom"`

    CreatedAt   time.Time              `json:"createdAt"`
}

type IndicatorTemplate struct {
    Type       string                 `json:"type"`      // "RSI", "MA", "MACD"
    Parameters map[string]interface{} `json:"parameters"`
    Colors     map[string]string      `json:"colors"`
    Visible    bool                   `json:"visible"`
}
```

#### 2.2 Template Persistence APIs

```
POST   /api/chart-templates                  # Save new template
GET    /api/chart-templates                  # List templates
GET    /api/chart-templates/{id}             # Get template
PUT    /api/chart-templates/{id}             # Update template
DELETE /api/chart-templates/{id}             # Delete template
POST   /api/chart-templates/{id}/apply       # Apply to current chart
```

#### 2.3 Frontend Chart Components Update

**Update**: `clients/desktop/src/components/TradingChart.tsx`

```typescript
const TradingChart: React.FC<TradingChartProps> = ({ symbol, timeframe }) => {
  const [template, setTemplate] = useState<ChartTemplate | null>(null);

  const saveAsTemplate = async (name: string) => {
    const template = {
      symbol,
      timeframe,
      chartType: chartType,
      indicators: selectedIndicators,
      // ... more config
    };
    await saveChartTemplate(template);
  };

  const applyTemplate = async (templateId: string) => {
    const template = await loadChartTemplate(templateId);
    // Apply template settings
    setChartType(template.chartType);
    loadIndicators(template.indicators);
    // ... more
  };

  return (
    <div>
      <select onChange={(e) => applyTemplate(e.target.value)}>
        <option>Load Template...</option>
        {templates.map(t => (
          <option key={t.id} value={t.id}>{t.name}</option>
        ))}
      </select>
      <button onClick={() => saveAsTemplate('My Template')}>
        Save Template
      </button>
      {/* Chart component */}
    </div>
  );
};
```

**Effort**: 3-4 weeks

---

### Phase 3: Enhanced Settings (Weeks 8-9)

#### 3.1 Expanded Settings

**Update**: `clients/desktop/src/store/useAppStore.ts`

```typescript
interface ProfileSettings {
  // Server
  serverName: string;

  // Chart Defaults
  defaultSymbol: string;
  defaultTimeframe: string;
  defaultChartType: 'candlestick' | 'line' | 'area';

  // Trade Defaults
  defaultVolume: number;
  defaultDeviation: number;
  oneClickTrade: boolean;

  // Chart Behavior
  showTradeHistory: boolean;
  showTradeLevels: boolean;
  dragTradeableLevels: boolean;
  preloadChartData: boolean;

  // Appearance
  theme: 'light' | 'dark';
  fontSize: number;

  // Notifications
  enableNotifications: boolean;
  enableSound: boolean;
  enableEmail: boolean;
}
```

#### 3.2 Enhanced Options Dialog

**Update**: `clients/desktop/src/components/settings/OptionsDialog.tsx`

Make all tabs functional with proper state management and API calls.

**Effort**: 2 weeks

---

### Phase 4: Import/Export (Weeks 10-11)

#### 4.1 Profile Export Format

**JSON Format** (.profile extension):
```json
{
  "version": "1.0",
  "profile": {
    "name": "Scalping Setup",
    "description": "Fast-paced 5m scalping",
    "settings": {...},
    "charts": [
      {
        "symbol": "EURUSD",
        "timeframe": "5m",
        "indicators": [...],
        "drawings": [...]
      }
    ]
  },
  "exportedAt": "2026-02-02T10:00:00Z",
  "exportedBy": "username@trading-engine"
}
```

#### 4.2 Export/Import APIs

```go
// Export endpoint
func ExportProfile(w http.ResponseWriter, r *http.Request) {
  profileID := mux.Vars(r)["id"]
  profile := db.GetProfile(profileID)

  data, _ := json.Marshal(profile)
  w.Header().Set("Content-Disposition",
    fmt.Sprintf("attachment; filename=%s.profile", profile.Name))
  w.Write(data)
}

// Import endpoint
func ImportProfile(w http.ResponseWriter, r *http.Request) {
  file, _ := r.FormFile("file")
  defer file.Close()

  var profile Profile
  json.NewDecoder(file).Decode(&profile)

  profile.ID = uuid.New().String()
  db.SaveProfile(profile)

  json.NewEncoder(w).Encode(profile)
}
```

#### 4.3 Frontend Import/Export UI

**New Component**: `clients/desktop/src/components/ProfileExportImport.tsx`

```typescript
export const ProfileExportImport: React.FC = () => {
  return (
    <div className="flex gap-2">
      <button
        onClick={() => exportProfile(activeProfile)}
        className="px-4 py-2 bg-blue-600 rounded"
      >
        Export Profile
      </button>

      <label className="px-4 py-2 bg-green-600 rounded cursor-pointer">
        Import Profile
        <input
          type="file"
          accept=".profile"
          onChange={(e) => importProfile(e.target.files[0])}
          className="hidden"
        />
      </label>
    </div>
  );
};
```

**Effort**: 2-3 weeks

---

### Phase 5: Advanced Features (Weeks 12+)

#### 5.1 Workspace Layout Persistence

High effort (4-6 weeks) - requires:
- Window position tracking
- Multi-monitor detection
- Layout restoration on startup
- Docking system redesign

#### 5.2 Indicator Parameter Templates

Medium effort (3-4 weeks) - requires:
- Indicator system first
- Parameter storage per indicator
- Quick-apply templates

#### 5.3 Custom Toolbar Support

Medium effort (3-4 weeks) - requires:
- Toolbar customization UI
- Button reordering
- Custom button creation
- Persistence with profile

---

## Feature Implementation Priority Matrix

```
High Impact, Low Effort:
  1. ✅ Profile System (foundation)
  2. ✅ Chart Templates
  3. ✅ Settings Persistence
  4. ✅ Import/Export

Medium Impact, Medium Effort:
  5. ⚠️ Profile Switching UI
  6. ⚠️ Custom Toolbar Support
  7. ⚠️ Indicator Settings

Low Impact, High Effort:
  8. 🔵 Workspace Layout
  9. 🔵 Multi-Monitor Support
  10. 🔵 EA Profile System
```

---

## Estimated Timeline

### MVP Profile System (8-10 weeks)

**Phase 1**: Foundation (2-3w)
- Profile database schema
- Backend CRUD APIs
- Frontend profile selector

**Phase 2**: Chart Templates (3-4w)
- Chart template model
- Save/load template APIs
- Template application UI

**Phase 3**: Settings (1-2w)
- Expand persisted settings
- Settings UI updates

**Phase 4**: Import/Export (2-3w)
- Export profile as file
- Import profile from file
- File format specification

**Total MVP**: 8-10 weeks
**Effort**: 1 senior developer

---

## Implementation Checklist

### Backend Tasks

- [ ] Define Profile data model
- [ ] Define ChartConfig data model
- [ ] Define ChartTemplate data model
- [ ] Create database migrations
- [ ] Implement Profile CRUD endpoints
- [ ] Implement ChartTemplate CRUD endpoints
- [ ] Implement profile export endpoint
- [ ] Implement profile import endpoint
- [ ] Add profile validation
- [ ] Add error handling
- [ ] Write integration tests

### Frontend Tasks

- [ ] Create ProfileSelector component
- [ ] Create ProfileManager component
- [ ] Create ChartTemplateSelector component
- [ ] Create SaveTemplateDialog
- [ ] Update OptionsDialog tabs (make functional)
- [ ] Update useAppStore (add profile state)
- [ ] Update TradingChart component
- [ ] Create ProfileExportImport component
- [ ] Update TopToolbar (add profile selector)
- [ ] Add profile switching logic
- [ ] Write component tests
- [ ] Test multi-profile workflows

### Database Tasks

- [ ] Add profiles table
- [ ] Add chart_templates table
- [ ] Add profile_settings table
- [ ] Create indexes on frequently queried fields
- [ ] Plan migration strategy

### Testing Tasks

- [ ] Unit tests for profile CRUD
- [ ] Integration tests for profile workflows
- [ ] E2E tests for profile switching
- [ ] E2E tests for template save/load
- [ ] E2E tests for import/export
- [ ] Performance tests (loading many profiles)
- [ ] Multi-browser testing

---

## Success Metrics

### Functional Metrics

| Metric | Target | Success Criteria |
|--------|--------|---|
| Profiles Created | Unlimited | Can create 100+ profiles |
| Profile Switching | <500ms | Switch in under 500ms |
| Template Save Time | <1s | Save template in <1s |
| Template Load Time | <1s | Load and apply in <1s |
| Export File Size | <100KB | Average profile export <100KB |
| Import Success Rate | 99.9% | Import compatible profiles reliably |

### User Experience Metrics

| Metric | Target | Success Criteria |
|--------|--------|---|
| Time to Create Profile | <30s | Easy profile creation |
| Time to Switch Profile | <5s | Quick profile switching |
| User Satisfaction | >4/5 | Traders prefer profile system |
| Adoption Rate | >60% | Majority of users use profiles |

---

## Risk Assessment

### High Risk

**Risk**: Profile data corruption on import/export
- **Mitigation**: Version control, validation, backup before import
- **Impact**: Could lose user settings

**Risk**: Performance degradation with many profiles
- **Mitigation**: Database indexing, pagination, caching
- **Impact**: Slow profile listing/switching

### Medium Risk

**Risk**: Backward compatibility on settings changes
- **Mitigation**: Version field in profiles, migration scripts
- **Impact**: Old profiles might not load

**Risk**: Multi-user conflicts (if shared accounts)
- **Mitigation**: Per-user profile namespace
- **Impact**: Settings overwriting between users

---

## Competitive Analysis

### MT5 Comparison

**Strengths of MT5 Profile System**:
- Unlimited profiles
- Instant switching
- Comprehensive settings coverage
- Good export/import

**Our Advantages**:
- Modern tech stack (React)
- Cloud-native architecture
- Better UI/UX potential
- More flexible customization

### Market Differentiators

To compete with MT5, we should:

1. **Faster Switching**: <500ms vs MT5's ~1-2s
2. **Better UI**: Modern drag-and-drop profile management
3. **Easier Sharing**: Direct profile sharing via URL/link
4. **Presets**: Pre-built professional profiles for different strategies
5. **Collaboration**: Share profiles with team members
6. **Version Control**: Profile history and rollback

---

## Roadmap Timeline

```
Week 1-3:   Foundation (Profile System)
Week 4-7:   Chart Templates
Week 8-9:   Enhanced Settings
Week 10-11: Import/Export
Week 12-14: Advanced Features (optional)
```

**MVP Release**: Week 8 (basic profiles + templates)
**Full Release**: Week 12 (all core features)

---

## Conclusion

### Current State

The Trading Engine has minimal profile support:
- Only 4 settings persisted
- No profile concept
- No chart templates
- Non-functional options dialog

### Recommended Approach

**Short-term (MVP)**: Focus on Phases 1-4
- Get profiles system working
- Add chart templates
- Import/export capability
- This gives users 80% of MT5 functionality

**Medium-term**: Add advanced features
- Workspace layouts
- Indicator settings
- Custom toolbars
- Multi-monitor support

**Long-term**: Enterprise features
- Profile sharing
- Collaboration
- Profile marketplace
- Performance analytics

### Success Definition

Platform has achieved parity with MT5 profiles when:
1. ✅ Users can create multiple named profiles
2. ✅ Users can switch profiles instantly
3. ✅ Chart templates persist across sessions
4. ✅ Settings persist across sessions
5. ✅ Profiles can be imported/exported
6. ✅ Professional traders prefer our system

---

**Document Version**: 1.0
**Created**: 2026-02-02
**Last Updated**: 2026-02-02
**Status**: Research Complete, Ready for Implementation Planning

---

## Appendix A: Data Model Examples

### Sample Profile Structure

```json
{
  "id": "prof-001",
  "accountId": "acc-123",
  "name": "Scalping EURUSD 5min",
  "description": "Fast-paced scalping strategy with RSI and MA",
  "isDefault": false,
  "settings": {
    "defaultSymbol": "EURUSD",
    "defaultTimeframe": "5m",
    "defaultVolume": 0.5,
    "oneClickTrade": true,
    "showTradeHistory": true
  },
  "charts": [
    {
      "id": "chart-001",
      "name": "Main Chart",
      "symbol": "EURUSD",
      "timeframe": "5m",
      "chartType": "candlestick",
      "indicators": [
        {
          "type": "RSI",
          "parameters": {
            "period": 14
          },
          "colors": {
            "line": "#FF6B6B",
            "level30": "#4ECDC4",
            "level70": "#4ECDC4"
          }
        },
        {
          "type": "MA",
          "parameters": {
            "period": 20,
            "method": "EMA"
          },
          "colors": {
            "line": "#45B7D1"
          }
        }
      ],
      "drawings": []
    }
  ],
  "createdAt": "2026-02-01T15:30:00Z",
  "updatedAt": "2026-02-02T10:00:00Z"
}
```

---

## Appendix B: API Specifications

### Create Profile Endpoint

```
POST /api/profiles
Content-Type: application/json

Request:
{
  "name": "My Profile",
  "description": "Trading strategy",
  "settings": {...}
}

Response (201 Created):
{
  "id": "prof-001",
  "accountId": "acc-123",
  "name": "My Profile",
  ...
}
```

### Switch Profile Endpoint

```
POST /api/profiles/{id}/activate
Authorization: Bearer {token}

Response (200 OK):
{
  "success": true,
  "message": "Profile activated",
  "profile": {...}
}
```

---

**End of Document**
