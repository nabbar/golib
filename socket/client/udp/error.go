/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package udp

import "fmt"

var (
	// ErrInstance is returned when a nil client instance is used for operations.
	// This typically indicates a programming error where a method is called on
	// a nil pointer or an uninitialized client.
	ErrInstance = fmt.Errorf("invalid instance")

	// ErrConnection is returned when attempting to perform I/O operations
	// on a client that hasn't called Connect(), or when the underlying socket
	// is nil or invalid. Call Connect() before performing operations.
	ErrConnection = fmt.Errorf("invalid connection")

	// ErrAddress is returned by New() when the provided address is empty,
	// malformed, or cannot be resolved as a valid UDP address. The address
	// must be in the format "host:port" (e.g., "localhost:8080", "192.168.1.1:9000").
	ErrAddress = fmt.Errorf("invalid dial address")
)
