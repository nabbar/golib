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

package cobra

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"

	liblog "github.com/nabbar/golib/logger"
	spfcbr "github.com/spf13/cobra"
)

func (c *cobra) getDefaultPath(baseName string) (string, error) {
	path := ""

	// Find home directory.
	home, err := homedir.Dir()
	c.getLog().CheckError(liblog.WarnLevel, liblog.InfoLevel, "Loading home dir", err)

	// set configname based on package name
	if baseName == "" {
		return "", fmt.Errorf("arguments missing: requires the destination file path")
	}

	path = filepath.Clean(home + string(filepath.Separator) + baseName + ".json")

	if path == "." || path == ".json" {
		return "", fmt.Errorf("arguments missing: requires the destination file path")
	}

	return path, nil
}

func (c *cobra) AddCommandConfigure(basename string, defaultConfig func() io.Reader) {
	pkg := c.getPackageName()

	if basename == "" && pkg != "" {
		basename = "." + strings.ToLower(pkg)
	}

	var cfgFile string

	c.c.AddCommand(&spfcbr.Command{
		Use:     "configure <file path with valid extension (json, yaml, toml, ...) to be generated>",
		Example: "configure ~/." + strings.ToLower(pkg) + ".yml",
		Short:   "Generate config file",
		Long: `Generates a configuration file based on giving existing config flag
override by passed flag in command line and completed with default for non existing values.`,

		Run: func(cmd *spfcbr.Command, args []string) {
			var fs *os.File

			defer func() {
				if fs != nil {
					_ = fs.Close()
				}
			}()

			buf, err := ioutil.ReadAll(defaultConfig())
			c.getLog().CheckError(liblog.FatalLevel, liblog.DebugLevel, "reading default config", err)

			if len(path.Ext(cfgFile)) > 0 && strings.ToLower(path.Ext(cfgFile)) != ".json" {
				var mod = make(map[string]interface{}, 0)

				err = json.Unmarshal(buf, &mod)
				c.getLog().CheckError(liblog.FatalLevel, liblog.DebugLevel, "transform json default config", err)

				switch strings.ToLower(path.Ext(cfgFile)) {
				case ".toml":
					buf, err = toml.Marshal(mod)
				case ".yml", ".yaml":
					buf, err = yaml.Marshal(mod)
				default:
					c.getLog().CheckError(liblog.FatalLevel, liblog.DebugLevel, "get encode for extension file", fmt.Errorf("extension file '%s' not compatible", path.Ext(cfgFile)))
				}
			}

			fs, err = os.OpenFile(cfgFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
			c.getLog().CheckError(liblog.FatalLevel, liblog.DebugLevel, "opening destination config file for exclusive write with truncate", err)

			_, err = fs.Write(buf)
			c.getLog().CheckError(liblog.FatalLevel, liblog.DebugLevel, fmt.Sprintf("writing config to file '%s'", cfgFile), err)

			err = os.Chmod(cfgFile, 0600)
			if !c.getLog().CheckError(liblog.ErrorLevel, liblog.InfoLevel, fmt.Sprintf("setting permission for config file '%s'", cfgFile), err) {
				println(fmt.Sprintf("\n\t>> Config File '%s' has been created and file permission have been set.", cfgFile))
				println("\t>> To explicitly specify this config file when you call this tool, use the '-c' flag like this: ")
				println(fmt.Sprintf("\t\t\t %s -c %s <cmd>...\n", pkg, cfgFile))
			}
		},

		Args: func(cmd *spfcbr.Command, args []string) error {
			if len(args) < 1 {
				var err error
				cfgFile, err = c.getDefaultPath(basename)
				return err
			} else if len(args) > 1 {
				return fmt.Errorf("arguments error: too many file path specify")
			} else {
				cfgFile = args[0]
			}

			return nil
		},
	})
}
