/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

// Package console provides enhanced terminal I/O utilities for Go applications.
// It offers colored output, interactive prompts, text formatting, and UTF-8 text padding.
//
// The package wraps fatih/color for ANSI color support and adds:
//   - Thread-safe color management with multiple color types
//   - UTF-8 aware text padding (left, right, center)
//   - Interactive user prompts with type validation
//   - Hierarchical output with indentation
//   - Buffer writing for testing and non-terminal output
//
// Example usage:
//
//	import (
//	    "github.com/fatih/color"
//	    "github.com/nabbar/golib/console"
//	)
//
//	// Set colors
//	console.SetColor(console.ColorPrint, int(color.FgCyan), int(color.Bold))
//
//	// Print colored text
//	console.ColorPrint.Println("Hello, World!")
//
//	// Get user input
//	name, _ := console.PromptString("Enter your name")
//
//	// Format text
//	title := console.PadCenter("My App", 40, "=")
package console

import (
	"github.com/fatih/color"
	libatm "github.com/nabbar/golib/atomic"
)

// lst is the thread-safe atomic map storing ColorType to color.Color mappings.
// It enables concurrent color operations without explicit locking.
var lst = libatm.NewMapTyped[ColorType, color.Color]()

// ColorType represents a color scheme identifier.
// Each ColorType can have its own color configuration (foreground, background, attributes).
// This allows different parts of the application to use different color schemes.
type ColorType uint8

const (
	// ColorPrint is the default color type for standard output.
	// Use this for general application output, messages, and formatted text.
	ColorPrint ColorType = iota

	// ColorPrompt is the color type for user prompts and interactive input.
	// Use this for questions, input requests, and interactive elements.
	ColorPrompt
)

// GetColorType converts a uint8 to a ColorType.
// This is useful when loading color types from configuration files or external sources.
//
// Parameters:
//   - id: The numeric identifier (0 = ColorPrint, 1 = ColorPrompt, etc.)
//
// Returns:
//   - ColorType corresponding to the given id
//
// Example:
//
//	ct := console.GetColorType(0) // Returns ColorPrint
func GetColorType(id uint8) ColorType {
	return ColorType(id)
}

// SetColor configures the color attributes for a ColorType.
// Multiple attributes can be combined (e.g., color + bold + underline).
//
// Parameters:
//   - id: The ColorType to configure
//   - value: Variadic list of color attributes (from fatih/color package)
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Example:
//
//	// Set single attribute
//	console.SetColor(console.ColorPrint, int(color.FgRed))
//
//	// Set multiple attributes
//	console.SetColor(console.ColorPrint,
//	    int(color.FgYellow),
//	    int(color.Bold),
//	    int(color.Underline))
func SetColor(id ColorType, value ...int) {
	var cols = make([]color.Attribute, 0)

	for _, v := range value {
		cols = append(cols, color.Attribute(v))
	}

	a := color.New(cols...)
	if a == nil {
		lst.Store(id, color.Color{})
	} else {
		lst.Store(id, *a)
	}
}

// GetColor retrieves the color.Color instance for a ColorType.
// Returns an empty color.Color if the ColorType hasn't been configured.
//
// Parameters:
//   - id: The ColorType to retrieve
//
// Returns:
//   - Pointer to color.Color instance (never nil)
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Example:
//
//	c := console.GetColor(console.ColorPrint)
//	c.Println("Colored text")
func GetColor(id ColorType) *color.Color {
	if v, k := lst.Load(id); k {
		return &v
	} else {
		return &color.Color{}
	}
}

// DelColor removes the color configuration for a ColorType.
// After deletion, the ColorType will use an empty color (no coloring).
//
// Parameters:
//   - id: The ColorType to remove
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Example:
//
//	console.DelColor(console.ColorPrint)
func DelColor(id ColorType) {
	lst.Delete(id)
}
