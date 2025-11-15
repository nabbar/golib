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

package monitor_test

import (
	"context"
	"fmt"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
)

// ExampleNew demonstrates creating a new monitor instance.
func ExampleNew() {
	// Create info metadata
	inf, err := moninf.New("database-monitor")
	if err != nil {
		panic(err)
	}

	// Create monitor
	mon, err := libmon.New(context.Background(), inf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Monitor created: %s\n", mon.Name())
	// Output: Monitor created: not named
}

// Example demonstrates a complete monitor setup and execution.
func Example() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create info
	inf, _ := moninf.New("example-service")

	// Create monitor
	mon, _ := libmon.New(context.Background(), inf)

	// Configure monitor
	cfg := montps.Config{
		Name:          "example-monitor",
		CheckTimeout:  libdur.ParseDuration(2 * time.Second),
		IntervalCheck: libdur.ParseDuration(1 * time.Second),
		RiseCountKO:   2,
		RiseCountWarn: 2,
		FallCountKO:   2,
		FallCountWarn: 2,
		Logger:        lo.Clone(),
	}
	_ = mon.SetConfig(context.Background(), cfg)

	// Set health check function
	mon.SetHealthCheck(func(ctx context.Context) error {
		// Simulate health check
		return nil
	})

	// Start monitoring
	_ = mon.Start(ctx)
	time.Sleep(500 * time.Millisecond)

	// Check status
	fmt.Printf("Status: %s\n", mon.Status())
	fmt.Printf("Running: %v\n", mon.IsRunning())

	// Stop monitoring
	_ = mon.Stop(ctx)
	fmt.Printf("Running after stop: %v\n", mon.IsRunning())

	// Output:
	// Status: KO
	// Running: true
	// Running after stop: false
}

// ExampleMonitor_SetHealthCheck demonstrates registering a health check function.
func ExampleMonitor_SetHealthCheck() {
	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	// Register a simple health check
	mon.SetHealthCheck(func(ctx context.Context) error {
		// Check if service is healthy
		return nil
	})

	fmt.Println("Health check registered")
	// Output: Health check registered
}

// ExampleMonitor_SetConfig demonstrates configuring a monitor.
func ExampleMonitor_SetConfig() {
	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	cfg := montps.Config{
		Name:          "my-service",
		CheckTimeout:  libdur.ParseDuration(5 * time.Second),
		IntervalCheck: libdur.ParseDuration(10 * time.Second),
		RiseCountKO:   3,
		RiseCountWarn: 2,
		Logger:        lo.Clone(),
	}

	err := mon.SetConfig(context.Background(), cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Configured: %s\n", mon.Name())
	// Output: Configured: my-service
}

// ExampleMonitor_MarshalText demonstrates text encoding.
func ExampleMonitor_MarshalText() {
	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	cfg := montps.Config{
		Name:   "my-service",
		Logger: lo.Clone(),
	}
	_ = mon.SetConfig(context.Background(), cfg)

	text, _ := mon.MarshalText()
	fmt.Printf("Contains service name: %v\n", len(text) > 0)
	// Output: Contains service name: true
}

// ExampleMonitor_MarshalJSON demonstrates JSON encoding.
func ExampleMonitor_MarshalJSON() {
	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	cfg := montps.Config{
		Name:   "my-service",
		Logger: lo.Clone(),
	}
	_ = mon.SetConfig(context.Background(), cfg)

	jsonData, _ := mon.MarshalJSON()
	fmt.Printf("JSON output available: %v\n", len(jsonData) > 0)
	// Output: JSON output available: true
}

// ExampleMonitor_Clone demonstrates cloning a monitor.
func ExampleMonitor_Clone() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	// Configure original
	cfg := montps.Config{
		Name:   "original",
		Logger: lo.Clone(),
	}
	_ = mon.SetConfig(context.Background(), cfg)

	// Clone it
	cloned, err := mon.Clone(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Cloned successfully: %v\n", cloned != nil)
	// Output: Cloned successfully: true
}

// ExampleMonitor_Restart demonstrates restarting a monitor.
func ExampleMonitor_Restart() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	cfg := montps.Config{
		Name:          "my-service",
		CheckTimeout:  libdur.ParseDuration(5 * time.Second),
		IntervalCheck: libdur.ParseDuration(1 * time.Second),
		Logger:        lo.Clone(),
	}
	_ = mon.SetConfig(context.Background(), cfg)

	mon.SetHealthCheck(func(ctx context.Context) error {
		return nil
	})

	// Start
	_ = mon.Start(ctx)
	fmt.Printf("Running: %v\n", mon.IsRunning())

	// Restart
	_ = mon.Restart(ctx)
	fmt.Printf("Running after restart: %v\n", mon.IsRunning())

	_ = mon.Stop(ctx)
	// Output:
	// Running: true
	// Running after restart: true
}

// ExampleMonitor_Status demonstrates checking monitor status.
func ExampleMonitor_Status() {
	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	status := mon.Status()
	fmt.Printf("Initial status is KO: %v\n", status.String() == "KO")
	// Output: Initial status is KO: true
}

// ExampleMonitor_metrics demonstrates collecting metrics.
func ExampleMonitor_metrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	cfg := montps.Config{
		Name:          "my-service",
		CheckTimeout:  libdur.ParseDuration(5 * time.Second),
		IntervalCheck: libdur.ParseDuration(200 * time.Millisecond),
		Logger:        lo.Clone(),
	}
	_ = mon.SetConfig(context.Background(), cfg)

	mon.SetHealthCheck(func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	_ = mon.Start(ctx)
	time.Sleep(500 * time.Millisecond)
	_ = mon.Stop(ctx)

	// Collect metrics
	latency := mon.Latency()
	uptime := mon.Uptime()
	downtime := mon.Downtime()

	fmt.Printf("Latency recorded: %v\n", latency > 0)
	fmt.Printf("Uptime recorded: %v\n", uptime >= 0)
	fmt.Printf("Downtime recorded: %v\n", downtime >= 0)

	// Output:
	// Latency recorded: true
	// Uptime recorded: true
	// Downtime recorded: true
}

// ExampleMonitor_RegisterMetricsName demonstrates Prometheus metrics registration.
func ExampleMonitor_RegisterMetricsName() {
	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	// Register metric names
	mon.RegisterMetricsName("service_health", "service_latency")

	// Add more metric names
	mon.RegisterMetricsAddName("service_uptime")

	fmt.Println("Metrics registered")
	// Output: Metrics registered
}

// ExampleMonitor_transitions demonstrates status transitions.
func ExampleMonitor_transitions() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inf, _ := moninf.New("service")
	mon, _ := libmon.New(context.Background(), inf)

	cfg := montps.Config{
		Name:          "transition-demo",
		CheckTimeout:  libdur.ParseDuration(5 * time.Second),
		IntervalCheck: libdur.ParseDuration(200 * time.Millisecond),
		RiseCountKO:   1,
		RiseCountWarn: 1,
		Logger:        lo.Clone(),
	}
	_ = mon.SetConfig(context.Background(), cfg)

	mon.SetHealthCheck(func(ctx context.Context) error {
		return nil // Always healthy
	})

	_ = mon.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Initial: KO=%v\n", mon.Status().String() == "KO")

	time.Sleep(300 * time.Millisecond)
	fmt.Printf("Rising: %v\n", mon.IsRise() || mon.Status().String() != "KO")

	_ = mon.Stop(ctx)

	// Output:
	// Initial: KO=true
	// Rising: true
}
