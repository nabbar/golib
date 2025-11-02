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

package cobra_test

import (
	"time"

	libcbr "github.com/nabbar/golib/cobra"
	libver "github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cobra Flags", func() {
	var (
		cobra   libcbr.Cobra
		version libver.Version
	)

	BeforeEach(func() {
		cobra = libcbr.New()
		version = libver.NewVersion(
			libver.License_MIT,
			"testapp",
			"Test Description",
			"2024-01-01",
			"abc123",
			"v1.0.0",
			"Test Author",
			"test-app",
			struct{}{},
			0,
		)
		cobra.SetVersion(version)
		cobra.Init()
	})

	Describe("Config Flag", func() {
		It("should add persistent config flag", func() {
			var configFile string
			err := cobra.SetFlagConfig(true, &configFile)

			Expect(err).ToNot(HaveOccurred())
			Expect(cobra.Cobra().PersistentFlags().Lookup("config")).ToNot(BeNil())
		})

		It("should add non-persistent config flag", func() {
			var configFile string
			err := cobra.SetFlagConfig(false, &configFile)

			Expect(err).ToNot(HaveOccurred())
			Expect(cobra.Cobra().Flags().Lookup("config")).ToNot(BeNil())
		})

		It("should have config shorthand 'c'", func() {
			var configFile string
			err := cobra.SetFlagConfig(true, &configFile)

			Expect(err).ToNot(HaveOccurred())
			flag := cobra.Cobra().PersistentFlags().Lookup("config")
			Expect(flag.Shorthand).To(Equal("c"))
		})
	})

	Describe("Verbose Flag", func() {
		It("should add persistent verbose flag", func() {
			var verbose int
			cobra.SetFlagVerbose(true, &verbose)

			Expect(cobra.Cobra().PersistentFlags().Lookup("verbose")).ToNot(BeNil())
		})

		It("should add non-persistent verbose flag", func() {
			var verbose int
			cobra.SetFlagVerbose(false, &verbose)

			Expect(cobra.Cobra().Flags().Lookup("verbose")).ToNot(BeNil())
		})

		It("should have verbose shorthand 'v'", func() {
			var verbose int
			cobra.SetFlagVerbose(true, &verbose)

			flag := cobra.Cobra().PersistentFlags().Lookup("verbose")
			Expect(flag.Shorthand).To(Equal("v"))
		})
	})

	Describe("String Flag", func() {
		It("should add persistent string flag", func() {
			var value string
			cobra.AddFlagString(true, &value, "test", "t", "default", "test usage")

			flag := cobra.Cobra().PersistentFlags().Lookup("test")
			Expect(flag).ToNot(BeNil())
			Expect(flag.DefValue).To(Equal("default"))
		})

		It("should add non-persistent string flag", func() {
			var value string
			cobra.AddFlagString(false, &value, "test", "t", "default", "test usage")

			flag := cobra.Cobra().Flags().Lookup("test")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Count Flag", func() {
		It("should add persistent count flag", func() {
			var value int
			cobra.AddFlagCount(true, &value, "count", "n", "count usage")

			flag := cobra.Cobra().PersistentFlags().Lookup("count")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent count flag", func() {
			var value int
			cobra.AddFlagCount(false, &value, "count", "n", "count usage")

			flag := cobra.Cobra().Flags().Lookup("count")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Bool Flag", func() {
		It("should add persistent bool flag", func() {
			var value bool
			cobra.AddFlagBool(true, &value, "enable", "e", false, "enable feature")

			flag := cobra.Cobra().PersistentFlags().Lookup("enable")
			Expect(flag).ToNot(BeNil())
			Expect(flag.DefValue).To(Equal("false"))
		})

		It("should add non-persistent bool flag with default true", func() {
			var value bool
			cobra.AddFlagBool(false, &value, "enable", "e", true, "enable feature")

			flag := cobra.Cobra().Flags().Lookup("enable")
			Expect(flag).ToNot(BeNil())
			Expect(flag.DefValue).To(Equal("true"))
		})
	})

	Describe("Duration Flag", func() {
		It("should add persistent duration flag", func() {
			var value time.Duration
			cobra.AddFlagDuration(true, &value, "timeout", "t", 30*time.Second, "timeout duration")

			flag := cobra.Cobra().PersistentFlags().Lookup("timeout")
			Expect(flag).ToNot(BeNil())
			Expect(flag.DefValue).To(Equal("30s"))
		})

		It("should add non-persistent duration flag", func() {
			var value time.Duration
			cobra.AddFlagDuration(false, &value, "timeout", "t", 1*time.Minute, "timeout duration")

			flag := cobra.Cobra().Flags().Lookup("timeout")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Float32 Flag", func() {
		It("should add persistent float32 flag", func() {
			var value float32
			cobra.AddFlagFloat32(true, &value, "ratio", "r", 1.5, "ratio value")

			flag := cobra.Cobra().PersistentFlags().Lookup("ratio")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent float32 flag", func() {
			var value float32
			cobra.AddFlagFloat32(false, &value, "ratio", "r", 2.5, "ratio value")

			flag := cobra.Cobra().Flags().Lookup("ratio")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Float64 Flag", func() {
		It("should add persistent float64 flag", func() {
			var value float64
			cobra.AddFlagFloat64(true, &value, "percent", "p", 50.5, "percentage")

			flag := cobra.Cobra().PersistentFlags().Lookup("percent")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent float64 flag", func() {
			var value float64
			cobra.AddFlagFloat64(false, &value, "percent", "p", 75.5, "percentage")

			flag := cobra.Cobra().Flags().Lookup("percent")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Int Flag", func() {
		It("should add persistent int flag", func() {
			var value int
			cobra.AddFlagInt(true, &value, "num", "n", 10, "number")

			flag := cobra.Cobra().PersistentFlags().Lookup("num")
			Expect(flag).ToNot(BeNil())
			Expect(flag.DefValue).To(Equal("10"))
		})

		It("should add non-persistent int flag", func() {
			var value int
			cobra.AddFlagInt(false, &value, "num", "n", 20, "number")

			flag := cobra.Cobra().Flags().Lookup("num")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Int8 Flag", func() {
		It("should add persistent int8 flag", func() {
			var value int8
			cobra.AddFlagInt8(true, &value, "small", "s", 5, "small number")

			flag := cobra.Cobra().PersistentFlags().Lookup("small")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent int8 flag", func() {
			var value int8
			cobra.AddFlagInt8(false, &value, "small", "s", 10, "small number")

			flag := cobra.Cobra().Flags().Lookup("small")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Int16 Flag", func() {
		It("should add persistent int16 flag", func() {
			var value int16
			cobra.AddFlagInt16(true, &value, "medium", "m", 100, "medium number")

			flag := cobra.Cobra().PersistentFlags().Lookup("medium")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent int16 flag", func() {
			var value int16
			cobra.AddFlagInt16(false, &value, "medium", "m", 200, "medium number")

			flag := cobra.Cobra().Flags().Lookup("medium")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Int32 Flag", func() {
		It("should add persistent int32 flag", func() {
			var value int32
			cobra.AddFlagInt32(true, &value, "large", "l", 1000, "large number")

			flag := cobra.Cobra().PersistentFlags().Lookup("large")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent int32 flag", func() {
			var value int32
			cobra.AddFlagInt32(false, &value, "large", "l", 2000, "large number")

			flag := cobra.Cobra().Flags().Lookup("large")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Int32Slice Flag", func() {
		It("should add persistent int32 slice flag", func() {
			var value []int32
			cobra.AddFlagInt32Slice(true, &value, "nums", "n", []int32{1, 2, 3}, "number list")

			flag := cobra.Cobra().PersistentFlags().Lookup("nums")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent int32 slice flag", func() {
			var value []int32
			cobra.AddFlagInt32Slice(false, &value, "nums", "n", []int32{4, 5, 6}, "number list")

			flag := cobra.Cobra().Flags().Lookup("nums")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Int64 Flag", func() {
		It("should add persistent int64 flag", func() {
			var value int64
			cobra.AddFlagInt64(true, &value, "huge", "h", 1000000, "huge number")

			flag := cobra.Cobra().PersistentFlags().Lookup("huge")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent int64 flag", func() {
			var value int64
			cobra.AddFlagInt64(false, &value, "huge", "h", 2000000, "huge number")

			flag := cobra.Cobra().Flags().Lookup("huge")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Int64Slice Flag", func() {
		It("should add persistent int64 slice flag", func() {
			var value []int64
			cobra.AddFlagInt64Slice(true, &value, "bigNums", "b", []int64{100, 200}, "big numbers")

			flag := cobra.Cobra().PersistentFlags().Lookup("bigNums")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent int64 slice flag", func() {
			var value []int64
			cobra.AddFlagInt64Slice(false, &value, "bigNums", "b", []int64{300, 400}, "big numbers")

			flag := cobra.Cobra().Flags().Lookup("bigNums")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Uint Flag", func() {
		It("should add persistent uint flag", func() {
			var value uint
			cobra.AddFlagUint(true, &value, "count", "c", 10, "count")

			flag := cobra.Cobra().PersistentFlags().Lookup("count")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent uint flag", func() {
			var value uint
			cobra.AddFlagUint(false, &value, "count", "c", 20, "count")

			flag := cobra.Cobra().Flags().Lookup("count")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("UintSlice Flag", func() {
		It("should add persistent uint slice flag", func() {
			var value []uint
			cobra.AddFlagUintSlice(true, &value, "ports", "p", []uint{80, 443}, "ports")

			flag := cobra.Cobra().PersistentFlags().Lookup("ports")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent uint slice flag", func() {
			var value []uint
			cobra.AddFlagUintSlice(false, &value, "ports", "p", []uint{8080, 8443}, "ports")

			flag := cobra.Cobra().Flags().Lookup("ports")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Uint8 Flag", func() {
		It("should add persistent uint8 flag", func() {
			var value uint8
			cobra.AddFlagUint8(true, &value, "byte", "b", 255, "byte value")

			flag := cobra.Cobra().PersistentFlags().Lookup("byte")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent uint8 flag", func() {
			var value uint8
			cobra.AddFlagUint8(false, &value, "byte", "b", 128, "byte value")

			flag := cobra.Cobra().Flags().Lookup("byte")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Uint16 Flag", func() {
		It("should add persistent uint16 flag", func() {
			var value uint16
			cobra.AddFlagUint16(true, &value, "port", "p", 8080, "port number")

			flag := cobra.Cobra().PersistentFlags().Lookup("port")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent uint16 flag", func() {
			var value uint16
			cobra.AddFlagUint16(false, &value, "port", "p", 9090, "port number")

			flag := cobra.Cobra().Flags().Lookup("port")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Uint32 Flag", func() {
		It("should add persistent uint32 flag", func() {
			var value uint32
			cobra.AddFlagUint32(true, &value, "id", "i", 12345, "identifier")

			flag := cobra.Cobra().PersistentFlags().Lookup("id")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent uint32 flag", func() {
			var value uint32
			cobra.AddFlagUint32(false, &value, "id", "i", 54321, "identifier")

			flag := cobra.Cobra().Flags().Lookup("id")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Uint64 Flag", func() {
		It("should add persistent uint64 flag", func() {
			var value uint64
			cobra.AddFlagUint64(true, &value, "bigId", "B", 1234567890, "big identifier")

			flag := cobra.Cobra().PersistentFlags().Lookup("bigId")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent uint64 flag", func() {
			var value uint64
			cobra.AddFlagUint64(false, &value, "bigId", "B", 9876543210, "big identifier")

			flag := cobra.Cobra().Flags().Lookup("bigId")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("StringArray Flag", func() {
		It("should add persistent string array flag", func() {
			var value []string
			cobra.AddFlagStringArray(true, &value, "names", "n", []string{"a", "b"}, "names")

			flag := cobra.Cobra().PersistentFlags().Lookup("names")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent string array flag", func() {
			var value []string
			cobra.AddFlagStringArray(false, &value, "names", "n", []string{"c", "d"}, "names")

			flag := cobra.Cobra().Flags().Lookup("names")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("StringToInt Flag", func() {
		It("should add persistent string to int flag", func() {
			var value map[string]int
			cobra.AddFlagStringToInt(true, &value, "map", "m", map[string]int{"key": 1}, "mapping")

			flag := cobra.Cobra().PersistentFlags().Lookup("map")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent string to int flag", func() {
			var value map[string]int
			cobra.AddFlagStringToInt(false, &value, "map", "m", map[string]int{"key": 2}, "mapping")

			flag := cobra.Cobra().Flags().Lookup("map")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("StringToInt64 Flag", func() {
		It("should add persistent string to int64 flag", func() {
			var value map[string]int64
			cobra.AddFlagStringToInt64(true, &value, "bigMap", "B", map[string]int64{"key": 100}, "big mapping")

			flag := cobra.Cobra().PersistentFlags().Lookup("bigMap")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent string to int64 flag", func() {
			var value map[string]int64
			cobra.AddFlagStringToInt64(false, &value, "bigMap", "B", map[string]int64{"key": 200}, "big mapping")

			flag := cobra.Cobra().Flags().Lookup("bigMap")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("StringToString Flag", func() {
		It("should add persistent string to string flag", func() {
			var value map[string]string
			cobra.AddFlagStringToString(true, &value, "labels", "l", map[string]string{"env": "prod"}, "labels")

			flag := cobra.Cobra().PersistentFlags().Lookup("labels")
			Expect(flag).ToNot(BeNil())
		})

		It("should add non-persistent string to string flag", func() {
			var value map[string]string
			cobra.AddFlagStringToString(false, &value, "labels", "l", map[string]string{"env": "dev"}, "labels")

			flag := cobra.Cobra().Flags().Lookup("labels")
			Expect(flag).ToNot(BeNil())
		})
	})

	Describe("Multiple Flags", func() {
		It("should add multiple flags without conflict", func() {
			var (
				str      string
				num      int
				enable   bool
				duration time.Duration
			)

			cobra.AddFlagString(true, &str, "string", "s", "default", "string flag")
			cobra.AddFlagInt(true, &num, "number", "n", 10, "number flag")
			cobra.AddFlagBool(true, &enable, "enable", "e", false, "enable flag")
			cobra.AddFlagDuration(true, &duration, "timeout", "t", 30*time.Second, "timeout flag")

			Expect(cobra.Cobra().PersistentFlags().Lookup("string")).ToNot(BeNil())
			Expect(cobra.Cobra().PersistentFlags().Lookup("number")).ToNot(BeNil())
			Expect(cobra.Cobra().PersistentFlags().Lookup("enable")).ToNot(BeNil())
			Expect(cobra.Cobra().PersistentFlags().Lookup("timeout")).ToNot(BeNil())
		})
	})
})
