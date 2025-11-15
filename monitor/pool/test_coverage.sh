#!/bin/bash

# Test Coverage Script for monitor/pool package
# Usage: ./test_coverage.sh [options]
#
# Options:
#   -h, --html     Generate HTML coverage report
#   -r, --race     Run with race detector
#   -v, --verbose  Verbose test output
#   -b, --bench    Run benchmarks
#   help           Show this help message

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
RUN_HTML=false
RUN_RACE=false
VERBOSE=""
RUN_BENCH=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        -h|--html)
            RUN_HTML=true
            shift
            ;;
        -r|--race)
            RUN_RACE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -b|--bench)
            RUN_BENCH=true
            shift
            ;;
        help)
            head -n 13 "$0" | tail -n 12
            exit 0
            ;;
        *)
            ;;
    esac
done

echo -e "${BLUE}=== Monitor/Pool Test Coverage ===${NC}\n"

# Run tests with coverage
echo -e "${YELLOW}Running tests...${NC}"
if [ "$RUN_RACE" = true ]; then
    echo -e "${YELLOW}(with race detector enabled)${NC}"
    CGO_ENABLED=1 go test $VERBOSE -race -coverprofile=coverage.out -covermode=atomic ./...
else
    go test $VERBOSE -coverprofile=coverage.out -covermode=atomic ./...
fi

if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passed!${NC}\n"
else
    echo -e "\n${RED}✗ Tests failed!${NC}\n"
    exit 1
fi

# Show coverage summary
echo -e "${BLUE}=== Coverage Summary ===${NC}"
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "${GREEN}Total Coverage: ${TOTAL_COVERAGE}${NC}\n"

# Show functions with < 80% coverage
echo -e "${YELLOW}Functions with < 80% coverage:${NC}"
go tool cover -func=coverage.out | awk '$3 != "100.0%" && $3 < 80.0 {print $1 "\t" $2 "\t" $3}'

# Show top 10 lowest coverage functions
echo -e "\n${YELLOW}Top 10 functions needing attention:${NC}"
go tool cover -func=coverage.out | grep -v "100.0%" | sort -k3 -n | head -10

# Generate HTML report if requested
if [ "$RUN_HTML" = true ]; then
    echo -e "\n${BLUE}Generating HTML coverage report...${NC}"
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}✓ HTML report generated: coverage.html${NC}"
    
    # Try to open in browser (Linux/Mac)
    if command -v xdg-open > /dev/null; then
        xdg-open coverage.html 2>/dev/null &
    elif command -v open > /dev/null; then
        open coverage.html 2>/dev/null &
    fi
fi

# Run benchmarks if requested
if [ "$RUN_BENCH" = true ]; then
    echo -e "\n${BLUE}=== Running Benchmarks ===${NC}"
    go test -bench=. -benchmem ./...
fi

echo -e "\n${GREEN}Done!${NC}"
