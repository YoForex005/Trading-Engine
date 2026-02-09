@echo off
echo ==========================================
echo      Installing Claude Code CLI
echo ==========================================

echo.
echo [1/3] Checking Node.js installation...
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Node.js is not found in your PATH.
    echo Please install Node.js from https://nodejs.org/ and try again.
    pause
    exit /b 1
)
node --version
echo Node.js is installed.

echo.
echo [2/3] Installing @anthropic-ai/claude-code...
echo This may take a few minutes...
call npm install -g @anthropic-ai/claude-code

if %errorlevel% neq 0 (
    echo.
    echo [WARNING] Default installation failed. Trying simple 'claude-code'...
    call npm install -g claude-code
)

echo.
echo [3/3] Verifying installation...
where claude >nul 2>nul
if %errorlevel% equ 0 (
    echo.
    echo [SUCCESS] Claude Code installed successfully!
    echo Run 'claude' to get started.
    claude --version
) else (
    echo.
    echo [ERROR] Installation might have failed or 'claude' is not in PATH.
    echo Please check the npm output above for errors.
)

pause
