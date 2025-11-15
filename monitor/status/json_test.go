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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/monitor/status"
)

var _ = Describe("JSON Marshaling", func() {
	Describe("MarshalJSON", func() {
		Context("with KO status", func() {
			It("should marshal to \"KO\"", func() {
				s := status.KO
				data, err := s.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"KO"`))
			})
		})

		Context("with Warn status", func() {
			It("should marshal to \"Warn\"", func() {
				s := status.Warn
				data, err := s.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"Warn"`))
			})
		})

		Context("with OK status", func() {
			It("should marshal to \"OK\"", func() {
				s := status.OK
				data, err := s.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"OK"`))
			})
		})

		Context("with unknown status", func() {
			It("should marshal unknown status as \"KO\"", func() {
				s := status.Status(99)
				data, err := s.MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"KO"`))
			})
		})

		Context("using standard json.Marshal", func() {
			It("should work with json.Marshal for OK", func() {
				s := status.OK
				data, err := json.Marshal(s)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"OK"`))
			})

			It("should work with json.Marshal for Warn", func() {
				s := status.Warn
				data, err := json.Marshal(s)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"Warn"`))
			})

			It("should work with json.Marshal for KO", func() {
				s := status.KO
				data, err := json.Marshal(s)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"KO"`))
			})
		})

		Context("in struct", func() {
			type TestStruct struct {
				Status status.Status `json:"status"`
				Name   string        `json:"name"`
			}

			It("should marshal struct with Status field", func() {
				ts := TestStruct{
					Status: status.OK,
					Name:   "test",
				}
				data, err := json.Marshal(ts)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`{"status":"OK","name":"test"}`))
			})

			It("should handle multiple status values in struct", func() {
				type MultiStatus struct {
					Status1 status.Status `json:"status1"`
					Status2 status.Status `json:"status2"`
					Status3 status.Status `json:"status3"`
				}

				ms := MultiStatus{
					Status1: status.KO,
					Status2: status.Warn,
					Status3: status.OK,
				}
				data, err := json.Marshal(ms)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`{"status1":"KO","status2":"Warn","status3":"OK"}`))
			})
		})
	})
})

var _ = Describe("JSON Unmarshaling", func() {
	Describe("UnmarshalJSON", func() {
		Context("with string values", func() {
			It("should unmarshal \"OK\" to OK status", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"OK"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})

			It("should unmarshal \"ok\" (lowercase) to OK status", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"ok"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})

			It("should unmarshal \"Warn\" to Warn status", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"Warn"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.Warn))
			})

			It("should unmarshal \"warn\" (lowercase) to Warn status", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"warn"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.Warn))
			})

			It("should unmarshal \"KO\" to KO status", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"KO"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should unmarshal \"ko\" (lowercase) to KO status", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"ko"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should unmarshal unknown strings to KO", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"unknown"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})
		})

		Context("with numeric strings (treated as invalid)", func() {
			It("should unmarshal numeric string \"0\" to KO status (not recognized as status string)", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"0"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should unmarshal numeric string \"1\" to KO status (not recognized as status string)", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"1"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should unmarshal numeric string \"2\" to KO status (not recognized as status string)", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"2"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should unmarshal numeric string \"999\" to KO status", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`"999"`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})
		})

		Context("with null value", func() {
			It("should unmarshal null to KO status", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`null`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})
		})

		Context("with invalid JSON", func() {
			It("should handle empty quotes as KO", func() {
				var s status.Status
				err := s.UnmarshalJSON([]byte(`""`))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})
		})

		Context("using standard json.Unmarshal", func() {
			It("should work with json.Unmarshal for string values", func() {
				var s status.Status
				err := json.Unmarshal([]byte(`"OK"`), &s)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})

			It("should work with json.Unmarshal for Warn string", func() {
				var s status.Status
				err := json.Unmarshal([]byte(`"Warn"`), &s)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.Warn))
			})
		})

		Context("in struct", func() {
			type TestStruct struct {
				Status status.Status `json:"status"`
				Name   string        `json:"name"`
			}

			It("should unmarshal struct with Status field from string", func() {
				var ts TestStruct
				err := json.Unmarshal([]byte(`{"status":"OK","name":"test"}`), &ts)
				Expect(err).NotTo(HaveOccurred())
				Expect(ts.Status).To(Equal(status.OK))
				Expect(ts.Name).To(Equal("test"))
			})

			It("should unmarshal struct with Status field from Warn string", func() {
				var ts TestStruct
				err := json.Unmarshal([]byte(`{"status":"Warn","name":"test"}`), &ts)
				Expect(err).NotTo(HaveOccurred())
				Expect(ts.Status).To(Equal(status.Warn))
				Expect(ts.Name).To(Equal("test"))
			})

			It("should handle null status in struct", func() {
				var ts TestStruct
				err := json.Unmarshal([]byte(`{"status":null,"name":"test"}`), &ts)
				Expect(err).NotTo(HaveOccurred())
				Expect(ts.Status).To(Equal(status.KO))
				Expect(ts.Name).To(Equal("test"))
			})
		})
	})
})

var _ = Describe("JSON Round-trip", func() {
	It("should maintain status through marshal and unmarshal for KO", func() {
		original := status.KO
		data, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())

		var unmarshaled status.Status
		err = json.Unmarshal(data, &unmarshaled)
		Expect(err).NotTo(HaveOccurred())
		Expect(unmarshaled).To(Equal(original))
	})

	It("should maintain status through marshal and unmarshal for Warn", func() {
		original := status.Warn
		data, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())

		var unmarshaled status.Status
		err = json.Unmarshal(data, &unmarshaled)
		Expect(err).NotTo(HaveOccurred())
		Expect(unmarshaled).To(Equal(original))
	})

	It("should maintain status through marshal and unmarshal for OK", func() {
		original := status.OK
		data, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())

		var unmarshaled status.Status
		err = json.Unmarshal(data, &unmarshaled)
		Expect(err).NotTo(HaveOccurred())
		Expect(unmarshaled).To(Equal(original))
	})

	It("should handle multiple round-trips", func() {
		s := status.OK
		for i := 0; i < 5; i++ {
			data, err := json.Marshal(s)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled status.Status
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(unmarshaled).To(Equal(s))
			s = unmarshaled
		}
	})

	It("should handle array of statuses", func() {
		original := []status.Status{status.KO, status.Warn, status.OK}
		data, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal(`["KO","Warn","OK"]`))

		var unmarshaled []status.Status
		err = json.Unmarshal(data, &unmarshaled)
		Expect(err).NotTo(HaveOccurred())
		Expect(unmarshaled).To(Equal(original))
	})

	It("should handle map with status values", func() {
		original := map[string]status.Status{
			"first":  status.OK,
			"second": status.Warn,
			"third":  status.KO,
		}
		data, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())

		var unmarshaled map[string]status.Status
		err = json.Unmarshal(data, &unmarshaled)
		Expect(err).NotTo(HaveOccurred())
		Expect(unmarshaled).To(Equal(original))
	})
})
