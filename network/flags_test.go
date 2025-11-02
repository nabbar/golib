/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package network_test

import (
	"net"

	. "github.com/nabbar/golib/network"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flags Functions", func() {
	Describe("FindFlagInList()", func() {
		Context("with flag present in list", func() {
			It("should return true for FlagUp", func() {
				list := []string{"up", "broadcast", "multicast"}
				result := FindFlagInList(list, net.FlagUp)
				Expect(result).To(BeTrue())
			})

			It("should return true for FlagBroadcast", func() {
				list := []string{"up", "broadcast", "multicast"}
				result := FindFlagInList(list, net.FlagBroadcast)
				Expect(result).To(BeTrue())
			})

			It("should return true for FlagLoopback", func() {
				list := []string{"up", "loopback", "running"}
				result := FindFlagInList(list, net.FlagLoopback)
				Expect(result).To(BeTrue())
			})

			It("should return true for FlagPointToPoint", func() {
				list := []string{"up", "pointtopoint", "running"}
				result := FindFlagInList(list, net.FlagPointToPoint)
				Expect(result).To(BeTrue())
			})

			It("should return true for FlagMulticast", func() {
				list := []string{"up", "broadcast", "multicast"}
				result := FindFlagInList(list, net.FlagMulticast)
				Expect(result).To(BeTrue())
			})
		})

		Context("with flag not present in list", func() {
			It("should return false when flag is missing", func() {
				list := []string{"up", "broadcast"}
				result := FindFlagInList(list, net.FlagLoopback)
				Expect(result).To(BeFalse())
			})

			It("should return false for empty list", func() {
				list := []string{}
				result := FindFlagInList(list, net.FlagUp)
				Expect(result).To(BeFalse())
			})

			It("should return false when searching in wrong flags", func() {
				list := []string{"loopback", "running"}
				result := FindFlagInList(list, net.FlagBroadcast)
				Expect(result).To(BeFalse())
			})
		})

		Context("with various flag combinations", func() {
			It("should match exact flag strings", func() {
				// Test all common flags
				flags := map[net.Flags]string{
					net.FlagUp:           "up",
					net.FlagBroadcast:    "broadcast",
					net.FlagLoopback:     "loopback",
					net.FlagPointToPoint: "pointtopoint",
					net.FlagMulticast:    "multicast",
				}

				for flag, flagStr := range flags {
					list := []string{flagStr}
					result := FindFlagInList(list, flag)
					Expect(result).To(BeTrue(), "Flag %s should be found", flagStr)
				}
			})
		})

		Context("edge cases", func() {
			It("should handle list with single flag", func() {
				list := []string{"up"}
				Expect(FindFlagInList(list, net.FlagUp)).To(BeTrue())
				Expect(FindFlagInList(list, net.FlagBroadcast)).To(BeFalse())
			})

			It("should handle list with many flags", func() {
				list := []string{"up", "broadcast", "multicast", "running", "simplex"}
				Expect(FindFlagInList(list, net.FlagUp)).To(BeTrue())
				Expect(FindFlagInList(list, net.FlagBroadcast)).To(BeTrue())
				Expect(FindFlagInList(list, net.FlagMulticast)).To(BeTrue())
			})

			It("should be case-sensitive", func() {
				list := []string{"UP", "BROADCAST"}
				// These should not match because flag strings are lowercase
				result := FindFlagInList(list, net.FlagUp)
				// This depends on how net.Flags.String() works
				// It should return lowercase "up"
				Expect(result).To(BeFalse())
			})

			It("should handle duplicate flags in list", func() {
				list := []string{"up", "up", "broadcast"}
				result := FindFlagInList(list, net.FlagUp)
				Expect(result).To(BeTrue())
			})
		})
	})

	Describe("FindAllFlagInList()", func() {
		Context("when all flags are present", func() {
			It("should return true for single flag", func() {
				list := []string{"up", "broadcast", "multicast"}
				flags := []net.Flags{net.FlagUp}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeTrue())
			})

			It("should return true for multiple flags", func() {
				list := []string{"up", "broadcast", "multicast"}
				flags := []net.Flags{net.FlagUp, net.FlagBroadcast}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeTrue())
			})

			It("should return true when all flags match", func() {
				list := []string{"up", "broadcast", "multicast", "running"}
				flags := []net.Flags{net.FlagUp, net.FlagBroadcast, net.FlagMulticast}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeTrue())
			})
		})

		Context("when some flags are missing", func() {
			It("should return false if one flag is missing", func() {
				list := []string{"up", "broadcast"}
				flags := []net.Flags{net.FlagUp, net.FlagBroadcast, net.FlagMulticast}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeFalse())
			})

			It("should return false if first flag is missing", func() {
				list := []string{"broadcast", "multicast"}
				flags := []net.Flags{net.FlagUp, net.FlagBroadcast}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeFalse())
			})

			It("should return false if last flag is missing", func() {
				list := []string{"up", "broadcast"}
				flags := []net.Flags{net.FlagUp, net.FlagBroadcast, net.FlagMulticast}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeFalse())
			})

			It("should return false if all flags are missing", func() {
				list := []string{"up", "running"}
				flags := []net.Flags{net.FlagBroadcast, net.FlagMulticast}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeFalse())
			})
		})

		Context("with empty inputs", func() {
			It("should return true for empty flags list", func() {
				list := []string{"up", "broadcast"}
				flags := []net.Flags{}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeTrue()) // No flags to check = all present
			})

			It("should return false for empty flag list with flags to check", func() {
				list := []string{}
				flags := []net.Flags{net.FlagUp}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeFalse())
			})

			It("should return true when both are empty", func() {
				list := []string{}
				flags := []net.Flags{}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeTrue())
			})
		})

		Context("real-world scenarios", func() {
			It("should validate typical ethernet interface flags", func() {
				// Typical ethernet interface: up, broadcast, multicast
				list := []string{"up", "broadcast", "multicast", "running"}

				// Check for up and broadcast
				result := FindAllFlagInList(list, []net.Flags{net.FlagUp, net.FlagBroadcast})
				Expect(result).To(BeTrue())

				// Check for loopback (should fail)
				result2 := FindAllFlagInList(list, []net.Flags{net.FlagLoopback})
				Expect(result2).To(BeFalse())
			})

			It("should validate loopback interface flags", func() {
				// Typical loopback interface: up, loopback, running
				list := []string{"up", "loopback", "running"}

				// Check for up and loopback
				result := FindAllFlagInList(list, []net.Flags{net.FlagUp, net.FlagLoopback})
				Expect(result).To(BeTrue())

				// Check for broadcast (should fail - loopback doesn't broadcast)
				result2 := FindAllFlagInList(list, []net.Flags{net.FlagBroadcast})
				Expect(result2).To(BeFalse())
			})

			It("should validate point-to-point interface flags", func() {
				// Point-to-point interface (e.g., VPN)
				list := []string{"up", "pointtopoint", "running", "multicast"}

				// Check for up and point-to-point
				result := FindAllFlagInList(list, []net.Flags{net.FlagUp, net.FlagPointToPoint})
				Expect(result).To(BeTrue())
			})
		})

		Context("order independence", func() {
			It("should return same result regardless of flag order", func() {
				list := []string{"up", "broadcast", "multicast"}

				result1 := FindAllFlagInList(list, []net.Flags{net.FlagUp, net.FlagBroadcast})
				result2 := FindAllFlagInList(list, []net.Flags{net.FlagBroadcast, net.FlagUp})

				Expect(result1).To(Equal(result2))
				Expect(result1).To(BeTrue())
			})

			It("should return same result regardless of list order", func() {
				list1 := []string{"up", "broadcast", "multicast"}
				list2 := []string{"multicast", "up", "broadcast"}
				flags := []net.Flags{net.FlagUp, net.FlagBroadcast}

				result1 := FindAllFlagInList(list1, flags)
				result2 := FindAllFlagInList(list2, flags)

				Expect(result1).To(Equal(result2))
				Expect(result1).To(BeTrue())
			})
		})

		Context("with duplicate flags", func() {
			It("should handle duplicate flags in list", func() {
				list := []string{"up", "up", "broadcast", "broadcast"}
				flags := []net.Flags{net.FlagUp, net.FlagBroadcast}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeTrue())
			})

			It("should handle duplicate flags in check list", func() {
				list := []string{"up", "broadcast", "multicast"}
				flags := []net.Flags{net.FlagUp, net.FlagUp, net.FlagBroadcast}
				result := FindAllFlagInList(list, flags)
				Expect(result).To(BeTrue())
			})
		})
	})

	Describe("Integration tests", func() {
		It("should work together for filtering interfaces", func() {
			// Simulate filtering interfaces by flags
			interfaces := []struct {
				name  string
				flags []string
			}{
				{"eth0", []string{"up", "broadcast", "multicast", "running"}},
				{"lo", []string{"up", "loopback", "running"}},
				{"wlan0", []string{"up", "broadcast", "multicast"}},
				{"tun0", []string{"up", "pointtopoint", "running", "multicast"}},
				{"eth1", []string{"broadcast", "multicast"}}, // Down interface
			}

			// Find all UP interfaces
			upInterfaces := 0
			for _, iface := range interfaces {
				if FindFlagInList(iface.flags, net.FlagUp) {
					upInterfaces++
				}
			}
			Expect(upInterfaces).To(Equal(4)) // eth0, lo, wlan0, tun0

			// Find all UP + BROADCAST interfaces
			upBroadcastInterfaces := 0
			for _, iface := range interfaces {
				if FindAllFlagInList(iface.flags, []net.Flags{net.FlagUp, net.FlagBroadcast}) {
					upBroadcastInterfaces++
				}
			}
			Expect(upBroadcastInterfaces).To(Equal(2)) // eth0, wlan0

			// Find loopback interfaces
			loopbackInterfaces := 0
			for _, iface := range interfaces {
				if FindFlagInList(iface.flags, net.FlagLoopback) {
					loopbackInterfaces++
				}
			}
			Expect(loopbackInterfaces).To(Equal(1)) // lo
		})
	})

	Describe("Performance", func() {
		It("should handle repeated calls efficiently", func() {
			list := []string{"up", "broadcast", "multicast", "running"}
			flags := []net.Flags{net.FlagUp, net.FlagBroadcast}

			Expect(func() {
				for i := 0; i < 10000; i++ {
					_ = FindFlagInList(list, net.FlagUp)
					_ = FindAllFlagInList(list, flags)
				}
			}).NotTo(Panic())
		})

		It("should handle large flag lists efficiently", func() {
			list := []string{
				"up", "broadcast", "multicast", "running",
				"promisc", "allmulti", "master", "slave",
				"debug", "dormant", "echo",
			}
			flags := []net.Flags{net.FlagUp, net.FlagBroadcast, net.FlagMulticast}

			Expect(func() {
				for i := 0; i < 1000; i++ {
					_ = FindAllFlagInList(list, flags)
				}
			}).NotTo(Panic())
		})
	})
})
