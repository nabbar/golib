/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package fileDescriptor_test

import (
	"fmt"
	"log"

	"github.com/nabbar/golib/ioutils/fileDescriptor"
)

// Example_queryLimits demonstrates the simplest usage: querying current file descriptor limits.
// This is useful for diagnostics and monitoring.
func Example_queryLimits() {
	// Query current limits without modification (newValue = 0)
	current, max, err := fileDescriptor.SystemFileDescriptor(0)
	if err != nil {
		log.Fatalf("Failed to query limits: %v", err)
	}

	fmt.Printf("Current limit: %d\n", current)
	fmt.Printf("Maximum limit: %d\n", max)
	fmt.Printf("Available headroom: %d\n", max-current)

	// Output format varies by system, so we don't specify exact output
}

// Example_simpleIncrease demonstrates increasing the file descriptor limit.
// This is commonly done during application startup.
func Example_simpleIncrease() {
	// Try to increase limit to 4096
	desired := 4096
	current, max, err := fileDescriptor.SystemFileDescriptor(desired)

	if err != nil {
		// Permission error: may need elevated privileges
		log.Printf("Cannot increase limit: %v", err)
		return
	}

	if current >= desired {
		fmt.Printf("Successfully set limit to %d (max: %d)\n", current, max)
	} else {
		fmt.Printf("Limit set to %d, below desired %d (max: %d)\n", current, desired, max)
	}
}

// Example_checkBeforeIncrease demonstrates checking if an increase is needed before attempting it.
// This avoids unnecessary syscalls and potential permission errors.
func Example_checkBeforeIncrease() {
	required := 2048

	// First, check current limit
	current, max, err := fileDescriptor.SystemFileDescriptor(0)
	if err != nil {
		log.Fatalf("Cannot query limits: %v", err)
	}

	// Check if increase is needed
	if current >= required {
		fmt.Printf("Current limit (%d) already meets requirement (%d)\n", current, required)
		return
	}

	// Check if increase is possible
	if max < required {
		log.Fatalf("System maximum (%d) is below requirement (%d)", max, required)
	}

	// Attempt increase
	newCurrent, newMax, err := fileDescriptor.SystemFileDescriptor(required)
	if err != nil {
		log.Fatalf("Cannot increase limit: %v (may need elevated privileges)", err)
	}

	fmt.Printf("Increased limit from %d to %d (max: %d)\n", current, newCurrent, newMax)
}

// Example_gracefulFallback demonstrates handling permission errors gracefully.
// Applications should work with available limits when increases fail.
func Example_gracefulFallback() {
	desired := 8192
	minimum := 1024

	// Query initial state
	initial, _, _ := fileDescriptor.SystemFileDescriptor(0)

	// Try to increase to desired
	current, _, err := fileDescriptor.SystemFileDescriptor(desired)
	if err != nil {
		log.Printf("Warning: Cannot increase to %d: %v", desired, err)
		current = initial
	}

	// Check if we meet minimum requirement
	if current < minimum {
		log.Fatalf("Insufficient file descriptors: have %d, need at least %d", current, minimum)
	}

	fmt.Printf("Proceeding with %d file descriptors\n", current)

	// Adjust application behavior based on available limit
	maxConnections := (current * 80) / 100 // Use 80% of limit
	fmt.Printf("Max concurrent connections: %d\n", maxConnections)
}

// Example_serverInitialization demonstrates a complete server initialization pattern.
// This shows how high-performance servers should handle file descriptor limits.
func Example_serverInitialization() {
	const (
		minRequired = 2048  // Absolute minimum for operation
		preferred   = 16384 // Optimal for high traffic
	)

	// Query initial state
	initial, max, err := fileDescriptor.SystemFileDescriptor(0)
	if err != nil {
		log.Fatalf("Cannot query file descriptor limits: %v", err)
	}

	fmt.Printf("Initial: current=%d, max=%d\n", initial, max)

	// Check minimum requirement
	if initial < minRequired {
		log.Printf("Current limit (%d) below minimum (%d), attempting increase...", initial, minRequired)

		current, _, err := fileDescriptor.SystemFileDescriptor(minRequired)
		if err != nil || current < minRequired {
			log.Fatalf("Cannot meet minimum requirement: have %d, need %d", current, minRequired)
		}

		fmt.Printf("Increased to minimum: %d\n", current)
	}

	// Try to reach preferred limit
	if initial < preferred && max >= preferred {
		current, newMax, err := fileDescriptor.SystemFileDescriptor(preferred)
		if err == nil && current >= preferred {
			fmt.Printf("Reached preferred limit: %d (max: %d)\n", current, newMax)
		} else {
			fmt.Printf("Using %d file descriptors (could not reach preferred %d)\n", initial, preferred)
		}
	} else {
		fmt.Printf("Using current limit: %d\n", initial)
	}

	// Calculate safe connection pool size
	current, _, _ := fileDescriptor.SystemFileDescriptor(0)
	safeConnections := (current * 75) / 100 // Reserve 25% for overhead
	fmt.Printf("Safe concurrent connections: %d\n", safeConnections)
}

// Example_databaseApplication demonstrates limit management for database applications.
// Shows how to calculate required limits based on connection pool sizes.
func Example_databaseApplication() {
	// Application configuration
	dbPools := 3    // Number of database pools
	poolSize := 50  // Connections per pool
	overhead := 200 // Reserved for logs, temp files, etc.
	required := dbPools*poolSize + overhead

	// Check and set limits
	current, max, err := fileDescriptor.SystemFileDescriptor(0)
	if err != nil {
		log.Fatalf("Cannot query limits: %v", err)
	}

	fmt.Printf("Application requires %d file descriptors\n", required)
	fmt.Printf("Current limit: %d, Maximum: %d\n", current, max)

	if current < required {
		if max < required {
			log.Fatalf("System maximum (%d) insufficient for requirements (%d)", max, required)
		}

		// Attempt increase
		newCurrent, _, err := fileDescriptor.SystemFileDescriptor(required)
		if err != nil {
			log.Fatalf("Cannot increase limit: %v", err)
		}

		if newCurrent < required {
			// Reduce pool sizes proportionally
			available := newCurrent - overhead
			if available < dbPools {
				log.Fatalf("Insufficient descriptors even for minimal pools")
			}

			newPoolSize := available / dbPools
			fmt.Printf("Reducing pool size from %d to %d per pool\n", poolSize, newPoolSize)
			poolSize = newPoolSize
		} else {
			fmt.Printf("Successfully increased limit to %d\n", newCurrent)
		}
	}

	fmt.Printf("Final configuration: %d pools × %d connections = %d total\n",
		dbPools, poolSize, dbPools*poolSize)
}

// Example_platformAware demonstrates writing platform-aware code.
// Shows how to adjust expectations based on the operating system.
func Example_platformAware() {
	// Platform-specific defaults
	var target int
	// On Windows, max is 8192 (C runtime limit)
	// On Unix/Linux, typical max is 65536 or higher
	// We adjust our target accordingly
	target = 16384 // Will be capped on Windows

	current, max, err := fileDescriptor.SystemFileDescriptor(target)
	if err != nil {
		log.Printf("Cannot set limit: %v", err)
		current, max, _ = fileDescriptor.SystemFileDescriptor(0)
	}

	fmt.Printf("Running with: current=%d, max=%d\n", current, max)

	// Adjust application behavior
	if max <= 8192 {
		fmt.Println("Platform: Windows or constrained environment")
		fmt.Println("Strategy: Use connection pooling and aggressive timeouts")
	} else {
		fmt.Println("Platform: Unix/Linux with higher limits available")
		fmt.Println("Strategy: Can support more concurrent connections")
	}
}

// Example_monitoring demonstrates periodic monitoring of file descriptor limits.
// Useful for long-running services that need to track resource availability.
func Example_monitoring() {
	// Initial state
	initial, _, _ := fileDescriptor.SystemFileDescriptor(0)
	fmt.Printf("Service started with %d file descriptors\n", initial)

	// In a real application, this would run periodically
	checkLimits := func() {
		current, max, err := fileDescriptor.SystemFileDescriptor(0)
		if err != nil {
			log.Printf("Warning: Cannot check limits: %v", err)
			return
		}

		// Calculate usage percentage (rough estimate)
		usage := float64(current) / float64(max) * 100

		fmt.Printf("FD Limits - Current: %d, Max: %d (%.1f%% of max)\n",
			current, max, usage)

		// Alert if approaching limit
		if usage >= 80.0 {
			log.Printf("WARNING: File descriptor limit usage high (%.1f%%)", usage)
			// In production: send alert, log to monitoring system
		}
	}

	// Simulate periodic check
	checkLimits()

	// Output format:
	// Service started with 1024 file descriptors
	// FD Limits - Current: 1024, Max: 4096 (25.0% of max)
}

// Example_errorHandling demonstrates comprehensive error handling patterns.
// Shows how to distinguish between different types of errors.
func Example_errorHandling() {
	desired := 100000

	current, max, err := fileDescriptor.SystemFileDescriptor(desired)

	if err != nil {
		// Error during syscall - could be permission denied or other system error
		log.Printf("ERROR: Cannot set limit to %d: %v", desired, err)
		log.Printf("Current: %d, Max: %d", current, max)

		// Check if it's a permission issue
		if current < desired && max >= desired {
			log.Println("Likely cause: insufficient privileges to increase limit")
			log.Println("Solution: run with elevated privileges or adjust system limits")
		} else if max < desired {
			log.Println("Likely cause: system hard limit is below requested value")
			log.Printf("Solution: increase system limit (current max: %d)", max)
		}

		// Fallback: use current limit
		fmt.Printf("Falling back to current limit: %d\n", current)
		return
	}

	// Success case
	if current >= desired {
		fmt.Printf("Successfully set limit to %d\n", current)
	} else {
		fmt.Printf("Partial success: limit set to %d (requested %d)\n", current, desired)
	}
}

// Example_multiStageIncrease demonstrates incrementally increasing limits.
// Useful when the optimal limit is unknown or system-dependent.
func Example_multiStageIncrease() {
	// Try progressively higher limits until one succeeds or we reach the target
	targets := []int{1024, 2048, 4096, 8192, 16384}

	initial, _, _ := fileDescriptor.SystemFileDescriptor(0)
	fmt.Printf("Starting with: %d\n", initial)

	var achieved int
	for _, target := range targets {
		if target <= initial {
			continue // Already above this target
		}

		current, _, err := fileDescriptor.SystemFileDescriptor(target)
		if err != nil {
			fmt.Printf("Cannot reach %d: %v\n", target, err)
			break
		}

		achieved = current
		fmt.Printf("Increased to: %d\n", current)

		if current < target {
			// Hit the limit
			break
		}
	}

	if achieved > initial {
		fmt.Printf("Final limit: %d (increased from %d)\n", achieved, initial)
	} else {
		fmt.Printf("Could not increase beyond initial: %d\n", initial)
	}
}
