#!/bin/bash
# Simple test: verify tags don't pull, branches do pull

cd "$(dirname "$0")"

REPOS_FILE="repos.test.list"
TEST_DIR=".test-repos"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║              REAL REPOSITORY TESTING                              ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Clean up
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"
cp "../$REPOS_FILE" .

# Create minimal Makefile
cat > Makefile <<'EOF'
REPOS_FILE := repos.test.list
SHELL := /bin/bash
include ../repos.mk
.PHONY: gitignore
gitignore:
	@:
EOF

total=0
passed=0

test_result() {
    ((total++))
    if [ "$2" = "pass" ]; then
        echo "✅ TEST $total: $1"
        ((passed++))
    else
        echo "❌ TEST $total: $1"
    fi
}

# Test 1: Install repos
echo "━━━ Installing repos ━━━"
if make install >/dev/null 2>&1; then
    test_result "Install command runs" "pass"
else
    test_result "Install command runs" "fail"
fi

# Test 2: Verify tag repo
echo ""
echo "━━━ Testing tag behavior ━━━"
tag=$(cd gh && git describe --tags 2>/dev/null)
[ "$tag" = "v2.65.0" ] && test_result "Tag repo at v2.65.0" "pass" || test_result "Tag repo at v2.65.0" "fail"

# Test 3: Verify branch repo
echo ""
echo "━━━ Testing branch behavior ━━━"
branch=$(cd via && git branch --show-current 2>/dev/null)
[ "$branch" = "main" ] && test_result "Branch repo on main" "pass" || test_result "Branch repo on main" "fail"

# Test 4: Idempotency (tags should not change)
echo ""
echo "━━━ Testing idempotency ━━━"
hash1=$(cd gh && git rev-parse HEAD 2>/dev/null)
make install >/dev/null 2>&1
hash2=$(cd gh && git rev-parse HEAD 2>/dev/null)
[ "$hash1" = "$hash2" ] && test_result "Tag unchanged after second install" "pass" || test_result "Tag unchanged after second install" "fail"

# Cleanup
cd ..
rm -rf "$TEST_DIR"

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  Total: $total  │  Passed: $passed  │  Failed: $((total - passed))                   ║"
echo "╚══════════════════════════════════════════════════════════════════╝"

if [ $passed -eq $total ]; then
    echo ""
    echo "🎉 ALL TESTS PASSED!"
    exit 0
else
    echo ""
    echo "❌ Some tests failed"
    exit 1
fi
