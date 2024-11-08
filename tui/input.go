package tui

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var prompt string

func Input(text string) string {
	prompt = strings.TrimSpace(text)
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	return m.(InputModel).textInput.Value()
}

type InputModel struct {
	textInput textinput.Model
	err       error
}

func initialModel() InputModel {
	ti := textinput.New()
	ti.Placeholder = "name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return InputModel{
		textInput: ti,
		err:       nil,
	}
}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		prompt,
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
