/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package retro

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	Describe("isValidVersion", func() {
		Context("when validating correct semantic versions", func() {
			It("should accept valid version without operator", func() {
				Expect(isValidVersion("v1.0.0")).To(BeTrue())
			})

			It("should accept version with >= operator", func() {
				Expect(isValidVersion(">=v1.0.0")).To(BeTrue())
			})

			It("should accept version with <= operator", func() {
				Expect(isValidVersion("<=v2.3.4")).To(BeTrue())
			})

			It("should accept version with > operator", func() {
				Expect(isValidVersion(">v0.0.1")).To(BeTrue())
			})

			It("should accept version with < operator", func() {
				Expect(isValidVersion("<v10.20.30")).To(BeTrue())
			})

			It("should accept default version", func() {
				Expect(isValidVersion("default")).To(BeTrue())
			})
		})

		Context("when validating incorrect versions", func() {
			It("should reject version without v prefix", func() {
				Expect(isValidVersion("1.0.0")).To(BeFalse())
			})

			It("should reject version with only two parts", func() {
				Expect(isValidVersion("v1.0")).To(BeFalse())
			})

			It("should reject version with four parts", func() {
				Expect(isValidVersion("v1.0.0.0")).To(BeFalse())
			})

			It("should reject empty version", func() {
				Expect(isValidVersion("")).To(BeFalse())
			})

			It("should reject version with letters", func() {
				Expect(isValidVersion("v1.0.a")).To(BeFalse())
			})

			It("should reject invalid operator", func() {
				Expect(isValidVersion("!=v1.0.0")).To(BeFalse())
			})
		})
	})

	Describe("compareVersions", func() {
		Context("when comparing equal versions", func() {
			It("should return 0 for identical versions", func() {
				Expect(compareVersions("1.0.0", "1.0.0")).To(Equal(0))
			})

			It("should return 0 for equal versions with leading zeros", func() {
				Expect(compareVersions("1.0.0", "1.00.0")).To(Equal(0))
			})
		})

		Context("when first version is greater", func() {
			It("should return 1 when major version is greater", func() {
				Expect(compareVersions("2.0.0", "1.0.0")).To(Equal(1))
			})

			It("should return 1 when minor version is greater", func() {
				Expect(compareVersions("1.2.0", "1.1.0")).To(Equal(1))
			})

			It("should return 1 when patch version is greater", func() {
				Expect(compareVersions("1.0.5", "1.0.3")).To(Equal(1))
			})
		})

		Context("when first version is lesser", func() {
			It("should return -1 when major version is lesser", func() {
				Expect(compareVersions("1.0.0", "2.0.0")).To(Equal(-1))
			})

			It("should return -1 when minor version is lesser", func() {
				Expect(compareVersions("1.1.0", "1.2.0")).To(Equal(-1))
			})

			It("should return -1 when patch version is lesser", func() {
				Expect(compareVersions("1.0.3", "1.0.5")).To(Equal(-1))
			})
		})
	})

	Describe("checkCondition", func() {
		Context("when using > operator", func() {
			It("should return true when version is greater", func() {
				Expect(checkCondition("v1.0.1", "v1.0.0", ">")).To(BeTrue())
			})

			It("should return false when version is equal", func() {
				Expect(checkCondition("v1.0.0", "v1.0.0", ">")).To(BeFalse())
			})

			It("should return false when version is lesser", func() {
				Expect(checkCondition("v0.9.0", "v1.0.0", ">")).To(BeFalse())
			})
		})

		Context("when using >= operator", func() {
			It("should return true when version is greater", func() {
				Expect(checkCondition("v2.0.0", "v1.0.0", ">=")).To(BeTrue())
			})

			It("should return true when version is equal", func() {
				Expect(checkCondition("v1.0.0", "v1.0.0", ">=")).To(BeTrue())
			})

			It("should return false when version is lesser", func() {
				Expect(checkCondition("v0.9.0", "v1.0.0", ">=")).To(BeFalse())
			})
		})

		Context("when using < operator", func() {
			It("should return true when version is lesser", func() {
				Expect(checkCondition("v0.9.0", "v1.0.0", "<")).To(BeTrue())
			})

			It("should return false when version is equal", func() {
				Expect(checkCondition("v1.0.0", "v1.0.0", "<")).To(BeFalse())
			})

			It("should return false when version is greater", func() {
				Expect(checkCondition("v2.0.0", "v1.0.0", "<")).To(BeFalse())
			})
		})

		Context("when using <= operator", func() {
			It("should return true when version is lesser", func() {
				Expect(checkCondition("v0.9.0", "v1.0.0", "<=")).To(BeTrue())
			})

			It("should return true when version is equal", func() {
				Expect(checkCondition("v1.0.0", "v1.0.0", "<=")).To(BeTrue())
			})

			It("should return false when version is greater", func() {
				Expect(checkCondition("v2.0.0", "v1.0.0", "<=")).To(BeFalse())
			})
		})

		Context("when using == operator", func() {
			It("should return true when versions are equal", func() {
				Expect(checkCondition("v1.0.0", "v1.0.0", "==")).To(BeTrue())
			})

			It("should return false when versions are different", func() {
				Expect(checkCondition("v1.0.1", "v1.0.0", "==")).To(BeFalse())
			})
		})

		Context("when using invalid operator", func() {
			It("should return false", func() {
				Expect(checkCondition("v1.0.0", "v1.0.0", "!=")).To(BeFalse())
			})
		})
	})

	Describe("parseOperator", func() {
		Context("when parsing version with operators", func() {
			It("should parse >= operator", func() {
				op, ver := parseOperator(">=v1.0.0")
				Expect(op).To(Equal(">="))
				Expect(ver).To(Equal("v1.0.0"))
			})

			It("should parse <= operator", func() {
				op, ver := parseOperator("<=v2.0.0")
				Expect(op).To(Equal("<="))
				Expect(ver).To(Equal("v2.0.0"))
			})

			It("should parse > operator", func() {
				op, ver := parseOperator(">v1.5.0")
				Expect(op).To(Equal(">"))
				Expect(ver).To(Equal("v1.5.0"))
			})

			It("should parse < operator", func() {
				op, ver := parseOperator("<v3.0.0")
				Expect(op).To(Equal("<"))
				Expect(ver).To(Equal("v3.0.0"))
			})

			It("should handle version without operator", func() {
				op, ver := parseOperator("v1.0.0")
				Expect(op).To(BeEmpty())
				Expect(ver).To(Equal("v1.0.0"))
			})
		})
	})

	Describe("validRetroTag", func() {
		Context("when validating correct retro tags", func() {
			It("should accept single version", func() {
				Expect(validRetroTag([]string{"v1.0.0"})).To(BeTrue())
			})

			It("should accept multiple versions", func() {
				Expect(validRetroTag([]string{"v1.0.0", "v2.0.0"})).To(BeTrue())
			})

			It("should accept default tag", func() {
				Expect(validRetroTag([]string{"default"})).To(BeTrue())
			})

			It("should accept version with operators", func() {
				Expect(validRetroTag([]string{">v1.0.0", "<v2.0.0"})).To(BeTrue())
			})

			It("should accept mixed versions", func() {
				Expect(validRetroTag([]string{"default", "v1.0.0", ">v2.0.0"})).To(BeTrue())
			})
		})

		Context("when validating incorrect retro tags", func() {
			It("should reject duplicate > operators", func() {
				Expect(validRetroTag([]string{">v1.0.0", ">v2.0.0"})).To(BeFalse())
			})

			It("should reject duplicate < operators", func() {
				Expect(validRetroTag([]string{"<v1.0.0", "<v2.0.0"})).To(BeFalse())
			})

			It("should reject invalid version format", func() {
				Expect(validRetroTag([]string{"1.0.0"})).To(BeFalse())
			})

			It("should reject mixed valid and invalid versions", func() {
				Expect(validRetroTag([]string{"v1.0.0", "invalid"})).To(BeFalse())
			})
		})
	})

	Describe("detectedBoundaries", func() {
		Context("when detecting dual boundaries", func() {
			It("should detect both > and < operators", func() {
				Expect(detectedBoundaries([]string{">v1.0.0", "<v2.0.0"})).To(BeTrue())
			})

			It("should detect >= and <= operators", func() {
				Expect(detectedBoundaries([]string{">=v1.0.0", "<=v2.0.0"})).To(BeTrue())
			})

			It("should not detect only > operator", func() {
				Expect(detectedBoundaries([]string{">v1.0.0"})).To(BeFalse())
			})

			It("should not detect only < operator", func() {
				Expect(detectedBoundaries([]string{"<v2.0.0"})).To(BeFalse())
			})

			It("should not detect standalone versions", func() {
				Expect(detectedBoundaries([]string{"v1.0.0", "v2.0.0"})).To(BeFalse())
			})

			It("should detect boundaries with additional standalone versions", func() {
				Expect(detectedBoundaries([]string{">v1.0.0", "<v2.0.0", "v1.5.0"})).To(BeTrue())
			})
		})
	})

	Describe("isVersionSupported", func() {
		Context("when checking empty retro tag", func() {
			It("should always return true for empty retro tag", func() {
				Expect(isVersionSupported("v1.0.0", "")).To(BeTrue())
				Expect(isVersionSupported("v2.0.0", "")).To(BeTrue())
				Expect(isVersionSupported("default", "")).To(BeTrue())
			})
		})

		Context("when checking default version", func() {
			It("should return true when retro tag contains default", func() {
				Expect(isVersionSupported("default", "default")).To(BeTrue())
			})

			It("should return true when retro tag has default with others", func() {
				Expect(isVersionSupported("default", "default,v1.0.0")).To(BeTrue())
			})

			It("should return false when retro tag does not contain default", func() {
				Expect(isVersionSupported("default", "v1.0.0")).To(BeFalse())
			})
		})

		Context("when checking standalone versions", func() {
			It("should match exact version", func() {
				Expect(isVersionSupported("v1.0.0", "v1.0.0")).To(BeTrue())
			})

			It("should not match different version", func() {
				Expect(isVersionSupported("v1.0.1", "v1.0.0")).To(BeFalse())
			})

			It("should match one of multiple standalone versions", func() {
				Expect(isVersionSupported("v1.0.3", "v1.0.0,v1.0.3")).To(BeTrue())
			})
		})

		Context("when checking greater than operator", func() {
			It("should return true for version greater than constraint", func() {
				Expect(isVersionSupported("v1.0.1", ">v1.0.0")).To(BeTrue())
			})

			It("should return false for version equal to constraint", func() {
				Expect(isVersionSupported("v1.0.0", ">v1.0.0")).To(BeFalse())
			})

			It("should return false for version less than constraint", func() {
				Expect(isVersionSupported("v0.9.0", ">v1.0.0")).To(BeFalse())
			})
		})

		Context("when checking greater than or equal operator", func() {
			It("should return true for version greater than constraint", func() {
				Expect(isVersionSupported("v1.0.1", ">=v1.0.0")).To(BeTrue())
			})

			It("should return true for version equal to constraint", func() {
				Expect(isVersionSupported("v1.0.0", ">=v1.0.0")).To(BeTrue())
			})

			It("should return false for version less than constraint", func() {
				Expect(isVersionSupported("v0.9.0", ">=v1.0.0")).To(BeFalse())
			})
		})

		Context("when checking less than operator", func() {
			It("should return true for version less than constraint", func() {
				Expect(isVersionSupported("v0.9.0", "<v1.0.0")).To(BeTrue())
			})

			It("should return false for version equal to constraint", func() {
				Expect(isVersionSupported("v1.0.0", "<v1.0.0")).To(BeFalse())
			})

			It("should return false for version greater than constraint", func() {
				Expect(isVersionSupported("v1.0.1", "<v1.0.0")).To(BeFalse())
			})
		})

		Context("when checking less than or equal operator", func() {
			It("should return true for version less than constraint", func() {
				Expect(isVersionSupported("v0.9.0", "<=v1.0.0")).To(BeTrue())
			})

			It("should return true for version equal to constraint", func() {
				Expect(isVersionSupported("v1.0.0", "<=v1.0.0")).To(BeTrue())
			})

			It("should return false for version greater than constraint", func() {
				Expect(isVersionSupported("v1.0.1", "<=v1.0.0")).To(BeFalse())
			})
		})

		Context("when checking dual boundaries", func() {
			It("should return true for version within range", func() {
				Expect(isVersionSupported("v1.0.2", ">v1.0.0,<v1.0.4")).To(BeTrue())
			})

			It("should return false for version below range", func() {
				Expect(isVersionSupported("v1.0.0", ">v1.0.0,<v1.0.4")).To(BeFalse())
			})

			It("should return false for version above range", func() {
				Expect(isVersionSupported("v1.0.4", ">v1.0.0,<v1.0.4")).To(BeFalse())
			})

			It("should handle inclusive boundaries", func() {
				Expect(isVersionSupported("v1.0.0", ">=v1.0.0,<=v1.0.2")).To(BeTrue())
				Expect(isVersionSupported("v1.0.2", ">=v1.0.0,<=v1.0.2")).To(BeTrue())
			})
		})

		Context("when checking standalone exceptions with operators", func() {
			It("should return true for exception version outside range", func() {
				Expect(isVersionSupported("v0.0.3", ">v1.0.0,v0.0.3")).To(BeTrue())
			})

			It("should return true for version within range", func() {
				Expect(isVersionSupported("v1.0.1", ">v1.0.0,v0.0.3")).To(BeTrue())
			})

			It("should return false for version outside both range and exception", func() {
				Expect(isVersionSupported("v0.9.0", ">v1.0.0,v0.0.3")).To(BeFalse())
			})
		})

		Context("when checking invalid retro tags", func() {
			It("should return false for conflicting operators", func() {
				Expect(isVersionSupported("v1.0.2", ">v1.0.0,>v1.0.3")).To(BeFalse())
			})

			It("should return false for invalid version format", func() {
				Expect(isVersionSupported("v1.0.0", "1.0.0")).To(BeFalse())
			})
		})

		Context("when checking complex real-world scenarios", func() {
			It("should handle lastName field constraint: v1.0.0,>v1.0.3", func() {
				Expect(isVersionSupported("v1.0.0", "v1.0.0,>v1.0.3")).To(BeTrue())
				Expect(isVersionSupported("v1.0.1", "v1.0.0,>v1.0.3")).To(BeFalse())
				Expect(isVersionSupported("v1.0.3", "v1.0.0,>v1.0.3")).To(BeFalse())
				Expect(isVersionSupported("v1.0.4", "v1.0.0,>v1.0.3")).To(BeTrue())
			})

			It("should handle status field constraint: default,>v1.0.1,<v1.0.4", func() {
				Expect(isVersionSupported("default", "default,>v1.0.1,<v1.0.4")).To(BeTrue())
				Expect(isVersionSupported("v1.0.2", "default,>v1.0.1,<v1.0.4")).To(BeTrue())
				Expect(isVersionSupported("v1.0.3", "default,>v1.0.1,<v1.0.4")).To(BeTrue())
				Expect(isVersionSupported("v1.0.1", "default,>v1.0.1,<v1.0.4")).To(BeFalse())
				Expect(isVersionSupported("v1.0.4", "default,>v1.0.1,<v1.0.4")).To(BeFalse())
			})

			It("should handle sex field constraint with exception: >v1.0.0,<=v1.0.3,v0.0.3,default", func() {
				Expect(isVersionSupported("default", ">v1.0.0,<=v1.0.3,v0.0.3,default")).To(BeTrue())
				Expect(isVersionSupported("v0.0.3", ">v1.0.0,<=v1.0.3,v0.0.3,default")).To(BeTrue())
				Expect(isVersionSupported("v1.0.1", ">v1.0.0,<=v1.0.3,v0.0.3,default")).To(BeTrue())
				Expect(isVersionSupported("v1.0.0", ">v1.0.0,<=v1.0.3,v0.0.3,default")).To(BeFalse())
				Expect(isVersionSupported("v1.0.4", ">v1.0.0,<=v1.0.3,v0.0.3,default")).To(BeFalse())
			})
		})
	})
})
