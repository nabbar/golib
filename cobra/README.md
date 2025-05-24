## `cobra` Package

The `cobra` package provides a wrapper and utility layer around the [spf13/cobra](https://github.com/spf13/cobra) library, simplifying the creation of CLI applications in Go. It adds helpers for configuration management, command completion, error code printing, and flag handling, as well as integration with logging and versioning.

### Features

- Simplified CLI application setup using Cobra
- Automatic version and description management
- Built-in commands for shell completion, configuration file generation, and error code listing
- Extensive helpers for adding flags of various types (string, int, bool, slices, maps, etc.)
- Integration with custom logger and version modules
- Support for persistent and local flags

---

### Main Types

- **Cobra**: Main interface for building CLI applications
- **FuncInit, FuncLogger, FuncViper, FuncPrintErrorCode**: Function types for custom initialization, logging, configuration, and error printing

---

### Quick Start

#### Create a New CLI Application

```go
import (
    "github.com/nabbar/golib/cobra"
    "github.com/nabbar/golib/version"
)

func main() {
    app := cobra.New()
    app.SetVersion(version.New(...)) // Set your version info
    app.Init()
    // Add commands, flags, etc.
    app.Execute()
}
```

#### Add Built-in Commands

- **Completion**: Generates shell completion scripts (bash, zsh, fish, PowerShell)
- **Configure**: Generates a configuration file in JSON, YAML, or TOML
- **Error**: Prints error codes for the application

```go
app.AddCommandCompletion()
app.AddCommandConfigure("conf", "myapp", defaultConfigFunc)
app.AddCommandPrintErrorCode(func(code, desc string) {
    fmt.Printf("%s: %s\n", code, desc)
})
```

#### Add Flags

```go
var configPath string
app.SetFlagConfig(true, &configPath)

var verbose int
app.SetFlagVerbose(true, &verbose)

var myString string
app.AddFlagString(false, &myString, "name", "n", "default", "Description of the flag")
```

#### Add Custom Commands

```go
cmd := app.NewCommand("hello", "Say hello", "Prints hello world", "", "")
cmd.Run = func(cmd *cobra.Command, args []string) {
    fmt.Println("Hello, world!")
}
app.AddCommand(cmd)
```

---

### API Overview

- **SetVersion(v Version)**: Set version information
- **SetFuncInit(fct FuncInit)**: Set custom initialization function
- **SetLogger(fct FuncLogger)**: Set custom logger
- **SetViper(fct FuncViper)**: Set custom viper config handler
- **SetFlagConfig(persistent, \*string)**: Add a config file flag
- **SetFlagVerbose(persistent, \*int)**: Add a verbose flag
- **AddFlagXXX(...)**: Add various types of flags (see code for full list)
- **AddCommandCompletion()**: Add shell completion command
- **AddCommandConfigure(alias, basename, defaultConfigFunc)**: Add config file generation command
- **AddCommandPrintErrorCode(fct)**: Add error code listing command
- **Execute()**: Run the CLI application

---

### Error Handling

All commands and helpers return standard Go `error` values. Always check and handle errors when using the API.

---

### Notes

- The package is designed to be extensible and integrates with custom logger and version modules from `golib`.
- Configuration file generation supports JSON, YAML, and TOML formats.
- Shell completion scripts can be generated for bash, zsh, fish, and PowerShell.
