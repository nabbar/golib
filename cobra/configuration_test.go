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
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	libcbr "github.com/nabbar/golib/cobra"
	libver "github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cobra Configuration", func() {
	var (
		cobra       libcbr.Cobra
		version     libver.Version
		tempDir     string
		silentPrint func(pkg, file string)
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "cobra-test-*")
		Expect(err).ToNot(HaveOccurred())

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

		// Silent print function to avoid polluting test output
		silentPrint = func(pkg, file string) {
			// Do nothing - suppress output during tests
		}
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	Describe("AddCommandConfigure", func() {
		It("should add configure command", func() {
			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"test": "value"}`)
			}

			cobra.AddCommandConfigure("config", ".testapp", defaultConfig)

			rootCmd := cobra.Cobra()
			var found bool
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "configure <file path with valid extension (json, yaml, toml, ...) to be generated>" {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should add configure command with alias", func() {
			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"test": "value"}`)
			}

			cobra.AddCommandConfigure("cfg", ".testapp", defaultConfig)

			rootCmd := cobra.Cobra()
			var foundCmd *struct {
				Aliases []string
			}

			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "configure") {
					// Command exists
					foundCmd = &struct{ Aliases []string }{Aliases: cmd.Aliases}
					break
				}
			}

			Expect(foundCmd).ToNot(BeNil())
		})

		It("should support empty alias", func() {
			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"test": "value"}`)
			}

			cobra.AddCommandConfigure("", ".testapp", defaultConfig)

			rootCmd := cobra.Cobra()
			var found bool
			for _, cmd := range rootCmd.Commands() {
				if strings.Contains(cmd.Use, "configure") {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})
	})

	Describe("ConfigureCheckArgs", func() {
		It("should accept valid file path", func() {
			testPath := filepath.Join(tempDir, "config.json")
			err := cobra.ConfigureCheckArgs(".testapp", []string{testPath})

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle no arguments with basename", func() {
			err := cobra.ConfigureCheckArgs(".testapp", []string{})

			// Should generate default path
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject too many arguments", func() {
			err := cobra.ConfigureCheckArgs(".testapp", []string{"file1", "file2"})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("too many"))
		})

		It("should handle empty basename with no args", func() {
			err := cobra.ConfigureCheckArgs("", []string{})

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ConfigureWriteConfig", func() {
		It("should write JSON config file", func() {
			configPath := filepath.Join(tempDir, "test.json")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"key": "value", "number": 123}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			// Verify file exists
			_, err = os.Stat(configPath)
			Expect(err).ToNot(HaveOccurred())

			// Verify content
			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("key"))
			Expect(string(content)).To(ContainSubstring("value"))
		})

		It("should write YAML config file", func() {
			configPath := filepath.Join(tempDir, "test.yaml")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"key": "value", "number": 123}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			// Verify file exists
			_, err = os.Stat(configPath)
			Expect(err).ToNot(HaveOccurred())

			// Verify YAML content
			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("key:"))
			Expect(string(content)).To(ContainSubstring("value"))
		})

		It("should write TOML config file", func() {
			configPath := filepath.Join(tempDir, "test.toml")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"key": "value", "number": 123}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			// Verify file exists
			_, err = os.Stat(configPath)
			Expect(err).ToNot(HaveOccurred())

			// Verify TOML content
			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("key"))
			Expect(string(content)).To(ContainSubstring("="))
		})

		It("should handle .yml extension", func() {
			configPath := filepath.Join(tempDir, "test.yml")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"key": "value"}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			_, err = os.Stat(configPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should default to JSON for unknown extensions", func() {
			configPath := filepath.Join(tempDir, "test.conf")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"key": "value"}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			// Should create .json file
			jsonPath := filepath.Join(tempDir, "test.json")
			_, err = os.Stat(jsonPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should set file permissions to 0600", func() {
			configPath := filepath.Join(tempDir, "test.json")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"key": "value"}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			info, err := os.Stat(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))
		})

		It("should call custom print message function", func() {
			configPath := filepath.Join(tempDir, "test.json")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"key": "value"}`)
			}

			var printedPkg, printedFile string
			printMsg := func(pkg, file string) {
				printedPkg = pkg
				printedFile = file
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, printMsg)
			Expect(err).ToNot(HaveOccurred())

			Expect(printedPkg).ToNot(BeEmpty())
			Expect(printedFile).To(Equal(configPath))
		})

		It("should handle nested directory paths", func() {
			nestedPath := filepath.Join(tempDir, "nested", "dir", "config.json")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"key": "value"}`)
			}

			// Create parent directory
			err := os.MkdirAll(filepath.Dir(nestedPath), 0755)
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureCheckArgs(".testapp", []string{nestedPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			_, err = os.Stat(nestedPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle empty config content", func() {
			configPath := filepath.Join(tempDir, "empty.json")

			defaultConfig := func() io.Reader {
				return bytes.NewReader([]byte{})
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("0 byte"))
		})

		It("should overwrite existing file", func() {
			configPath := filepath.Join(tempDir, "overwrite.json")

			// Write initial file
			err := os.WriteFile(configPath, []byte(`{"old": "data"}`), 0600)
			Expect(err).ToNot(HaveOccurred())

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"new": "data"}`)
			}

			err = cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			// Verify new content
			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("new"))
			Expect(string(content)).ToNot(ContainSubstring("old"))
		})
	})

	Describe("Format Conversion", func() {
		It("should convert JSON to YAML correctly", func() {
			configPath := filepath.Join(tempDir, "convert.yaml")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{
					"string": "value",
					"number": 42,
					"bool": true,
					"nested": {"key": "nested_value"}
				}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())

			yamlContent := string(content)
			Expect(yamlContent).To(ContainSubstring("string:"))
			Expect(yamlContent).To(ContainSubstring("number:"))
			Expect(yamlContent).To(ContainSubstring("bool:"))
			Expect(yamlContent).To(ContainSubstring("nested:"))
		})

		It("should convert JSON to TOML correctly", func() {
			configPath := filepath.Join(tempDir, "convert.toml")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{
					"string": "value",
					"number": 42,
					"bool": true
				}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())

			tomlContent := string(content)
			Expect(tomlContent).To(ContainSubstring("string"))
			Expect(tomlContent).To(ContainSubstring("="))
			Expect(tomlContent).To(ContainSubstring("42"))
		})

		It("should handle complex nested structures in YAML", func() {
			configPath := filepath.Join(tempDir, "complex.yaml")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{
					"database": {
						"host": "localhost",
						"port": 5432,
						"credentials": {
							"username": "admin",
							"password": "secret"
						}
					},
					"servers": ["server1", "server2", "server3"]
				}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())

			yamlContent := string(content)
			Expect(yamlContent).To(ContainSubstring("database:"))
			Expect(yamlContent).To(ContainSubstring("host:"))
			Expect(yamlContent).To(ContainSubstring("credentials:"))
			Expect(yamlContent).To(ContainSubstring("servers:"))
		})
	})

	Describe("Edge Cases", func() {
		It("should handle special characters in config", func() {
			configPath := filepath.Join(tempDir, "special.json")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"special": "value with spaces & symbols!"}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("&"))
			Expect(string(content)).To(ContainSubstring("!"))
		})

		It("should handle unicode in config", func() {
			configPath := filepath.Join(tempDir, "unicode.json")

			defaultConfig := func() io.Reader {
				return strings.NewReader(`{"message": "Hello 世界 🌍"}`)
			}

			err := cobra.ConfigureCheckArgs(".testapp", []string{configPath})
			Expect(err).ToNot(HaveOccurred())

			err = cobra.ConfigureWriteConfig(".testapp", defaultConfig, silentPrint)
			Expect(err).ToNot(HaveOccurred())

			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(content)).To(BeNumerically(">", 0))
		})
	})
})
