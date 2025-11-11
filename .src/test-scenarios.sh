#!/bin/bash
# Comprehensive scenario testing for fork workflow
# Tests all 5 things DRY_RUN cannot test:
# 1. State-dependent logic (which menu shows)
# 2. Git operations (branch creation, commits)
# 3. Error conditions (uncommitted changes, detached HEAD)
# 4. Conditional menus based on state
# 5. Integration between commands

set -e
cd "$(dirname "$0")"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║              COMPREHENSIVE SCENARIO TESTING                       ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Clean up any leftover test branches from previous runs
cd via
git checkout main 2>/dev/null || true
git branch | grep -vE "^\*|main" | xargs git branch -D 2>/dev/null || true
cd ..

# Helper function to set up a scenario
setup_scenario() {
    local scenario=$1
    echo "━━━ Setting up scenario: $scenario ━━━"

    cd via
    git checkout main 2>/dev/null || true
    git reset --hard HEAD 2>/dev/null || true
    git clean -fd 2>/dev/null || true

    case $scenario in
        "main-clean")
            # Already set up
            echo "✅ State: main branch, no changes"
            ;;
        "main-dirty")
            echo "test" > .test-scenario
            git add .test-scenario
            echo "✅ State: main branch, uncommitted changes"
            ;;
        "feature-clean")
            git checkout -b test-scenario-branch 2>/dev/null || git checkout test-scenario-branch
            echo "✅ State: feature branch, no changes"
            ;;
        "feature-dirty")
            git checkout -b test-scenario-branch 2>/dev/null || git checkout test-scenario-branch
            echo "test" > .test-scenario
            git add .test-scenario
            echo "✅ State: feature branch, uncommitted changes"
            ;;
        "detached")
            git checkout HEAD~0 2>/dev/null || true
            echo "✅ State: detached HEAD"
            ;;
    esac
    cd ..
    echo ""
}

cleanup_scenario() {
    cd via
    git checkout main 2>/dev/null || true
    git reset --hard HEAD 2>/dev/null || true
    git clean -fd 2>/dev/null || true
    git branch -D test-scenario-branch 2>/dev/null || true
    cd ..
}

# Test counter
total_tests=0
passed_tests=0

run_test() {
    local test_name=$1
    local scenario=$2
    local command=$3
    local expected=$4

    ((total_tests++))
    echo "TEST $total_tests: $test_name"

    setup_scenario "$scenario"

    # Run command and capture output
    output=$(eval "$command" 2>&1 || true)

    if echo "$output" | grep -q "$expected"; then
        echo "✅ PASS: Found '$expected'"
        ((passed_tests++))
    else
        echo "❌ FAIL: Expected '$expected'"
        echo "Full output:"
        echo "$output" | head -15
    fi

    cleanup_scenario
    echo ""
}

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 1: State-Dependent Logic (Menu Variations)          ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# This tests what menu options appear based on state
run_test \
    "fork on main with no changes shows 'Start new work'" \
    "main-clean" \
    "make fork DRY_RUN=1" \
    "Start new work"

run_test \
    "fork on main with changes shows 'Create branch and save changes'" \
    "main-dirty" \
    "make fork DRY_RUN=1" \
    "Create branch and save changes"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 2: Git Operations (With Parameters)                 ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

run_test \
    "fork-branch actually creates branch" \
    "main-clean" \
    "make fork-branch BRANCH=test-git-op && git -C via branch | grep test-git-op" \
    "test-git-op"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 3: Error Conditions                                  ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

run_test \
    "Detached HEAD is detected and shows error" \
    "detached" \
    "make fork DRY_RUN=1" \
    "Detached HEAD"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 4: Conditional Menus                                 ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

run_test \
    "fork-commit on main shows warning menu" \
    "main-dirty" \
    "make fork-commit DRY_RUN=1" \
    "Warning: You're on main"

run_test \
    "fork-save on main shows warning menu" \
    "main-dirty" \
    "make fork-save DRY_RUN=1" \
    "Warning: You're on main"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         TEST 5: Command Integration                               ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

run_test \
    "fork-switch with non-existent branch calls fork-branch" \
    "main-clean" \
    "make fork-switch BRANCH=integration-test 2>&1" \
    "→ Running: make fork-branch"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                         RESULTS                                   ║"
echo "╠══════════════════════════════════════════════════════════════════╣"
printf "║  Total:  %2d                                                       ║\n" $total_tests
printf "║  Passed: %2d                                                       ║\n" $passed_tests
printf "║  Failed: %2d                                                       ║\n" $((total_tests - passed_tests))
echo "╚══════════════════════════════════════════════════════════════════╝"

if [ $passed_tests -eq $total_tests ]; then
    echo ""
    echo "🎉 ALL SCENARIOS PASSED!"
    echo "✅ All 5 limitations are now testable:"
    echo "  1. State-dependent logic ✅"
    echo "  2. Git operations ✅"
    echo "  3. Error conditions ✅"
    echo "  4. Conditional menus ✅"
    echo "  5. Command integration ✅"
    exit 0
else
    exit 1
fi
