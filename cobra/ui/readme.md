# UI Package
The main services of the UI package is to help giving a real time UI that is assigned by
questions in a dynamic way.
- The package can be used alone 
- The package can be used with cobra CLI

## Example with cobra CLI
```
package main

import (
	"fmt"
	"github.com/nabbar/golib/cobra"
	"github.com/nabbar/golib/cobra/ui"
	spfcbr "github.com/spf13/cobra"
	"github/sabouaram/testui/release"
)

var (
	tui   ui.UI
	cbr   cobra.Cobra
	input string
)

var rootCmd = &spfcbr.Command{
	Use:   "test",
	Short: "This for test purposes",
	Run: func(cmd *spfcbr.Command, args []string) {
		fmt.Printf("the question input: %v\n", input)
	},
}

func main() {
	tui = ui.New()
	tui.SetQuestions([]ui.Question{
		{
			Text:    "What is your preferred car:",
			Options: []string{"BMW", "Toyota", "Nissan"},
			Handler: func(s string) (e error) {
				input = s
				return nil
			},
		},
	})
	cbr = cobra.New()
	cbr.SetVersion(release.GetVersion())
	cbr.SetFuncInit(func() {
	})
	cbr.Init()
	cbr.Cobra().AddCommand(rootCmd)
	tui.SetCobra(rootCmd)
	tui.BeforeRun()
	err := cbr.Cobra().Execute()
	if err != nil {
		fmt.Println(err)
	}

}

```
In this example the UI package is used with cobra CLI:
 - We define our cobra command
 - We use the golib/cobra pkg to add our cobra command
 - We set our questions using the ui pkg
 - We set the cobra command in the ui
 - Just before the CLI execution we can choose when to run the UI => here we choosed BeforeRun so we can handle the user anwsers after in the Run thanks to the question handler
 -  The ui can be run before or after the cobra defined PreRun or Run functions so we have to use one of the corresponding functions:  
 ```BeforeRun() ```  
 ```AfterRun() ```  
 ```AfterPreRun```  
 ```BeforePreRun```

## Useful General Infos
In the scenario when we have a question without options  
this means that a user input is attended from the user  
by default if we have an input type question we can in the handler  
check the attended input type in case the type is wrong an error should be returned  
the UI pkg in this case will show the error and re-ask the question dynamically  

## Using the UI pkg alone
``` 
package main

import (
	"github.com/nabbar/golib/cobra/ui"
)

func main() {

	tui := ui.New()
	tui.SetQuestions([]ui.Question{
		{
			Text:    "What is your preferred car:",
			Options: []string{"BMW", "Toyota", "Nissan"},
			Handler: func(s string) (e error) {
				return nil
			},
		},
	})
	tui.RunInteractiveUI()
}
```

## Protected Input Question
``` 
q := ui.Question{
		Text: "Enter your password:",
		PasswordType: true,
		Handler: func(s string) error {
			return nil
		},
	}
```