/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package console

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func printPrompt(text string) {
	if text != "" {
		ColorPrompt.Printf("%s: ", text)
	}
}

func PromptString(text string) (string, error) {
	var (
		scn *bufio.Scanner = bufio.NewScanner(os.Stdin)
		res string
		err error
	)

	printPrompt(text)

	for scn.Scan() {
		res = scn.Text()
		err = scn.Err()
		break
	}

	return res, err
}

func PromptInt(text string) (int64, error) {
	if str, err := PromptString(text); err != nil {
		return 0, err
	} else {
		return strconv.ParseInt(str, 10, 64)
	}
}

func PromptUrl(text string) (*url.URL, error) {
	if str, err := PromptString(text); err != nil {
		return nil, err
	} else {
		return url.Parse(str)
	}
}

func PromptBool(text string) (bool, error) {
	if str, err := PromptString(text); err != nil {
		return false, err
	} else {
		return strconv.ParseBool(str)
	}
}

func PromptPassword(text string) (string, error) {
	printPrompt(text)
	//nolint #unconvert
	res, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")

	return string(res), err
}
