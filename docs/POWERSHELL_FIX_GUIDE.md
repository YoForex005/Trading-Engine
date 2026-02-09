# PowerShell Fix Guide - Complete Troubleshooting

This guide provides comprehensive solutions for PowerShell issues including installation, PATH configuration, VS Code integration, and execution policy fixes.

---

## Table of Contents
1. [Quick Diagnostics](#quick-diagnostics)
2. [Reinstalling PowerShell](#reinstalling-powershell)
3. [Adding PowerShell to PATH](#adding-powershell-to-path)
4. [Configuring VS Code Terminal](#configuring-vs-code-terminal)
5. [Fixing Execution Policy Issues](#fixing-execution-policy-issues)
6. [Alternative Terminal Options](#alternative-terminal-options)
7. [Troubleshooting Common Errors](#troubleshooting-common-errors)

---

## Quick Diagnostics

Before starting fixes, run these checks to identify the issue:

### Check PowerShell Installation
1. Press `Win + R`, type `powershell`, hit Enter
   - **Works?** PowerShell is installed but may not be in PATH
   - **Doesn't work?** PowerShell needs reinstalling

2. Open CMD (Win + R, type `cmd`) and run:
```cmd
where powershell
```
- **Shows path?** PowerShell exists but PATH may be incorrect
- **No output?** PowerShell is not in PATH or not installed

3. Check PowerShell version:
```powershell
$PSVersionTable.PSVersion
```

---

## 1. Reinstalling PowerShell

### Option A: Repair Windows PowerShell (Built-in)

Windows PowerShell comes with Windows. If it's missing or corrupted:

**Step 1: Enable Windows PowerShell via Windows Features**
1. Press `Win + R`, type `optionalfeatures.exe`, hit Enter
2. Scroll down and find "Windows PowerShell 2.0"
3. If unchecked, check it and click OK
4. Restart your computer

**Step 2: Run System File Checker**
1. Open CMD as Administrator:
   - Press `Win + X`
   - Select "Command Prompt (Admin)" or "Windows Terminal (Admin)"
2. Run:
```cmd
sfc /scannow
```
3. Wait for completion (may take 10-30 minutes)
4. Restart your computer

**Step 3: Run DISM Tool**
If SFC didn't fix the issue:
1. Open CMD as Administrator
2. Run:
```cmd
DISM /Online /Cleanup-Image /RestoreHealth
```
3. Wait for completion (may take 20-40 minutes)
4. Run `sfc /scannow` again
5. Restart your computer

### Option B: Install PowerShell 7+ (Modern, Recommended)

PowerShell 7 is cross-platform and more modern than Windows PowerShell 5.1.

**Method 1: Using winget (Windows 10/11)**
1. Open CMD or existing terminal:
```cmd
winget install --id Microsoft.Powershell --source winget
```

**Method 2: Using MSI Installer**
1. Download from: https://github.com/PowerShell/PowerShell/releases/latest
2. Look for `PowerShell-7.x.x-win-x64.msi`
3. Download and run the installer
4. Follow installation wizard (use default settings)
5. Check "Add PowerShell to PATH" option
6. Complete installation

**Method 3: Using Chocolatey**
If you have Chocolatey installed:
```cmd
choco install powershell-core -y
```

**Verify Installation:**
```cmd
pwsh --version
```

---

## 2. Adding PowerShell to PATH Manually

If PowerShell is installed but not found in terminal:

### Step 1: Locate PowerShell Executable

**Windows PowerShell (5.1):**
- 64-bit: `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`
- 32-bit: `C:\Windows\SysWOW64\WindowsPowerShell\v1.0\powershell.exe`

**PowerShell 7+:**
- Default: `C:\Program Files\PowerShell\7\pwsh.exe`

### Step 2: Add to System PATH

**Via GUI (Recommended):**
1. Press `Win + X`, select "System"
2. Click "Advanced system settings" (right sidebar)
3. Click "Environment Variables" button
4. Under "System variables", find and select "Path"
5. Click "Edit"
6. Click "New"
7. Add one of these paths:
   - For PowerShell 5.1: `C:\Windows\System32\WindowsPowerShell\v1.0`
   - For PowerShell 7: `C:\Program Files\PowerShell\7`
8. Click "OK" on all dialogs
9. **IMPORTANT:** Close and reopen all terminal windows

**Via PowerShell (As Administrator):**
```powershell
# For PowerShell 7
$newPath = "C:\Program Files\PowerShell\7"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
[Environment]::SetEnvironmentVariable("Path", "$currentPath;$newPath", "Machine")
```

**Via CMD (As Administrator):**
```cmd
setx /M PATH "%PATH%;C:\Program Files\PowerShell\7"
```

### Step 3: Verify PATH Addition

Close all terminals and open a new one:
```cmd
where powershell
where pwsh
echo %PATH%
```

---

## 3. Configuring VS Code Terminal Profiles

### Method 1: Via VS Code Settings UI

1. Open VS Code
2. Press `Ctrl + ,` (Settings)
3. Search for "terminal.integrated.profiles.windows"
4. Click "Edit in settings.json"

### Method 2: Direct settings.json Edit

1. Press `Ctrl + Shift + P`
2. Type "Preferences: Open User Settings (JSON)"
3. Add this configuration:

```json
{
  "terminal.integrated.profiles.windows": {
    "PowerShell": {
      "source": "PowerShell",
      "icon": "terminal-powershell"
    },
    "PowerShell 7": {
      "path": "C:\\Program Files\\PowerShell\\7\\pwsh.exe",
      "icon": "terminal-powershell",
      "args": ["-NoLogo"]
    },
    "Command Prompt": {
      "path": "C:\\Windows\\System32\\cmd.exe",
      "icon": "terminal-cmd"
    },
    "Git Bash": {
      "path": "C:\\Program Files\\Git\\bin\\bash.exe",
      "icon": "terminal-bash"
    }
  },
  "terminal.integrated.defaultProfile.windows": "PowerShell 7"
}
```

### Method 3: Workspace-Specific Configuration

1. Create `.vscode` folder in project root
2. Create `settings.json` inside it
3. Add terminal profiles (same as above)

### Troubleshooting VS Code Terminal

**Issue: "The terminal process failed to launch"**

Solution 1 - Check Path:
```json
{
  "terminal.integrated.profiles.windows": {
    "PowerShell": {
      "path": "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"
    }
  }
}
```

Solution 2 - Use source instead:
```json
{
  "terminal.integrated.profiles.windows": {
    "PowerShell": {
      "source": "PowerShell"
    }
  }
}
```

**Issue: Terminal opens but shows errors**

Check execution policy (see Section 4)

---

## 4. Fixing Execution Policy Issues

### Understanding Execution Policies

PowerShell execution policies control script execution:
- `Restricted` - No scripts allowed (default on Windows clients)
- `AllSigned` - Only signed scripts allowed
- `RemoteSigned` - Local scripts allowed, downloaded scripts must be signed
- `Unrestricted` - All scripts allowed (with warnings)
- `Bypass` - Nothing blocked, no warnings

### Check Current Policy

```powershell
Get-ExecutionPolicy
Get-ExecutionPolicy -List
```

### Fix Method 1: Set Policy for Current User (Recommended)

**Least intrusive, doesn't require admin rights:**

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Fix Method 2: Set Policy System-Wide

**Requires Administrator PowerShell:**

1. Right-click PowerShell, select "Run as administrator"
2. Run:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope LocalMachine
```

### Fix Method 3: Bypass for Single Session

**Temporary fix, no permanent changes:**

```powershell
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
```

### Fix Method 4: Unblock Downloaded Scripts

If you downloaded scripts from the internet:

```powershell
# Unblock single file
Unblock-File -Path "C:\path\to\script.ps1"

# Unblock all scripts in folder
Get-ChildItem -Path "C:\path\to\folder" -Recurse -Filter *.ps1 | Unblock-File
```

### Fix Method 5: VS Code Specific

Add to VS Code `settings.json`:

```json
{
  "terminal.integrated.profiles.windows": {
    "PowerShell": {
      "source": "PowerShell",
      "args": ["-ExecutionPolicy", "Bypass", "-NoLogo"]
    }
  }
}
```

### Group Policy Override

If execution policy is enforced by Group Policy:

1. Press `Win + R`, type `gpedit.msc`
2. Navigate to: `Computer Configuration → Administrative Templates → Windows Components → Windows PowerShell`
3. Double-click "Turn on Script Execution"
4. Set to "Enabled" and select "Allow local scripts and remote signed scripts"
5. Click OK and run `gpupdate /force`

---

## 5. Alternative Terminal Options

### Option A: Git Bash (Recommended for Cross-Platform Scripts)

**Install:**
1. Download from: https://git-scm.com/downloads
2. Run installer with default options
3. Ensure "Git Bash Here" is checked

**Use in VS Code:**
```json
{
  "terminal.integrated.profiles.windows": {
    "Git Bash": {
      "path": "C:\\Program Files\\Git\\bin\\bash.exe",
      "icon": "terminal-bash"
    }
  },
  "terminal.integrated.defaultProfile.windows": "Git Bash"
}
```

**Verify:**
```bash
git --version
bash --version
```

### Option B: Windows Command Prompt (CMD)

Built-in, always available:

**Use in VS Code:**
```json
{
  "terminal.integrated.profiles.windows": {
    "Command Prompt": {
      "path": "C:\\Windows\\System32\\cmd.exe",
      "icon": "terminal-cmd"
    }
  },
  "terminal.integrated.defaultProfile.windows": "Command Prompt"
}
```

### Option C: Windows Terminal (Modern, Recommended)

**Install via Microsoft Store:**
1. Open Microsoft Store
2. Search "Windows Terminal"
3. Click Install

**Or via winget:**
```cmd
winget install Microsoft.WindowsTerminal
```

**Features:**
- Multiple tabs
- Multiple profiles (PowerShell, CMD, Git Bash, WSL)
- Better customization
- GPU-accelerated rendering

### Option D: WSL (Windows Subsystem for Linux)

For Linux environment on Windows:

**Enable WSL:**
```cmd
wsl --install
```

**Use in VS Code:**
```json
{
  "terminal.integrated.profiles.windows": {
    "WSL": {
      "path": "wsl.exe",
      "icon": "terminal-ubuntu"
    }
  }
}
```

### Option E: PowerShell Core (pwsh.exe)

Modern cross-platform PowerShell:
- Faster startup
- Better Linux/Mac compatibility
- More features

**Already installed?** Check:
```cmd
pwsh --version
```

**Set as default in VS Code:**
```json
{
  "terminal.integrated.defaultProfile.windows": "PowerShell 7"
}
```

---

## 6. Troubleshooting Common Errors

### Error: "The term 'powershell' is not recognized"

**Cause:** PowerShell not in PATH

**Solutions:**
1. Follow [Section 2: Adding PowerShell to PATH](#2-adding-powershell-to-path-manually)
2. Use full path instead:
```cmd
"C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" -Command "Get-Date"
```

### Error: "File cannot be loaded because running scripts is disabled"

**Cause:** Execution policy is Restricted

**Solutions:**
1. Follow [Section 4: Fixing Execution Policy Issues](#4-fixing-execution-policy-issues)
2. Quick fix:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Error: "The system cannot find the path specified"

**Cause:** PowerShell path in VS Code settings is incorrect

**Solutions:**
1. Check actual PowerShell location:
```cmd
where powershell
where pwsh
```
2. Update VS Code `settings.json` with correct path
3. Or use `"source": "PowerShell"` instead of hardcoded path

### Error: "Access is denied" when changing execution policy

**Cause:** Insufficient privileges or Group Policy enforcement

**Solutions:**
1. Run PowerShell as Administrator
2. Use CurrentUser scope instead:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```
3. Check Group Policy (see Section 4, Fix Method 5)

### Error: VS Code terminal crashes immediately

**Cause:** Shell integration issues or corrupted profile

**Solutions:**
1. Disable shell integration temporarily:
```json
{
  "terminal.integrated.shellIntegration.enabled": false
}
```
2. Rename PowerShell profile to disable it:
```powershell
# Check if profile exists
Test-Path $PROFILE
# Rename it temporarily
Rename-Item $PROFILE "$PROFILE.backup"
```
3. Reset VS Code settings:
   - Delete `%APPDATA%\Code\User\settings.json`
   - Restart VS Code

### Error: "pwsh.exe" not found but PowerShell 7 is installed

**Cause:** Installation path is different or not in PATH

**Solutions:**
1. Find pwsh.exe location:
```cmd
dir "C:\Program Files\PowerShell" /s /b | findstr pwsh.exe
```
2. Add found path to VS Code settings
3. Or reinstall PowerShell 7 with default path

---

## 7. Quick Reference Commands

### PowerShell Diagnostics
```powershell
# Check version
$PSVersionTable.PSVersion

# Check execution policy
Get-ExecutionPolicy -List

# Check PATH
$env:Path -split ';' | Where-Object { $_ -like '*PowerShell*' }

# Test if command exists
Get-Command powershell -ErrorAction SilentlyContinue
Get-Command pwsh -ErrorAction SilentlyContinue
```

### CMD Diagnostics
```cmd
# Find PowerShell
where powershell
where pwsh

# Check PATH
echo %PATH%

# Check system variables
set PATH
```

### VS Code Terminal Commands

**Open terminal:** `` Ctrl + ` ``
**New terminal:** `Ctrl + Shift + ` `
**Select default profile:** `Ctrl + Shift + P` → "Terminal: Select Default Profile"
**Kill terminal:** Click trash icon or `Ctrl + Shift + P` → "Terminal: Kill Active Terminal Instance"

---

## 8. Recommended Setup for This Project

For the Trading Engine project, here's the optimal setup:

### Step 1: Install PowerShell 7
```cmd
winget install --id Microsoft.Powershell --source winget
```

### Step 2: Configure VS Code
Create/update `.vscode/settings.json`:
```json
{
  "terminal.integrated.profiles.windows": {
    "PowerShell 7": {
      "path": "C:\\Program Files\\PowerShell\\7\\pwsh.exe",
      "args": ["-NoLogo", "-ExecutionPolicy", "Bypass"],
      "icon": "terminal-powershell"
    },
    "Git Bash": {
      "path": "C:\\Program Files\\Git\\bin\\bash.exe",
      "icon": "terminal-bash"
    },
    "Command Prompt": {
      "path": "C:\\Windows\\System32\\cmd.exe",
      "icon": "terminal-cmd"
    }
  },
  "terminal.integrated.defaultProfile.windows": "PowerShell 7",
  "terminal.integrated.shellIntegration.enabled": true
}
```

### Step 3: Set Execution Policy
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Step 4: Verify Setup
```powershell
# In VS Code terminal
$PSVersionTable.PSVersion
Get-ExecutionPolicy
npm --version
node --version
go version
```

---

## 9. Emergency Fallback

If nothing works, use CMD for Go/Node commands:

### Create batch file: `run-backend.bat`
```batch
@echo off
cd backend
go run cmd/server/main.go
```

### Create batch file: `run-frontend.bat`
```batch
@echo off
cd clients\desktop
npm run dev
```

### Use in VS Code tasks.json
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Run Backend",
      "type": "shell",
      "command": "cmd",
      "args": ["/c", "run-backend.bat"],
      "problemMatcher": []
    }
  ]
}
```

---

## Support Resources

- **PowerShell Documentation:** https://docs.microsoft.com/en-us/powershell/
- **PowerShell GitHub:** https://github.com/PowerShell/PowerShell
- **VS Code Terminal Docs:** https://code.visualstudio.com/docs/terminal/profiles
- **Stack Overflow PowerShell Tag:** https://stackoverflow.com/questions/tagged/powershell

---

## Summary Checklist

- [ ] PowerShell is installed (run `powershell` in Run dialog)
- [ ] PowerShell is in PATH (run `where powershell` in CMD)
- [ ] VS Code terminal profiles are configured
- [ ] Execution policy is set to RemoteSigned or Bypass
- [ ] Default terminal profile is selected in VS Code
- [ ] All terminals closed and reopened after changes
- [ ] Can run `npm --version` and `go version` in terminal

**If all checked, your PowerShell setup is complete!**
