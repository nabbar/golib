/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

/*
Package info provides a robust and thread-safe metadata management system for monitored components.
It acts as a dynamic repository for identifying information (Name) and descriptive key-value pairs (Data)
about a service, application, or internal component.

# Core Philosophy

The 'info' package is designed to bridge the gap between static configuration and dynamic runtime state.
It allows a component to have a permanent identity (Default Name) while also supporting manual overrides
and real-time data injection through provider functions.

# Architecture & Internal Logic

The internal state is managed by an 'inf' structure which utilizes a high-performance atomic map (libatm.Map)
to ensure non-blocking, concurrent access from multiple health check routines or API endpoints.

## Naming Resolution Hierarchy

When the Name() method is invoked, the package follows a strict priority-based resolution logic:

 1. Dynamic Provider: If a function is registered via RegisterName(), it is executed. If it returns
    a non-empty string, this result is used.
 2. Manual Override: If no dynamic result is available, the package checks for a name set manually
    via SetName().
 3. Default Fallback: If neither of the above exists, the default name provided at initialization
    (via New()) is returned.

Naming Dataflow Diagram:

	[ Call Name() ]
	      |
	      +--> [ Check Dynamic Provider (RegisterName) ] --(exists & non-empty)--> [ RETURN RESULT ]
	      |
	      +--> [ Check Manual Override (SetName) ] --------(exists)--------------> [ RETURN RESULT ]
	      |
	      +--> [ Default Name (New) ] -------------------------------------------> [ RETURN RESULT ]

## Metadata (Data) Resolution Logic

The Data() method aggregates information from two distinct layers:

 1. Static Store: Key-value pairs explicitly added via AddData() or SetData().
 2. Dynamic Provider: A map returned by the function registered via RegisterData().

The resolution logic merges these layers into a single map. In the event of a key collision,
the data from the Dynamic Provider takes precedence, ensuring that real-time runtime values
always override stale static configuration.

Metadata Dataflow Diagram:

	[ Call Data() ]
	      |
	      +--[ Retrieve all static keys from Atomic Map ]
	      |
	      +--[ Execute Dynamic Provider (RegisterData) ]
	      |
	      +--[ Merge Logic: Dynamic values OVERWRITE static values ]
	      |
	      +--> [ RETURN MERGED MAP ]

# Key Features

  - Thread-Safe: Atomic operations ensure consistency without the overhead of heavy mutexes.
  - Zero-Caching Policy: Provider functions are executed on every call, ensuring that data like
    CPU usage, memory stats, or current version is always fresh.
  - LIFO-like Cleanliness: SetData() replaces only user-defined metadata, preserving internal
    naming logic and registered functions.
  - Standard Interfaces: Implements encoding.TextMarshaler and json.Marshaler for seamless
    integration with logging frameworks and REST APIs.

# Usage Examples

## Initializing and Adding Static Data

	inf, _ := info.New("my-database-service")
	inf.AddData("environment", "production")
	inf.AddData("cluster", "eu-west-1")

## Registering Dynamic Metadata

This is useful for exposing real-time metrics alongside the component's identity.

	inf.RegisterData(func() (map[string]interface{}, error) {
	    return map[string]interface{}{
	        "goroutines": runtime.NumGoroutine(),
	        "uptime":     time.Since(startTime).String(),
	    }, nil
	})

## Manual Name Override

	inf.SetName("temporary-maintenance-mode")
	// Name() now returns "temporary-maintenance-mode" instead of "my-database-service".

# Thread Safety & Performance

The internal implementation uses a lock-free approach where possible, relying on atomic stores
and loads. This ensures that even under heavy load (e.g., thousands of status requests per second),
the impact on the monitored component's performance is negligible.

# Integration with Monitor

This package is a core dependency of the 'monitor' package. In a typical monitoring setup,
the 'Info' instance is passed to the monitor, which then uses it to enrich health check
reports and Prometheus metrics with the component's identity and metadata.
*/
package info
