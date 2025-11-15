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

package console

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/term"
)

// PromptString displays a prompt and reads a line of text from stdin.
// The prompt text is displayed using ColorPrompt colors.
//
// Parameters:
//   - text: The prompt message to display (e.g., "Enter your name")
//
// Returns:
//   - string: The user's input (trimmed of newline)
//   - error: Any error during reading (e.g., EOF, I/O error)
//
// Example:
//
//	name, err := console.PromptString("Enter your name")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Hello, %s!\n", name)
func PromptString(text string) (string, error) {
	var (
		scn = bufio.NewScanner(os.Stdin)
		res string
		err error
	)

	printPrompt(text)

	for scn.Scan() {
		res = scn.Text()
		err = scn.Err()
		break
	}

	return res, err
}

// PromptInt displays a prompt and reads an integer from stdin.
// The input is parsed as a base-10, 64-bit signed integer.
//
// Parameters:
//   - text: The prompt message to display (e.g., "Enter your age")
//
// Returns:
//   - int64: The parsed integer value
//   - error: Input error or parse error (invalid integer format)
//
// Accepts: Any valid base-10 integer (-9223372036854775808 to 9223372036854775807)
//
// Example:
//
//	age, err := console.PromptInt("Enter your age")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("You are %d years old\n", age)
func PromptInt(text string) (int64, error) {
	if str, err := PromptString(text); err != nil {
		return 0, err
	} else {
		return strconv.ParseInt(str, 10, 64)
	}
}

// PromptUrl displays a prompt and reads a URL from stdin.
// The input is parsed and validated as a URL.
//
// Parameters:
//   - text: The prompt message to display (e.g., "Enter API endpoint")
//
// Returns:
//   - *url.URL: The parsed URL object
//   - error: Input error or parse error (invalid URL format)
//
// Accepts: Any valid URL format (http://example.com, https://api.example.com/v1, etc.)
//
// Example:
//
//	endpoint, err := console.PromptUrl("Enter API endpoint")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Connecting to: %s\n", endpoint.String())
func PromptUrl(text string) (*url.URL, error) {
	if str, err := PromptString(text); err != nil {
		return nil, err
	} else {
		return url.Parse(str)
	}
}

// PromptBool displays a prompt and reads a boolean value from stdin.
// Accepts multiple formats for true/false values.
//
// Parameters:
//   - text: The prompt message to display (e.g., "Continue? (true/false)")
//
// Returns:
//   - bool: The parsed boolean value
//   - error: Input error or parse error (invalid boolean format)
//
// Accepted values:
//   - True: "true", "TRUE", "True", "t", "T", "1"
//   - False: "false", "FALSE", "False", "f", "F", "0"
//
// Example:
//
//	confirm, err := console.PromptBool("Do you want to continue? (true/false)")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if confirm {
//	    fmt.Println("Proceeding...")
//	}
func PromptBool(text string) (bool, error) {
	if str, err := PromptString(text); err != nil {
		return false, err
	} else {
		return strconv.ParseBool(str)
	}
}

// PromptPassword displays a prompt and reads a password from stdin with hidden input.
// The input is not echoed to the terminal for security.
// A newline is printed after input to maintain proper terminal formatting.
//
// Parameters:
//   - text: The prompt message to display (e.g., "Enter password")
//
// Returns:
//   - string: The password entered by the user
//   - error: Any error during reading (e.g., terminal not available)
//
// Security Notes:
//   - Input is not displayed on screen (no echo)
//   - Password remains in memory as a string (consider using []byte and zeroing)
//   - Use HTTPS/TLS when transmitting passwords over network
//
// Example:
//
//	password, err := console.PromptPassword("Enter password")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Use password securely...
//	// Consider zeroing the password when done:
//	// for i := range password { password[i] = 0 }
func PromptPassword(text string) (string, error) {
	printPrompt(text)
	//nolint #unconvert
	res, err := term.ReadPassword(int(syscall.Stdin))
	if err == nil {
		// Only print newline if password was successfully read
		_, _ = fmt.Fprintf(os.Stdout, "\n")
	}

	return string(res), err
}

// printPrompt is an internal helper that prints the prompt text with ColorPrompt.
// Adds a colon and space after the prompt text for consistent formatting.
func printPrompt(text string) {
	if text != "" {
		ColorPrompt.Printf("%s: ", text)
	}
}
