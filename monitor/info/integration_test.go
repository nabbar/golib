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

package info_test

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/monitor/info"
)

var _ = Describe("Info Integration and Real-World Usage", func() {
	Describe("Standard interface compatibility", func() {
		Context("with encoding.TextMarshaler", func() {
			It("should be usable in contexts requiring TextMarshaler", func() {
				i, err := info.New("test-service")
				Expect(err).NotTo(HaveOccurred())

				// Verify it implements encoding.TextMarshaler
				func(tm interface{ MarshalText() ([]byte, error) }) {
					text, err := tm.MarshalText()
					Expect(err).NotTo(HaveOccurred())
					Expect(text).NotTo(BeEmpty())
				}(i)
			})
		})

		Context("with json.Marshaler", func() {
			It("should be usable in contexts requiring json.Marshaler", func() {
				i, err := info.New("test-service")
				Expect(err).NotTo(HaveOccurred())

				// Verify it implements json.Marshaler
				func(jm json.Marshaler) {
					data, err := jm.MarshalJSON()
					Expect(err).NotTo(HaveOccurred())
					Expect(data).NotTo(BeEmpty())
				}(i)
			})

			It("should work with json.Marshal directly", func() {
				i, err := info.New("test-service")
				Expect(err).NotTo(HaveOccurred())

				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"version": "1.0.0",
						"env":     "test",
					}, nil
				})

				data, err := json.Marshal(i)
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result["Name"]).To(Equal("test-service"))
			})

			It("should work in complex JSON structures", func() {
				i, err := info.New("test-service")
				Expect(err).NotTo(HaveOccurred())

				type ComplexStruct struct {
					ServiceInfo info.Info `json:"service_info"`
					Metadata    string    `json:"metadata"`
				}

				complex := ComplexStruct{
					ServiceInfo: i,
					Metadata:    "test-metadata",
				}

				data, err := json.Marshal(complex)
				Expect(err).NotTo(HaveOccurred())
				Expect(json.Valid(data)).To(BeTrue())
			})
		})
	})

	Describe("Real-world usage patterns", func() {
		Context("with service discovery pattern", func() {
			It("should provide service metadata for discovery", func() {
				i, err := info.New("api-gateway")
				Expect(err).NotTo(HaveOccurred())

				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"version":  "1.2.3",
						"port":     8080,
						"protocol": "http",
						"health":   "/health",
						"ready":    "/ready",
						"tags":     []string{"api", "gateway", "production"},
					}, nil
				})

				result := i.Info()
				Expect(result["version"]).To(Equal("1.2.3"))
				Expect(result["port"]).To(Equal(8080))
				Expect(result["tags"]).To(HaveLen(3))
			})
		})

		Context("with dynamic runtime information", func() {
			It("should capture runtime metrics", func() {
				i, err := info.New("runtime-service")
				Expect(err).NotTo(HaveOccurred())

				i.RegisterInfo(func() (map[string]interface{}, error) {
					var m runtime.MemStats
					runtime.ReadMemStats(&m)

					return map[string]interface{}{
						"goroutines":  runtime.NumGoroutine(),
						"alloc_bytes": m.Alloc,
						"sys_bytes":   m.Sys,
						"num_gc":      m.NumGC,
					}, nil
				})

				result := i.Info()
				Expect(result).To(HaveKey("goroutines"))
				Expect(result).To(HaveKey("alloc_bytes"))
			})

			It("should capture environment information", func() {
				i, err := info.New("env-service")
				Expect(err).NotTo(HaveOccurred())

				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"hostname": getHostname(),
						"os":       runtime.GOOS,
						"arch":     runtime.GOARCH,
						"go":       runtime.Version(),
						"cpus":     runtime.NumCPU(),
					}, nil
				})

				result := i.Info()
				Expect(result["os"]).To(Equal(runtime.GOOS))
				Expect(result["arch"]).To(Equal(runtime.GOARCH))
			})
		})

		Context("with health check integration", func() {
			It("should provide health status information", func() {
				i, err := info.New("health-monitor")
				Expect(err).NotTo(HaveOccurred())

				healthy := true
				i.RegisterInfo(func() (map[string]interface{}, error) {
					status := "healthy"
					if !healthy {
						status = "unhealthy"
					}

					return map[string]interface{}{
						"status":      status,
						"uptime":      "1h23m45s",
						"checks":      3,
						"checks_ok":   2,
						"checks_fail": 1,
					}, nil
				})

				result := i.Info()
				Expect(result["status"]).To(Equal("healthy"))
			})
		})

		Context("with configuration management", func() {
			It("should expose configuration metadata", func() {
				i, err := info.New("config-service")
				Expect(err).NotTo(HaveOccurred())

				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{
						"config_file":    "/etc/service/config.yaml",
						"config_version": "v2",
						"reload_count":   5,
						"last_reload":    "2024-01-01T12:00:00Z",
					}, nil
				})

				result := i.Info()
				Expect(result).To(HaveKey("config_file"))
				Expect(result).To(HaveKey("reload_count"))
			})
		})
	})

	Describe("Error handling patterns", func() {
		Context("with transient errors", func() {
			It("should handle function errors gracefully", func() {
				i, err := info.New("error-test")
				Expect(err).NotTo(HaveOccurred())

				callCount := 0
				i.RegisterInfo(func() (map[string]interface{}, error) {
					callCount++
					if callCount == 1 {
						return nil, fmt.Errorf("transient error")
					}
					return map[string]interface{}{"success": true}, nil
				})

				// First call fails
				result1 := i.Info()
				Expect(result1).To(BeNil())

				// Need to re-register for retry
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"success": true}, nil
				})

				result2 := i.Info()
				Expect(result2).NotTo(BeNil())
			})
		})

		Context("with partial data availability", func() {
			It("should handle partial info gracefully", func() {
				i, err := info.New("partial-service")
				Expect(err).NotTo(HaveOccurred())

				i.RegisterInfo(func() (map[string]interface{}, error) {
					// Simulate partial data availability
					data := map[string]interface{}{
						"available": "yes",
					}
					// Some fields might be missing
					return data, nil
				})

				result := i.Info()
				Expect(result).To(HaveKey("available"))
			})
		})
	})

	Describe("Performance characteristics", func() {
		Context("with caching behavior", func() {
			It("should demonstrate caching efficiency", func() {
				i, err := info.New("cache-test")
				Expect(err).NotTo(HaveOccurred())

				callCount := 0
				i.RegisterName(func() (string, error) {
					callCount++
					return fmt.Sprintf("name-%d", callCount), nil
				})

				// First call executes function
				name1 := i.Name()
				Expect(callCount).To(Equal(1))

				// Subsequent calls use cache
				name2 := i.Name()
				Expect(callCount).To(Equal(1))
				Expect(name1).To(Equal(name2))
			})

			It("should invalidate cache on re-registration", func() {
				i, err := info.New("invalidate-test")
				Expect(err).NotTo(HaveOccurred())

				i.RegisterName(func() (string, error) {
					return "first", nil
				})

				name1 := i.Name()
				Expect(name1).To(Equal("first"))

				// Re-register invalidates cache
				i.RegisterName(func() (string, error) {
					return "second", nil
				})

				name2 := i.Name()
				Expect(name2).To(Equal("second"))
			})
		})
	})
})

// getHostname returns the system hostname or "unknown" if it fails.
func getHostname() string {
	if name, err := os.Hostname(); err == nil {
		return name
	}
	return "unknown"
}
