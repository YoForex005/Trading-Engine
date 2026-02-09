# Backend Path Resolution Fixes

## Problem
The backend server was crashing when run from different working directories because config file paths were hardcoded relative paths that didn't account for whether the executable was run from project root or the backend/ subdirectory.

## Root Cause
1. **.env file location**: Located at `backend/.env` but `godotenv.Load()` was only looking in current directory
2. **LP config path**: Code used `data/lp_config.json` but actual file is at `backend/data/lp_config.json`
3. **Duplicate simulation code**: Live data simulation goroutine was placed before `hub` initialization, causing undefined variable error

## Fixes Applied

### 1. Config Loading (backend/config/config.go)
**Before:**
```go
func Load() (*Config, error) {
    // Try to load .env file (ignore error if not found)
    _ = godotenv.Load()
```

**After:**
```go
func Load() (*Config, error) {
    // Try to load .env file from multiple locations (ignore errors if not found)
    // Try project root first, then backend/ subdirectory
    if err := godotenv.Load(".env"); err != nil {
        _ = godotenv.Load("backend/.env")
    }
```

### 2. LP Config Path (backend/cmd/server/main.go)
**Before:**
```go
lpConfigPath := getProjectPath("data/lp_config.json")
```

**After:**
```go
lpConfigPath := getProjectPath("backend/data/lp_config.json")
```

### 3. Removed Duplicate Simulation Code
- Removed lines 334-414 (duplicate simulation goroutine that was before hub initialization)
- Kept lines 462-542 (correct placement after hub initialization)

## Path Resolution Strategy

The code already has a `getProjectPath()` helper function that:
1. Tries relative to current working directory
2. Tries relative to executable location
3. Falls back to the provided relative path

Combined with the working directory fix in main.go lines 285-292:
```go
// If running from 'backend' subdir, switch to parent (project root)
if filepath.Base(exeDir) == "backend" {
    projectRoot := filepath.Dir(exeDir)
    log.Printf("[Startup] Detected 'backend' subdirectory. Switching context to Project Root: %s", projectRoot)
    if err := os.Chdir(projectRoot); err != nil {
        log.Fatalf("Failed to switch to project root: %v", err)
    }
}
```

## Startup Script (start_services.bat)
The startup script correctly runs from project root:
```batch
@echo off
cd /d %~dp0           # Change to script directory (project root)
start "Trading Engine Backend" cmd /k "backend\server.exe"
```

## Result
✅ Backend can now be run from:
- Project root: `backend\server.exe`
- Backend directory: `.\server.exe`
- Startup script: `start_services.bat`

All config files resolve correctly in all scenarios.

## Testing
1. Run from project root: `backend\server.exe` ✓
2. Run from backend dir: `cd backend && .\server.exe` ✓
3. Run via startup script: `start_services.bat` ✓

## Files Modified
- `backend/config/config.go` - Multi-location .env loading
- `backend/cmd/server/main.go` - Fixed LP config path, removed duplicate code
- `backend/server.exe` - Rebuilt with fixes

Date: 2026-01-27
