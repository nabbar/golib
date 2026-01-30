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
 */

package shell_test

import (
	"fmt"
	"io"
	"os"

	"github.com/nabbar/golib/shell"
	"github.com/nabbar/golib/shell/command"
)

// ExampleNew demonstrates creating a new shell instance
func ExampleNew() {
	sh := shell.New(nil)
	fmt.Printf("Shell created: %T\n", sh)
	// Output:
	// Shell created: *shell.shell
}

// ExampleShell_Add demonstrates adding commands to the shell
func ExampleShell_Add() {
	sh := shell.New(nil)

	// Add commands without prefix
	sh.Add("",
		command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
			fmt.Fprintln(out, "Hello, World!")
		}),
		command.New("echo", "Echo arguments", func(out, err io.Writer, args []string) {
			fmt.Fprintln(out, args)
		}),
	)

	// Add commands with prefix
	sh.Add("sys:",
		command.New("info", "System info", func(out, err io.Writer, args []string) {
			fmt.Fprintln(out, "System Information")
		}),
	)

	fmt.Println("Commands added successfully")
	// Output:
	// Commands added successfully
}

// ExampleShell_Get demonstrates retrieving commands
func ExampleShell_Get() {
	sh := shell.New(nil)

	sh.Add("", command.New("hello", "Say hello", nil))

	cmd, found := sh.Get("hello")
	if found {
		fmt.Printf("Found command: %s\n", cmd.Name())
	}

	_, found = sh.Get("nonexistent")
	if !found {
		fmt.Println("Command not found")
	}

	// Output:
	// Found command: hello
	// Command not found
}

// ExampleShell_Desc demonstrates getting command descriptions
func ExampleShell_Desc() {
	sh := shell.New(nil)

	sh.Add("",
		command.New("hello", "Say hello", nil),
		command.New("echo", "Echo text", nil),
	)

	desc := sh.Desc("hello")
	fmt.Printf("hello: %s\n", desc)

	// Output:
	// hello: Say hello
}

// ExampleShell_Walk demonstrates iterating over commands
func ExampleShell_Walk() {
	sh := shell.New(nil)

	sh.Add("",
		command.New("cmd1", "Command 1", nil),
		command.New("cmd2", "Command 2", nil),
		command.New("cmd3", "Command 3", nil),
	)

	count := 0
	sh.Walk(func(name string, item command.Command) bool {
		count++
		return true
	})

	fmt.Printf("Total commands: %d\n", count)
	// Output:
	// Total commands: 3
}

// ExampleShell_Run demonstrates executing commands
func ExampleShell_Run() {
	sh := shell.New(nil)

	sh.Add("", command.New("hello", "Say hello", func(out, err io.Writer, args []string) {
		if len(args) > 0 {
			fmt.Fprintf(out, "Hello, %s!", args[0])
		} else {
			fmt.Fprint(out, "Hello, World!")
		}
	}))

	// Execute without arguments
	sh.Run(os.Stdout, os.Stderr, []string{"hello"})
	fmt.Println()

	// Execute with arguments
	sh.Run(os.Stdout, os.Stderr, []string{"hello", "Alice"})
	fmt.Println()

	// Output:
	// Hello, World!
	// Hello, Alice!
}

// Example_workflow demonstrates a complete workflow
func Example_workflow() {
	sh := shell.New(nil)

	// Register commands
	sh.Add("",
		command.New("greet", "Greet someone", func(out, err io.Writer, args []string) {
			name := "World"
			if len(args) > 0 {
				name = args[0]
			}
			fmt.Fprintf(out, "Greetings, %s!", name)
		}),
	)

	sh.Add("sys:",
		command.New("version", "Show version", func(out, err io.Writer, args []string) {
			fmt.Fprint(out, "v1.0.0")
		}),
	)

	// Get command information
	cmd, found := sh.Get("greet")
	if found {
		fmt.Printf("Command: %s - %s\n", cmd.Name(), cmd.Describe())
	}

	// Execute commands
	sh.Run(os.Stdout, os.Stderr, []string{"greet", "User"})
	fmt.Println()

	sh.Run(os.Stdout, os.Stderr, []string{"sys:version"})
	fmt.Println()

	// Count all commands
	count := 0
	sh.Walk(func(name string, item command.Command) bool {
		count++
		return true
	})
	fmt.Printf("Total registered commands: %d\n", count)

	// Output:
	// Command: greet - Greet someone
	// Greetings, User!
	// v1.0.0
	// Total registered commands: 2
}

// Example_namespaces demonstrates using command prefixes
func Example_namespaces() {
	sh := shell.New(nil)

	// Add commands to different namespaces
	sh.Add("sys:", command.New("info", "System info", func(out, err io.Writer, args []string) {
		fmt.Fprint(out, "System Info")
	}))

	sh.Add("user:", command.New("info", "User info", func(out, err io.Writer, args []string) {
		fmt.Fprint(out, "User Info")
	}))

	// Execute commands from different namespaces
	sh.Run(os.Stdout, os.Stderr, []string{"sys:info"})
	fmt.Println()

	sh.Run(os.Stdout, os.Stderr, []string{"user:info"})
	fmt.Println()

	// Output:
	// System Info
	// User Info
}
