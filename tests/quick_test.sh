#!/bin/bash

# Quick Test Runner for Perl Compiler
# Runs all .pl test files in tests/quick/
# Works on Linux, macOS, and Windows (MINGW/Git Bash)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Detect OS and set executable extension
EXE=""
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
    EXE=".exe"
fi

# Build perlc
echo -e "${BLUE}Building perlc...${NC}"
go build -o "perlc${EXE}" ./cmd/perlc
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build OK${NC}"
echo

# Test modes
MODE=${1:-both}  # interp, compile, or both

TOTAL=0
PASSED=0
FAILED=0

run_interp_test() {
    local file=$1
    local name=$(basename "$file" .pl)
    
    TOTAL=$((TOTAL + 1))
    
    echo -e "${CYAN}[$name] Interpreter...${NC}"
    output=$("./perlc${EXE}" "$file" 2>&1)
    
    if echo "$output" | grep -q "FAIL"; then
        echo -e "${RED}  FAILED${NC}"
        echo "$output" | grep "FAIL"
        FAILED=$((FAILED + 1))
        return 1
    elif echo "$output" | grep -qi "error"; then
        echo -e "${RED}  ERROR${NC}"
        echo "$output" | head -5
        FAILED=$((FAILED + 1))
        return 1
    else
        echo -e "${GREEN}  PASSED${NC}"
        PASSED=$((PASSED + 1))
        return 0
    fi
}

run_compile_test() {
    local file=$1
    local name=$(basename "$file" .pl)
    
    TOTAL=$((TOTAL + 1))
    
    echo -e "${CYAN}[$name] Compiled...${NC}"
    
    # Compile only first
    compile_out=$("./perlc${EXE}" -c -o "${name}_test" "$file" 2>&1)
    if [ $? -ne 0 ]; then
        echo -e "${RED}  COMPILE ERROR${NC}"
        echo "$compile_out" | head -10
        FAILED=$((FAILED + 1))
        return 1
    fi
    
    # Run the compiled exe
    if [ -f "${name}_test${EXE}" ]; then
        output=$("./${name}_test${EXE}" 2>&1)
    elif [ -f "${name}_test" ]; then
        output=$("./${name}_test" 2>&1)
    else
        echo -e "${RED}  EXE NOT FOUND${NC}"
        FAILED=$((FAILED + 1))
        return 1
    fi
    
    # Cleanup exe
    rm -f "${name}_test${EXE}" "${name}_test" 2>/dev/null
    
    if echo "$output" | grep -q "FAIL"; then
        echo -e "${RED}  FAILED${NC}"
        echo "$output" | grep "FAIL"
        FAILED=$((FAILED + 1))
        return 1
    else
        echo -e "${GREEN}  PASSED${NC}"
        PASSED=$((PASSED + 1))
        return 0
    fi
}

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  Quick Tests for Perl Compiler${NC}"
echo -e "${BLUE}======================================${NC}"
echo

# Find all test files
TEST_FILES=$(find tests/quick -name "*.pl" 2>/dev/null | sort)

if [ -z "$TEST_FILES" ]; then
    echo -e "${YELLOW}No test files found in tests/quick/${NC}"
    exit 0
fi

# Run tests
for file in $TEST_FILES; do
    echo -e "${YELLOW}--- Testing: $(basename $file) ---${NC}"
    
    if [ "$MODE" = "both" ] || [ "$MODE" = "interp" ]; then
        run_interp_test "$file"
    fi
    
    if [ "$MODE" = "both" ] || [ "$MODE" = "compile" ]; then
        run_compile_test "$file"
    fi
    
    echo
done

# Cleanup
rm -f "perlc${EXE}"
rm -f *.txt 2>/dev/null || true

# Summary
echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  Summary${NC}"
echo -e "${BLUE}======================================${NC}"
echo "Total:  $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
fi
