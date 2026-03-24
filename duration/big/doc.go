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

/*
Package big provides a custom duration type that extends the standard `time.Duration`
to support much larger time scales, up to billions of years. It is designed for applications
that need to handle very long durations, such as in astronomical calculations, geological
time scales, or long-term planning simulations.

# Quick Start

To get started, you can parse a duration string or create a duration from units:

	// Parse a duration string
	d, err := big.Parse("1000d5h10m")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d.String()) // Output: 1000d5h10m

	// Create a duration from units
	d2 := big.Days(500) + big.Hours(12)
	fmt.Println(d2.Hours()) // Output: 12012

# Architecture and Data Flow

The `big.Duration` type is fundamentally an `int64` representing a number of seconds.
This is a key difference from `time.Duration`, which is an `int64` of nanoseconds.
This trade-off (losing nanosecond precision) allows `big.Duration` to represent a
vastly larger range of time.

Here's a conceptual data flow:

 1. *Creation*: Durations can be created from strings (`Parse`), standard units like
    days/hours/minutes/seconds (`Days`, `Hours`, etc.), or from other numeric types
    (`ParseFloat64`, `ParseDuration`).

 2. *Manipulation*: The package provides functions for arithmetic operations (implicitly
    through standard operators `+`, `-`), rounding (`Round`), and truncation (`Truncate`).

 3. *Formatting*: Durations can be formatted back into strings (`String()`) or converted
    to standard Go types like `int64`, `float64`, or `time.Duration` (if within range).

 4. *Serialization*: The `big.Duration` type supports out-of-the-box serialization and
    deserialization for JSON, YAML, and TOML, making it easy to use in configuration
    files and APIs.

# Use Cases

*Long-Term Scheduling*: Scheduling tasks or events that are years or decades in the future.

	// A maintenance task scheduled for 50 years from now.
	maintenanceCycle := big.Days(50 * 365)
	nextMaintenance := time.Now().Add(maintenanceCycle.Time()) // Note: .Time() will fail if > 292 years

*Simulations*: Modeling processes over geological or astronomical time.

	// Half-life of a radioactive isotope.
	halfLife := big.Days(1600 * 365) // Approx. 1600 years for Radium-226
	fmt.Printf("Half-life is %d days\n", halfLife.Days())

*Configuration*: Specifying very long timeout or retention periods in configuration files.

	// config.yaml
	// session_timeout: "30d"
	// data_retention_policy: "10y" // Note: 'y' for year is not a standard unit, would need parsing logic.
	// A custom parser could handle this, or use days.

# Code Examples

See the `examples_test.go` file for runnable examples demonstrating various
features of the package.
*/
package big
