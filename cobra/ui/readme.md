# UI Package

The UI package empowers developers to craft immersive and interactive user interfaces seamlessly integrated into their applications.  
## Features

- File Input with Pagination: Facilitate the selection of files from a directory with built-in pagination support (10 files/page). Users can navigate through large sets of files seamlessly.  
- Error Handling: Handle errors gracefully during user interactions. When a handler encounters an error, it returns an error object containing relevant information about the error. The UI displays the error message to the user and prompts the question again.  
- Single-Choice Questions: Allow users to select one option from a list of choices.
- Text Input Questions: Prompt users to enter text-based inputs.  
- Password Input Questions: Securely collect password inputs from users, hiding the entered characters for privacy.  
- Dynamic Pagination: Automatically paginate choices for single-choice questions with more than 10 options, ensuring a smooth user experience without overwhelming them with too many choices at once.  

These features collectively empower developers to create engaging and user-friendly interfaces, whether they are standalone applications or CLI tools integrated with Cobra.   

## Standalone Usage Examples:
1- Single-choice Question:  
``` 
import (
    "fmt"
    "github.com/nabbar/cobra/ui"
)

func main() {
    var choice string
    tui := ui.New()
    tui.SetQuestions([]ui.Question{
        {
            Text: "What is your preferred programming language?",
            Options: []string{"Go", "Python", "JavaScript", "Java"},
            Handler: func(input string) error {
                choice = input
                return nil
            },
        },
    })
    // if the options > 10 => pagination will be done dynamically
    tui.RunInteractiveUI()
    fmt.Println(choice)
}
```  
2- Text Input Question
``` 
package main

import (
	"fmt"
	"strconv"
	"errors"
	"github.com/nabbar/cobra/ui"
)

func main() {
	var(
	    age int
	    err error
	)
	tui := ui.New()
	tui.SetQuestions([]ui.Question{
		{
			Text:    "Enter your age:",
			Handler: func(input string) error {
				age,err = strconv.Atoi(input)
				if err !=nil{
				  return errors.New("Age must be an integer")
				}
			},
		},
	})
	tui.RunInteractiveUI()
	fmt.Printf("Hello, %s!\n", userName)
}
```  
3- Password Input Question  
``` 
package main

import (
    "fmt"
	"github.com/nabbar/cobra/ui"
)

func main() {
	var passwordEntered bool
	tui := ui.New()
	tui.SetQuestions([]dynamicui.Question{
		{
			Text:          "Enter your password:",
			PasswordInput: true,
			Handler: func(password string) error {
				passwordEntered = true
				return nil
			},
		},
	})
	tui.RunInteractiveUI()
	if passwordEntered {
		fmt.Println("Password entered.")
	}
}
```  
4- File Input Question  
``` 
package main

import (
	"fmt"
	"github.com/nabbar/cobra/ui"
)

func main() {
	var selectedFile string
	tui := ui.New()
	tui.SetQuestions([]ui.Question{
		{
			Text:     "Select a file:",
			FilePath: true,
			Handler: func(filePath string) error {
				selectedFile = filePath
				return nil
			},
		},
	})
	tui.RunInteractiveUI()
	fmt.Printf("Selected file full path: %s\n", selectedFile)
}

```  
5- Cobra CLI Integration
``` 
package main

import (
	"fmt"
	"github.com/nabbar/cobra/ui"
	"github.com/spf13/cobra"
)
var choice string
var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "A sample application using",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to myapp!")
		fmt.Println("Thank you for choosing ", choice)
	},
}

func main() {
	tui := ui.New()
	tui.SetQuestions([]ui.Question{
		{
			Text:    "What is your preferred programming language?",
			Options: []string{"Go", "Python", "JavaScript"},
			Handler: func(input string) error {
				choice = input
				return nil
			},
		},
	})
	tui.SetCobra(rootCmd)
	tui.BeforeRun()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

```