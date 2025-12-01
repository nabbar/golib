/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package gorm_test

import (
	"fmt"
	"time"

	liblog "github.com/nabbar/golib/logger"
	loggorm "github.com/nabbar/golib/logger/gorm"
	gorlog "gorm.io/gorm/logger"
)

// Example_basic demonstrates the simplest usage of the GORM logger adapter.
// This is suitable for basic applications where default behavior is acceptable.
func Example_basic() {
	// Create a mock logger (in real usage, use a configured golib logger)
	mockLogger := NewMockLogger()

	// Create GORM logger adapter with basic configuration
	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		false,                // Do not ignore record not found errors
		100*time.Millisecond, // Slow query threshold: 100ms
	)

	// gormLogger can now be used with GORM
	// db, err := gorm.Open(driver, &gorm.Config{Logger: gormLogger})
	_ = gormLogger // Use the logger
	fmt.Println("GORM logger created successfully")

	// Output:
	// GORM logger created successfully
}

// Example_withSlowQueryDetection shows how to configure slow query monitoring.
// Queries exceeding the threshold will be logged as warnings.
func Example_withSlowQueryDetection() {
	mockLogger := NewMockLogger()

	// Configure with 50ms threshold for performance-sensitive applications
	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		false,
		50*time.Millisecond, // Low threshold for strict performance monitoring
	)

	_ = gormLogger // Use the logger
	fmt.Println("Slow query threshold: 50ms")

	// Any query taking longer than 50ms will be logged as WarnLevel
	// with "SLOW Query >= 50ms" error message

	// Output:
	// Slow query threshold: 50ms
}

// Example_ignoreRecordNotFound demonstrates filtering of ErrRecordNotFound.
// Useful when "not found" is expected behavior rather than an error.
func Example_ignoreRecordNotFound() {
	mockLogger := NewMockLogger()

	// Enable ignoring of record not found errors
	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		true, // Ignore ErrRecordNotFound
		100*time.Millisecond,
	)

	// With this configuration:
	// - ErrRecordNotFound is logged as InfoLevel
	// - Other errors are still logged as ErrorLevel
	_ = gormLogger // Use the logger
	fmt.Println("ErrRecordNotFound will be logged as info")

	// Output:
	// ErrRecordNotFound will be logged as info
}

// Example_disableSlowQueryDetection shows how to disable slow query warnings.
// Set threshold to 0 to log all queries at InfoLevel regardless of duration.
func Example_disableSlowQueryDetection() {
	mockLogger := NewMockLogger()

	// Disable slow query detection by setting threshold to 0
	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		false,
		0, // Disabled: no slow query warnings
	)

	_ = gormLogger // Use the logger
	fmt.Println("Slow query detection disabled")

	// All queries will be logged at InfoLevel, never WarnLevel for slowness

	// Output:
	// Slow query detection disabled
}

// Example_dynamicLogLevel demonstrates changing log levels at runtime.
// Useful for adjusting verbosity without restarting the application.
func Example_dynamicLogLevel() {
	mockLogger := NewMockLogger()

	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		false,
		100*time.Millisecond,
	)

	// Set to Silent to disable all logging
	gormLogger.LogMode(gorlog.Silent)
	fmt.Println("Level: Silent")

	// Set to Info for detailed query logging
	gormLogger.LogMode(gorlog.Info)
	fmt.Println("Level: Info")

	// Set to Warn for warnings and errors only
	gormLogger.LogMode(gorlog.Warn)
	fmt.Println("Level: Warn")

	// Set to Error for errors only
	gormLogger.LogMode(gorlog.Error)
	fmt.Println("Level: Error")

	// Output:
	// Level: Silent
	// Level: Info
	// Level: Warn
	// Level: Error
}

// Example_productionConfiguration shows a recommended setup for production.
// Balances performance monitoring with noise reduction.
func Example_productionConfiguration() {
	mockLogger := NewMockLogger()

	// Production-optimized configuration
	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		true,                 // Ignore not found errors (common in production)
		200*time.Millisecond, // 200ms threshold (reasonable for most apps)
	)

	// Start with Warn level in production
	gormLogger.LogMode(gorlog.Warn)

	fmt.Println("Production configuration applied")
	fmt.Println("- ErrRecordNotFound ignored")
	fmt.Println("- Slow threshold: 200ms")
	fmt.Println("- Level: Warn")

	// Output:
	// Production configuration applied
	// - ErrRecordNotFound ignored
	// - Slow threshold: 200ms
	// - Level: Warn
}

// Example_developmentConfiguration shows a recommended setup for development.
// Maximizes visibility for debugging and optimization.
func Example_developmentConfiguration() {
	mockLogger := NewMockLogger()

	// Development-optimized configuration
	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		false, // Log all errors including not found
		0,     // Disable slow query detection (see all queries)
	)

	// Set to Info level to see all query details
	gormLogger.LogMode(gorlog.Info)

	fmt.Println("Development configuration applied")
	fmt.Println("- All errors logged")
	fmt.Println("- Slow threshold: disabled")
	fmt.Println("- Level: Info")

	// Output:
	// Development configuration applied
	// - All errors logged
	// - Slow threshold: disabled
	// - Level: Info
}

// Example_highPerformanceMonitoring demonstrates aggressive performance tracking.
// Suitable for performance-critical applications requiring tight SLAs.
func Example_highPerformanceMonitoring() {
	mockLogger := NewMockLogger()

	// Aggressive performance monitoring
	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		true,
		30*time.Millisecond, // Very strict: 30ms threshold
	)

	gormLogger.LogMode(gorlog.Warn)

	fmt.Println("High-performance monitoring enabled")
	fmt.Println("- Threshold: 30ms (strict)")
	fmt.Println("- Only warnings and errors logged")

	// This will catch even slightly slow queries
	// Useful for identifying optimization opportunities

	// Output:
	// High-performance monitoring enabled
	// - Threshold: 30ms (strict)
	// - Only warnings and errors logged
}

// Example_analyticsConfiguration shows setup for analytical workloads.
// Long-running queries are expected and should not trigger warnings.
func Example_analyticsConfiguration() {
	mockLogger := NewMockLogger()

	// Analytics-optimized configuration
	gormLogger := loggorm.New(
		func() liblog.Logger { return mockLogger },
		true,
		5*time.Second, // High threshold for complex analytical queries
	)

	gormLogger.LogMode(gorlog.Warn)

	fmt.Println("Analytics configuration applied")
	fmt.Println("- Threshold: 5s (permissive)")
	fmt.Println("- Suitable for OLAP queries")

	// Complex aggregations and joins won't trigger slow query warnings
	// unless they exceed 5 seconds

	// Output:
	// Analytics configuration applied
	// - Threshold: 5s (permissive)
	// - Suitable for OLAP queries
}

// Example_withLoggerFactory demonstrates using a factory function for dynamic logger switching.
// This allows changing the underlying logger without recreating the GORM logger.
func Example_withLoggerFactory() {
	// Shared logger reference (can be replaced at runtime)
	var currentLogger liblog.Logger = NewMockLogger()

	// Factory function returns the current logger
	loggerFactory := func() liblog.Logger {
		return currentLogger
	}

	gormLogger := loggorm.New(
		loggerFactory,
		false,
		100*time.Millisecond,
	)

	fmt.Println("Logger factory configured")

	// Later, you can replace the logger:
	// currentLogger = newConfiguredLogger()
	// All subsequent GORM logs will use the new logger

	_ = gormLogger // Use the logger

	// Output:
	// Logger factory configured
}
