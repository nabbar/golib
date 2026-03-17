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

// Package listmandatory provides a thread-safe mechanism for managing a collection
// of mandatory component groups. Each group defines a set of component keys, an
// associated validation mode (e.g., Must, Should, Ignore), and descriptive metadata.
//
// This package is designed for scenarios where different sets of components have
// varying levels of importance for the overall system health. For example, a
// database connection might be a 'Must' have, while a connection to a logging
// service might be a 'Should' have.
//
// # Key Features
//
//   - **Thread-Safe**: All operations are safe for concurrent use by multiple goroutines,
//     thanks to the internal use of a thread-safe map.
//   - **Group Management**: Easily add, remove, and iterate through groups of mandatory components.
//   - **Dynamic Configuration**: Modify the mandatory status of components at runtime.
//   - **Flexible Mode Resolution**: Quickly determine the validation mode for any given
//     component key by searching through the list of groups.
//
// # Architecture and Data Flow
//
// The package is centered around the `ListMandatory` interface and its default
// implementation, `model`.
//
// ## Core Components:
//
//   - `ListMandatory`: The public interface defining the contract for managing a list of
//     mandatory groups.
//   - `model`: The internal struct that implements the `ListMandatory` interface. It uses
//     a `libatm.MapTyped[string, stsmdt.Mandatory]` which is a wrapper around `sync.Map`
//     to store the groups, ensuring thread safety.
//   - `stsmdt.Mandatory`: Represents a single group of components. It contains a list of
//     component keys (strings), a single `stsctr.Mode` that applies to all keys within that group,
//     and a map of descriptive metadata.
//   - `stsctr.Mode`: An enumeration (`Ignore`, `Should`, `Must`) that specifies the level of
//     importance for a component's status.
//
// ## Data Flow Diagram:
//
// The following diagram illustrates the typical flow of operations:
//
//	+-----------------------+
//	|     Application       |
//	+-----------------------+
//	           |
//	           v
//	+------------------------+      Initializes with or adds      +----------------------+
//	|  listmandatory.New()   | ---------------------------------> | stsmdt.Mandatory     |
//	| listmandatory.Add(...) |                                    | (Group of keys/mode) |
//	+------------------------+                                    +----------------------+
//	           |
//	           v
//	+----------------------------------+
//	|      ListMandatory Instance      |
//	| (contains a thread-safe map of   |
//	|      Mandatory groups)           |
//	+----------------------------------+
//	     |           ^            |
//	     | Walk()    | Add/Del()  | GetMode(key)
//	     v           |            v
//	+-----------------+   +-----------------+   +------------------------+
//	| Iterate over    |   | Modify group    |   | Search for key in all  |
//	| all groups      |   | list            |   | groups, return mode of |
//	|                 |   |                 |   | the first match        |
//	+-----------------+   +-----------------+   +------------------------+
//
// ## Workflow Details:
//
//  1. **Instantiation**: A `ListMandatory` instance is created via `listmandatory.New()`.
//     It can be initialized with a set of `stsmdt.Mandatory` groups.
//
//  2. **Group Management**: The `Add` and `Del` methods are used to dynamically modify the
//     collection of groups. Each group is stored in the internal map using its name as the key.
//
//  3. **Mode Resolution**: The `GetMode(key)` method is the primary way to query the required
//     status for a component. It iterates through all the groups in the list. The first group
//     found to contain the specified `key` determines the `stsctr.Mode` that is returned.
//     This "first match wins" approach allows for creating override rules by ordering how
//     groups are conceptually managed. If no group contains the key, `stsctr.Ignore` is returned.
//
//  4. **Self-Healing**: The implementation includes robustness checks. During iteration
//     (e.g., in `Len`, `Walk`, `GetMode`), if an invalid entry (like a `nil` group or a group
//     with no keys) is encountered, it is automatically purged from the map. This ensures
//     the list maintains a clean and valid state.
//
// # Usage Example
//
// Here is a practical example of how to use the `listmandatory` package:
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/nabbar/golib/status/control"
//		"github.com/nabbar/golib/status/listmandatory"
//		"github.com/nabbar/golib/status/mandatory"
//	)
//
//	func main() {
//		// 1. Define mandatory groups for different parts of the system.
//		// A group for core database dependencies.
//		dbGroup := mandatory.New()
//		dbGroup.SetName("Database Dependencies")
//		dbGroup.SetMode(control.Must)
//		dbGroup.KeyAdd("postgres-main", "redis-session")
//
//		// A group for optional but recommended services.
//		monitoringGroup := mandatory.New()
//		monitoringGroup.SetName("Monitoring Services")
//		monitoringGroup.SetMode(control.Should)
//		monitoringGroup.KeyAdd("prometheus-exporter", "jaeger-agent")
//
//		// 2. Create a new list and add the defined groups.
//		list := listmandatory.New(dbGroup, monitoringGroup)
//
//		// 3. Query the mandatory mode for specific components.
//		fmt.Printf("Mode for 'postgres-main': %s\n", list.GetMode("postgres-main"))
//		fmt.Printf("Mode for 'jaeger-agent': %s\n", list.GetMode("jaeger-agent"))
//		fmt.Printf("Mode for 'unknown-service': %s\n", list.GetMode("unknown-service"))
//
//		// 4. Iterate over all groups to print their details.
//		fmt.Println("\n--- All Mandatory Groups ---")
//		list.Walk(func(name string, m mandatory.Mandatory) bool {
//			fmt.Printf("Group '%s' (Mode: %s) requires keys: %v\n", name, m.GetMode(), m.KeyList())
//			return true // Continue walking
//		})
//	}
//
// # Integration
//
// This package is part of the `golib/status` ecosystem and is designed to work
// seamlessly with:
//   - `github.com/nabbar/golib/status/mandatory`: For defining individual mandatory groups.
//   - `github.com/nabbar/golib/status/control`: For the `Mode` enumeration.
//   - `github.com/nabbar/golib/status`: For overall application status management.
//
// See Also:
//   - `github.com/nabbar/golib/status/mandatory`
//   - `github.com/nabbar/golib/status/control`
package listmandatory
