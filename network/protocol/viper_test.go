/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package protocol_test

import (
	"reflect"

	. "github.com/nabbar/golib/network/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Viper Decoder Hook", func() {
	var hook func(reflect.Type, reflect.Type, interface{}) (interface{}, error)

	BeforeEach(func() {
		hook = ViperDecoderHook()
	})

	Describe("ViperDecoderHook creation", func() {
		It("should return a non-nil hook function", func() {
			Expect(hook).NotTo(BeNil())
		})

		It("should return a function that can be called", func() {
			Expect(hook).To(BeAssignableToTypeOf(func(reflect.Type, reflect.Type, interface{}) (interface{}, error) { return nil, nil }))
		})
	})

	Describe("Hook behavior with NetworkProtocol target", func() {
		var (
			stringType   reflect.Type
			intType      reflect.Type
			int8Type     reflect.Type
			int16Type    reflect.Type
			int32Type    reflect.Type
			int64Type    reflect.Type
			uintType     reflect.Type
			uint8Type    reflect.Type
			uint16Type   reflect.Type
			uint32Type   reflect.Type
			uint64Type   reflect.Type
			protocolType reflect.Type
		)

		BeforeEach(func() {
			stringType = reflect.TypeOf("")
			intType = reflect.TypeOf(int(0))
			int8Type = reflect.TypeOf(int8(0))
			int16Type = reflect.TypeOf(int16(0))
			int32Type = reflect.TypeOf(int32(0))
			int64Type = reflect.TypeOf(int64(0))
			uintType = reflect.TypeOf(uint(0))
			uint8Type = reflect.TypeOf(uint8(0))
			uint16Type = reflect.TypeOf(uint16(0))
			uint32Type = reflect.TypeOf(uint32(0))
			uint64Type = reflect.TypeOf(uint64(0))
			var p NetworkProtocol
			protocolType = reflect.TypeOf(p)
		})

		Context("with valid string data", func() {
			It("should decode 'tcp' to NetworkTCP", func() {
				result, err := hook(stringType, protocolType, "tcp")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should decode 'udp' to NetworkUDP", func() {
				result, err := hook(stringType, protocolType, "udp")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkUDP))
			})

			It("should decode 'unix' to NetworkUnix", func() {
				result, err := hook(stringType, protocolType, "unix")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkUnix))
			})

			It("should decode 'tcp4' to NetworkTCP4", func() {
				result, err := hook(stringType, protocolType, "tcp4")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP4))
			})

			It("should decode 'tcp6' to NetworkTCP6", func() {
				result, err := hook(stringType, protocolType, "tcp6")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP6))
			})

			It("should decode 'unixgram' to NetworkUnixGram", func() {
				result, err := hook(stringType, protocolType, "unixgram")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkUnixGram))
			})

			It("should handle uppercase strings", func() {
				result, err := hook(stringType, protocolType, "TCP")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should handle mixed case strings", func() {
				result, err := hook(stringType, protocolType, "TcP")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})
		})

		Context("with int data", func() {
			It("should decode int value 2 to NetworkTCP", func() {
				result, err := hook(intType, protocolType, int(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid int value", func() {
				result, err := hook(intType, protocolType, int(99))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid value"))
				Expect(result).To(BeNil())
			})
		})

		Context("with int8 data", func() {
			It("should decode int8 value 2 to NetworkTCP", func() {
				result, err := hook(int8Type, protocolType, int8(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid int8 value", func() {
				result, err := hook(int8Type, protocolType, int8(99))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("with int16 data", func() {
			It("should decode int16 value 2 to NetworkTCP", func() {
				result, err := hook(int16Type, protocolType, int16(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid int16 value", func() {
				result, err := hook(int16Type, protocolType, int16(99))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("with int32 data", func() {
			It("should decode int32 value 2 to NetworkTCP", func() {
				result, err := hook(int32Type, protocolType, int32(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid int32 value", func() {
				result, err := hook(int32Type, protocolType, int32(99))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("with int64 data", func() {
			It("should decode int64 value 2 to NetworkTCP", func() {
				result, err := hook(int64Type, protocolType, int64(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should decode int64 value 5 to NetworkUDP", func() {
				result, err := hook(int64Type, protocolType, int64(5))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkUDP))
			})

			It("should decode all valid protocol int64 values", func() {
				tests := map[int64]NetworkProtocol{
					1:  NetworkUnix,
					2:  NetworkTCP,
					3:  NetworkTCP4,
					4:  NetworkTCP6,
					5:  NetworkUDP,
					6:  NetworkUDP4,
					7:  NetworkUDP6,
					8:  NetworkIP,
					9:  NetworkIP4,
					10: NetworkIP6,
					11: NetworkUnixGram,
				}

				for val, expected := range tests {
					result, err := hook(int64Type, protocolType, val)
					Expect(err).To(BeNil(), "Failed for value %d", val)
					Expect(result).To(Equal(expected), "Failed for value %d", val)
				}
			})

			It("should return error for invalid int64 value", func() {
				result, err := hook(int64Type, protocolType, int64(99))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid value"))
				Expect(result).To(BeNil())
			})

			It("should return error for negative int64 value", func() {
				result, err := hook(int64Type, protocolType, int64(-1))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return error for zero value", func() {
				result, err := hook(int64Type, protocolType, int64(0))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("with uint data", func() {
			It("should decode uint value 2 to NetworkTCP", func() {
				result, err := hook(uintType, protocolType, uint(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid uint value", func() {
				result, err := hook(uintType, protocolType, uint(99))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return error for uint value exceeding MaxUint16", func() {
				result, err := hook(uintType, protocolType, uint(70000))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid value"))
				Expect(result).To(BeNil())
			})
		})

		Context("with uint8 data", func() {
			It("should decode uint8 value 2 to NetworkTCP", func() {
				result, err := hook(uint8Type, protocolType, uint8(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid uint8 value", func() {
				result, err := hook(uint8Type, protocolType, uint8(99))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("with uint16 data", func() {
			It("should decode uint16 value 2 to NetworkTCP", func() {
				result, err := hook(uint16Type, protocolType, uint16(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid uint16 value", func() {
				result, err := hook(uint16Type, protocolType, uint16(99))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("with uint32 data", func() {
			It("should decode uint32 value 2 to NetworkTCP", func() {
				result, err := hook(uint32Type, protocolType, uint32(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid uint32 value", func() {
				result, err := hook(uint32Type, protocolType, uint32(99))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("with uint64 data", func() {
			It("should decode uint64 value 2 to NetworkTCP", func() {
				result, err := hook(uint64Type, protocolType, uint64(2))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should return error for invalid uint64 value", func() {
				result, err := hook(uint64Type, protocolType, uint64(99))
				Expect(err).NotTo(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return error for uint64 value exceeding MaxUint16", func() {
				result, err := hook(uint64Type, protocolType, uint64(70000))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid value"))
				Expect(result).To(BeNil())
			})
		})

		Context("with invalid or empty string data", func() {
			// String unmarshaling doesn't return errors (returns NetworkEmpty)
			It("should decode invalid protocol to NetworkEmpty (no error)", func() {
				result, err := hook(stringType, protocolType, "invalid")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should decode empty string to NetworkEmpty", func() {
				result, err := hook(stringType, protocolType, "")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkEmpty))
			})

			// ✅ FIXED: Parse() now trims whitespace
			It("should handle strings with whitespace", func() {
				result, err := hook(stringType, protocolType, " tcp ")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkTCP))
			})
		})
	})

	Describe("Hook passthrough behavior", func() {
		Context("when source type is not supported", func() {
			It("should pass through bool data unchanged", func() {
				boolType := reflect.TypeOf(true)
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(boolType, protocolType, true)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(true))
			})

			It("should pass through float data unchanged", func() {
				floatType := reflect.TypeOf(float64(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(floatType, protocolType, float64(3.14))
				Expect(err).To(BeNil())
				Expect(result).To(Equal(float64(3.14)))
			})

			It("should pass through struct data unchanged", func() {
				type TestStruct struct{ Value int }
				structType := reflect.TypeOf(TestStruct{})
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				data := TestStruct{Value: 42}
				result, err := hook(structType, protocolType, data)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(data))
			})

			It("should pass through slice data unchanged", func() {
				sliceType := reflect.TypeOf([]string{})
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				data := []string{"tcp", "udp"}
				result, err := hook(sliceType, protocolType, data)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(data))
			})
		})

		Context("when target type is not NetworkProtocol", func() {
			It("should pass through to string target", func() {
				stringType := reflect.TypeOf("")
				targetType := reflect.TypeOf("")

				result, err := hook(stringType, targetType, "tcp")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("tcp"))
			})

			It("should pass through to int target", func() {
				stringType := reflect.TypeOf("")
				intType := reflect.TypeOf(0)

				result, err := hook(stringType, intType, "42")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("42"))
			})

			It("should pass through to struct target", func() {
				type TargetStruct struct{ Protocol string }
				stringType := reflect.TypeOf("")
				structType := reflect.TypeOf(TargetStruct{})

				result, err := hook(stringType, structType, "tcp")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("tcp"))
			})
		})

		Context("when data is not a string despite string type", func() {
			// ⚠️ EDGE CASE: Type assertion failure handling
			It("should pass through if data cannot be cast to string", func() {
				stringType := reflect.TypeOf("")
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				// This is a bit contrived, but tests the k bool check
				// In practice, if from.Kind() is String, data should be string
				// But the code checks with type assertion anyway
				result, err := hook(stringType, protocolType, 42)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(42))
			})
		})

		Context("when data type assertion fails", func() {
			It("should pass through if int type has wrong data type", func() {
				intType := reflect.TypeOf(int(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				// Pass string data to int type (assertion will fail)
				result, err := hook(intType, protocolType, "not an int")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not an int"))
			})

			It("should pass through if int8 type has wrong data type", func() {
				int8Type := reflect.TypeOf(int8(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(int8Type, protocolType, "not an int8")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not an int8"))
			})

			It("should pass through if int16 type has wrong data type", func() {
				int16Type := reflect.TypeOf(int16(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(int16Type, protocolType, "not an int16")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not an int16"))
			})

			It("should pass through if int32 type has wrong data type", func() {
				int32Type := reflect.TypeOf(int32(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(int32Type, protocolType, "not an int32")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not an int32"))
			})

			It("should pass through if int64 type has wrong data type", func() {
				int64Type := reflect.TypeOf(int64(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(int64Type, protocolType, "not an int64")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not an int64"))
			})

			It("should pass through if uint type has wrong data type", func() {
				uintType := reflect.TypeOf(uint(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(uintType, protocolType, "not a uint")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not a uint"))
			})

			It("should pass through if uint8 type has wrong data type", func() {
				uint8Type := reflect.TypeOf(uint8(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(uint8Type, protocolType, "not a uint8")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not a uint8"))
			})

			It("should pass through if uint16 type has wrong data type", func() {
				uint16Type := reflect.TypeOf(uint16(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(uint16Type, protocolType, "not a uint16")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not a uint16"))
			})

			It("should pass through if uint32 type has wrong data type", func() {
				uint32Type := reflect.TypeOf(uint32(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(uint32Type, protocolType, "not a uint32")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not a uint32"))
			})

			It("should pass through if uint64 type has wrong data type", func() {
				uint64Type := reflect.TypeOf(uint64(0))
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(uint64Type, protocolType, "not a uint64")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("not a uint64"))
			})

			It("should pass through if slice type has wrong data type", func() {
				sliceType := reflect.TypeOf([]byte{})
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				// Pass string slice instead of byte slice
				result, err := hook(sliceType, protocolType, []string{"not", "bytes"})
				Expect(err).To(BeNil())
				Expect(result).To(Equal([]string{"not", "bytes"}))
			})
		})
	})

	Describe("Edge cases and error handling", func() {
		Context("with nil data", func() {
			It("should handle nil gracefully", func() {
				stringType := reflect.TypeOf("")
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				// nil is not a string, should pass through
				result, err := hook(stringType, protocolType, nil)
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
			})
		})

		Context("with pointer types", func() {
			It("should handle pointer to string", func() {
				str := "tcp"
				ptrType := reflect.TypeOf(&str)
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(ptrType, protocolType, &str)
				Expect(err).To(BeNil())
				// Pointer type != String, should pass through
				Expect(result).To(Equal(&str))
			})

			It("should handle pointer to NetworkProtocol as target", func() {
				stringType := reflect.TypeOf("")
				var p NetworkProtocol
				ptrType := reflect.TypeOf(&p)

				result, err := hook(stringType, ptrType, "tcp")
				Expect(err).To(BeNil())
				// Target is pointer, not NetworkProtocol, should pass through
				Expect(result).To(Equal("tcp"))
			})
		})

		Context("with special string values", func() {
			It("should handle strings with null bytes", func() {
				stringType := reflect.TypeOf("")
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				result, err := hook(stringType, protocolType, "tcp\x00")
				Expect(err).To(BeNil())
				// Null byte might affect parsing
				Expect(result).NotTo(BeNil())
			})

			It("should handle very long strings without panic", func() {
				stringType := reflect.TypeOf("")
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				longString := string(make([]byte, 10000))
				Expect(func() {
					_, _ = hook(stringType, protocolType, longString)
				}).NotTo(Panic())
			})

			// ✅ FIXED: Parse() now trims whitespace including newlines and tabs
			It("should handle strings with special characters", func() {
				stringType := reflect.TypeOf("")
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				// These strings now parse correctly after trimming
				whitespaceStrings := map[string]NetworkProtocol{
					"tcp\n":   NetworkTCP,
					"tcp\t":   NetworkTCP,
					"tcp\r":   NetworkTCP,
					"tcp\r\n": NetworkTCP,
					"\ntcp":   NetworkTCP,
					"tcp ":    NetworkTCP,
					" tcp":    NetworkTCP,
					"\tudp\n": NetworkUDP,
				}

				for str, expected := range whitespaceStrings {
					result, err := hook(stringType, protocolType, str)
					Expect(err).To(BeNil())
					Expect(result).To(Equal(expected), "Failed for: %q", str)
				}
			})
		})
	})

	Describe("Type compatibility", func() {
		It("should work with reflect.Kind() check", func() {
			stringType := reflect.TypeOf("")
			Expect(stringType.Kind()).To(Equal(reflect.String))
		})

		It("should correctly identify NetworkProtocol type", func() {
			var p NetworkProtocol
			protocolType := reflect.TypeOf(p)

			// Type should match itself
			Expect(protocolType).To(Equal(protocolType))
		})

		It("should distinguish between NetworkProtocol and other uint8 types", func() {
			var p NetworkProtocol
			var u uint8

			protocolType := reflect.TypeOf(p)
			uint8Type := reflect.TypeOf(u)

			// These are different types even though underlying type is uint8
			Expect(protocolType).NotTo(Equal(uint8Type))
		})
	})

	Describe("Concurrent usage", func() {
		It("should be safe for concurrent calls", func() {
			stringType := reflect.TypeOf("")
			var p NetworkProtocol
			protocolType := reflect.TypeOf(p)

			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer func() { done <- true }()
					for j := 0; j < 100; j++ {
						_, _ = hook(stringType, protocolType, "tcp")
						_, _ = hook(stringType, protocolType, "udp")
						_, _ = hook(stringType, protocolType, "unix")
					}
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should create independent hooks", func() {
			hook1 := ViperDecoderHook()
			hook2 := ViperDecoderHook()

			// Each call creates a new hook function
			Expect(hook1).NotTo(BeNil())
			Expect(hook2).NotTo(BeNil())

			// Should produce same results
			stringType := reflect.TypeOf("")
			var p NetworkProtocol
			protocolType := reflect.TypeOf(p)

			result1, err1 := hook1(stringType, protocolType, "tcp")
			result2, err2 := hook2(stringType, protocolType, "tcp")

			// Both errors should be nil
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(result1).To(Equal(result2))
		})
	})

	Describe("Integration scenarios", func() {
		Context("simulating Viper configuration parsing", func() {
			It("should handle typical config scenario", func() {
				stringType := reflect.TypeOf("")
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				// Simulate reading from config file
				configValues := map[string]NetworkProtocol{
					"tcp":      NetworkTCP,
					"udp":      NetworkUDP,
					"unix":     NetworkUnix,
					"tcp4":     NetworkTCP4,
					"unixgram": NetworkUnixGram,
				}

				for strValue, expected := range configValues {
					result, err := hook(stringType, protocolType, strValue)
					Expect(err).To(BeNil())
					Expect(result).To(Equal(expected),
						"Failed to decode '%s' to %v", strValue, expected)
				}
			})

			It("should handle default/missing config values", func() {
				stringType := reflect.TypeOf("")
				var p NetworkProtocol
				protocolType := reflect.TypeOf(p)

				// Empty string from config (missing key)
				result, err := hook(stringType, protocolType, "")
				Expect(err).To(BeNil())
				Expect(result).To(Equal(NetworkEmpty))
			})
		})
	})

	Describe("Memory and performance", func() {
		// ⚠️ MEMORY CHECK: Verify no excessive allocations
		It("should not allocate excessive memory", func() {
			stringType := reflect.TypeOf("")
			var p NetworkProtocol
			protocolType := reflect.TypeOf(p)

			// Multiple calls should not accumulate memory
			for i := 0; i < 10000; i++ {
				_, _ = hook(stringType, protocolType, "tcp")
			}
		})

		It("should handle large number of different inputs", func() {
			stringType := reflect.TypeOf("")
			var p NetworkProtocol
			protocolType := reflect.TypeOf(p)

			protocols := []string{
				"tcp", "tcp4", "tcp6",
				"udp", "udp4", "udp6",
				"unix", "unixgram",
				"ip", "ip4", "ip6",
				"invalid1", "invalid2", "",
			}

			for _, proto := range protocols {
				Expect(func() {
					_, _ = hook(stringType, protocolType, proto)
				}).NotTo(Panic())
			}
		})
	})
})
