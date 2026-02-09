# Windows PowerShell Troubleshooting Guide

**Last Updated:** January 2026

This comprehensive guide addresses the most common Windows PowerShell issues and provides actionable solutions based on current best practices and the latest fixes.

---

## Table of Contents
1. [PowerShell Not Found in PATH](#1-powershell-not-found-in-path)
2. [Execution Policy Restrictions](#2-execution-policy-restrictions)
3. [Windows Defender and Antivirus Blocking](#3-windows-defender-and-antivirus-blocking)
4. [Corrupted PowerShell Installation](#4-corrupted-powershell-installation)
5. [Windows Updates Breaking PowerShell](#5-windows-updates-breaking-powershell)
6. [PowerShell vs PowerShell Core Differences](#6-powershell-vs-powershell-core-differences)

---

## 1. PowerShell Not Found in PATH

### Problem Description
Error messages like:
- `exec: "powershell": executable file not found in %PATH%`
- `PowerShell is not recognized as an internal or external command`
- `pwsh.exe not recognized`

These errors indicate that PowerShell executables are not in your system's PATH environment variable.

### Root Causes
- Missing or corrupted PATH environment variable
- PowerShell executable location not registered
- System configuration issues after Windows updates
- Manual PowerShell installation in non-standard locations

### Solutions

#### Solution 1: Add PowerShell to PATH Environment Variable (Permanent Fix)

**Steps:**
1. Open Control Panel → System and Security → System
2. Click "Advanced system settings" on the left
3. Click "Environment Variables" button
4. In "System variables" section, find and select "Path"
5. Click "Edit"
6. Add these paths (if missing):
   - Windows PowerShell 5.1: `C:\Windows\System32\WindowsPowerShell\v1.0`
   - PowerShell 7+: `C:\Program Files\PowerShell\7` (or your installation path)
7. Click OK on all dialogs
8. **Restart your terminal or reboot system**

**PowerShell Command Alternative:**
```powershell
# Run as Administrator
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Windows\System32\WindowsPowerShell\v1.0", [EnvironmentVariableTarget]::Machine)
```

#### Solution 2: No-Reboot Workaround

If you don't want to reboot:
1. Open Task Manager (Ctrl+Shift+Esc)
2. Find "Windows Explorer" process
3. Right-click → Restart
4. This refreshes environment variables without full reboot

#### Solution 3: Manual PowerShell Shortcut

1. Navigate to: `C:\Windows\System32\WindowsPowerShell\v1.0`
2. Right-click `powershell.exe`
3. Select "Create shortcut"
4. Move shortcut to Desktop or preferred location
5. Pin to taskbar for easy access

#### Solution 4: Repair System Files

Run in Command Prompt (Admin):
```cmd
sfc /scannow
DISM /Online /Cleanup-Image /RestoreHealth
```

These commands repair damaged system files that may affect PATH registration.

#### Solution 5: Verify PowerShell Installation

**Check if PowerShell exists:**
```cmd
where powershell
where pwsh
```

**Check current PATH:**
```cmd
echo %PATH%
```

---

## 2. Execution Policy Restrictions

### Problem Description
Error messages like:
- `File cannot be loaded because running scripts is disabled on this system`
- `The execution of scripts is disabled on this system`
- Scripts fail to run despite being legitimate

### Understanding Execution Policies

**IMPORTANT:** PowerShell execution policies are NOT a security feature. They are a safety mechanism to prevent accidental script execution. As noted by Microsoft, they were "never meant to be a security control but were intended to prevent administrators from shooting themselves in the foot."

**Common Execution Policies:**
- **Restricted** (Default): No scripts can run
- **AllSigned**: Only scripts signed by trusted publisher
- **RemoteSigned**: Downloaded scripts must be signed; local scripts run freely
- **Unrestricted**: All scripts run (with prompt for downloaded scripts)
- **Bypass**: Nothing blocked, no warnings

### Solutions

#### Solution 1: Bypass Policy Temporarily (Recommended for Testing)

```powershell
# Run single script bypassing policy
PowerShell.exe -ExecutionPolicy Bypass -File .\script.ps1

# Start PowerShell session with bypass
PowerShell.exe -ExecutionPolicy Bypass
```

**Pros:** No permanent system changes, safe for testing
**Cons:** Must specify for each script execution

#### Solution 2: Set Execution Policy for Current User

```powershell
# Run as regular user (no admin needed)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Pros:** Only affects your user account, no admin rights needed
**Cons:** Applies to all scripts you run

#### Solution 3: Set Execution Policy System-Wide

```powershell
# Run as Administrator
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope LocalMachine
```

**Pros:** Applies to all users
**Cons:** Requires admin rights, affects entire system

#### Solution 4: Alternative Bypass Methods

**Pipe Script to PowerShell:**
```powershell
Get-Content .\script.ps1 | PowerShell.exe -NoProfile -
```

**Read and Execute:**
```powershell
Invoke-Expression (Get-Content .\script.ps1 -Raw)
```

**Download and Execute (from web):**
```powershell
Invoke-Expression (Invoke-WebRequest https://example.com/script.ps1).Content
```

#### Solution 5: Check Current Policy

```powershell
# Check current execution policy
Get-ExecutionPolicy

# Check all scope policies
Get-ExecutionPolicy -List
```

### Security Best Practices (2026)

Since execution policies are easily bypassed, implement these additional controls:

1. **AppLocker / Windows Defender Application Control (WDAC)**
   - Restrict which scripts and executables can run
   - True security boundary unlike execution policies

2. **PowerShell Script Block Logging**
   ```powershell
   # Enable via Group Policy or Registry
   # Computer Configuration → Administrative Templates → Windows Components → Windows PowerShell
   # Enable "Turn on PowerShell Script Block Logging"
   ```

3. **Constrained Language Mode**
   - Limits PowerShell capabilities in untrusted sessions
   - Prevents access to .NET types and COM objects

4. **Attack Surface Reduction Rules**
   - Built into Windows Defender
   - Blocks obfuscated scripts and Office VBA macros launching PowerShell

### Environment-Specific Recommendations

| Environment | Recommended Policy | Additional Controls |
|-------------|-------------------|---------------------|
| Development Workstations | RemoteSigned | Script Block Logging |
| Production Servers | AllSigned | AppLocker + Logging |
| Automation/CI-CD | Bypass (for specific scripts) | Pipeline security, code signing |
| End User Desktops | Restricted | WDAC, Attack Surface Reduction |

---

## 3. Windows Defender and Antivirus Blocking

### Problem Description
- PowerShell scripts blocked with "malicious content" warnings
- False positives on legitimate scripts
- Antivirus software preventing script execution
- Recent 2026 issues with Microsoft Activation Scripts (MAS) false positives

### Recent Issues (January 2026)

**Microsoft Defender False Positive Crisis:**
Windows Defender has been incorrectly flagging legitimate PowerShell scripts as `Trojan:PowerShell/FakeMas.DA!MTB`. This happened after cybercriminals exploited typosquatted domains to distribute malware, causing Microsoft to update Defender signatures too aggressively.

**Affected Scripts:**
- Microsoft Activation Scripts (MAS)
- Windows utility scripts
- System administration tools
- Legitimate automation scripts

### Solutions

#### Solution 1: Add Exclusions to Windows Defender (Use with Caution)

**Via Windows Security UI:**
1. Open Windows Security
2. Go to Virus & threat protection
3. Click "Manage settings" under Virus & threat protection settings
4. Scroll to "Exclusions"
5. Click "Add or remove exclusions"
6. Add your script file or folder

**Via PowerShell (Admin):**
```powershell
# Exclude specific file
Add-MpPreference -ExclusionPath "C:\Scripts\MyScript.ps1"

# Exclude entire folder
Add-MpPreference -ExclusionPath "C:\Scripts"

# Exclude PowerShell process (NOT RECOMMENDED)
Add-MpPreference -ExclusionProcess "powershell.exe"

# View current exclusions
Get-MpPreference | Select-Object -ExpandProperty ExclusionPath
```

**WARNING:** Only exclude scripts you trust completely. Verify script source and hash before excluding.

#### Solution 2: Verify and Sign Your Scripts

Scripts without digital signatures are more likely to be flagged.

**Sign a PowerShell Script:**
```powershell
# Get your code signing certificate
$cert = Get-ChildItem -Path Cert:\CurrentUser\My -CodeSigningCert

# Sign the script
Set-AuthenticodeSignature -FilePath .\script.ps1 -Certificate $cert
```

#### Solution 3: Use Defender's Cloud-Delivered Protection

Submit false positives to Microsoft:
1. Go to: https://www.microsoft.com/wdsi/filesubmission
2. Submit your script file
3. Provide details about why it's a false positive
4. Microsoft typically responds within 24-48 hours

#### Solution 4: Temporarily Disable Real-Time Protection (Testing Only)

**Via Windows Security:**
1. Open Windows Security
2. Virus & threat protection
3. Manage settings
4. Toggle off "Real-time protection"

**Via PowerShell (Admin):**
```powershell
# Disable (temporary - automatically re-enables)
Set-MpPreference -DisableRealtimeMonitoring $true

# Re-enable
Set-MpPreference -DisableRealtimeMonitoring $false
```

**WARNING:** Only for testing. System is vulnerable while disabled.

#### Solution 5: Third-Party Antivirus Configuration

**McAfee, Norton, Bitdefender, etc.:**
- Each has its own exclusion/whitelist system
- Look for "Application Control" or "Script Control" settings
- Some block PowerShell with `-ExecutionPolicy Unrestricted` flag

**Bitdefender Example:**
When powershell.exe is detected as potentially malicious:
1. Open Bitdefender
2. Go to Protection → Advanced Threat Defense
3. Add powershell.exe to exceptions (if trusted)

#### Solution 6: Download Scripts Safely

**Unblock Downloaded Scripts:**
```powershell
# Unblock single file
Unblock-File -Path .\script.ps1

# Unblock all scripts in folder
Get-ChildItem -Path C:\Scripts -Recurse | Unblock-File
```

Windows marks files downloaded from internet with "Zone.Identifier" alternate data stream. `Unblock-File` removes this mark.

### Best Practices for 2026

1. **Keep Defender Updated:** Ensure latest definitions to minimize false positives
2. **Use Microsoft Store Scripts:** Pre-vetted scripts less likely to be flagged
3. **Verify Script Hashes:** Compare with official sources before execution
4. **Monitor Defender Logs:** Review quarantined items regularly
   ```powershell
   Get-MpThreatDetection
   ```
5. **Use AppLocker:** Better than antivirus for script control

---

## 4. Corrupted PowerShell Installation

### Problem Description
- PowerShell commands fail unexpectedly
- Cmdlets missing or not recognized
- PowerShell crashes on startup
- Module loading failures
- SFC reports PowerShell files as corrupted

### Diagnosis

**Check PowerShell Version:**
```powershell
$PSVersionTable
```

**Test Basic Cmdlets:**
```powershell
Get-Command
Get-Process
Get-Help
```

**Check Module Paths:**
```powershell
$env:PSModulePath -split ';'
```

### Solutions

#### Solution 1: System File Checker (SFC) and DISM

**Most Common Fix for Corrupted PowerShell:**

```cmd
# Run Command Prompt as Administrator

# Step 1: Repair Windows Image
DISM /Online /Cleanup-Image /ScanHealth
DISM /Online /Cleanup-Image /RestoreHealth

# Step 2: Repair System Files
sfc /scannow
```

**Process Explanation:**
- **DISM**: Repairs Windows component store (source for system files)
- **SFC**: Replaces corrupted files using repaired component store
- Run DISM first, then SFC (order matters!)

**Expected Output:**
```
Windows Resource Protection found corrupt files and successfully repaired them.
```

#### Solution 2: Reset PowerShell Features

**For Windows 10/11:**
1. Open Control Panel
2. Programs → Programs and Features
3. Click "Turn Windows features on or off"
4. Find "Windows PowerShell 2.0" (if present)
5. Uncheck it → OK → Restart
6. Check it again → OK → Restart

**Via PowerShell (Admin):**
```powershell
# Disable and re-enable Windows PowerShell feature
Disable-WindowsOptionalFeature -Online -FeatureName MicrosoftWindowsPowerShellV2
Enable-WindowsOptionalFeature -Online -FeatureName MicrosoftWindowsPowerShellV2
```

#### Solution 3: Reinstall PowerShell 7+ (Not 5.1)

**PowerShell 5.1 CANNOT be uninstalled/reinstalled** (it's part of Windows), but PowerShell 7+ can:

**Via WinGet:**
```powershell
# Uninstall
winget uninstall Microsoft.PowerShell

# Reinstall latest
winget install Microsoft.PowerShell
```

**Via MSI (Manual):**
1. Download from: https://github.com/PowerShell/PowerShell/releases
2. Run the MSI installer
3. Follow installation wizard

**Via Microsoft Store:**
1. Search "PowerShell" in Microsoft Store
2. Install/Reinstall

#### Solution 4: In-Place Upgrade (Repair Install)

If DISM and SFC fail, perform an in-place Windows upgrade:

1. Download Windows Media Creation Tool
2. Create installation media or upgrade directly
3. Run setup.exe
4. Choose "Keep personal files and apps"
5. Complete upgrade

**This repairs Windows while keeping:**
- Personal files
- Installed applications
- User accounts and settings
- PowerShell installation

#### Solution 5: Manual Module Reinstallation

If specific modules are corrupted:

```powershell
# List installed modules
Get-Module -ListAvailable

# Remove corrupted module
Remove-Module -Name ModuleName -Force
Uninstall-Module -Name ModuleName

# Reinstall from PowerShell Gallery
Install-Module -Name ModuleName -Force

# Example: Reinstall PSReadLine
Uninstall-Module -Name PSReadLine
Install-Module -Name PSReadLine -Force
```

#### Solution 6: Check for Defender Module Corruption

SFC may flag Windows Defender PowerShell modules as corrupted. This is often a false positive.

**Known Issue:**
- Files: `C:\Windows\System32\WindowsPowerShell\v1.0\Modules\Defender\*`
- Microsoft KB: Files are actually intact despite SFC warnings

**Workaround:**
Ignore SFC warnings about Defender module unless Defender cmdlets actually fail.

### Prevention

1. **Regular Windows Updates:** Keeps system files current
2. **Backup System Regularly:** Create restore points before major changes
3. **Avoid Manual File Editing:** Don't modify PowerShell system files
4. **Use Package Managers:** WinGet, Chocolatey for clean installations

---

## 5. Windows Updates Breaking PowerShell

### Problem Description
- PowerShell functionality breaks after Windows updates
- Scripts that worked before updates now fail
- New error messages after Patch Tuesday
- PowerShell Direct (PSDirect) connection failures
- Module compatibility issues post-update

### Recent Issues (2026)

#### January 2026 Security Updates

**Affected Updates:**
- **KB5073724** (Windows 10 22H2)
- **KB5074109** / **KB5073455** (Windows 11)
- **KB5078127** (Out-of-band emergency update)

**Known PowerShell Issues:**

1. **PowerShell Direct Connection Failures:**
   - **Affected:** Devices with September 2025 updates (KB5065474, KB5065426)
   - **Symptom:** PSDirect fails between host and VM when not both fully updated
   - **Fix:** Install KB5066360 or later on both host and guest

2. **PowerShell Script Execution Error (CVE-2025-54100):**
   - **Affected:** Pre-December 2025 systems
   - **Fix:** KB5071546 (December security update)

3. **Remote Desktop and PowerShell Remoting:**
   - January 2026 updates caused RDP issues that affect PowerShell remoting
   - Monitor Windows Health Dashboard for updates

### Solutions

#### Solution 1: Verify Update Installation

**Check installed updates:**
```powershell
# Check for specific KB
Get-HotFix -Id KB5073724

# List all updates from January 2026
Get-HotFix | Where-Object {$_.InstalledOn -gt "2026-01-01"} | Sort-Object InstalledOn
```

#### Solution 2: Install Missing Updates

**Via Windows Update:**
1. Settings → Windows Update
2. Check for updates
3. Install all available updates
4. Restart

**Via PowerShell (Admin):**
```powershell
# Install PSWindowsUpdate module
Install-Module -Name PSWindowsUpdate -Force

# Check for updates
Get-WindowsUpdate

# Install all available updates
Install-WindowsUpdate -AcceptAll -AutoReboot
```

#### Solution 3: Rollback Problematic Update

**If an update breaks PowerShell:**

```powershell
# Uninstall specific update
wusa /uninstall /kb:5073724

# Or use PowerShell
Remove-WindowsUpdate -KBArticleID KB5073724 -NoRestart
```

**Via Settings:**
1. Settings → Windows Update → Update history
2. Uninstall updates
3. Select problematic update → Uninstall

**IMPORTANT:** Only uninstall updates if you're certain they caused the issue. Security updates are critical.

#### Solution 4: Fix Windows Update Corruption

**If updates fail to install or PowerShell breaks during update:**

```cmd
# Run as Administrator

# Clean Windows Update components
net stop wuauserv
net stop cryptSvc
net stop bits
net stop msiserver

ren C:\Windows\SoftwareDistribution SoftwareDistribution.old
ren C:\Windows\System32\catroot2 catroot2.old

net start wuauserv
net start cryptSvc
net start bits
net start msiserver
```

**Then run DISM and SFC:**
```cmd
DISM /Online /Cleanup-Image /RestoreHealth
sfc /scannow
```

#### Solution 5: Use Windows Update Troubleshooter

**Automatic Fix:**
1. Settings → System → Troubleshoot → Other troubleshooters
2. Run "Windows Update" troubleshooter
3. Follow on-screen instructions

**Or download:**
- Microsoft Update Troubleshooter: https://aka.ms/wudiag

#### Solution 6: Monitor Microsoft Known Issues

**Stay informed:**
- **Microsoft Update Catalog:** https://www.catalog.update.microsoft.com/
- **Windows Health Dashboard:** https://learn.microsoft.com/windows/release-health/
- **Microsoft Q&A:** https://learn.microsoft.com/answers/

**Subscribe to notifications:**
- Windows release notes RSS feeds
- Microsoft Tech Community posts

### Best Practices

1. **Test Updates in Non-Production First:**
   - Use a VM or test machine
   - Wait 24-48 hours after Patch Tuesday before deploying

2. **Enable System Restore:**
   ```powershell
   # Enable System Protection
   Enable-ComputerRestore -Drive "C:\"

   # Create restore point before updates
   Checkpoint-Computer -Description "Before Windows Update"
   ```

3. **Backup Critical Scripts:**
   - Version control (Git)
   - Cloud backup
   - External storage

4. **Document Working Configurations:**
   ```powershell
   # Export PowerShell configuration
   $PSVersionTable | Out-File "PS-Config-Backup.txt"
   Get-Module -ListAvailable | Out-File "PS-Modules-Backup.txt"
   ```

5. **Use Windows Server Update Services (WSUS)** for enterprise:
   - Test updates before deployment
   - Control rollout timing
   - Quick rollback capabilities

### Update Verification Commands

```powershell
# Verify PowerShell works after update
Test-Connection localhost
Get-Process
Get-Service

# Check module functionality
Import-Module -Name Microsoft.PowerShell.Management
Get-Command -Module Microsoft.PowerShell.Management

# Test remoting
Test-WSMan -ComputerName localhost

# Verify execution policy unchanged
Get-ExecutionPolicy -List
```

---

## 6. PowerShell vs PowerShell Core Differences

### Overview

As of 2026, there are two main PowerShell versions:

| Feature | Windows PowerShell 5.1 | PowerShell 7+ |
|---------|------------------------|---------------|
| **Platform** | Windows only | Windows, macOS, Linux |
| **Framework** | .NET Framework 4.x | .NET Core / .NET 6+ |
| **Installation** | Built into Windows | Separate download |
| **Side-by-Side** | N/A | Runs alongside Windows PowerShell |
| **Updates** | Security fixes only | Active feature development |
| **Performance** | Slower | Faster, more efficient |
| **Module Compatibility** | All Windows modules | Core-compatible modules + compatibility layer |

### Key Differences

#### Platform and Architecture

**Windows PowerShell 5.1:**
- Windows-only
- Built on .NET Framework (full framework)
- Integrated into Windows OS
- Cannot be uninstalled
- Last feature update was 2016

**PowerShell 7+:**
- Cross-platform (Windows, macOS, Linux, ARM)
- Built on .NET Core / .NET 6+
- Standalone application
- Can be installed/uninstalled freely
- Regular feature updates (every 6 months)
- Optimized for cloud and automation

#### Module Compatibility

**Compatibility Issues:**

Not all Windows PowerShell modules work in PowerShell 7+, particularly:
- **Active Directory** (RSAT tools)
- **Group Policy**
- **Windows Workflow Foundation** (removed entirely)
- Some Exchange cmdlets
- Legacy SCOM/SCCM modules

**Check Module Compatibility:**
```powershell
# In PowerShell 7+
Get-Module -ListAvailable | Where-Object { $_.CompatiblePSEditions -notcontains 'Core' }
```

#### Windows PowerShell Compatibility Feature

PowerShell 7+ includes a compatibility layer for Windows modules:

```powershell
# Import Windows PowerShell module into PowerShell 7
Import-Module ActiveDirectory -UseWindowsPowerShell

# Create implicit remoting session to Windows PowerShell
New-PSSession -ConfigurationName Microsoft.PowerShell
```

**How it works:**
- Creates implicit remoting session to Windows PowerShell 5.1
- Proxies cmdlets from Windows PowerShell to PowerShell 7
- Automatic serialization/deserialization of objects

**Limitations:**
- Performance overhead
- Some object types don't serialize well
- Not all modules work via implicit remoting

#### Performance Comparison

**PowerShell 7+ Advantages:**
- **Faster startup:** .NET Core loads faster than .NET Framework
- **Better memory usage:** More efficient garbage collection
- **Pipeline performance:** Up to 2x faster for large datasets
- **Parallel processing:** `ForEach-Object -Parallel` (new in v7)

**Benchmark Example:**
```powershell
# Windows PowerShell 5.1: ~2.5 seconds
Measure-Command { 1..1000 | ForEach-Object { Start-Sleep -Milliseconds 1 } }

# PowerShell 7+ with parallel: ~0.1 seconds
Measure-Command { 1..1000 | ForEach-Object -Parallel { Start-Sleep -Milliseconds 1 } -ThrottleLimit 50 }
```

### Solutions and Best Practices

#### Solution 1: Use Both Side-by-Side

**Install PowerShell 7+ alongside Windows PowerShell:**

```powershell
# Install via WinGet
winget install Microsoft.PowerShell

# Or via Chocolatey
choco install powershell-core

# Or download from GitHub
# https://github.com/PowerShell/PowerShell/releases
```

**Access each version:**
- Windows PowerShell: `powershell.exe`
- PowerShell 7+: `pwsh.exe`

#### Solution 2: Migration Strategy

**For Organizations (2026 Recommendations):**

1. **Audit Current Scripts:**
   ```powershell
   # Check script compatibility
   Test-PSScriptAnalyzer -Path .\script.ps1 -Settings PSGallery
   ```

2. **Identify Dependencies:**
   - List all modules used
   - Check which are Core-compatible
   - Plan for incompatible modules

3. **Gradual Migration:**
   - Start with new scripts in PowerShell 7+
   - Migrate non-Windows-specific scripts
   - Keep Windows-specific scripts in 5.1

4. **Testing Environment:**
   - Test scripts in both environments
   - Validate automation workflows
   - Check scheduled tasks

#### Solution 3: Handle Compatibility in Scripts

**Detect PowerShell Version:**
```powershell
if ($PSVersionTable.PSVersion.Major -ge 7) {
    Write-Host "Running in PowerShell 7+"
    # Use PowerShell 7+ features
} else {
    Write-Host "Running in Windows PowerShell"
    # Use 5.1 compatible code
}
```

**Cross-Version Compatible Script Template:**
```powershell
#Requires -Version 5.1
<#
.SYNOPSIS
    Cross-version compatible script
#>

# Check edition
$isCore = $PSVersionTable.PSEdition -eq 'Core'

if ($isCore) {
    # PowerShell 7+ specific code
    $items | ForEach-Object -Parallel { Process-Item $_ } -ThrottleLimit 10
} else {
    # Windows PowerShell 5.1 fallback
    $items | ForEach-Object { Process-Item $_ }
}
```

#### Solution 4: Module Management

**Install Modules for Specific Versions:**

```powershell
# In Windows PowerShell 5.1
Install-Module -Name Az -Repository PSGallery -Scope CurrentUser

# In PowerShell 7+
Install-Module -Name Az -Repository PSGallery -Scope CurrentUser -Force
```

**Check Module Path:**
```powershell
# Different paths for each version
$env:PSModulePath -split ';'

# Windows PowerShell includes:
# C:\Program Files\WindowsPowerShell\Modules
# C:\Windows\System32\WindowsPowerShell\v1.0\Modules

# PowerShell 7+ includes:
# C:\Program Files\PowerShell\7\Modules
# C:\Program Files\PowerShell\Modules (shared)
```

#### Solution 5: Update Scripts for PowerShell 7+ Features

**Leverage new PowerShell 7+ features:**

**Parallel Processing:**
```powershell
# Old way (5.1)
$results = $servers | ForEach-Object { Test-Connection $_ }

# New way (7+) - much faster
$results = $servers | ForEach-Object -Parallel { Test-Connection $_ } -ThrottleLimit 20
```

**Ternary Operator:**
```powershell
# Old way (5.1)
$status = if ($service.Status -eq 'Running') { 'OK' } else { 'Failed' }

# New way (7+)
$status = $service.Status -eq 'Running' ? 'OK' : 'Failed'
```

**Null Coalescing:**
```powershell
# Old way (5.1)
$value = if ($null -ne $config) { $config } else { 'default' }

# New way (7+)
$value = $config ?? 'default'
```

**Pipeline Chain Operators:**
```powershell
# Run second command only if first succeeds
Get-Process powershell && Write-Host "PowerShell is running"

# Run second command only if first fails
Get-Process nonexistent || Write-Host "Process not found"
```

### When to Use Which Version (2026 Guidance)

#### Use Windows PowerShell 5.1 When:
- Managing Active Directory (RSAT required)
- Using Group Policy cmdlets
- Working with Exchange on-premises
- Running legacy enterprise scripts
- Module compatibility issues
- Required by enterprise policies

#### Use PowerShell 7+ When:
- Writing new scripts
- Cross-platform requirements
- Performance is critical
- Cloud automation (Azure, AWS)
- Modern module development
- CI/CD pipelines
- Container/Docker environments
- Need latest PowerShell features

#### Migration Timeline Recommendation

| Timeframe | Action |
|-----------|--------|
| **Now (2026)** | Install PowerShell 7+ alongside 5.1, start using for new projects |
| **2026-2027** | Migrate non-Windows-specific scripts, train team on v7 features |
| **2027-2028** | Primary use PowerShell 7+, maintain 5.1 for legacy only |
| **2028+** | PowerShell 7+ becomes standard, Windows PowerShell for exceptions |

**Note:** Windows PowerShell 5.1 will continue receiving security updates but no new features.

### Resources and Documentation

**Official Documentation:**
- [Differences between Windows PowerShell and PowerShell 7.x](https://learn.microsoft.com/en-us/powershell/scripting/whats-new/differences-from-windows-powershell)
- [Migrating from Windows PowerShell 5.1 to PowerShell 7](https://learn.microsoft.com/en-us/powershell/scripting/whats-new/migrating-from-windows-powershell-51-to-powershell-7)

**Download PowerShell:**
- https://github.com/PowerShell/PowerShell/releases
- Microsoft Store: "PowerShell"
- WinGet: `winget install Microsoft.PowerShell`

---

## Quick Reference: Diagnostic Commands

```powershell
# Check PowerShell Version
$PSVersionTable

# Check Execution Policy
Get-ExecutionPolicy -List

# Check PATH
$env:Path -split ';' | Select-String -Pattern 'PowerShell'

# Test PowerShell Health
Get-Command
Get-Module -ListAvailable
Test-WSMan

# Check Windows Updates
Get-HotFix | Sort-Object InstalledOn -Descending | Select-Object -First 10

# System File Check
# Run in CMD as Admin:
# DISM /Online /Cleanup-Image /RestoreHealth
# sfc /scannow

# Check Defender Exclusions
Get-MpPreference | Select-Object -ExpandProperty ExclusionPath

# Module Compatibility Check (PowerShell 7+)
Get-Module -ListAvailable | Select-Object Name, CompatiblePSEditions
```

---

## Additional Resources

### Official Microsoft Resources
- [PowerShell Documentation](https://learn.microsoft.com/en-us/powershell/)
- [PowerShell GitHub Repository](https://github.com/PowerShell/PowerShell)
- [Windows Update Troubleshooting](https://learn.microsoft.com/troubleshoot/windows-client/installing-updates-features-roles/fix-windows-update-errors)
- [Windows Defender False Positive Submission](https://www.microsoft.com/wdsi/filesubmission)

### Community Resources
- [PowerShell Community](https://devblogs.microsoft.com/powershell/)
- [PowerShell Gallery](https://www.powershellgallery.com/)
- [PowerShell Forums](https://forums.powershell.org/)
- [Reddit /r/PowerShell](https://reddit.com/r/PowerShell)

### Tools
- [PowerShell Script Analyzer](https://github.com/PowerShell/PSScriptAnalyzer)
- [Windows Terminal](https://github.com/microsoft/terminal)
- [VS Code PowerShell Extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode.PowerShell)

---

## Sources

This guide was compiled using the latest information from January 2026:

1. [What might be the issue if PowerShell not found in PATH](https://learn.microsoft.com/en-us/answers/questions/2194288/what-might-be-the-issue-if-the-error-message-exec)
2. [What to Do When Windows Cannot Find PowerShell](https://www.makeuseof.com/windows-cannot-find-powershell-fix/)
3. [15 Ways to Bypass the PowerShell Execution Policy](https://www.netspi.com/blog/technical-blog/network-pentesting/15-ways-to-bypass-the-powershell-execution-policy/)
4. [Microsoft Learn: About Execution Policies](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_execution_policies)
5. [PowerShell Execution Policy Security in 2026](https://www.realtimecyber.net/blog/powershell-execution-policy-bypass-how-attackers-do-it-and-how-to-prevent-it)
6. [Microsoft Defender Blocks Legitimate MAS Scripts](https://cybersecuritynews.com/microsoft-defenders-blocks-legitimate-mas/)
7. [Defender Blocks PowerShell Files Discussion](https://www.tenforums.com/antivirus-firewalls-system-security/166215-defender-sees-powershell-file-virus.html)
8. [How to re-install Windows PowerShell](https://www.elevenforum.com/t/how-to-re-install-windows-powershell.34167/)
9. [Guide: How to Reinstall Windows PowerShell](https://www.ninjaone.com/blog/how-to-reinstall-windows-powershell/)
10. [Need to repair PowerShell 5.1](https://learn.microsoft.com/en-us/answers/questions/2194267/need-to-repair-the-windows-powershell-5-1-to-fix-b)
11. [January 2026 Patch Tuesday Updates](https://www.archyde.com/january-2026-patch-tuesday-critical-windows-10-22h2-kb5073724-and-windows-11-kb5074109-kb5073455-security-updates-unveiled/)
12. [Windows 11 January 2026 Update Issues](https://www.windowslatest.com/2026/01/18/microsoft-confirms-windows-11-january-2026-update-issues-releases-fix-for-at-least-problems/)
13. [Differences between Windows PowerShell 5.1 and PowerShell 7.x](https://learn.microsoft.com/en-us/powershell/scripting/whats-new/differences-from-windows-powershell)
14. [What's the Difference Between PowerShell and PowerShell Core](https://www.majorgeeks.com/content/page/whats_the_difference_between_powershell_and_powershell_core.html)
15. [PowerShell v5 vs v7—Which to use and when](https://4sysops.com/archives/powershell-v5-and-v7which-to-use-and-when/)

---

**Document Version:** 1.0
**Last Updated:** January 27, 2026
**Maintained By:** Trading Engine Documentation Team
