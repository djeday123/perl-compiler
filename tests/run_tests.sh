#!/bin/bash

# Perl Compiler Integration Test Runner
# Usage: ./run_tests.sh [options]
#
# Options:
#   -v          Verbose output
#   -f FILTER   Run only tests matching filter
#   -c          Run only compilation tests
#   -i          Run only interpreter tests
#   -q          Quick test (subset)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse arguments
VERBOSE=""
FILTER=""
TEST_TYPE=""

while getopts "vf:ciq" opt; do
    case $opt in
        v) VERBOSE="-v" ;;
        f) FILTER="$OPTARG" ;;
        c) TEST_TYPE="compile" ;;
        i) TEST_TYPE="interpret" ;;
        q) FILTER="Basic" ;;
        *) echo "Unknown option: $opt" >&2; exit 1 ;;
    esac
done

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  Perl Compiler Integration Tests${NC}"
echo -e "${BLUE}======================================${NC}"
echo

# Build perlc first
echo -e "${YELLOW}Building perlc...${NC}"
go build -o perlc ./cmd/perlc
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful${NC}"
echo

# Move to tests directory
cd tests

# Build test filter
TEST_ARGS=""
if [ -n "$FILTER" ]; then
    TEST_ARGS="-run $FILTER"
fi

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
echo

if go test $VERBOSE $TEST_ARGS -count=1 .; then
    echo
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}  All tests passed!${NC}"
    echo -e "${GREEN}======================================${NC}"
else
    echo
    echo -e "${RED}======================================${NC}"
    echo -e "${RED}  Some tests failed!${NC}"
    echo -e "${RED}======================================${NC}"
    exit 1
fi

# Cleanup
cd "$PROJECT_ROOT"
rm -f perlc perlc.exe

echo
echo -e "${BLUE}Done.${NC}"
