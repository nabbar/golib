#!/bin/bash

# Script pour exécuter les tests AWS avec différentes options

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Display usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -v, --verbose           Verbose output
    -c, --coverage          Run with coverage
    -f, --focus PATTERN     Run only tests matching PATTERN
    -s, --skip PATTERN      Skip tests matching PATTERN
    -m, --minio             Ensure MinIO is used (no config.json)
    -a, --all               Run all tests (default)
    --s3                    Run only S3 tests
    --iam                   Run only IAM tests
    --bucket                Run only S3 Bucket tests
    --object                Run only S3 Object tests
    --pusher                Run only Pusher tests
    --user                  Run only IAM User tests
    --group                 Run only IAM Group tests
    --role                  Run only IAM Role tests
    --policy                Run only IAM Policy tests
    --cors                  Run only CORS tests
    --tags                  Run only object tagging tests
    --stress                Run only stress tests
    --advanced              Run only advanced features tests

Examples:
    $0                      # Run all tests
    $0 -c                   # Run all tests with coverage
    $0 --s3                 # Run only S3 tests
    $0 --iam -v             # Run only IAM tests with verbose output
    $0 -f "Bucket.*creation" # Run only tests matching pattern

EOF
    exit 0
}

# Parse arguments
VERBOSE=""
COVERAGE=""
FOCUS=""
SKIP=""
FORCE_MINIO=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -c|--coverage)
            COVERAGE="--cover --coverprofile=coverage.out"
            shift
            ;;
        -f|--focus)
            FOCUS="--focus=$2"
            shift 2
            ;;
        -s|--skip)
            SKIP="--skip=$2"
            shift 2
            ;;
        -m|--minio)
            FORCE_MINIO=true
            shift
            ;;
        -a|--all)
            # Default behavior
            shift
            ;;
        --s3)
            FOCUS="--focus=S3"
            shift
            ;;
        --iam)
            FOCUS="--focus=IAM"
            shift
            ;;
        --bucket)
            FOCUS="--focus=S3 Bucket"
            shift
            ;;
        --object)
            FOCUS="--focus=S3 Object"
            shift
            ;;
        --pusher)
            FOCUS="--focus=Pusher"
            shift
            ;;
        --user)
            FOCUS="--focus=IAM User"
            shift
            ;;
        --group)
            FOCUS="--focus=IAM Group"
            shift
            ;;
        --role)
            FOCUS="--focus=IAM Role"
            shift
            ;;
        --policy)
            FOCUS="--focus=IAM Policy"
            shift
            ;;
        --cors)
            FOCUS="--focus=CORS"
            shift
            ;;
        --tags)
            FOCUS="--focus=Tagging"
            shift
            ;;
        --stress)
            FOCUS="--focus=Stress"
            shift
            ;;
        --advanced)
            FOCUS="--focus=Advanced"
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

# Force MinIO mode if requested
if [ "$FORCE_MINIO" = true ]; then
    if [ -f "config.json" ]; then
        echo -e "${YELLOW}Renaming config.json to config.json.bak to force MinIO mode${NC}"
        mv config.json config.json.bak
        trap "mv config.json.bak config.json" EXIT
    fi
fi

# Check if minio binary exists
if [ ! -f "./minio" ]; then
    echo -e "${YELLOW}Warning: minio binary not found in current directory${NC}"
    echo -e "${YELLOW}Tests will run with MinIO mode but may fail without the binary${NC}"
    echo ""
fi

# Display test configuration
echo -e "${GREEN}=== AWS Test Runner ===${NC}"
echo ""
echo "Test directory: $SCRIPT_DIR"
echo "Verbose:        $([ -n "$VERBOSE" ] && echo "Yes" || echo "No")"
echo "Coverage:       $([ -n "$COVERAGE" ] && echo "Yes" || echo "No")"
echo "Focus:          $([ -n "$FOCUS" ] && echo "${FOCUS#--focus=}" || echo "All tests")"
echo "Skip:           $([ -n "$SKIP" ] && echo "${SKIP#--skip=}" || echo "None")"
echo ""

# Run tests with ginkgo if available, otherwise use go test
if command -v ginkgo &> /dev/null; then
    echo -e "${GREEN}Running tests with Ginkgo...${NC}"
    echo ""
    ginkgo $VERBOSE $COVERAGE $FOCUS $SKIP .
    TEST_RESULT=$?
else
    echo -e "${YELLOW}Ginkgo not found, using go test...${NC}"
    echo ""
    
    # Build go test command
    GO_TEST_CMD="go test"
    [ -n "$VERBOSE" ] && GO_TEST_CMD="$GO_TEST_CMD -v"
    [ -n "$COVERAGE" ] && GO_TEST_CMD="$GO_TEST_CMD -cover -coverprofile=coverage.out"
    [ -n "$FOCUS" ] && GO_TEST_CMD="$GO_TEST_CMD -run ${FOCUS#--focus=}"
    
    eval $GO_TEST_CMD
    TEST_RESULT=$?
fi

# Display coverage if generated
if [ -n "$COVERAGE" ] && [ -f "coverage.out" ]; then
    echo ""
    echo -e "${GREEN}=== Coverage Report ===${NC}"
    go tool cover -func=coverage.out | tail -1
    echo ""
    echo "To view detailed HTML coverage report, run:"
    echo "  go tool cover -html=coverage.out"
fi

# Exit with test result
if [ $TEST_RESULT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ All tests passed!${NC}"
else
    echo ""
    echo -e "${RED}✗ Some tests failed${NC}"
fi

exit $TEST_RESULT
