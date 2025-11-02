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

package console

import (
	"fmt"

	"github.com/fatih/color"
)

// SetColor sets the color configuration for this ColorType using a color.Color object.
// This method provides an alternative to the package-level SetColor function.
//
// Parameters:
//   - col: The color.Color instance to use (nil resets to no coloring)
//
// Thread-safe: Can be called concurrently.
//
// Example:
//
//	c := color.New(color.FgGreen, color.BgBlack)
//	console.ColorPrint.SetColor(c)
func (c ColorType) SetColor(col *color.Color) {
	if col == nil {
		lst.Store(c, color.Color{})
	} else {
		lst.Store(c, *col)
	}
}

// Println prints the text to stdout with the ColorType's color, followed by a newline.
// Output goes directly to os.Stdout.
//
// Parameters:
//   - text: The text to print
//
// Example:
//
//	console.ColorPrint.Println("Hello, World!")
func (c ColorType) Println(text string) {
	_, _ = GetColor(c).Println(text)
}

// Print prints the text to stdout with the ColorType's color, without a newline.
// Output goes directly to os.Stdout.
//
// Parameters:
//   - text: The text to print
//
// Example:
//
//	console.ColorPrint.Print("Hello")
//	console.ColorPrint.Print(" World")
func (c ColorType) Print(text string) {
	_, _ = GetColor(c).Print(text)
}

// Sprintf formats the string with the ColorType's color and returns it.
// Does not print to stdout - returns the formatted string with ANSI color codes.
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for format string
//
// Returns:
//   - Formatted string with color codes
//
// Example:
//
//	colored := console.ColorPrint.Sprintf("Hello %s", "World")
//	fmt.Println(colored) // Prints colored text
func (c ColorType) Sprintf(format string, args ...interface{}) string {
	return GetColor(c).Sprintf(format, args...)
}

// Printf prints formatted text to stdout with the ColorType's color, without a newline.
// Equivalent to Print(fmt.Sprintf(format, args...)).
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for format string
//
// Example:
//
//	console.ColorPrint.Printf("Hello %s", "World")
func (c ColorType) Printf(format string, args ...interface{}) {
	c.Print(fmt.Sprintf(format, args...))
}

// PrintLnf prints formatted text to stdout with the ColorType's color, followed by a newline.
// Equivalent to Println(fmt.Sprintf(format, args...)).
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for format string
//
// Example:
//
//	console.ColorPrint.PrintLnf("Hello %s!", "World")
func (c ColorType) PrintLnf(format string, args ...interface{}) {
	c.Println(fmt.Sprintf(format, args...))
}
