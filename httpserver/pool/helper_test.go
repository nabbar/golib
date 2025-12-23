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

package pool_test

import (
	"net/http"

	libhtp "github.com/nabbar/golib/httpserver"
)

// testHandler returns a minimal HTTP handler for testing purposes.
// This handler simply returns a 404 Not Found response.
func testHandler() map[string]http.Handler {
	return map[string]http.Handler{
		"": http.NotFoundHandler(),
	}
}

// makeTestConfig creates a server configuration with a test handler.
// This is a helper function to simplify test setup by providing a valid
// configuration with all required fields.
//
// Parameters:
//   - name: Server name identifier
//   - listen: Bind address (e.g., "127.0.0.1:8080")
//   - expose: Public address (e.g., "http://localhost:8080")
//
// Returns:
//   - libhtp.Config: Fully configured server configuration ready for testing
func makeTestConfig(name, listen, expose string) libhtp.Config {
	cfg := libhtp.Config{
		Name:   name,
		Listen: listen,
		Expose: expose,
	}
	cfg.RegisterHandlerFunc(testHandler)
	return cfg
}
