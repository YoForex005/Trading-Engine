# Trading Profile System - Implementation Checklist

## Phase 1: Foundation & Database (Week 1)

### Database Design & Migrations
- [ ] Create migration file: `backend/migrations/008_add_user_profiles_tables.sql`
- [ ] Define user_profiles table structure
- [ ] Add JSONB columns for each settings category
- [ ] Create indexes on (account_id, is_active) and (account_id, is_default)
- [ ] Add profile_audit_log table with full audit trail
- [ ] Create indicator_templates table
- [ ] Create drawing_templates table
- [ ] Create price_alerts table
- [ ] Create keyboard_shortcuts table
- [ ] Test migrations on local PostgreSQL instance
- [ ] Add rollback migrations for all tables

### Backend Models & Types (Go)
- [ ] Create `backend/internal/models/profile.go`:
  ```go
  type TradingProfile struct {
    ProfileID string `json:"profileId"`
    AccountID string `json:"accountId"`
    Name string `json:"name"`
    Category string `json:"category"`
    // ... all settings fields
    ChartSettings json.RawMessage `json:"chartSettings"`
    LayoutSettings json.RawMessage `json:"layoutSettings"`
    // ... continue for all categories
  }
  ```
- [ ] Create `backend/internal/models/profile_audit.go` for audit log
- [ ] Create type definitions for all settings categories
- [ ] Add validation functions for each profile type
- [ ] Create ProfileService interface:
  ```go
  type ProfileService interface {
    CreateProfile(ctx context.Context, profile *TradingProfile) error
    GetProfile(ctx context.Context, profileID string) (*TradingProfile, error)
    ListProfiles(ctx context.Context, accountID string) ([]*TradingProfile, error)
    UpdateProfile(ctx context.Context, profileID string, updates map[string]interface{}) error
    DeleteProfile(ctx context.Context, profileID string) error
    SetActiveProfile(ctx context.Context, profileID string) error
    ExportProfile(ctx context.Context, profileID string) ([]byte, error)
    ImportProfile(ctx context.Context, data []byte) (*TradingProfile, error)
  }
  ```

### Database Access Layer (Go)
- [ ] Create `backend/internal/repository/profile_repository.go`:
  - [ ] `CreateProfile(profile *TradingProfile) error`
  - [ ] `GetProfileByID(id string) (*TradingProfile, error)`
  - [ ] `GetProfilesByAccountID(accountID string) ([]*TradingProfile, error)`
  - [ ] `UpdateProfile(profile *TradingProfile) error`
  - [ ] `DeleteProfile(id string) error`
  - [ ] `GetActiveProfile(accountID string) (*TradingProfile, error)`
- [ ] Create `backend/internal/repository/audit_repository.go`:
  - [ ] `LogProfileChange(log *ProfileAuditLog) error`
  - [ ] `GetAuditLog(profileID string) ([]*ProfileAuditLog, error)`
- [ ] Add database connection pooling and error handling
- [ ] Create unit tests for repository layer

### Service Layer (Go)
- [ ] Create `backend/internal/service/profile_service.go`:
  - [ ] Implement ProfileService interface
  - [ ] Add input validation
  - [ ] Add authorization checks (user owns profile)
  - [ ] Add caching for active profile
  - [ ] Add profile switch logic
- [ ] Create `backend/internal/service/template_service.go`:
  - [ ] Indicator template CRUD
  - [ ] Drawing template CRUD
  - [ ] Template application logic
- [ ] Create `backend/internal/service/alert_service.go`:
  - [ ] Price alert CRUD
  - [ ] Alert validation
  - [ ] Alert notification logic

---

## Phase 2: API Endpoints (Week 2)

### Profile API Handlers
- [ ] Create `backend/internal/api/handlers/profiles.go`:
  - [ ] `HandleListProfiles(w, r)` - GET /api/user/profiles
  - [ ] `HandleGetProfile(w, r)` - GET /api/user/profiles/:id
  - [ ] `HandleCreateProfile(w, r)` - POST /api/user/profiles
  - [ ] `HandleUpdateProfile(w, r)` - PUT /api/user/profiles/:id
  - [ ] `HandleDeleteProfile(w, r)` - DELETE /api/user/profiles/:id
  - [ ] `HandleActivateProfile(w, r)` - POST /api/user/profiles/:id/activate
  - [ ] `HandleCloneProfile(w, r)` - POST /api/user/profiles/:id/clone
  - [ ] `HandleExportProfile(w, r)` - GET /api/user/profiles/:id/export
  - [ ] `HandleImportProfile(w, r)` - POST /api/user/profiles/import

- [ ] Add request/response structures:
  ```go
  type CreateProfileRequest struct {
    Name string `json:"name"`
    Category string `json:"category"`
    Description string `json:"description"`
    Settings json.RawMessage `json:"settings"`
  }

  type ProfileResponse struct {
    ProfileID string `json:"profileId"`
    Name string `json:"name"`
    // ... include all profile fields
  }
  ```

### Template API Handlers
- [ ] Create `backend/internal/api/handlers/templates.go`:
  - [ ] `HandleListIndicatorTemplates(w, r)` - GET /api/user/indicator-templates
  - [ ] `HandleCreateIndicatorTemplate(w, r)` - POST /api/user/indicator-templates
  - [ ] `HandleApplyIndicatorTemplate(w, r)` - POST /api/user/indicator-templates/:id/apply
  - [ ] Similar for drawing templates

### Alert API Handlers
- [ ] Create `backend/internal/api/handlers/alerts.go`:
  - [ ] `HandleListAlerts(w, r)` - GET /api/user/alerts
  - [ ] `HandleCreateAlert(w, r)` - POST /api/user/alerts
  - [ ] `HandleUpdateAlert(w, r)` - PUT /api/user/alerts/:id
  - [ ] `HandleDeleteAlert(w, r)` - DELETE /api/user/alerts/:id
  - [ ] `HandleTestAlert(w, r)` - POST /api/user/alerts/:id/test

### Keyboard Shortcut Handlers
- [ ] Create `backend/internal/api/handlers/shortcuts.go`:
  - [ ] `HandleListShortcuts(w, r)` - GET /api/user/shortcuts
  - [ ] `HandleCreateShortcut(w, r)` - POST /api/user/shortcuts
  - [ ] `HandleUpdateShortcut(w, r)` - PUT /api/user/shortcuts/:id
  - [ ] `HandleDeleteShortcut(w, r)` - DELETE /api/user/shortcuts/:id

### Route Registration
- [ ] Update `backend/cmd/server/main.go` or routing file to register all endpoints
- [ ] Add authentication middleware to protect user endpoints
- [ ] Add request validation middleware
- [ ] Add CORS headers for profile endpoints
- [ ] Add rate limiting for profile operations

### API Documentation
- [ ] Add OpenAPI/Swagger documentation for all endpoints
- [ ] Create API documentation in markdown
- [ ] Add example requests/responses
- [ ] Document error codes and responses

### Testing
- [ ] Create integration tests for all profile endpoints
- [ ] Test CRUD operations
- [ ] Test profile activation/switching
- [ ] Test export/import functionality
- [ ] Test error cases (unauthorized, not found, invalid data)
- [ ] Load test profile endpoints

---

## Phase 3: Frontend Service Layer (Week 3)

### Profile Manager Service
- [ ] Create `clients/desktop/src/services/profileManager.ts`:
  ```typescript
  export class ProfileManager {
    async listProfiles(): Promise<TradingProfile[]>
    async getProfile(profileId: string): Promise<TradingProfile>
    async createProfile(profile: TradingProfile): Promise<string>
    async updateProfile(profileId: string, updates: Partial<TradingProfile>): Promise<void>
    async deleteProfile(profileId: string): Promise<void>
    async activateProfile(profileId: string): Promise<void>
    async cloneProfile(profileId: string, newName: string): Promise<string>
    async exportProfile(profileId: string): Promise<Blob>
    async importProfile(file: File): Promise<string>
  }
  ```

### Local Storage Migration
- [ ] Create `clients/desktop/src/services/profileMigration.ts`:
  ```typescript
  export function migrateExistingSettingsToProfile(): TradingProfile {
    // Load from localStorage keys
    const appStorage = JSON.parse(localStorage.getItem('trading-app-storage') || '{}');
    const windowState = JSON.parse(localStorage.getItem('windowManagerState') || '{}');
    const marketWatchCols = JSON.parse(localStorage.getItem('rtx5_marketwatch_cols') || '[]');
    // ... load all settings

    // Create new profile
    return {
      profileId: generateUUID(),
      name: 'Default - Migrated',
      category: 'custom',
      chartPreferences: {
        chartType: appStorage.chartType || 'candlestick',
        timeframe: appStorage.timeframe || '1m',
        // ... map all settings
      },
      // ... continue mapping
    };
  }
  ```

### Zustand Store Update
- [ ] Update `clients/desktop/src/store/useAppStore.ts`:
  ```typescript
  interface AppState {
    // ... existing state
    currentProfile: TradingProfile | null;
    profiles: TradingProfile[];

    // New actions
    setCurrentProfile: (profile: TradingProfile) => void;
    setProfiles: (profiles: TradingProfile[]) => void;
    updateProfileSetting: (key: string, value: any) => void;
  }
  ```

- [ ] Keep existing localStorage for backward compatibility during transition
- [ ] Add profile sync on app load
- [ ] Add profile auto-save debounce (500ms)

### API Client Update
- [ ] Create `clients/desktop/src/services/api.ts` additions:
  ```typescript
  export const API = {
    profiles: {
      list: () => fetch(`${API_BASE}/api/user/profiles`),
      get: (id: string) => fetch(`${API_BASE}/api/user/profiles/${id}`),
      create: (profile: TradingProfile) =>
        fetch(`${API_BASE}/api/user/profiles`, {
          method: 'POST',
          body: JSON.stringify(profile)
        }),
      // ... etc
    }
  };
  ```

### TypeScript Types
- [ ] Create `clients/desktop/src/types/profile.ts`:
  ```typescript
  export interface TradingProfile {
    profileId: string;
    accountId: string;
    name: string;
    category: ProfileCategory;
    // ... all settings categories
  }
  ```

- [ ] Update `clients/desktop/src/types/trading.ts` if needed
- [ ] Export all profile types

### Context Provider (Optional)
- [ ] Create `clients/desktop/src/contexts/ProfileContext.tsx`:
  ```typescript
  export const ProfileContext = React.createContext<{
    currentProfile: TradingProfile | null;
    profiles: TradingProfile[];
    switchProfile: (id: string) => Promise<void>;
  }>(null);

  export function ProfileProvider({ children }) {
    // Provide profile state to entire app
  }
  ```

---

## Phase 4: UI Components (Week 4)

### Profile Selector Component
- [ ] Create `clients/desktop/src/components/ProfileSelector.tsx`:
  - [ ] Dropdown/combobox showing all profiles
  - [ ] "Active" indicator on current profile
  - [ ] Quick action buttons (edit, delete, clone)
  - [ ] "New Profile" button
  - [ ] Sort options (recent, name, category)

### Profile Management Modal
- [ ] Create `clients/desktop/src/components/ProfileModal.tsx`:
  - [ ] Create new profile form
  - [ ] Edit profile form
  - [ ] Delete confirmation
  - [ ] Profile preview/summary
  - [ ] Category selection with recommendations

### Settings Organization
- [ ] Reorganize settings into collapsible sections:
  - [ ] Chart Settings panel
  - [ ] Layout Settings panel
  - [ ] Trading Execution panel
  - [ ] Market Watch panel
  - [ ] Indicators panel
  - [ ] Drawings panel
  - [ ] Alerts panel
  - [ ] Notifications panel
  - [ ] Appearance panel

### Template Builder Components
- [ ] Create `clients/desktop/src/components/IndicatorTemplateBuilder.tsx`:
  - [ ] Add/remove indicators
  - [ ] Configure indicator parameters
  - [ ] Set colors and styling
  - [ ] Save as template
  - [ ] Load from template

- [ ] Create `clients/desktop/src/components/DrawingTemplateBuilder.tsx`:
  - [ ] Similar structure for drawing templates

### Profile Export/Import UI
- [ ] Create `clients/desktop/src/components/ProfileImportExport.tsx`:
  - [ ] Export button (downloads JSON)
  - [ ] Import button (file upload)
  - [ ] Show import preview
  - [ ] Merge or replace options

### Integration with Existing UI
- [ ] Add ProfileSelector to TopToolbar
- [ ] Add Settings access to MenuBar
- [ ] Add Profile info to StatusBar
- [ ] Update SettingsDialog to show profiles

### Styling
- [ ] Style all profile components with app theme
- [ ] Add dark/light mode support
- [ ] Add hover states and animations
- [ ] Ensure responsive design (responsive dropdowns, mobile-friendly modals)

---

## Phase 5: Testing & Polish (Week 5)

### Unit Tests (Frontend)
- [ ] Test ProfileManager service
  - [ ] API calls
  - [ ] Error handling
  - [ ] Data transformation

- [ ] Test profile migration
  - [ ] localStorage → profile conversion
  - [ ] Data integrity
  - [ ] Fallback defaults

- [ ] Test profile reducer/store
  - [ ] Profile switching
  - [ ] Setting updates
  - [ ] Profile sync

### Integration Tests (Frontend)
- [ ] E2E test: Create profile
- [ ] E2E test: Switch between profiles
- [ ] E2E test: Update profile settings
- [ ] E2E test: Export/import profile
- [ ] E2E test: Delete profile
- [ ] E2E test: Apply indicator template
- [ ] E2E test: Create price alert

### Backend Tests
- [ ] API endpoint tests (all CRUD operations)
- [ ] Authorization/authentication tests
- [ ] Input validation tests
- [ ] Error handling tests
- [ ] Database transaction tests
- [ ] Audit log tests

### Load Testing
- [ ] Test profile switching with 100 profiles
- [ ] Concurrent profile updates
- [ ] Large profile exports
- [ ] Import large profile files

### Migration Testing
- [ ] Test migration from existing app
- [ ] Verify no data loss
- [ ] Test rollback scenarios
- [ ] Performance impact assessment

### Documentation
- [ ] User guide: Creating and switching profiles
- [ ] User guide: Saving indicator templates
- [ ] User guide: Saving drawing templates
- [ ] Admin guide: Profile backup/restore
- [ ] Developer guide: Profile API
- [ ] API documentation (Swagger/OpenAPI)

### Bug Fixes & Polish
- [ ] Fix any UI glitches identified in testing
- [ ] Optimize performance (profile loading speed)
- [ ] Add loading indicators for async operations
- [ ] Add error notifications/toasts
- [ ] Improve accessibility (keyboard navigation, screen readers)
- [ ] Add tooltips/help text

### Release Preparation
- [ ] Feature flags (enable profiles gradually)
- [ ] A/B testing setup (old UI vs profiles)
- [ ] User communication/tutorials
- [ ] Backward compatibility testing
- [ ] Release notes preparation

---

## Quick Wins (Can Implement Immediately)

### Low-Effort, High-Value Features

1. **Persist Indicator Settings to localStorage** (1-2 hours)
   ```typescript
   // In indicatorManager.ts
   private saveIndicators(): void {
     localStorage.setItem('saved_indicators', JSON.stringify(this.indicators));
   }

   private loadIndicators(): void {
     const saved = localStorage.getItem('saved_indicators');
     if (saved) this.indicators = new Map(JSON.parse(saved));
   }
   ```

2. **Persist Drawing State to localStorage** (1-2 hours)
   ```typescript
   // In drawingManager.ts
   private saveDrawings(): void {
     localStorage.setItem('saved_drawings', JSON.stringify(Array.from(this.drawings)));
   }
   ```

3. **Add Basic Theme Toggle** (30 minutes)
   ```typescript
   export function toggleTheme() {
     const current = localStorage.getItem('theme') || 'dark';
     const next = current === 'dark' ? 'light' : 'dark';
     localStorage.setItem('theme', next);
     document.documentElement.classList.toggle('dark', next === 'dark');
   }
   ```

4. **Add Keyboard Shortcut Customization** (2-3 hours)
   - Simple form to rebind hotkeys
   - Store in localStorage: `rtx5_custom_shortcuts`
   - Load on app init

5. **Create 3 Default Profiles** (1 hour)
   - "Scalping" (1m chart, RSI, tight stops)
   - "Swing Trading" (1h chart, EMA, wider stops)
   - "Analysis" (4h chart, multiple indicators)

---

## File Structure After Implementation

```
Trading-Engine2/
├── backend/
│   ├── migrations/
│   │   └── 008_add_user_profiles_tables.sql
│   ├── internal/
│   │   ├── models/
│   │   │   ├── profile.go                    # NEW
│   │   │   ├── profile_audit.go              # NEW
│   │   │   └── template.go                   # NEW
│   │   ├── repository/
│   │   │   ├── profile_repository.go         # NEW
│   │   │   └── alert_repository.go           # NEW
│   │   ├── service/
│   │   │   ├── profile_service.go            # NEW
│   │   │   ├── template_service.go           # NEW
│   │   │   └── alert_service.go              # NEW
│   │   └── api/
│   │       └── handlers/
│   │           ├── profiles.go               # NEW
│   │           ├── templates.go              # NEW
│   │           ├── alerts.go                 # NEW
│   │           └── shortcuts.go              # NEW
│   └── cmd/
│       └── server/
│           └── main.go                       # UPDATED (register routes)
│
├── clients/
│   └── desktop/
│       └── src/
│           ├── services/
│           │   ├── profileManager.ts         # NEW
│           │   ├── profileMigration.ts       # NEW
│           │   ├── templateService.ts        # NEW
│           │   └── alertService.ts           # NEW
│           │
│           ├── stores/
│           │   └── useAppStore.ts            # UPDATED (add profile state)
│           │
│           ├── contexts/
│           │   └── ProfileContext.tsx        # NEW
│           │
│           ├── components/
│           │   ├── ProfileSelector.tsx       # NEW
│           │   ├── ProfileModal.tsx          # NEW
│           │   ├── SettingsPanel.tsx         # NEW (reorganized)
│           │   ├── IndicatorTemplateBuilder.tsx  # NEW
│           │   ├── DrawingTemplateBuilder.tsx    # NEW
│           │   └── ProfileImportExport.tsx   # NEW
│           │
│           └── types/
│               ├── profile.ts                # NEW
│               └── trading.ts                # UPDATED if needed
│
└── docs/
    ├── SETTINGS_PERSISTENCE_ARCHITECTURE.md
    ├── PROFILE_ARCHITECTURE_SUMMARY.md
    └── PROFILE_IMPLEMENTATION_GUIDE.md       # NEW (more detailed)
```

---

## Success Criteria

### Phase 1 Complete ✓
- [ ] Database tables created and tested
- [ ] All migrations working (up/down)
- [ ] ProfileService interface defined
- [ ] Basic CRUD operations in database

### Phase 2 Complete ✓
- [ ] All API endpoints working
- [ ] API endpoints tested with Postman/curl
- [ ] Error handling and validation working
- [ ] Documentation complete

### Phase 3 Complete ✓
- [ ] ProfileManager service functional
- [ ] Migration script working (localStorage → DB)
- [ ] Zustand store updated
- [ ] API client integration complete

### Phase 4 Complete ✓
- [ ] Profile selector visible in UI
- [ ] Profile switching works (UI updates reflected)
- [ ] All settings save to active profile
- [ ] Import/export functionality working
- [ ] UI looks polished and matches theme

### Phase 5 Complete ✓
- [ ] All tests passing
- [ ] No performance degradation
- [ ] Documentation complete
- [ ] Users can migrate from old settings
- [ ] Release notes ready

---

## Rollout Strategy

### Beta Phase (Days 1-7)
- [ ] Deploy to staging environment
- [ ] Internal testing (team uses new profiles)
- [ ] Collect feedback
- [ ] Fix critical bugs

### Canary Phase (Days 8-14)
- [ ] Deploy to 10% of users
- [ ] Monitor for errors
- [ ] Gather real-world feedback
- [ ] Iterate on UX

### General Availability (Day 15+)
- [ ] Deploy to 100% of users
- [ ] Monitor performance
- [ ] Provide user support
- [ ] Plan enhancements based on usage

### Deprecation Timeline (Month 2-3)
- [ ] Continue supporting old localStorage settings
- [ ] Add migration prompt for remaining users
- [ ] Eventually deprecate old system

---

## Risks & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Data loss in migration | Low | High | Test migration thoroughly, keep backups, rollback plan |
| Performance degradation | Medium | Medium | Profile caching, lazy loading, performance tests |
| User confusion with profiles | Medium | Medium | Clear documentation, in-app tutorials, gradual rollout |
| API bottleneck | Low | High | Rate limiting, caching, pagination |
| Storage explosion | Low | Medium | JSONB compression, old profile cleanup |
| Browser compatibility | Low | Low | Test in all supported browsers, polyfills |

---

## Budget Estimate

**Backend Development**: 40 hours
- Database design: 4h
- Models & service layer: 12h
- API endpoints: 16h
- Testing: 8h

**Frontend Development**: 50 hours
- Profile manager service: 8h
- Zustand store integration: 6h
- UI components: 20h
- Testing: 10h
- Polish & bug fixes: 6h

**DevOps & Testing**: 10 hours
- Database migrations: 2h
- Deployment setup: 3h
- Integration testing: 5h

**Documentation**: 8 hours
- API documentation: 3h
- User guides: 3h
- Developer guide: 2h

**Total: ~108 hours (~2.7 weeks for one full-time developer, or 1 week with 3 developers)**

