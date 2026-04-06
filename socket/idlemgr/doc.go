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

// Package idlemgr provides a highly efficient, thread-safe, and scalable mechanism
// for managing idle socket connections or any client-like entities that require
// automated cleanup based on inactivity periods.
//
// OVERVIEW
//
// The Idle Manager is designed to solve the common "leaky connection" problem in
// networked applications. In high-concurrency environments, tracking thousands of
// active connections and identifying those that have become stale is a non-trivial
// task. This package simplifies the process by providing a central Manager that
// monitors registered Clients and automatically closes them once their inactivity
// counter exceeds a predefined threshold.
//
// CORE CONCEPTS
//
// 1. The Manager: The central authority responsible for tracking clients and
// executing the cleanup cycle. It runs as a background service using a Ticker.
//
// 2. The Client: An interface representing any entity that can be managed.
// A client must be able to identify itself, increment its own inactivity counter,
// and be closed by the manager.
//
// 3. Sharding: To ensure high performance under heavy load, the manager uses
// a sharded storage system (32 shards by default). This drastically reduces
// lock contention between concurrent registrations, unregistrations, and the
// background cleanup process.
//
// ARCHITECTURE
//
// The architecture follows a "shared-nothing" approach between shards as much
// as possible. The Manager acts as an orchestrator for these shards.
//
// +---------------------------------------------------------------+
// |                            MANAGER                            |
// |  +---------------------------------------------------------+  |
// |  |                         TICKER                          |  |
// |  | (Periodically triggers the 'run' method for all shards) |  |
// |  +---------------------------------------------------------+  |
// |          |                    |                    |          |
// |          v                    v                    v          |
// |  +---------------+    +---------------+    +---------------+  |
// |  |    SHARD 0    |    |    SHARD 1    |    |   SHARD N     |  |
// |  | +-----------+ |    | +-----------+ |    | +-----------+ |  |
// |  | | Mutex     | |    | | Mutex     | |    | | Mutex     | |  |
// |  | +-----------+ |    | +-----------+ |    | +-----------+ |  |
// |  | | Map[Ref]C | |    | | Map[Ref]C | |    | | Map[Ref]C | |  |
// |  | +-----------+ |    | +-----------+ |    | +-----------+ |  |
// |  +---------------+    +---------------+    +---------------+  |
// |          ^                    ^                    ^          |
// |          |                    |                    |          |
// |          +--------------------+--------------------+          |
// |                               |                               |
// |                      +------------------+                     |
// |                      | HASHING FUNCTION |                     |
// |                      |    (FNV-1a)      |                     |
// |                      +------------------+                     |
// |                               ^                               |
// |                               |                               |
// |                         [Registration]                        |
// +---------------------------------------------------------------+
//
// DATA FLOW SCHEMATICS
//
// 1. Registration Flow:
//
//    [User Code] ----(Client Obj)----> [Manager.Register]
//                                              |
//                                       [Hash(Client.Ref)]
//                                              |
//                                       [Shard = Hash % 32]
//                                              |
//                                       [Shard.Lock (Write)]
//                                              |
//                                       [Shard.Map[Ref] = Client]
//                                              |
//                                       [Shard.Unlock]
//
// 2. Background Execution Flow (The Cleanup Cycle):
//
//     [Ticker Trigger]
//            |
//     +------+------------------------------------------+
//     | For Each Shard (0..31)                          |
//     |   |                                             |
//     |   |-- [Shard.Lock (Read)]                       |
//     |   |-- [Iterate Map]                             |
//     |   |      |                                      |
//     |   |      |-- [Client.Inc()]                     |
//     |   |      |-- [IF Client.Get() > Threshold]      |
//     |   |             |-- [Add to Cleanup List]       |
//     |   |-- [Shard.Unlock (Read)]                     |
//     |                                                 |
//     |   |-- [IF Cleanup List NOT Empty]               |
//     |          |-- [Spawn Goroutine]                  |
//     |                |                                |
//     |                |-- [Shard.Lock (Write)]         |
//     |                |-- [For Each in Cleanup List]   |
//     |                |      |-- [Client.Close()]      |
//     |                |      |-- [Delete from Map]     |
//     |                |-- [Shard.Unlock (Write)]       |
//     +-------------------------------------------------+
//
// 3. Manual Unregistration Flow:
//
//    [User Code] ----(Client Obj)----> [Manager.Unregister]
//                                              |
//                                       [Hash(Client.Ref)]
//                                              |
//                                       [Shard = Hash % 32]
//                                              |
//                                       [Shard.Lock (Write)]
//                                              |
//                                       [Client.Close()]
//                                              |
//                                       [Delete from Shard.Map]
//                                              |
//                                       [Shard.Unlock]
//
// DATA FLOW DESCRIPTION
//
// 1. Client Registration:
//    - User calls Manager.Register(client).
//    - Manager calculates the FNV-1a hash of the client's Reference string.
//    - Manager determines the appropriate shard: hash % 32.
//    - Manager acquires a Write lock on the shard and stores the client.
//
// 2. Background Monitoring (The 'Tick' cycle):
//    - Every 'tick' interval (defined at creation), the Ticker calls 'run'.
//    - The Manager iterates through all 32 shards.
//    - For each shard:
//        a. Acquire a Read lock.
//        b. Iterate through all clients in the map.
//        c. Call Client.Inc() on each client.
//        d. Call Client.Get() to check the current counter.
//        e. If counter > idle threshold, add client reference to a 'deletion' list.
//        f. Release Read lock.
//        g. If 'deletion' list is not empty, start a goroutine to:
//             i. Acquire Write lock on the shard.
//            ii. Double-check client existence.
//           iii. Call Client.Close().
//            iv. Remove client from the map.
//             v. Release Write lock.
//
// 3. Client Activity:
//    - Usually, the implementation of the Client interface should reset its
//      internal counter whenever activity is detected (e.g., on Read or Write).
//      Note: The Manager itself only increments (Inc). It's the Client's responsibility
//      to provide a mechanism to reset if needed, although the interface only
//      requires Inc and Get.
//
// 4. Client Unregistration:
//    - User calls Manager.Unregister(client).
//    - Manager calculates the shard, acquires Write lock, closes the client,
//      and removes it from the map.
//
// COMPONENT SPECIFICATIONS
//
// Interface: Client
//
// The Client interface is the contract for objects that want to be managed by
// the idlemgr.
//
//    type Client interface {
//        io.Closer    // Standard Close() error method.
//        Ref() string // Unique identifier for the client (used for hashing).
//        Inc()        // Increments the internal idle counter.
//        Get() uint32 // Returns the current value of the idle counter.
//    }
//
// Interface: Manager
//
// The Manager interface extends the Runner and Closer interfaces, adding methods
// for client lifecycle management.
//
//    type Manager interface {
//        libsrv.Runner             // Start, Stop, Restart, IsRunning, Uptime.
//        io.Closer                 // Close all clients and stop the manager.
//        Register(Client) error   // Add a client to the manager.
//        Unregister(Client) error // Remove and close a client.
//    }
//
// USAGE & LIFECYCLE
//
// Initialization
//
// To create a new Manager, use the New function:
//
//    mgr, err := idlemgr.New(ctx, idleDuration, tickDuration)
//
// - ctx: A context.Context used to control the manager's lifetime. If the context
//   is canceled, the manager will stop its background routine.
// - idleDuration: The maximum time a client can remain idle. Internally, the
//   manager works with units derived from the tickDuration.
// - tickDuration: The frequency at which the manager checks all clients.
//
// Starting the Manager
//
// The Manager is a Runner. It will not start monitoring clients until Start() is called.
//
//    err := mgr.Start(ctx)
//
// Monitoring and Cleanup
//
// Once started, the Manager periodically visits every shard. It is optimized to
// minimize blocking. By using 32 shards, even with a massive number of clients,
// the impact of the Write lock during cleanup is localized to 1/32nd of the total
// client base.
//
// USE CASES
//
// 1. TCP/UDP Server Management:
//    Keep track of thousands of open sockets. Automatically close connections that
//    haven't sent heartbeats or data within a specific timeframe.
//
// 2. WebSocket Orchestration:
//    Manage persistent WebSocket connections. Ensure that "ghost" connections
//    (where the client disconnected without a proper handshake) are cleaned up.
//
// 3. Resource Pooling:
//    Monitor idle resources in a custom pool (e.g., database connections, workers)
//    and retire them after periods of inactivity to save system resources.
//
// 4. Session Timeout:
//    Track user sessions in memory and automatically invalidate them after they
//    expire.
//
// QUICK START
//
// Below is a minimal example of how to implement the Client interface and use
// the Manager.
//
//    // 1. Implement the Client interface
//    type MyClient struct {
//        id      string
//        counter uint32
//        mu      sync.Mutex
//    }
//
//    func (c *MyClient) Ref() string { return c.id }
//    func (c *MyClient) Inc()        { atomic.AddUint32(&c.counter, 1) }
//    func (c *MyClient) Get() uint32 { return atomic.LoadUint32(&c.counter) }
//    func (c *MyClient) Close() error {
//        fmt.Printf("Client %s is being closed due to idleness\n", c.id)
//        return nil
//    }
//    func (c *MyClient) Reset()      { atomic.StoreUint32(&c.counter, 0) }
//
//    // 2. Initialize the Manager
//    // Inactivity threshold: 30 units. Check every 1 second.
//    // Total idle time: 30 seconds.
//    ctx := context.Background()
//    mgr, _ := idlemgr.New(ctx, durbig.New(30 * time.Second), durbig.New(time.Second))
//
//    // 3. Start the Manager
//    mgr.Start(ctx)
//
//    // 4. Register a client
//    client := &MyClient{id: "user-123"}
//    mgr.Register(client)
//
//    // 5. If activity occurs, reset the counter (implementation specific)
//    // client.Reset()
//
// ADVANCED IMPLEMENTATION DETAILS
//
// Hashing and Sharding
//
// The choice of FNV-1a for sharding is intentional. It is a non-cryptographic
// hash function with excellent distribution properties and extremely low
// computational overhead. Because it's implemented inline, it avoids the heap
// allocations typical of more complex hashing libraries.
//
// Thread Safety
//
// The Manager is designed for highly concurrent environments. All public methods
// (Register, Unregister, Start, Stop, etc.) are thread-safe. The internal state
// is protected by a combination of:
// - Sharding: Dividing the workload to reduce lock contention.
// - RWMutex: Allowing multiple concurrent readers (during the 'run' check)
//   while ensuring exclusive access for writers (during registration or deletion).
// - Goroutines: Offloading the 'Close' and 'Delete' operations to background
//   goroutines during the cleanup cycle to avoid blocking the main ticker loop.
//
// Stopping and Cleanup
//
// When Manager.Close() is called:
// 1. It iterates through all shards.
// 2. It acquires a Write lock on each shard.
// 3. It calls Close() on every registered client.
// 4. It deletes all clients from its internal maps.
// 5. It stops the background Ticker.
//
// This ensures a clean shutdown of all resources managed by the package.
//
// Error Handling
//
// The package defines two primary error constants:
// - ErrInvalidInstance: Returned when a method is called on a nil Manager.
// - ErrInvalidClient: Returned when attempting to register or unregister a nil Client.
//
// Methods like Register and Unregister return these errors to allow callers to
// handle edge cases gracefully.
//
// PERFORMANCE CONSIDERATIONS
//
// - Memory Usage: The overhead per client is minimal (a map entry and a pointer).
// - CPU Usage: The 'run' cycle's complexity is O(N/32) per shard, where N is the
//   total number of clients. Since the operations inside the Read lock (Inc and Get)
//   are very fast, the manager can handle tens of thousands of clients with
//   negligible CPU impact.
// - GC Pressure: The use of string keys in maps and the periodic deletion of
//   clients are the primary sources of GC activity. The sharding helps spread
//   this load.
//
// DESIGN PHILOSOPHY
//
// This package prioritizes "correctness under load." It assumes that the number
// of clients might be very large and that multiple components might be
// registering or unregistering clients simultaneously. By decoupling the
// "increment" logic from the "reset" logic (leaving the latter to the user's
// implementation), the Manager remains agnostic of the specific protocol or
// data type managed.
//
// SUMMARY OF RESPONSIBILITIES
//
// Manager Responsibilities:
// - Track client lifecycle (Register/Unregister).
// - Maintain a periodic tick.
// - Increment client counters.
// - Identify stale clients.
// - Execute cleanup (Close and Remove).
//
// Client Responsibilities:
// - Provide a unique Reference string.
// - Maintain an internal counter (atomically).
// - Implement a Close method that releases underlying resources (sockets, etc.).
// - (Optional) Provide a way for external activity to reset the counter.
//
// FREQUENTLY ASKED QUESTIONS
//
// Q: Why doesn't the Client interface have a Reset() method?
// A: The Manager doesn't need to reset the counter; only the user's logic
//    (which detects activity) needs to. Keeping the interface minimal makes it
//    easier to implement.
//
// Q: Can I change the number of shards?
// A: Currently, the number of shards is fixed at 32, which provides an optimal
//    balance for most modern multi-core processors.
//
// Q: What happens if a client's Close() method blocks?
// A: The Manager executes cleanup in a separate goroutine per shard, so a
//    blocking Close() on one client won't stop the Manager's main loop, but it
//    might delay further cleanup for that specific shard until the goroutine
//    finishes. It is recommended that Client.Close() implementations be
//    non-blocking or have their own timeouts.
//
// Q: Is the Reference string case-sensitive?
// A: Yes, the hashing is performed on the raw bytes of the string returned by
//    Ref(), so "ClientA" and "clienta" will hash to different values and be
//    treated as different entities.
//
// Q: Can I use this for non-socket resources?
// A: Absolutely. Any struct that implements the four methods of the Client
//    interface can be managed. This includes file handles, temporary files,
//    cached items, or even virtual actors in a simulation.
//
// EXTENSIBILITY
//
// The Manager interface is designed to be easily embeddable or wrapped. If you
// need to add logging, metrics (like Prometheus), or additional validation,
// you can create a wrapper struct that implements the Manager interface and
// delegates calls to the underlying idlemgr instance while adding your custom
// logic.
//
// Example Wrapper for Metrics:
//
//    type MetricsManager struct {
//        idlemgr.Manager
//        registerCount prometheus.Counter
//    }
//
//    func (m *MetricsManager) Register(c idlemgr.Client) error {
//        m.registerCount.Inc()
//        return m.Manager.Register(c)
//    }
//
// This flexibility makes idlemgr a robust building block for larger, more
// complex systems.
//
// CONCLUSION
//
// The idlemgr package provides a battle-tested pattern for idle resource
// management. Its focus on performance through sharding and its clean,
// interface-based design make it a versatile tool for any Go developer
// dealing with long-lived, potentially idle connections or resources.
//
// By adhering to the principles of simplicity and efficiency, it allows
// developers to focus on their core application logic while trusting the
// management of idleness to a specialized, robust component.
//
// For more details on the underlying runner system, see the
// github.com/nabbar/golib/runner package.
//
// For details on the duration handling, see the
// github.com/nabbar/golib/duration/big package.
//
// Happy coding!

package idlemgr
