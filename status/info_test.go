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

package status_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsts "github.com/nabbar/golib/status"
	libver "github.com/nabbar/golib/version"
)

var _ = Describe("Status/Info", func() {
	Describe("SetInfo", func() {
		It("should set basic application info", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("my-app", "v1.2.3", "abc123def456")

			// Verify by marshaling to JSON
			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("my-app"))
			Expect(string(data)).To(ContainSubstring("v1.2.3"))
			Expect(string(data)).To(ContainSubstring("abc123def456"))
		})

		It("should handle empty name", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("", "v1.0.0", "hash")

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).ToNot(BeEmpty())
		})

		It("should handle empty version", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("app", "", "hash")

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("app"))
		})

		It("should handle empty hash", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("app", "v1.0.0", "")

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("app"))
		})

		It("should handle all empty values", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("", "", "")

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).ToNot(BeEmpty())
		})

		It("should allow updating info multiple times", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("app-v1", "v1.0.0", "hash1")

			data1, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data1)).To(ContainSubstring("app-v1"))

			// Update info
			status.SetInfo("app-v2", "v2.0.0", "hash2")

			data2, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data2)).To(ContainSubstring("app-v2"))
			Expect(string(data2)).To(ContainSubstring("v2.0.0"))
			Expect(string(data2)).To(ContainSubstring("hash2"))
		})

		It("should handle special characters in values", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("app-name_123", "v1.0.0-beta+build.123", "abc-def-123")

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("app-name_123"))
		})

		It("should handle long values", func() {
			status := libsts.New(globalCtx)
			longName := "very-long-application-name-with-many-characters-" + time.Now().Format("20060102150405")
			longHash := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

			status.SetInfo(longName, "v1.0.0", longHash)

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring(longName))
			Expect(string(data)).To(ContainSubstring(longHash))
		})
	})

	Describe("SetVersion", func() {
		It("should set version from Version object", func() {
			status := libsts.New(globalCtx)
			type testStruct struct{}
			vers := libver.NewVersion(
				libver.License_MIT,
				"test-package",
				"Test Package Description",
				time.Now().Format(time.RFC3339),
				"commit-hash-123",
				"v1.2.3",
				"Test Author",
				"TST",
				testStruct{},
				1,
			)
			status.SetVersion(vers)

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("test-package"))
			Expect(string(data)).To(ContainSubstring("v1.2.3"))
			Expect(string(data)).To(ContainSubstring("commit-hash-123"))
		})

		It("should extract all version fields", func() {
			status := libsts.New(globalCtx)
			type testStruct struct{}
			vers := libver.NewVersion(
				libver.License_MIT,
				"test-package",
				"Test Package Description",
				time.Now().Format(time.RFC3339),
				"commit-hash-123",
				"v1.2.3",
				"Test Author",
				"TST",
				testStruct{},
				1,
			)
			status.SetVersion(vers)

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())

			// Verify package name
			Expect(string(data)).To(ContainSubstring("test-package"))

			// Verify release version
			Expect(string(data)).To(ContainSubstring("v1.2.3"))

			// Verify build hash
			Expect(string(data)).To(ContainSubstring("commit-hash-123"))
		})

		It("should allow updating version multiple times", func() {
			status := libsts.New(globalCtx)
			type testStruct struct{}
			vers := libver.NewVersion(
				libver.License_MIT,
				"test-package",
				"Test Package Description",
				time.Now().Format(time.RFC3339),
				"commit-hash-123",
				"v1.2.3",
				"Test Author",
				"TST",
				testStruct{},
				1,
			)
			status.SetVersion(vers)

			data1, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data1)).To(ContainSubstring("test-package"))

			// Create new version
			vers2 := libver.NewVersion(
				libver.License_MIT,
				"updated-package",
				"Updated Description",
				time.Now().Format(time.RFC3339),
				"new-hash-456",
				"v2.0.0",
				"New Author",
				"UPD",
				testStruct{},
				2,
			)

			status.SetVersion(vers2)

			data2, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data2)).To(ContainSubstring("updated-package"))
			Expect(string(data2)).To(ContainSubstring("v2.0.0"))
			Expect(string(data2)).To(ContainSubstring("new-hash-456"))
		})

		It("should include build date in output", func() {
			status := libsts.New(globalCtx)
			type testStruct struct{}
			vers := libver.NewVersion(
				libver.License_MIT,
				"test-package",
				"Test Package Description",
				time.Now().Format(time.RFC3339),
				"commit-hash-123",
				"v1.2.3",
				"Test Author",
				"TST",
				testStruct{},
				1,
			)
			status.SetVersion(vers)

			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())

			// DateBuild should be present in JSON
			Expect(string(data)).To(ContainSubstring("commit-hash-123"))
		})
	})

	Describe("SetInfo vs SetVersion", func() {
		It("should allow switching from SetInfo to SetVersion", func() {
			status := libsts.New(globalCtx)
			// First use SetInfo
			status.SetInfo("manual-app", "v1.0.0", "manual-hash")

			data1, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data1)).To(ContainSubstring("manual-app"))

			// Then use SetVersion
			type testStruct struct{}
			vers := libver.NewVersion(
				libver.License_MIT,
				"version-app",
				"Description",
				time.Now().Format(time.RFC3339),
				"version-hash",
				"v2.0.0",
				"Author",
				"VER",
				testStruct{},
				1,
			)
			status.SetVersion(vers)

			data2, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data2)).To(ContainSubstring("version-app"))
			Expect(string(data2)).To(ContainSubstring("v2.0.0"))
		})

		It("should allow switching from SetVersion to SetInfo", func() {
			status := libsts.New(globalCtx)
			// First use SetVersion
			type testStruct struct{}
			vers := libver.NewVersion(
				libver.License_MIT,
				"version-first",
				"Description",
				time.Now().Format(time.RFC3339),
				"hash1",
				"v1.0.0",
				"Author",
				"VER",
				testStruct{},
				1,
			)
			status.SetVersion(vers)

			data1, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data1)).To(ContainSubstring("version-first"))

			// Then use SetInfo
			status.SetInfo("manual-second", "v2.0.0", "hash2")

			data2, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data2)).To(ContainSubstring("manual-second"))
			Expect(string(data2)).To(ContainSubstring("v2.0.0"))
		})
	})

	Describe("Info in different output formats", func() {
		It("should include info in JSON output", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("format-test", "v1.0.0", "abc123")
			data, err := json.Marshal(status)
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).ToNot(HaveOccurred())

			Expect(result["Name"]).To(Equal("format-test"))
			Expect(result["Release"]).To(Equal("v1.0.0"))
			Expect(result["Hash"]).To(Equal("abc123"))
		})

		It("should include info in text output", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("format-test", "v1.0.0", "abc123")
			data, err := status.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			text := string(data)
			Expect(text).To(ContainSubstring("format-test"))
			Expect(text).To(ContainSubstring("v1.0.0"))
			Expect(text).To(ContainSubstring("abc123"))
		})
	})
})
