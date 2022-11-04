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

package logger_test

import (
	"context"
	"path/filepath"
	"time"

	libsem "github.com/nabbar/golib/semaphore"

	"github.com/nabbar/golib/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func getLogFileTemp() string {
	fsp, err := GetTempFile()
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		err = DelTempFile(fsp)
		Expect(err).ToNot(HaveOccurred())
	}()

	return filepath.Base(fsp)
}

var _ = Describe("Logger", func() {
	Context("Create New Logger with Default Config", func() {
		It("Must succeed", func() {
			log := logger.New(GetContext())
			log.SetLevel(logger.DebugLevel)
			err := log.SetOptions(&logger.Options{})
			Expect(err).ToNot(HaveOccurred())
			log.LogDetails(logger.InfoLevel, "test logger", nil, nil, nil)
		})
	})
	Context("Create New Logger with Default Config and trace", func() {
		It("Must succeed", func() {
			log := logger.New(GetContext())
			log.SetLevel(logger.DebugLevel)
			err := log.SetOptions(&logger.Options{
				EnableTrace: true,
			})
			Expect(err).ToNot(HaveOccurred())
			log.LogDetails(logger.InfoLevel, "test logger with trace", nil, nil, nil)
		})
	})
	Context("Create New Logger with field", func() {
		It("Must succeed", func() {
			log := logger.New(GetContext())
			log.SetLevel(logger.DebugLevel)
			err := log.SetOptions(&logger.Options{
				EnableTrace: true,
			})
			log.SetFields(logger.NewFields().Add("test-field", "ok"))
			Expect(err).ToNot(HaveOccurred())
			log.LogDetails(logger.InfoLevel, "test logger with field", nil, nil, nil)
		})
	})
	Context("Create New Logger with file", func() {
		It("Must succeed", func() {
			log := logger.New(GetContext())
			log.SetLevel(logger.DebugLevel)

			fsp, err := GetTempFile()

			defer func() {
				err = DelTempFile(fsp)
				Expect(err).ToNot(HaveOccurred())
			}()

			Expect(err).ToNot(HaveOccurred())

			err = log.SetOptions(&logger.Options{
				EnableTrace: true,
				LogFile: []logger.OptionsFile{
					{
						LogLevel:   nil,
						Filepath:   fsp,
						Create:     true,
						CreatePath: true,
						//FileMode:         0644,
						//PathMode:         0755,
						DisableStack:     false,
						DisableTimestamp: false,
						EnableTrace:      true,
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			log.SetFields(logger.NewFields().Add("test-field", "ok"))
			log.LogDetails(logger.InfoLevel, "test logger with field", nil, nil, nil)
		})
	})
	Context("Create New Logger with file in multithreading mode", func() {
		It("Must succeed", func() {
			log := logger.New(GetContext())
			log.SetLevel(logger.DebugLevel)

			fsp, err := GetTempFile()

			defer func() {
				err = DelTempFile(fsp)
				Expect(err).ToNot(HaveOccurred())
			}()

			Expect(err).ToNot(HaveOccurred())

			err = log.SetOptions(&logger.Options{
				EnableTrace: true,
				LogFile: []logger.OptionsFile{
					{
						LogLevel:   nil,
						Filepath:   fsp,
						Create:     true,
						CreatePath: true,
						//FileMode:         0644,
						//PathMode:         0755,
						DisableStack:     false,
						DisableTimestamp: false,
						EnableTrace:      true,
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			log.SetFields(logger.NewFields().Add("test-field", "ok"))
			log.LogDetails(logger.InfoLevel, "test logger with field", nil, nil, nil)

			var sub logger.Logger
			sub, err = log.Clone()
			Expect(err).ToNot(HaveOccurred())

			go func(log logger.Logger) {
				defer func() {
					se := log.Close()
					Expect(se).ToNot(HaveOccurred())
				}()

				log.SetFields(logger.NewFields().Add("logger", "sub"))
				for i := 0; i < 10; i++ {
					log.Entry(logger.InfoLevel, "test multithreading logger").FieldAdd("id", i).Log()
				}
			}(sub)

			log.SetFields(logger.NewFields().Add("logger", "main"))
			sem := libsem.NewSemaphoreWithContext(context.Background(), 0)
			defer sem.DeferMain()

			for i := 0; i < 25; i++ {
				Expect(sem.NewWorker()).ToNot(HaveOccurred())

				go func(id int) {
					defer sem.DeferWorker()
					ent := log.Entry(logger.InfoLevel, "test multithreading logger")
					ent.FieldAdd("id", id)
					ent.Log()
				}(i)
			}

			Expect(sem.WaitAll()).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
		})
	})
})
