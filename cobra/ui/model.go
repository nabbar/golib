/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	spfcbr "github.com/spf13/cobra"
)

const pageSize = 10

type ui struct {
	cobra       *spfcbr.Command
	questions   []Question
	index       int
	input       string
	cursor      int
	errorMsg    string
	userAnswers []string
	LastMessage string
}

func (u *ui) SetCobra(cobra *spfcbr.Command) {
	u.cobra = cobra
}

func (u *ui) Init() tea.Cmd {
	return nil
}

func (u *ui) SetLastMessage(msg string) {
	u.LastMessage = msg
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
				if u.questions[u.index].FilePath {
					selectedFile := ""
					var fullPath string
					if u.input == "." {
						fullPath, _ = filepath.Abs(u.input)
					} else {
						fullPath = u.input
					}

					if _, err := os.Stat(fullPath); err == nil {
						files, _ := filepath.Glob(filepath.Join(fullPath, "*"))
						var filesList []string
						for _, file := range files {
							fileInfo, err := os.Stat(file)
							if err == nil && !fileInfo.IsDir() {
								filesList = append(filesList, file)
							}
						}
						if len(filesList) > 0 {
							fileIndex := u.cursor
							if fileIndex < len(filesList) {
								selectedFile = filesList[fileIndex]
							}
						}
						err = u.questions[u.index].Handler(selectedFile)
					} else {
						u.errorMsg = "Directory does not exist"
					}
				} else if len(u.questions[u.index].Options) > 0 {
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
				u.userAnswers = append(u.userAnswers, u.input)
				u.input = ""
				u.index++
				u.cursor = 0
				return u, nil
			} else {
				return nil, tea.Quit
			}
		case tea.KeyDown:
			if u.questions[u.index].FilePath {
				if _, err := os.Stat(u.input); err == nil {
					files, _ := filepath.Glob(filepath.Join(u.input, "*"))
					var filesList []string
					for _, file := range files {
						fileInfo, err := os.Stat(file)
						if err == nil && !fileInfo.IsDir() {
							filesList = append(filesList, file)
						}
					}
					if len(filesList) > 0 {
						if u.cursor < len(filesList)-1 {
							u.cursor++
						}
					}
				}
			} else if len(u.questions[u.index].Options) > 0 {
				if u.cursor < len(u.questions[u.index].Options)-1 {
					u.cursor++
				}
			}

		case tea.KeyUp:
			if u.cursor > 0 {
				u.cursor--
			}
		case tea.KeyRight:
			if u.questions[u.index].FilePath {
				if _, err := os.Stat(u.input); err == nil {
					files, _ := filepath.Glob(filepath.Join(u.input, "*"))
					page := u.cursor + pageSize
					if page >= len(files) {
						u.cursor = 0
					} else {
						u.cursor = page
					}
				}
			} else if len(u.questions[u.index].Options) > 0 {
				nextPage := (u.cursor/pageSize + 1) * pageSize
				if nextPage < len(u.questions[u.index].Options) {
					u.cursor = nextPage
				} else {
					u.cursor = len(u.questions[u.index].Options) - 1
				}
			}
		case tea.KeyLeft:
			if u.cursor >= pageSize {
				u.cursor -= pageSize
			} else {
				u.cursor = 0
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
	var dir string
	if u.index >= len(u.questions) {
		return u.LastMessage
	}
	view := u.questions[u.index].Text

	if u.questions[u.index].FilePath {
		dir += u.input
		view += u.input + "\n"
		if u.errorMsg != "" {
			view += "Error: " + u.errorMsg + "\n"
		}
		fileInfo, err := os.Stat(u.input)
		if err == nil && fileInfo.IsDir() {
			files, _ := filepath.Glob(filepath.Join(u.input, "*"))
			var filesList []string
			for _, file := range files {
				fileInfo, err := os.Stat(file)
				if err == nil && !fileInfo.IsDir() {
					filesList = append(filesList, file)
				}
			}
			if len(filesList) > 0 {
				view += "Files in folder:\n"
				totalPages := (len(filesList) + pageSize - 1) / pageSize
				currentPage := u.cursor/pageSize + 1
				start := (currentPage - 1) * pageSize
				end := start + pageSize
				if end > len(filesList) {
					end = len(filesList)
				}
				if start >= len(filesList) {
					u.cursor = 0
					currentPage = 1
					start = 0
					end = pageSize
				}
				for i := start; i < end; i++ {
					cursor := " "
					if i == u.cursor {
						cursor = "→"
					}
					view += fmt.Sprintf("%s %d. %s\n", cursor, i+1, filesList[i])
				}
				view += fmt.Sprintf("\nPage %d/%d\n", currentPage, totalPages)
			} else {
				view += "No files in folder\n"
				u.cursor = 0
			}
		} else {
			if u.input != "" {
				view += "Invalid directory path\n"
			}

		}
	} else if len(u.questions[u.index].Options) > 0 {
		view += "\n"
		totalOptions := len(u.questions[u.index].Options)
		totalPages := (totalOptions + pageSize - 1) / pageSize
		currentPage := u.cursor/pageSize + 1
		start := (currentPage - 1) * pageSize
		end := start + pageSize
		if end > totalOptions {
			end = totalOptions
		}
		if start >= totalOptions {
			u.cursor = 0
			currentPage = 1
			start = 0
			end = pageSize
		}
		for i := start; i < end; i++ {
			cursor := " "
			if i == u.cursor {
				cursor = "→"
			}
			view += fmt.Sprintf("%s %d. %s\n", cursor, i+1, u.questions[u.index].Options[i])
		}
		view += fmt.Sprintf("\nPage %d/%d\n", currentPage, totalPages)
	} else {
		if len(u.input) > 0 {
			if u.questions[u.index].PasswordType {
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
