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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	loglvl "github.com/nabbar/golib/logger/level"
	"github.com/pelletier/go-toml"
	spfcbr "github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func (c *cobra) getDefaultPath(baseName string) (string, error) {
	path := ""

	// Find home directory.
	home, err := homedir.Dir()
	c.getLog().CheckError(loglvl.WarnLevel, loglvl.InfoLevel, "Loading home dir", err)

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

func (c *cobra) ConfigureCheckArgs(basename string, args []string) error {
	if len(args) < 1 {
		var err error
		c.f, err = c.getDefaultPath(basename)
		return err
	} else if len(args) > 1 {
		return fmt.Errorf("arguments error: too many file path specify")
	} else {
		c.f = args[0]
	}

	return nil
}

func (c *cobra) ConfigureWriteConfig(basename string, defaultConfig func() io.Reader, printMsg func(pkg, file string)) error {
	pkg := c.getPackageName()

	// Use c.f (set by ConfigureCheckArgs) if available, otherwise use basename
	filePath := c.f
	if len(filePath) < 1 {
		filePath = basename
	}

	if len(filePath) < 1 && len(pkg) > 0 {
		filePath = "." + strings.ToLower(pkg)
	}

	var (
		fs  *os.File
		rt  *os.Root
		ext string
		buf io.Reader
		nbr int64
		err error
	)

	defer func() {
		if rt != nil {
			_ = rt.Close()
		}
	}()

	defer func() {
		if fs != nil {
			_ = fs.Close()
		}
	}()

	ext = strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".toml", ".tml":
		if buf, err = c.jsonToToml(defaultConfig()); err != nil {
			return err
		}
	case ".yaml", ".yml":
		if buf, err = c.jsonToYaml(defaultConfig()); err != nil {
			return err
		}
	default:
		buf = defaultConfig()
		filePath = strings.TrimSuffix(filePath, ext) + ".json"
	}

	// Update c.f with the final file path
	c.f = filePath

	rt, err = os.OpenRoot(filepath.Dir(filePath))

	if err != nil {
		return err
	}

	fs, err = rt.OpenFile(filepath.Base(filePath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)

	if err != nil {
		return err
	}

	if nbr, err = io.Copy(fs, buf); err != nil {
		return err
	} else if nbr < 1 {
		return fmt.Errorf("error wrting 0 byte to config file")
	} else if err = fs.Close(); err != nil {
		fs = nil
		return err
	} else {
		fs = nil
	}

	err = os.Chmod(filePath, 0600)
	if err != nil {
		return err
	}

	if printMsg == nil {
		println(fmt.Sprintf("\n\t>> Config File '%s' has been created and file permission have been set.", filePath))
		println("\t>> To explicitly specify this config file when you call this tool, use the '-c' flag like this: ")
		println(fmt.Sprintf("\t\t\t %s -c %s <cmd>...\n", pkg, filePath))
	} else {
		printMsg(pkg, filePath)
	}

	return nil
}

func (c *cobra) jsonToToml(r io.Reader) (io.Reader, error) {
	var (
		e   error
		p   = make([]byte, 0)
		buf = bytes.NewBuffer(p)
		mod = make(map[string]interface{}, 0)
	)

	if p, e = io.ReadAll(r); e != nil {
		return nil, e
	} else if e = json.Unmarshal(p, &mod); e != nil {
		return nil, e
	} else if p, e = toml.Marshal(mod); e != nil {
		return nil, e
	} else {
		buf.Write(p)
		return buf, nil
	}
}

func (c *cobra) jsonToYaml(r io.Reader) (io.Reader, error) {
	var (
		e   error
		p   = make([]byte, 0)
		buf = bytes.NewBuffer(p)
		mod = make(map[string]interface{}, 0)
	)

	if p, e = io.ReadAll(r); e != nil {
		return nil, e
	} else if e = json.Unmarshal(p, &mod); e != nil {
		return nil, e
	} else if p, e = yaml.Marshal(mod); e != nil {
		return nil, e
	} else {
		buf.Write(p)
		return buf, nil
	}
}

func (c *cobra) AddCommandConfigure(alias, basename string, defaultConfig func() io.Reader) {
	pkg := c.getPackageName()

	if basename == "" && pkg != "" {
		basename = "." + strings.ToLower(pkg)
	}

	cmd := &spfcbr.Command{
		Use:     "configure <file path with valid extension (json, yaml, toml, ...) to be generated>",
		Example: "configure ~/." + strings.ToLower(pkg) + ".yml",
		Short:   "Generate config file",
		Long: `Generates a configuration file based on giving existing config flag
override by passed flag in command line and completed with default for non existing values.`,

		RunE: func(cmd *spfcbr.Command, args []string) error {
			return c.ConfigureWriteConfig(basename, defaultConfig, nil)
		},

		Args: func(cmd *spfcbr.Command, args []string) error {
			return c.ConfigureCheckArgs(basename, args)
		},
	}

	if len(alias) > 0 {
		cmd.Aliases = []string{alias}
	}

	c.c.AddCommand(cmd)
}
