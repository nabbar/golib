#!/bin/bash
#
# MIT License
#
# Copyright (c) 2026 Nicolas JUHEL
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
#

#set -e

# Default package to current directory if not provided
PKG="${1:-.}"
TIMEOUT="2m"

# Determine Output Directory
# Strip common suffixes to find the directory part
CLEAN_PKG="$PKG"
CLEAN_PKG="${CLEAN_PKG%...}"
CLEAN_PKG="${CLEAN_PKG%/}"

if [ -d "$CLEAN_PKG" ]; then
    LOG_DIR="$CLEAN_PKG"
else
    # If the package path isn't a directory (e.g. a go module path), default to current
    LOG_DIR="."
fi

echo "Running tests for package: $PKG"
echo "Logs and metrics will be stored in: $LOG_DIR"

# Define output files
F_COV="$LOG_DIR/res_coverage.log"
F_COV_RACE="$LOG_DIR/res_coverage_race.log"
F_LOG_TEST="$LOG_DIR/res_test.log"
F_LOG_RACE="$LOG_DIR/res_test_race.log"
F_LOG_BENCH="$LOG_DIR/res_bench.log"
F_CPULST="$LOG_DIR/res_cpu-list.log"
F_CPUSVG="$LOG_DIR/res_cpu.svg"
F_MEMLST="$LOG_DIR/res_mem-list.log"
F_MEMSVG="$LOG_DIR/res_mem.svg"
F_MEMTOP="$LOG_DIR/res_mem-top.log"
F_REPORT="$LOG_DIR/res_report.log"
F_LSHW="$LOG_DIR/res_lshw.log"
F_LOG_SEC="$LOG_DIR/res_gosec.log"
F_LOG_LINT="$LOG_DIR/res_golint.log"

# Clean up previous artifacts
rm -f "$F_COV" "$F_COV.out" "$F_COV_RACE" "$F_COV_RACE.out" "$F_LOG_TEST" "$F_LOG_RACE"
rm -f "$F_LOG_BENCH" "$F_CPULST" "$F_CPULST.out" "$F_CPUSVG" "$F_MEMLST" "$F_MEMLST.out" "$F_MEMSVG" "$F_MEMTOP"
rm -f "$F_REPORT" "$F_LSHW" "$F_LOG_SEC" "$F_LOG_LINT"

# Testing sudo
sudo echo "ok"

# 1. Print Material used for test
# Capture both script messages and command output to file
{
    echo "----------------------------------------------------------------------"
    echo "Listing materials..."
    echo "----------------------------------------------------------------------"
    sudo lshw
} > "$F_LSHW" 2>&1

echo "Step 1/7: Listing materials. Logs: $F_LSHW"

# 2. Calling Reports script
# Capture both script messages and command output to file
$(dirname "${0}")/coverage-report.sh -o "$(basename "$F_REPORT")" "${CLEAN_PKG#./}"
echo "Step 2/7: Report called. Logs: $F_REPORT"

# 3. Normal Test Mode with Coverage
# Capture both script messages and command output to file
{
    echo "----------------------------------------------------------------------"
    echo "Running Tests (Normal Mode) with Coverage..."
    echo "Package: $PKG"
    echo "Timeout: $TIMEOUT"
    echo "Mode: atomic"
    echo "----------------------------------------------------------------------"
    go test -v -timeout "$TIMEOUT" -covermode=atomic -coverprofile="$F_COV.out" "$PKG"
    go tool cover -func="$F_COV.out" -o="$F_COV"
    rm -f "$F_COV.out"
} > "$F_LOG_TEST" 2>&1

echo "Step 3/7: Normal Tests completed. Logs: $F_LOG_TEST"

# 4. Benchmarks (Normal Mode)
{
    echo "----------------------------------------------------------------------"
    echo "Running Benchmarks..."
    echo "Package: $PKG"
    echo "Flags: -bench=. -benchmem"
    echo "----------------------------------------------------------------------"
    go test -v -run=^$ -bench=. -benchmem -cpuprofile="$F_CPULST.out" -memprofile="$F_MEMLST.out" "$PKG"
    go tool pprof -svg "$F_CPULST.out" > "$F_CPUSVG"
    go tool pprof -list . "$F_CPULST.out" > "$F_CPULST"
    go tool pprof -svg "$F_MEMLST.out" > "$F_MEMSVG"
    go tool pprof -list . "$F_MEMLST.out" > "$F_MEMLST"
    go tool pprof -top . "$F_MEMLST.out" > "$F_MEMTOP"
    rm -f "$F_CPULST.out" "$F_MEMLST.out"
} > "$F_LOG_BENCH" 2>&1

echo "Step 4/7: Benchmarks completed. Logs: $F_LOG_BENCH"

# 5. Race Test Mode with Coverage
{
    echo "----------------------------------------------------------------------"
    echo "Running Tests (Race Mode) with Coverage..."
    echo "Package: $PKG"
    echo "Timeout: $TIMEOUT"
    echo "Mode: atomic + race"
    echo "----------------------------------------------------------------------"
    export CGO_ENABLED=1
    go test -v -race -timeout "$TIMEOUT" -covermode=atomic -coverprofile="$F_COV_RACE.out" "$PKG"
    export CGO_ENABLED=0
    go tool cover -func="$F_COV_RACE.out" -o="$F_COV_RACE"
    rm -f "$F_COV_RACE.out"
} > "$F_LOG_RACE" 2>&1

echo "Step 5/7: Race Tests completed. Logs: $F_LOG_RACE"

# 6. Checking security static code
{
    echo "----------------------------------------------------------------------"
    echo "Checking static security ..."
    echo "Package: $PKG"
    echo "----------------------------------------------------------------------"
    gosec -sort "$PKG"
} > "$F_LOG_SEC" 2>&1

echo "Step 6/7: Checking static security completed. Logs: $F_LOG_SEC"

# 7. Verify Golint
{
    echo "----------------------------------------------------------------------"
    echo "Checking / Updating format & imports..."
    echo "Package: $PKG"
    echo "----------------------------------------------------------------------"
    for ITM in $(find "$LOG_DIR" -type f -name '*.go' | grep -v '/vendor/')
    do
      gofmt -w "$ITM"
      go fmt "$ITM"
      goimports -w "$ITM"
    done
    echo "----------------------------------------------------------------------"
    echo "Checking linters..."
    echo "Package: $PKG"
    echo "----------------------------------------------------------------------"
    golangci-lint --config .golangci.yml run "$PKG"
} > "$F_LOG_LINT" 2>&1

echo "Step 7/7: Checking format, imports & linter completed. Logs: $F_LOG_LINT"

echo "----------------------------------------------------------------------"
echo "All operations completed successfully."
echo "Artifacts in $LOG_DIR:"
echo " - Logs: test.log, test_race.log, bench.log"
echo " - Coverage: coverage.out, coverage_race.out"
echo " - Profiles: cpu.out, mem.out"
echo " - Quality: gosec.log golint.log"
echo " - Reports: lshw.log, report.log"
echo "----------------------------------------------------------------------"
