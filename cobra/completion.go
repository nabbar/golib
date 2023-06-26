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
	"os"
	"path/filepath"
	"strings"

	loglvl "github.com/nabbar/golib/logger/level"
	spfcbr "github.com/spf13/cobra"
)

func (c *cobra) AddCommandCompletion() {
	pkg := c.getPackageName()

	desc := "This command will create a completion shel script for simplify the use of this app.\n" +
		"To do this," +
		"\n\t 1- generate a completion script for your shell, like this : " +
		"\n\t\t" + pkg + " completion bash /etc/bash_completion.d/" + pkg +
		"\n\n 2- enable completion into your shell" +
		"\n\t\t example to bash, you need to install the package `bash-completion`" +
		"\n\n 3- enable completion into your shell profile" +
		"\n\t\t example to bash, you need to uncomment the completion section into your /home/<user>/.bashrc" +
		"\n\n"

	cmd := &spfcbr.Command{
		Use:     "completion <Bash|Zsh|PowerShell|Fish> <Completion File to be write>",
		Example: "completion bash /etc/bash_completion.d/" + pkg,
		Short:   "Generate a completion scripts for bash, zsh, fish or powershell",
		Long:    desc,
		Run: func(cmd *spfcbr.Command, args []string) {
			var file string

			if len(args) < 1 {
				c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "missing args", cmd.Usage())
				os.Exit(1)
			} else if len(args) >= 2 {
				file = filepath.Clean(args[1])
				c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "create file path", os.MkdirAll(filepath.Dir(file), 0755))
			}

			switch strings.ToLower(args[0]) {
			case "bash":
				if file == "" {
					c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "generating bash completion", c.c.GenBashCompletionV2(os.Stdout, true))
				} else {
					c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "generating bash completion", c.c.GenBashCompletionFileV2(file, true))
				}
			case "fish":
				if file == "" {
					c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "generating bash completion", c.c.GenFishCompletion(os.Stdout, true))
				} else {
					c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "generating bash completion", c.c.GenFishCompletionFile(file, true))
				}
			case "powershell":
				if file == "" {
					c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "generating bash completion", c.c.GenPowerShellCompletionWithDesc(os.Stdout))
				} else {
					c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "generating bash completion", c.c.GenPowerShellCompletionFileWithDesc(file))
				}
			case "zsh":
				if file == "" {
					c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "generating bash completion", c.c.GenZshCompletion(os.Stdout))
				} else {
					c.getLog().CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "generating bash completion", c.c.GenZshCompletionFile(file))
				}
			}
		},
	}
	c.c.AddCommand(cmd)
}
