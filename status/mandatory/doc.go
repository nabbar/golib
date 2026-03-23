/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

// Package mandatory provides a thread-safe, high-performance mechanism for managing
// groups of components that share a common validation mode and descriptive metadata
// within the status monitoring ecosystem.
//
// This package is a fundamental building block for the status monitoring system.
// It allows developers to define a logical set of components (e.g., "all critical databases",
// "all optional caches") and associate them with a single control mode (e.g., "Must",
// "Should", "AnyOf") and rich metadata (e.g., description, runbook links). This
// abstraction simplifies the configuration and evaluation of complex health check policies.
//
// # Key Features
//
//   - **Thread-Safety**: All operations are designed to be safe for concurrent use.
//     The implementation leverages atomic operations to ensure non-blocking, lock-free reads,
//     which is critical for high-throughput health check endpoints.
//   - **Dynamic Management**: Component keys can be added or removed from a group at runtime,
//     allowing the system to adapt to changes in the application's topology.
//   - **Rich Metadata**: Each group can be annotated with a map of `Info` containing
//     details like a human-readable description or links to dashboards and runbooks,
//     which are then exposed in the main status response.
//   - **Control Mode Association**: Each group is tightly coupled with a `control.Mode`,
//     ensuring that all components in the group are evaluated consistently according to the
//     specified policy.
//   - **Sanitization**: Input names are automatically sanitized to ensure consistency
//     and safety when used as identifiers.
//
// # Architecture and Data Flow
//
// The following diagram illustrates how the `Mandatory` package interacts with the
// broader status monitoring workflow:
//
//		+------------------+         +--------------------+
//		|   Configuration  |         |   Status Monitor   |
//		| (Static/Dynamic) |         |     (Poller)       |
//		+--------+---------+         +---------+----------+
//		         |                             |
//		         v                             v
//		+--------+-----------------------------+----------+
//		|                  Mandatory                      |
//		|                                                 |
//		|  +------------+          +-------------------+  |
//		|  |  Key Set   | <------- | Is Key Mandatory? |  |
//		|  | {A, B, C}  |          |     (KeyHas)      |  |
//		|  +------------+          +-------------------+  |
//		|        ^                                        |
//		|        | Add/Del                                |
//		|        v                                        |
//		|  +------------+          +-------------------+  |
//		|  | Validation | -------> |   Get Strategy    |  |
//		|  |    Mode    |          |     (GetMode)     |  |
//		|  +------------+          +-------------------+  |
//		|                                                 |
//		+-------------------------------------------------+
//
//	 1. **Configuration Phase**: The application defines groups (e.g., "Critical Services")
//	    and populates them with component keys using `KeyAdd`. The validation strategy
//	    is set using `SetMode` (e.g., `control.Must`), and descriptive metadata is
//	    added via `SetInfo` or `AddInfo`.
//
// 2. **Monitoring Phase**: When the status system evaluates the overall health:
//   - It iterates over registered components.
//   - For each component, it queries the `Mandatory` group (`KeyHas`) to see if
//     the component belongs to the group.
//   - If it does, the component's status is aggregated according to the group's
//     `Mode`, and the group's `Info` is used to enrich the response.
//
// # Usage Example
//
//	import (
//	    "fmt"
//	    "github.com/nabbar/golib/status/control"
//	    "github.com/nabbar/golib/status/mandatory"
//	)
//
//	func main() {
//	    // Create a new mandatory group for critical database components.
//	    dbGroup := mandatory.New()
//	    dbGroup.SetName("critical-db")
//
//	    // Set the validation mode to 'Must'.
//	    dbGroup.SetMode(control.Must)
//
//	    // Add descriptive metadata.
//	    dbGroup.SetInfo(map[string]interface{}{
//	        "description": "Primary database cluster",
//	        "runbook":     "https://wiki.example.com/db-failover",
//	    })
//
//	    // Add the keys of the components that belong to this group.
//	    dbGroup.KeyAdd("primary-db", "read-replica-db")
//
//	    // Check if a specific component is part of this mandatory group.
//	    if dbGroup.KeyHas("primary-db") {
//	        fmt.Println("The primary database is a mandatory component.")
//	    }
//
//	    // Retrieve the list of all keys in the group.
//	    keys := dbGroup.KeyList()
//	    fmt.Printf("Mandatory database components: %v\n", keys)
//
//	    // Retrieve the metadata.
//	    info := dbGroup.GetInfo()
//	    fmt.Printf("Group Info: %v\n", info)
//	}
package mandatory
