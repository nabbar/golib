#!/bin/bash

##############################################################################
# Enhanced Coverage Report Script for golib
# 
# This script provides comprehensive test analysis across all Go packages:
# - Real per-package coverage (using go tool cover)
# - Detailed test metrics (specs, assertions, pending, skip, benchmarks)
# - Race detection support (with CGO_ENABLED=1)
# - Optional path targeting or full repository scan
#
# Usage: ./coverage-report.sh [options] [path]
#
# Arguments:
#   [path]            Optional path to test (default: current directory and below)
#
# Options:
#   -v, --verbose     Show detailed output during analysis
#   -h, --help        Show this help message
#   -o, --output FILE Save report to file
#   -m, --min PCT     Highlight packages below minimum coverage (default: 75)
#   -r, --race        Enable race detection (CGO_ENABLED=1, may take >10min)
#   --no-color        Disable colored output
#   -t, --timeout DUR Test timeout duration (default: 15m for normal, 30m for race)
##############################################################################

set -e

# Force C locale for consistent number formatting
export LC_NUMERIC=C

# Default color codes
_RED='\033[0;31m'
_GREEN='\033[0;32m'
_YELLOW='\033[1;33m'
_BLUE='\033[0;34m'
_CYAN='\033[0;36m'
_MAGENTA='\033[0;35m'
_NC='\033[0m' # No Color
_BOLD='\033[1m'

# Configuration
VERBOSE=0
OUTPUT_FILE=""
MIN_COVERAGE=75
RACE_MODE=0
USE_COLOR=1
TIMEOUT="15m"
TARGET_PATH=""
CURRENT_DIR="$(pwd)"

# Statistics
TOTAL_PACKAGES=0
TESTED_PACKAGES=0
UNTESTED_PACKAGES=0
TOTAL_SPECS=0
TOTAL_ASSERTIONS=0
TOTAL_PENDING=0
TOTAL_SKIPPED=0
TOTAL_BENCHMARKS=0
TOTAL_COVERAGE=0

# Arrays to store detailed results per package
declare -a PKG_NAMES
declare -a PKG_COVERAGE
declare -a PKG_SPECS
declare -a PKG_ASSERTIONS
declare -a PKG_PENDING
declare -a PKG_SKIPPED
declare -a PKG_BENCHMARKS
declare -a UNTESTED_LIST

##############################################################################
# Helper Functions
##############################################################################

# Initialize colors based on USE_COLOR flag
init_colors() {
    if [ "$USE_COLOR" -eq 0 ]; then
        RED=""
        GREEN=""
        YELLOW=""
        BLUE=""
        CYAN=""
        MAGENTA=""
        NC=""
        BOLD=""
    else
        RED="$_RED"
        GREEN="$_GREEN"
        YELLOW="$_YELLOW"
        BLUE="$_BLUE"
        CYAN="$_CYAN"
        MAGENTA="$_MAGENTA"
        NC="$_NC"
        BOLD="$_BOLD"
    fi
}

show_help() {
    grep '^#' "$0" | grep -v '#!/bin/bash' | sed 's/^# //g' | sed 's/^#//g'
    exit 0
}

log_verbose() {
    if [ "$VERBOSE" -eq 1 ]; then
        echo -e "${CYAN}[VERBOSE]${NC} $1"
    fi
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

##############################################################################
# Parse Arguments
##############################################################################

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=1
                shift
                ;;
            -h|--help)
                show_help
                ;;
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            -m|--min)
                MIN_COVERAGE="$2"
                shift 2
                ;;
            -r|--race)
                RACE_MODE=1
                # Increase default timeout for race mode
                if [ "$TIMEOUT" = "10m" ]; then
                    TIMEOUT="30m"
                fi
                shift
                ;;
            --no-color)
                USE_COLOR=0
                shift
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -*)
                log_error "Unknown option: $1"
                show_help
                ;;
            *)
                # First non-option argument is the target path
                if [ -z "$TARGET_PATH" ]; then
                    TARGET_PATH="$1"
                else
                    log_error "Multiple paths specified. Only one path is allowed."
                    exit 1
                fi
                shift
                ;;
        esac
    done
    
    # Initialize colors after parsing arguments
    init_colors
    
    # Validate go.mod exists in current directory
    if [ ! -f "${CURRENT_DIR}/go.mod" ]; then
        log_error "No go.mod found in current directory: $CURRENT_DIR"
        log_error "Please run this script from the root of a Go module"
        exit 1
    fi
    
    # Set default target path if not specified (current directory)
    if [ -z "$TARGET_PATH" ]; then
        TARGET_PATH="$CURRENT_DIR"
    else
        # Target path is relative to current directory
        if [[ "$TARGET_PATH" = /* ]]; then
            # Absolute path provided
            if [[ ! "$TARGET_PATH" == "$CURRENT_DIR"* ]]; then
                log_error "Target path must be within current directory: $CURRENT_DIR"
                exit 1
            fi
        else
            # Relative path - make it absolute
            TARGET_PATH="${CURRENT_DIR}/${TARGET_PATH}"
        fi
        
        # Verify the target path exists
        if [ ! -d "$TARGET_PATH" ]; then
            log_error "Target path does not exist: $TARGET_PATH"
            exit 1
        fi
    fi
}

##############################################################################
# Package Discovery
##############################################################################

find_packages() {
    log_info "Scanning for Go packages in: $TARGET_PATH" >&2
    
    # Find all directories containing non-test .go files
    # This ensures we only check packages with actual source code
    local packages=$(find "$TARGET_PATH" -type f -name "*.go" \
        ! -name "*_test.go" \
        -not -path "*/vendor/*" \
        -not -path "*/.*" \
        -not -path "*/_*" \
        -exec dirname {} \; | sort -u)
    
    if [ -z "$packages" ]; then
        log_error "No Go packages found in $TARGET_PATH" >&2
        exit 1
    fi
    
    local pkg_count=$(echo "$packages" | wc -l)
    log_verbose "Found $pkg_count packages with source files" >&2
    
    echo "$packages"
}

##############################################################################
# Test Metrics Extraction
##############################################################################

count_test_metrics() {
    local pkg_dir="$1"
    local specs=0
    local assertions=0
    local pending=0
    local skipped=0
    local benchmarks=0
    
    if ! ls "$pkg_dir"/*_test.go >/dev/null 2>&1; then
        echo "0 0 0 0 0"
        return
    fi
    
    # Count Ginkgo/Gomega specifications
    # It(), Specify(), Entry() from table-driven tests
    specs=$(grep -rh '\(^\s*It(\|^\s*Specify(\|^\s*Entry(\)' "$pkg_dir"/*_test.go 2>/dev/null | wc -l || echo "0")
    
    # If no Ginkgo specs, count regular Test functions
    if [ "$specs" -eq 0 ]; then
        specs=$(grep -rh '^func Test' "$pkg_dir"/*_test.go 2>/dev/null | wc -l || echo "0")
    fi
    
    # Count assertions: Expect(), Eventually(), Consistently()
    assertions=$(grep -rh '\(Expect(\|Eventually(\|Consistently(\)' "$pkg_dir"/*_test.go 2>/dev/null | wc -l || echo "0")
    
    # Count pending specs: PIt(), PSpecify(), or XIt(), XSpecify()
    pending=$(grep -rh '\(^\s*PIt(\|^\s*PSpecify(\|^\s*XIt(\|^\s*XSpecify(\)' "$pkg_dir"/*_test.go 2>/dev/null | wc -l || echo "0")
    
    # Count skipped: Skip(), t.Skip()
    skipped=$(grep -rh '\(\.Skip(\|t\.Skip(\)' "$pkg_dir"/*_test.go 2>/dev/null | wc -l || echo "0")
    
    # Count benchmarks: Benchmark functions and gmeasure
    benchmarks=$(grep -rh '\(^func Benchmark\|Measure(\)' "$pkg_dir"/*_test.go 2>/dev/null | wc -l || echo "0")
    
    echo "$specs $assertions $pending $skipped $benchmarks"
}

##############################################################################
# Coverage Analysis
##############################################################################

run_tests_with_coverage() {
    local test_path="$1"
    local coverprofile="$2"
    
    log_info "Running test suite..."
    if [ "$RACE_MODE" -eq 1 ]; then
        log_warning "Race detection enabled - this may take more than 10 minutes"
        log_info "Command: CGO_ENABLED=1 go test -race -timeout=$TIMEOUT -coverprofile=$coverprofile -covermode=atomic ./..."
    else
        log_info "Command: go test -timeout=$TIMEOUT -coverprofile=$coverprofile -covermode=atomic ./..."
    fi
    echo ""
    
    cd "$test_path"
    
    local test_cmd
    if [ "$RACE_MODE" -eq 1 ]; then
        export CGO_ENABLED=1
        test_cmd="go test -race -timeout=$TIMEOUT -coverprofile=$coverprofile -covermode=atomic ./..."
    else
        test_cmd="go test -timeout=$TIMEOUT -coverprofile=$coverprofile -covermode=atomic ./..."
    fi
    
    # Run tests and capture output
    local test_output
    local test_exit_code=0
    
    if [ "$VERBOSE" -eq 1 ]; then
        # Show full output in verbose mode
        if ! $test_cmd 2>&1; then
            test_exit_code=$?
        fi
    else
        # Show only package results in normal mode
        if ! $test_cmd 2>&1 | grep -E "(ok|FAIL|\\?|coverage:)"; then
            test_exit_code=$?
        fi
    fi
    
    echo ""
    
    if [ $test_exit_code -eq 0 ]; then
        if [ -f "$coverprofile" ]; then
            log_success "Test suite completed successfully"
        else
            log_warning "Tests completed but no coverage profile generated"
        fi
    else
        log_warning "Some tests failed or had issues (exit code: $test_exit_code)"
        exit $test_exit_code
    fi
    
    return 0
}

extract_package_coverage() {
    local pkg_dir="$1"
    local pkg_import_path="$2"
    local coverprofile="$3"
    
    # Extract all coverage lines for this package from the master profile
    # This includes coverage from tests in other packages that exercise this package's code
    local tmp_profile="${coverprofile}.tmp.$$"
    echo "mode: atomic" > "$tmp_profile"
    
    # Get coverage for this specific package
    grep "^${pkg_import_path}/" "$coverprofile" 2>/dev/null >> "$tmp_profile" || true
    
    # Calculate coverage percentage using go tool cover
    local coverage_pct=""
    if [ -s "$tmp_profile" ] && [ "$(wc -l < "$tmp_profile")" -gt 1 ]; then
        coverage_pct=$(go tool cover -func="$tmp_profile" 2>/dev/null | grep "total:" | awk '{print $3}' | sed 's/%//' || echo "")
    fi
    
    rm -f "$tmp_profile"
    
    if [ -z "$coverage_pct" ]; then
        echo "0.0"
    else
        echo "$coverage_pct"
    fi
}

##############################################################################
# Analysis Engine
##############################################################################

analyze_packages() {
    local packages="$1"
    local total=$(echo "$packages" | wc -l)
    
    log_info "Analyzing $total packages..."
    echo ""
    
    # Generate coverage profile for all packages
    local coverprofile="${TARGET_PATH}/.coverage-full-$$.out"
    run_tests_with_coverage "$TARGET_PATH" "$coverprofile"
    
    # Check if coverage profile exists
    if [ ! -f "$coverprofile" ]; then
        log_warning "Coverage profile not generated - some metrics may be unavailable"
    fi
    
    # Get module path from go.mod in current directory
    local module_path=$(grep "^module " "${CURRENT_DIR}/go.mod" | awk '{print $2}')
    local module_root="$CURRENT_DIR"
    
    log_verbose "Module root: $module_root"
    log_verbose "Module path: $module_path"
    
    # Process each package
    log_info "Extracting metrics for each package..."
    echo ""
    
    local count=0
    while IFS= read -r pkg_dir; do
        count=$((count + 1))
        
        # Get relative package path (relative to TARGET_PATH for display)
        local pkg_path="${pkg_dir#${TARGET_PATH}/}"
        if [ "$pkg_path" = "$pkg_dir" ]; then
            pkg_path="."
        fi
        
        # Construct full import path (relative to module root for coverage extraction)
        local pkg_rel_to_module="${pkg_dir#${module_root}/}"
        local pkg_import_path
        
        if [ "$pkg_rel_to_module" = "$pkg_dir" ]; then
            # Package is at module root
            pkg_import_path="$module_path"
        else
            # Package is in subdirectory
            pkg_import_path="${module_path}/${pkg_rel_to_module}"
        fi
        
        # Show progress
        printf "\r${CYAN}Progress:${NC} [%d/%d] Processing: %-50s" "$count" "$total" "$pkg_path"
        
        # Check if package has test files
        local has_tests=0
        if ls "$pkg_dir"/*_test.go >/dev/null 2>&1; then
            has_tests=1
        fi
        
        # Count test metrics
        read -r specs assertions pending skipped benchmarks <<< $(count_test_metrics "$pkg_dir")
        
        # Extract real coverage for this package (includes coverage from other packages' tests)
        local coverage_pct="0.0"
        if [ -f "$coverprofile" ] && [ "$has_tests" -eq 1 ]; then
            coverage_pct=$(extract_package_coverage "$pkg_dir" "$pkg_import_path" "$coverprofile")
        fi
        
        TOTAL_PACKAGES=$((TOTAL_PACKAGES + 1))
        
        if [ "$has_tests" -eq 0 ]; then
            UNTESTED_PACKAGES=$((UNTESTED_PACKAGES + 1))
            UNTESTED_LIST+=("$pkg_path")
        else
            # Package has tests
            TESTED_PACKAGES=$((TESTED_PACKAGES + 1))
            PKG_NAMES+=("$pkg_path")
            PKG_COVERAGE+=("$coverage_pct")
            PKG_SPECS+=("$specs")
            PKG_ASSERTIONS+=("$assertions")
            PKG_PENDING+=("$pending")
            PKG_SKIPPED+=("$skipped")
            PKG_BENCHMARKS+=("$benchmarks")
            
            # Update totals
            TOTAL_COVERAGE=$(echo "$TOTAL_COVERAGE + $coverage_pct" | bc)
            TOTAL_SPECS=$((TOTAL_SPECS + specs))
            TOTAL_ASSERTIONS=$((TOTAL_ASSERTIONS + assertions))
            TOTAL_PENDING=$((TOTAL_PENDING + pending))
            TOTAL_SKIPPED=$((TOTAL_SKIPPED + skipped))
            TOTAL_BENCHMARKS=$((TOTAL_BENCHMARKS + benchmarks))
        fi
        
    done <<< "$packages"
    
    # Clean up coverage file
    rm -f "$coverprofile"
    rm -f "${coverprofile}".tmp.*
    
    echo "" # New line after progress
    echo ""
}

##############################################################################
# Report Generation
##############################################################################

print_report() {
    local avg_coverage=0
    if [ "$TESTED_PACKAGES" -gt 0 ]; then
        avg_coverage=$(echo "scale=2; $TOTAL_COVERAGE / $TESTED_PACKAGES" | bc)
    fi
    
    # Header
    echo -e "${BOLD}═══════════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}                      GOLIB TEST COVERAGE REPORT${NC}"
    echo -e "${BOLD}═══════════════════════════════════════════════════════════════════════════${NC}"
    echo ""
    
    # Configuration Info
    echo -e "${BOLD}CONFIGURATION${NC}"
    echo "───────────────────────────────────────────────────────────────────────────"
    printf "%-30s: %s\n" "Target Path" "$TARGET_PATH"
    printf "%-30s: %s\n" "Timeout" "$TIMEOUT"
    printf "%-30s: %s\n" "Race Detection" "$([ "$RACE_MODE" -eq 1 ] && echo "${GREEN}Enabled${NC}" || echo "Disabled")"
    printf "%-30s: %s\n" "Minimum Coverage Threshold" "${MIN_COVERAGE}%"
    echo ""
    
    # Summary
    echo -e "${BOLD}SUMMARY${NC}"
    echo "───────────────────────────────────────────────────────────────────────────"
    printf "%-30s: %s\n" "Total Packages" "$TOTAL_PACKAGES"
    printf "%-30s: ${GREEN}%s${NC}\n" "Packages with Tests" "$TESTED_PACKAGES"
    printf "%-30s: ${YELLOW}%s${NC}\n" "Packages without Tests" "$UNTESTED_PACKAGES"
    echo ""
    printf "%-30s: %s\n" "Total Specifications" "$TOTAL_SPECS"
    printf "%-30s: %s\n" "Total Assertions" "$TOTAL_ASSERTIONS"
    printf "%-30s: ${YELLOW}%s${NC}\n" "Total Pending" "$TOTAL_PENDING"
    printf "%-30s: ${CYAN}%s${NC}\n" "Total Skipped" "$TOTAL_SKIPPED"
    printf "%-30s: ${MAGENTA}%s${NC}\n" "Total Benchmarks" "$TOTAL_BENCHMARKS"
    echo ""
    printf "%-30s: ${BOLD}${GREEN}%.2f%%${NC}\n" "Average Coverage" "$avg_coverage"
    echo ""
    
    # Detailed coverage by package
    if [ "$TESTED_PACKAGES" -gt 0 ]; then
        echo -e "${BOLD}DETAILED METRICS BY PACKAGE${NC}"
        echo "───────────────────────────────────────────────────────────────────────────"
        printf "%-40s %9s %7s %7s %5s %5s %7s\n" \
            "PACKAGE" "COVERAGE" "SPECS" "ASSERT" "PEND" "SKIP" "BENCH"
        echo "───────────────────────────────────────────────────────────────────────────"
        
        for i in "${!PKG_NAMES[@]}"; do
            local pkg="${PKG_NAMES[$i]}"
            local cov="${PKG_COVERAGE[$i]}"
            local specs="${PKG_SPECS[$i]}"
            local asserts="${PKG_ASSERTIONS[$i]}"
            local pend="${PKG_PENDING[$i]}"
            local skip="${PKG_SKIPPED[$i]}"
            local bench="${PKG_BENCHMARKS[$i]}"
            
            # Color code based on coverage
            local color="$GREEN"
            if (( $(echo "$cov < $MIN_COVERAGE" | bc -l) )); then
                color="$YELLOW"
            fi
            if (( $(echo "$cov < 50" | bc -l) )); then
                color="$RED"
            fi
            
            # Format package name (truncate if too long)
            local pkg_display="$pkg"
            if [ ${#pkg_display} -gt 40 ]; then
                pkg_display="...${pkg_display: -37}"
            fi
            
            printf "%-40s ${color}%8.2f%%${NC} %7s %7s %5s %5s %7s\n" \
                "$pkg_display" "$cov" "$specs" "$asserts" "$pend" "$skip" "$bench"
        done
        echo ""
    fi
    
    # Packages without tests
    if [ "$UNTESTED_PACKAGES" -gt 0 ]; then
        echo -e "${BOLD}PACKAGES WITHOUT TESTS${NC}"
        echo "───────────────────────────────────────────────────────────────────────────"
        for pkg in "${UNTESTED_LIST[@]}"; do
            echo -e "${YELLOW}•${NC} $pkg"
        done
        echo ""
    fi
    
    # Packages below minimum coverage
    local low_coverage_count=0
    for i in "${!PKG_NAMES[@]}"; do
        local cov="${PKG_COVERAGE[$i]}"
        if (( $(echo "$cov < $MIN_COVERAGE" | bc -l) )); then
            low_coverage_count=$((low_coverage_count + 1))
        fi
    done
    
    if [ "$low_coverage_count" -gt 0 ]; then
        echo -e "${BOLD}PACKAGES BELOW ${MIN_COVERAGE}% COVERAGE${NC}"
        echo "───────────────────────────────────────────────────────────────────────────"
        for i in "${!PKG_NAMES[@]}"; do
            local pkg="${PKG_NAMES[$i]}"
            local cov="${PKG_COVERAGE[$i]}"
            if (( $(echo "$cov < $MIN_COVERAGE" | bc -l) )); then
                printf "${YELLOW}•${NC} %-60s %8.2f%%\n" "$pkg" "$cov"
            fi
        done
        echo ""
    fi
    
    # Footer
    echo -e "${BOLD}═══════════════════════════════════════════════════════════════════════════${NC}"
    
    # Recommendations
    if [ "$UNTESTED_PACKAGES" -gt 0 ] || [ "$low_coverage_count" -gt 0 ] || [ "$TOTAL_PENDING" -gt 0 ]; then
        echo ""
        echo -e "${YELLOW}RECOMMENDATIONS:${NC}"
        if [ "$UNTESTED_PACKAGES" -gt 0 ]; then
            echo "  • Add test coverage for $UNTESTED_PACKAGES untested packages"
        fi
        if [ "$low_coverage_count" -gt 0 ]; then
            echo "  • Improve coverage for $low_coverage_count packages below ${MIN_COVERAGE}%"
        fi
        if [ "$TOTAL_PENDING" -gt 0 ]; then
            echo "  • Complete or remove $TOTAL_PENDING pending test specifications"
        fi
        echo ""
    fi
}

save_report() {
    if [ -n "$OUTPUT_FILE" ]; then
        log_info "Saving report to $OUTPUT_FILE..."
        # Disable colors for file output
        local original_use_color=$USE_COLOR
        USE_COLOR=0
        init_colors
        {
            print_report
        } > "$OUTPUT_FILE" 2>&1
        USE_COLOR=$original_use_color
        init_colors
        log_success "Report saved to $OUTPUT_FILE"
    fi
}

##############################################################################
# Main Execution
##############################################################################

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Show startup info
    log_info "Starting enhanced coverage analysis..."
    log_info "Working directory: $CURRENT_DIR"
    
    # Show target if different from current directory
    if [ "$TARGET_PATH" != "$CURRENT_DIR" ]; then
        local target_rel="${TARGET_PATH#${CURRENT_DIR}/}"
        log_info "Target path: $target_rel"
    else
        log_info "Target: All packages"
    fi
    
    if [ "$RACE_MODE" -eq 1 ]; then
        log_warning "Race detection mode enabled (CGO_ENABLED=1)"
        log_warning "Tests may take more than 10 minutes to complete"
    fi
    echo ""
    
    # Find all packages
    packages=$(find_packages)
    
    # Analyze each package
    analyze_packages "$packages"
    
    # Print report
    print_report
    
    # Save to file if requested
    save_report
    
    log_success "Coverage analysis complete!"
}

# Run main function with all arguments
main "$@"
