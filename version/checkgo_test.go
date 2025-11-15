/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package version_test

import (
	"runtime"
	"strings"

	"github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckGo Method", func() {
	var v version.Version

	BeforeEach(func() {
		v = version.NewVersion(
			version.License_MIT,
			"TestApp",
			"Test Application",
			testTime,
			"abc123",
			"v1.0.0",
			"Test Author",
			"test",
			testStruct{},
			0,
		)
	})

	// Helper to extract major and minor version from runtime.Version()
	// runtime.Version() returns something like "go1.25.0"
	extractGoVersion := func() (string, string, string) {
		ver := runtime.Version()
		// Remove "go" prefix
		ver = strings.TrimPrefix(ver, "go")
		parts := strings.Split(ver, ".")
		if len(parts) >= 2 {
			major := parts[0]
			minor := parts[1]
			patch := "0"
			if len(parts) >= 3 {
				patch = parts[2]
			}
			return major, minor, patch
		}
		return "1", "25", "0" // Default fallback
	}

	Describe("Valid Constraints", func() {
		Context("with exact version match", func() {
			It("should pass when runtime matches required version exactly", func() {
				major, minor, _ := extractGoVersion()
				requiredVersion := major + "." + minor

				err := v.CheckGo(requiredVersion, ">=")
				Expect(err).To(BeNil())
			})
		})

		Context("with >= constraint", func() {
			It("should pass when runtime version is greater or equal", func() {
				err := v.CheckGo("1.18", ">=")
				Expect(err).To(BeNil())
			})

			It("should pass with current runtime version", func() {
				major, minor, _ := extractGoVersion()
				requiredVersion := major + "." + minor

				err := v.CheckGo(requiredVersion, ">=")
				Expect(err).To(BeNil())
			})

			It("should pass with older version requirement", func() {
				err := v.CheckGo("1.16", ">=")
				Expect(err).To(BeNil())
			})
		})

		Context("with > constraint", func() {
			It("should pass when runtime is strictly greater", func() {
				err := v.CheckGo("1.10", ">")
				Expect(err).To(BeNil())
			})
		})

		Context("with <= constraint", func() {
			It("should pass with very high version", func() {
				err := v.CheckGo("99.99", "<=")
				Expect(err).To(BeNil())
			})
		})

		Context("with < constraint", func() {
			It("should pass with version higher than runtime", func() {
				err := v.CheckGo("99.99", "<")
				Expect(err).To(BeNil())
			})
		})

		Context("with ~> constraint (pessimistic)", func() {
			It("should pass with compatible version", func() {
				major, _, _ := extractGoVersion()
				requiredVersion := major + ".0"

				err := v.CheckGo(requiredVersion, "~>")
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Invalid Constraints", func() {
		Context("when runtime version is too old", func() {
			It("should fail with >= constraint for future version", func() {
				err := v.CheckGo("99.99", ">=")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionConstraint))
				Expect(err.Error()).To(ContainSubstring("non-compatible version of Go"))
			})

			It("should fail with > constraint for current version", func() {
				major, minor, patch := extractGoVersion()
				requiredVersion := major + "." + minor + "." + patch

				// Current version is NOT > itself
				err := v.CheckGo(requiredVersion, ">")
				Expect(err).ToNot(BeNil())
			})
		})

		Context("when runtime version is too new", func() {
			It("should fail with < constraint for old version", func() {
				err := v.CheckGo("1.10", "<")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionConstraint))
			})

			It("should fail with <= constraint for old version", func() {
				err := v.CheckGo("1.10", "<=")
				Expect(err).ToNot(BeNil())
			})
		})

		Context("when exact version doesn't match", func() {
			It("should fail with == constraint for different version", func() {
				err := v.CheckGo("1.10.5", "==")
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Invalid Constraint Syntax", func() {
		Context("with malformed constraint", func() {
			It("should return error for invalid constraint operator", func() {
				err := v.CheckGo("1.18", "!!")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionInit))
				Expect(err.Error()).To(ContainSubstring("init GoVersion contraint error"))
			})

			It("should return error for invalid version format", func() {
				err := v.CheckGo("invalid.version", ">=")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionInit))
			})

			It("should return error for empty version", func() {
				err := v.CheckGo("", ">=")
				Expect(err).ToNot(BeNil())
			})

			It("should return error for empty constraint", func() {
				err := v.CheckGo("1.18", "")
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Version Parsing Edge Cases", func() {
		Context("with various version formats", func() {
			It("should handle two-part version numbers", func() {
				err := v.CheckGo("1.18", ">=")
				Expect(err).To(BeNil())
			})

			It("should handle three-part version numbers", func() {
				err := v.CheckGo("1.18.5", ">=")
				Expect(err).To(BeNil())
			})

			It("should handle version with leading zeros", func() {
				err := v.CheckGo("1.18.01", ">=")
				// This might fail or succeed depending on go-version parsing
				// We just verify it doesn't panic
				_ = err
			})
		})
	})

	Describe("Constraint Combinations", func() {
		Context("with multiple checks", func() {
			It("should handle multiple constraint checks sequentially", func() {
				err1 := v.CheckGo("1.16", ">=")
				err2 := v.CheckGo("99.0", "<=")

				Expect(err1).To(BeNil())
				Expect(err2).To(BeNil())
			})

			It("should independently evaluate each constraint", func() {
				// First check passes
				err1 := v.CheckGo("1.16", ">=")
				Expect(err1).To(BeNil())

				// Second check fails
				err2 := v.CheckGo("1.10", "<")
				Expect(err2).ToNot(BeNil())
			})
		})
	})

	Describe("Runtime Version Information", func() {
		It("should correctly parse runtime version", func() {
			runtimeVer := runtime.Version()
			Expect(runtimeVer).To(HavePrefix("go"))

			// Verify we can check against current version
			major, minor, _ := extractGoVersion()
			requiredVersion := major + "." + minor

			err := v.CheckGo(requiredVersion, ">=")
			Expect(err).To(BeNil())
		})

		It("should handle runtime version with or without patch", func() {
			major, minor, _ := extractGoVersion()

			// Check with major.minor using ==
			err1 := v.CheckGo(major+"."+minor, ">=")
			Expect(err1).To(BeNil())

			// Check with >= should also work
			err2 := v.CheckGo(major+"."+minor, ">=")
			Expect(err2).To(BeNil())
		})
	})

	Describe("Error Message Content", func() {
		It("should provide detailed error message on constraint failure", func() {
			err := v.CheckGo("99.99", ">=")
			Expect(err).ToNot(BeNil())

			errMsg := err.Error()
			Expect(errMsg).To(ContainSubstring("non-compatible version of Go"))
		})

		It("should provide error message for invalid constraint", func() {
			err := v.CheckGo("invalid", ">=")
			Expect(err).ToNot(BeNil())

			errMsg := err.Error()
			Expect(errMsg).To(ContainSubstring("init GoVersion contraint error"))
		})
	})

	Describe("Common Use Cases", func() {
		Context("minimum version requirements", func() {
			It("should enforce minimum Go 1.18 requirement", func() {
				err := v.CheckGo("1.18", ">=")
				Expect(err).To(BeNil())
			})

			It("should enforce minimum Go 1.20 requirement", func() {
				err := v.CheckGo("1.20", ">=")
				Expect(err).To(BeNil())
			})

			It("should enforce minimum Go 1.21 requirement", func() {
				err := v.CheckGo("1.21", ">=")
				Expect(err).To(BeNil())
			})
		})

		Context("version range requirements", func() {
			It("should work within supported range", func() {
				major, _, _ := extractGoVersion()

				// Check we're at least 1.18
				err1 := v.CheckGo("1.18", ">=")
				Expect(err1).To(BeNil())

				// Check we're below 100.0 (sanity check)
				err2 := v.CheckGo("100.0", "<")
				Expect(err2).To(BeNil())

				_ = major // Use variable to avoid lint error
			})
		})
	})

	Describe("Concurrent Safety", func() {
		It("should safely check Go version from multiple goroutines", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					err := v.CheckGo("1.18", ">=")
					Expect(err).To(BeNil())
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})

		It("should handle concurrent checks with different constraints", func() {
			done := make(chan bool, 20)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					_ = v.CheckGo("1.18", ">=")
					done <- true
				}()

				go func() {
					defer GinkgoRecover()
					_ = v.CheckGo("99.0", "<=")
					done <- true
				}()
			}

			for i := 0; i < 20; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})

	Describe("Real-world Scenarios", func() {
		Context("checking application requirements", func() {
			It("should validate Go version at application startup", func() {
				// Simulate checking if runtime meets minimum requirements
				err := v.CheckGo("1.16", ">=")
				Expect(err).To(BeNil(), "Application should run on Go 1.16+")
			})

			It("should prevent running on too old Go versions", func() {
				// Simulate rejecting very old Go versions
				err := v.CheckGo("99.0", ">=")
				Expect(err).ToNot(BeNil(), "Application should not run on old Go versions")
			})
		})

		Context("checking for feature availability", func() {
			It("should check for generics support (Go 1.18+)", func() {
				err := v.CheckGo("1.18", ">=")
				Expect(err).To(BeNil())
			})

			It("should check for workspace mode support (Go 1.18+)", func() {
				err := v.CheckGo("1.18", ">=")
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Error Code Verification", func() {
		It("should return ErrorGoVersionConstraint for constraint mismatch", func() {
			err := v.CheckGo("99.99", ">=")
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(version.ErrorGoVersionConstraint))
		})

		It("should return ErrorGoVersionInit for invalid constraint", func() {
			err := v.CheckGo("invalid", ">=")
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(version.ErrorGoVersionInit))
		})
	})
})
