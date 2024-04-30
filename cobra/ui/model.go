package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	spfcbr "github.com/spf13/cobra"
	"os"
	"strings"
)

type ui struct {
	cobra       *spfcbr.Command
	questions   []Question
	index       int
	input       string
	cursor      int
	errorMsg    string
	userAnswers []string
}

func (u *ui) SetCobra(cobra *spfcbr.Command) {
	u.cobra = cobra
}

func (u *ui) Init() tea.Cmd {
	return nil
}

func (u *ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	if u.index >= len(u.questions) {
		return u, tea.Quit
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return u, tea.Quit
		}
		switch msg.Type {
		case tea.KeyEnter:
			if u.index < len(u.questions) {
				if len(u.questions[u.index].Options) > 0 {
					err = u.questions[u.index].Handler(u.questions[u.index].Options[u.cursor])
				} else {
					err = u.questions[u.index].Handler(u.input)
				}
				if err != nil {
					u.errorMsg = err.Error()
					u.input = ""
					u.cursor = 0
					return u, nil
				}
				// Store the user's answer
				u.userAnswers = append(u.userAnswers, u.input)
				u.input = ""
				u.index++
				u.cursor = 0
				return u, nil
			} else {
				return nil, tea.Quit
			}
		case tea.KeyDown:
			if len(u.questions[u.index].Options) > 0 {
				u.cursor = (u.cursor + 1) % len(u.questions[u.index].Options)
			}
		case tea.KeyUp:
			if len(u.questions[u.index].Options) > 0 {
				u.cursor = (u.cursor - 1 + len(u.questions[u.index].Options)) % len(u.questions[u.index].Options)
			}
		case tea.KeyTab:

		case tea.KeySpace:

		case tea.KeyBackspace:
			if len(u.input) > 0 {
				u.input = u.input[:len(u.input)-1]
			}
		default:
			u.input += msg.String()
		}
	}

	return u, nil
}

func (u *ui) View() string {
	if u.index >= len(u.questions) {
		return ""
	}
	if u.errorMsg != "" {
		return "" + u.errorMsg + "\n" + u.questions[u.index].Text + u.input
	}

	view := u.questions[u.index].Text

	if len(u.questions[u.index].Options) > 0 {
		view += "\n"
		for i, option := range u.questions[u.index].Options {
			cursor := " "
			if i == u.cursor {
				cursor = "→"
			}
			view += fmt.Sprintf("%s %d. %s\n", cursor, i+1, option)
		}
	} else {
		if len(u.input) > 0 {
			if u.questions[u.index].PasswordType == true {
				view += strings.Repeat("*", len(u.input)) + "\n"
			} else {
				view += "" + u.input + "\n"
			}
		} else {
			view += "\n"
		}
	}

	return view
}

func (u *ui) SetQuestions(questions []Question) {
	u.questions = questions
}

func (u *ui) RunInteractiveUI() {
	u.index = 0
	u.cursor = 0
	p := tea.NewProgram(u)
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func (u *ui) GetAnswers() []string {
	return u.userAnswers
}

func (u *ui) AfterPreRun() {
	if u.cobra == nil {
		fmt.Println("Cobra instance is not set")
	}
	existingPreRun := u.cobra.PreRun
	if existingPreRun != nil {
		u.cobra.PreRun = func(cmd *spfcbr.Command, args []string) {
			existingPreRun(cmd, args)
			u.RunInteractiveUI()
		}
	} else {
		u.cobra.PreRun = func(cmd *spfcbr.Command, args []string) {
			u.RunInteractiveUI()
		}
	}
}

func (u *ui) BeforePreRun() {
	if u.cobra == nil {
		fmt.Println("Cobra instance is not set")
	}
	existingPreRun := u.cobra.PreRun
	if existingPreRun != nil {
		u.cobra.PreRun = func(cmd *spfcbr.Command, args []string) {
			u.RunInteractiveUI()
			existingPreRun(cmd, args)
		}
	} else {
		u.cobra.PreRun = func(cmd *spfcbr.Command, args []string) {
			u.RunInteractiveUI()
		}
	}
}

func (u *ui) BeforeRun() {
	if u.cobra == nil {
		fmt.Println("Cobra instance is not set")
	}
	existingRun := u.cobra.Run
	if existingRun != nil {
		u.cobra.Run = func(cmd *spfcbr.Command, args []string) {
			u.RunInteractiveUI()
			existingRun(cmd, args)
		}
	} else {
		u.cobra.Run = func(cmd *spfcbr.Command, args []string) {
			u.RunInteractiveUI()
		}
	}
}

func (u *ui) AfterRun() {
	if u.cobra == nil {
		fmt.Println("Cobra instance is not set")
	}
	existingRun := u.cobra.Run
	if existingRun != nil {
		u.cobra.Run = func(cmd *spfcbr.Command, args []string) {
			existingRun(cmd, args)
			u.RunInteractiveUI()
		}
	} else {
		u.cobra.Run = func(cmd *spfcbr.Command, args []string) {
			u.RunInteractiveUI()
		}
	}
}
