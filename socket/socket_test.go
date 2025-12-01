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
 */

package socket_test

import (
	"fmt"
	"testing"

	libsck "github.com/nabbar/golib/socket"
)

// TestErrorFilter tests the ErrorFilter function with various error scenarios.
func TestErrorFilter(t *testing.T) {
	tests := []struct {
		nam string
		err error
		exp error
	}{
		{
			nam: "nil error",
			err: nil,
			exp: nil,
		},
		{
			nam: "closed connection error",
			err: fmt.Errorf("use of closed network connection"),
			exp: nil,
		},
		{
			nam: "normal error",
			err: fmt.Errorf("connection timeout"),
			exp: fmt.Errorf("connection timeout"),
		},
		{
			nam: "connection refused",
			err: fmt.Errorf("connection refused"),
			exp: fmt.Errorf("connection refused"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.nam, func(t *testing.T) {
			res := libsck.ErrorFilter(tc.err)

			if tc.exp == nil {
				if res != nil {
					t.Errorf("Expected nil, got %v", res)
				}
			} else {
				if res == nil {
					t.Errorf("Expected error, got nil")
				} else if res.Error() != tc.exp.Error() {
					t.Errorf("Expected %v, got %v", tc.exp, res)
				}
			}
		})
	}
}

// TestConnState_String tests the String method for all connection states.
func TestConnState_String(t *testing.T) {
	tests := []struct {
		sta libsck.ConnState
		exp string
	}{
		{libsck.ConnectionDial, "Dial Connection"},
		{libsck.ConnectionNew, "New Connection"},
		{libsck.ConnectionRead, "Read Incoming Stream"},
		{libsck.ConnectionCloseRead, "Close Incoming Stream"},
		{libsck.ConnectionHandler, "Run HandlerFunc"},
		{libsck.ConnectionWrite, "Write Outgoing Steam"},
		{libsck.ConnectionCloseWrite, "Close Outgoing Stream"},
		{libsck.ConnectionClose, "Close Connection"},
		{libsck.ConnState(255), "unknown connection state"},
	}

	for _, tc := range tests {
		t.Run(tc.exp, func(t *testing.T) {
			if got := tc.sta.String(); got != tc.exp {
				t.Errorf("ConnState(%d).String() = %q, want %q", tc.sta, got, tc.exp)
			}
		})
	}
}

// TestConnState_Values tests that all connection state constants have expected values.
func TestConnState_Values(t *testing.T) {
	if libsck.ConnectionDial != 0 {
		t.Errorf("ConnectionDial should be 0, got %d", libsck.ConnectionDial)
	}
	if libsck.ConnectionNew != 1 {
		t.Errorf("ConnectionNew should be 1, got %d", libsck.ConnectionNew)
	}
	if libsck.ConnectionRead != 2 {
		t.Errorf("ConnectionRead should be 2, got %d", libsck.ConnectionRead)
	}
	if libsck.ConnectionCloseRead != 3 {
		t.Errorf("ConnectionCloseRead should be 3, got %d", libsck.ConnectionCloseRead)
	}
	if libsck.ConnectionHandler != 4 {
		t.Errorf("ConnectionHandler should be 4, got %d", libsck.ConnectionHandler)
	}
	if libsck.ConnectionWrite != 5 {
		t.Errorf("ConnectionWrite should be 5, got %d", libsck.ConnectionWrite)
	}
	if libsck.ConnectionCloseWrite != 6 {
		t.Errorf("ConnectionCloseWrite should be 6, got %d", libsck.ConnectionCloseWrite)
	}
	if libsck.ConnectionClose != 7 {
		t.Errorf("ConnectionClose should be 7, got %d", libsck.ConnectionClose)
	}
}

// TestDefaultBufferSize tests the default buffer size constant.
func TestDefaultBufferSize(t *testing.T) {
	exp := 32 * 1024
	if libsck.DefaultBufferSize != exp {
		t.Errorf("DefaultBufferSize = %d, want %d", libsck.DefaultBufferSize, exp)
	}
}

// TestEOL tests the end-of-line constant.
func TestEOL(t *testing.T) {
	if libsck.EOL != '\n' {
		t.Errorf("EOL = %q, want %q", libsck.EOL, '\n')
	}
}

// BenchmarkErrorFilter benchmarks the ErrorFilter function.
func BenchmarkErrorFilter(b *testing.B) {
	err := fmt.Errorf("connection timeout")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = libsck.ErrorFilter(err)
	}
}

// BenchmarkErrorFilter_Nil benchmarks ErrorFilter with nil error.
func BenchmarkErrorFilter_Nil(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = libsck.ErrorFilter(nil)
	}
}

// BenchmarkErrorFilter_Closed benchmarks ErrorFilter with closed connection error.
func BenchmarkErrorFilter_Closed(b *testing.B) {
	err := fmt.Errorf("use of closed network connection")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = libsck.ErrorFilter(err)
	}
}

// BenchmarkConnState_String benchmarks the ConnState.String method.
func BenchmarkConnState_String(b *testing.B) {
	state := libsck.ConnectionNew
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = state.String()
	}
}
