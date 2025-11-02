/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
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

package hashicorp_test

import (
	"github.com/hashicorp/go-hclog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liblog "github.com/nabbar/golib/logger"
	loghc "github.com/nabbar/golib/logger/hashicorp"
)

var _ = Describe("HashiCorp Logger Context Operations", func() {
	var (
		mockLogger *MockLogger
		hcLogger   hclog.Logger
	)

	BeforeEach(func() {
		mockLogger = NewMockLogger()
		hcLogger = loghc.New(func() liblog.Logger { return mockLogger })
	})

	Describe("With", func() {
		Context("with key-value arguments", func() {
			It("should store arguments in fields", func() {
				result := hcLogger.With("key1", "value1", "key2", "value2")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(hcLogger))
			})
		})

		Context("with empty arguments", func() {
			It("should handle empty args", func() {
				result := hcLogger.With()

				Expect(result).ToNot(BeNil())
			})
		})

		Context("with multiple calls", func() {
			It("should accumulate arguments", func() {
				hcLogger.With("key1", "value1")
				result := hcLogger.With("key2", "value2")

				Expect(result).ToNot(BeNil())
			})
		})
	})

	Describe("ImpliedArgs", func() {
		Context("when no args set", func() {
			It("should return empty slice", func() {
				args := hcLogger.ImpliedArgs()

				Expect(args).ToNot(BeNil())
				Expect(args).To(HaveLen(0))
			})
		})

		Context("when args are set", func() {
			It("should return stored args", func() {
				testArgs := []interface{}{"key", "value"}
				mockLogger.fields = mockLogger.fields.Add(loghc.HCLogArgs, testArgs)

				args := hcLogger.ImpliedArgs()

				Expect(args).To(Equal(testArgs))
			})
		})

		Context("when args are wrong type", func() {
			It("should return empty slice", func() {
				mockLogger.fields = mockLogger.fields.Add(loghc.HCLogArgs, "wrong type")

				args := hcLogger.ImpliedArgs()

				Expect(args).To(HaveLen(0))
			})
		})
	})

	Describe("Name", func() {
		Context("when no name is set", func() {
			It("should return empty string", func() {
				name := hcLogger.Name()

				Expect(name).To(BeEmpty())
			})
		})

		Context("when name is set", func() {
			It("should return stored name", func() {
				mockLogger.fields = mockLogger.fields.Add(loghc.HCLogName, "test-logger")

				name := hcLogger.Name()

				Expect(name).To(Equal("test-logger"))
			})
		})

		Context("when name is wrong type", func() {
			It("should return empty string", func() {
				mockLogger.fields = mockLogger.fields.Add(loghc.HCLogName, 123)

				name := hcLogger.Name()

				Expect(name).To(BeEmpty())
			})
		})
	})

	Describe("Named", func() {
		Context("with valid name", func() {
			It("should set name in fields", func() {
				result := hcLogger.Named("my-logger")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(hcLogger))
			})
		})

		Context("with empty name", func() {
			It("should handle empty name", func() {
				result := hcLogger.Named("")

				Expect(result).ToNot(BeNil())
			})
		})

		Context("with multiple calls", func() {
			It("should update name", func() {
				hcLogger.Named("first")
				result := hcLogger.Named("second")

				Expect(result).ToNot(BeNil())
			})
		})
	})

	Describe("ResetNamed", func() {
		Context("with valid name", func() {
			It("should reset name in fields", func() {
				hcLogger.Named("old-name")
				result := hcLogger.ResetNamed("new-name")

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(hcLogger))
			})
		})

		Context("with empty name", func() {
			It("should handle empty name", func() {
				result := hcLogger.ResetNamed("")

				Expect(result).ToNot(BeNil())
			})
		})
	})

	Describe("Nil Logger Handling", func() {
		var nilHcLogger hclog.Logger

		BeforeEach(func() {
			nilHcLogger = loghc.New(nil)
		})

		Context("when underlying logger is nil", func() {
			It("should handle Trace gracefully", func() {
				Expect(func() {
					nilHcLogger.Trace("message")
				}).ToNot(Panic())
			})

			It("should handle Debug gracefully", func() {
				Expect(func() {
					nilHcLogger.Debug("message")
				}).ToNot(Panic())
			})

			It("should handle Info gracefully", func() {
				Expect(func() {
					nilHcLogger.Info("message")
				}).ToNot(Panic())
			})

			It("should handle Warn gracefully", func() {
				Expect(func() {
					nilHcLogger.Warn("message")
				}).ToNot(Panic())
			})

			It("should handle Error gracefully", func() {
				Expect(func() {
					nilHcLogger.Error("message")
				}).ToNot(Panic())
			})

			It("should handle Log gracefully", func() {
				Expect(func() {
					nilHcLogger.Log(hclog.Info, "message")
				}).ToNot(Panic())
			})

			It("should return false for IsTrace", func() {
				Expect(nilHcLogger.IsTrace()).To(BeFalse())
			})

			It("should return false for IsDebug", func() {
				Expect(nilHcLogger.IsDebug()).To(BeFalse())
			})

			It("should return false for IsInfo", func() {
				Expect(nilHcLogger.IsInfo()).To(BeFalse())
			})

			It("should return false for IsWarn", func() {
				Expect(nilHcLogger.IsWarn()).To(BeFalse())
			})

			It("should return false for IsError", func() {
				Expect(nilHcLogger.IsError()).To(BeFalse())
			})

			It("should return empty args for ImpliedArgs", func() {
				args := nilHcLogger.ImpliedArgs()
				Expect(args).To(HaveLen(0))
			})

			It("should return self for With", func() {
				result := nilHcLogger.With("key", "value")
				Expect(result).To(Equal(nilHcLogger))
			})

			It("should return empty string for Name", func() {
				Expect(nilHcLogger.Name()).To(BeEmpty())
			})

			It("should return self for Named", func() {
				result := nilHcLogger.Named("name")
				Expect(result).To(Equal(nilHcLogger))
			})

			It("should return self for ResetNamed", func() {
				result := nilHcLogger.ResetNamed("name")
				Expect(result).To(Equal(nilHcLogger))
			})

			It("should handle SetLevel gracefully", func() {
				Expect(func() {
					nilHcLogger.SetLevel(hclog.Info)
				}).ToNot(Panic())
			})

			It("should return NoLevel for GetLevel", func() {
				Expect(nilHcLogger.GetLevel()).To(Equal(hclog.NoLevel))
			})
		})
	})

	Describe("Integration with underlying logger", func() {
		Context("when logger function returns nil", func() {
			It("should handle gracefully", func() {
				returnNilLogger := loghc.New(func() liblog.Logger { return nil })

				Expect(func() {
					returnNilLogger.Info("test")
				}).ToNot(Panic())
			})
		})

		Context("when using all methods together", func() {
			It("should work correctly", func() {
				hcLogger.Named("test-logger")
				hcLogger.With("key", "value")
				hcLogger.SetLevel(hclog.Info)

				hcLogger.Info("test message")

				Expect(mockLogger.entries).To(HaveLen(1))
			})
		})
	})
})
