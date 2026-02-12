#!/bin/bash
# cleanup_debug_statements.sh
# Removes fmt.Println debug statements from production code

set -e

echo "=== Backend Debug Statement Cleanup ==="
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Directories to exclude (CLI tools where fmt.Println is acceptable)
EXCLUDE_DIRS=(
    "./cmd/migrate"
    "./cmd/test_e2e"
    "./logging/examples"
)

# Build exclusion pattern for find
EXCLUDE_PATTERN=""
for dir in "${EXCLUDE_DIRS[@]}"; do
    EXCLUDE_PATTERN="$EXCLUDE_PATTERN -path '$dir' -prune -o"
done

# Count before cleanup
BEFORE_COUNT=$(grep -r "fmt.Println" --include="*.go" . | \
    grep -v "./cmd/migrate" | \
    grep -v "./cmd/test_e2e" | \
    grep -v "./logging/examples" | \
    wc -l)

echo -e "${YELLOW}Found $BEFORE_COUNT fmt.Println statements to clean${NC}"
echo ""

# Backup before modifying
BACKUP_DIR="./scripts/backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
echo -e "${GREEN}Created backup at: $BACKUP_DIR${NC}"

# Find and replace fmt.Println with log.Printf
FILES_MODIFIED=0
find . -type f -name "*.go" \
    ! -path "./cmd/migrate/*" \
    ! -path "./cmd/test_e2e/*" \
    ! -path "./logging/examples/*" \
    -exec grep -l "fmt\.Println" {} \; | \
    while read file; do
        # Backup original
        cp "$file" "$BACKUP_DIR/$(basename $file).bak"

        # Replace fmt.Println with log.Printf
        # This handles: fmt.Println("message") → log.Printf("message\n")
        sed -i 's/fmt\.Println(\(.*\))/log.Printf(\1 + "\\n")/g' "$file"

        # Better: replace with proper log.Printf format
        # fmt.Println(var) → log.Printf("%v\n", var)
        # This is safer but requires manual review

        echo -e "${GREEN}✓ Cleaned: $file${NC}"
        ((FILES_MODIFIED++))
    done

echo ""
echo -e "${GREEN}Modified $FILES_MODIFIED files${NC}"

# Count after cleanup
AFTER_COUNT=$(grep -r "fmt.Println" --include="*.go" . | \
    grep -v "./cmd/migrate" | \
    grep -v "./cmd/test_e2e" | \
    grep -v "./logging/examples" | \
    wc -l)

echo -e "${YELLOW}Remaining fmt.Println statements: $AFTER_COUNT${NC}"

if [ $AFTER_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓ All debug statements cleaned!${NC}"
else
    echo -e "${YELLOW}⚠ Manual review needed for remaining statements${NC}"
fi

echo ""
echo "=== Cleanup Summary ==="
echo "Before: $BEFORE_COUNT statements"
echo "After:  $AFTER_COUNT statements"
echo "Files:  $FILES_MODIFIED modified"
echo "Backup: $BACKUP_DIR"

# Suggest next steps
echo ""
echo "=== Next Steps ==="
echo "1. Review changes: git diff"
echo "2. Run build: go build ./cmd/server/"
echo "3. Run tests: go test ./..."
echo "4. If issues, restore: cp $BACKUP_DIR/*.bak <original-locations>"
