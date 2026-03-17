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

// Package control provides the data types and logic for defining how component health statuses
// influence the overall application health. This is a core part of the status monitoring system,
// enabling granular control over failure propagation.
//
// # Overview
//
// In a distributed system or a complex application, "health" is rarely a binary state.
// Some components are critical (e.g., the primary database), while others are optional
// (e.g., an analytics service) or redundant (e.g., one of many cache nodes).
//
// The `control` package introduces the concept of a `Mode`. A `Mode` is attached to each
// monitored component or group of components and dictates how that component's status
// (OK, WARN, KO) aggregates into the global status.
//
// # Control Modes & Data Flow
//
// The following modes are available, ordered by their impact logic:
//
// 1. Ignore
//   - Description: The component is monitored, but its status is completely disregarded for the global health.
//   - Use Case: Experimental features, non-critical background jobs, or debugging.
//   - Data Flow:
//     [Component Status: KO/WARN/OK] --(Ignore)--> [Global Impact: None (Always OK)]
//
// 2. Should
//   - Description: The component is important but not critical. Failure results in a degraded state (WARN) rather than a system outage.
//   - Use Case: Caching layers, secondary features, metrics exporters.
//   - Data Flow:
//     [Component Status: OK]   --(Should)--> [Global Impact: OK]
//     [Component Status: WARN] --(Should)--> [Global Impact: WARN]
//     [Component Status: KO]   --(Should)--> [Global Impact: WARN]
//
// 3. Must
//   - Description: The component is critical. Its failure directly causes a global system failure.
//   - Use Case: Primary database, authentication service, main API listener.
//   - Data Flow:
//     [Component Status: OK]   --(Must)--> [Global Impact: OK]
//     [Component Status: WARN] --(Must)--> [Global Impact: WARN]
//     [Component Status: KO]   --(Must)--> [Global Impact: KO]
//
// 4. AnyOf
//   - Description: Used for a group of redundant components. The group is considered healthy if *at least one* component is healthy.
//   - Use Case: HA clusters, multiple upstream providers, load-balanced read replicas.
//   - Data Flow:
//     [Comp1: KO, Comp2: OK] --(AnyOf)--> [Global Impact: OK]
//     [Comp1: KO, Comp2: KO] --(AnyOf)--> [Global Impact: KO]
//
// 5. Quorum
//   - Description: Used for consensus-based groups. The group is considered healthy if a *majority* (>50%) of components are healthy.
//   - Use Case: Raft/Paxos clusters, distributed storage nodes.
//   - Data Flow:
//     [Comp1: OK, Comp2: OK, Comp3: KO] (2/3 OK) --(Quorum)--> [Global Impact: OK]
//     [Comp1: OK, Comp2: KO, Comp3: KO] (1/3 OK) --(Quorum)--> [Global Impact: KO]
//
// # Interoperability
//
// The `Mode` type implements standard interfaces for serialization and formatted output:
//   - `fmt.Stringer`: Returns PascalCase strings ("Must", "Should").
//   - `json.Marshaler` / `json.Unmarshaler`: JSON support.
//   - `yaml.Marshaler` / `yaml.Unmarshaler`: YAML support.
//   - `toml.Marshaler` / `toml.Unmarshaler`: TOML support.
//   - `cbor.Marshaler` / `cbor.Unmarshaler`: CBOR support.
//   - `mapstructure` hook: Integration with Viper for configuration loading.
//
// # Numeric Representation
//
// Modes can also be represented as numeric values, useful for compact storage or database columns:
//   - 0: Ignore
//   - 1: Should
//   - 2: Must
//   - 3: AnyOf
//   - 4: Quorum
package control
