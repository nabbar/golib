## `console` Package

The `console` package provides utilities for enhanced console input/output in Go applications. It offers colored output, user prompts, and string formatting helpers to improve CLI user experience.

### Features

- Colored printing and formatting using [fatih/color](https://github.com/fatih/color)
- Customizable color types for standard and prompt outputs
- User input prompts for strings, integers, booleans, URLs, and passwords (with hidden input)
- String padding and tabulated printing helpers
- Error handling with custom error codes

---

### Main Types & Functions

#### Color Output

- **SetColor(col colorType, value ...int)**: Set color attributes for a color type.
- **ColorPrint / ColorPrompt**: Predefined color types for standard and prompt outputs.
- **Print, Println, Printf, Sprintf**: Print text with or without color.
- **BuffPrintf**: Print formatted text to an `io.Writer` with color support.

#### User Prompts

- **PromptString(text string) (string, error)**: Prompt user for a string.
- **PromptInt(text string) (int64, error)**: Prompt user for an integer.
- **PromptBool(text string) (bool, error)**: Prompt user for a boolean.
- **PromptUrl(text string) (\*url.URL, error)**: Prompt user for a URL.
- **PromptPassword(text string) (string, error)**: Prompt user for a password (input hidden).

#### String Formatting Helpers

- **PadLeft/PadRight/PadCenter(str string, len int, pad string) string**: Pad a string to a given length.
- **PrintTabf(tablLevel int, format string, args ...interface{})**: Print formatted text with indentation.

#### Error Handling

- Custom error codes for parameter validation and I/O errors.
- Errors are wrapped using the `github.com/nabbar/golib/errors` package.

---

### Example Usage

```go
import "github.com/nabbar/golib/console"

func main() {
    // Set prompt color to green
    console.SetColor(console.ColorPrompt, 32)
    name, _ := console.PromptString("Enter your name")
    console.ColorPrint.Printf("Hello, %s!\n", name)

    // Print padded and colored output
    padded := console.PadCenter("Welcome", 20, "-")
    console.ColorPrint.Println(padded)
}
```

---

### Notes

- Colors can be customized using ANSI attributes.
- Password prompts use hidden input (no echo).
- All prompt functions return errors; always check them in production code.
- Deprecated functions are marked and will be removed in future versions.
