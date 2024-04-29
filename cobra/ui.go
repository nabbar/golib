package cobra

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type Question struct {
	Text    string
	Options []string
	Handler func(string) error
}

type promptModel struct {
	questions []Question
	index     int
	input     string
	cursor    int
	errorMsg  string
}

func (m *promptModel) Init() tea.Cmd {
	return nil
}

func (m *promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	if m.index >= len(m.questions) {
		return m, tea.Quit
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyEnter:
			if m.index < len(m.questions) {
				if len(m.questions[m.index].Options) > 0 {
					err = m.questions[m.index].Handler(m.questions[m.index].Options[m.cursor])
				} else {
					err = m.questions[m.index].Handler(m.input)
				}
				if err != nil {
					m.errorMsg = err.Error()
					m.input = ""
					m.cursor = 0
					return m, nil
				}
				m.input = ""
				m.index++
				m.cursor = 0
				return m, nil
			} else {
				return nil, tea.Quit
			}
		case tea.KeyDown:
			if len(m.questions[m.index].Options) > 0 {
				m.cursor = (m.cursor + 1) % len(m.questions[m.index].Options)
			}
		case tea.KeyUp:
			if len(m.questions[m.index].Options) > 0 {
				m.cursor = (m.cursor - 1 + len(m.questions[m.index].Options)) % len(m.questions[m.index].Options)
			}
		case tea.KeyTab:

		case tea.KeySpace:

		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			m.input += msg.String()
		}
	}

	return m, nil
}

func (m *promptModel) View() string {
	if m.index >= len(m.questions) {
		return ""
	}
	if m.errorMsg != "" {
		return "" + m.errorMsg + "\n" + m.questions[m.index].Text + m.input
	}

	view := m.questions[m.index].Text

	if len(m.questions[m.index].Options) > 0 {
		view += "\n"
		for i, option := range m.questions[m.index].Options {
			cursor := " "
			if i == m.cursor {
				cursor = "→"
			}
			view += fmt.Sprintf("%s %d. %s\n", cursor, i+1, option)
		}
	} else {
		if len(m.input) > 0 {
			view += "" + m.input + "\n"
		} else {
			view += "\n"
		}
	}

	return view
}
