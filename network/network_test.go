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

package network_test

import (
	. "github.com/nabbar/golib/network"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Network Package Integration", func() {
	Describe("Number and Bytes interoperability", func() {
		It("should convert between Number and Bytes seamlessly", func() {
			// Number -> Bytes -> Number
			n1 := Number(12345)
			b := n1.AsBytes()
			n2 := b.AsNumber()
			Expect(n2).To(Equal(n1))

			// Bytes -> Number -> Bytes
			b1 := Bytes(67890)
			n := b1.AsNumber()
			b2 := n.AsBytes()
			Expect(b2).To(Equal(b1))
		})

		It("should format same values differently", func() {
			// 5000 as Number vs Bytes
			n := Number(5000)
			b := Bytes(5000)

			// Number uses decimal (K = 10^3), so 5000 = 5K
			nResult := n.FormatUnitInt()
			Expect(nResult).To(ContainSubstring("K"))
			Expect(nResult).NotTo(ContainSubstring("KB"))

			// Bytes uses binary (KB = 2^10), so 5000 bytes = 4.88 KB
			bResult := b.FormatUnitInt()
			Expect(bResult).To(ContainSubstring("KB"))
		})
	})

	Describe("Stats type integration", func() {
		It("should format StatBytes using Bytes formatting", func() {
			n := Number(1024)

			// Direct bytes formatting
			bytesResult := n.AsBytes().FormatUnitInt()

			// Stats formatting for bytes
			statsResult := StatBytes.FormatUnitInt(n)

			Expect(statsResult).To(Equal(bytesResult))
			Expect(statsResult).To(ContainSubstring("KB"))
		})

		It("should format other stats using Number formatting", func() {
			n := Number(5000)

			// Direct number formatting
			numberResult := n.FormatUnitInt()

			// Stats formatting for packets
			statsResult := StatPackets.FormatUnitInt(n)

			Expect(statsResult).To(Equal(numberResult))
			Expect(statsResult).To(ContainSubstring("K"))
		})

		It("should provide consistent labeling", func() {
			n := Number(1000)

			stats := []Stats{
				StatBytes,
				StatPackets,
				StatFifo,
				StatDrop,
				StatErr,
			}

			for _, stat := range stats {
				labeled := stat.FormatLabelUnit(n)
				Expect(labeled).To(ContainSubstring(stat.String() + ":"))

				paddedLabeled := stat.FormatLabelUnitPadded(n)
				Expect(paddedLabeled).To(ContainSubstring(stat.String() + ":"))
			}
		})
	})

	Describe("Real-world usage scenarios", func() {
		Context("network traffic statistics", func() {
			It("should format network counters correctly", func() {
				// Simulate network interface statistics
				bytesReceived := Number(1073741824) // 1 GB
				packetsSent := Number(5000000)      // 5 M packets
				droppedPackets := Number(150)       // 150 packets
				errors := Number(5)                 // 5 errors

				// Format as would be displayed
				trafficStr := StatBytes.FormatLabelUnitPadded(bytesReceived)
				packetsStr := StatPackets.FormatLabelUnitPadded(packetsSent)
				dropStr := StatDrop.FormatLabelUnitPadded(droppedPackets)
				errStr := StatErr.FormatLabelUnitPadded(errors)

				// Verify formatting
				Expect(trafficStr).To(ContainSubstring("Traffic:"))
				Expect(trafficStr).To(ContainSubstring("GB"))

				Expect(packetsStr).To(ContainSubstring("Packets:"))
				Expect(packetsStr).To(ContainSubstring("M"))

				Expect(dropStr).To(ContainSubstring("Drop:"))
				Expect(errStr).To(ContainSubstring("Error:"))
			})
		})

		Context("bandwidth monitoring", func() {
			It("should format bandwidth with precision", func() {
				// 1.5 Mbps = 1,500,000 bits per second
				bandwidth := Number(1500000)
				result := bandwidth.FormatUnitFloat(2)

				Expect(result).To(ContainSubstring("M"))
				Expect(result).To(MatchRegexp(`1\.50`))
			})

			It("should format transfer speed", func() {
				// 10.5 MB/s = 10,485,760 bytes per second
				transferSpeed := Bytes(10485760)
				result := transferSpeed.FormatUnitFloat(1)

				Expect(result).To(ContainSubstring("MB"))
				Expect(result).To(MatchRegexp(`10\.\d`))
			})
		})

		Context("storage and memory", func() {
			It("should format file sizes correctly", func() {
				tests := map[Bytes]string{
					Bytes(1024):             "KB", // 1 KB
					Bytes(1048576):          "MB", // 1 MB
					Bytes(1073741824):       "GB", // 1 GB
					Bytes(1099511627776):    "TB", // 1 TB
					Bytes(1125899906842624): "PB", // 1 PB
				}

				for size, expectedUnit := range tests {
					result := size.FormatUnitInt()
					Expect(result).To(ContainSubstring(expectedUnit))
				}
			})
		})
	})

	Describe("Performance at scale", func() {
		It("should handle large-scale statistics efficiently", func() {
			// Simulate processing many network interfaces
			Expect(func() {
				for i := 0; i < 1000; i++ {
					// Simulate interface stats
					bytesIn := Number(uint64(i) * 1000000)
					bytesOut := Number(uint64(i) * 500000)
					packets := Number(uint64(i) * 10000)

					_ = StatBytes.FormatUnit(bytesIn)
					_ = StatBytes.FormatUnit(bytesOut)
					_ = StatPackets.FormatUnit(packets)
				}
			}).NotTo(Panic())
		})

		It("should handle concurrent access safely", func() {
			// Verify no race conditions
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(id int) {
					n := Number(uint64(id) * 1000000)
					b := Bytes(uint64(id) * 1024)

					for j := 0; j < 100; j++ {
						_ = n.FormatUnitInt()
						_ = b.FormatUnitFloat(2)
						_ = StatBytes.FormatUnit(n)
					}

					done <- true
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})

	Describe("Edge cases across types", func() {
		It("should handle zero across all types", func() {
			n := Number(0)
			b := Bytes(0)

			Expect(n.String()).To(Equal("0"))
			Expect(b.String()).To(Equal("0"))
			Expect(n.FormatUnitInt()).NotTo(BeEmpty())
			Expect(b.FormatUnitInt()).NotTo(BeEmpty())

			for _, stat := range []Stats{StatBytes, StatPackets, StatFifo, StatDrop, StatErr} {
				Expect(stat.FormatUnit(n)).NotTo(BeEmpty())
			}
		})

		It("should handle MaxUint64 across types", func() {
			maxVal := uint64(18446744073709551615)
			n := Number(maxVal)
			b := Bytes(maxVal)

			Expect(n.AsUint64()).To(Equal(maxVal))
			Expect(b.AsUint64()).To(Equal(maxVal))
			Expect(n.FormatUnitInt()).NotTo(BeEmpty())
			Expect(b.FormatUnitInt()).NotTo(BeEmpty())
		})
	})

	Describe("Formatting consistency", func() {
		It("should maintain consistent padding", func() {
			values := []Number{
				Number(5),
				Number(50),
				Number(500),
				Number(5000),
			}

			results := make([]string, len(values))
			for i, val := range values {
				results[i] = val.FormatUnitInt()
			}

			// All should have consistent structure
			for _, result := range results {
				Expect(len(result)).To(BeNumerically(">=", 4))
			}
		})

		It("should maintain consistent precision", func() {
			values := []Number{
				Number(1234),
				Number(5678900),
				Number(9876543210),
			}

			for _, val := range values {
				result1 := val.FormatUnitFloat(1)
				result2 := val.FormatUnitFloat(2)
				result3 := val.FormatUnitFloat(3)

				Expect(result1).To(MatchRegexp(`\d+\.\d{1}`))
				Expect(result2).To(MatchRegexp(`\d+\.\d{2}`))
				Expect(result3).To(MatchRegexp(`\d+\.\d{3}`))
			}
		})
	})
})
