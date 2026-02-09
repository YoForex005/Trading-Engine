# Profile Features - Relevant Codebase Locations

**Reference Guide** | 2026-02-02

---

## Current Implementation Files

### Frontend (React/TypeScript)

#### Settings & State Management

**File**: `clients/desktop/src/store/useAppStore.ts`
- **Purpose**: Global application state (Zustand)
- **Currently Persists**: 4 settings only
  - `selectedSymbol`
  - `chartType`
  - `timeframe`
  - `orderVolume`
- **Issue**: Should expand to ~50 settings, move to profiles concept
- **Action**: Refactor to use profile-based persistence

```typescript
// Current persistence (lines 216-224)
{
  name: 'trading-app-storage',
  partialize: (state) => ({
    selectedSymbol: state.selectedSymbol,
    chartType: state.chartType,
    timeframe: state.timeframe,
    orderVolume: state.orderVolume,
  }),
}
```

**Lines**: 1-230

---

#### Settings Dialog (Non-functional)

**File**: `clients/desktop/src/components/settings/OptionsDialog.tsx`
- **Purpose**: Settings UI with tabs
- **Status**: ⚠️ Placeholder - most tabs non-functional
- **Tabs Defined**:
  - Server (network settings)
  - Charts (chart behavior)
  - Trade (trading defaults)
  - Automation (EA/bot settings)
  - Events (sound notifications)
  - Notifications (push/mobile)
  - Email (alert emails)
  - FTP (report publishing)
  - Community (integrations)
  - Signals (EA marketplace)
- **Issue**: UI exists but no backend integration
- **Action**: Implement tab functionality, connect to profiles API

**Key Sections**:
- Line 27: Tab type definition
- Line 43-54: Tab list
- Line 84-94: Tab content routing
- Line 111-146: ServerTab (empty)
- Line 148-194: ChartsTab (empty)
- Line 196-235: TradeTab (empty)
- Line 237-288: AutomationTab (mentions profiles but not implemented!)

**Interesting Note** (Line 251):
```typescript
<span>Disable strategies when the profile has been changed</span>
```
Someone anticipated profiles but never built the system!

**Lines**: 1-299

---

#### Toolbar Component

**File**: `clients/desktop/src/components/layout/TopToolbar.tsx`
- **Purpose**: Main toolbar with buttons
- **Current Features**: Fixed buttons only
- **Issue**: No customization, no toolbar settings
- **Action**: Will need integration with profile toolbar configs

---

#### Chart Components

**File**: `clients/desktop/src/components/TradingChart.tsx`
- **Purpose**: Main trading chart
- **Current State**: Individual component state
- **Issue**: Chart settings not persisted
- **Action**: Connect to profile system, support chart templates

**Related Files**:
- `clients/desktop/src/components/ChartWithIndicators.tsx`
- `clients/desktop/src/components/ChartWithHistory.tsx`
- `clients/desktop/src/components/ChartTabs.tsx`

---

#### App Root

**File**: `clients/desktop/src/App.tsx`
- **Purpose**: Main app component
- **Will Need**: Profile initialization on app startup
- **Action**: Add profile loader at boot

---

### Backend (Go)

#### Client Profile (Wrong Type!)

**File**: `backend/cbook/client_profile.go`
- **Purpose**: B-Book routing decision - tracks client trading behavior
- **NOT For**: User workspace profiles
- **Contains**: Win rates, toxicity scores, risk classification
- **Issue**: Confusing naming - this is NOT workspace profiles
- **Action**: Keep separate, don't confuse with UI profiles

**Key Struct** (lines 20-53):
```go
type ClientProfile struct {
    AccountID      int64
    UserID         string
    Username       string
    TotalTrades    int64
    WinRate        float64
    MaxDrawdown    float64
    ToxicityScore  float64
    Classification ClientClassification
    // ...
}
```

**Lines**: 1-407

---

#### Configuration

**File**: `backend/config/config.go`
- **Purpose**: Server configuration
- **Will Need**: Profile storage configuration
- **Action**: Add database connection for profiles

---

#### API Server

**File**: `backend/api/server.go`
- **Purpose**: Main API server setup
- **Will Need**: Register new profile endpoints
- **Action**: Add profile route handlers

---

#### Database

**File**: `backend/database/migrate.go`
- **Purpose**: Database migrations
- **Will Need**: New migration for profiles table
- **Action**: Create migration 008_create_profiles.go

---

## Files That Need Creation

### Backend

#### 1. Profile Models

**New File**: `backend/models/profile.go`

```go
package models

import "time"

type Profile struct {
    ID          string                 `json:"id" db:"id"`
    AccountID   string                 `json:"accountId" db:"account_id"`
    Name        string                 `json:"name" db:"name"`
    Description string                 `json:"description" db:"description"`
    IsDefault   bool                   `json:"isDefault" db:"is_default"`
    UIState     json.RawMessage        `json:"uiState" db:"ui_state"`
    Settings    json.RawMessage        `json:"settings" db:"settings"`
    CreatedAt   time.Time              `json:"createdAt" db:"created_at"`
    UpdatedAt   time.Time              `json:"updatedAt" db:"updated_at"`
}

type ChartTemplate struct {
    ID          string                 `json:"id" db:"id"`
    ProfileID   string                 `json:"profileId" db:"profile_id"`
    Name        string                 `json:"name" db:"name"`
    Symbol      string                 `json:"symbol" db:"symbol"`
    Timeframe   string                 `json:"timeframe" db:"timeframe"`
    ChartType   string                 `json:"chartType" db:"chart_type"`
    Config      json.RawMessage        `json:"config" db:"config"`
    CreatedAt   time.Time              `json:"createdAt" db:"created_at"`
    UpdatedAt   time.Time              `json:"updatedAt" db:"updated_at"`
}

type ProfileSettings struct {
    DefaultSymbol    string  `json:"defaultSymbol"`
    DefaultVolume    float64 `json:"defaultVolume"`
    OneClickTrade    bool    `json:"oneClickTrade"`
    ShowTradeHistory bool    `json:"showTradeHistory"`
    // ... expand as needed
}
```

---

#### 2. Profile Repository

**New File**: `backend/models/profile_repository.go`

```go
package models

type ProfileRepository interface {
    CreateProfile(profile *Profile) error
    GetProfile(id string) (*Profile, error)
    ListProfiles(accountID string) ([]Profile, error)
    UpdateProfile(profile *Profile) error
    DeleteProfile(id string) error
    SetDefaultProfile(accountID, profileID string) error
}
```

---

#### 3. Profile API Handler

**New File**: `backend/api/profiles.go`

```go
package api

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
)

// Handler functions
func CreateProfile(w http.ResponseWriter, r *http.Request) {
    // POST /api/profiles
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
    // GET /api/profiles/{id}
}

func ListProfiles(w http.ResponseWriter, r *http.Request) {
    // GET /api/profiles
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
    // PUT /api/profiles/{id}
}

func DeleteProfile(w http.ResponseWriter, r *http.Request) {
    // DELETE /api/profiles/{id}
}

func SwitchProfile(w http.ResponseWriter, r *http.Request) {
    // POST /api/profiles/{id}/activate
}

func ExportProfile(w http.ResponseWriter, r *http.Request) {
    // GET /api/profiles/{id}/export
}

func ImportProfile(w http.ResponseWriter, r *http.Request) {
    // POST /api/profiles/import
}

func SaveChartTemplate(w http.ResponseWriter, r *http.Request) {
    // POST /api/profiles/{id}/templates
}

func ListChartTemplates(w http.ResponseWriter, r *http.Request) {
    // GET /api/profiles/{id}/templates
}

func DeleteChartTemplate(w http.ResponseWriter, r *http.Request) {
    // DELETE /api/profiles/{id}/templates/{templateId}
}
```

---

#### 4. Database Migration

**New File**: `backend/db/migrations/008_create_profiles.go`

```go
package migrations

func init() {
    RegisterMigration(8, "Create profiles and templates tables", up008, down008)
}

func up008(db *sql.DB) error {
    // CREATE TABLE profiles
    // CREATE TABLE chart_templates
    // CREATE INDEX on account_id
    return nil
}

func down008(db *sql.DB) error {
    // DROP TABLE chart_templates
    // DROP TABLE profiles
    return nil
}
```

---

### Frontend

#### 1. Profile Store

**New File**: `clients/desktop/src/store/useProfileStore.ts`

```typescript
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

export interface Profile {
  id: string;
  accountId: string;
  name: string;
  description: string;
  isDefault: boolean;
  uiState: Record<string, any>;
  settings: ProfileSettings;
  createdAt: string;
  updatedAt: string;
}

export interface ChartTemplate {
  id: string;
  profileId: string;
  name: string;
  symbol: string;
  timeframe: string;
  config: Record<string, any>;
}

interface ProfileState {
  profiles: Profile[];
  activeProfileId: string | null;
  chartTemplates: ChartTemplate[];
  isLoading: boolean;
  error: string | null;

  // Actions
  loadProfiles: () => Promise<void>;
  createProfile: (profile: Omit<Profile, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  switchProfile: (id: string) => Promise<void>;
  deleteProfile: (id: string) => Promise<void>;
  updateProfile: (id: string, updates: Partial<Profile>) => Promise<void>;

  // Chart templates
  loadTemplates: (profileId: string) => Promise<void>;
  saveTemplate: (template: Omit<ChartTemplate, 'id'>) => Promise<void>;
  deleteTemplate: (id: string) => Promise<void>;
  applyTemplate: (id: string) => Promise<void>;

  // Import/Export
  exportProfile: (id: string) => Promise<Blob>;
  importProfile: (file: File) => Promise<Profile>;
}

export const useProfileStore = create<ProfileState>()(
  devtools(
    persist(
      (set) => ({
        profiles: [],
        activeProfileId: null,
        chartTemplates: [],
        isLoading: false,
        error: null,

        // Action implementations
        loadProfiles: async () => { /* ... */ },
        // ... other actions
      }),
      {
        name: 'profile-storage',
      }
    )
  )
);
```

---

#### 2. Profile Selector Component

**New File**: `clients/desktop/src/components/ProfileSelector.tsx`

```typescript
import React, { useEffect } from 'react';
import { useProfileStore } from '../store/useProfileStore';

export const ProfileSelector: React.FC = () => {
  const { profiles, activeProfileId, switchProfile, loadProfiles } = useProfileStore();

  useEffect(() => {
    loadProfiles();
  }, []);

  return (
    <div className="flex items-center gap-2">
      <label className="text-xs font-medium text-gray-400">Profile:</label>
      <select
        value={activeProfileId || ''}
        onChange={(e) => switchProfile(e.target.value)}
        className="bg-[#1e1e1e] border border-gray-700 rounded px-2 py-1 text-sm"
      >
        <option value="">Select Profile...</option>
        {profiles.map((profile) => (
          <option key={profile.id} value={profile.id}>
            {profile.name}
          </option>
        ))}
      </select>
      <button
        onClick={() => createNewProfile()}
        className="px-2 py-1 text-xs bg-blue-600 hover:bg-blue-500 rounded"
      >
        + New
      </button>
    </div>
  );
};
```

---

#### 3. Profile Manager Dialog

**New File**: `clients/desktop/src/components/ProfileManager.tsx`

```typescript
import React, { useState } from 'react';
import { useProfileStore } from '../store/useProfileStore';

export const ProfileManager: React.FC<{ isOpen: boolean; onClose: () => void }> = ({
  isOpen,
  onClose,
}) => {
  const { profiles, deleteProfile, createProfile, exportProfile, importProfile } =
    useProfileStore();
  const [newName, setNewName] = useState('');

  // Implementation details...

  return (
    // Modal UI for managing profiles
    <div>{/* ... */}</div>
  );
};
```

---

#### 4. Chart Template UI

**Update File**: `clients/desktop/src/components/TradingChart.tsx`

Add:
```typescript
// Template selector
const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);

const saveAsTemplate = async (name: string) => {
  // Save current chart config as template
};

const applyTemplate = async (templateId: string) => {
  // Load and apply template
};

// UI additions
<div className="flex gap-2">
  <select
    onChange={(e) => applyTemplate(e.target.value)}
    className="..."
  >
    <option>Load Template...</option>
    {chartTemplates.map((t) => (
      <option key={t.id} value={t.id}>{t.name}</option>
    ))}
  </select>
  <button onClick={() => saveAsTemplate('My Template')}>
    Save Template
  </button>
</div>
```

---

#### 5. Profile Export/Import Component

**New File**: `clients/desktop/src/components/ProfileExportImport.tsx`

```typescript
import React, { useRef } from 'react';
import { useProfileStore } from '../store/useProfileStore';

export const ProfileExportImport: React.FC = () => {
  const { profiles, activeProfileId, exportProfile, importProfile } = useProfileStore();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleExport = async () => {
    if (!activeProfileId) return;
    const blob = await exportProfile(activeProfileId);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `profile-${activeProfileId}.profile`;
    a.click();
  };

  const handleImport = async (file: File) => {
    await importProfile(file);
  };

  return (
    <div className="flex gap-2">
      <button
        onClick={handleExport}
        className="px-4 py-2 bg-blue-600 text-white rounded"
      >
        Export Profile
      </button>
      <label className="px-4 py-2 bg-green-600 text-white rounded cursor-pointer">
        Import Profile
        <input
          ref={fileInputRef}
          type="file"
          accept=".profile"
          onChange={(e) => e.target.files?.[0] && handleImport(e.target.files[0])}
          className="hidden"
        />
      </label>
    </div>
  );
};
```

---

## Integration Points

### 1. App Initialization

**Location**: `clients/desktop/src/main.tsx` or `clients/desktop/src/App.tsx`

```typescript
// On app startup
useEffect(() => {
  const { loadProfiles, switchProfile } = useProfileStore();

  // Load all profiles
  await loadProfiles();

  // Switch to default profile
  const defaultProfile = profiles.find(p => p.isDefault);
  if (defaultProfile) {
    await switchProfile(defaultProfile.id);
  }
}, []);
```

---

### 2. Settings Save/Load

**Current**: `clients/desktop/src/store/useAppStore.ts`
**Needed**: Integration with profile system

```typescript
// Instead of individual persist:
// Use profile-based persistence via useProfileStore
```

---

### 3. Chart Configuration

**Current**: `clients/desktop/src/components/TradingChart.tsx`
**Needed**: Template loading in chart effect

```typescript
useEffect(() => {
  const loadChartTemplate = async () => {
    if (!selectedTemplateId) return;
    const template = await fetchTemplate(selectedTemplateId);
    applyTemplateToChart(template);
  };

  loadChartTemplate();
}, [selectedTemplateId]);
```

---

### 4. Server Integration

**Location**: `backend/cmd/server/main.go`

```go
// Register profile routes
router.HandleFunc("/api/profiles", handlers.ListProfiles).Methods("GET")
router.HandleFunc("/api/profiles", handlers.CreateProfile).Methods("POST")
router.HandleFunc("/api/profiles/{id}", handlers.GetProfile).Methods("GET")
router.HandleFunc("/api/profiles/{id}", handlers.UpdateProfile).Methods("PUT")
router.HandleFunc("/api/profiles/{id}", handlers.DeleteProfile).Methods("DELETE")
router.HandleFunc("/api/profiles/{id}/activate", handlers.SwitchProfile).Methods("POST")
router.HandleFunc("/api/profiles/{id}/export", handlers.ExportProfile).Methods("GET")
router.HandleFunc("/api/profiles/import", handlers.ImportProfile).Methods("POST")
```

---

## File Creation Checklist

### Backend
- [ ] `backend/models/profile.go`
- [ ] `backend/models/profile_repository.go`
- [ ] `backend/api/profiles.go`
- [ ] `backend/db/migrations/008_create_profiles.go`

### Frontend
- [ ] `clients/desktop/src/store/useProfileStore.ts`
- [ ] `clients/desktop/src/components/ProfileSelector.tsx`
- [ ] `clients/desktop/src/components/ProfileManager.tsx`
- [ ] `clients/desktop/src/components/ProfileExportImport.tsx`

### Updates Required
- [ ] `backend/config/config.go` - Add profile DB config
- [ ] `backend/api/server.go` - Register profile routes
- [ ] `clients/desktop/src/App.tsx` - Initialize profiles on startup
- [ ] `clients/desktop/src/components/TradingChart.tsx` - Add template support
- [ ] `clients/desktop/src/components/settings/OptionsDialog.tsx` - Make functional
- [ ] `clients/desktop/src/components/layout/TopToolbar.tsx` - Add profile selector
- [ ] `clients/desktop/src/store/useAppStore.ts` - Integrate with profiles

---

## Testing Locations

### Backend Tests
- Create: `backend/api/profiles_test.go`
- Create: `backend/models/profile_repository_test.go`

### Frontend Tests
- Create: `clients/desktop/src/store/useProfileStore.test.ts`
- Create: `clients/desktop/src/components/ProfileSelector.test.tsx`
- Create: `clients/desktop/src/components/ProfileManager.test.tsx`

---

## API Endpoints to Implement

### Profile Management
```
POST   /api/profiles                      # Create new profile
GET    /api/profiles                      # List all profiles for account
GET    /api/profiles/{id}                 # Get specific profile
PUT    /api/profiles/{id}                 # Update profile
DELETE /api/profiles/{id}                 # Delete profile
POST   /api/profiles/{id}/activate        # Switch active profile
PUT    /api/profiles/{id}/set-default     # Set as default
```

### Chart Templates
```
POST   /api/profiles/{id}/templates       # Save new template
GET    /api/profiles/{id}/templates       # List templates in profile
GET    /api/profiles/{id}/templates/{tid} # Get specific template
PUT    /api/profiles/{id}/templates/{tid} # Update template
DELETE /api/profiles/{id}/templates/{tid} # Delete template
POST   /api/profiles/{id}/templates/{tid}/apply # Apply to chart
```

### Import/Export
```
GET    /api/profiles/{id}/export          # Export as file
POST   /api/profiles/import               # Import from file
POST   /api/profiles/{id}/export-all      # Export all profiles
POST   /api/profiles/import-all           # Import multiple profiles
```

---

## Database Schema (SQL)

### Profiles Table
```sql
CREATE TABLE profiles (
  id VARCHAR(36) PRIMARY KEY,
  account_id VARCHAR(36) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  is_default BOOLEAN DEFAULT FALSE,
  ui_state JSON,
  settings JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (account_id) REFERENCES accounts(id),
  INDEX idx_account_id (account_id),
  INDEX idx_is_default (is_default)
);
```

### Chart Templates Table
```sql
CREATE TABLE chart_templates (
  id VARCHAR(36) PRIMARY KEY,
  profile_id VARCHAR(36) NOT NULL,
  name VARCHAR(255) NOT NULL,
  symbol VARCHAR(20) NOT NULL,
  timeframe VARCHAR(10) NOT NULL,
  chart_type VARCHAR(50),
  config JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (profile_id) REFERENCES profiles(id),
  INDEX idx_profile_id (profile_id),
  INDEX idx_symbol_timeframe (symbol, timeframe)
);
```

---

**Reference Document**: PROFILE_CODEBASE_LOCATIONS.md
**Status**: Complete navigation guide
**Last Updated**: 2026-02-02

For full specifications, see: `MT5_PROFILE_FEATURES_COMPARISON.md`
For quick reference, see: `PROFILE_COMPARISON_MATRIX.md`
