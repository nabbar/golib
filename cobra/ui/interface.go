package ui

import "github.com/nabbar/golib/cobra"

type Question struct {
	Text    string
	Options []string
	Handler func(string) error
}
type UI interface {
	SetQuestions(questions []Question)
	RunInteractiveUI()
	SetCobra(cobra cobra.Cobra)
	AfterPreRun()
	BeforePreRun()
	AfterRun()
	BeforeRun()
}

func New() UI {
	return &ui{
		cobra:     nil,
		questions: nil,
		index:     0,
		input:     "",
		cursor:    0,
		errorMsg:  "",
	}
}
