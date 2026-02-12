#!/bin/bash
# fix_panic_calls.sh
# Identifies and helps fix unsafe panic() calls in production code

set -e

echo "=== Unsafe Panic Call Audit ==="
echo ""

# Find all panic calls (excluding tests)
PANIC_FILES=$(grep -rn "panic(" --include="*.go" . | \
    grep -v "_test.go" | \
    grep -v "logging/examples" || true)

if [ -z "$PANIC_FILES" ]; then
    echo "✓ No unsafe panic() calls found!"
    exit 0
fi

echo "Found panic() calls in production code:"
echo ""
echo "$PANIC_FILES"
echo ""

echo "=== Recommended Fixes ==="
echo ""

# Parse each panic call and suggest fix
echo "$PANIC_FILES" | while IFS=: read -r file line code; do
    echo "File: $file:$line"
    echo "Code: $code"

    # Determine fix based on context
    if [[ "$code" == *"environment variable"* ]] || [[ "$code" == *"config"* ]]; then
        echo "Fix:  Replace with log.Fatalf() at startup (acceptable for missing config)"
        echo "      log.Fatalf(\"CRITICAL: %s\", \"$code\")"
    else
        echo "Fix:  Replace with proper error return"
        echo "      return fmt.Errorf(\"panic avoided: %w\", err)"
    fi

    echo ""
done

echo "=== Manual Actions Required ==="
echo ""
echo "The following files need manual review:"
echo ""

# List unique files
echo "$PANIC_FILES" | cut -d: -f1 | sort -u | while read file; do
    echo "  - $file"
done

echo ""
echo "Guidelines:"
echo "  1. Startup panics (missing config) → log.Fatalf()"
echo "  2. Runtime panics → return error"
echo "  3. Impossible states → log.Fatalf() + incident reporting"
echo "  4. External failures → circuit breaker pattern"

# Check specific known panics
echo ""
echo "=== Known Issues ==="
echo ""

if grep -q "panic(err)" fix/credentials.go 2>/dev/null; then
    echo "⚠ fix/credentials.go:409, 424"
    echo "  These panics occur when parsing credentials fails"
    echo "  Recommended: Return error to caller instead of panicking"
    echo ""
fi

if grep -q "panic.*TOTP_ENCRYPTION_KEY" auth/totp.go 2>/dev/null; then
    echo "⚠ auth/totp.go:28, 34"
    echo "  TOTP initialization panics on missing/invalid key"
    echo "  Recommended: Keep panic OR convert to log.Fatalf()"
    echo "  Reason: 2FA is critical security feature, server should not start without it"
    echo ""
fi
