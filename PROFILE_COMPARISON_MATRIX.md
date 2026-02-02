# MT5 vs Trading Engine - Profile Features Comparison Matrix

**Quick Visual Reference** | Updated: 2026-02-02

---

## Feature Comparison at a Glance

### Scoring Legend
- ✅ **Fully Implemented** = 100% feature parity
- ⚠️ **Partially Implemented** = 25-75% complete
- ❌ **Not Implemented** = 0% (missing)
- N/A = Not applicable

---

## Core Profile Features

### 1. Profiles & Workspaces

| Feature | MT5 | Trading Engine | Gap | Priority | Effort |
|---------|-----|---|---|---|---|
| **Named Profiles** | ✅ | ❌ | Create multiple profiles | CRITICAL | 2-3w |
| **Profile Switching** | ✅ | ❌ | Quick switch between profiles | CRITICAL | 1-2w |
| **Default Profile** | ✅ | ❌ | Auto-load on startup | HIGH | 1w |
| **Profile Deletion** | ✅ | ❌ | Delete unused profiles | MEDIUM | 1w |
| **Profile Renaming** | ✅ | ❌ | Rename profiles | MEDIUM | 1w |

**Current State**: Profile system foundation = **0% complete**
**MVP Requirement**: All features needed
**Complexity**: Low-Medium

---

### 2. Chart Templates & Indicators

| Feature | MT5 | Trading Engine | Gap | Priority | Effort |
|---------|-----|---|---|---|---|
| **Save Chart Template** | ✅ | ❌ | Save chart config as template | CRITICAL | 2w |
| **Load Chart Template** | ✅ | ❌ | Load saved chart config | CRITICAL | 1w |
| **Template List** | ✅ | ❌ | Browse saved templates | HIGH | 1w |
| **Template Names** | ✅ | ❌ | Name templates descriptively | MEDIUM | 1w |
| **Indicator Persistence** | ✅ | ❌ | Save indicator settings with template | CRITICAL | 2-3w |
| **Drawing Objects** | ✅ | ❌ | Save trendlines, channels, etc. | MEDIUM | 2-3w |

**Current State**: Chart templates = **0% complete**
**MVP Requirement**: Save/Load templates minimum
**Complexity**: Medium

---

### 3. Settings & Configuration

| Feature | MT5 | Trading Engine | Gap | Priority | Effort |
|---------|-----|---|---|---|---|
| **Default Symbol** | ✅ | ✅ | selectedSymbol persisted | DONE | - |
| **Default Timeframe** | ✅ | ✅ | timeframe persisted | DONE | - |
| **Chart Type** | ✅ | ✅ | chartType persisted | DONE | - |
| **Default Volume** | ✅ | ✅ | orderVolume persisted | DONE | - |
| **One-Click Trading** | ✅ | ⚠️ | Toggle setting only (UI exists) | MEDIUM | 1w |
| **Server Selection** | ✅ | ❌ | Select trading server | MEDIUM | 1w |
| **Notification Settings** | ✅ | ❌ | Sound/push/email alerts | MEDIUM | 2-3w |
| **Chart Behavior** | ✅ | ⚠️ | Options UI exists but not functional | MEDIUM | 1w |
| **Trade Defaults** | ✅ | ⚠️ | Options UI exists but not functional | MEDIUM | 1w |

**Current State**: Settings persistence = **20% complete** (4 of ~50 core settings)
**MVP Requirement**: Expand to ~30 settings minimum
**Complexity**: Low

---

### 4. Data Persistence & Storage

| Feature | MT5 | Trading Engine | Gap | Priority | Effort |
|---------|-----|---|---|---|---|
| **Local Storage** | ✅ | ⚠️ | Uses localStorage (4 settings) | MEDIUM | 1w |
| **Database Storage** | ✅ | ❌ | Backend profile storage needed | CRITICAL | 2-3w |
| **Cloud Sync** | ✅ | ❌ | Sync profiles across devices | LOW | 3-4w |
| **Encryption** | ✅ | ❌ | Encrypt sensitive settings | MEDIUM | 2w |
| **Backup/Restore** | ✅ | ❌ | Auto backup profiles | LOW | 2w |

**Current State**: Data persistence = **10% complete** (localStorage only)
**MVP Requirement**: Database backend needed
**Complexity**: Medium

---

### 5. Import/Export & Sharing

| Feature | MT5 | Trading Engine | Gap | Priority | Effort |
|---------|-----|---|---|---|---|
| **Export Profile** | ✅ | ❌ | Download profile as file | HIGH | 1-2w |
| **Import Profile** | ✅ | ❌ | Upload profile file | HIGH | 1-2w |
| **File Format** | ✅ | ❌ | Define .profile format | MEDIUM | 1w |
| **Version Control** | ✅ | ❌ | Profile versioning | LOW | 2w |
| **Share via URL** | ✅ | ❌ | Share profile link | LOW | 2-3w |
| **Validate Import** | ✅ | ❌ | Check profile compatibility | MEDIUM | 1w |

**Current State**: Import/Export = **0% complete**
**MVP Requirement**: Basic export/import needed
**Complexity**: Low-Medium

---

### 6. Workspace Layout Management

| Feature | MT5 | Trading Engine | Gap | Priority | Effort |
|---------|-----|---|---|---|---|
| **Window Positions** | ✅ | ❌ | Remember window positions | MEDIUM | 2w |
| **Window Sizes** | ✅ | ❌ | Remember window sizes | MEDIUM | 2w |
| **Multi-Monitor** | ✅ | ❌ | Support multi-screen layouts | LOW | 3-4w |
| **Panel Visibility** | ✅ | ⚠️ | Panels always visible (fixed) | MEDIUM | 1w |
| **Docking System** | ✅ | ⚠️ | Fixed docking only (no flexibility) | HIGH | 4-6w |
| **Zoom Levels** | ✅ | ❌ | Remember UI zoom | LOW | 1w |

**Current State**: Layout persistence = **5% complete** (fixed layout only)
**MVP Requirement**: Can defer (not critical for MVP)
**Complexity**: High

---

### 7. Advanced Features

| Feature | MT5 | Trading Engine | Gap | Priority | Effort |
|---------|-----|---|---|---|---|
| **Expert Advisors** | ✅ | ❌ | Different EAs per profile | N/A | 16+w |
| **Custom Toolbars** | ✅ | ⚠️ | Fixed toolbar only | LOW | 3-4w |
| **Macro/Scripts** | ✅ | ❌ | Automation scripts | LOW | 4-6w |
| **Market Depth Display** | ✅ | ❌ | Level 2 quotes | LOW | 2-3w |
| **Alert Conditions** | ✅ | ❌ | Profile-specific alerts | LOW | 2w |

**Current State**: Advanced features = **0% complete**
**MVP Requirement**: Can defer all (depends on EA system)
**Complexity**: Very High

---

## Overall Comparison Summary

### Feature Completeness by Category

```
PROFILES & SWITCHING:        ❌ 0%
├─ Named profiles            ❌ 0/1
├─ Profile switching         ❌ 0/1
├─ Default profiles          ❌ 0/1
└─ Profile management        ❌ 0/4

CHART TEMPLATES:             ❌ 0%
├─ Save template             ❌ 0/1
├─ Load template             ❌ 0/1
├─ Template list             ❌ 0/1
└─ Indicator persistence     ❌ 0/1

SETTINGS:                    ⚠️  20%
├─ Symbol                    ✅ 1/1
├─ Timeframe                 ✅ 1/1
├─ Chart type                ✅ 1/1
├─ Volume                    ✅ 1/1
├─ Trade settings            ⚠️  0.5/1
├─ Notifications             ❌ 0/1
└─ Other (~40 settings)      ❌ 0/40

DATA PERSISTENCE:            ⚠️  10%
├─ Local storage             ⚠️  0.5/1
├─ Database storage          ❌ 0/1
├─ Cloud sync                ❌ 0/1
└─ Encryption                ❌ 0/1

IMPORT/EXPORT:               ❌ 0%
├─ Export profile            ❌ 0/1
├─ Import profile            ❌ 0/1
├─ File format               ❌ 0/1
└─ Validation                ❌ 0/1

LAYOUT MANAGEMENT:           ⚠️  5%
├─ Window positions          ❌ 0/1
├─ Window sizes              ❌ 0/1
├─ Multi-monitor             ❌ 0/1
└─ Docking                   ⚠️  0.5/1

ADVANCED FEATURES:           ❌ 0%
├─ Expert Advisors           ❌ 0/1
├─ Custom toolbars           ❌ 0/1
├─ Macros                    ❌ 0/1
└─ Other                     ❌ 0/2

OVERALL PROFILE PARITY:      ⚠️  6% vs 100% in MT5
```

---

## Priority Grouping

### 🔴 CRITICAL (Must Have for MVP)
**Timeline: Weeks 1-10**

1. **Profile System** (2-3w)
   - Create, name, switch, delete profiles
   - Database backend

2. **Chart Templates** (3-4w)
   - Save/load chart configurations
   - Indicator persistence

3. **Settings Persistence** (1-2w)
   - Expand from 4 to ~30 settings
   - Make OptionsDialog functional

4. **Import/Export** (2-3w)
   - Export profile as file
   - Import profile from file

**Subtotal**: 8-10 weeks
**Effort**: 1 senior developer
**Business Impact**: HIGH (feature parity with MT5)

---

### 🟡 IMPORTANT (Should Have for v2)
**Timeline: Weeks 11-18**

5. **Workspace Layouts** (4-6w)
   - Window position/size persistence
   - Docking system improvements

6. **Indicator Settings** (3-4w)
   - Advanced indicator parameter templates
   - (Depends on indicator system)

7. **Custom Toolbars** (3-4w)
   - Toolbar customization UI
   - Button reordering/grouping

**Subtotal**: 10-14 weeks
**Effort**: 1-2 developers
**Business Impact**: MEDIUM (professional features)

---

### 🔵 NICE-TO-HAVE (Could Have for v3)
**Timeline: Weeks 19+**

8. **Multi-Monitor Support** (3-4w)
9. **Expert Advisor Profiles** (16+w, depends on EA system)
10. **Cloud Sync** (3-4w)
11. **Profile Marketplace** (4-6w)
12. **Collaboration Features** (6-8w)

**Subtotal**: 32+ weeks
**Effort**: 2-3 developers
**Business Impact**: LOW-MEDIUM (differentiation)

---

## Implementation Roadmap

### Timeline

```
Month 1 (Weeks 1-4): Foundation
  ├─ Week 1-2: Profile System (backend + frontend)
  ├─ Week 3-4: Chart Templates (basic)
  └─ Milestone: Profiles exist, can switch

Month 2 (Weeks 5-8): Templates & Settings
  ├─ Week 5-7: Chart Templates (complete)
  ├─ Week 8: Settings expansion (basic)
  └─ Milestone: Can save/load chart setups

Month 3 (Weeks 9-12): Polish & Export
  ├─ Week 9-10: Settings expansion (complete)
  ├─ Week 11-12: Import/Export
  └─ Milestone: MVP COMPLETE - Feature parity with MT5

Optional Phase 2 (Weeks 13+): Advanced
  ├─ Weeks 13-18: Layouts, toolbars, indicators
  ├─ Weeks 19+: Multi-monitor, EA support, marketplace
  └─ Milestone: v2 COMPLETE - Exceed MT5 features
```

---

## Resource Requirements

### Minimum Team
- **1 Senior Developer** (backend + frontend)
- **1 QA Engineer** (testing)
- **0.5 Product Manager** (planning/prioritization)

### Skill Requirements
- Backend: Go, database design, REST APIs
- Frontend: React, state management (Zustand), TypeScript
- Testing: Integration tests, E2E tests
- DevOps: Database migrations, CI/CD

### Tools & Technologies
- Backend Framework: Existing (Go/REST)
- Database: SQL (PostgreSQL/MySQL)
- Frontend: React, Zustand, TypeScript
- File Handling: JSON export/import
- Testing: Jest, Cypress, Go testing

---

## Success Metrics

### Feature Coverage
```
Target: 100% MT5 parity on core features (Phases 1-4)

Categories: 7 total
  ✅ Profiles & Switching
  ✅ Chart Templates
  ✅ Settings
  ✅ Data Persistence
  ✅ Import/Export
  ⚠️  Layout (defer to phase 2)
  ⚠️  Advanced Features (defer to phase 2+)

MVP Success = 5/7 categories complete = 71% coverage
```

### Performance Targets
| Metric | Target | Current |
|--------|--------|---------|
| Profile Switch Time | <500ms | N/A |
| Template Save Time | <1s | N/A |
| Settings Persist | <100ms | ✅ Fast |
| Import File Size | <200KB | N/A |
| Load Profiles | <2s (10 profiles) | N/A |

### User Adoption
| Metric | Target | Method |
|--------|--------|--------|
| Feature Awareness | >80% | In-app tutorial |
| Active Users | >60% | Analytics |
| User Satisfaction | >4.2/5 | Surveys |
| Support Tickets | <5% | Help desk tracking |

---

## Risk Assessment

### High Risk
1. **Profile Data Corruption**
   - Risk: Data loss on import/export
   - Mitigation: Validation, backup before import
   - Probability: Medium

2. **Performance Degradation**
   - Risk: Slow with many profiles
   - Mitigation: Database indexing, pagination
   - Probability: Low

### Medium Risk
3. **Backward Compatibility**
   - Risk: Old profiles incompatible
   - Mitigation: Version field, migration scripts
   - Probability: Medium

4. **User Confusion**
   - Risk: Complex UI, unclear features
   - Mitigation: Tutorials, good UX
   - Probability: Low

### Low Risk
5. **Technical Complexity**
   - Risk: Implementation issues
   - Mitigation: Clear specs, testing
   - Probability: Very Low (straightforward feature)

---

## Competitive Positioning

### How This Helps Us

1. **Feature Parity**: Match MT5 on profiles
2. **Better UX**: Modern React UI vs legacy MT5
3. **Faster Switching**: <500ms vs MT5's 1-2s
4. **Better Sharing**: Direct profile sharing
5. **Professional Appeal**: Attract traders from MT5

### Market Impact

- **Acquisition**: Traders switching from MT5
- **Retention**: More settings = more stickiness
- **Satisfaction**: Professional features = higher NPS
- **Differentiation**: Profile marketplace (future)

---

## Quick Decision Matrix

| Question | Answer | Action |
|----------|--------|--------|
| **Is this important?** | YES - core feature | ✅ Proceed |
| **Do we have time?** | YES - 8-10w fit | ✅ Schedule it |
| **Can we build it?** | YES - well-scoped | ✅ Allocate team |
| **Will users love it?** | YES - MT5 parity | ✅ Approve |
| **What's blocking?** | Nothing | ✅ Start now |

---

## Recommendation: GO/NO-GO

### Recommendation: **GO** ✅

**Rationale**:
1. Feature is well-understood (MT5 reference)
2. Effort is reasonable (8-10 weeks)
3. High user impact (professional feature parity)
4. Low technical risk (straightforward implementation)
5. Competitive necessity (trading platform baseline)

### Next Step
1. Review this document with team
2. Finalize feature scope for MVP
3. Create detailed database schema
4. Break into sprint tasks
5. Allocate developer resources
6. Start Phase 1 (Week 1)

---

**Document**: PROFILE_COMPARISON_MATRIX.md
**Format**: Quick reference visual guide
**Status**: Ready for stakeholder review
**Last Updated**: 2026-02-02

---

## Quick Links

- [Full Analysis](./MT5_PROFILE_FEATURES_COMPARISON.md)
- [Executive Summary](./PROFILE_RESEARCH_SUMMARY.md)
- [Codebase Locations](./CODEBASE_LOCATIONS.md) - See below

---

## Relevant Codebase Locations

### Frontend Components
- **Settings Dialog**: `clients/desktop/src/components/settings/OptionsDialog.tsx`
- **App Store**: `clients/desktop/src/store/useAppStore.ts`
- **Main Chart**: `clients/desktop/src/components/TradingChart.tsx`
- **Top Toolbar**: `clients/desktop/src/components/layout/TopToolbar.tsx`

### Backend Files
- **Client Profile** (wrong type): `backend/cbook/client_profile.go`
- **Config**: `backend/config/config.go`
- **API Server**: `backend/api/server.go`

### New Files To Create
- `backend/models/profile.go` - Profile data models
- `backend/models/chart_template.go` - Template data models
- `backend/api/profiles.go` - Profile endpoints
- `clients/desktop/src/components/ProfileSelector.tsx` - Profile UI
- `clients/desktop/src/components/ProfileManager.tsx` - Management UI
- `clients/desktop/src/store/useProfileStore.ts` - Profile state

---

**End of Quick Reference**
