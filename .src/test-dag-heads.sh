#!/bin/bash
# Test the new DAG head commands (workflow-based)
# Tests: fork-start, fork-continue, fork-finish

set -e
cd "$(dirname "$0")"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║              DAG HEAD WORKFLOW TESTING                            ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Clean up
cd via
git checkout main 2>/dev/null || true
git branch | grep -vE "^\*|main" | xargs git branch -D 2>/dev/null || true
git reset --hard HEAD 2>/dev/null || true
git clean -fd 2>/dev/null || true
cd ..

total_tests=0
passed_tests=0

cleanup_test() {
    cd via
    git checkout main 2>/dev/null || true
    git reset --hard HEAD 2>/dev/null || true
    git clean -fd 2>/dev/null || true
    git branch | grep -vE "^\*|main" | xargs git branch -D 2>/dev/null || true
    cd ..
}

run_test() {
    local test_name=$1
    local command=$2
    local expected=$3

    ((total_tests++))
    echo "TEST $total_tests: $test_name"

    # Clean before test
    cleanup_test >/dev/null 2>&1

    output=$(eval "$command" 2>&1 || true)

    if echo "$output" | grep -q "$expected"; then
        echo "✅ PASS: Found '$expected'"
        ((passed_tests++))
    else
        echo "❌ FAIL: Expected '$expected'"
        echo "Full output:"
        echo "$output" | head -20
    fi

    # Clean after test
    cleanup_test >/dev/null 2>&1
    echo ""
}

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 1: fork-start (Start New Work)                       ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

run_test \
    "fork-start on main shows 'Start New Work'" \
    "make fork-start DRY_RUN=1" \
    "🌱 Start New Work"

run_test \
    "fork-start calls fork-branch" \
    "make fork-start DRY_RUN=1" \
    "→ Running: make fork-branch"

run_test \
    "fork-start creates branch" \
    "make fork-start DRY_RUN=1" \
    "🌿 Branch Management"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 2: fork-continue (Save Progress)                     ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Set up: create branch with changes
cd via
git checkout -b test-continue-branch 2>/dev/null
echo "test" > test-file.md
git add test-file.md
cd ..

run_test \
    "fork-continue on feature branch shows 'Continue Working'" \
    "make fork-continue DRY_RUN=1" \
    "💾 Continue Working"

run_test \
    "fork-continue calls fork-save" \
    "make fork-continue DRY_RUN=1" \
    "→ Running: make fork-save"

run_test \
    "fork-continue commits and pushes" \
    "make fork-continue DRY_RUN=1" \
    "🚀 Quick Save & Push"

# Clean up
cd via
git checkout main 2>/dev/null || true
git branch -D test-continue-branch 2>/dev/null || true
git reset --hard HEAD 2>/dev/null || true
cd ..

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 3: fork-continue Error Handling                      ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

run_test \
    "fork-continue on main branch shows warning" \
    "make fork-continue DRY_RUN=1" \
    "⚠️  Warning: You're on main branch"

run_test \
    "fork-continue with no changes shows success" \
    "cd via && git checkout -b test-no-changes 2>/dev/null && cd .. && make fork-continue DRY_RUN=1 && cd via && git checkout main 2>/dev/null && git branch -D test-no-changes 2>/dev/null" \
    "✅ No changes to save"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 4: fork-finish (Done with Work)                      ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Set up: create branch
cd via
git checkout -b test-finish-branch 2>/dev/null
cd ..

run_test \
    "fork-finish on feature branch shows 'Finish Work'" \
    "make fork-finish DRY_RUN=1" \
    "✅ Finish Work"

run_test \
    "fork-finish switches back to main" \
    "make fork-finish DRY_RUN=1 2>&1 | head -20" \
    "→ Running: git checkout main"

# Clean up
cd via
git checkout main 2>/dev/null || true
git branch -D test-finish-branch 2>/dev/null || true
cd ..

run_test \
    "fork-finish on main shows 'Already on main'" \
    "make fork-finish DRY_RUN=1" \
    "✅ Already on main"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 5: Workflow Integration                              ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

run_test \
    "fork-start shows repo context" \
    "make fork-start DRY_RUN=1" \
    "(via)"

run_test \
    "fork-continue shows repo context" \
    "cd via && git checkout -b test-context 2>/dev/null && echo 'test' > test.md && git add test.md && cd .. && make fork-continue DRY_RUN=1 && cd via && git checkout main 2>/dev/null && git branch -D test-context 2>/dev/null" \
    "(via)"

run_test \
    "fork-finish shows repo context" \
    "cd via && git checkout -b test-finish-ctx 2>/dev/null && cd .. && make fork-finish DRY_RUN=1 2>&1 | head -10 && cd via && git checkout main 2>/dev/null && git branch -D test-finish-ctx 2>/dev/null" \
    "(via)"

# Final cleanup
cd via
git checkout main 2>/dev/null || true
git reset --hard HEAD 2>/dev/null || true
git clean -fd 2>/dev/null || true
git branch | grep -vE "^\*|main" | xargs git branch -D 2>/dev/null || true
cd ..

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                         RESULTS                                   ║"
echo "╠══════════════════════════════════════════════════════════════════╣"
printf "║  Total:  %2d                                                       ║\n" $total_tests
printf "║  Passed: %2d                                                       ║\n" $passed_tests
printf "║  Failed: %2d                                                       ║\n" $((total_tests - passed_tests))
echo "╚══════════════════════════════════════════════════════════════════╝"

if [ $passed_tests -eq $total_tests ]; then
    echo ""
    echo "🎉 ALL DAG HEAD TESTS PASSED!"
    echo "✅ Workflow commands tested:"
    echo "  1. fork-start ✅"
    echo "  2. fork-continue ✅"
    echo "  3. fork-finish ✅"
    echo "  4. Integration ✅"
    echo "  5. Error handling ✅"
    exit 0
else
    exit 1
fi
